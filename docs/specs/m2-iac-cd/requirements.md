# M2 IaC + CD Loop -- Requirements Document

> **Milestone:** M2 -- IaC + CD Loop (second of seven)
> **Version:** 1.0
> **Date:** 2026-04-20
> **Status:** Draft, awaiting user approval
> **Companion docs:** [design-specification.md](./design-specification.md), [technical-specification.md](./technical-specification.md)

---

## 1. Purpose

M2 builds the Integration and Delivery loop on top of the M1 substrate. Where M1 proved the platform is installable, reachable, and wired, M2 proves that a human or an agent can commit code in Gitea, see cost and policy gates fire in a pull request, and have a merged change flow through the platform to a running workload on the cluster, with drift correction running underneath for the cluster add-ons that surround it.

This document captures what M2 must do. It does not describe how -- that is the Design Specification's job.

## 2. Context

M1 shipped the substrate: a k3d cluster with OpenChoreo (control, data, observability, and workflow planes), Gitea, cert-manager, external-secrets, openbao, Argo Workflows (OpenChoreo-internal), and a Backstage skeleton. All pods reach `Ready`. The three M1 spec documents and the `rr-policy-guards` plugin are in place. The `emoji-guard`, `bash-guard`, and `brew-guard` hooks are registered.

M2 extends that substrate with an end-to-end IaC + CD loop. Per the user's locked-in tool list, the Integration and Delivery stack is: **Gitea + Backstage Software Catalog and Score + OpenTofu + Gitea Actions + an in-cluster OCI registry**. The initial draft named Gitea's OCI package registry, but the implemented M2 image path uses the dedicated `local-registry` module so both Gitea Actions and k3s containerd can push/pull over stable in-cluster HTTP endpoints. Flux was added on 2026-04-20 as the GitOps controller for cluster add-ons drift correction; the canonical tool list should be updated to reflect it when this spec is approved. Infracost is cross-cutting -- the user's list categorizes it under Observability (stacked with OpenCost and Cloud Custodian), but its M2 use is pre-deploy cost estimation in the Gitea Actions PR gate. Its post-deploy dashboards land in M3/M4.

OPA/Gatekeeper is pulled forward from M6 for a narrow set of pipeline-scoped constraints. The broader runtime policy surface stays M6.

## 3. Stakeholders

| Stakeholder | Role | Interest in M2 |
|---|---|---|
| User (platform owner) | Sole decision maker | A working commit-to-deploy pipeline they can demonstrate end to end |
| Team members (future) | Developers consuming the IDP | Push a repo to Gitea, Backstage surfaces it, Score renders it, cluster runs it |
| AI coding agents (CC, Codex, Qwen-Code, OpenCode, Mistral Vibe) | First-class consumers | Same pipeline as humans; token-gated access lands in M7 |
| Rational Reserve (RR) | Future consumer | M7 integration; M2 does not couple to RR |
| Finance (future) | Budget owner | Infracost comments in PRs give visibility before cost lands |
| Security (future) | Policy owner | Pipeline-scoped Gatekeeper constraints plus openbao-sourced secrets prove the shape that M6 will generalize |

## 4. Functional Requirements

### 4.1 OpenTofu

- **FR-1** OpenTofu SHALL be available on the operator's host PATH at a pinned version.
- **FR-2** A canonical OpenTofu root module SHALL exist at `iac/` in the developer-portal repo that provisions M2's Kubernetes-side resources (namespaces, helm releases for Flux and Gatekeeper, Gitea Actions runner install, Gitea Actions runner registration secret stub) via the Kubernetes and Helm providers.
- **FR-3** The OpenTofu state backend SHALL be the native `kubernetes` backend, storing state as a Secret in a dedicated `tofu-state` namespace with RBAC restricting access to the Gitea Actions runner ServiceAccount. Local state files SHALL NOT be committed. Snapshotting state to openbao for off-cluster durability is explicitly deferred to a later milestone.
- **FR-4** `tofu init && tofu plan` SHALL succeed from a clean clone against the M1 cluster, producing a plan with no errors.
- **FR-5** `tofu apply` SHALL be invoked only by the Gitea Actions pipeline, never by a human directly. A pre-commit hook SHALL reject `tofu apply` invocations from a shell tool use.

### 4.2 Gitea Actions

