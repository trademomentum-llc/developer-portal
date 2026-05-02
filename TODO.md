# TODO

> Action list ordered by priority and dependency.

**Snapshot date:** 2026-05-02

---

## M2 IaC + CD Loop -- where we actually are

M2 is **architecturally validated end-to-end** through Pod creation, but **score2openchoreo's output model is wrong for the real OpenChoreo CRDs** and the **k3d cluster does not trust the in-cluster local-registry** for image pulls. Both are scoped, well-understood follow-ups (see m2-renderer-rewrite and m2i-7 below).

What was proved this session (2026-05-02): a push to `openchoreo/hello-m2` triggers CI, CI builds + pushes image, CI renders + commits the Component to `platform-config/environments/dev/`, **Flux pulls platform-config**, **Flux applies the (manually-corrected) Component+Workload+Project triplet**, **OpenChoreo creates a ComponentRelease, then a Deployment, then a Pod** in an auto-named per-environment data-plane namespace (`dp-default-default-development-<hash>`). The Pod cannot pull its image (k3d/containerd registry trust gap), but everything upstream of that works.

### Outstanding M2 work

| Task | Status | Notes |
|------|--------|-------|
| T21 install-m2.sh end-to-end | DONE 2026-05-02 | Cluster healthy with all M2 namespaces; tofu apply ran successfully; m2i-1..m2i-6 closed |
| T22 first pipeline run on hello-m2 | PARTIAL 2026-05-02 | Run #17 (sha dc407cc) completed CI in ~88s; chain validated through Pod creation; pod blocked on image pull (m2i-7) |
| Score2openchoreo renderer rewrite | OPEN -- next session | See m2-renderer-rewrite below; ~2-4 hours of focused work |
| Push 42b2231 to gitea-com | DONE 2026-05-02 | Pushed via one-shot ephemeral PAT; PAT then revoked by operator |
| Push to local Gitea origin | DEFERRED | origin URL is stale (`localhost:3002`); update to 3333 and push when next convenient |

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
| **m2i-7** | **k3d/containerd does not trust in-cluster local-registry** | **NEW 2026-05-02** -- pods get ImagePullBackOff: containerd resolves `registry.local-registry.svc.cluster.local` via host DNS (not cluster DNS) and defaults to HTTPS on a HTTP-only registry. Fix: add `/etc/rancher/k3s/registries.yaml` mirror entry on k3d node OR expose registry via NodePort/hostPort. Affects install-m1.sh / cluster bootstrap, not M2 codebase |

---

## m2-renderer-rewrite -- the score2openchoreo redesign

**Why:** the current `tools/score2openchoreo/convert.go` emits a single Component CRD with `spec.workloadTemplate`, `spec.environment`, and `spec.owner.project`. The actual `openchoreo.dev/v1alpha1/Component` schema rejects all three -- it expects `spec.componentType` (a ref to a ClusterComponentType like `deployment/service`), `spec.owner.projectName`, and stores the workload definition in a separate `Workload` CRD. The renderer was written against an outdated conception of Component.

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

**Required work (Path B, next session):**

| ID | Item |
|---|---|
| score-6a | Replace types.go with `Component`, `Workload`, optional `Project`, optional `SecretReference` shapes matching openchoreo.dev/v1alpha1 |
| score-6b | Convert function returns a list of resources (multi-document YAML), not a single Component |
| score-6c | Decide componentType inference: heuristic from Score (e.g., `service.ports` -> `deployment/service`, no ports -> `worker`) OR a `pipeline.m2/component-type` annotation in score.yaml |
| score-6d | Namespace strategy: emit Components in same namespace as Project (currently `default` per cluster install). Drop `--namespace openchoreo-data-plane` flag or repurpose it |
| score-6e | Rewrite golden fixtures (minimal, with-secret) to multi-document YAML |
| score-6f | Update convert_test, schema_test, main_test for new shape |
| score-6g | Document the new conversion conventions in README; update technical-specification.md |
| score-6h | Update CI workflow (hello-m2/.gitea/workflows/ci.yaml) if `--namespace` flag drops |

---

## M2 done backlog (chronological closeout this session)

| Item | Commit | Date |
|---|---|---|
| Track CLAUDE.md, document Gitea port migration in plan | a99d97e | 2026-05-02 |
| **Complete CRD-group migration: score2openchoreo + gatekeeper + flux platform-config watch** | **42b2231** | **2026-05-02** |

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

origin (local Gitea) URL still points at port 3002; Gitea is now on 3333. Update with `git remote set-url origin http://localhost:3333/openchoreo/developer-portal.git` or similar before next push to local origin.

---

## M3-M7 (unchanged)

| Milestone | Scope | Status |
|---|---|---|
| M3 | OpenTelemetry + SigNoz + Infracost post-deploy dashboards | deferred |
| M4 | OpenCost + Cilium + Envoy Gateway | deferred |
| M5 | RabbitMQ or Kafka + OpenResty front-door | deferred |
| M6 | OPA/Gatekeeper runtime policies + MISP + TheHive + Cortex + Velociraptor + Cloud Custodian | deferred |
| M7 | MCP plugin surfacing Backstage + RR to OpenChoreo + per-agent Gitea tokens | deferred |
