# Technical Specification: OpenBao Production-Grade Storage

**Document ID:** BAO-STORAGE-TECH-001
**Version:** 0.2
**Date:** 2026-08-26
**Status:** APPROVED by user 2026-08-26 -- v0.2 folds in simulation corrections C5-C9 (BAO-STORAGE-SIM-001); cleared for live implementation
**Predecessors:** BAO-STORAGE-REQ-001 (`2026-08-24-OpenBao-Production-Storage-Requirements.md`), BAO-STORAGE-DES-001 (`2026-08-24-OpenBao-Production-Storage-Design-Specification.md`, approved 2026-08-26)

---

## Purpose, evidence discipline, and traceability

This document is the implementation-grade companion to BAO-STORAGE-DES-001. It
specifies the exact helm values file, the full text of every new or changed
script, the migration runbook, the FR-4 inverse-proof lane, the rollback
procedure, and the acceptance checklist. It changes nothing by itself: the only
file created at spec time is this document. Implementation follows Phase 1/2 of
DES-001 section 11.

**Evidence discipline.** Cluster and chart facts cited below were measured
2026-08-26 with read-only commands against context `k3d-openchoreo`, or read
from the chart tarball `openbao-0.25.6` fetched to `/tmp/obchart` with
`helm pull openbao/openbao --version 0.25.6` (repo
`https://openbao.github.io/openbao-helm`). Anything not executed is marked
*spec-time assertion*; nothing in this document is presented as a measured
result that was not measured.

**Verified facts used below (all measured 2026-08-26 unless noted).**

- Helm release `openbao`, namespace `openbao`, chart `openbao-0.25.6`, app
  `v2.5.1`, revision 1, installed 2026-04-11 (`helm list -n openbao`).
- User-supplied values in full (`helm get values openbao -n openbao`):
  `injector.enabled=false`; `server.dev.enabled=true`;
  `server.dev.devRootToken=root`; `server.postStart` = the seeding script
  quoted verbatim in section 1.2.
- StatefulSet `openbao`: `updateStrategy.type=OnDelete`;
  `persistentVolumeClaimRetentionPolicy={whenDeleted: Retain, whenScaled:
  Retain}`; readiness probe `exec ["/bin/sh","-ec","bao status
  -tls-skip-verify"]` (initialDelaySeconds 5, periodSeconds 5,
  failureThreshold 6); container args `bao server -dev`.
