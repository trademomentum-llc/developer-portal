# TODO

> Action list ordered by priority and dependency.

**Snapshot date:** 2026-05-17

---

## M2 IaC + CD Loop -- where we actually are

M2 is **architecturally validated end-to-end** through Pod creation, and the two major follow-ups from 2026-05-02 are now implemented locally:

- `score2openchoreo` now renders schema-valid OpenChoreo `Component` + `Workload` multi-document YAML.
- k3d/containerd registry trust is configured through a NodePort-backed local registry mirror.

Remaining closeout is a fresh `hello-m2` CI run so the new renderer output is committed by automation and the expected image tag is pushed into the local registry. The existing live pod still uses the old `dc407cc` tag and now fails with `not found`, which means registry resolution is working but that stale image tag is absent.

What was proved on 2026-05-02: a push to `openchoreo/hello-m2` triggers CI, CI builds + pushes image, CI renders + commits the Component to `platform-config/environments/dev/`, **Flux pulls platform-config**, **Flux applies the (manually-corrected) Component+Workload+Project triplet**, **OpenChoreo creates a ComponentRelease, then a Deployment, then a Pod** in an auto-named per-environment data-plane namespace (`dp-default-default-development-<hash>`). On 2026-05-17, the renderer output and registry-trust configuration were fixed; the next proof point is a fresh CI run.

### Outstanding M2 work

| Task | Status | Notes |
|------|--------|-------|
| T21 install-m2.sh end-to-end | DONE 2026-05-02 | Cluster healthy with all M2 namespaces; tofu apply ran successfully; m2i-1..m2i-6 closed |
| T22 first pipeline run on hello-m2 | PARTIAL 2026-05-17 | Chain validated through Pod creation. Needs fresh CI run after renderer rewrite + registry trust so the image tag exists and the pod can run |
| Score2openchoreo renderer rewrite | DONE 2026-05-17 | Emits Component + SecretReference + Workload multi-doc YAML; Go tests, score smoke, and live server-side dry-run passed |
| Push to gitea-com | BLOCKED 2026-05-21 | `git push gitea-com main` could not connect to `gitea.com:443` from this environment, even with network escalation |
| Push to local Gitea origin | DONE 2026-05-21 | Created `openchoreo/developer-portal` in local Gitea and pushed `main` through the localhost:3333 port-forward |

---

## M2 install blockers (status as of 2026-05-02)

| ID | Item | Status |
|---|---|---|
| m2i-1 | OpenChoreo CRD group drift -- `core.choreo.dev` -> `openchoreo.dev` | **DONE 2026-05-02 commit 42b2231** -- score2openchoreo, gatekeeper constraints (both `policies/` and the `seed-repos/` mirror), and the technical-spec doc all migrated. Earlier commit 692f200 had only fixed the Environment module |
| m2i-2 | k3d cluster CPU exhausted | **DONE** -- archive memory: colima upgrade (6 CPU / 10 GB) addressed |
| m2i-3 | Stale `tofu-state` ns from 2026-04-21 | **DONE** -- imported during the successful tofu apply |
| m2i-4 | ExternalSecret cache misses under cluster pressure | **DONE** (resolved with m2i-2) |
| m2i-5 | Deprecated `k3d-m1-substrate` cluster | **DONE** -- archive memory: torn down |
| m2i-6 | OpenBao dev-mode `inmem` storage loses kv on restart | **OPEN, low-priority** -- production-readiness item, not M2 closeout |
| **m2i-7** | **k3d/containerd does not trust in-cluster local-registry** | **DONE 2026-05-17** -- `scripts/install-m1.sh` writes `/etc/rancher/k3s/registries.yaml`; local-registry Service is NodePort `30082`; k3s mirror maps `registry.local-registry.svc.cluster.local:5000` to `http://127.0.0.1:30082`. Current pod pull error is now `not found` for old tag `dc407cc`, not DNS/HTTPS trust |

---

## m2-renderer-rewrite -- the score2openchoreo redesign

**Status:** DONE 2026-05-17.

**Why it mattered:** the old `tools/score2openchoreo/convert.go` emitted a single Component CRD with `spec.workloadTemplate`, `spec.environment`, and `spec.owner.project`. The actual `openchoreo.dev/v1alpha1/Component` schema rejects all three -- it expects `spec.componentType` (a ref to a ClusterComponentType like `deployment/service`), `spec.owner.projectName`, and stores the workload definition in a separate `Workload` CRD. The renderer was written against an outdated conception of Component.

