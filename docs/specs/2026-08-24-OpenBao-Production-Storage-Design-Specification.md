# Design Specification: OpenBao Production-Grade Storage

**Document ID:** BAO-STORAGE-DES-001
**Version:** 0.1
**Date:** 2026-08-24
**Status:** APPROVED by user 2026-08-26 -- basis for the Technical Specification
**Predecessor:** BAO-STORAGE-REQ-001 (`2026-08-24-OpenBao-Production-Storage-Requirements.md`)

---

## 1. Design philosophy

OpenBao already runs on the target cluster under the right chart with the
right extension points; the failure is a configuration choice (dev mode),
not a missing component. The design therefore changes values and adds
bootstrap mechanics, not infrastructure.

Three principles govern every design element:

- **Reuse before build.** The deployed chart (openbao-0.25.6) already
  ships Raft integrated-storage support, a `dataStorage`
  volumeClaimTemplate, `extraContainers`/`extraVolumes` sidecar hooks,
  and `persistentVolumeClaimRetentionPolicy: Retain`. Every mechanism in
  this design is a chart value or a script; no new image is introduced
  (the unseal sidecar reuses the `quay.io/openbao/openbao:2.5.1` image
  already pulled on the node), and no new standing workload is created
  (NFR-1).
- **Deterministic first.** Bootstrap, migration, and recovery are
  idempotent scripts with fixed ordering, invoked through the existing
  install-script lifecycle (NFR-4). Nothing depends on a human
  remembering a step after the initial bootstrap.
- **Honest limits.** Auto-unseal on a single-node local cluster protects
  persistence semantics, not against a cluster-admin attacker: the
  unseal key lives in etcd and on the host, so anyone with cluster or
  host access can unseal. The local-path PVC shares fate with the host
  disk; node loss is data loss unless a snapshot was taken. Neither
  property is dressed up as more than it is (section 6, section 9).

The design is stated at the level of mechanics and contracts; exact
values files, script code, and flag parsing belong to the Technical
Specification.

---

## 2. Current state (verified evidence)

All findings in this section were measured 2026-08-24 with read-only
`kubectl` / `helm` against context `k3d-openchoreo`; no mutation was
performed.

| Fact | Evidence |
|---|---|
| Helm release `openbao`, ns `openbao`, chart `openbao-0.25.6`, app `v2.5.1`, revision 1, installed 2026-04-11 | `helm list -n openbao` |
| User-supplied values: `injector.enabled=false`; `server.dev.enabled=true`; `server.dev.devRootToken=root`; a `server.postStart` script | `helm get values openbao -n openbao` |
| Container args are `bao server -dev`; env pins `VAULT_DEV_ROOT_TOKEN_ID=root` | StatefulSet spec |
| No PVC in the `openbao` namespace; only an `emptyDir` `home` volume | `kubectl -n openbao get pvc` (none); StatefulSet volumes |
| StatefulSet `updateStrategy: OnDelete`; `persistentVolumeClaimRetentionPolicy` Retain/Retain | StatefulSet spec |
| Single-node cluster: one node `k3d-openchoreo-server-0`, K3s v1.32.9 | `kubectl get nodes` |
| StorageClass `local-path` (default), provisioner `rancher.io/local-path`, reclaim Delete, `WaitForFirstConsumer`, no volume expansion | `kubectl get sc` |
| `openbao-0` shows RESTARTS=3 within 13h age at survey time (recurrence ongoing) | `kubectl -n openbao get pods` |

The chart's relevant extension points (verified in computed values):
`server.dataStorage` (mountPath `/openbao/data`, default size 10Gi,
`storageClass: null` = cluster default), `server.ha.enabled` /
`server.ha.replicas` / `server.ha.raft.enabled` / `server.ha.raft.config`,
`server.extraContainers`, `server.extraVolumes`,
`server.extraInitContainers`, `server.updateStrategyType: OnDelete`.

Downstream consumers that this design must not break:

- `scripts/seed-openbao-m2-paths.sh` -- idempotent seeder for the four
  M2 keys; authenticates as `OPENBAO_TOKEN` (default `root`).
