# SESSION HANDOFF

> Read this file FIRST in the next session. It tells you where we stopped,
> what is now committed that was not before, what is still outstanding, and
> exactly what to do first.

**Last updated:** 2026-05-23
**Reason for handoff:** M2 T22 is complete locally. Backstage catalog validation now shows the developer portal entities; M3 observability kickoff specs have started; gitea-com push needs refreshed cloud authentication; and the Backstage dependency audit has existing critical/high advisories.

---

## 1. The single most important thing

**M2 no longer needs the Path B renderer rewrite, the m2i-7 registry-trust code change, or the fresh T22 pipeline proof.** All three are complete locally.

Current validation:
- `score2openchoreo` emits OpenChoreo `Component` + `SecretReference` + `Workload` multi-document YAML when Score secrets are used.
- `go test ./...` and `go build -o bin/score2openchoreo .` pass under `tools/score2openchoreo`.
- `./scripts/smoke-score.sh` passes.
- `kubectl --context k3d-openchoreo apply --dry-run=server -f /private/tmp/hello-m2-rendered.yaml` accepts the generated `hello-m2` OpenChoreo resources.
- local-registry is a NodePort Service on `30082`.
- `/etc/rancher/k3s/registries.yaml` on `k3d-openchoreo-server-0` maps `registry.local-registry.svc.cluster.local:5000` to `http://127.0.0.1:30082`.
- Gitea Actions run #24 on `hello-m2` commit `5d88625` completed successfully on 2026-05-22.
- Flux applied the CI-committed `Component` + `SecretReference` + `Workload` output from platform-config.
- `releasebinding.openchoreo.dev/hello-m2-development` is `Ready=True` and `ResourcesReady=True`.
- `deployment.apps/hello-m2-development-95297084` is `1/1` available in `dp-default-default-development-f8e58905`; pod `hello-m2-development-95297084-85c9fd7bcc-hd5qt` is `1/1 Running`.
- `./scripts/smoke-openbao.sh` passes after reseeding OpenBao's dev-mode kv mounts.
- `externalsecret/gitea-runner-token` is `SecretSynced=True` again.
- Backstage local dev runs on `http://127.0.0.1:3001` with backend `http://127.0.0.1:7008` in this workstation's current port layout.
- The Backstage catalog loads the repo root `catalog-info.yaml` plus `seed-repos/hello-m2/catalog-info.yaml`; rendered catalog smoke confirms `Component:default/developer-portal` and `Component:default/hello-m2`.
- `yarn npm audit --all --recursive` reaches the registry with network escalation but fails on existing transitive advisories, including critical/high `vm2`, `protobufjs`, `axios`, `tar`, `undici`, `fast-uri`, `fast-xml-builder`, and `basic-ftp` findings. No dependency files changed in the Backstage catalog work.
- M3 observability kickoff specs are added under `docs/specs/m3-observability/`; no M3 cluster resources have been installed.

The remaining external closeout is `git push gitea-com main` after refreshing the cloud Gitea credential/PAT. The 2026-05-23 retry reached `gitea.com` but failed authentication. The local Gitea origin has the M2 commits.

---

## 2. Git state at handoff

- **Branch:** `main`
- **Local HEAD:** see `git log --oneline -5`; the 2026-05-22 M2 closeout work
  should be the latest local commit once this handoff is committed.
- **gitea-com:** push attempt on 2026-05-21 failed because `gitea.com:443` was unreachable; 2026-05-23 retry reached `gitea.com` but failed authentication.
- **origin (local Gitea):** URL is `http://localhost:3333/openchoreo/developer-portal.git`; `main` has been pushed there.
- **Working tree:** should be clean after the Backstage catalog commit unless later M3 work has started. Do not revert unrelated changes.

Earlier 2026-05-02 commits:
```
a99d97e docs(m2): finalize -- track CLAUDE.md, document Gitea port migration
42b2231 fix(m2): complete CRD-group migration -- score2openchoreo + gatekeeper + flux platform-config watch
```

---

## 3. What was built / proved this session

### Documentation finalization (commit a99d97e)

- Added `CLAUDE.md` to repo (was untracked previously)
- Plan doc `2026-04-20-m2-iac-cd.md` got a Host Port Selection appendix capturing the 3002 -> 3333 Gitea port migration with substitution checklist for downstream files