- No PVC in namespace `openbao` (`kubectl -n openbao get pvc` -> "No resources
  found"). Secret `openbao-root-token` (ns `external-secrets`) currently holds
  the literal string `root` (read back and base64-decoded 2026-08-26).
- Chart defaults from the pulled 0.25.6 tarball: `server.dataStorage`
  (mountPath `/openbao/data`, size 10Gi, `storageClass: null`),
  `server.ha.replicas: 3`, `server.ha.raft.enabled: false`,
  `server.updateStrategyType: "OnDelete"` (values.yaml:393),
  `server.postStart: []` (values.yaml:671),
  `server.persistentVolumeClaimRetentionPolicy: {}` (values.yaml:917),
  `snapshotAgent.enabled: false`.
- Chart template semantics (0.25.6): `server.extraVolumes` entries
  `{type, name, defaultMode}` render as volumes named `userconfig-<name>`
  (`_helpers.tpl`, `define "openbao.volumes"`, lines 205-214) and are
  auto-mounted into the *server* container at `/openbao/userconfig/<name>`
  (`define "openbao.mounts"`, lines 273-277); `server.extraContainers` is
  appended verbatim to the StatefulSet container list
  (`server-statefulset.yaml:221-223`); the volumeClaimTemplate named `data`
  renders only when mode is not `dev` and `ha.raft.enabled` is true
  (`define "openbao.volumeclaims"`, lines 288-307); `server.ha.raft.config`
  is the exact key path rendered into ConfigMap `openbao-config`
  (`define "openbao.config"`, line 1136).
- Chart creates a PodDisruptionBudget in ha mode; with `replicas: 1` the chart
  special-cases `maxUnavailable: 0` (`define "openbao.pdb.maxUnavailable"`,
  lines 132-140). A PDB with maxUnavailable 0 blocks voluntary *evictions*
  (drain), not direct `kubectl delete pod`, so the FR-4 lane is unaffected.
  Recorded here so the operator is not surprised on a future node drain; left
  at chart default (minimal change).
- Chart ships an optional `snapshotAgent` CronJob (defaults to S3 targets).
  It stays `enabled: false`: DES-001 section 7 (OQ-2 disposition) specifies
  snapshot-on-script, and NFR-1 forbids new standing workloads.

**One unresolved origin (UNVERIFIED, does not block):** the live StatefulSet
reports `persistentVolumeClaimRetentionPolicy` Retain/Retain while the computed
helm values show `{}` and the chart default is `{}`; the template does not
render the field from an empty map. The current retention setting's origin is
unresolved at spec time. This specification therefore pins the policy
*explicitly* in the new values file (section 3), so post-change behavior no
longer depends on the unexplained state.

---

## 1. Current state in detail

### 1.1 Consumer inventory (verified, repo + cluster)

| Consumer | Path / mechanism | Evidence |
|---|---|---|
| ExternalSecrets ClusterSecretStore `openbao-kv` | `http://openbao.openbao.svc:8200`, path `kv`, v2, token from Secret `openbao-root-token` (ns `external-secrets`) | `iac/modules/external-secrets-wiring/main.tf:16-39`; live Secret value `root` |
| ExternalSecret `gitea-deploy-key` (flux-system) | `kv/flux/gitea-deploy-key`, refreshInterval 1h | `iac/modules/external-secrets-wiring/main.tf:41-60` |
| ExternalSecret `hello-m2-example-secret` (openchoreo-data-plane) | `kv/apps/hello-m2/dev/example-secret`, refreshInterval 1h | `iac/modules/external-secrets-wiring/main.tf:62-81` |
| ExternalSecret `gitea-runner-token` (gitea-runners) | `kv/gitea/runners/token`, refreshInterval 1h; act-runner consumes the synced K8s Secret via `existingSecret` | `iac/modules/gitea-runner/main.tf:5-24,47-53` |
| OpenChoreo default ClusterSecretStore | `secret/` mount via `kubernetes` auth method, policies, roles (written today by postStart) | postStart script (section 1.2) |
| Observability-plane / backstage consumers | thirteen static placeholder secrets under `secret/` | postStart script (section 1.2) |
| `scripts/seed-openbao-m2-paths.sh` | idempotent seeder, authenticates as `OPENBAO_TOKEN` default `root` | `scripts/seed-openbao-m2-paths.sh:12` |
| `scripts/smoke-openbao.sh` | asserts presence of the four M2 keys, token default `root` | `scripts/smoke-openbao.sh:11` |

### 1.2 The postStart script, verbatim (measured 2026-08-26)

This is the exact current `server.postStart` value from
`helm get values openbao -n openbao`. The bootstrap script in section 4.1
carries this content forward; the values change removes it from the release.

```sh
sleep 5
export BAO_ADDR=http://127.0.0.1:8200
export BAO_TOKEN=root

bao auth enable kubernetes 2>/dev/null || true
bao write auth/kubernetes/config \
  kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"

bao policy write openchoreo-secret-reader-policy - <<POLICY
path "secret/data/*" { capabilities = ["read"] }
path "secret/metadata/*" { capabilities = ["list", "read"] }
POLICY

bao policy write openchoreo-secret-writer-policy - <<POLICY
path "secret/data/*" { capabilities = ["create", "read", "update", "delete"] }
path "secret/metadata/*" { capabilities = ["create", "read", "update", "delete", "list"] }
POLICY

bao write auth/kubernetes/role/openchoreo-secret-reader-role \
  bound_service_account_names=default \
  bound_service_account_namespaces="dp*" \
  policies=openchoreo-secret-reader-policy ttl=20m

bao write auth/kubernetes/role/openchoreo-secret-writer-role \
  bound_service_account_names="*" \
  bound_service_account_namespaces="openbao,openchoreo-workflow-plane" \
  policies=openchoreo-secret-writer-policy ttl=20m

# Sample apps
bao kv put secret/npm-token value="fake-npm-token-for-development"
bao kv put secret/docker-username value="dev-user"
bao kv put secret/docker-password value="dev-password"
bao kv put secret/github-pat value="fake-github-token-for-development"
bao kv put secret/username value="dev-user"
bao kv put secret/password value="dev-password"

# Backstage (web console)
bao kv put secret/backstage-backend-secret value="local-dev-backend-secret"
bao kv put secret/backstage-client-secret value="backstage-portal-secret"
bao kv put secret/backstage-jenkins-api-key value="placeholder-not-in-use"
# Observer (observability)
bao kv put secret/observer-oauth-client-secret value="openchoreo-observer-resource-reader-client-secret"
# RCA Agent (observability)
bao kv put secret/rca-oauth-client-secret value="openchoreo-rca-agent-secret"
# OpenSearch (observability)
bao kv put secret/opensearch-username value="admin"
bao kv put secret/opensearch-password value="ThisIsTheOpenSearchPassword1"
```

**Count clarification (carried to section 11):** DES-001 and REQ-derived notes
say "eleven platform secrets"; the measured postStart writes **thirteen** kv
keys under `secret/` (6 sample-app, 3 backstage, 1 observer, 1 rca, 2
opensearch). This specification preserves all thirteen verbatim. "Preserve" is
the binding requirement; the count in the design was a summary, not a subset.

---

## 2. OQ-3 resolution: CI dependency on dev-mode OpenBao

REQ OQ-3 asks whether CI (act-runner dind) holds any dev-mode dependency that a
persistent backend would break. **Resolved at spec time from repo evidence:
none.**

- No CI workflow references OpenBao or Vault at all:
  `grep -ri 'bao|vault'` over `seed-repos/**/*.yaml` and
  `iac/**/*.yaml/yml/tf/sh` (run 2026-08-26) matches only the tofu modules
  `gitea-runner/main.tf` and `external-secrets-wiring/main.tf` -- never a
  workflow step. The canonical pipeline (`seed-repos/hello-m2/.gitea/workflows/ci.yaml`)
  and the template (`iac/templates/ci.yaml`) contain no OpenBao call.
- The runner consumes its registration token from the Kubernetes Secret
  `gitea-runners/gitea-runner-token` via the chart's `existingSecret`
  (`iac/modules/gitea-runner/main.tf:47-53`). That Secret is synced by
  ExternalSecrets from `kv/gitea/runners/token`
  (`iac/modules/gitea-runner/main.tf:5-24`). Once synced, the K8s Secret
  persists independently of OpenBao availability; the migration preserves the
  backing key in Raft, so the next 1h refresh succeeds unchanged.
- The migration briefly restarts OpenBao. A CI job already in flight does not
  contact OpenBao; a runner that (re)registers during the window reads the
  already-synced K8s Secret, not OpenBao directly.

Residual measured-at-implementation check (Phase 2, after migration step 5):

```sh
kubectl -n gitea-runners get pods            # runner pod Running
kubectl -n gitea-runners get secret gitea-runner-token -o jsonpath='{.data.token}' | base64 -d | wc -c
# expect: non-zero token length
./scripts/smoke-actions.sh                   # existing lane; must stay green
```

---

## 3. The helm values change

### 3.1 New file: `scripts/openbao-values.yaml` (byte-exact)

This file is the COMPLETE user-supplied values set. It is applied without
`--reuse-values`, so every key not listed returns to chart default -- which is
exactly what removes `server.dev.devRootToken` (default empty) and
`server.postStart` (default `[]`, values.yaml:671). Exact key paths verified
against the 0.25.6 templates (see "Verified facts").

```yaml
# scripts/openbao-values.yaml
# G5 (BAO-STORAGE-DES-001 section 4): OpenBao production-grade storage.
# Raft integrated storage on a 1Gi local-path PVC, single replica, with an
# in-pod auto-unseal sidecar fed by the bootstrap-created Secret
# openbao-unseal-key. Applied ONLY by scripts/install-openbao-storage.sh:
#   helm upgrade openbao openbao/openbao --version 0.25.6 \
#     --namespace openbao --values scripts/openbao-values.yaml
injector:
  enabled: false
server:
  dev:
    enabled: false
  ha:
    enabled: true
    replicas: 1
    raft:
      enabled: true
      config: |
        ui = true

        listener "tcp" {
          tls_disable = 1
          address = "[::]:8200"
          cluster_address = "[::]:8201"
        }

        storage "raft" {
          path = "/openbao/data"
        }

        service_registration "kubernetes" {}
  dataStorage:
    enabled: true
    size: 1Gi
    storageClass: local-path
    accessMode: ReadWriteOnce
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Retain
    whenScaled: Retain
  extraVolumes:
    - type: secret
      name: openbao-unseal-key
      defaultMode: 256
  extraContainers:
    - name: openbao-unseal
      image: quay.io/openbao/openbao:2.5.1
      imagePullPolicy: IfNotPresent
      env:
        - name: BAO_ADDR
          value: "http://127.0.0.1:8200"
      command: ["/bin/sh", "-c"]
      args:
        - |
          while true; do
            status=$(bao status -format=json 2>/dev/null || true)
            case "$status" in
              *'"sealed":true'*|*'"sealed": true'*)
                key=$(cat /openbao/unseal/key)
                if [ -n "$key" ]; then
                  bao operator unseal "$key" >/dev/null 2>&1 || true
                fi
                ;;
            esac
            sleep 5
          done
      volumeMounts:
        - name: userconfig-openbao-unseal-key
          mountPath: /openbao/unseal
          readOnly: true
      resources:
        requests:
          cpu: "5m"
          memory: "16Mi"
        limits:
          cpu: "50m"
          memory: "64Mi"
```

### 3.2 Before -> after (exact key paths, chart openbao-0.25.6)

| Key path | Before (measured) | After |
|---|---|---|
| `injector.enabled` | `false` | `false` (unchanged) |
| `server.dev.enabled` | `true` | `false` |
| `server.dev.devRootToken` | `root` | removed (chart default empty) |
| `server.postStart` | script (section 1.2) | removed (chart default `[]`) |
| `server.ha.enabled` | `false` | `true` |
| `server.ha.replicas` | `3` (chart default) | `1` |
| `server.ha.raft.enabled` | `false` | `true` |
| `server.ha.raft.config` | chart raft example | raft storage at `/openbao/data` (above) |
| `server.dataStorage.enabled` | `true` (inert under dev mode) | `true` (now live: volumeClaimTemplate renders only in ha/raft mode) |
| `server.dataStorage.size` | `10Gi` | `1Gi` |
| `server.dataStorage.storageClass` | `null` | `local-path` (explicit, not cluster default) |
| `server.dataStorage.accessMode` | `ReadWriteOnce` | `ReadWriteOnce` (unchanged, pinned) |
| `server.persistentVolumeClaimRetentionPolicy` | `{}` in values (live STS shows Retain/Retain, origin UNVERIFIED) | `{whenDeleted: Retain, whenScaled: Retain}` (pinned) |
| `server.extraVolumes` | `[]` | secret volume `openbao-unseal-key`, `defaultMode: 256` (0400) |
| `server.extraContainers` | `null` | unseal sidecar (above) |
| `server.updateStrategyType` | `OnDelete` | `OnDelete` (chart default; not set in the file) |
| `snapshotAgent.enabled` | `false` | `false` (unchanged; OQ-2 disposition is snapshot-on-script) |

### 3.3 Sidecar sizing rationale (NFR-1)

The sidecar is a sleeping `/bin/sh` loop plus one `bao` invocation per 5s poll
in the same `quay.io/openbao/openbao:2.5.1` image already pulled on the node
(no new image dependency). Limits `cpu: 50m / memory: 64Mi` bound its worst
case; steady-state RSS of the loop is single-digit MiB (spec-time assertion).
Together with Raft overhead at this secret volume (kilobytes of payload), the
NFR-1 delta budget (<100 MiB) is bounded by construction at 64 MiB sidecar
limit plus Raft steady-state; measured at acceptance (section 9, lane D).

### 3.4 Sidecar unseal loop semantics

- `bao status -format=json` exits 2 while sealed but still prints JSON, so the
  `|| true` keeps the loop alive; the case match accepts both
  `"sealed":true` and `"sealed": true` renderings. (Exit-code contract per the
  chart's own probe comment, `server-statefulset.yaml`: 0 unsealed, 1 error,
  2 sealed.)
- Before initialization, and while the Secret still holds the placeholder,
  unseal attempts fail harmlessly (`|| true`) and the loop retries. The
  bootstrap (section 4.1) unseals the server directly on first boot, so the
  migration does not depend on kubelet's secret-volume refresh delay; the
  sidecar covers every subsequent restart, when the Secret already holds the
  real key (spec-time assertion; proven by the FR-4 lane, section 6).
- The volume name `userconfig-openbao-unseal-key` is the chart's generated
  name for `extraVolumes` entry `openbao-unseal-key` (verified,
  `_helpers.tpl` `openbao.volumes`).

---

## 4. Script change points (full listings)

Repo bash conventions honored throughout: `set -euo pipefail` (the teardown
keeps its existing `set -uo pipefail`), `ROOT="$(cd "$(dirname
"${BASH_SOURCE[0]}")/.." && pwd)"`, all variable expansions quoted
(rr-bash-guard blocks unquoted expansion), no emoji (rr-emoji-guard).

### 4.1 NEW `scripts/bootstrap-openbao-persistent.sh` (D-06, FR-3, FR-6)

One-time bootstrap replacing the helm postStart script. Idempotent: an
already-initialized backend recovers key material from custody instead of
re-initializing; all writes are check-before-put.

```bash
#!/usr/bin/env bash
# scripts/bootstrap-openbao-persistent.sh
#
# One-time bootstrap for the Raft-backed OpenBao server (BAO-STORAGE-DES-001
# D-06). Replaces the dev-mode helm postStart script (content preserved
# verbatim below, measured 2026-08-26). Run once per storage lifetime by
# scripts/install-openbao-storage.sh; safe to re-run (idempotent).
#
# Steps: wait for API -> init or recover from custody -> sync the unseal
# Secret -> unseal + verify token -> enable mounts -> auth/policies/roles ->
# thirteen platform secrets -> M2 seed -> root-token handoff to
# ExternalSecrets.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
ESO_NS=${ESO_NS:-external-secrets}
ESO_SECRET=${ESO_SECRET:-openbao-root-token}
UNSEAL_SECRET=${UNSEAL_SECRET:-openbao-unseal-key}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}

info() { printf "\033[1;36m[openbao-bootstrap]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-bootstrap ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

UNSEAL_KEY=""
ROOT_TOKEN=""

# bao invocations that need no auth (init, init -status, status, unseal).
bao_raw() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- bao "$@"
}

# bao invocations authenticated with the current root token.
bao_auth() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" bao "$@"
}

# bao operator init -status: exit 0 initialized, 2 uninitialized; both mean
# the API is answering. Any other exit means the server is still coming up.
# (CLI contract carried from vault-compatible bao; spec-time assertion,
# exercised at Phase 2.)
wait_for_api() {
    local tries=0 rc=0
    while [ "$tries" -lt 60 ]; do
        rc=0
        bao_raw operator init -status >/dev/null 2>&1 || rc=$?
        if [ "$rc" -eq 0 ] || [ "$rc" -eq 2 ]; then
            return 0
        fi
        tries=$((tries + 1))
        sleep 3
    done
    fail "openbao API did not answer within 180s"
}

init_or_recover() {
    if bao_raw operator init -status >/dev/null 2>&1; then
        info "Raft backend already initialized; recovering key material from custody"
        [ -r "$CUSTODY_DIR/unseal-key" ] || fail "initialized backend but $CUSTODY_DIR/unseal-key is missing"
        [ -r "$CUSTODY_DIR/root-token" ] || fail "initialized backend but $CUSTODY_DIR/root-token is missing"
        UNSEAL_KEY=$(cat "$CUSTODY_DIR/unseal-key")
        ROOT_TOKEN=$(cat "$CUSTODY_DIR/root-token")
    else
        info "Initializing Raft backend (Shamir key-shares=1 key-threshold=1)"
        local out
        out=$(bao_raw operator init -key-shares=1 -key-threshold=1) || fail "bao operator init failed"
        UNSEAL_KEY=$(printf '%s\n' "$out" | sed -n 's/^Unseal Key 1: //p')
        ROOT_TOKEN=$(printf '%s\n' "$out" | sed -n 's/^Initial Root Token: //p')
        [ -n "$UNSEAL_KEY" ] || fail "could not parse unseal key from init output"
        [ -n "$ROOT_TOKEN" ] || fail "could not parse root token from init output"
        umask 077
        mkdir -p "$CUSTODY_DIR"
        chmod 700 "$CUSTODY_DIR"
        printf '%s\n' "$UNSEAL_KEY" > "$CUSTODY_DIR/unseal-key"
        printf '%s\n' "$ROOT_TOKEN" > "$CUSTODY_DIR/root-token"
        chmod 600 "$CUSTODY_DIR/unseal-key" "$CUSTODY_DIR/root-token"
        info "custody written: $CUSTODY_DIR (dir 700, files 600)"
    fi
}

sync_unseal_secret() {
    kubectl -n "$OPENBAO_NS" create secret generic "$UNSEAL_SECRET" \
        --from-literal=key="$UNSEAL_KEY" --dry-run=client -o yaml \
        | kubectl apply -f - >/dev/null
    info "secret $UNSEAL_SECRET synced (ns $OPENBAO_NS)"
}

ensure_unsealed() {
    local status
    status=$(bao_raw status -format=json 2>/dev/null || true)
    case "$status" in
        *'"sealed":true'*|*'"sealed": true'*)
            info "server sealed; unsealing (one-time bootstrap action; the sidecar covers later restarts)"
            bao_raw operator unseal "$UNSEAL_KEY" >/dev/null || fail "bao operator unseal failed"
            ;;
    esac
    kubectl -n "$OPENBAO_NS" wait --for=condition=Ready "pod/$OPENBAO_POD" --timeout=120s \
        || fail "pod/$OPENBAO_POD did not become Ready after unseal"
    # Sanity: the custody root token is valid against this backend. A mismatch
    # means custody and cluster disagree; fail loudly rather than re-init.
    bao_auth token lookup >/dev/null 2>&1 \
        || fail "custody root token rejected by the backend (custody/cluster mismatch)"
}

# Fresh Raft has NO mounts (dev mode's pre-mounted secret/ is gone). Enable
# both kv-v2 mounts before any secret write. Idempotent.
# C6 (BAO-STORAGE-SIM-001 D3): plain grep instead of grep -q -- the -q early
# exit SIGPIPEs the framed kubectl-exec producer under load and pipefail then
# misreads a present mount as absent.
ensure_mounts() {
    if ! bao_auth secrets list 2>/dev/null | grep '^kv/' >/dev/null; then
        bao_auth secrets enable -path=kv -version=2 kv >/dev/null
    fi
    if ! bao_auth secrets list 2>/dev/null | grep '^secret/' >/dev/null; then
        bao_auth secrets enable -path=secret -version=2 kv >/dev/null
    fi
}

# Durable configuration, verbatim from the retired postStart (section 1.2),
# minus the dev-mode token export. All writes are idempotent against Raft.
apply_durable_config() {
    bao_auth auth enable kubernetes >/dev/null 2>&1 || true
    # $KUBERNETES_PORT_443_TCP_ADDR expands IN THE POD (single-quoted sh -c).
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" sh -c \
        'bao write auth/kubernetes/config kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"' >/dev/null

    kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" \
        bao policy write openchoreo-secret-reader-policy - >/dev/null <<'POLICY'
path "secret/data/*" { capabilities = ["read"] }
path "secret/metadata/*" { capabilities = ["list", "read"] }
POLICY

    kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" \
        bao policy write openchoreo-secret-writer-policy - >/dev/null <<'POLICY'
path "secret/data/*" { capabilities = ["create", "read", "update", "delete"] }
path "secret/metadata/*" { capabilities = ["create", "read", "update", "delete", "list"] }
POLICY

    bao_auth write auth/kubernetes/role/openchoreo-secret-reader-role \
        bound_service_account_names=default \
        bound_service_account_namespaces="dp*" \
        policies=openchoreo-secret-reader-policy ttl=20m >/dev/null

    bao_auth write auth/kubernetes/role/openchoreo-secret-writer-role \
        bound_service_account_names="*" \
        bound_service_account_namespaces="openbao,openchoreo-workflow-plane" \
        policies=openchoreo-secret-writer-policy ttl=20m >/dev/null
}

put_if_absent() {
    local path="$1" value="$2"
    if ! bao_auth kv get "$path" >/dev/null 2>&1; then
        bao_auth kv put "$path" value="$value" >/dev/null
    fi
}

# The thirteen static platform secrets from the retired postStart, values
# verbatim (section 1.2). Existing values are preserved; rotation is an
# explicit future change, never a bootstrap side effect.
seed_platform_secrets() {
    put_if_absent secret/npm-token "fake-npm-token-for-development"
    put_if_absent secret/docker-username "dev-user"
    put_if_absent secret/docker-password "dev-password"
    put_if_absent secret/github-pat "fake-github-token-for-development"
    put_if_absent secret/username "dev-user"
    put_if_absent secret/password "dev-password"
    put_if_absent secret/backstage-backend-secret "local-dev-backend-secret"
    put_if_absent secret/backstage-client-secret "backstage-portal-secret"
    put_if_absent secret/backstage-jenkins-api-key "placeholder-not-in-use"
    put_if_absent secret/observer-oauth-client-secret "openchoreo-observer-resource-reader-client-secret"
    put_if_absent secret/rca-oauth-client-secret "openchoreo-rca-agent-secret"
    put_if_absent secret/opensearch-username "admin"
    put_if_absent secret/opensearch-password "ThisIsTheOpenSearchPassword1"
}

# D-07: the ClusterSecretStore openbao-kv reads its token from this Secret;
# after bootstrap it must hold the generated root token, not the literal root.
handoff_root_token() {
    kubectl -n "$ESO_NS" create secret generic "$ESO_SECRET" \
        --from-literal=token="$ROOT_TOKEN" --dry-run=client -o yaml \
        | kubectl apply -f - >/dev/null
    info "secret $ESO_SECRET (ns $ESO_NS) now holds the generated root token"
}

main() {
    wait_for_api
    init_or_recover
    sync_unseal_secret
    ensure_unsealed
    ensure_mounts
    apply_durable_config
    seed_platform_secrets
    info "seeding the four M2 kv paths"
    OPENBAO_TOKEN="$ROOT_TOKEN" "$ROOT/scripts/seed-openbao-m2-paths.sh"
    handoff_root_token
    info "bootstrap complete"
}

main "$@"
```

### 4.2 NEW `scripts/install-openbao-storage.sh` (DES-001 section 5; FR-3, NFR-4)

The migration orchestrator. Five script-owned steps; idempotent greenfield
path; serialized heavy operation (never run concurrently with other smokes or
cluster-affecting work).

```bash
#!/usr/bin/env bash
# scripts/install-openbao-storage.sh
#
# G5 migration orchestrator: dev-mode inmem -> Raft on a local-path PVC
# (BAO-STORAGE-DES-001 section 5). Steps:
#   1. FR-4 inverse-proof lane, pre-change (FAIL expected; PASS aborts)
#   2. helm upgrade to scripts/openbao-values.yaml + pod restart
#   3. bootstrap (scripts/bootstrap-openbao-persistent.sh)
#   4. FR-4 inverse-proof lane, post-change (PASS required)
#   5. smoke-openbao.sh + smoke-all.sh (serialized)
# Set OPENBAO_STORAGE_SKIP_FULL_SMOKE=1 when invoked from install-m2.sh
# (task 6 owns the smoke run there).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OPENBAO_NS=${OPENBAO_NS:-openbao}
RELEASE=${RELEASE:-openbao}
CHART_VERSION=0.25.6
VALUES_FILE="$ROOT/scripts/openbao-values.yaml"
UNSEAL_SECRET=${UNSEAL_SECRET:-openbao-unseal-key}

info() { printf "\033[1;36m[openbao-storage]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-storage ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

already_persistent() {
    kubectl -n "$OPENBAO_NS" get pvc data-openbao-0 >/dev/null 2>&1 || return 1
    # C9 (BAO-STORAGE-SIM-001 D5): a retained PVC alone does not mean the Raft
    # template is live -- after a rollback the release is dev-mode again while
    # the PVC persists, and the re-run path then fails against the dev backend
    # ("custody root token rejected"). Skip migration steps 1-2 only when the
    # current StatefulSet carries the unseal sidecar (the raft-template
    # marker); otherwise fall through to the full path, whose upgrade
    # re-attaches the retained PVC and whose bootstrap recovers the persisted
    # Raft state.
    kubectl -n "$OPENBAO_NS" get statefulset "$RELEASE" \
        -o jsonpath='{.spec.template.spec.containers[*].name}' 2>/dev/null \
        | tr ' ' '\n' | grep -qx openbao-unseal
}

step_1_inverse_proof() {
    info "Step 1: FR-4 inverse-proof lane (pre-change; FAIL expected)"
    if "$ROOT/scripts/smoke-openbao.sh" --with-restart; then
        fail "restart lane PASSED against dev-mode storage; the check cannot fail and is unverified (inverse-proof convention). Investigate before proceeding."
    fi
    info "Step 1 recorded: lane FAILED as expected (inmem data loss on pod deletion)"
}

step_2_upgrade() {
    info "Step 2: helm upgrade to Raft values (chart $CHART_VERSION)"
    helm repo add openbao https://openbao.github.io/openbao-helm >/dev/null 2>&1 || true
    # The unseal Secret must exist before the new pod template starts: the
    # sidecar mounts it as a volume, and a missing Secret blocks container
    # creation. Placeholder value; the bootstrap (step 3) writes the real key
    # and unseals directly, so first boot does not depend on secret-volume
    # refresh.
    if ! kubectl -n "$OPENBAO_NS" get secret "$UNSEAL_SECRET" >/dev/null 2>&1; then
        kubectl -n "$OPENBAO_NS" create secret generic "$UNSEAL_SECRET" \
            --from-literal=key=PENDING-BOOTSTRAP >/dev/null
    fi
    # volumeClaimTemplates is immutable on a StatefulSet: the Raft template
    # cannot be patched onto the existing StatefulSet. Orphan-delete it (the
    # pod keeps running); helm upgrade then recreates the StatefulSet with
    # the data volume. (Kubernetes API behavior; verified constraint, see
    # section 11.)
    kubectl -n "$OPENBAO_NS" delete statefulset "$RELEASE" --cascade=orphan --ignore-not-found >/dev/null
    # No --reuse-values: scripts/openbao-values.yaml is the complete set, so
    # dev.devRootToken and postStart return to chart defaults (removed).
    helm upgrade "$RELEASE" openbao/openbao \
        --version "$CHART_VERSION" \
        --namespace "$OPENBAO_NS" \
        --values "$VALUES_FILE"
    # updateStrategyType OnDelete: the running pod still has the old template
    # until it is deleted.
    kubectl -n "$OPENBAO_NS" delete pod openbao-0 --ignore-not-found >/dev/null
    info "waiting for openbao-0 to run with the Raft template"
    # C7 (BAO-STORAGE-SIM-001 D4): a bare kubectl wait errors NotFound when
    # the recreated pod does not exist yet and set -e kills the orchestrator.
    # Poll within the same 180s budget instead.
    local wstart=$SECONDS
    until kubectl -n "$OPENBAO_NS" wait --for=jsonpath='{.status.phase}'=Running pod/openbao-0 --timeout=15s >/dev/null 2>&1; do
        if [ $((SECONDS - wstart)) -ge 180 ]; then
            fail "openbao-0 did not reach Running within 180s; see rollback (BAO-STORAGE-TECH-001 section 7)"
        fi
        sleep 2
    done
}

step_3_bootstrap() {
    info "Step 3: bootstrap (init/custody/unseal/auth/policies/secrets/seed/handoff)"
    "$ROOT/scripts/bootstrap-openbao-persistent.sh"
}

step_4_inverse_proof_pass() {
    info "Step 4: FR-4 inverse-proof lane (post-change; PASS required)"
    "$ROOT/scripts/smoke-openbao.sh" --with-restart \
        || fail "post-change restart lane FAILED; do not claim acceptance (section 6)"
}

step_5_smokes() {
    if [ "${OPENBAO_STORAGE_SKIP_FULL_SMOKE:-0}" = "1" ]; then
        info "Step 5: deferred to the invoking orchestrator (install-m2 task 6)"
        return 0
    fi
    info "Step 5: serialized smoke suites"
    "$ROOT/scripts/smoke-openbao.sh" || fail "smoke-openbao failed after migration"
    "$ROOT/scripts/smoke-all.sh" || fail "smoke-all failed after migration"
}

main() {
    [ -f "$VALUES_FILE" ] || fail "missing $VALUES_FILE"
    if already_persistent; then
        # Greenfield-after-migration / re-run path (DES-001 section 5 ordering
        # note): no dev-mode phase exists, so steps 1-2 are skipped.
        info "PVC data-openbao-0 already present; skipping migration steps 1-2"
        step_3_bootstrap
        "$ROOT/scripts/smoke-openbao.sh" || fail "smoke-openbao failed"
        info "OpenBao persistent storage verified (re-run path)"
        return 0
    fi
    step_1_inverse_proof
    step_2_upgrade
    step_3_bootstrap
    step_4_inverse_proof_pass
    step_5_smokes
    info "OpenBao persistent storage migration complete."
}

main "$@"
```

### 4.3 NEW `scripts/backup-openbao.sh` (DES-001 section 7; OQ-2)

Raft snapshot to host custody, weekly by operator convention and before any
cluster surgery. No CronJob (NFR-1). Streams the snapshot out with `cat` over
kubectl exec (no in-container `tar` dependency that `kubectl cp` would need).

```bash
#!/usr/bin/env bash
# scripts/backup-openbao.sh
#
# Raft snapshot backup (BAO-STORAGE-DES-001 section 7; OQ-2 disposition:
# snapshot-on-script, weekly cadence by operator convention, no CronJob).
# Run weekly and before any cluster surgery. Snapshots land in
# ~/.rational-reserve/backups/openbao/ (host disk; off-host copy is a
# recorded out-of-scope follow-up).
set -euo pipefail

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}
BACKUP_DIR=${OPENBAO_BACKUP_DIR:-"$HOME/.rational-reserve/backups/openbao"}
KEEP=${OPENBAO_BACKUP_KEEP:-8}

info() { printf "\033[1;36m[openbao-backup]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-backup ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

[ -r "$CUSTODY_DIR/root-token" ] || fail "missing $CUSTODY_DIR/root-token (run bootstrap first)"
ROOT_TOKEN=$(cat "$CUSTODY_DIR/root-token")

umask 077
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

TS=$(date -u +%Y%m%d-%H%M%S)
SNAP="$BACKUP_DIR/openbao-$TS.snap"
POD_TMP=/tmp/openbao-snapshot.snap

kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" \
    bao operator raft snapshot save "$POD_TMP" >/dev/null || fail "raft snapshot save failed"
kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- cat "$POD_TMP" > "$SNAP" || fail "snapshot stream failed"
kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- rm -f "$POD_TMP" >/dev/null
[ -s "$SNAP" ] || fail "snapshot $SNAP is empty"
chmod 600 "$SNAP"

# Retention: keep the newest KEEP snapshots.
ls -1t "$BACKUP_DIR"/openbao-*.snap 2>/dev/null | tail -n "+$((KEEP + 1))" | while read -r old; do
    rm -f "$old"
done

info "snapshot written: $SNAP"
```

### 4.4 NEW `scripts/rollback-openbao-storage.sh` (DES-001 section 9; FR-5 preserved)

DES-001 section 9 step 4 says "the rollback script performs this"; this names
and defines that script. Rollback restores dev mode without destroying the
persistent state; a later re-upgrade resumes from Raft with no reseed.

```bash
#!/usr/bin/env bash
# scripts/rollback-openbao-storage.sh
#
# Roll back to the pre-G5 dev-mode release (BAO-STORAGE-DES-001 section 9).
# Preserves the data-openbao-0 PVC, the openbao-unseal-key Secret, the host
# custody files, and all snapshots. Usage:
#   ./scripts/rollback-openbao-storage.sh [revision]   # default 1
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OPENBAO_NS=${OPENBAO_NS:-openbao}
RELEASE=${RELEASE:-openbao}
REVISION=${1:-1}
ESO_NS=${ESO_NS:-external-secrets}
ESO_SECRET=${ESO_SECRET:-openbao-root-token}
# C8 v2 (BAO-STORAGE-SIM-001 D2): wait budget overridable; default 120s kept.
ROLLBACK_WAIT=${OPENBAO_ROLLBACK_WAIT_SECONDS:-120}

info() { printf "\033[1;36m[openbao-rollback]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-rollback ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

info "orphaning the Raft StatefulSet (volumeClaimTemplates is immutable; pod keeps running)"
kubectl -n "$OPENBAO_NS" delete statefulset "$RELEASE" --cascade=orphan --ignore-not-found >/dev/null

info "helm rollback $RELEASE to revision $REVISION (dev mode + postStart)"
helm rollback "$RELEASE" "$REVISION" --namespace "$OPENBAO_NS"

info "restarting the pod onto the dev-mode template (updateStrategy OnDelete)"
kubectl -n "$OPENBAO_NS" delete pod openbao-0 --ignore-not-found >/dev/null
# C8 (BAO-STORAGE-SIM-001 D4, same class as C7): a bare kubectl wait races
# pod recreation (NotFound kills the script under set -e). Poll within the
# same budget instead.
wstart=$SECONDS
until kubectl -n "$OPENBAO_NS" wait --for=condition=Ready pod/openbao-0 --timeout=15s >/dev/null 2>&1; do
    if [ $((SECONDS - wstart)) -ge "$ROLLBACK_WAIT" ]; then
        fail "openbao-0 did not become Ready on the rolled-back template"
    fi
    sleep 2
done

info "restoring literal dev root token for ExternalSecrets"
kubectl -n "$ESO_NS" create secret generic "$ESO_SECRET" \
    --from-literal=token=root --dry-run=client -o yaml \
    | kubectl apply -f - >/dev/null

info "reseeding the fresh inmem backend (the seeder's pre-G5 recovery role)"
OPENBAO_TOKEN=root "$ROOT/scripts/seed-openbao-m2-paths.sh"

info "rollback complete. PVC data-openbao-0, secret openbao-unseal-key,"
info "custody ~/.rational-reserve/openbao/, and snapshots are preserved;"
info "a re-run of install-openbao-storage.sh resumes from persisted Raft state."
```

### 4.5 CHANGE `scripts/seed-openbao-m2-paths.sh` (D-06 token source; FR-3)

One change: the default token source becomes the D-05 custody file;
`OPENBAO_TOKEN` still overrides; the literal `root` remains only as the
pre-migration dev-mode fallback. The header comment's dev-mode wording is
updated to match the new role (one-time bootstrap component, still
idempotent). Exact diff:

```diff
 #!/usr/bin/env bash
-# Seed the OpenBao kv v2 paths M2 relies on. Idempotent.
-#
-# Authenticates as the dev-mode root token (VAULT_TOKEN=root by default;
-# override via OPENBAO_TOKEN env). openbao's dev-mode inmem storage loses
-# mounts and contents whenever the pod restarts, so we (re-)enable the
-# kv-v2 mounts here before writing.
+# Seed the OpenBao kv v2 paths M2 relies on. Idempotent.
+#
+# Post-G5 this is a one-time bootstrap component invoked by
+# scripts/bootstrap-openbao-persistent.sh, not a recovery tool (BAO-STORAGE-
+# DES-001 D-06). Authenticates with the root token from the custody file
+# ~/.rational-reserve/openbao/root-token by default; OPENBAO_TOKEN overrides;
+# the literal "root" fallback only matches the pre-migration dev-mode
+# backend. Mounts are (re-)enabled before writing so the script is also safe
+# against a fresh backend.
 set -euo pipefail
 
 OPENBAO_POD=${OPENBAO_POD:-openbao-0}
 OPENBAO_NS=${OPENBAO_NS:-openbao}
-OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}
+CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}
+if [ -z "${OPENBAO_TOKEN:-}" ] && [ -r "$CUSTODY_DIR/root-token" ]; then
+    OPENBAO_TOKEN=$(cat "$CUSTODY_DIR/root-token")
+fi
+OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}
 GITEA_RUNNER_SECRET_NS=${GITEA_RUNNER_SECRET_NS:-gitea-runners}
 GITEA_RUNNER_SECRET_NAME=${GITEA_RUNNER_SECRET_NAME:-gitea-runner-token}
```

The D3/C6 correction (BAO-STORAGE-SIM-001) also lands in the mount check:

```diff
 ensure_kv_v2_mount() {
     local mount="$1"
 
-    if ! exec_bao secrets list 2>/dev/null | grep -q "^${mount}/"; then
+    # C6 (BAO-STORAGE-SIM-001 D3): grep -q exits on first match; kubectl
+    # exec streams frames, so under node load the producer is SIGPIPEd and
+    # pipefail misreads a present mount as absent. Plain grep reads to EOF.
+    if ! exec_bao secrets list 2>/dev/null | grep "^${mount}/" >/dev/null; then
         exec_bao secrets enable -path="$mount" -version=2 kv >/dev/null
     fi
 }
```

No other line of the seeder changes. Its idempotence, runner-token recovery
fallback, and app-secret rotation flags are untouched.

### 4.6 CHANGE `scripts/smoke-openbao.sh` (D-08; FR-4) -- full replacement

The default invocation keeps today's fast four-key presence check (so
`smoke-m2.sh` stays lightweight). The opt-in `--with-restart` lane is the
FR-4 inverse proof. Flag parsing composes with the shared
`smoke-json.sh` argument handling (`SMOKE_JSON_ARGS`).