- `scripts/smoke-openbao.sh` -- asserts the four M2 keys:
  `kv/gitea/runners/token`, `kv/flux/gitea-deploy-key`,
  `secret/apps/hello-m2/dev/example-secret`,
  `kv/apps/hello-m2/dev/example-secret`.
- `iac/modules/external-secrets-wiring/main.tf` -- ClusterSecretStore
  `openbao-kv` (`http://openbao.openbao.svc:8200`, path `kv`, v2) using
  the token in Secret `openbao-root-token` (ns `external-secrets`),
  currently the literal string `root`.
- OpenChoreo's default ClusterSecretStore -- reads the `secret/` mount
  via the `kubernetes` auth method, policies, and roles that the
  postStart script writes on every pod start.
- The postStart script additionally seeds eleven static platform
  secrets under `secret/` (npm/docker/github placeholders, backstage,
  observer, rca, opensearch credentials) that observability-plane and
  backstage consumers rely on.

---

## 3. Architecture overview

Target topology: one OpenBao server pod with Raft integrated storage on
a local-path PVC, plus a co-located unseal sidecar in the same pod. All
consumers keep their existing addresses, mounts, and auth paths; the
only consumer-visible change is that the root token's value changes
from the literal `root` to a generated token (section 8).

```text
  k3d-openchoreo-server-0 (single node)
  +-------------------------------------------------------------+
  | pod/openbao-0 (ns openbao)                                  |
  |                                                             |
  |  +-----------------------+   +---------------------------+  |
  |  | container: openbao    |   | sidecar: openbao-unseal   |  |
  |  | image 2.5.1           |   | same image, shell loop    |  |
  |  | bao server (raft)     |   | watches `bao status`;     |  |
  |  |                       |   | unseals when sealed       |  |
  |  +----------+------------+   +-------------+-------------+  |
  |             | volumeClaimTemplate          | secret vol  |  |
  |  +----------v------------+   +---------------v-----------+  |
  |  | PVC data-openbao-0    |   | Secret openbao-unseal-key |  |
  |  | local-path, 1Gi, RWO  |   | (bootstrap-created)       |  |
  |  +-----------------------+   +---------------------------+  |
  +-------------------------------------------------------------+

  host custody (NFR-3): ~/.rational-reserve/openbao/
    unseal-key   mode 600
    root-token   mode 600
  host backups (section 9): ~/.rational-reserve/backups/openbao/
    raft snapshot files
```

Design elements (D-01 through D-09) follow; each names the requirement
it implements and its evidence.

### D-01 -- Raft integrated storage via chart ha.raft (FR-1, REQ option A)

Values change: `server.dev.enabled: false`, `server.ha.enabled: true`,
`server.ha.replicas: 1`, `server.ha.raft.enabled: true`,
`server.ha.raft.config` carrying `storage "raft" { path =
"/openbao/data" }`. Raft is chosen over the `file` backend (REQ option
B, fallback) because it keeps the production topology shape -- the
upgrade path to 3 nodes is a replicas bump plus two PVCs, not a
migration. The ha/raft chart path with `replicas: 1` is the chart's
supported single-node Raft pattern: a one-node Raft cluster in which
the sole member is the voter.

**Implements:** FR-1. **Evidence:** chart computed values (section 2);
REQ section 6 option A verdict.

### D-02 -- PVC sizing and StorageClass (FR-1, NFR-1)

`server.dataStorage.enabled: true`, `size: 1Gi`,
`storageClass: local-path` (set explicitly rather than relying on the
cluster default, for determinism), `accessMode: ReadWriteOnce`. 1Gi is
sized for the actual content: four M2 keys, eleven platform
placeholder secrets, auth/policy metadata, and Raft overhead (log,
snapshots, WAL) -- kilobytes of payload; 1Gi is headroom, not need.
The local-path provisioner does not support volume expansion (verified,
section 2), so the size cannot be grown in place; 1Gi against a
secret-count growth of orders of magnitude is still trivially
sufficient, and a resize path (snapshot, new PVC, restore) exists if
that assumption ever fails (section 9).

**Implements:** FR-1, NFR-1. **Evidence:** StorageClass survey
(section 2); secret inventory (section 2).

