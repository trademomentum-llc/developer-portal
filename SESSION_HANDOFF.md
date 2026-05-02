# SESSION HANDOFF

> Read this file FIRST in the next session. It tells you where we stopped,
> what is now committed that was not before, what is still outstanding, and
> exactly what to do first.

**Last session ended:** 2026-05-02 (mid-afternoon)
**Reason for handoff:** M2 architectural validation complete; substantial Path B (score2openchoreo rewrite) deserves a fresh session with clean energy.

---

## 1. The single most important thing

**M2 is architecturally validated end-to-end through Pod creation.** Push -> CI -> render -> commit -> Flux pulls -> Flux applies -> OpenChoreo reconciles -> ComponentRelease -> Deployment -> Pod. **Six links of the seven-link chain proved on a live cluster.** The final link (Pod runs) is blocked by an out-of-scope k3d/containerd registry trust gap (m2i-7), not anything in the M2 codebase.

But this only works with **manually-corrected Component+Workload YAML**. The score2openchoreo renderer emits the wrong CRD shape -- a substantial rewrite (Path B) is required before the chain works without manual intervention. See `TODO.md` -> `m2-renderer-rewrite` for the spec.

---

## 2. Git state at handoff

- **Branch:** `main`
- **Local HEAD:** `42b2231 fix(m2): complete CRD-group migration -- score2openchoreo + gatekeeper + flux platform-config watch`
- **gitea-com:** has 42b2231 (pushed this session via ephemeral PAT, then PAT revoked)
- **origin (local Gitea):** still 9 commits behind; URL points at stale port 3002 (Gitea is now on 3333)
- **Working tree:** clean

Two new commits this session:
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

### Path B: score2openchoreo rewrite (next session, ~2-4 hours)

See `TODO.md` -> `m2-renderer-rewrite` for the full spec (8 line items: score-6a..6h). Summary:

- Replace types.go with Component+Workload+Project+SecretReference shapes
- Convert function returns multi-document YAML (a list of resources)
- Decide componentType inference (heuristic vs annotation)
- Move default namespace from `openchoreo-data-plane` to `default`
- Rewrite all golden fixtures (multi-document)
- Update tests (convert_test, schema_test, main_test)
- Update README + technical-specification.md
- Update hello-m2 CI workflow if `--namespace` flag semantics change

### m2i-7: k3d registry trust (next session, ~30-60 min)

Add `/etc/rancher/k3s/registries.yaml` mirror entry on the k3d node so containerd resolves and pulls from the in-cluster registry. The fix likely belongs in `scripts/install-m1.sh` (cluster bootstrap), not M2 codebase. Without this, even after Path B, hello-m2 still won't run.

### Push 42b2231 to local Gitea origin

Stale `localhost:3002` URL; Gitea is on 3333. Update remote URL and push when port-forward is conveniently up. Not blocking anything.

---

## 5. Live state at handoff

- **k3d-openchoreo cluster:** healthy. All M2 namespaces present.
- **Gitea port-forward:** I started one this session at localhost:3333 (background task `b6zf5th1l`). Still running unless terminated. Safe to leave.
- **platform-config repo state:** has hand-written Component+Workload+Project YAML at `environments/dev/hello-m2.yaml` (v3, commit c6656c8 in platform-config). Flux is currently applying this. **Heads up:** when Path B's renderer ships and a fresh CI run happens, that file will be overwritten with the new auto-generated multi-document YAML. The hand-written copy was for chain validation only.
- **OpenChoreo workload state:** Component `hello-m2`, Workload `hello-m2-workload`, ComponentRelease `hello-m2-6b5cd77c7b` all live in `default` namespace. Deployment in `dp-default-default-development-f8e58905`. Pod stuck ImagePullBackOff.

---

## 6. Skills / agents to reach for in the next session

- **superpowers:writing-plans** to spec the Path B rewrite before touching code (the score2openchoreo redesign is multi-step enough to warrant a plan).
- **superpowers:test-driven-development** for the actual rewrite -- new fixtures + new tests are the natural TDD signal.
- **superpowers:verification-before-completion** at the end -- a Path B that "passes its own tests" but doesn't render schema-valid YAML against the live cluster CRD is the same trap that bit us in 2026-04-24.

---

## 7. What to do first in the next session

In this exact order:

1. Read this file.
2. Read `TODO.md` (especially `m2-renderer-rewrite` and m2i-7).
3. Read `PROJECT_SUMMARY.md` for the wider context.
4. `git status` and `git log --oneline origin/main..HEAD` (or gitea-com/main..HEAD) to verify state.
5. Confirm cluster is still healthy: `kubectl --context k3d-openchoreo get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded` should return only the long-Completed cluster-agent-dataplane pod.
6. If Gitea port-forward is no longer running: `kubectl --context k3d-openchoreo -n gitea port-forward svc/gitea-http 3333:3000 &`
7. Decide with the operator: start Path B (the renderer rewrite) OR address m2i-7 first (so when Path B ships, the demo actually shows a running pod).

---

## 8. State of the three projects in one line each

- **openchoreo** (`/Users/nnos/Projects/openchoreo/`): unchanged, cluster healthy, used as reference for canonical Component+Workload patterns.
- **rational-reserve** (`/Users/nnos/Projects/rational-reserve/`): unchanged this session.
- **developer-portal** (`/Users/nnos/Projects/developer-portal/`): M1 + M2 architecturally complete; M2 chain validated through Pod creation; score2openchoreo renderer rewrite (Path B) outstanding; k3d registry trust (m2i-7) outstanding.