```bash
#!/usr/bin/env bash
# scripts/smoke-openbao.sh
#
# Default lane: presence of the four M2 keys (fast; part of smoke-m2).
# --with-restart: FR-4 inverse-proof lane (BAO-STORAGE-DES-001 D-08) --
# records exact values, deletes openbao-0, requires Ready (which proves the
# unseal sidecar acted, because the readiness probe fails while sealed),
# asserts value equality with no reseed. Heavy and serialized: standalone or
# via smoke-all.sh --with-openbao-restart, never concurrent with other
# smokes. Expected FAIL against dev-mode inmem storage; expected PASS
# against the Raft backend.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin openbao

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}

# Post-G5 default: the custody root token. OPENBAO_TOKEN overrides; the
# literal "root" fallback only matches the pre-migration dev-mode backend.
if [ -z "${OPENBAO_TOKEN:-}" ] && [ -r "$CUSTODY_DIR/root-token" ]; then
    OPENBAO_TOKEN=$(cat "$CUSTODY_DIR/root-token")
fi
OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}

WITH_RESTART=0
for arg in ${SMOKE_JSON_ARGS[@]+"${SMOKE_JSON_ARGS[@]}"}; do
    case "$arg" in
        --with-restart) WITH_RESTART=1 ;;
        *) echo "smoke-openbao: unknown argument: $arg" >&2; exit 2 ;;
    esac
done

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$OPENBAO_TOKEN" bao "$@"
}

M2_KEYS="
kv/gitea/runners/token
kv/flux/gitea-deploy-key
secret/apps/hello-m2/dev/example-secret
kv/apps/hello-m2/dev/example-secret
"

check_presence() {
    local rc=0 key
    for key in $M2_KEYS; do
        if exec_bao kv get "$key" >/dev/null; then
            smoke_json_count pass
        else
            smoke_json_count fail
            rc=1
        fi
    done
    if [ "$rc" -eq 0 ]; then echo "PASS"; else echo "FAIL"; fi
    return "$rc"
}

check_restart() {
    local keys=() before=() after="" i=0 key start elapsed
    local restart_wait=${OPENBAO_RESTART_WAIT_SECONDS:-120}
    for key in $M2_KEYS; do
        keys[$i]="$key"
        if ! before[$i]=$(exec_bao kv get "$key" 2>&1); then
            echo "FAIL (pre-restart read of $key)"
            smoke_json_count fail
            return 1
        fi
        i=$((i + 1))
    done

    start=$SECONDS
    kubectl -n "$OPENBAO_NS" delete pod "$OPENBAO_POD" --wait=true --timeout=60s >/dev/null
    # Ready requires the readiness probe (bao status) to pass; the probe fails
    # while sealed, so Ready proves the sidecar unsealed the server with no
    # human step. This lane never invokes any seed script -- it is the only
    # actor, so value equality below is the no-reseed proof.
    # C5 (BAO-STORAGE-SIM-001 D4): tolerate the recreation gap (the pod is
    # briefly NotFound after deletion) by polling within the wait budget
    # instead of dying on the first error under set -e.
    until kubectl -n "$OPENBAO_NS" wait --for=condition=Ready "pod/$OPENBAO_POD" --timeout=15s >/dev/null 2>&1; do
        if [ $((SECONDS - start)) -ge "$restart_wait" ]; then
            echo "FAIL (pod/$OPENBAO_POD not Ready within ${restart_wait}s after deletion)"
            smoke_json_count fail
            return 1
        fi
        sleep 2
    done
    elapsed=$((SECONDS - start))

    i=0
    while [ "$i" -lt "${#keys[@]}" ]; do
        if ! after=$(exec_bao kv get "${keys[$i]}" 2>&1); then
            echo "FAIL (post-restart ${keys[$i]} absent -- no reseed was run)"
            smoke_json_count fail
            return 1
        fi
        if [ "$after" != "${before[$i]}" ]; then
            echo "FAIL (post-restart ${keys[$i]} value changed)"
            smoke_json_count fail
            return 1
        fi
        smoke_json_count pass
        i=$((i + 1))
    done
    echo "PASS (restart lane, ${elapsed}s deletion-to-secrets; NFR-2 bound 120s)"
    if [ "$elapsed" -ge 120 ]; then
        echo "FAIL (recovery ${elapsed}s exceeds the 120s NFR-2 bound)"
        smoke_json_count fail
        return 1
    fi
    return 0
}

if [ "$WITH_RESTART" = "1" ]; then
    check_restart
else
    check_presence
fi
```