- **FR-6** A Gitea Actions runner SHALL be installed in the cluster in a dedicated `gitea-runners` namespace, registered against the Gitea instance from M1.
- **FR-7** The runner SHALL be able to execute at least one smoke-test workflow that prints a line and exits 0.
- **FR-8** Every pipeline workflow SHALL run in an isolated ephemeral pod (no state persisted between runs beyond what is pushed to Gitea or written to openbao).
- **FR-9** Runner registration tokens SHALL be stored in openbao and synced into the runner namespace via external-secrets; they SHALL NOT appear in any committed file, helm values, or env var written to disk.
- **FR-10** The CI entry point for an application repository SHALL be `.gitea/workflows/ci.yaml`. A canonical template SHALL live at `iac/templates/ci.yaml` and be the reference copy Backstage scaffolders (later milestone) use.

### 4.3 Flux

- **FR-11** Flux SHALL be installed in the cluster in the `flux-system` namespace at a pinned version.
- **FR-12** Flux SHALL reconcile from a single "cluster add-ons" repository hosted in Gitea named `platform-addons`.
- **FR-13** Flux's reconciliation scope SHALL be restricted to cluster add-ons (Gatekeeper policies, external-secrets resources that M2 creates, Gitea Actions runner helm release, Flux itself via GitOpsKustomization patterns). Flux SHALL NOT reconcile application workload manifests -- those are OpenChoreo's responsibility.
- **FR-14** Flux SHALL emit a reconciliation event to its standard logs for every sync. No additional observability stack is required in M2 (M3 covers tracing).

### 4.4 Score and score2openchoreo

- **FR-15** Application authors SHALL describe workloads using Score (`score.yaml`), not directly in OpenChoreo Component YAML.
- **FR-16** A small converter binary named `score2openchoreo` SHALL be written in Go, with stdlib-only dependencies except for a pinned Score schema library, and SHALL take a `score.yaml` file on input and emit OpenChoreo `Component` and `Workload` CRD manifests on stdout.
- **FR-17** The converter SHALL support, at minimum: container image reference, container resource requests (cpu and memory), container ports, environment variables, and at least one workload-scoped secret sourced via external-secrets from openbao.
- **FR-18** A workload author SHALL NOT need to read OpenChoreo CRD schemas to ship a workload. Score is the authoring surface.

### 4.5 Infracost

- **FR-19** Infracost SHALL be available to the Gitea Actions runner and SHALL run on every pull request that changes anything under `iac/`.
- **FR-20** Infracost output SHALL be posted as a PR comment via the Gitea Actions runner using Gitea's API.
- **FR-21** A configurable cost-delta threshold SHALL gate merges. If a PR's estimated monthly cost delta exceeds the threshold, a Gatekeeper constraint SHALL mark the PR pipeline as failed (the runner checks the Infracost artifact against the policy). The default threshold is $50/month delta for M2.
- **FR-22** Infracost usage in M2 is pre-deploy only. Post-deploy cost attribution (OpenCost dashboards) is explicitly deferred to M3/M4.

### 4.6 OPA/Gatekeeper (pipeline-scoped)

- **FR-23** Gatekeeper SHALL be installed in the cluster in the `gatekeeper-system` namespace at a pinned version.
- **FR-24** Three and only three ConstraintTemplates SHALL be defined in M2:
  - **C-1:** `platform-addons` repo `main` branch SHALL reject direct commits. All changes SHALL go through a merged PR.
  - **C-2:** Any workload manifest emitted by the pipeline SHALL pass Score schema validation before `kubectl apply` or equivalent.
  - **C-3:** An Infracost breakdown artifact SHALL exist in the pipeline workspace and its monthly delta SHALL be below the configured threshold (FR-21).
- **FR-25** All three constraints SHALL fire as part of the Gitea Actions pipeline. Runtime pod-admission constraints are explicitly deferred to M6.
- **FR-26** Gatekeeper's audit controller SHALL log any constraint violation to cluster logs only; no external alerting in M2.

### 4.7 Environment promotion

- **FR-27** Two OpenChoreo Environments SHALL exist: `dev` and `staging`. Both live in the M1 cluster, separated by OpenChoreo-managed namespaces.
- **FR-28** A commit to an application repo's `main` branch SHALL deploy to `dev` automatically.
- **FR-29** Promotion from `dev` to `staging` SHALL be a commit to the `platform-config` config repo (see Section 4.8) that updates the Component's `targetEnvironment` field. No automatic promotion in M2.
- **FR-30** A rollback SHALL be a revert commit to `platform-config`. No bespoke rollback tooling.