### M2 install completion + retrospective review

Verified the M2 install completed successfully on 2026-04-24 (per session memory archive). Cluster state at session start:
- 1/1 nodes Ready
- 19 namespaces, all M2 components healthy: flux-system, gatekeeper-system, gitea, gitea-runners, local-registry, external-secrets, openbao, tofu-state, plus the openchoreo planes

Filed retrospective review thread on gitea.com: Issue #1 at https://gitea.com/trademomentum.net/developer-portal/issues/1 -- captures the 9 finalization commits and pending review prompts.

### Backstage catalog validation

On 2026-05-22, the Backstage app was moved off the scaffold catalog state:
- `backstage/app-config.yaml` title is `Developer Portal`.
- Catalog locations now include `../../../catalog-info.yaml` and `../../../seed-repos/hello-m2/catalog-info.yaml`.
- Root `catalog-info.yaml` now has a `Component` entry for `developer-portal` in addition to the `System`.
- `scripts/start-backstage.sh` defaults to `127.0.0.1:3001` for the app and `127.0.0.1:7008` for the backend, and prefers Homebrew `node@24` when present.
- `backstage/packages/app/e2e-tests/app.test.ts` validates Guest sign-in plus the `developer-portal` and `hello-m2` component links.

### CRD-group migration completion (commit 42b2231)

Task 22 acceptance test exposed two gaps after the post-handoff finalization:

**Gap 1: score2openchoreo + Gatekeeper still emitted/matched the stale `core.choreo.dev` API group.** Earlier commit 692f200 (2026-04-24) had fixed only the openchoreo-environments tofu module, missing:
- `tools/score2openchoreo/convert.go:39` (renamed to `openchoreo.dev/v1alpha1`)
- `tools/score2openchoreo/fixtures/*.component.yaml` (golden file regeneration)
- `tools/score2openchoreo/convert_test.go` (test expectation)
- `tools/score2openchoreo/README.md` (doc)
- `policies/C2-constraint.yaml`, `policies/C3-constraint.yaml`, `seed-repos/platform-addons/clusters/default/gatekeeper/constraints.yaml` (all updated)
- `docs/specs/m2-iac-cd/technical-specification.md` line 1008 (spec updated)

**Gap 2: Flux only watched `platform-addons`, never `platform-config`.** The CD chain stopped dead at "CI commits Component to platform-config" because nothing pulled it back into the cluster.

Fix: added GitRepository + dev/staging Kustomizations for platform-config in `iac/modules/flux/main.tf`. Tofu apply was clean: 3 to add, 0 to change, 0 to destroy. Flux now watches both repos.

### Path A end-to-end validation (no commits -- transient platform-config edits)

After the commit, triggered CI run #17 on hello-m2 (sha dc407cc). Run completed in ~88 seconds and committed a fresh `platform-config/environments/dev/hello-m2.yaml` with the corrected `openchoreo.dev/v1alpha1` API group.

Flux pulled it -- but Flux schema validation FAILED:
> `.spec.environment: field not declared in schema`

This exposed **Gap 3 (the deepest one): score2openchoreo's entire output model is wrong for the real CRD.** The cluster's `components.openchoreo.dev` CRD requires `spec.componentType` + `spec.owner.projectName` and uses a separate `Workload` CRD for container/image/env/ports. score2openchoreo conflated them.

Hand-wrote a schema-valid Component+Workload+Project triplet using the canonical pattern from `~/Projects/openchoreo/samples/from-image/go-greeter-service/greeter-service.yaml`. After three iterations (v1: only Component+Workload, no Project; v2: added Project but in wrong namespace; v3: moved everything to `default` ns to colocate with the existing DeploymentPipeline + Project), Flux applied successfully and OpenChoreo reconciled:

```
Component (Ready=True, ComponentReleaseReady)
  -> ComponentRelease hello-m2-6b5cd77c7b
    -> Deployment in dp-default-default-development-f8e58905
      -> Pod hello-m2-development-95297084-7676dd9cc-brrt8 (ImagePullBackOff)
```

Pod ImagePullBackOff cause: k3d/containerd doesn't resolve cluster DNS for `registry.local-registry.svc.cluster.local` and defaults to HTTPS on the HTTP-only registry. **This is m2i-7 in TODO** -- a cluster bootstrap config gap, not part of M2.