### D-03 -- Single-node, not 3-node Raft (NFR-1, NFR-2)

Rejected alternative: `replicas: 3`. On a single-node k3d cluster all
three pods land on `k3d-openchoreo-server-0` and all three local-path
PVCs land on the same host disk, so the quorum buys zero failure-domain
coverage -- every failure that kills one pod (node flap, Colima stop,
containerd down; REQ section 2) kills all three. The cost is real:
triple the memory footprint against the 2 vCPU host budget and the
NFR-1 footprint bound, plus Raft leader-election noise on every
restart. A single voter has no election latency, which also serves the
NFR-2 recovery bound. If a multi-node cluster ever materializes, D-01's
config shape scales by raising replicas.

**Implements:** NFR-1, NFR-2. **Evidence:** single node (section 2);
REQ section 2 failure events are all node-scoped.

### D-04 -- Automatic unseal via in-pod sidecar (FR-2, OQ-1 disposition)

OQ-1 asks: Shamir keys in `~/.rational-reserve/` with manual unseal, vs
auto-unseal via a local file-based key. This design dispositions OQ-1
in favor of automatic unseal, because manual unseal contradicts the
requirements' own goal and bounds: the REQ Goal is that deleting
`openbao-0` is "a non-event" with "zero manual reseed steps," and NFR-2
bounds recovery from pod deletion to under 2 minutes. A human typing
`bao operator unseal` after every node flap -- the exact events in REQ
section 2 -- is a manual step on the critical path and cannot meet
either.

Mechanism: a sidecar container in the openbao-0 pod (chart
`server.extraContainers`, same `quay.io/openbao/openbao:2.5.1` image --
no new image dependency) running a shell loop: poll `bao status`
against `http://127.0.0.1:8200`; when sealed, read the unseal key from
a mounted Secret volume (`openbao-unseal-key`) and run `bao operator
unseal`. The sidecar is inside the existing openbao-0 pod, so no new
standing workload is created (NFR-1); its footprint is a sleeping shell
plus the bao binary, bounded by explicit resource requests/limits
(sized in the Technical Specification against the <100 MiB NFR-1
budget; design-time assertion, to be measured at acceptance).

Seal configuration: Shamir with `key-shares=1, key-threshold=1`. A
5-of-3 split on a single-operator local cluster is ceremony -- the
shares would be stored in the same `~/.rational-reserve/` directory
under one host account, providing no independent custody.

Honest limits (stated, not hidden): with the unseal key in a Kubernetes
Secret and on the host, encryption-at-rest protects the data only
against actors who hold neither cluster nor host access. That is the
correct posture for a local single-node development platform; the
documented upgrade path for a stronger posture is a KMS or transit
seal, which requires infrastructure this environment deliberately does
not have (a second Bao for transit would be a new standing workload and
a bootstrapping circle; cloud KMS violates the self-hosted constraint).

**Implements:** FR-2. **Evidence:** REQ Goal + NFR-2 (manual unseal
infeasible); chart `extraContainers`/`extraVolumes` extension points
(section 2).

### D-05 -- Key custody (NFR-3)

Bootstrap writes the unseal key and root token to
`~/.rational-reserve/openbao/unseal-key` and
`~/.rational-reserve/openbao/root-token`, directory mode 700, file mode
600, host-side -- matching the existing custody pattern
(`~/.rational-reserve/m1-gitea-admin-password`,
`~/.rational-reserve/logs/`). The host copy is the recovery source of
truth; the cluster copy (`Secret openbao-unseal-key`, ns `openbao`) is
created from it by the install script and exists only to feed the D-04
sidecar. Neither location enters git; `~/.rational-reserve/` is already
outside all remotes, and the provenance listing is updated to record
the new custody artifacts (NFR-3). Keys never enter git (FR-2).

**Implements:** FR-2, NFR-3. **Evidence:** existing
`~/.rational-reserve/` custody pattern; REQ NFR-3.

### D-06 -- One-time bootstrap replaces the postStart script (FR-3, FR-6)