### 4.7 CHANGE `scripts/smoke-all.sh` (D-08 wiring) -- exact diff

The FR-4 lane is heavy (deletes a pod), so it is opt-in and always runs last,
after every other suite has completed (serialized-heavy-operations
convention). Default `smoke-all.sh` behavior is unchanged.

```diff
 smoke_json_parse_args "$@"
 smoke_json_begin all
 
+WITH_OPENBAO_RESTART=0
+for arg in ${SMOKE_JSON_ARGS[@]+"${SMOKE_JSON_ARGS[@]}"}; do
+    case "$arg" in
+        --with-openbao-restart) WITH_OPENBAO_RESTART=1 ;;
+        *) echo "smoke-all: unknown argument: $arg" >&2; exit 2 ;;
+    esac
+done
+
 # FR-34: child suites append their own records to the same JSONL file;
```

and immediately before the final `if [ ${#FAILED[@]} -eq 0 ]` block:

```diff
+# G5 FR-4 lane (opt-in, heavy, serialized last): deletes openbao-0 and
+# asserts the four M2 keys survive with no reseed.
+if [ "$WITH_OPENBAO_RESTART" = "1" ]; then
+    echo "=== Running smoke-openbao.sh --with-restart (FR-4 lane) ==="
+    if "${ROOT_DIR}/scripts/smoke-openbao.sh" --with-restart; then
+        echo "=== smoke-openbao.sh --with-restart PASSED ==="
+        smoke_json_count pass
+    else
+        echo "=== smoke-openbao.sh --with-restart FAILED ===" >&2
+        FAILED+=("openbao-restart")
+        smoke_json_count fail
+    fi
+    echo
+fi
+
 if [ ${#FAILED[@]} -eq 0 ]; then
```

