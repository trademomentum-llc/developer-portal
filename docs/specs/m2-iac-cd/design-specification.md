# M2 IaC + CD Loop -- Design Specification

> **Milestone:** M2 -- IaC + CD Loop
> **Version:** 1.0
> **Date:** 2026-04-20
> **Status:** Draft, awaiting user approval
> **Companion docs:** [requirements.md](./requirements.md), [technical-specification.md](./technical-specification.md)

---

## 1. Purpose

This document describes *how* the M2 IaC + CD loop is shaped. It sits one level above the Technical Specification: it explains why the pieces are arranged the way they are, where the boundaries between them live, and what trade-offs were accepted. It does not list file paths, pinned versions, or command snippets -- those belong in the Technical Specification.

Read this document to understand architecture. Read the Technical Specification to know what to type.

## 2. Context diagram

```
                 +----------------------------------------------+
                 |             Operator's macOS laptop          |
                 |   Host CLIs: tofu, flux, infracost, score,   |
                 |   score2openchoreo, Backstage (yarn dev),    |
                 |   Claude Code + rr-policy-guards             |
                 +-----------+------------------+---------------+
                             |                  |
                             | git push         | catalog + proxy
                             v                  v
         +-------------------+------------------+----------------+
         |                 Colima VM (docker daemon)             |
         |                                                       |
         |    +---------------------------------------------+    |
         |    |              k3d cluster (M1)                |    |
         |    |                                              |    |
         |    |  Repos (hosted in Gitea, from M1):           |    |
         |    |   +-- platform-addons  <----- watched by Flux |    |
         |    |   +-- platform-config  <----- watched by      |    |
         |    |   |                           OpenChoreo      |    |
         |    |   +-- hello-m2         <----- CI source       |    |
         |    |                                              |    |
         |    |  New namespaces in M2:                       |    |
         |    |   +-- flux-system       Flux controllers     |    |
         |    |   +-- gatekeeper-system OPA/Gatekeeper       |    |
         |    |   +-- gitea-runners     Actions runner pod   |    |
         |    |                                              |    |
         |    |  Reused from M1:                             |    |
         |    |   +-- openbao           Tofu state + secrets |    |
         |    |   +-- external-secrets  Secret materializer  |    |
         |    |   +-- openchoreo-*      Workload runtime     |    |
         |    |   +-- gitea             Repo host + OCI reg  |    |
         |    +---------------------------------------------+    |
         +-------------------------------------------------------+
```

The flow of a code change through M2, at a glance:

```
dev pushes to hello-m2 (PR)
    -> Gitea Actions CI fires on the runner
        -> Score schema validation
        -> tofu plan on iac/ (no-op unless IaC changed)
        -> Infracost breakdown, comment to PR
        -> Gatekeeper pipeline constraints evaluated
        -> build image, tag sha, push to Gitea OCI registry (build only; push on merge)
    (merge)
    -> CI on main:
        -> render Component via score2openchoreo
        -> commit rendered manifest to platform-config repo (dev environment)
        -> OpenChoreo reconciles, hello-m2 starts in dev namespace

dev opens PR on platform-config moving hello-m2 from dev to staging
    -> same CI gates run
    (merge)
    -> OpenChoreo reconciles, hello-m2 starts in staging namespace

Flux, underneath all of this, reconciles platform-addons every minute,
reverting any drift in Flux's own scope (Gatekeeper policies, runner
helm release, add-on charts).
```

## 3. Component boundaries

M2 has eight components, each with a single clearly-owned responsibility. Six are in-cluster additions, one is a host-side tool, one is a new PreToolUse hook.