Today's postStart script authenticates as `BAO_TOKEN=root` and
re-establishes auth methods, policies, roles, and eleven platform
secrets on every pod start -- correct behavior for ephemeral inmem
storage, broken under persistent storage (the token `root` will not
exist, and the re-writes are unnecessary because Raft persists them).
The values change therefore removes `server.postStart` entirely and
moves its content into a new repo-owned bootstrap script
(`scripts/bootstrap-openbao-persistent.sh`, created by the Technical
Specification), run once per storage lifetime, which:

1. Initializes the server if `bao operator init -status` reports
   uninitialized (`key-shares=1, key-threshold=1`), else recovers the
   existing key material from the D-05 host custody files.
2. Persists custody files per D-05 and (re)creates the
   `openbao-unseal-key` Secret from them.
3. Waits for unseal (D-04 sidecar) and readiness.
4. Re-applies the postStart's durable configuration exactly once:
   `kubernetes` auth method and config, the two
   `openchoreo-secret-reader/writer` policies, the two auth roles --
   all idempotent writes against persisted Raft state.
5. Re-seeds the eleven static platform secrets under `secret/`
   (idempotent; preserves existing values unless rotation is explicitly
   requested), keeping observability-plane and backstage consumers
   whole.
6. Invokes `scripts/seed-openbao-m2-paths.sh` for the four M2 keys
   (FR-6: `kv/gitea/runners/token` and `secret/apps/...` both covered).
7. Updates Secret `openbao-root-token` (ns `external-secrets`) to the
   generated root token (section 8).

`seed-openbao-m2-paths.sh` changes in one respect: its default token
source becomes the D-05 custody file (`OPENBAO_TOKEN` env still
overrides). It thereby becomes a one-time bootstrap component instead
of a recovery tool (FR-3); its idempotence is unchanged, so running it
again is always safe.

**Implements:** FR-3, FR-6. **Evidence:** postStart contents (section
2); `seed-openbao-m2-paths.sh` current behavior (section 2).

### D-07 -- Root-token handoff to ExternalSecrets (FR-6, NFR-4)

The ClusterSecretStore `openbao-kv` reads its token from Secret
`openbao-root-token` (ns `external-secrets`), today the literal `root`
(`iac/modules/external-secrets-wiring/main.tf`). After bootstrap, that
Secret must hold the generated root token. The update is performed by
the bootstrap script (D-06 step 7) -- a scripted lifecycle action, not
an ad-hoc mutation (NFR-4). Token auth is retained for this change;
migrating ExternalSecrets to `kubernetes` auth (the module's own
comment anticipates it) is a separate hardening change, explicitly out
of G5 scope, recorded in section 13. Design-time assertion, verified at
acceptance: the ExternalSecrets' 1h refreshInterval means a brief
stale-token window is self-healing and no tofu re-apply is needed for
the token value change.

**Implements:** FR-6, NFR-4. **Evidence:**
`iac/modules/external-secrets-wiring/main.tf` (section 2).

### D-08 -- FR-4 inverse-proof lane (FR-4)

`scripts/smoke-openbao.sh` gains an opt-in lane (flag `--with-restart`;
default invocation keeps today's fast four-key presence check so the
routine smoke stays lightweight):

1. Record the current values of the four M2 keys (exact values, not
   mere presence).
2. `kubectl delete pod openbao-0`.
3. Wait for the pod to be Ready -- which, because the readiness probe
   fails while sealed, also proves the D-04 sidecar unsealed the server
   without any human step.
4. Re-read the four keys and assert value equality with step 1.
5. Assert that no seed script was run during the lane (the lane itself
   is the only actor; it never invokes the seeder).

Inverse-proof property: pre-change, the lane must fail -- inmem storage
loses all four keys on pod deletion (verified twice 2026-08-24 per REQ
section 4, FR-4). Post-change, it must pass. The lane is executed once
before the migration (expected FAIL, recorded as evidence) and once
after (expected PASS, acceptance criterion 1). Because the lane deletes
a pod, it is a heavy operation: per the serialized-heavy-operations
convention it runs standalone or as a dedicated lane in
`smoke-all.sh`, never concurrently with other smokes (exact wiring in
the Technical Specification).

**Implements:** FR-4. **Evidence:** REQ FR-4 (pre-change failure
verified twice); StatefulSet readiness-probe behavior (section 2).