### 4.8 CHANGE `scripts/teardown-m2.sh` (D-09; FR-5) -- exact diff

Default behavior preserves the PVC, the unseal Secret, and the host custody
files; `--wipe-secrets` deletes all three with a confirmation prompt on an
interactive terminal. (The teardown keeps its existing `set -uo pipefail`.)

```diff
 #!/usr/bin/env bash
 set -uo pipefail
 
 ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
 
 info() { printf "\033[1;36m[m2-teardown]\033[0m %s\n" "$*"; }
 
+WIPE_SECRETS=0
+for arg in "$@"; do
+    case "$arg" in
+        --wipe-secrets) WIPE_SECRETS=1 ;;
+        *) echo "teardown-m2: unknown argument: $arg" >&2; exit 2 ;;
+    esac
+done
+
 "$ROOT/scripts/remove-tofu-hook-from-settings.sh" 2>/dev/null || true
```

and after the namespace deletion loop, before the final info line:

```diff
 "$ROOT/scripts/delete-m2-gitea-repos.sh" 2>/dev/null || true
 
+# G5 / FR-5 (BAO-STORAGE-DES-001 D-09): OpenBao persistent state survives M2
+# teardown by default -- the data-openbao-0 PVC, the openbao-unseal-key
+# Secret, and the host custody files. --wipe-secrets deletes all three.
+CUSTODY_DIR="$HOME/.rational-reserve/openbao"
+if [ "$WIPE_SECRETS" = "1" ]; then
+    if [ -t 0 ]; then
+        printf "Wipe OpenBao PVC, unseal Secret, and custody files? [y/N] "
+        read -r answer
+        if [ "$answer" != "y" ]; then
+            info "wipe aborted; OpenBao state preserved."
+            exit 0
+        fi
+    fi
+    kubectl -n openbao delete pvc data-openbao-0 --ignore-not-found --timeout=2m || true
+    kubectl -n openbao delete secret openbao-unseal-key --ignore-not-found || true
+    rm -rf "$CUSTODY_DIR"
+    info "OpenBao secrets wiped."
+else
+    info "OpenBao PVC data-openbao-0, secret openbao-unseal-key, and $CUSTODY_DIR preserved (use --wipe-secrets to delete)."
+fi
+
 info "M2 torn down. M1 preserved."
```