**Reference shape** (validated by hand this session against the cluster's CRD schema):

```yaml
apiVersion: openchoreo.dev/v1alpha1
kind: Component
metadata: {name, namespace}
spec:
  owner: {projectName: <project>}
  autoDeploy: true
  componentType: {kind: ClusterComponentType, name: deployment/<service|web-application|worker|scheduled-task>}
---
apiVersion: openchoreo.dev/v1alpha1
kind: Workload
metadata: {name: <component>-workload, namespace}
spec:
  owner: {componentName: <component>, projectName: <project>}
  endpoints: {http: {type: HTTP, port: <port>, visibility: [external]}}
  container:
    image: <oci-ref>
    env: [{key, value | valueFrom.secretKeyRef}]
```

**Completion evidence:**

- `go test ./...` passed in `tools/score2openchoreo`.
- `go build -o bin/score2openchoreo .` passed.
- `./scripts/smoke-score.sh` passed.
- `kubectl --context k3d-openchoreo apply --dry-run=server -f /private/tmp/hello-m2-rendered.yaml` accepted the generated `Component` and `Workload`.

**Completed work:**

| ID | Item |
|---|---|
| score-6a | DONE -- types now include Component, Workload, optional Project, and optional SecretReference shapes matching openchoreo.dev/v1alpha1 |
| score-6b | DONE -- Convert returns `[]OpenChoreoResource` and CLI writes multi-document YAML |
| score-6c | DONE -- heuristic: Score service ports -> `deployment/service`, otherwise `deployment/worker`; `pipeline.m2/component-type` annotation overrides |
| score-6d | DONE -- default namespace and project are `default`, matching the local cluster Project and DeploymentPipeline |
| score-6e | DONE -- golden fixtures are multi-document YAML |
| score-6f | DONE -- convert and CLI/golden tests updated for new shape |
| score-6g | DONE -- README and M2 specs updated in the M2 closeout commit |
| score-6h | DONE -- hello-m2 CI render step no longer passes stale namespace/project flags |

---

## M2 done backlog (chronological closeout this session)

| Item | Commit | Date |
|---|---|---|
| Track CLAUDE.md, document Gitea port migration in plan | a99d97e | 2026-05-02 |
| **Complete CRD-group migration: score2openchoreo + gatekeeper + flux platform-config watch** | **42b2231** | **2026-05-02** |
| Implement score2openchoreo Component+Workload renderer rewrite | M2 closeout commit | 2026-05-17 |
| Implement k3d/containerd local-registry trust via NodePort mirror | M2 closeout commit | 2026-05-17 |

---

## Post-M2 tech debt -- closed before this session

(Carried forward for reference; all DONE before 2026-04-30.)

### Guards (rr-policy-guards)

guard-1 through guard-6 -- DONE 2026-04-23. See git log for commit-level detail.

### score2openchoreo

score-1 through score-5 -- DONE 2026-04-23. (Note: score-6 above is a NEW item from this session, distinct from those.)

### Gatekeeper / Install

gk-1, inst-1, inst-2 -- DONE.

---

## Push / remote resolution -- update

The 2026-04-21 entry suggested gitea.com URL had embedded credentials. As of 2026-05-02:
- `.git/config` has clean URLs (no embedded creds)
- gitea-com auth uses ephemeral PATs minted per-push and immediately revoked
- osxkeychain caches credentials but expires when PATs are revoked

origin (local Gitea) now points at `http://localhost:3333/openchoreo/developer-portal.git` and has received `main`. gitea-com push is blocked by external network reachability from this environment.

---

## M3-M7 (unchanged)

| Milestone | Scope | Status |
|---|---|---|
| M3 | OpenTelemetry + SigNoz + Infracost post-deploy dashboards | deferred |
| M4 | OpenCost + Cilium + Envoy Gateway | deferred |
| M5 | RabbitMQ or Kafka + OpenResty front-door | deferred |
| M6 | OPA/Gatekeeper runtime policies + MISP + TheHive + Cortex + Velociraptor + Cloud Custodian | deferred |
| M7 | MCP plugin surfacing Backstage + RR to OpenChoreo + per-agent Gitea tokens | deferred |