| Component | Responsibility | Runs where | Language | New in M2 |
|---|---|---|---|---|
| **Flux** | Drift correction for cluster add-ons | In cluster, `flux-system` | Go (upstream) | yes |
| **Gatekeeper** | Admission-time pipeline constraints | In cluster, `gatekeeper-system` | Go (upstream) | yes |
| **Gitea Actions runner** | CI executor, one pod per job | In cluster, `gitea-runners` | Go (upstream) | yes |
| **OpenTofu modules** | Cluster-side IaC (Flux, Gatekeeper, runner, OpenChoreo Environment CRDs) | Invoked from runner | HCL | yes |
| **score2openchoreo** | Score YAML to OpenChoreo Component CRD converter | CLI, invoked from runner and host | **Go** | yes |
| **Infracost** | Pre-deploy cost estimation, PR comment | Invoked from runner | Go (upstream) | yes |
| **rr-tofu-guard** | Block direct `tofu apply` from Bash tool use | PreToolUse hook on host | **Go** | yes |
| **Seeded Gitea repos** | `platform-addons`, `platform-config`, `hello-m2` | In Gitea (from M1) | -- | yes (content) |

Reused unchanged from M1: Gitea (repos and OCI registry), openbao (state and secrets), external-secrets (materializer), OpenChoreo (workload reconciler), Backstage (catalog surfacer).

No component owns another component's responsibilities. Communication happens only via published APIs: OpenTofu reads and writes Kubernetes and Helm via their providers; Flux watches git via Gitea's HTTPS git protocol; the runner calls Gitea's REST API; score2openchoreo reads YAML on stdin and emits YAML on stdout; the tofu-guard reads JSON on stdin per Claude Code's PreToolUse contract.

## 4. Design decisions and rationale

### 4.1 Why Flux for drift correction, not Argo CD, and not for application workloads

Three options were considered for the CD loop's reconciliation layer:

1. **Argo CD as the single deployer.** Argo CD watches a repo and reconciles every manifest. Would require OpenChoreo to step back from workload reconciliation. Contradicts M1's explicit decision that OpenChoreo is load-bearing for workload delivery.
2. **Flux only for cluster add-ons; OpenChoreo owns workloads** (chosen). Flux watches `platform-addons`. OpenChoreo watches `platform-config`. No overlap: Flux never touches a workload Deployment, OpenChoreo never touches a Gatekeeper ConstraintTemplate.
3. **No GitOps controller at all.** OpenChoreo covers workloads; add-ons drift silently until someone notices. Simplest, but the whole point of GitOps is continuous reconciliation.

Option 2 was confirmed by the operator on 2026-04-20 after a brief exchange about the roadmap's earlier "Argo-style GitOps" placeholder. What is already running in the cluster named "argo" is Argo Workflows bundled by OpenChoreo, a completely different piece of software.

**Trade-off accepted:** Two reconciliation loops to reason about. Mitigated by the strict repo boundary -- `platform-addons` is Flux territory, `platform-config` is OpenChoreo territory, and they never overlap.

**Reversal cost:** low. If a future milestone wants Flux to reconcile workloads too (e.g., if OpenChoreo is swapped out), moving `platform-config` under Flux is a single Kustomization addition.

### 4.2 Why three repos (platform-addons, platform-config, hello-m2), not one

A monorepo where add-ons, rendered workload manifests, and application source coexist was considered and rejected. Three reasons:

1. **Reconciliation scope.** Flux watches one path. OpenChoreo watches another. A monorepo would force both tools to watch the same revision stream and tooling-filter it, doubling the surface for "a commit changed the wrong thing." Separate repos make scope obvious from the URL alone.
2. **Permissions and audit.** `platform-addons` and `platform-config` are platform-managed; only the CI runner should push to them. `hello-m2` is application-owned. Splitting them lets Gitea branch protection and the Gatekeeper C-1 constraint enforce cleanly.
3. **Blast radius on rollback.** A bad Component render should not be revertable in the same commit stream as a Gatekeeper policy. Separate repos -> separate `git revert` targets -> smaller blast radius per rollback.

**Trade-off accepted:** Three remotes to keep in sync during a major refactor instead of one. Mitigation: a rendering pipeline step in CI handles almost all cross-repo commits; humans rarely push to `platform-config` directly.