### D-09 -- Teardown preserves the PVC (FR-5)

`scripts/teardown-m2.sh` currently has no OpenBao handling (verified:
no openbao or pvc references), so it cannot destroy what it does not
touch -- but FR-5 requires the guarantee to be explicit. The Technical
Specification adds an OpenBao section to the teardown script: default
behavior preserves the `data-openbao-0` PVC, the `openbao-unseal-key`
Secret, and the D-05 host custody files; a new explicit
`--wipe-secrets` flag deletes all three (with a confirmation prompt on
an interactive terminal). The chart's
`persistentVolumeClaimRetentionPolicy: Retain` (already in effect,
verified section 2) additionally protects the PVC if the StatefulSet
itself is ever deleted.

**Implements:** FR-5. **Evidence:** teardown-m2.sh survey; StatefulSet
retention policy (section 2).

---

## 4. Helm values change set

The release was installed from the sibling openchoreo checkout; this
design moves ownership of the values into this repo as
`scripts/openbao-values.yaml` (created by the Technical Specification)
so the change flows through an install script (NFR-4) and the desired
state is reviewable in git. From -> to:

| Value | From (verified) | To |
|---|---|---|
| `injector.enabled` | `false` | `false` (unchanged) |
| `server.dev.enabled` | `true` | `false` |
| `server.dev.devRootToken` | `root` | removed |
| `server.postStart` | auth/policy/secret seeding script | removed (replaced by D-06 bootstrap) |
| `server.ha.enabled` | `false` | `true` |
| `server.ha.replicas` | (default 3) | `1` (D-03) |
| `server.ha.raft.enabled` | `false` | `true` |
| `server.ha.raft.config` | (default raft example) | raft storage at `/openbao/data` (D-01) |
| `server.dataStorage` | inert under dev | `enabled: true`, `size: 1Gi`, `storageClass: local-path` (D-02) |
| `server.extraContainers` | `null` | unseal sidecar (D-04) |
| `server.extraVolumes` | `[]` | Secret volume `openbao-unseal-key` (D-04) |
| `server.updateStrategyType` | `OnDelete` | `OnDelete` (unchanged; migration deletes the pod explicitly, section 5) |

---

## 5. Migration path: dev-mode inmem to Raft (FR-3, NFR-4)