### 4.9 CHANGE `scripts/install-m2.sh` (DES-001 section 5 wiring; FR-3, NFR-4) -- exact diff

New task 4.5, wired **after** `task_4_tofu_apply` and before
`task_5_wait_flux`. Deviation from DES-001's "ahead of the existing seed
task"; justification in section 11 (tofu owns Secret `openbao-root-token` and
would revert the generated token on apply, so the bootstrap handoff must run
after the last tofu apply of the same lifecycle run; this placement also makes
every install-m2 rerun self-healing against that drift).

```diff
 task_4_tofu_apply() {
     info "Task 4: tofu apply"
     cd "$ROOT/iac"
     export RR_TOFU_GUARD_BYPASS=1
     tofu init -reconfigure
     tofu apply -auto-approve
 }
 
+task_4_5_openbao_storage() {
+    info "Task 4.5: OpenBao persistent storage (G5)"
+    OPENBAO_STORAGE_SKIP_FULL_SMOKE=1 "$ROOT/scripts/install-openbao-storage.sh"
+}
+
 task_5_wait_flux() {
```

```diff
     task_3_build_score2openchoreo
     task_4_tofu_apply
+    task_4_5_openbao_storage
     task_5_wait_flux
```

Greenfield behavior in this position: task 1 seeds dev mode (idempotent,
harmless), task 4 applies tofu (creates `openbao-root-token` = `root`, valid
in dev mode), task 4.5 migrates to Raft and hands off the generated token
*after* tofu, tasks 5-6 verify. On a rerun after migration, the orchestrator's
re-run path re-applies bootstrap idempotently, including the token handoff.

### 4.10 CHANGE `provenance/PROVENANCE.md` (NFR-3)

Append the following block (custody-artifact record per the provenance-duty
convention; values are never committed):

```markdown
## Host custody artifacts (G5, 2026-08-26)

Non-repo artifacts held under ~/.rational-reserve/ (outside all remotes):
openbao/unseal-key and openbao/root-token (dir 700, files 600) -- Shamir
1-of-1 unseal key and generated root token for the Raft-backed OpenBao
(BAO-STORAGE-DES-001 D-05); backups/openbao/*.snap (mode 600) -- raft
snapshots written by scripts/backup-openbao.sh. These are secrets custody,
not third-party works; recorded here so their existence and location are
auditable.
```

---

## 5. Migration runbook (FR-3, NFR-4)

Single entry point, serialized (no other smoke suite, Backstage boot, or CI
run concurrently -- the 2026-08-24 node-flap lesson). Capture evidence:

```sh
mkdir -p ~/.rational-reserve/logs
LOG=~/.rational-reserve/logs/openbao-migration-$(date -u +%Y%m%d-%H%M%S).log
```

### Step 0 -- preflight (read-only)

```sh
kubectl config current-context          # expect: k3d-openchoreo
helm list -n openbao                    # expect: openbao  openbao-0.25.6  revision 1
helm get values openbao -n openbao      # expect: dev.enabled=true, devRootToken=root, postStart present
kubectl -n openbao get pvc              # expect: No resources found
kubectl -n openbao get pods             # expect: openbao-0 Running
./scripts/smoke-openbao.sh              # expect: PASS (four keys present in dev mode)
```

**Abort criteria:** wrong context; release missing or not `openbao-0.25.6`;
revision not 1 (a prior partial attempt -- investigate with `helm history -n
openbao` first); a PVC already exists (use the orchestrator's re-run path
instead of the full migration); the baseline smoke fails (fix dev-mode state
first so step 1's FAIL is attributable to storage, not to pre-existing
breakage).

### Step 1 -- pre-change inverse proof (expected FAIL)

```sh
./scripts/smoke-openbao.sh --with-restart 2>&1 | tee -a "$LOG"
```

**Expected output (spec-time expectation; the mechanism was verified twice
2026-08-24 per REQ section 4):** the four pre-restart reads succeed; the pod
is deleted and returns Ready (dev mode is never sealed); then:

```text
FAIL (post-restart kv/gitea/runners/token absent -- no reseed was run)
```

exit status 1, and the JSONL record (when `--json` is used) shows
`"failed":1` for suite `openbao`.

**Abort criterion:** if the lane PASSES here, the check cannot fail and is
unverified per the inverse-proof convention -- stop and investigate before
mutating anything. (The orchestrator enforces this itself; running it
manually first is optional because the orchestrator re-runs the lane.)

After this step the dev-mode backend has lost the four keys (the lane deleted
the pod). Reseed immediately if the migration will not proceed now:

```sh
OPENBAO_TOKEN=root ./scripts/seed-openbao-m2-paths.sh   # expect: openbao seeded
```

### Step 2 -- helm upgrade to Raft values

Performed by the orchestrator; the exact internal commands (section 4.2
`step_2_upgrade`) and why:

```sh
helm repo add openbao https://openbao.github.io/openbao-helm
kubectl -n openbao create secret generic openbao-unseal-key --from-literal=key=PENDING-BOOTSTRAP   # if absent
kubectl -n openbao delete statefulset openbao --cascade=orphan
helm upgrade openbao openbao/openbao --version 0.25.6 --namespace openbao --values scripts/openbao-values.yaml
kubectl -n openbao delete pod openbao-0
kubectl -n openbao wait --for=jsonpath='{.status.phase}'=Running pod/openbao-0 --timeout=180s
```

(The orchestrator polls this condition in a retry loop with output suppressed
rather than issuing one bare wait -- v0.2 correction C7; shown one-shot here
for readability.)

**Expected outputs:** `helm upgrade` prints `Release "openbao" has been
upgraded. Happy Helming!` and revision 2; the wait prints `pod/openbao-0
condition met`. The pod is Running but NOT Ready (uninitialized/sealed) --
that is expected at this point. `kubectl -n openbao get pvc` now shows
`data-openbao-0` Bound, `1Gi`, `local-path` (local-path is
WaitForFirstConsumer: it binds once the pod schedules).

**Abort criteria:** helm upgrade fails (e.g. chart fetch) -> nothing has
changed except the orphaned StatefulSet; re-running the orchestrator recreates
it via helm. Pod not Running within 180s -> inspect
`kubectl -n openbao describe pod openbao-0` (the usual suspect would be the
secret volume if step ordering were broken); rollback per section 7.

### Step 3 -- bootstrap

```sh
./scripts/bootstrap-openbao-persistent.sh 2>&1 | tee -a "$LOG"
```

(run by the orchestrator; shown standalone for the runbook)

**Expected output (representative; token values never printed):**

```text
[openbao-bootstrap] Initializing Raft backend (Shamir key-shares=1 key-threshold=1)
[openbao-bootstrap] custody written: /Users/<you>/.rational-reserve/openbao (dir 700, files 600)
[openbao-bootstrap] secret openbao-unseal-key synced (ns openbao)
[openbao-bootstrap] server sealed; unsealing (one-time bootstrap action; the sidecar covers later restarts)
pod/openbao-0 condition met
[openbao-bootstrap] seeding the four M2 kv paths
openbao seeded
[openbao-bootstrap] secret openbao-root-token (ns external-secrets) now holds the generated root token
[openbao-bootstrap] bootstrap complete
```

**Abort criteria:** init parse failure, unseal failure, `custody root token
rejected by the backend`, or seeder failure -> do NOT re-init; the script is
idempotent, so a transient failure is re-runnable. Persistent failure ->
rollback (section 7); custody files and the PVC are preserved either way.

### Step 4 -- post-change inverse proof (expected PASS)

```sh
./scripts/smoke-openbao.sh --with-restart 2>&1 | tee -a "$LOG"
```

**Expected PASS criteria (acceptance criterion 1):** exit 0 and

```text
PASS (restart lane, Ns deletion-to-secrets; NFR-2 bound 120s)
```

with all four key values byte-identical across the deletion and `N < 120`.
Because the readiness probe fails while sealed, reaching Ready also proves
the sidecar performed the unseal with no human step.

**Abort criterion:** any FAIL -> do not claim acceptance; rollback (section
7) or diagnose with `kubectl -n openbao logs openbao-0 -c openbao-unseal`.

### Step 5 -- serialized smokes (acceptance criteria 2 proxy and 3)

```sh
./scripts/smoke-openbao.sh   # expect: PASS
./scripts/smoke-all.sh       # expect: ALL SMOKE SUITES PASSED (...)
```

Acceptance criterion 2 (green after a full Colima stop/start) is verified
separately in lane B (section 9) because it requires a host-level restart.

### Whole-migration single command

Steps 1-5 are owned end-to-end by:

```sh
./scripts/install-openbao-storage.sh 2>&1 | tee -a "$LOG"
```

### Post-migration operational notes

- Authentication everywhere defaults to the custody files; the convenience
  token `root` is dead by design (FR-2).
- **tofu drift note (expected, benign):** Secret `openbao-root-token` is
  tofu-managed (`iac/modules/external-secrets-wiring/main.tf:5-14`) with
  configured value `root`. After the handoff, `tofu plan` shows a
  one-attribute diff on that Secret (`token` generated -> `root`). Any future
  `tofu apply` reverts the live Secret to `root`; the ExternalSecrets' 1h
  refreshInterval then self-heals only after the handoff is re-applied.
  Remediation: re-run `./scripts/install-openbao-storage.sh` (re-run path
  re-applies the handoff idempotently) or full `install-m2.sh` (task 4.5 runs
  after task 4, so the lifecycle self-heals every run). The tofu module stays
  untouched per DES-001 section 9; the auth migration that would remove this
  drift is the recorded follow-up (section 12).
- The chart-created PodDisruptionBudget (`maxUnavailable: 0`, section 3.2
  note) blocks `kubectl drain` of the node while openbao-0 runs; direct pod
  deletion (the FR-4 lane, restarts) is unaffected.

---

## 6. FR-4 inverse-proof lane (FR-4; D-08)

**Lane:** `scripts/smoke-openbao.sh --with-restart` (section 4.6), also
reachable as the final serialized lane of `smoke-all.sh
--with-openbao-restart` (section 4.7).

**Pre-change expected FAIL (capture at migration step 1):** exit 1 with
`FAIL (post-restart kv/gitea/runners/token absent -- no reseed was run)`.
Prior evidence this mechanism fails pre-change: REQ section 4 records two
verifications on 2026-08-24; step 1 re-captures it on the migration day and
writes it to the migration log. If the lane passes pre-change, the gate is
unverified and the migration aborts (the orchestrator enforces this).

**Post-change PASS criteria (acceptance criterion 1):** exit 0; all four M2
keys re-read with byte-identical values after pod deletion; pod Ready
attained without any human unseal step; elapsed deletion-to-secrets time
reported and < 120s (NFR-2 measurement, lane D records the number).

**No-reseed assertion:** the lane is the only actor in its window and never
invokes the seeder; value equality across a deletion is therefore the proof
that persistence, not reseeding, preserved the keys.

---

## 7. Rollback procedure (NFR-4; D-09 guarantees)

```sh
./scripts/rollback-openbao-storage.sh           # rolls back to revision 1
./scripts/rollback-openbao-storage.sh <rev>     # explicit revision if history differs
```

What it does (section 4.4), and why each step exists:

1. `kubectl -n openbao delete statefulset openbao --cascade=orphan` --
   volumeClaimTemplates is immutable, so the dev-mode template cannot be
   patched over the Raft one either; orphan-delete lets helm recreate the
   StatefulSet. The pod keeps running throughout.
2. `helm rollback openbao 1 -n openbao` -- restores revision 1 values (dev
   mode, postStart, no volumes). Expected output: `Rollback was a success!
   ... happy helming!`.
3. `kubectl delete pod openbao-0` + wait Ready -- OnDelete strategy requires
   the explicit deletion; dev mode starts unsealed with token `root`.
4. Restore Secret `openbao-root-token` (ns external-secrets) to literal
   `root` -- the generated token does not exist in the fresh inmem backend;
   this also reconverges the tofu-managed value (no residual drift).
5. `OPENBAO_TOKEN=root ./scripts/seed-openbao-m2-paths.sh` -- the seeder
   resumes its pre-G5 recovery role against the fresh inmem backend.
   Expected: `openbao seeded`.

**State preservation through rollback:** `data-openbao-0` PVC (Retain policy,
now pinned in values), Secret `openbao-unseal-key`, custody files
`~/.rational-reserve/openbao/`, and snapshots are all preserved. Re-running
`./scripts/install-openbao-storage.sh` after a rollback resumes from the
persisted Raft state: the orchestrator sees the retained PVC but also sees
the dev-mode StatefulSet (no unseal sidecar), so it falls through to the full
migration -- the upgrade re-attaches the PVC and the bootstrap recovers the
persisted Raft state from custody, needing no reseed (v0.2 correction C9).

**Script rollback:** ordinary git revert of the section-4 change points. No
tofu state is touched by G5 (the external-secrets-wiring module is
untouched).

---

## 8. Backup, restore, and disaster recovery (OQ-2)

**Backup (weekly by operator convention, and before any cluster surgery):**

```sh
./scripts/backup-openbao.sh
# expect: [openbao-backup] snapshot written: /Users/<you>/.rational-reserve/backups/openbao/openbao-YYYYMMDD-HHMMSS.snap
```

Retention: newest 8 snapshots kept (configurable via `OPENBAO_BACKUP_KEEP`).

**Restore from snapshot** (host/cluster loss with a surviving snapshot;
spec-time procedure, exercised at Phase 2 only if a DR drill is scheduled):

```sh
# 1. Fresh, empty backend:
kubectl -n openbao delete pvc data-openbao-0 --ignore-not-found   # ONLY in DR; never in routine ops
./scripts/install-openbao-storage.sh                              # re-init + bootstrap on the empty PVC
# 2. Restore the snapshot over the fresh Raft:
SNAP=~/.rational-reserve/backups/openbao/<file>.snap
kubectl -n openbao exec -i openbao-0 -- sh -c 'cat > /tmp/restore.snap' < "$SNAP"
kubectl -n openbao exec openbao-0 -- env VAULT_TOKEN="$(cat ~/.rational-reserve/openbao/root-token)" \
    bao operator raft snapshot restore /tmp/restore.snap
kubectl -n openbao exec openbao-0 -- rm -f /tmp/restore.snap
# 3. The restored snapshot predates the current root token only if the token
#    was rotated after the snapshot; otherwise custody still matches. Verify:
./scripts/smoke-openbao.sh
```

**DR without a snapshot (the honest common case, DES-001 section 7):** rerun
the migration on a fresh PVC; the bootstrap re-derives everything -- platform
placeholders are static, the Flux deploy key comes from
`~/.rational-reserve/m1-gitea-admin-password`, the runner token from the
surviving Kubernetes Secret `gitea-runners/gitea-runner-token` or Gitea, and
the demo app secret rotates. Worst-case loss is secrets written outside the
bootstrap inventory; none are known (spec-time assertion, inventory in
section 1.1).

**Honest limit (carried):** local-path volumes and `~/.rational-reserve/`
snapshots share the host disk's fate; off-host copies are out of G5 scope.

---

## 9. Test and acceptance plan

All lanes serialized. Static checks run at implementation time (Phase 1);
live lanes run at Phase 2.

**Lane 0 -- static (Phase 1, no cluster mutation):**

```sh
bash -n scripts/bootstrap-openbao-persistent.sh scripts/install-openbao-storage.sh \
        scripts/backup-openbao.sh scripts/rollback-openbao-storage.sh \
        scripts/smoke-openbao.sh scripts/smoke-all.sh scripts/teardown-m2.sh scripts/install-m2.sh
# expect: no output, exit 0
helm template openbao openbao/openbao --version 0.25.6 -n openbao -f scripts/openbao-values.yaml \
  | grep -c 'name: openbao-unseal'        # expect: 1 (sidecar rendered)
helm template openbao openbao/openbao --version 0.25.6 -n openbao -f scripts/openbao-values.yaml \
  | grep 'storageClassName: local-path'   # expect: match in the volumeClaimTemplate
helm template openbao openbao/openbao --version 0.25.6 -n openbao -f scripts/openbao-values.yaml \
  | grep -c 'postStart'                   # expect: 0 (postStart gone)
```

**Lane A -- FR-4 inverse proof:** migration steps 1 and 4 (section 5/6):
expected FAIL then expected PASS, both recorded in the migration log.

**Lane B -- cold-restart acceptance (REQ acceptance criterion 2):**

```sh
colima stop && colima start             # host-level restart
kubectl -n openbao wait --for=condition=Ready pod/openbao-0 --timeout=180s
./scripts/smoke-openbao.sh              # expect: PASS with no manual steps
```

(The scripts poll this same condition in a retry loop rather than one bare
wait -- v0.2 corrections C5/C8; the one-shot form is fine in this manual
lane.)

**Lane C -- full smoke (REQ acceptance criterion 3):** `./scripts/smoke-all.sh`
green, and once with `--with-openbao-restart` green (final serialized lane).

**Lane D -- NFR measurements:**

```sh
# NFR-1: memory delta bounded <100 MiB (measure before migration and after):
kubectl top pod openbao-0 -n openbao --containers
# NFR-2: the FR-4 lane's reported elapsed seconds < 120 (recorded in lane A post-change).
# NFR-3: custody hygiene:
stat -f '%Lp %N' ~/.rational-reserve/openbao ~/.rational-reserve/openbao/*
# expect: 700 dir, 600 files
git status --porcelain ~/.rational-reserve 2>/dev/null; git check-ignore -v ~/.rational-reserve 2>/dev/null
# expect: outside all repos/remotes (path is in $HOME, not under any remote)
```

**Lane E -- teardown inverse proof (FR-5), static form:** full teardown
execution destroys M2 and is not part of G5 acceptance; the guarantee is
verified by code inspection plus a targeted static check:

```sh
grep -q 'data-openbao-0' scripts/teardown-m2.sh \
  && grep -q -- '--wipe-secrets' scripts/teardown-m2.sh \
  && grep -q 'preserved' scripts/teardown-m2.sh   # expect: exit 0
# negative form: without --wipe-secrets the script contains no delete path
# that reaches the PVC (the deletes sit inside the WIPE_SECRETS branch).
```

The next *routine* `teardown-m2.sh` run (whenever M2 is legitimately torn
down) is the live confirmation that `data-openbao-0` survives; record it in
SESSION_HANDOFF when it happens.