### 4.3 Why score2openchoreo is a separate converter, not a teach-OpenChoreo-Score effort

Three options were considered to get Score input into OpenChoreo's Component model:

1. **Teach OpenChoreo to accept Score directly.** Would require an upstream contribution to OpenChoreo -- weeks of review, potential rejection, and creates a fork risk. Out of scope for M2.
2. **Use `score-k8s` and render raw Deployments and Services.** Bypasses OpenChoreo entirely. Contradicts M1.
3. **Write `score2openchoreo` as a small Go converter** (chosen). Reads Score on stdin, emits an OpenChoreo Component YAML on stdout. ~200 to 400 lines of Go. Runs in CI as a step. No changes to OpenChoreo.

The converter is the bridge between the authoring surface (Score, per the user's locked-in tool list) and the orchestrator surface (OpenChoreo). It has one job and can be tested exhaustively against fixture Score files.

**Trade-off accepted:** Another small Go binary to maintain. Offset by the fact that Score's schema is stable and OpenChoreo's Component schema is stable; the converter changes only when one of those changes. Golden-file tests catch drift loudly.

**Reversal cost:** zero-ish. If OpenChoreo later accepts Score natively, delete the converter and replace the CI step with `kubectl apply -f score.yaml`.

### 4.4 Why the kubernetes backend for OpenTofu state

Four options were considered for Tofu state:

1. **Local state files.** Never. State files contain resource IDs and sometimes secrets. Never committed, never on a laptop that could be lost.
2. **MinIO in-cluster with S3 backend.** Works, but introduces a new service in M2.
3. **openbao kv v2 via the generic `http` backend plus a shim.** openbao is already running, but OpenTofu has no native openbao backend -- this route requires an in-cluster sidecar (~150 lines of Go) that translates the http backend protocol to openbao's kv v2 API. Too much M2 surface for a problem with a simpler solution.
4. **OpenTofu's native `kubernetes` backend** (chosen). Stores state as a Secret in a dedicated `tofu-state` namespace. Zero new services, zero custom shims. RBAC restricts access to the runner ServiceAccount. State is etcd-stored, which for a local k3d cluster is acceptable at M2 scale.

**Trade-off accepted:** State is not off-cluster at M2, so a cluster loss loses state. Mitigation: the state is reproducible by re-running `tofu plan` against the cluster (since the resources in state are cluster-internal). A scheduled snapshot to openbao is a natural M3 or M4 addition and does not require changing the backend -- `tofu state pull` and push to openbao kv works today.

**Reversal cost:** low. If we ever need a cloud-ready state backend (S3), it is a `backend "s3"` block change plus a one-time `tofu state push`.

### 4.5 Why two environments (dev and staging), not one or five

One environment (dev) would not exercise promotion at all; the whole point of M2 is to prove a commit-driven promotion model. Five environments (dev, qa, perf, staging, prod) would pad M2 with environment management that does not belong in a local k3d substrate -- that is cloud-cluster territory (M-later).

Two is the minimum to prove promotion works. Trade-off accepted: if the operator later wants qa and perf, each is a namespace addition plus one OpenChoreo Environment CRD; not a redesign.

### 4.6 Why the Gitea Actions runner runs in-cluster, not on host

Three options were considered:

1. **Host runner** (`act_runner` on the laptop). Simplest. Runs as the operator's user. Has access to everything the operator has access to -- the wrong blast radius.
2. **Containerized runner on the Colima docker daemon outside the cluster.** Better isolation, but creates a second "place where platform things run" (cluster and docker-compose). M1 already rejected this shape for Gitea.
3. **In-cluster runner via the official `act_runner` helm chart** (chosen). Runs as a constrained ServiceAccount in its own namespace. Uses Kubernetes RBAC. Ephemeral pod per job. Inherits openbao via external-secrets like every other component. Teardown is a namespace delete.

**Trade-off accepted:** In-cluster runner pods compete with the M1 substrate for cluster resources. Mitigation: NFR-2 already budgeted 8 GB, observed usage after M1 is well below that.

### 4.7 Why Infracost lives in the PR gate, not Backstage (yet)

Infracost produces a JSON breakdown. Two consumption surfaces were considered:

1. **PR comment via the runner** (chosen for M2). Every IaC-touching PR gets a comment showing monthly delta. Reviewer sees it inline. This is where cost decisions are made.
2. **Backstage plugin visualizing Infracost artifacts.** Correct surface long-term, wrong timing for M2. Building a Backstage plugin means entity mapping, storage, auth. Deferred to the same milestone that brings OpenCost (M4), at which point a unified cost view makes sense.

**Trade-off accepted:** Non-IaC PRs on `hello-m2` do not get a cost comment (nothing changed in `iac/`). Acceptable because those PRs do not change cost, by definition.

### 4.8 Why only three Gatekeeper constraints in M2

Gatekeeper can express many runtime policies (pod security, allowed images, required labels, ingress TLS, etc.). M2 installs and configures Gatekeeper but restricts its enforcement surface to three pipeline-scoped constraints:

- **C-1:** `platform-addons` main branch rejects direct commits
- **C-2:** Score schema validates before render
- **C-3:** Infracost monthly delta below threshold

Rationale for the minimalism: Gatekeeper's broader policy surface is M6's assignment. Pulling all of it forward into M2 would conflate two milestones. Three constraints are enough to prove the pipeline has admission gates; every subsequent constraint added in M6 uses the same ConstraintTemplate shape.

**Trade-off accepted:** A misbehaving pod can still run in M2 because there is no pod-security constraint. Acceptable -- the cluster is a laptop-local dev cluster, not internet-facing.

### 4.9 Why promotion is a commit, not a button

Two patterns were considered:

1. **Backstage UI button ("promote to staging").** Gives operators a one-click experience. Requires a Backstage action plugin, auth, audit trail inside Backstage. Scope creep into Backstage work that belongs later.
2. **Commit to `platform-config` that updates `targetEnvironment`** (chosen). Every promotion is a reviewable git commit. Audit trail is `git log`. Rollback is `git revert`. No additional tooling required in M2.

**Trade-off accepted:** More friction for a promotion than a button. Acceptable, arguably preferred -- promotions should leave a trail.

## 5. Data flow

### 5.1 Install flow

```
operator runs scripts/install-m2.sh
    |
    +-- [Task 0] build rr-tofu-guard binary
    |     +-- register in ~/.claude/settings.json as PreToolUse hook
    |     +-- update brew-guard tap allow-list (flux, score-spec)
    |
    +-- [Task 1] brew install opentofu infracost flux score-k8s   (validated by brew-guard)
    |
    +-- [Task 2] cd into iac/  (developer-portal repo)
    |     +-- tofu init   (openbao http backend)
    |     +-- tofu apply
    |           +-- creates namespaces: flux-system, gatekeeper-system, gitea-runners
    |           +-- helm_release.flux
    |           +-- helm_release.gatekeeper
    |           +-- helm_release.gitea_actions_runner
    |           +-- kubernetes_manifest for OpenChoreo Environments (dev, staging)
    |           +-- external_secrets_io_v1beta1.ExternalSecret manifests for runner token, app secrets
    |
    +-- [Task 3] seed Gitea repos via the Gitea API
    |     +-- create openchoreo organization
    |     +-- create platform-addons repo (seeded with Flux Kustomization)
    |     +-- create platform-config repo (seeded with empty Environments dirs)
    |     +-- create hello-m2 repo (seeded with main.go, Dockerfile, score.yaml, catalog-info.yaml, .gitea/workflows/ci.yaml)
    |
    +-- [Task 4] wait for Flux to reconcile platform-addons once
    |
    +-- [Task 5] run scripts/smoke-m2.sh
          +-- every per-tool check must pass
```

### 5.2 PR flow (hello-m2)

```
dev opens PR on hello-m2 with code + score.yaml change
    |
    v
Gitea Actions workflow triggered in runner pod
    |
    +-- checkout + set up Go
    |
    +-- validate score.yaml against Score schema  (constraint C-2 contributes here)
    |
    +-- if iac/ changed in the PR: tofu init && tofu plan && infracost breakdown
    |     +-- post Infracost output as a PR comment via Gitea API
    |     +-- constraint C-3: fail the workflow if monthly delta > threshold
    |
    +-- build container image (no push)
    |
    +-- if anything failed: runner marks the workflow red, reviewer sees in PR UI
    +-- if all green: PR is mergeable
```

### 5.3 Merge flow (hello-m2 -> platform-config)

```
PR merged on hello-m2 main
    |
    v
Gitea Actions workflow on main branch
    |
    +-- checkout
    |
    +-- build + push image
    |     +-- image tag = git short SHA
    |     +-- push target = gitea.gitea.svc.cluster.local:3000/openchoreo/hello-m2:<sha>
    |
    +-- render OpenChoreo Component via score2openchoreo
    |     +-- input: score.yaml
    |     +-- output: Component YAML with image ref pinned to :<sha>
    |
    +-- clone platform-config
    |     +-- write rendered YAML to environments/dev/hello-m2.yaml
    |     +-- commit with "chore: hello-m2 dev -> <sha>"
    |     +-- push via runner's service-account token (sourced from openbao)
    |
    +-- OpenChoreo controller sees the platform-config change
          +-- reconciles Component into Deployment + Service in dev namespace
```

### 5.4 Promotion flow (dev -> staging)

```
dev opens PR on platform-config
    |
    v
PR diff: environments/staging/hello-m2.yaml now points to the same image SHA as dev
    |
    +-- pipeline validates against Score schema (C-2)
    +-- no IaC change -> no Infracost step
    +-- Gatekeeper admits
    |
    (merge)
    |
    v
OpenChoreo controller sees the platform-config change
    +-- reconciles Component into Deployment + Service in staging namespace
    +-- dev remains running at whatever SHA was last promoted to dev
```

### 5.5 Drift correction flow (Flux on platform-addons)

```
Flux Source Controller polls platform-addons every minute
    |
    v
If revision changed:
    +-- Kustomization Controller applies the new manifest set
    +-- Gatekeeper, runner helm release, external-secrets resources converge

If revision unchanged:
    +-- Kustomization Controller still applies the same manifest set
    +-- Any manual drift (kubectl edit, someone poking at a helm release) is reverted
```

## 6. File structure (top-level, fully materialized in Technical Spec)

```
/Users/nnos/Projects/developer-portal/
+-- docs/
|   +-- specs/
|       +-- m1-substrate/                    # from M1
|       +-- m2-iac-cd/
|           +-- requirements.md
|           +-- design-specification.md      # this file
|           +-- technical-specification.md
|   +-- superpowers/
|       +-- plans/
|           +-- 2026-04-20-m2-iac-cd.md      # produced later by writing-plans
+-- iac/
|   +-- README.md
|   +-- main.tf
|   +-- variables.tf
|   +-- outputs.tf
|   +-- modules/
|   |   +-- flux/
|   |   +-- gatekeeper/
|   |   +-- gitea-runner/
|   |   +-- openchoreo-environments/
|   |   +-- external-secrets-wiring/
|   +-- templates/
|       +-- ci.yaml                          # canonical Gitea Actions workflow
+-- plugins/
|   +-- rr-policy-guards/
|       +-- tools/
|       |   +-- tofu-guard/                  # new in M2
|       +-- bin/rr-tofu-guard                # build output (gitignored)
+-- tools/
|   +-- score2openchoreo/
|       +-- go.mod
|       +-- main.go
|       +-- convert.go
|       +-- main_test.go
|       +-- convert_test.go
|       +-- fixtures/                        # golden input/output pairs
+-- policies/
|   +-- C1-platform-addons-main-protected.rego
|   +-- C2-score-schema-valid.rego
|   +-- C3-infracost-delta.rego
+-- scripts/
|   +-- install-m2.sh
|   +-- teardown-m2.sh
|   +-- smoke-m2.sh
|   +-- smoke-tofu.sh
|   +-- smoke-actions.sh
|   +-- smoke-flux.sh
|   +-- smoke-score.sh
|   +-- smoke-infracost.sh
|   +-- smoke-gatekeeper.sh
|   +-- smoke-openbao.sh
+-- seed-repos/                              # one-time seed content for the three Gitea repos
    +-- platform-addons/
    +-- platform-config/
    +-- hello-m2/
```

Exact contents of every file are the Technical Specification's job.

## 7. Interfaces

### 7.1 score2openchoreo CLI

- **Input:** a Score YAML document on stdin. Alternately, `score2openchoreo --input path/to/score.yaml`.
- **Output:** an OpenChoreo `Component` CRD manifest on stdout.
- **Exit codes:** 0 on success, 1 on Score validation failure (with error on stderr), 2 on output encoding failure.
- **Flags:**
  - `--environment <name>` -- sets metadata.labels to pin the emitted Component to an Environment
  - `--image <ref>` -- overrides the Score container image reference (used in CI to pin to git SHA)
  - `--validate-only` -- run Score schema validation and exit without emitting output

### 7.2 rr-tofu-guard

- **Input:** JSON on stdin, matching Claude Code's PreToolUse protocol.
- **Output:** none on allow; human-readable error on stderr on block.
- **Exit codes:** 0 = allow, 2 = block, 1 = internal error (logged, defaults to block).
- **Environment variables:**
  - `RR_TOFU_GUARD_BYPASS=1` -- override (logged as bypass)
  - `RR_TOFU_GUARD_AUDIT_LOG` -- override audit log path

### 7.3 Gitea Actions API consumed

M2 calls the following Gitea endpoints from the runner:

- `POST /api/v1/repos/{owner}/{repo}/issues/{index}/comments` -- post Infracost PR comment
- `GET /api/v1/repos/{owner}/{repo}/actions/runs` -- enumerate workflow runs (used by smoke-actions)
- `POST /api/v1/repos/{owner}/{repo}/contents/{filepath}` -- commit rendered manifest to platform-config
- `POST /api/v1/orgs` -- one-time org creation
- `POST /api/v1/orgs/{org}/repos` -- one-time repo creation
- `POST /api/v1/repos/{owner}/{repo}/packages` (OCI) -- image push uses Docker auth, not REST, but ACLs are checked via this surface

### 7.4 openbao kv paths consumed

| Path | Reader | Purpose |
|---|---|---|
| `kv/gitea/runners/token` | external-secrets -> runner ServiceAccount | Runner registration |
| `kv/flux/gitea-deploy-key` | external-secrets -> flux-system Secret | Flux git auth against Gitea |
| `kv/apps/hello-m2/dev/example-secret` | external-secrets -> Deployment env | Demo workload secret |

OpenTofu state lives separately, in a Kubernetes Secret managed by the native `kubernetes` backend, not in openbao (see Section 4.4).

## 8. Error handling and failure modes

| Failure | Observable symptom | Response |
|---|---|---|
| Flux cannot pull from Gitea | Kustomization condition NotReady | `flux events` shows auth error; operator rotates the Flux deploy key |
| Gatekeeper ConstraintTemplate fails to compile | Template object has status not Ready | install-m2.sh detects and fails loudly; no Constraint is created |
| Gitea Actions runner cannot register | runner pod crashloops | runner logs show token error; operator verifies openbao path and external-secrets sync |
| score2openchoreo emits invalid Component YAML | OpenChoreo controller rejects the manifest in platform-config | OpenChoreo status shows validation error; CI pipeline Red on the next run because Component is not Ready |
| Infracost run fails (offline or token error) | PR comment missing | runner marks the workflow yellow but does not block merge; operator decides case by case |
| openbao unsealed but network partition to runner | runner pod crashloops fetching secrets | external-secrets logs show timeout; operator restarts runner pod after network restored |
| tofu plan differs from applied state outside CI | no automatic alert in M2 | caught on next `tofu plan` in CI; M3 will add drift alerting |
| rr-tofu-guard binary missing when Claude Code tries `tofu apply` via Bash | PreToolUse hook exits non-zero | Claude Code surfaces the error; operator runs Task 0 of install |

## 9. Alternatives considered and rejected

| Alternative | Rejected because |
|---|---|
| Argo CD as the single deployer | Contradicts M1's OpenChoreo decision; user confirmed Flux for drift-only |
| Monorepo for addons + config + app | Reconciliation scope and rollback blast radius favor three repos |
| Score rendering raw Deployments | Bypasses OpenChoreo; contradicts M1 |
| Teaching OpenChoreo to read Score natively | Upstream work, out of scope |
| Host-based Gitea Actions runner | Too-broad blast radius; breaks cluster-isolation pattern |
| MinIO for OpenTofu state | Adds a new service; openbao already runs |
| Local state files | Never -- state can contain secrets |
| Backstage button for promotion | Nice UX, scope creep; commit-based promotion leaves a git trail |
| All of Gatekeeper's runtime policies pulled into M2 | Conflates M6 with M2 |
| Automatic dev-to-staging promotion | Human approval is a feature, not a bug |
| GitHub Actions or Jenkins instead of Gitea Actions | User's locked-in stack is Gitea Actions |

## 10. Open questions resolved in this document

- **Q1: Argo CD or Flux?** -> Flux, cluster add-ons only. Section 4.1.
- **Q2: One repo or many?** -> Three repos: platform-addons, platform-config, hello-m2. Section 4.2.
- **Q3: How does Score reach OpenChoreo?** -> score2openchoreo converter in Go. Section 4.3.
- **Q4: Where does Tofu state live?** -> openbao kv v2. Section 4.4.
- **Q5: How many environments?** -> Two (dev, staging). Section 4.5.
- **Q6: Where does the runner run?** -> In-cluster, `gitea-runners` namespace. Section 4.6.
- **Q7: Where does Infracost surface?** -> PR comment via runner. Section 4.7.
- **Q8: How many Gatekeeper constraints in M2?** -> Three pipeline-scoped. Section 4.8.
- **Q9: How does promotion work?** -> Commit to platform-config. Section 4.9.

## 11. Open questions deferred to Technical Specification

- **Q10:** Exact pinned versions of Flux, Gatekeeper, act_runner chart
- **Q11:** Exact openbao policy and role configuration for the runner, Tofu state, and external-secrets
- **Q12:** Exact openbao http backend wrapper (shim script, ConfigMap, or sidecar)
- **Q13:** Exact score2openchoreo Go module layout and Score library dependency
- **Q14:** Exact Gatekeeper ConstraintTemplate Rego text
- **Q15:** Exact Gitea Actions workflow YAML (ci.yaml canonical template)
- **Q16:** Exact `platform-config` directory layout per environment

## 12. Success criteria revisited

This document is well-formed when:

- Every functional requirement in `requirements.md` maps to a component or design decision here
- Every design decision names its alternatives and explains the trade-off
- Every interface between components is documented
- Every alternative considered is recorded
- Nothing in this document contradicts `requirements.md`
- The "what is Flux's scope vs OpenChoreo's scope" question has exactly one answer (Flux = cluster add-ons, OpenChoreo = application workloads) and that answer is visible in at least three places (component boundaries, repo topology, reconciliation flow)

These six tests are the self-review checklist.