### 4.8 Repository topology

- **FR-31** Three Gitea repositories SHALL exist after M2 install:
  - `platform-addons` -- watched by Flux, contains Kustomization for cluster add-ons
  - `platform-config` -- watched by OpenChoreo, contains rendered OpenChoreo manifests per environment
  - `hello-m2` -- the demo application repo used to validate the M2 pipeline end to end
- **FR-32** `platform-addons` and `platform-config` SHALL be owned by an `openchoreo` Gitea organization, not by `gitea_admin`.
- **FR-33** A Gitea webhook or poll-based sync SHALL fire a CI workflow on push to any of the three M2 repos.

### 4.9 Local OCI Registry

- **FR-34** The in-cluster `local-registry` SHALL be used to store container images built from the demo application.
- **FR-35** A Gitea Actions workflow SHALL build, tag (git SHA), and push the demo image to the local registry on merge to `main`.
- **FR-36** OpenChoreo manifests in `platform-config` SHALL reference images by `registry.local-registry.svc.cluster.local:5000/hello-m2:<sha>` (cluster-internal form) and SHALL NOT pull from external registries for M2.

### 4.10 openbao integration

- **FR-37** openbao (already running from M1) SHALL be the source of truth for:
  - Gitea Actions runner registration token (kv v2 path `kv/gitea/runners/token`)
  - Demo application's example secret (OpenChoreo default-store kv v2 path `secret/apps/hello-m2/dev/example-secret`; `kv/apps/hello-m2/dev/example-secret` is seeded as a compatibility mirror)
  - Flux Gitea deploy key (kv v2 path `kv/flux/gitea-deploy-key`)
- **FR-38** All Kubernetes Secret objects consumed in M2 SHALL be materialized by external-secrets from openbao. No `kubectl create secret` in the install path except for the initial bootstrap of external-secrets' own auth against openbao (already present from M1).
- **FR-39** No secret value, token, or key SHALL appear in any committed file, spec document, or log line.

### 4.11 Backstage surfacing

- **FR-40** Backstage (from M1) SHALL continue to discover repos from Gitea and SHALL pick up `hello-m2` via its `catalog-info.yaml`.
- **FR-41** Backstage SHALL NOT receive a custom M2-specific plugin in this milestone. M2 adds one proxy endpoint only: `/api/proxy/gitea-actions` targeting Gitea's actions API so a future plugin (M3+) can enumerate workflow runs.
- **FR-42** A `links` section SHALL be added to `hello-m2`'s `catalog-info.yaml` pointing at: Gitea repo UI, Gitea Actions last run, and Infracost artifact URL. These render as clickable links in Backstage's component view without any plugin work.

### 4.12 Per-tool smoke checks

For operator confidence and debuggability, every M2 tool SHALL have an individually runnable smoke check. These mirror the user's request that M2 deliver option (c) end-to-end plus option (a)'s per-tool individual checks.

- **FR-43** `scripts/smoke-tofu.sh` SHALL run `tofu version`, `tofu init` against `iac/`, and `tofu plan -detailed-exitcode` against `iac/`, reporting pass or fail.
- **FR-44** `scripts/smoke-actions.sh` SHALL trigger the `hello-world` workflow in `hello-m2` and wait for its success via Gitea's API.
- **FR-45** `scripts/smoke-flux.sh` SHALL run `flux reconcile source git flux-system` and `flux get kustomizations` and report all ready.
- **FR-46** `scripts/smoke-score.sh` SHALL run `score2openchoreo < fixtures/score.yaml` and validate the output against OpenChoreo's Component CRD schema.
- **FR-47** `scripts/smoke-infracost.sh` SHALL run `infracost breakdown --path iac/` and report a numeric monthly figure.
- **FR-48** `scripts/smoke-gatekeeper.sh` SHALL run `kubectl get constrainttemplates` and `kubectl get constraints` and report counts.
- **FR-49** `scripts/smoke-openbao.sh` SHALL confirm the kv paths listed in FR-37, including the demo app secret mirror, are reachable and populated.
- **FR-50** A single orchestrator `scripts/smoke-m2.sh` SHALL invoke every per-tool smoke check in sequence and exit 0 only if every one passes.

### 4.13 Operational commands