No data migration is required or possible: inmem contents are
ephemeral by definition, and every secret the platform needs is either
static (platform placeholders), derivable (Flux deploy key from the m1
password file; runner token recoverable from the existing Kubernetes
Secret per the seeder's fallback), or rotatable (demo app secret). The
migration is a storage swap plus reseed, executed by a new
`scripts/install-openbao-storage.sh` (created by the Technical
Specification), wired as a task in `scripts/install-m2.sh` ahead of the
existing seed task. Steps:

1. Run the D-08 inverse-proof lane; record the expected FAIL (evidence
   that the gate can fail -- inverse-proof convention).
2. `helm upgrade` the release with `scripts/openbao-values.yaml`
   (section 4). Because `updateStrategyType: OnDelete`, the running pod
   keeps the old template; the script then deletes `openbao-0` so the
   StatefulSet recreates it with Raft config and the PVC.
3. Run the D-06 bootstrap (init, custody, unseal secret, auth/policies,
   platform secrets, M2 seed, root-token handoff).
4. Run the D-08 lane again; record the expected PASS (acceptance 1).
5. Run `smoke-openbao.sh` and `smoke-all.sh` (acceptance 2 and 3), each
   serialized per the heavy-operations convention.

Ordering note for a fresh machine: on a new install the same script is
idempotent -- step 1's lane fails as expected on dev mode only if dev
mode was ever deployed; on a greenfield run the script detects an
already-persistent backend and skips to bootstrap-verification.

---

## 6. Unseal strategy summary and custody decision

Disposition of OQ-1 (design decision, pending approval of this
document): **automatic unseal by in-pod sidecar (D-04), Shamir 1-of-1,
dual custody -- host files under `~/.rational-reserve/openbao/` (mode
600, recovery source of truth, NFR-3) and a cluster Secret feeding the
sidecar.** Manual Shamir unseal is rejected for this environment
because it violates the REQ Goal and NFR-2 (section D-04). The honest
security boundary is stated in D-04: this protects persistence, not
against cluster/host-level actors; KMS/transit seals are the documented
upgrade path and are out of scope for G5.

---

## 7. Backup, restore, and disaster recovery (OQ-2 disposition)

Disposition of OQ-2: **snapshot-on-script, weekly cadence by
convention, no CronJob** -- a CronJob would be a new standing workload
(NFR-1) and this host is routinely stopped. Two artifacts:

- `scripts/backup-openbao.sh` (Technical Specification): `bao operator
  raft snapshot save` via `kubectl exec`, written to
  `~/.rational-reserve/backups/openbao/openbao-YYYYMMDD-HHMMSS.snap`
  (mode 600). Run weekly by operator convention and before any cluster
  surgery; cadence recorded in the runbook.
- Restore procedure (runbook material for the Technical Specification):
  reinstall the release (section 5 steps 2-3 with a fresh, empty PVC),
  then `bao operator raft snapshot restore <snap>` before the platform
  reseed steps.

Disaster recovery when no snapshot exists (the honest common case on a
local platform): rerun section 5 from step 2 with a fresh PVC; the D-06
bootstrap re-derives everything -- platform placeholders are static,
the Flux deploy key comes from `~/.rational-reserve/m1-gitea-admin-password`,
the runner token from its surviving Kubernetes Secret or Gitea, and the
demo app secret rotates. Worst-case data loss is therefore the delta
since the last snapshot plus any secret written outside the bootstrap
inventory -- currently none are known (design-time assertion; the
inventory is re-verified in the Technical Specification).

Honest limits: local-path volumes live on the host disk, and snapshots
written to `~/.rational-reserve/backups/` share that fate. Host disk
loss without an off-host copy loses both. Off-host snapshot copy is
out of scope for G5.

---

## 8. Root-token and auth posture

The generated root token replaces the literal `root` everywhere it was
assumed: the seeder's default (D-06), the ExternalSecrets handoff
(D-07), and any smoke tooling that authenticates (`smoke-openbao.sh`
reads the D-05 custody file when `OPENBAO_TOKEN` is unset). The
`kubernetes` auth method, roles, and policies persist in Raft after D-06
applies them once; OpenChoreo's ClusterSecretStore is unaffected. The
dev-mode convenience of a well-known token ends -- that is the point of
FR-2 -- and the only new operational rule is: authenticate via the D-05
custody files.

---

## 9. Rollback plan

Rollback restores the previous dev-mode behavior without destroying the
new persistent state:

1. `helm rollback openbao 1` restores revision 1 values (dev mode,
   postStart).
2. Delete `openbao-0` (OnDelete strategy) so the old template takes
   effect.
3. Run `scripts/seed-openbao-m2-paths.sh` once -- it resumes its
   pre-change role as recovery tool against the fresh inmem backend.
4. Restore Secret `openbao-root-token` to the literal `root` (the
   rollback script performs this; the generated token is invalid in dev
   mode).

The PVC, unseal Secret, custody files, and snapshots are all preserved
through rollback (D-09 guarantees; chart retention policy), so a
subsequent re-upgrade resumes from the persisted Raft state with no
reseed. Rollback of the scripts themselves is ordinary git revert; no
tofu state is touched by this change (no tofu resources are modified --
the external-secrets-wiring module is untouched).

---

## 10. Resource and recovery budget (NFR-1, NFR-2)

Design-time assertions, to be measured at acceptance (not executed
here):

- Memory delta vs today: PVC bookkeeping plus Raft overhead plus the
  sidecar's sleeping shell -- bounded <100 MiB by construction
  (sidecar carries explicit limits; Raft's steady-state overhead at
  this secret volume is single-digit MiB in upstream documentation,
  UNVERIFIED locally until acceptance).
- Recovery time from `kubectl delete pod openbao-0` to
  secrets-available: pod reschedule on the same node (image already
  pulled), Raft replay of kilobytes, sidecar unseal on its first poll
  interval -- expected well under the 2-minute NFR-2 bound; the D-08
  lane measures it end-to-end at acceptance.

---

## 11. Phasing

### Phase 0 -- Approval

- **Scope:** approve Requirements and this Design Specification;
  disposition OQ-1 and OQ-2 as decided here (or amend).
- **Exit criteria:** user approval recorded; Technical Specification
  drafted from sections 4-9.

### Phase 1 -- Technical Specification and implementation

- **Scope:** values file, bootstrap/backup scripts, seeder and smoke
  changes, teardown flag, install-m2 wiring, provenance update.
- **Exit criteria:** review complete; no cluster mutation has occurred
  (per REQ acceptance criterion 4).

### Phase 2 -- Execution and acceptance

- **Scope:** section 5 migration steps 1-5, in order, serialized.
- **Exit criteria:** REQ acceptance criteria 1-3 met with measured
  results; SESSION_HANDOFF snapshot addendum per standing directive.

---

## 12. Traceability

| Design element | Implements | Evidence |
|---|---|---|
| D-01 Raft via chart ha.raft | FR-1 | Chart computed values (s2); REQ s6 option A |
| D-02 PVC sizing/StorageClass | FR-1, NFR-1 | StorageClass survey (s2); secret inventory (s2) |
| D-03 Single-node choice | NFR-1, NFR-2 | Single node verified (s2); REQ s2 node-scoped failures |
| D-04 Sidecar auto-unseal | FR-2 (OQ-1) | REQ Goal + NFR-2; chart extension points (s2) |
| D-05 Key custody | FR-2, NFR-3 | Existing `~/.rational-reserve/` pattern; REQ NFR-3 |
| D-06 Bootstrap replaces postStart | FR-3, FR-6 | postStart contents (s2); seeder behavior (s2) |
| D-07 Root-token handoff | FR-6, NFR-4 | `external-secrets-wiring/main.tf` (s2) |
| D-08 Inverse-proof lane | FR-4 | REQ FR-4 (pre-change failure verified twice); readiness probe (s2) |
| D-09 Teardown preserves PVC | FR-5 | teardown-m2.sh survey; retention policy (s2) |
| Section 4 values change set | FR-1, FR-3, NFR-4 | `helm get values` user-supplied (s2) |
| Section 5 migration path | FR-3, NFR-4 | install-m2.sh task structure (s2); OnDelete strategy (s2) |
| Section 6 unseal decision | FR-2, NFR-2, NFR-3 | OQ-1 disposition (D-04, D-05) |
| Section 7 backup/DR | FR-5 (data lifecycle), OQ-2 | Snapshot mechanics; derivation inventory (s2) |
| Section 8 auth posture | FR-2, FR-6 | Consumer inventory (s2) |
| Section 9 rollback | NFR-4 (lifecycle discipline) | helm revision 1 (s2); retention policy (s2) |
| Section 10 budgets | NFR-1, NFR-2 | Design-time assertions; acceptance measurement via D-08 |

NFR-4 is additionally satisfied structurally: every cluster-affecting
step in this design is executed by an install/lifecycle script
(section 5, D-09), matching the repo's tofu-guard and
no-ad-hoc-mutation conventions.

---

## 13. Carried open questions and follow-ups

- **OQ-3 (from REQ):** whether CI (act-runner dind) holds any dev-mode
  dependency. Design-time assertion: none -- CI consumes the runner
  registration token only indirectly (the runner registers before jobs
  run) and no workflow step talks to OpenBao. Assigned to the Technical
  Specification for verification, per REQ section 7.
- **Follow-up (out of G5 scope):** migrate ExternalSecrets from
  root-token auth to `kubernetes` auth, as the
  external-secrets-wiring module's own comment anticipates. Recorded
  here so it is not lost; belongs in a later hardening change.
- **Follow-up (out of G5 scope):** off-host snapshot copies (section 7
  honest limits).

---

**End of Design Specification**

The Technical Specification (exact `scripts/openbao-values.yaml`,
`bootstrap-openbao-persistent.sh`, `backup-openbao.sh`,
`install-openbao-storage.sh`, seeder/smoke/teardown diffs, sidecar
resource sizing, and the OQ-3 verification) completes the triad. No
cluster mutation occurs before that document is approved (REQ
acceptance criterion 4).