---

## 10. Acceptance checklist (traceability)

| Requirement / design element | Spec section | Verification lane |
|---|---|---|
| FR-1 persistent Raft on local-path PVC | s3 (values), s4.2 (upgrade) | s5 step 2 (`kubectl get pvc`: `data-openbao-0` Bound 1Gi local-path); lane A post |
| FR-2 seal/unseal strategy; keys never in git | s3.4 (sidecar), s4.1 (custody, 600/700) | lane A post (Ready-without-human proof); lane D `stat`; custody outside remotes |
| FR-3 install-script bootstrap; seeder demoted | s4.1, s4.5, s4.2, s4.9 | s5 step 3 expected output; lane 0 `bash -n`; greenfield path in s4.2 |
| FR-4 inverse-proof lane | s4.6, s4.7, s6 | lane A (FAIL then PASS, recorded) |
| FR-5 teardown preserves PVC | s4.8 | lane E (static) + next routine teardown (recorded in SESSION_HANDOFF) |
| FR-6 runner-token and app-secret paths covered | s1.1 inventory, s4.1 (seed + handoff) | lane A covers all four keys incl. `kv/gitea/runners/token` and `secret/apps/...`; lane C smoke-all |
| NFR-1 no new standing workload; <100 MiB delta | s3.3, s3 (snapshotAgent stays off) | lane D `kubectl top`; sidecar is in-pod (no new pod/workload) |
| NFR-2 recovery < 2 min | s4.6 (elapsed assertion), s10 lane D | lane A post-change elapsed report |
| NFR-3 custody under ~/.rational-reserve mode 600; provenance | s4.1, s4.10 | lane D `stat`; PROVENANCE.md block present |
| NFR-4 lifecycle via scripts; no ad-hoc steady state | s4.2, s4.9, s5 | all cluster mutations flow through install-openbao-storage.sh / install-m2.sh; rollback scripted (s4.4) |
| D-01 Raft via ha.raft | s3.1 | lane 0 helm template + s5 step 2 |
| D-02 1Gi local-path explicit StorageClass | s3.1 | lane 0 grep + `kubectl get pvc` |
| D-03 single replica | s3.1 (`replicas: 1`) | `kubectl -n openbao get pods` shows exactly openbao-0 |
| D-04 sidecar auto-unseal | s3.1, s3.4 | lane A post (Ready proof), logs `kubectl logs openbao-0 -c openbao-unseal` |
| D-05 dual custody host + Secret | s4.1 (`sync_unseal_secret`, custody write) | lane D; `kubectl -n openbao get secret openbao-unseal-key` |
| D-06 bootstrap replaces postStart | s4.1 (verbatim carry-forward), s3.2 (postStart removed) | lane 0 `postStart` count 0; s5 step 3 output |
| D-07 root-token handoff | s4.1 (`handoff_root_token`), s5 drift note | `kubectl -n external-secrets get secret openbao-root-token -o jsonpath='{.data.token}' \| base64 -d` != `root`; lane C (ESO-backed smokes green) |
| D-08 `--with-restart` lane + smoke-all wiring | s4.6, s4.7 | lane A; lane C with flag |
| D-09 teardown default-preserve + `--wipe-secrets` | s4.8 | lane E |
| OQ-1 (auto-unseal + custody) | s3.4, s4.1 | lane A post |
| OQ-2 (weekly snapshot, no CronJob) | s4.3, s8 | `./scripts/backup-openbao.sh` expected output; snapshot file mode 600 |
| OQ-3 (no CI dependency) | s2 (resolved) | residual check in s2 at Phase 2 |
| REQ acceptance 1 (lane passes) | s5 step 4, s6 | lane A |
| REQ acceptance 2 (cold restart green) | s9 lane B | lane B |
| REQ acceptance 3 (smoke-all green) | s5 step 5, s9 lane C | lane C |
| REQ acceptance 4 (triad approved before mutation) | this document | process: Phase 2 starts only after approval |

---

## 11. Deviations and clarifications vs BAO-STORAGE-DES-001

Each item below deviates from or tightens the approved design; all preserve
its intent and are justified by measured evidence.

1. **Platform-secret count: thirteen, not eleven.** The measured postStart
   (section 1.2) writes 13 kv keys; DES-001 said eleven. The bootstrap
   preserves all thirteen verbatim (section 4.1). "Preserve existing" is the
   binding requirement.
2. **install-m2.sh wiring position: after `task_4_tofu_apply`, not ahead of
   the seed task.** DES-001 section 5 said "ahead of the existing seed task".
   tofu owns Secret `openbao-root-token`
   (`iac/modules/external-secrets-wiring/main.tf:5-14`) and any `tofu apply`
   reverts the generated token to `root`. Placing the bootstrap handoff after
   the last tofu apply of the run makes every install-m2 run self-healing
   instead of self-breaking. The migration orchestrator's internal five-step
   order is unchanged; greenfield seeds twice (dev mode at task 1, Raft at
   task 4.5), both idempotent.
3. **StatefulSet orphan-delete added to upgrade and rollback.**
   `volumeClaimTemplates` is immutable on a StatefulSet, so helm cannot patch
   the dev-mode StatefulSet into the Raft one (or back). Both
   `install-openbao-storage.sh` and `rollback-openbao-storage.sh` therefore
   `kubectl delete statefulset openbao --cascade=orphan` first (pod keeps
   running; helm recreates the StatefulSet; OnDelete then swaps the pod).
   DES-001 step 2 implied a plain `helm upgrade`; this is the mechanics the
   API actually requires.
4. **Placeholder-first for Secret `openbao-unseal-key`.** The sidecar mounts
   the Secret as a volume; a missing Secret blocks container creation, but
   the real key only exists after init (which needs the pod running). The
   orchestrator creates the Secret with `PENDING-BOOTSTRAP` before the
   upgrade; bootstrap writes the real key and unseals directly, so first boot
   never depends on kubelet's secret-volume refresh delay.
5. **Bootstrap unseals directly on first boot.** D-04 assigns unsealing to
   the sidecar; the bootstrap performing the *initial* unseal (a one-time
   operator action) removes the refresh-delay dependency from the migration
   critical path and keeps the sidecar's role exactly as designed for every
   subsequent restart -- which is what the FR-4 lane proves.
6. **Mount enablement added to bootstrap.** Fresh Raft has no mounts (dev
   mode's pre-mounted `secret/` disappears); the bootstrap enables `kv/` and
   `secret/` before any secret write. Implied by D-06 ("keeping consumers
   whole") but not listed there.
7. **Rollback script named:** `scripts/rollback-openbao-storage.sh`
   implements DES-001 section 9's "the rollback script performs this".
8. **tofu drift hazard documented** (section 5, post-migration notes):
   DES-001's "no tofu re-apply is needed" is true for effecting the change;
   the corollary (a later apply reverts the Secret) is now explicit with the
   exact remediation. The module stays untouched as designed.
9. **Retention policy pinned in values.** Live STS shows Retain/Retain but
   computed values show `{}` (origin UNVERIFIED); the new values file sets
   `persistentVolumeClaimRetentionPolicy` explicitly so the guarantee no
   longer depends on unexplained state.
10. **Snapshot streaming via `cat`, not `kubectl cp`** (no in-container `tar`
    dependency), and retention fixed at newest 8 (`OPENBAO_BACKUP_KEEP`).

No deviation changes a design decision; all are implementation mechanics or
count/wording corrections against measured state.

### Simulation corrections folded (v0.2)

Folded from the scratch-namespace rehearsal (BAO-STORAGE-SIM-001 section 2);
each corrects a defect the simulation exposed in the v0.1 listings:

- **C5 (D4, `smoke-openbao.sh` restart lane):** bare `kubectl wait` after
  the `--wait=true` deletion raced pod recreation (NotFound under `set -e`);
  now a poll loop bounded by `OPENBAO_RESTART_WAIT_SECONDS` (default 120s).
- **C6 (D3, bootstrap `ensure_mounts` and seeder `ensure_kv_v2_mount`):**
  `grep -q` exits on first match and SIGPIPEs the framed kubectl-exec
  producer under load, so `pipefail` misread a present mount as absent;
  plain `grep` reads to EOF.
- **C7 (D4, `install-openbao-storage.sh` step 2):** same wait race on the
  recreated Raft pod; poll loop within the same 180s budget.
- **C8 (D4) + C8 v2 (D2, `rollback-openbao-storage.sh`):** same wait race on
  the rolled-back pod; poll loop, with the budget env-overridable via
  `OPENBAO_ROLLBACK_WAIT_SECONDS` (default 120s kept).
- **C9 (D5, `install-openbao-storage.sh` re-run gate):** a retained PVC
  alone did not prove the Raft template live -- post-rollback the release is
  dev-mode while the PVC persists, and the re-run failed with "custody root
  token rejected"; steps 1-2 now skip only when the StatefulSet carries the
  unseal sidecar, otherwise the full path re-attaches the PVC and the
  bootstrap recovers the persisted Raft state.

---

## 12. Honest limits and out-of-scope follow-ups (carried from DES-001)

- Auto-unseal with the key in a cluster Secret and on the host protects
  persistence semantics, not against a cluster-admin or host-level attacker
  (D-04). KMS/transit seals are the documented upgrade path and require
  infrastructure this environment deliberately does not have.
- local-path PVC and host snapshots share the host disk's fate; node loss is
  data loss without a snapshot; off-host copies are out of G5 scope.
- ExternalSecrets remains on root-token auth; migrating it to `kubernetes`
  auth (anticipated by the module's own comment at
  `iac/modules/external-secrets-wiring/main.tf:1-4`) is a later hardening
  change and removes the section-5 drift note.
- `snapshotAgent` CronJob stays disabled (NFR-1, OQ-2 disposition).
- The chart-created PDB (`maxUnavailable: 0` at replicas 1) blocks node
  drains; documented, left at default.

**End of Technical Specification**