- **FR-51** A single `scripts/install-m2.sh` SHALL provision the M2 delta on top of an M1-healthy cluster. It SHALL be idempotent and safe to re-run.
- **FR-52** A single `scripts/teardown-m2.sh` SHALL remove M2 add-ons (Flux, Gatekeeper, Gitea Actions runner) and seed repos, leaving the M1 substrate healthy. Tearing down M2 SHALL NOT tear down M1.
- **FR-53** `README.md` at the repo root SHALL be updated with an M2 section in under 150 additional words.

## 5. Non-Functional Requirements

### 5.1 Performance

- **NFR-1** Cold-start time from an M1-healthy cluster to an M2-green smoke test (`scripts/smoke-m2.sh` exit 0) SHALL NOT exceed 20 minutes.
- **NFR-2** A PR CI run on a change to `hello-m2` SHALL complete in under 5 minutes on the operator's Apple Silicon laptop.
- **NFR-3** Flux reconciliation interval SHALL be 1 minute for `platform-addons`; slower intervals add no value at M2 scale.

### 5.2 Determinism

- **NFR-4** New tools created in M2 (the `score2openchoreo` converter, any helper shim) SHALL be written in Go with stdlib plus explicitly pinned dependencies only.
- **NFR-5** OpenTofu module code SHALL be declarative HCL. Any computed logic that cannot be expressed declaratively SHALL be isolated into a Go helper invoked via `external` provider and explicitly called out.
- **NFR-6** Every install and teardown script SHALL be idempotent.

### 5.3 Security

- **NFR-7** No secret value SHALL be committed. External-secrets is the only path from openbao to a Kubernetes Secret.
- **NFR-8** The Gitea Actions runner SHALL run as a non-root user in its pod and have no privileged mounts.
- **NFR-9** OpenTofu state SHALL never be written to local disk outside `/tmp` in a runner pod. All persistence goes to openbao.
- **NFR-10** The pipeline SHALL reject any workflow that uses a third-party action not on an explicit allow-list. The allow-list starts empty; adding an entry is a code review.
- **NFR-11** A new PreToolUse hook in `plugins/rr-policy-guards/` SHALL reject `tofu apply` when invoked as a Bash tool use outside of a runner context. This hook SHALL be named `rr-tofu-guard`. It SHALL NOT reject `tofu plan`, `tofu init`, or `tofu validate`.

### 5.4 Observability