### Net findings about the actual OpenChoreo deployment model

Important learnings -- score2openchoreo's design assumptions are wrong:

1. **Components live in the same namespace as their Project + DeploymentPipeline** (here: `default`). NOT in `openchoreo-data-plane`.
2. **OpenChoreo auto-creates a per-environment data-plane namespace** for actual workloads, named `dp-<dataplane>-<project>-<environment>-<hash>`. Operators don't pick that namespace.
3. **Component is a thin abstraction**; the real deployable spec lives in `Workload`. Score's container/ports/env map to Workload, not Component.
4. **The four `ClusterComponentType`s** installed: `deployment/service`, `deployment/web-application`, `deployment/worker`, `cronjob/scheduled-task`. Score has no concept of this -- the renderer needs a heuristic or annotation to choose.

---

## 4. What is NOT yet done

### gitea-com push

External `gitea-com` push is blocked by cloud authentication as of 2026-05-23. The remote is reachable, but the cached credential was rejected. Local Gitea has the `developer-portal` repo and current `main`.

### m2i-6: OpenBao dev-mode storage

Still open and low-priority. It is production-readiness work, not M2 closeout.

### Backstage dependency audit remediation

`yarn npm audit --all --recursive` currently fails on existing critical/high transitive advisories. This needs a dedicated Backstage dependency alignment task before production hardening; it is not introduced by the catalog/dev-server validation change.

### M3 preflight and version inventory

M3 has specs only. The next implementation step is a read-only `scripts/preflight-m3.sh`, followed by pinned SigNoz, SigNoz K8s Infra, and OpenTelemetry Collector chart-version inventory. Do not install SigNoz or collector resources before that preflight is reviewed.

---

## 5. Live state at handoff

- **k3d-openchoreo cluster:** healthy. All M2 namespaces present.
- **Gitea local port state:** `localhost:3333` is not currently listening. A local `gitea` process is listening on `*:3000`, which is why Backstage dev defaults to `3001` here.
- **platform-config repo state:** run #24 committed renderer-generated multi-document YAML for `hello-m2`.
- **OpenChoreo workload state:** `releasebinding.openchoreo.dev/hello-m2-development` is Ready=True; deployment `hello-m2-development-95297084` is available and pod `hello-m2-development-95297084-85c9fd7bcc-hd5qt` is `1/1 Running`.
- **Backstage dev state:** the app was started at `http://127.0.0.1:3001` with backend `http://127.0.0.1:7008`; the browser may need a hard refresh after restarts because older console logs show transient connection-refused errors during restart windows.

---

## 6. Skills / agents to reach for in the next session

- Use the existing Go test and live dry-run loop before declaring M2 complete.
- For fresh pipeline validation, avoid direct `tofu apply`; use the repo scripts and Gitea workflow path.

---

## 7. What to do first in the next session

In this exact order:

1. Read this file.
2. Read `TODO.md`.
3. Read `PROJECT_SUMMARY.md` for the wider context.
4. `git status` and `git log --oneline origin/main..HEAD` (or gitea-com/main..HEAD) to verify state.
5. Confirm cluster is still healthy: `kubectl --context k3d-openchoreo get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded` should not show a stale `hello-m2` ImagePullBackOff after run #24.
6. If Gitea port-forward is no longer running: `kubectl --context k3d-openchoreo -n gitea port-forward svc/gitea-http 3333:3000 &`
7. For portal UI work, use `./scripts/start-backstage.sh`; on this workstation it defaults to `127.0.0.1:3001` and backend `127.0.0.1:7008`.
8. For M3, read `docs/specs/m3-observability/` and implement read-only `scripts/preflight-m3.sh` before any install work.

---

## 8. State of the three projects in one line each

- **openchoreo** (`/Users/nnos/Projects/openchoreo/`): unchanged, cluster healthy, used as reference for canonical Component+Workload patterns.
- **rational-reserve** (`/Users/nnos/Projects/rational-reserve/`): unchanged this session.
- **developer-portal** (`/Users/nnos/Projects/developer-portal/`): M1 + M2 architecturally complete and locally validated; Backstage catalog now exposes developer-portal plus hello-m2; M3 observability specs started; external gitea-com push needs refreshed cloud auth.