- **NFR-12** M2 SHALL emit no traces or metrics by default (M3's job). Logs produced by Flux, Gatekeeper, and the runner SHALL be inspectable via `kubectl logs`.
- **NFR-13** The `rr-tofu-guard` hook SHALL append every block to `~/.rational-reserve/logs/tofu-guard.jsonl` in the JSONL format established by the other guards.

### 5.5 Portability

- **NFR-14** M2 SHALL target Apple Silicon macOS with Colima, matching M1. Linux support is not actively excluded by design but is not tested.

### 5.6 Documentation

- **NFR-15** Every Go source file SHALL carry a file-level comment under 10 lines.
- **NFR-16** Every OpenTofu module SHALL have a `README.md` explaining inputs, outputs, and providers.
- **NFR-17** The three M2 spec documents SHALL be kept in sync during implementation, matching M1's discipline.

## 6. Prerequisites (required before M2 begins)

| Prerequisite | Minimum version | Source | Installer if missing |
|---|---|---|---|
| M1 substrate healthy | current SHA | M1 install script |  verified 2026-04-20 |
| OpenTofu | 1.8+ | host | `brew install opentofu` (gated by brew-guard) |
| Infracost | 0.10+ | host | `brew install infracost` (gated by brew-guard) |
| Score CLI (`score-k8s` or equivalent) | latest | host | `brew install score-spec/tap/score-k8s` (gated by brew-guard) -- tap add requires allow-list update |
| Flux CLI | 2.3+ | host | `brew install fluxcd/tap/flux` (gated by brew-guard) -- tap add requires allow-list update |
| Go | 1.21+ | host |  from M1 |

Three of the installs above require adding a new entry to the brew-guard tap allow-list, which is an explicit code change reviewable in this M2 work. Expand the allow-list in a single commit before running `install-m2.sh` for the first time.

## 7. Acceptance Criteria

M2 is complete when **all** of the following are true:

- [ ] `rr-tofu-guard` binary builds, all unit tests pass, hook registered in `~/.claude/settings.json`
- [ ] `tofu apply` from a direct Bash tool use is rejected by the hook; `tofu plan` is allowed
- [ ] `iac/` module applies cleanly; Flux, Gatekeeper, and the Gitea Actions runner are all Ready
- [ ] `flux get kustomizations -A` shows `platform-addons` Ready and SyncedRevision matches `platform-addons` HEAD
- [ ] `kubectl get pods -n gatekeeper-system` shows all Ready
- [ ] `kubectl get constrainttemplates` lists the three C-1/C-2/C-3 templates; `kubectl get constraints` lists three constraints; audit shows them Ready
- [ ] `kubectl get pods -n gitea-runners` shows the runner Ready and registered against Gitea
- [ ] A push to `hello-m2`'s `main` triggers a pipeline that: checks out, validates Score, runs `tofu plan`, runs Infracost, posts the Infracost comment, builds the image, pushes to local-registry, renders OpenChoreo resources, commits to `platform-config`
- [ ] A merged PR on `platform-config` that moves `hello-m2` from `dev` to `staging` results in a `hello-m2` Deployment running in the `staging` OpenChoreo Environment
- [ ] Backstage shows `hello-m2` with links to Gitea repo, Gitea Actions last run, and the Infracost artifact
- [ ] `scripts/smoke-m2.sh` exits 0
- [ ] `scripts/teardown-m2.sh` leaves M1 healthy; re-running `scripts/smoke-test-m1.sh` still exits 0
- [ ] README has an M2 section under 150 words
- [ ] All three M2 spec documents are present, checked in, and internally consistent

## 8. Out of Scope (explicitly deferred)

| Item | Deferred to |
|---|---|
| OpenTelemetry Collector and SigNoz instrumentation | M3 |
| OpenCost runtime cost attribution | M4 |
| Cloud Custodian | M6 |
| Cilium CNI | M4 |
| Envoy Gateway as primary gateway (currently kgateway via OpenChoreo) | M4 |
| RabbitMQ or Kafka with OpenResty front-door | M5 |
| MISP, TheHive, Cortex, Velociraptor | M6 |
| Broader OPA/Gatekeeper runtime policies (pod security, image allow-list, network) | M6 |
| MCP plugin surfacing the M2 pipeline to AI agents | M7 |
| Per-agent Gitea tokens and agent-aware policy | M7 |
| Rational Reserve to OpenChoreo wiring | M7 |
| Automatic dev-to-staging promotion | Never by default (explicit commit always required) |
| Production clusters, TLS, real DNS | Never at M2 (local dev only) |
| Linux and Intel macOS support | Revisit when team membership changes |

## 9. Assumptions

- **A-1** M1 is healthy at M2 install time. `scripts/smoke-test-m1.sh` exits 0.
- **A-2** openbao is running, unsealed, and has a policy that permits external-secrets to read under the paths in FR-37.
- **A-3** Gitea's helm values permit the `admin` user to create an organization (default true in Gitea chart 12.5).
- **A-4** The `openchoreo` organization in Gitea is owned by `gitea_admin` or a service account with org-create rights.
- **A-5** Score CLI and Flux CLI are installable via brew taps that the operator will add to brew-guard's allow-list before running install-m2.
- **A-6** The operator runs `scripts/install-m2.sh` from a Claude Code session where the hooks (emoji, bash, brew, and the new tofu guard) are active.
- **A-7** Colima is sized per M1's NFR-2 (8 GB RAM, 4 CPU, 40 GB disk). M2 adds approximately 300 MB for Flux, 200 MB for Gatekeeper, and 150 MB for the runner; still well inside budget.

## 10. Glossary

- **IaC** -- Infrastructure as Code (OpenTofu in this project)
- **CD loop** -- Continuous Delivery loop: commit -> CI -> cost/policy gate -> merge -> render -> deploy
- **Score** -- ScoreSpec workload specification language used as the authoring surface for workloads
- **score2openchoreo** -- Go converter this milestone introduces; reads Score and emits OpenChoreo Component, SecretReference, and Workload CRDs
- **Flux** -- GitOps controller used in M2 only for cluster add-ons drift correction
- **Gatekeeper** -- The OPA-based admission controller; in M2 used only for pipeline-scoped constraints
- **platform-addons** -- Gitea repo Flux watches
- **platform-config** -- Gitea repo OpenChoreo watches (rendered OpenChoreo resources)
- **hello-m2** -- the demo application repo used to prove the pipeline end to end
- **Environment** -- OpenChoreo's first-class concept for isolated deployment targets; in M2 these are `dev` and `staging`
- **rr-tofu-guard** -- the new PreToolUse hook this milestone introduces to block direct `tofu apply`
