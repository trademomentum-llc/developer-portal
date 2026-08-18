# Requirements and Roadmap: Five-Plane Collaborative Portal

**Document ID:** FIVE-PLANE-PORTAL-REQ-001
**Version:** 0.1
**Date:** 2026-08-18
**Status:** Draft for review -- nothing in this document is approved or decided yet
**Predecessors:** 2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md, 2026-06-30-M4-Cost-Visibility-Requirements.md, docs/specs/m2-iac-cd/requirements.md, TODO.md (M5-M7 deferred draft rows, Option C notes)

---

## 1. Purpose and relationship to existing milestone docs

This document defines the requirements and a phased roadmap for evolving the developer portal into a **web-based collaborative platform** in which a team hosts projects and traverses five planes within each project:

1. **Observation** (telemetry)
2. **Control** (project files)
3. **Orchestration** (toolsets)
4. **Security** (threat intelligence)
5. **Engagement** (execution testing)

This is an **umbrella requirements document**. It consolidates the verified plane maps produced on 2026-08-18 (five evidence-backed surveys of the repo: existing assets, verified gaps, traversal touchpoints, candidate fills, open questions) into one requirements baseline. It does not replace the milestone documents:

- M1-M4 requirements/design/technical triads remain the authoritative record for what is already built and smoke-validated.
- M5-M7 remain deferred one-line draft rows in `TODO.md:307-315`; this document does not ratify them.
- Per governance (`~/Projects/Sovereign/Structure/POLICIES.md` triad rule), any per-plane implementation work approved out of this umbrella gets its own Requirements, Design Specification, and Technical Specification before implementation begins.

A companion workstream -- **record immutability for project documents** -- is planned separately (TODO.md goal directive, slice 2). It is referenced here only as a dependency/touchpoint (see NFR-05, FR-40, OQ-10); it is not specified in this document.

**Evidence discipline:** every current-state claim in section 4 cites the repo path verified in the plane maps. Items the maps marked UNCERTAIN are carried forward as UNCERTAIN. Nothing in this document asserts live status that repo evidence does not support.

---

## 2. Goal and rationale

**Goal:** stand the portal up as the single web surface through which a team hosts a project and moves between its five planes -- observing telemetry, controlling project files, operating the toolsets that build and run the project, consuming security/threat-intelligence signals, and executing tests -- without leaving the portal context.

**Rationale (factual):**

- Today the portal is mostly a link farm: only cost telemetry renders live inside the UI (`CostCard.tsx:36-63`); every other plane is outbound links, some of them stale or dead (section 5). Each broken or absent link forces a context switch to a terminal or a separate UI, and each context switch costs the team the correlation keys (component, environment, git SHA, predicted namespace) the platform already computes.
- The correlation substrate already exists and is validated: trace resource attributes carry `openchoreo.project/component/environment/runtime_namespace` and `git.commit.sha` (`seed-repos/hello-m2/main.go:94-104`), and the namespace predictor is byte-equivalent to OpenChoreo name generation (`tools/namespace-predictor/main.go`, `namespace-predictor.ts`). The planes are instrumented at the data level; they are disconnected at the surface level.
- Smooth traversal supports better work discipline: when status is visible, live, and honestly labeled, a team checks the portal instead of asking around or re-running commands, and handoffs between roles (author, operator, reviewer) carry the same shared context. Fewer hunts for state, fewer stale assumptions, less rework -- that is the productivity mechanism. This document makes no claims beyond that.
- Collaboration is currently single-operator: `backstage/examples/org.yaml` defines two users, Gitea self-registration is disabled (`scripts/gitea-values.yaml:14`), and the agent-side preventive guards are registered per-user outside the repo (`plugins/rr-policy-guards/README.md:78-82`). A team cannot share the platform until identity, ownership, and guard distribution scale beyond one machine.

---

## 3. The five planes as scoped for this project

**Observation (telemetry).** All signals that describe what a project and the platform are doing: traces, metrics, logs, and cost allocation for workloads, plus telemetry about the platform components themselves. For this project the plane is anchored on SigNoz + the standalone OTEL collector (traces/metrics/logs pipeline) and OpenCost + Prometheus (cost), with Backstage entity tabs and cards as the in-portal surface and SigNoz/OpenCost UIs (via Envoy Gateway `.local` routes or port-forwards) as the external surface. Alerting, dashboards, and log collection are part of this plane's scope but are currently absent.

**Control (project files).** The authoritative content of a project and the means to create and change it: Gitea repositories (application code, Score descriptors, catalog entities, docs), the GitOps repos (`platform-config`, `platform-addons`) that record intended state per environment, the Backstage software catalog that models and discovers those files, TechDocs, and the scaffolder that should mint new projects. Ownership modeling (users/groups) belongs to this plane because it determines who may control what.

**Orchestration (toolsets).** The machinery that turns controlled files into running workloads and keeps them there: OpenChoreo Component/Workload reconciliation on the k3d-openchoreo cluster, the score2openchoreo renderer, Gitea Actions CI, Flux drift correction, the OpenTofu lifecycle (guarded by rr-tofu-guard), OpenBao/external-secrets wiring, Envoy Gateway networking, and the dev/staging environment model including promotion. The portal's role is visibility and -- where explicitly decided -- actuation; today it is read-only by design (`docs/specs/m2-iac-cd/design-specification.md:194`).

**Security (threat intelligence).** Two layers, honestly separated: (a) the preventive and detective controls that exist today -- the six rr-policy-guards, Gatekeeper pipeline constraints C1-C3, publication SCA gates (Semgrep, Gitleaks, yarn/npm audit, govulncheck), Gitea OAuth, the Backstage permission framework, OpenBao secrets; and (b) the plane's namesake capability, threat intelligence -- feeds, a threat-intel platform, vulnerability awareness, detection, incident response -- which has **zero implementation** today and exists only as deferred M6 draft rows (`TODO.md:314`). This document scopes the plane as both layers but treats layer (b) as gated on explicit decisions (OQ-19 through OQ-22).

**Engagement (execution testing).** Executing validation against a project or the platform and observing the results: the smoke suites (`scripts/smoke-*.sh`, `smoke-all.sh`), Playwright e2e, Go and Rego unit tests, CI validation steps, and the verify-guard CI-equivalent gate. The plane covers triggering tests, machine-readable results, per-project test surfaces, and -- pending capacity decisions -- load and failure testing. Today it is entirely terminal-side: no portal surface exists for this plane.

---

## 4. Current state per plane (verified)

Maturity scale: **live** (deployed and smoke-validated), **partial** (deployed but degraded, incomplete, or unverified in part), **scaffold** (installed/registered but not functional), **absent** (verified not present).

### 4.1 Observation plane -- assets

| Capability | Asset | Evidence (repo path) | Maturity |
|---|---|---|---|
| Trace/metric/log backend (SigNoz, chart 0.130.1, ns `signoz`) | OpenTofu helm_release + values | `iac/modules/observability/main.tf:17-28`, `observability/signoz/values.local.yaml` | live -- demo-grade: ClickHouse persistence disabled (emptyDir, `values.local.yaml:24-26`), sampling 0.1, retention traces 3d / logs 3d / metrics 7d |
| OTLP ingestion tier (standalone collector 0.155.0, ns `otel-system`) | helm_release; traces+metrics+logs pipelines to `signoz-otel-collector:4317` | `iac/modules/observability/main.tf:30-41`, `observability/otel/collector-values.local.yaml:56-77` | live |
| SigNoz collector OTLP enablement (OpAMP arg patch) | null_resource local-exec kubectl patch | `iac/modules/observability/main.tf:46-60` | live |
| Workload trace instrumentation (hello-m2) | OTel SDK, OTLP/HTTP exporter, `openchoreo.*` + `git.commit.sha` resource attrs | `seed-repos/hello-m2/main.go:69-120` | live -- traces only |
| CI telemetry injection | `--extra-env OTEL_EXPORTER_OTLP_ENDPOINT=...otel-system...:4318` | `seed-repos/hello-m2/.gitea/workflows/ci.yaml:91` | live |
| Live trace-ingestion assertion | request generation + ClickHouse query for `serviceName='hello-m2'` | `scripts/smoke-m3.sh:337-374` | live (22/22 per `SESSION_HANDOFF.md:65`) |
| Cost telemetry (OpenCost 2.5.25 + Prometheus 29.13.0, ns `opencost`) | OpenTofu module + values | `iac/modules/cost/main.tf`, `observability/cost/*.yaml` | live (`scripts/smoke-m4.sh` passes) |
| In-portal live cost telemetry | proxy `/api/proxy/opencost`; CostCard fetches live allocation per predicted namespace | `backstage/app-config.yaml:70-74`, `openchoreo-cards/CostCard.tsx:36-63`, `scripts/start-backstage.sh:63-77` | live -- the only telemetry rendered inside the portal UI |
| Pre/post-deploy cost artifacts | Infracost artifact committed per push; linked from CostCard | `scripts/ci/commit-cost-artifact.sh`, `scripts/smoke-m3.sh:376-390`, `CostCard.tsx:29-30,90-98` | live |
| Observability + Cost entity tabs | EntityContentBlueprint tabs registered in app | `openchoreo-entity-page/index.tsx:39-73`, `backstage/packages/app/src/App.tsx:5,13` | live (Playwright-verified per `SESSION_HANDOFF.md:149`) |
| SigNoz deep links from entity pages | Traces/Metrics/Logs links; dashboards link | `ObservabilityLinksCard.tsx:21,35-43`, `DeploymentCard.tsx:61-63`, `PlatformCard.tsx:30-32` | partial -- hardcode `http://localhost:8080` (see OBS-G6) |
| External telemetry UI access | Envoy HTTPRoutes `signoz.local`, `opencost.local`, `gitea.local` | `iac/modules/networking/variables.tf:47-72`, `scripts/smoke-m4-networking.sh` (3/3) | live -- requires /etc/hosts entry + NodePort (`scripts/update-local-hosts.sh`) |
| Telemetry-to-workload correlation key | namespace predictor (Go + TS ports), canonical vector `dp-default-default-development-f8e58905` | `tools/namespace-predictor/main.go`, `openchoreo-cards/namespace-predictor.ts` | live |
| Lifecycle scripts | install/teardown/preflight | `scripts/install-m3.sh`, `scripts/teardown-m3.sh`, `scripts/preflight-m3.sh`, `scripts/install-m4.sh`, `scripts/teardown-m4.sh` | live |

### 4.1b Observation plane -- verified gaps

| ID | Gap | Evidence |
|---|---|---|
| OBS-G1 | No log aggregation: collector has only an OTLP receiver (no filelog/k8s log receiver); hello-m2 writes stdout logs only; the "Logs in SigNoz" link lands on an empty view | `observability/otel/collector-values.local.yaml:30-77`; `seed-repos/hello-m2/main.go`; repo-wide grep for loki/fluent-bit/filelog finds only the unaligned `docs/specs/Engineering-Spec.md:19` draft |
| OBS-G2 | No alerting of any kind: alertmanager explicitly disabled in both values files; named requirement FR-PROD-2 in the M3 requirements is unimplemented | `observability/signoz/values.local.yaml:47-48`, `observability/cost/prometheus-values.local.yaml:10-11`, `docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md:72` |
| OBS-G3 | No workload metrics: hello-m2 imports only the trace SDK; no ServiceMonitor/PodMonitor anywhere; the collector metrics pipeline receives nothing | `seed-repos/hello-m2/main.go` imports; repo-wide grep ServiceMonitor/PodMonitor zero hits |
| OBS-G4 | No dashboards/SLO surfaces: `observability/dashboards/` is named in the M3 technical specification but the directory does not exist and nothing seeds it | `docs/specs/2026-05-28-M3-...-Technical-Specification.md:26,114`; `find observability -type f` returns only 4 values files |
| OBS-G5 | No in-portal trace/log/metric rendering: proxy endpoints are only `/gitea-actions` and `/opencost`; no SigNoz proxy, embed, or API client | `backstage/app-config.yaml:61-74`; grep across `backstage/` |
| OBS-G6 | SigNoz link target broken: cards hardcode `http://localhost:8080`, which no managed port-forward provides; AGENTS.md documents 3301; `signoz.local` exists via Envoy. Three inconsistent access paths. UNCERTAIN whether a manual 8080 forward ever existed | `ObservabilityLinksCard.tsx:21`, `DeploymentCard.tsx:61`, `PlatformCard.tsx:30`, `scripts/start-backstage.sh:50-77` |
| OBS-G7 | No telemetry durability: ClickHouse emptyDir; Prometheus has no PV and 2d retention; history dies with pod restart | `observability/signoz/values.local.yaml:24-26`, `observability/cost/prometheus-values.local.yaml:2-4` |
| OBS-G8 | No platform-component telemetry: only hello-m2 is instrumented; nothing in Backstage, Flux, Gatekeeper, or CI steps emits telemetry | repo-wide grep `OTEL_EXPORTER_OTLP_ENDPOINT` hits only hello-m2, its CI, and docs |

### 4.2 Control plane -- assets

| Capability | Asset | Evidence (repo path) | Maturity |
|---|---|---|---|
| Self-hosted git hosting (repos, PRs, web editing) | Gitea, PostgreSQL-backed, `DISABLE_REGISTRATION: true`; port-forwards 3333/3002 | `scripts/gitea-values.yaml`, `scripts/start-backstage.sh` | live |
| Org + repo provisioning (idempotent), branch protection | `openchoreo` org; `platform-addons`, `platform-config`, `hello-m2` | `scripts/seed-gitea-repos.sh:18-44`, `scripts/push-seed-content.sh` | live |
| CI for project repos | Gitea Actions act-runner; full push-to-deploy pipeline | `seed-repos/hello-m2/.gitea/workflows/ci.yaml` | live (run #30 per SESSION_HANDOFF.md) |
| Catalog auto-discovery of Gitea repos | `@backstage/plugin-catalog-backend-module-gitea`, org `openchoreo`, 30-min schedule | `backstage/app-config.yaml:100-109`, `packages/backend/src/index.ts:38-40` | live (smoke-m3.sh asserts both entities) |
| Entity modeling | Domain/System/Component with `openchoreo.dev/*` annotations | `catalog-info.yaml`, `seed-repos/hello-m2/catalog-info.yaml` | live, minimal scope |
| Ownership modeling | static org.yaml: 2 Users, 3 Groups | `backstage/examples/org.yaml`, `app-config.yaml:117-120` | partial (single real operator) |
| Score as versioned workload descriptor | score.yaml + Go converter with schema validation and fixtures | `seed-repos/hello-m2/score.yaml`, `tools/score2openchoreo/` | live |
| GitOps project-config repos; promotion-as-commit | `platform-config/environments/{dev,staging}`, `platform-addons/clusters/default`; CI write-back via contents API | `seed-repos/platform-config/`, `seed-repos/platform-addons/`, `scripts/ci/commit-to-platform-config.sh`, `scripts/ci/commit-cost-artifact.sh` | live |
| Scaffolder installed (backend + frontend + nav) | plugin-scaffolder-backend + notifications module | `packages/backend/src/index.ts:16-21,35-37`, `app/package.json:33`, `nav/Sidebar.tsx:36` | scaffold (see CTL-G1) |
| One example Template entity | stock Node.js example | `backstage/examples/template/template.yaml` | scaffold -- targets github.com, will fail at runtime |
| TechDocs installed + versioned docs | techdocs plugins; `mkdocs.yml` + `docs/`; `techdocs-ref: dir:.` on root entities | `packages/backend/src/index.ts:23-24`, `app-config.yaml:80-85`, `catalog-info.yaml:25,44` | partial (docker-based generation not live-verified) |
| Identity: Gitea OAuth mapped to catalog ownership | login maps to `user:default/<login>` + `group:default/openchoreo` | `packages/backend/src/modules/giteaAuth.ts:187-191`, `scripts/smoke-auth.sh` | live |
| Portal repo mirror Gitea -> GitHub | scheduled sync workflow | `.github/workflows/sync-from-gitea.yml` | live (gitea.com cloud auth blocked per TODO.md; local origin works) |

### 4.2b Control plane -- verified gaps

| ID | Gap | Evidence |
|---|---|---|
| CTL-G1 | No scaffolder publish-to-Gitea path: no `publish:gitea` action anywhere; the github module is present in package.json but deliberately not registered; the one template uses `publish:github` with `allowedHosts: github.com` and will fail. A team cannot create a new project repo from the portal today | `packages/backend/src/index.ts:18`, `packages/backend/package.json:38`, `examples/template/template.yaml:54-59`, repo-wide grep |
| CTL-G2 | No paved-path project template for this stack (Score + ci.yaml + catalog-info + Dockerfile); only one `kind: Template` exists in the repo | repo-wide grep `kind: Template` |
| CTL-G3 | No in-portal file browsing/editing; portal bounces users to the Gitea UI via entity links | `backstage/packages/app/package.json`, `packages/backend/package.json` (no editor/file-tree plugin) |
| CTL-G4 | No multi-user/team provisioning: self-registration disabled, no admin user/team scripts, org.yaml has 2 users | `scripts/gitea-values.yaml:14`, grep `admin/users|teams` across `scripts/` zero hits |
| CTL-G5 | Dangling ownership refs: root `catalog-info.yaml` uses `owner: user:default/nnos` (lines 14, 32, 56, 77, 97) but no such User entity exists | `catalog-info.yaml`, `backstage/examples/org.yaml` |
| CTL-G6 | Record immutability for project documents absent: no signed-commit enforcement, no append-only store; only the provenance-certificate digest pattern exists. (Owned by the companion workstream -- touchpoint only.) | repo-wide grep; `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md`; TODO.md goal slice 2 |
| CTL-G7 | No Gitea Actions CI for the developer-portal umbrella repo itself | root `.gitea/` does not exist; only `.github/workflows/sync-from-gitea.yml` |
| CTL-G8 | GitOps repos invisible in catalog: provider only ingests `catalog-info.yaml`; `platform-config` and `platform-addons` have none | `app-config.yaml:101-106`; `seed-repos/platform-config/`, `seed-repos/platform-addons/` |
| CTL-G9 | No ownership sync Gitea<->Backstage<->OpenChoreo: org.yaml is static; the gitea module ingests entities, not users/groups | `backstage/examples/org.yaml`; TODO.md Option C problem 4 |
| CTL-G10 | hello-m2 has no `backstage.io/techdocs-ref` annotation; per-project docs not surfaced | `seed-repos/hello-m2/catalog-info.yaml` |

### 4.3 Orchestration plane -- assets

| Capability | Asset | Evidence (repo path) | Maturity |
|---|---|---|---|
| Workload orchestration (Component/Workload reconcile) | OpenChoreo on k3d-openchoreo (sibling repo), driven from this repo | `tools/score2openchoreo/{types.go,convert.go,cli.go}`; live proof `TODO.md:140` (run #24), `SESSION_HANDOFF.md:54-60` (run #27, pod 1/1 Running) | live |
| Score validation + render to OpenChoreo CRDs | score2openchoreo Go CLI | `tools/score2openchoreo/schema.go`, `assets/score.schema.json`, `cli.go`; `scripts/smoke-score.sh` | live -- fidelity partial (no Traits/Workflows; grep `workflow|trait` in `tools/score2openchoreo` zero hits) |
| Environment model (dev, staging) | OpenTofu module emitting two Environment CRDs | `iac/modules/openchoreo-environments/main.tf` | live -- only dev is exercised by any pipeline |
| GitOps apply + drift correction | Flux 2.13.0, GitRepository x2, Kustomization x3 (incl. staging) | `iac/modules/flux/main.tf`; run #24 per `TODO.md:128` | live |
| CI execution | Gitea Actions act-runner; runner token from OpenBao via ExternalSecret | `iac/modules/gitea-runner/main.tf` | live (runs #24/#27/#30) |
| CI pipeline definition | live workflow + canonical template | `seed-repos/hello-m2/.gitea/workflows/ci.yaml`, `iac/templates/ci.yaml` | live -- drift: template uses `--environment dev`, live workflow uses `development` (`ci.yaml:89,94`) |
| Image registry | in-cluster local-registry + k3s mirror | `iac/modules/local-registry/main.tf` | live |
| Guarded IaC lifecycle | OpenTofu root composition + lifecycle scripts + rr-tofu-guard | `iac/main.tf`, `scripts/install-*.sh`, `scripts/teardown-*.sh`, `plugins/rr-policy-guards/tools/tofu-guard/` | live |
| Secrets delivery | ClusterSecretStore `openbao-kv` + ExternalSecrets | `iac/modules/external-secrets-wiring/main.tf` | live |
| Platform ingress | Envoy Gateway 1.3.1, HTTPRoutes for `gitea.local`, `signoz.local`, `opencost.local` | `iac/modules/networking/{main.tf,gateway/httproutes.tf,variables.tf}` | live (`scripts/smoke-m4-networking.sh` 3/3) |
| Promotion mechanism (commit-based) | env-parameterized commit script; design ratified in M2 spec | `scripts/ci/commit-to-platform-config.sh`; `docs/specs/m2-iac-cd/design-specification.md` sec 4.9/5.4 | partial -- script supports staging, nothing ever calls it with staging |
| Portal Deployment tab | entity content tab + card with predicted namespace and links | `openchoreo-entity-page/index.tsx:21-37`, `openchoreo-cards/DeploymentCard.tsx` | live -- static/link-based, no live status; some targets stale (ORC-G8) |
| Portal Platform tab | platform dependency links card | `openchoreo-cards/PlatformCard.tsx` | live -- link-based only |
| Portal Overview tab (OpenChoreo context card) | annotation-driven card on the default Overview tab via EntityCardBlueprint: project/component/environment, control-plane NS, predicted runtime NS, "View in OpenChoreo" link | `openchoreo-cards/OpenChoreoOverviewCard.tsx:8-47`, registered in `openchoreo-cards/index.tsx:16-33` | live -- annotation-driven, static + link-based; api-base link target subject to TRV-B5 |
| Namespace predictor | Go CLI + TS port, byte-equivalent to OpenChoreo namegen | `tools/namespace-predictor/main.go`, `openchoreo-cards/namespace-predictor.ts` | live |
| CI-run API proxy | `/api/proxy/gitea-actions` -> Gitea API | `backstage/app-config.yaml:63-69` | scaffold -- zero frontend consumers |
| Kubernetes plugin | frontend+backend deps; backend registered | `packages/backend/src/index.ts:64`, `app-config.yaml:122` | scaffold -- `kubernetes:` block empty, no cluster locator, zero frontend usage |
| Argo Workflows | runs in sibling cluster's `openchoreo-workflow-plane` (OpenChoreo-internal) | `provenance/PROVENANCE.md:369` (U13: "zero references in this repo"), `AGENTS.md:10` | present-in-cluster, unintegrated |

### 4.3b Orchestration plane -- verified gaps

| ID | Gap | Evidence |
|---|---|---|
| ORC-G1 | No in-portal CI/CD run visibility: the proxy is configured but nothing consumes it; FR-41 (`docs/specs/m2-iac-cd/requirements.md:119`) deferred this to "M3+" and it was never built | grep `gitea-actions` in `backstage/packages/app/src` zero hits |
| ORC-G2 | Kubernetes plugin installed but unconfigured: no cluster locator, no EntityKubernetesContent mounted; DeploymentCard predicts a namespace but cannot show what is actually there | `app-config.yaml:122-123`; grep `kubernetes` in `packages/app/src` only comments |
| ORC-G3 | No promotion execution path: CI hardcodes `development`; staging environment and its Flux kustomization are dead infrastructure; promotion is an undocumented manual copy+commit | grep `staging` in `scripts/` zero matches; `ci.yaml:89,94` |
| ORC-G4 | No workflow orchestration: no Workflow/WorkflowRun/Trait CRDs anywhere; score2openchoreo cannot render them; the workflow-plane sits idle | repo-wide greps; gap named in `TODO.md:283` |
| ORC-G5 | No live OpenChoreo status in catalog: cards read static annotations only; ReleaseBinding/ComponentRelease status unreachable except as raw API JSON | `DeploymentCard.tsx:20-25`; `TODO.md:286` ("missing Backstage OpenChoreo plugin surface") |
| ORC-G6 | No Gitea-aware scaffolding (cross-reference CTL-G1/CTL-G2): the self-service entry point points at the wrong forge | `packages/backend/src/index.ts:18`, `examples/template/template.yaml` |
| ORC-G7 | No fleet/multi-component orchestration board: per-entity cards only; no cross-project view of deployments, runs, or environments | review of `openchoreo-cards/` and `openchoreo-entity-page/` |
| ORC-G8 | Stale/broken deep links: `localhost:9090` (OpenChoreo API) forwarded only by `scripts/install-m1.sh:197` with no smoke coverage (reachability UNCERTAIN); `localhost:8080` served by nothing; `/iac/environments/...` is not a Backstage route | `DeploymentCard.tsx:58,61`, `ObservabilityLinksCard.tsx:21`, `PlatformCard.tsx:30`, `scripts/start-backstage.sh:50-77` |
| ORC-G9 | Canonical CI file drift: `iac/templates/ci.yaml` (stale, `dev`) vs live `seed-repos/hello-m2/.gitea/workflows/ci.yaml` (`development`, OTEL + cost artifact) | both files read in full |

### 4.4 Security plane -- assets

**Headline finding: the plane's namesake capability -- threat intelligence -- has zero implementation today.** What exists is real but almost entirely preventive, local, and pipeline-scoped. There is no detection, no runtime security, no image scanning, no SIEM, no incident response, and no security surface in the portal UI beyond a static link card. Everything M6 is a one-line deferred draft row (`TODO.md:314`); no M6 spec package exists (`docs/specs/` glob), so per the governance triad rule none of it is implementable yet.

| Capability | Asset | Evidence (repo path) | Maturity |
|---|---|---|---|
| Pipeline-scoped admission policy (3 constraints) | Gatekeeper C1 (main-protected), C2 (Score schema valid), C3 (Infracost delta) + Rego tests | `policies/C1-*.rego`, `C2-*.rego`, `C3-*.rego` + `*-constraint.yaml`; `policies/README.md` (6/6 PASS via `opa test --v0-compatible`) | live -- pipeline scope only |
| Gatekeeper deployment | OpenTofu module (helm + 3 `kubectl_manifest` constraints) | `iac/modules/gatekeeper/constraints.tf`, `main.tf`; `iac/main.tf:10-13` | live |
| Flux-delivered constraint mirror | platform-addons seed constraints | `seed-repos/platform-addons/clusters/default/gatekeeper/constraints.yaml` | live |
| Gatekeeper live smoke | template presence check in M2 smoke | `scripts/smoke-gatekeeper.sh` | live |
| Agent-side preventive guards (6) | rr-emoji/bash/brew/tofu/commit/verify-guard; stdlib-only Go binaries + git hooks | `plugins/rr-policy-guards/tools/<guard>/`, `plugins/rr-policy-guards/git-hooks/`, `scripts/install-git-hooks.sh` | live -- agent context only, not CI |
| Publication SCA gate | verify-guard runs Semgrep + Gitleaks + `yarn/npm audit` (High+) + `govulncheck`; blocks push | `plugins/rr-policy-guards/tools/verify-guard/exec.go:96-169` | live |
| Local audit trail (guards) | append-only JSONL, mode 0600, rotated | `plugins/rr-policy-guards/tools/verify-guard/audit.go`; writes to `~/.rational-reserve/logs/` | live -- local-only, not centralized or queryable |
| Security policy / coordinated disclosure | SECURITY.md with reporting channels and controls list | `SECURITY.md` (commit `24c33fb`) | live |
| Dependency remediation posture | Yarn resolutions for High/Critical advisories; grpc bump in hello-m2 | `backstage/package.json` resolutions; commits `24c33fb`, `67a17f9`; TODO.md dependency backlog (DONE 2026-06-30) | live |
| GitHub-side scanning (Dependabot, CodeQL) | referenced as running on the GitHub mirror | `SECURITY.md:121`, `.github/workflows/sync-from-gitea.yml` | partial -- no in-repo `dependabot.yml` or CodeQL workflow (`.github/` contains only the sync workflow); exact GitHub-side configuration UNCERTAIN |
| Portal authentication | custom Gitea OAuth provider; guest in dev only | `packages/backend/src/modules/giteaAuth.ts`, `packages/app/src/modules/giteaSignIn.tsx`, `scripts/smoke-auth.sh` | live |
| Production auth hardening | guest disabled, backend secret, `secure: true` OAuth, PostgreSQL | `backstage/app-config.production.yaml` | live |
| Authorization framework | Backstage permission framework enabled in production | `app-config.production.yaml:39-40`, `packages/backend/package.json:32-35` | partial -- policy module is `plugin-permission-backend-module-allow-all-policy`; framework on, everything permitted |
| Secrets management | OpenBao + external-secrets wiring | `iac/modules/external-secrets-wiring/`, `scripts/smoke-openbao.sh` | partial -- live but OpenBao runs dev-mode `inmem` storage (m2i-6 OPEN, `TODO.md:156`) |
| Policy angle in portal UI | PolicyCard + Policy tab on entity pages | `openchoreo-cards/PolicyCard.tsx`, `openchoreo-entity-page/index.tsx:79-86` | scaffold -- static links to C1/C2/C3; card states violations "will appear here once the M3 policy collector is wired" (`PolicyCard.tsx:50-53`) |
| Guard-layer governance specs | IDP Policy Guard Layer triad | `docs/specs/2026-05-28-IDP-Policy-Guard-Layer-{Requirements,Design-Specification,Technical-Specification}.md` | spec complete; partially implemented |

### 4.4b Security plane -- verified gaps

| ID | Gap | Evidence |
|---|---|---|
| SEC-G1 | Threat intelligence platform absent: no MISP, OpenCTI, or any feed ingestion; the plane's defining capability does not exist | repo-wide case-insensitive grep (trivy/grype/falco/MISP/OpenCTI/TheHive/Cortex/Velociraptor/custodian/SIEM/etc.) hits only docs (`TODO.md:314`, `docs/specs/m1-substrate/requirements.md:166`, `docs/specs/m2-iac-cd/requirements.md:216-221`, `docs/specs/Revised-Requirements.md`, `PROJECT_SUMMARY.md`) |
| SEC-G2 | No container image vulnerability scanning: CI builds and pushes images to the registry unscanned; source-level SCA does not cover the built artifact | grep scan/trivy/audit/vuln in `seed-repos/hello-m2/.gitea/workflows/ci.yaml` and `iac/templates/ci.yaml` -- no matches |
| SEC-G3 | No runtime security/intrusion detection (Falco, eBPF, runtime policies); explicitly deferred to M6 | same repo-wide grep; `docs/specs/m2-iac-cd/requirements.md:82` (FR-25), `:221` |
| SEC-G4 | No SIEM / centralized security event aggregation: guard audit events, Gatekeeper violations, and auth events are not aggregated, alerted on, or queryable in the portal | grep `SIEM`; guards write only local JSONL (`~/.rational-reserve/logs/verify-guard.jsonl`) |
| SEC-G5 | No incident response / case management (TheHive + Cortex): docs-only M6 rows | repo-wide grep |
| SEC-G6 | No DFIR endpoint visibility (Velociraptor): docs-only M6 rows | repo-wide grep |
| SEC-G7 | No cloud/resource governance (Cloud Custodian): docs-only | `docs/specs/m1-substrate/requirements.md:166`, `docs/specs/m2-iac-cd/requirements.md:216`, `TODO.md:314` |
| SEC-G8 | Policy violation visibility absent: PolicyCard collector not wired; no Gatekeeper audit/violation exporter anywhere; violations of C1-C3 are invisible in the portal | `PolicyCard.tsx:50-53`; repo-wide grep |
| SEC-G9 | No TLS on portal routes: all Envoy Gateway routes (`gitea.local`, `signoz.local`, `opencost.local`) are plain HTTP; cert-manager exists in the sibling cluster but nothing in this repo uses it | grep `tls|https|certificate|cert-manager` in `iac/modules/networking/` and `scripts/` |
| SEC-G10 | No security surface in Backstage: no Security tab or plugin; there is no place in the portal where a team member goes to "do security" | grep `security|permission` in `backstage/packages/app/package.json`; `openchoreo-entity-page/index.tsx` tab list |
| SEC-G11 | No in-repo Dependabot/CodeQL config: scanner posture is not version-controlled here and depends on GitHub-side settings surviving the mirror relationship | `.github/` listing -- only `workflows/sync-from-gitea.yml` |
| SEC-G12 | Authorization is allow-all: production enables the permission framework but with the allow-all policy; no real RBAC policy defined | `app-config.production.yaml:39-40`; `plugin-permission-backend-module-allow-all-policy` |
| SEC-G13 | Guard registration is per-user/per-machine and lives outside the repo; the guard README references an in-repo `hooks/hooks.json` (line 41) that does not exist -- doc drift and a portability gap for a multi-user portal | `plugins/rr-policy-guards/README.md:41,78-82`; registration at `~/Projects/Sovereign/Structure/hooks/hooks.json`, `~/.claude/settings.json` |
| SEC-G14 | OpenBao runs dev-mode inmem storage (m2i-6 OPEN) | `TODO.md:156` |
| SEC-G15 | No M6 spec package: everything M6 is a one-line draft row; per the governance triad rule none of it is implementable yet | `docs/specs/*` glob |

### 4.5 Engagement plane -- assets

| Capability | Asset | Evidence (repo path) | Maturity |
|---|---|---|---|
| Unified smoke umbrella (AUTH + M2 + M3 + M4 + BACKSTAGE-PRODUCTION) | `smoke-all.sh`, 5 suites, combined PASS/FAIL | `scripts/smoke-all.sh` | live (ALL SMOKE SUITES PASSED per SESSION_HANDOFF.md 2026-08-18) -- stdout-only reporting |
| M2 per-tool smokes (7) | tofu, actions, flux, score, infracost, gatekeeper, openbao | `scripts/smoke-m2.sh`, `scripts/smoke-{tofu,actions,flux,score,infracost,gatekeeper,openbao}.sh` | live |
| Live pipeline execution test | dispatches a real Gitea Actions run, polls conclusion up to 5 min | `scripts/smoke-actions.sh` | live -- the only asset that triggers execution rather than checking state |
| M3 full-spectrum harness (22 checks) | offline/cluster modes, predictor vectors, catalog assertions, live trace-ingestion assertion, cost-artifact presence | `scripts/smoke-m3.sh` (415 lines) | live (22/22 per handoff) |
| M4 cost smoke | OpenCost allocation endpoint | `scripts/smoke-m4.sh` | live |
| M4 networking smoke | Envoy routes, HTTP 200 for the three `.local` hostnames | `scripts/smoke-m4-networking.sh` | live (3/3) |
| Auth smoke | OAuth start-endpoint redirect check | `scripts/smoke-auth.sh` | live |
| Production Backstage smoke | HTTP 200 on :7009 | `scripts/smoke-backstage-production.sh` | live |
| Playwright e2e | guest sign-in; verifies two catalog links; HTML report to `e2e-test-report/` | `backstage/playwright.config.ts`, `packages/app/e2e-tests/app.test.ts` | partial (single test) |
| Jest unit test | single App render test | `backstage/packages/app/src/App.test.tsx` | scaffold (1 test) |
| Go unit tests: score2openchoreo | 4 test files incl. golden fixtures + schema-pin test | `tools/score2openchoreo/{convert,main,schema,schema_pin}_test.go`, `fixtures/` | live |
| Go unit tests: 6 policy guards | 27 test files | `plugins/rr-policy-guards/tools/*/*_test.go` | live |
| Rego policy tests (C1/C2/C3, 6 tests) | `opa test --v0-compatible policies/*.rego -v` | `policies/C1-test.rego`, `C2-test.rego`, `C3-test.rego` | live -- manual invocation |
| CI validation steps in app pipeline | Score validate, `tofu plan` (PR), Infracost delta + PR comment, docker build, render + commit | `seed-repos/hello-m2/.gitea/workflows/ci.yaml`, `iac/templates/ci.yaml` | live -- contains zero test-execution steps (see ENG-G4) |
| Pre-commit/pre-push CI-equivalent gate | verify-guard Phase 1 toolchain + Phase 2 forge checks; push adds SCA | `plugins/rr-policy-guards/tools/verify-guard/main.go:7`, `exec.go:43,128-170` | live -- local-agent gate, not a portal surface |
| Namespace predictor verification | smoke-m3.sh vectors + preflight | `scripts/smoke-m3.sh:87-137`, `scripts/preflight-m3.sh` | partial (indirect coverage; no Go test file) |
| Dormant Gitea Actions API proxy | `/api/proxy/gitea-actions` with admin token | `backstage/app-config.yaml:63-69` | scaffold -- intentionally deferred consumer (FR-41); zero consumers |
| Ad-hoc browser sign-in e2e scripts | Node + Python variants | `scripts/test-gitea-signin.mjs`, `scripts/test-gitea-signin.py` | orphaned -- zero references anywhere in repo |

### 4.5b Engagement plane -- verified gaps

| ID | Gap | Evidence |
|---|---|---|
| ENG-G1 | No in-portal CI/test run visibility: no Backstage CI/CD tab, card, or plugin consuming Gitea Actions runs; the five entity tabs are Deployment/Observability/Cost/Policy/Platform only | grep `gitea-actions` in `packages/app/src`; `openchoreo-entity-page/index.tsx` |
| ENG-G2 | No test triggering from the portal: no dispatch consumer of the proxy; dispatch exists only in CLI `smoke-actions.sh` | `backstage/packages/**`; `scripts/` |
| ENG-G3 | No machine-readable results or aggregation: smoke suites emit stdout PASS/FAIL only; no JUnit/xunit/report.json; Playwright HTML report stays local | grep `junit|xunit|tap format|report.json` in `scripts/` no hits |
| ENG-G4 | No test steps in app CI: hello-m2 (the only seeded project) ships with zero executed unit/e2e tests in its pipeline; new projects inherit a test-less pipeline | `seed-repos/hello-m2/.gitea/workflows/ci.yaml` and `iac/templates/ci.yaml` read in full |
| ENG-G5 | No CI for the developer-portal repo itself (cross-reference CTL-G7): platform test suites run only ad-hoc locally or via verify-guard inside agent sessions | `.github/workflows/` only the mirror; root `.gitea/` absent |
| ENG-G6 | No load/performance testing | repo-wide grep `k6|locust|jmeter|vegeta|fortio|gatling|artillery` -- only incidental hits |
| ENG-G7 | No chaos/failure testing | repo-wide grep `chaos-mesh|litmus|toxiproxy`, `chaos|fault injection` in docs -- no hits |
| ENG-G8 | No Go unit tests for the namespace predictor: a load-bearing deterministic tool has only smoke-vector coverage | `tools/namespace-predictor/` contains only `main.go` |
| ENG-G9 | No unit tests for the custom Backstage modules (openchoreo-cards, entity-page, giteaSignIn): breakage is caught only by the single Playwright sign-in test | glob `backstage/packages/*/src/**/*.test.{ts,tsx}` -- only `App.test.tsx` |
| ENG-G10 | No per-project test scoping: smoke suites are milestone-global; only hello-m2 has any CI; a team onboarding project N gets no project-scoped test surface | `scripts/smoke-m2.sh`, `smoke-m3.sh`, `smoke-m4.sh` structure |
| ENG-G11 | Kubernetes plugin dead (cross-reference ORC-G2), which also limits post-test runtime inspection | `app-config.yaml:122`; `packages/app/src` |

---

## 5. Traversal matrix: how a team moves between planes today

Legend: **W** = works and validated (evidence cited), **P** = partial (works with degradation: leaves the portal, stale targets, or one direction only), **B** = broken/stale, **N** = nothing exists. **(CLI)** = the traversal exists but only from a terminal, never in the portal UI.

| From \ To | Observation | Control | Orchestration | Security | Engagement |
|---|---|---|---|---|---|
| **Observation** | -- | W | P | N | P (one-way: engagement validates observation; no reverse path) |
| **Control** | P | -- | P | P (static Policy tab links only) | P (entity "CI runs" link leaves the portal) |
| **Orchestration** | P | W (one-way-out) | -- | W | N |
| **Security** | B | W (strongest) | W | -- | W (CLI) -- the only automated, validated traversal |
| **Engagement** | W (CLI) | P | W (CLI; reports nothing back) | W (CLI) | -- |

### 5.1 Working traversals (evidence)

- **Observation -> Control:** CostCard links to the Gitea `platform-config` cost artifact and CI runs (`CostCard.tsx:29-30,90-98`); PlatformCard links to repo and Actions (`PlatformCard.tsx:25-29`); `seed-repos/hello-m2/catalog-info.yaml:16-25` carries Gitea links. Works via the 3333/3002 port-forwards that `scripts/start-backstage.sh` guarantees.
- **Observation -> Orchestration (partial):** the strongest data-level touchpoint -- trace resource attributes carry `openchoreo.*` identity + `git.commit.sha` (`main.go:94-104`), verified live in ClickHouse; DeploymentCard surfaces the predicted runtime namespace and links to ReleaseBindings (`DeploymentCard.tsx:52-57`). Degraded by ORC-G8 (9090 link UNCERTAIN) and OBS-G6 (8080 links).
- **Control -> Orchestration (partial):** entity links to OpenChoreo API and CI runs (`catalog-info.yaml:16-25`); `commit-to-platform-config.sh` is the write-back path. Degraded by dead links (TRV-B2) and the broken scaffolder (TRV-B10).
- **Orchestration -> Control (works, one-way-out):** Gitea catalog provider auto-imports repos as entities (`app-config.yaml:100-109`); links jump to repo and Actions. Nothing in Gitea links back (TRV-B9).
- **Orchestration -> Security (works, second-strongest):** PolicyCard tab + Gatekeeper C1-C3 pipeline constraints (`policies/`) + Infracost delta gate in CI; orchestration secrets flow from OpenBao (`iac/modules/external-secrets-wiring/`).
- **Security -> Control (works, strongest):** guards gate what enters Gitea (rr-commit-guard via `git-hooks/` + PreToolUse; Gitleaks blocks secret commits); C1 protects `main` on platform-addons; `smoke-gatekeeper.sh` runs inside the M2 smoke suite.
- **Security -> Orchestration (works):** rr-tofu-guard forces OpenTofu mutations through approved lifecycle scripts; rr-verify-guard gates every clean push; C2/C3 gate pipeline output at admission.
- **Security <-> Engagement (works, CLI):** `smoke-gatekeeper.sh`, `smoke-auth.sh`, `smoke-openbao.sh` are automated and wrapped by `smoke-m2.sh`/`smoke-all.sh` -- the only automated, validated plane-to-plane traversal in the platform.
- **Engagement -> Observation (works, CLI):** `smoke-m3.sh:337-374` generates a live request and asserts trace ingestion in ClickHouse -- tests directly verify telemetry.
- **Engagement -> Orchestration (works, CLI):** `smoke-actions.sh` dispatches a real workflow; `smoke-flux.sh` reconciles Flux; `smoke-score.sh` executes score2openchoreo. Results are not reported back into any orchestration surface.

### 5.2 Traversal breakdown register

| ID | Breakdown | Evidence |
|---|---|---|
| TRV-B1 | SigNoz base URL hardcoded to `http://localhost:8080` in three cards; no managed port-forward serves it (`start-backstage.sh:50-77` manages only 3333/3002/29003); AGENTS.md documents 3301; `signoz.local` exists via Envoy. Three inconsistent access paths plus one unserved address. UNCERTAIN whether 8080 ever resolved | `ObservabilityLinksCard.tsx:21`, `DeploymentCard.tsx:61`, `PlatformCard.tsx:30` |
| TRV-B2 | DeploymentCard links `/iac/environments/${env}/kustomization.yaml` -- a relative portal route with no matching page anywhere in the app. Dead link | `DeploymentCard.tsx:58`; grep confirms no `/iac` route |
| TRV-B3 | DeploymentCard links `http://localhost:8080/dashboards?namespace=...` -- nothing in the repo serves port-8080 dashboards. UNCERTAIN what was intended (likely an Argo CD placeholder; Argo CD is explicitly not in the stack) | `DeploymentCard.tsx:61` |
| TRV-B4 | `/api/proxy/gitea-actions` is configured with an admin token but has zero frontend consumers -- a CI-runs-in-portal surface was started and abandoned (deferred FR-41) | `app-config.yaml:63-69`; grep `gitea-actions` in `packages/app/src` zero hits |
| TRV-B5 | The `openchoreo.dev/api-base` link target `http://localhost:9090` is forwarded only by `scripts/install-m1.sh:197` and has no smoke coverage; reachability UNCERTAIN from repo evidence | `scripts/install-m1.sh:197`; `scripts/start-backstage.sh:50-77` |
| TRV-B6 | No Security tab or surface: entity pages have Deployment/Observability/Cost/Policy/Platform only; the security plane has no home in the portal | `openchoreo-entity-page/index.tsx:75-91` |
| TRV-B7 | No engagement surface: no Tests/CI/Delivery tab; engagement is the one plane with no portal presence at all -- only outbound links and CLI suites | `openchoreo-entity-page/index.tsx`; `backstage/packages/**` |
| TRV-B8 | PolicyCard renders static links to C1/C2/C3 and states violations "will appear here once the M3 policy collector is wired" -- it is not wired | `PolicyCard.tsx:37-53` |
| TRV-B9 | Traversal is one-way-out: links leave the portal; nothing in Gitea, SigNoz, or OpenChoreo links back to the catalog entity; no reverse path from telemetry to test re-run or from a test result to its workload | plane maps, sections on Observation->Engagement and Orchestration->Control |
| TRV-B10 | Guided create->deploy traversal does not exist: the scaffolder is broken at the publish step (template targets `publish:github` / `allowedHosts: github.com`; no gitea module registered) | `examples/template/template.yaml:54-59`; `packages/backend/src/index.ts:18` |
| TRV-B11 | Guard registration lives in user-level settings outside the repo; a second team member on a fresh machine has no preventive guards until external settings are mirrored; the guard README references an in-repo `hooks/hooks.json` that does not exist | `plugins/rr-policy-guards/README.md:41,78-82` |
| TRV-B12 | The portal is read-only for orchestration and engagement: every mutation (deploy, promote, re-run, dispatch) happens via git push or CLI. This is a recorded design decision, not an accident -- changing it needs an explicit decision | `docs/specs/m2-iac-cd/design-specification.md:194`, `:452` ("human approval is a feature, not a bug") |

---

## 6. Requirements

Every functional requirement traces to a verified gap (section 4) or traversal breakdown (section 5) -- no invented features. Where a requirement depends on an undecided question, the OQ reference is given and the requirement is stated at the level of the decision, not the outcome.

### 6.1 Functional requirements -- traversal and portal coherence (cross-plane)

- FR-01: All external-tool base URLs used by portal cards and links (SigNoz, OpenCost, Gitea, OpenChoreo API) resolve from a single source of truth (app-config or entity annotation); no card hardcodes a localhost address. **Traces to:** OBS-G6, ORC-G8, TRV-B1, TRV-B3, TRV-B5.
- FR-02: Every link rendered on an entity page resolves to a working target or is removed; the dead `/iac/environments/...` route and `localhost:8080/dashboards` link are repaired (repointed at Gitea platform-config paths) or deleted per OQ-13. **Traces to:** TRV-B2, TRV-B3.
- FR-03: Every Component entity page presents a uniform five-plane layout (one section/tab per plane: Observation, Control, Orchestration, Security, Engagement); a plane with no wired data renders an explicit "not wired" state rather than static links or silence. **Traces to:** TRV-B6, TRV-B7, SEC-G10, ENG-G1.
- FR-04: CI run visibility inside the portal: a card consumes the existing dormant `/api/proxy/gitea-actions` proxy to list workflow runs and conclusions per project (fulfilling deferred FR-41 of `docs/specs/m2-iac-cd/requirements.md:119`). **Traces to:** ORC-G1, ENG-G1, TRV-B4.

### 6.2 Functional requirements -- observation plane

- FR-05: Workload and platform logs are collected into the existing SigNoz backend via the deployed standalone collector (or an equivalent single-backend mechanism per OQ-01), so the existing Logs link lands on real data. **Traces to:** OBS-G1.
- FR-06: Alerting is enabled (SigNoz alertmanager) with codified per-component alert rules and a decided notification channel (OQ-02); this implements the previously named but unimplemented FR-PROD-2 of the M3 requirements. **Traces to:** OBS-G2.
- FR-07: Workload metrics (RED/usage) are emitted and collected, completing the traces/metrics/logs triad the M3 requirements already promise. **Traces to:** OBS-G3.
- FR-08: Per-project dashboard definitions are materialized under `observability/dashboards/` and seeded by the install scripts, as the M3 technical specification (section 6.2 step 7) already names. **Traces to:** OBS-G4.
- FR-09: Traces are viewable without leaving the portal, via a SigNoz proxy endpoint and card mirroring the validated CostCard pattern. **Traces to:** OBS-G5, TRV-B9.
- FR-10: Telemetry durability is configurable (persistent volumes for ClickHouse and Prometheus), with defaults set by the retention decision (OQ-06). **Traces to:** OBS-G7.
- FR-11: Telemetry coverage extends beyond the hello-m2 demo workload to platform components (per the tenancy convention decided in OQ-05). **Traces to:** OBS-G8.

### 6.3 Functional requirements -- control plane

- FR-12: The scaffolder can publish a new project repository to Gitea end to end (upstream module or custom action per OQ-08); the publish step no longer targets github.com. **Traces to:** CTL-G1, ORC-G6, TRV-B10.
- FR-13: A paved-path project template for the locked-in stack exists (Score + `.gitea/workflows/ci.yaml` + `catalog-info.yaml` + Dockerfile), so project creation from the portal exercises the whole M2 loop. **Traces to:** CTL-G2.
- FR-14: Team provisioning and ownership synchronization exist (admin-provisioned Gitea users/teams mapped into Backstage users/groups per OQ-09), replacing hand-maintained static org.yaml as the ownership source. **Traces to:** CTL-G4, CTL-G9.
- FR-15: Dangling owner references in `catalog-info.yaml` resolve to real catalog entities. **Traces to:** CTL-G5.
- FR-16: `platform-config` and `platform-addons` are first-class catalog entities (kind per OQ-11). **Traces to:** CTL-G8.
- FR-17: The developer-portal repo has its own Gitea Actions CI running its existing test suites (Go tests, yarn test/tsc, opa test) per the OQ-27 decision on forge-run vs agent-mediated gating. **Traces to:** CTL-G7, ENG-G5.
- FR-18: Seed projects surface per-project TechDocs in the portal (`techdocs-ref` + minimal mkdocs), on a verified generator path. **Traces to:** CTL-G10.

### 6.4 Functional requirements -- orchestration plane

- FR-19: The already-installed Backstage kubernetes plugin is configured against k3d-openchoreo (auth model per OQ-17) and workload views are mounted on the Deployment tab, so the portal shows observed runtime objects, not only predicted namespaces. **Traces to:** ORC-G2, ENG-G11.
- FR-20: A documented, repeatable dev-to-staging promotion procedure exists (manual-commit runbook at minimum; portal-mediated promotion only if OQ-14 reverses the recorded manual-approval stance). **Traces to:** ORC-G3.
- FR-21: Observed OpenChoreo status (ReleaseBinding/ComponentRelease state) is visible in the catalog, complementing the predicted data shown today. **Traces to:** ORC-G5.
- FR-22: Workflow orchestration has a decided path: score2openchoreo Workflow/Trait fidelity and a team-facing workflow surface, or a documented decision that Argo Workflows stays OpenChoreo-internal (OQ-16). **Traces to:** ORC-G4.
- FR-23: A team operating multiple components has a cross-project view of deployments and runs. **Traces to:** ORC-G7.

### 6.5 Functional requirements -- security plane

- FR-24: PolicyCard shows live Gatekeeper violation state (via an audit exporter or API query) or an explicit not-wired indicator per NFR-04. **Traces to:** SEC-G8, TRV-B8.
- FR-25: Container images are scanned for vulnerabilities in CI before push to the registry (component per OQ-20 approval). **Traces to:** SEC-G2.
- FR-26: The portal has a Security surface (tab or section per FR-03) presenting live posture -- guard/SCA state, policy violations, scanner results -- or explicit not-wired states. **Traces to:** SEC-G10, TRV-B6.
- FR-27: A first threat-intelligence slice exists, implementing the scope decided in OQ-22 (full TIP vs feed-driven vulnerability awareness per component), gated on the OQ-19/OQ-20/OQ-21 decisions. **Traces to:** SEC-G1.
- FR-28: Dependency/scanner configuration is version-controlled in this repo (dependabot.yml, CodeQL workflow) or a documented decision records reliance on GitHub-side settings. **Traces to:** SEC-G11.
- FR-29: The Backstage authorization model is decided and implemented: a role-based permission policy, or documented acceptance of allow-all for the collaborative phase (OQ-25). **Traces to:** SEC-G12.
- FR-30: Preventive-guard distribution is portable: registration is a checked-in, install-script-managed asset so a second team member gets the same gates; the guard README's phantom `hooks/hooks.json` reference is fixed. **Traces to:** SEC-G13, TRV-B11.
- FR-31: Security events have a decided centralization path (OQ-26): guard JSONL + Gatekeeper violations shipped into the existing telemetry backend as a SIEM precursor, or a documented deferral. **Traces to:** SEC-G4.
- FR-32: The TLS posture of the `.local` routes is decided and implemented (cert-manager via Envoy Gateway) or documented as accepted plain HTTP (OQ-24). **Traces to:** SEC-G9.

### 6.6 Functional requirements -- engagement plane

- FR-33: The canonical app pipeline executes tests (`go test` for Go services; UI tests where a UI exists), and the template/live workflow drift (ORC-G9) is reconciled per OQ-18. **Traces to:** ENG-G4.
- FR-34: Smoke suites emit machine-readable results, and a committed results artifact exists following the proven cost-artifacts pattern (`scripts/ci/commit-cost-artifact.sh`), with the store location decided per OQ-28. **Traces to:** ENG-G3.
- FR-35: A portal-side test trigger exists (workflow dispatch through the existing proxy), if OQ-15/TRV-B12 decisions permit actuation; otherwise the CLI dispatch path is documented as the engagement trigger. **Traces to:** ENG-G2.
- FR-36: Per-project test results are visible in the portal (consuming FR-04 and FR-34). **Traces to:** ENG-G1, ENG-G3, TRV-B7.
- FR-37: The namespace predictor and the custom Backstage modules have unit tests (scope per OQ-30). **Traces to:** ENG-G8, ENG-G9.
- FR-38: New projects created from the paved-path template inherit a tested pipeline automatically (shape per OQ-31). **Traces to:** ENG-G10, CTL-G2.

### 6.7 Functional requirements -- collaboration (cross-plane)

- FR-39: Traversal is not one-way-out: project repositories and external surfaces carry a reference back to their catalog entity where the OQ-15 portal-philosophy decision makes this meaningful. **Traces to:** TRV-B9.
- FR-40: All new committed artifacts produced by this initiative (test results, cost artifacts, any future per-project records) are produced in an append-only, digest-friendly form compatible with the separately-specified record-immutability mechanism (companion workstream; OQ-10). **Traces to:** CTL-G6, ENG-G3.

### 6.8 Non-functional requirements

- NFR-01 **Open-source only.** Every added component must be open-source and license-compatible with the portfolio; no proprietary SaaS dependency may enter the traversal path.
- NFR-02 **Approval discipline.** Anything outside the locked-in stack (Gitea + Backstage Software Catalog and Score + OpenTofu + Gitea Actions + Gitea OCI registry; Flux and OPA/Gatekeeper previously approved) remains PROPOSED until it receives the same explicit user approval Flux (2026-04-20) and OPA/Gatekeeper (M2 pull-forward) received. This document approves nothing by itself.
- NFR-03 **Attribution triple maintenance.** Any dependency change regenerates `THIRD-PARTY-LICENSES.md` and `provenance/PROVENANCE.md` and re-issues `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md`; superseded certificates stay in git history. Attribution is recorded, never claimed.
- NFR-04 **No fabricated status.** Every portal surface shows live data or an explicit "not wired" state; placeholders must never be presented as live, and UNCERTAIN items stay labeled UNCERTAIN in docs and surfaces.
- NFR-05 **Record-immutability touchpoint.** Artifacts and stores designed under this initiative must be compatible with the separately-specified record-immutability mechanism (append-only, digest-able); the mechanism itself is out of scope here (section 10).
- NFR-06 **Accessibility.** New or modified portal UI surfaces meet WCAG 2.1 AA essentials: keyboard navigation, sufficient color contrast, labeled controls, and screen-reader-meaningful structure.
- NFR-07 **Local resource envelope.** Every added component declares its resource footprint and must fit the single-node k3d-openchoreo host (Colima) envelope, or be explicitly phased behind the capacity decision (OQ-21). Heavy stacks (e.g. a TIP plus its database dependencies) may not be installed on the shared cluster before that decision.
- NFR-08 **Non-regression.** `scripts/smoke-all.sh` (AUTH + M2 + M3 + M4 + BACKSTAGE-PRODUCTION) remains green after each phase, and the M2 delivery contract (how code flows to a running pod) is preserved.
- NFR-09 **Conventions.** Repo-driven, version-pinned, ASCII-only artifacts; script-driven install/teardown/smoke per component; deterministic logic preferred over interpreted where both are possible.
- NFR-10 **Governance.** Per-plane implementation work approved from this umbrella produces its own Requirements/Design/Technical specification triad per POLICIES.md before implementation.

---

## 7. Candidate component proposals (all PROPOSED -- nothing decided)

**Boundary restated.** The locked-in stack is: **Gitea + Backstage Software Catalog and Score + OpenTofu + Gitea Actions + Gitea OCI registry**, plus two explicitly approved additions: **Flux** (approved 2026-04-20, cluster add-on drift correction only) and **OPA/Gatekeeper** (approved M2 pull-forward, pipeline-scoped only). Every entry below is **PROPOSED** and pending explicit user approval under the same discipline (NFR-02). "Named in TODO" marks whether the idea already appears in the TODO.md M5-M7 draft rows or Option C notes; "new" means the plane maps surfaced it for the first time. No entry below is presented as decided.

| # | Plane | Addresses | Candidate (PROPOSED) | Named in TODO M5-M7 drafts? | Notes |
|---|---|---|---|---|---|
| P-01 | Observation | OBS-G1 | Extend the standalone OTel collector with a filelog/k8s-logs receiver (or SigNoz `k8s-infra` chart), writing into the existing ClickHouse | New | Loki alternative named only in the unaligned `docs/specs/Engineering-Spec.md:19` draft -- would add a second backend |
| P-02 | Observation | OBS-G2 | Enable SigNoz alertmanager (flip `values.local.yaml:47-48`) + codified per-component alert rules | Named in M3 specs (FR-PROD-2), not in the TODO M5-M7 table | Notification channel is a separate decision (OQ-02) |
| P-03 | Observation | OBS-G3 | OTel metrics SDK in the instrumentation pattern, or a prometheus receiver scraping annotated pods | New | Completes the triad the M3 requirements promise |
| P-04 | Observation | OBS-G4 | Materialize `observability/dashboards/` with SigNoz dashboard JSON seeded by `install-m3.sh` | Named in M3 technical specification (sec 6.2 step 7), unimplemented | Directory and seeding were never created |
| P-05 | Observation | OBS-G5 | `/api/proxy/signoz` + TracesCard querying by `service.name` + predicted namespace, mirroring CostCard | New | Reuses the validated CostCard integration shape |
| P-06 | Observation | OBS-G6 | SigNoz link repair: single base-URL source pointed at the existing `signoz.local` route | New | Bug fix, not a component |
| P-07 | Observation | OBS-G7 | ClickHouse + Prometheus PVCs (config-only change in two values files) | New | Blocked on retention decision (OQ-06) |
| P-08 | Observation | OBS-G8 | Instrument the score2openchoreo CI path; scrape Flux/Gatekeeper metrics via the collector | New | Extends the plane to the platform itself |
| P-09 | Control | CTL-G1 | `@backstage/plugin-scaffolder-backend-module-gitea` (upstream), or a custom `publish:gitea` action reusing existing Gitea integration credentials | New | Custom variant could also seed branch protection + webhooks, mirroring `seed-gitea-repos.sh` (OQ-08) |
| P-10 | Control | CTL-G2 | New Template entity `score-service` cloning the hello-m2 shape, with `fetch:template` + `publish:gitea` + `catalog:register` | New | The self-service golden path that exercises the M2 loop |
| P-11 | Control | CTL-G4, CTL-G9 | Gitea org/team -> Backstage User/Group sync (custom entity provider or scheduled script) + admin provisioning script | Named in TODO Option C notes (problem 4, unified identity), not the M5-M7 table | Kills the static org.yaml drift |
| P-12 | Control | CTL-G3 | Defer file editing to the Gitea web UI; portal stays discovery + launchpad | New | A decision, not a component (OQ-07) |
| P-13 | Control | CTL-G7 | `.gitea/workflows/ci.yaml` for developer-portal, modeled on `iac/templates/ci.yaml` | New | Same push-to-validate guarantee for the platform's own files |
| P-14 | Control | CTL-G8 | `catalog-info.yaml` (kind per OQ-11) added to `platform-config` and `platform-addons` seeds | New | Zero-code; provider auto-discovers on next scan |
| P-15 | Control | CTL-G5 | Add `user:default/nnos` to org.yaml, or re-point owners to `group:default/openchoreo` | New | One-file fix |
| P-16 | Control | CTL-G10 | `techdocs-ref` + minimal mkdocs for hello-m2; verify the generator path on this host (docker vs `runIn: local`) | New | Generator path currently unverified |
| P-17 | Orchestration | ORC-G1 | CI-runs card consuming the dormant `/api/proxy/gitea-actions` proxy | New | Fulfills deferred FR-41; no maintained official Gitea Actions Backstage plugin is known -- upstream availability UNCERTAIN, verify before building |
| P-18 | Orchestration | ORC-G2 | Configure the already-installed `@backstage/plugin-kubernetes` (cluster locator for k3d-openchoreo, EntityKubernetesContent on the Deployment tab) | New | Auth model per OQ-17 |
| P-19 | Orchestration | ORC-G5 | OpenChoreo catalog entity provider / status backend pushing observed ReleaseBinding state into entities | Named in TODO.md:295 structural backlog | Would require its own triad per governance |
| P-20 | Orchestration | ORC-G3 | Promotion action (scaffolder task or backend endpoint) wrapping `commit-to-platform-config.sh staging` | New | M2 design explicitly deferred the "promote button"; reverses the documented manual-approval stance -- OQ-14 |
| P-21 | Orchestration | ORC-G4 | Extend score2openchoreo to emit Workflow/Trait CRDs; expose Argo Server UI via an Envoy HTTPRoute | Fidelity gap named in TODO.md:283; Argo UI exposure new | Division of labor with Gitea Actions per OQ-16 |
| P-22 | Orchestration | ORC-G7 | Fleet board: thin page over the Gitea Actions API + OpenChoreo API | New | Cross-project team view |
| P-23 | Orchestration | ORC-G8 | Stale card link fixes (9090/8080/`/iac` -> `signoz.local`, managed forwards, or proxy endpoints) | New | Cheap correctness fix regardless of bigger choices |
| P-24 | Orchestration | (M7 draft) | MCP surfacing of orchestration to agents | Named in TODO.md M7 row -- draft, never decided | Carried here for completeness; not part of any RECOMMENDED phase |
| P-25 | Security | SEC-G1 | MISP (threat-intel platform) | Named in TODO.md:314 (M6 row) | Heavy on one k3d host -- OQ-21 |
| P-26 | Security | SEC-G1 | OpenCTI (alternative TIP, stronger graph correlation) | New | Alternative to P-25; pick at most one per OQ-22 |
| P-27 | Security | SEC-G2 | Trivy in Gitea Actions CI (image scan before push; Gatekeeper image allow-list later) | Named in draft `docs/specs/Revised-Requirements.md:35` (NFR-04), not in the TODO table | Smallest security-slice component |
| P-28 | Security | SEC-G3 | Falco (eBPF runtime detection) | New | Deferred to M6 triad work |
| P-29 | Security | SEC-G4 | Wazuh (full SIEM), or ship guard JSONL + Gatekeeper audit logs into the existing SigNoz/ClickHouse as a SIEM precursor | New (both options) | OQ-26 picks the path |
| P-30 | Security | SEC-G5 | TheHive + Cortex (incident response; Cortex analyzers double as intel enrichment) | Named in TODO.md:314 (M6 row) | Deferred to M6 triad work |
| P-31 | Security | SEC-G6 | Velociraptor (endpoint DFIR) | Named in TODO.md:314 (M6 row) | Deferred to M6 triad work |
| P-32 | Security | SEC-G7 | Cloud Custodian (resource governance; user files it under Observability) | Named in TODO.md:314 (M6 row) | Deferred to M6 triad work |
| P-33 | Security | SEC-G8 | Gatekeeper audit exporter -> PolicyCard | New | Smallest high-value fill: wires an existing scaffold card to real data |
| P-34 | Security | SEC-G9 | cert-manager integration with Envoy Gateway for the `.local` routes | New (as in-repo config) | cert-manager already runs in the sibling cluster |
| P-35 | Security | SEC-G10 | Security tab on entity pages, or a security-insights-style plugin | New | Gives the plane a home like the other four have |
| P-36 | Security | SEC-G11 | In-repo `.github/dependabot.yml` + CodeQL workflow | New | Version-controls what is currently GitHub-side-only (UNCERTAIN) |
| P-37 | Security | SEC-G12 | Role-based Backstage permission policy replacing allow-all | New | OQ-25 decides timing |
| P-38 | Engagement | ENG-G4 | Test stages in the canonical pipeline (`go test`; Playwright where a UI exists) + template/live drift fix | New | Smallest change; locked-in stack already runs CI per push |
| P-39 | Engagement | ENG-G3 | JUnit/machine-readable smoke output + committed results artifact, mirroring the `cost-artifacts/` pattern | New | Feeds portal display and the immutability touchpoint (FR-40); store per OQ-28 |
| P-40 | Engagement | ENG-G2 | Workflow-dispatch trigger from the portal through the existing proxy | New | `smoke-actions.sh` already proves the dispatch pattern; gated on TRV-B12 decision |
| P-41 | Engagement | ENG-G3 (alternative) | Allure Report or similar historical dashboard | New | Heavier alternative to P-39; pick one |
| P-42 | Engagement | ENG-G6 | k6 + k6-operator (load testing) | New | Scope on a single-node k3d per OQ-29 |
| P-43 | Engagement | ENG-G7 | Chaos Mesh or LitmusChaos (failure injection) | New | Heavyweight locally; OQ-29 |
| P-44 | Engagement | ENG-G8, ENG-G9 | Unit tests for `tools/namespace-predictor` and the custom Backstage modules | New | No new dependencies |
| P-45 | Engagement | (use of existing) | Argo Workflows (already in `openchoreo-workflow-plane`) as a test-orchestration engine for multi-step e2e | New use of an existing asset | Zero new components; depends on OQ-16 |

**Touchpoint -- not specified here.** The record-immutability workstream's own candidates (Gitea-enforced signed commits + branch protection verified in CI; the provenance-certificate SHA-256 digest pattern extended to document snapshots; per-agent Gitea tokens for machine actors) are partially named in TODO.md (goal directive slice 2; M7 per-agent tokens row). They are owned by the companion workstream and appear in this document only as NFR-05 / FR-40 / OQ-10.

---

## 8. Phased roadmap (RECOMMENDED -- pending approval)

The phasing below is **RECOMMENDED**, not decided. Reasoning, stated once and applied throughout:

1. **Wire before you install.** The cheapest, highest-traversal-value work is connecting existing scaffolds to live data (dormant proxy, static PolicyCard, hardcoded links, installed-but-dead plugins). It needs no new cluster components, no stack approvals, and no capacity headroom, and it immediately establishes the NFR-04 "live or not-wired" honesty the collaborative portal depends on.
2. **Self-service before depth.** A collaborative portal that cannot mint a project (TRV-B10) is a single-operator dashboard. The control loop (scaffolder -> repo -> CI -> deploy -> promote) is the spine every other plane hangs off.
3. **Heavy installs last and gated.** Security's namesake components (MISP/TheHive-class stacks with Elasticsearch/Cassandra dependencies) are the largest resource commitment on a single k3d/Colima host and all sit outside the locked-in stack; they are phased behind explicit approval (OQ-20), the capacity decision (OQ-21), and the slice-vs-M6 sequencing decision (OQ-19), which is flagged explicitly below.

### Phase 1 -- Wire what exists (RECOMMENDED first)

Scope: FR-01, FR-02, FR-03, FR-04, FR-15, FR-16, FR-24, FR-37; the drift-fix portion of FR-33 (OQ-18 housekeeping).

- **Why first:** every item converts an existing scaffold into a live surface or removes a lie from the UI at frontend/config cost only; zero new cluster components; zero capacity risk; highest traversal improvement per unit effort.
- **Entry criteria:** user approves the phase scope; dev environment running (`start-backstage.sh`); OQ-13 (fix vs remove dead links) and OQ-11 (catalog kind for GitOps repos) answered -- both are small.
- **Exit criteria:** no hardcoded dead localhost links on entity pages; CI runs visible in the portal via the existing proxy; PolicyCard shows live Gatekeeper violations or an explicit not-wired state; catalog owners resolve; `platform-config`/`platform-addons` appear in the catalog; `smoke-all.sh` green.
- **Dependencies:** none outside the repo.

### Phase 2 -- Self-service project lifecycle (RECOMMENDED second)

Scope: FR-12, FR-13, FR-38, FR-14, FR-17, FR-18, FR-20.

- **Why second:** restores the broken create->deploy traversal (TRV-B10) and makes the platform multi-operator (CTL-G4/G9); every later plane benefits because new projects arrive with tested pipelines and docs by construction.
- **Entry criteria:** Phase 1 exit met; OQ-08 (upstream module vs custom publish action), OQ-09 (team model), OQ-14 (promotion stance), OQ-27 (forge-run self-CI) answered; the chosen scaffolder publish dependency approved (NFR-02) and the attribution triple regenerated (NFR-03).
- **Exit criteria:** a new project created from the portal pushes through CI, deploys to dev, and promotes to staging via the documented procedure; developer-portal self-CI runs on Gitea Actions; new projects inherit test stages.
- **Dependencies:** Phase 1; Gitea admin access for provisioning scripts.

### Phase 3 -- Observation depth and engagement surfaces (RECOMMENDED third)

Scope: FR-05, FR-06, FR-07, FR-08, FR-09, FR-10, FR-11, FR-19, FR-34, FR-35, FR-36, FR-33 (test stages), FR-21, FR-39.

- **Why third:** completes the telemetry triad the M3 requirements promised and gives engagement a portal home; deliberately after Phases 1-2 because dashboards and alerts for a platform that cannot yet mint projects would serve one demo app only.
- **Entry criteria:** Phase 2 exit met; OQ-01 (log strategy), OQ-02 (alert channel), OQ-03 (SigNoz access), OQ-05 (tenancy), OQ-06 (retention), OQ-15 (portal philosophy), OQ-17 (k8s plugin auth), OQ-28 (results store) answered.
- **Exit criteria:** traces/metrics/logs for workloads visible from inside the portal; alerts fire to the decided channel; per-project dashboards seeded; Deployment tab shows observed runtime objects; test results machine-readable and visible per project; `smoke-all.sh` green.
- **Dependencies:** Phases 1-2; capacity check per NFR-07 for persistence volumes.

### Phase 4 -- Security vertical slice (RECOMMENDED fourth; gated)

Scope: FR-25, FR-26, FR-28, FR-29, FR-30, FR-31, FR-32, then FR-27 (threat-intel first slice) last.

- **Why fourth, and why gated:** FR-25/FR-26/FR-28/FR-30 are small and high-value (scan the artifact CI already builds; give the plane a portal home; version-control scanner config; make guards portable). FR-27 is different in kind: it is the plane's namesake capability, every candidate is outside the locked-in stack, and the leading candidates are heavy on one host. The **security minimal vertical slice question is flagged explicitly**: whether a slice (Trivy in CI + PolicyCard wiring + one intel feed) is pulled forward ahead of M5/M6, or the plane stays M6-owned, is an explicit user decision (OQ-19), not something this roadmap assumes.
- **Entry criteria:** Phases 1-3 exit met; OQ-19 (slice vs M6), OQ-20 (stack approvals), OQ-21 (capacity), OQ-22 (TIP vs feed-driven awareness), OQ-24 (TLS), OQ-25 (authorization), OQ-26 (audit centralization) answered; each approved component recorded with its approval date, Flux/Gatekeeper-style.
- **Exit criteria:** images scanned in CI before push; Security tab live or honestly not-wired; scanner config in-repo; decided authorization and TLS postures live; guard distribution portable; threat-intel slice (if approved) visible per component.
- **Dependencies:** Phase 1 (PolicyCard wiring builds on FR-24); explicit approvals.

### Explicitly deferred beyond Phase 4

Not scheduled here, each requiring its own triad and approval: Falco (P-28), TheHive + Cortex (P-30), Velociraptor (P-31), Cloud Custodian (P-32), full-TIP scope if chosen (P-25/P-26 at scale), MISP-class heavy installs generally; workflow orchestration surfaces (FR-22/P-21, pending OQ-16); fleet board (FR-23/P-22); load/chaos testing (P-42/P-43, pending OQ-29); MCP surfacing (P-24, M7 draft); M5 messaging (RabbitMQ/Kafka + OpenResty) -- untouched by this document.

---

## 9. Open questions

Consolidated and deduplicated from the five plane maps. Each entry states why it matters and what it blocks. UNCERTAIN markers are preserved from the maps.

| ID | Question | Why it matters | Blocks |
|---|---|---|---|
| OQ-01 | Log strategy: SigNoz-native collection (single backend) vs Loki (named only in the unaligned `Engineering-Spec.md` draft)? The M3 specs say "traces, metrics, logs (SigNoz)", which points one way. | Decides whether the platform runs one telemetry backend or two, and which query path the portal uses | FR-05 (P-01) |
| OQ-02 | Where should alerts notify a collaborative team -- Gitea issues, generic webhook, TheHive later? No channel exists today. | Alerting without a channel is silent; the channel choice couples observation to control or security tooling | FR-06 (P-02) |
| OQ-03 | Canonical SigNoz access: `signoz.local` via Envoy (needs /etc/hosts + NodePort), a managed 3301 port-forward per AGENTS.md, or a Backstage proxy? Cards currently point at a fourth, unserved address (UNCERTAIN whether 8080 ever resolved). | Three inconsistent paths already exist; every trace link and embed depends on picking one | FR-01, FR-09 (P-05, P-06) |
| OQ-04 | Is developer-portal's SigNoz the permanent observation plane, or interim until OpenChoreo's own observability-plane is adopted? It deliberately avoids `openchoreo-observability-plane` (`values.local.yaml:9-10`). | Sizes the investment in dashboards, durability, and in-portal rendering | FR-08 through FR-11 |
| OQ-05 | Multi-project telemetry tenancy: naming/ownership convention (`openchoreo.project` attribute as the per-project filter) and whether SigNoz community edition suffices for per-team scoping. Today everything keys off one demo workload. | A collaborative portal needs per-project isolation of signals | FR-11, collaboration scope |
| OQ-06 | Retention and durability targets: current demo settings (emptyDir, 3d traces, sampling 0.1) vs team-grade durability. | Persistence has a real disk/memory cost on one host; trend analysis is impossible without it | FR-10 (P-07) |
| OQ-07 | Is the Gitea web UI acceptable as the file-editing/PR surface, with the portal limited to discovery + launch links? | Decides whether any in-portal editing component is ever needed | FR-03 shape (P-12) |
| OQ-08 | Scaffolder publish: adopt the upstream gitea module, or write a custom action that also seeds branch protection and webhooks in one shot (mirroring `seed-gitea-repos.sh`)? | Trade-off between maintained upstream code and one-shot paved-path behavior | FR-12 (P-09) |
| OQ-09 | Team model: expected number of human users and orgs? Does `DISABLE_REGISTRATION: true` stay (admin-provisioned accounts)? Do M7 per-agent Gitea tokens share the same identity-mapping mechanism? | Sizes the provisioning and ownership-sync work; agent identity rides the same decision | FR-14, FR-29 (P-11) |
| OQ-10 | Record-immutability scope: signed-commit + protected-branch + CI verification sufficient, or a separate notarization/ledger (provenance-certificate pattern extended to all project documents)? **Owned by the companion workstream**; tracked here only as a touchpoint. | FR-40 must produce artifacts the chosen mechanism can notarize | FR-40 alignment |
| OQ-11 | Should `platform-config` / `platform-addons` be modeled as catalog entities, and as what kind (Resource vs System)? | Small but blocking for the Phase 1 catalog fix | FR-16 (P-14) |
| OQ-12 | Housekeeping: remove the dead `example-nodejs-template` and the unregistered github scaffolder module from `packages/backend/package.json`, or keep them as reference? | Dead scaffolding invites copy-paste of the wrong forge | Phase 1 hygiene |
| OQ-13 | The two dead DeploymentCard links (`/iac/...` route, `localhost:8080/dashboards`): repoint to Gitea platform-config paths, or remove? | Determines the FR-02 repair shape | FR-02 (P-23) |
| OQ-14 | Is promotion meant to stay manual-commit-only? The M2 design calls human approval "a feature, not a bug" (`design-specification.md:452`); an in-portal promote action reverses that stance. | Decides whether orchestration actuation ever enters the portal | FR-20 (P-20) |
| OQ-15 | Portal philosophy: deep-link out to Gitea/SigNoz/Argo UIs, or embed/proxy them into Backstage? Current cards do the former, and half the targets are stale. (Merges the engagement map's "custom CI card vs Gitea deep-links" question.) | One model must govern FR-03, FR-04, FR-09, FR-19, FR-36 or the portal drifts into a inconsistent mix again | FR-03, FR-04, FR-09, FR-19, FR-36, FR-39 |
| OQ-16 | Is Argo Workflows meant to stay OpenChoreo-internal, or become the team-facing job/workflow executor alongside Gitea Actions? Running both needs a stated division of labor. | Avoids two overlapping execution engines with no ownership line | FR-22 (P-21), P-45 |
| OQ-17 | Kubernetes plugin auth model: Backstage runs on host today (local kubeconfig works); the production path (`app-config.production.yaml`) changes how a cluster locator authenticates. Containerization is pending but not absent: `backstage/packages/backend/Dockerfile` and the `yarn build-image` script (`backstage/package.json:12`) exist, while the production runtime currently runs the backend on host via `scripts/start-backstage-production.sh`. | Wrong choice now means re-work when Backstage is containerized | FR-19 (P-18) |
| OQ-18 | Which CI file is canonical: `iac/templates/ci.yaml` (stale, `dev`) or the live `seed-repos/hello-m2/.gitea/workflows/ci.yaml` (`development`, OTEL + cost artifact)? Update the template or remove it. | The drift silently misinforms every new project scaffold | FR-33 (P-38) |
| OQ-19 | **Security minimal vertical slice: does the security plane stay owned by deferred M6, or does this roadmap pull a minimal slice (e.g. Trivy in CI + PolicyCard violation wiring + one intel feed) forward ahead of M5?** Flagged explicitly per the roadmap's gating. | The single largest sequencing decision in this document; changes what "collaborative portal" means for security | Phase 4 placement; FR-27 |
| OQ-20 | Stack approvals: MISP/OpenCTI, TheHive, Cortex, Velociraptor, Cloud Custodian, Falco, Trivy are all outside the locked-in stack and each needs the explicit approval Flux and OPA/Gatekeeper received. | Nothing security-side can be installed without it (NFR-02) | FR-25, FR-27, all M6 candidates |
| OQ-21 | Capacity: MISP + TheHive (with their Elasticsearch/Cassandra dependencies) on a single local k3d/Colima host is heavy -- is the plane scoped to local-demo fidelity or sized for real team use? | Determines whether the heavy candidates can run at all on the current host (NFR-07) | FR-27 (P-25, P-30), M6 sizing |
| OQ-22 | Threat-intel definition for the first slice: a full TIP (MISP/OpenCTI), or CVE/feed-driven vulnerability awareness (scanner + advisories surfaced per component in Backstage)? | The two scopes differ by an order of magnitude in components and capacity | FR-27 (P-25 vs P-26 vs feed-only) |
| OQ-23 | Guard portability: for a multi-user portal, does guard distribution become a checked-in, install-script-managed asset -- and should the README's phantom `hooks/hooks.json` reference be fixed? | A second team member currently gets no preventive guards at all | FR-30 |
| OQ-24 | TLS posture: keep plain-HTTP `.local` routes, or adopt cert-manager-signed certificates via Envoy Gateway now that M4 networking is live? | Plain HTTP is acceptable for loopback demos but not for team habits | FR-32 (P-34) |
| OQ-25 | Authorization model: is the allow-all permission policy acceptable for the collaborative phase, or must real roles land before team traversal is meaningful? | Multi-user without roles is structurally unenforced trust | FR-29 (P-37) |
| OQ-26 | Audit centralization: ship guard JSONL logs and Gatekeeper violations into SigNoz/ClickHouse as a SIEM precursor (reusing M3), or stand up a separate SIEM (Wazuh)? | Reuse is cheap but couples security events to a demo-grade retention setup | FR-31 (P-29) |
| OQ-27 | Is verify-guard's local, agent-mediated gate accepted as the platform's CI-equivalent, or should the forge itself run the platform test suites (developer-portal self-CI)? | Decides whether team execution without an agent session is covered | FR-17 (P-13) |
| OQ-28 | Test-results persistence: commit artifacts into `platform-config` (GitOps-visible but pollutes the deploy repo) vs a separate repo/store -- and should results fall under the record-immutability mechanism? | Store choice affects both portal display and FR-40 alignment | FR-34, FR-36 (P-39) |
| OQ-29 | Load/chaos scope on a single-node local k3d: is realistic load/failure testing in scope for a local IDP, or deferred to a future multi-node environment? | k6/Chaos Mesh are real resource commitments with unclear fidelity on one node | P-42, P-43 |
| OQ-30 | Namespace predictor: add real Go unit tests, or is the smoke-vector contract (canonical vector + edge cases in `smoke-m3.sh`) sufficient? | A load-bearing deterministic tool currently has only indirect coverage | FR-37 (P-44) |
| OQ-31 | Should every seeded project get the hello-m2 pipeline with test stages via the scaffolder template, making engagement automatic at project creation? | Determines whether engagement scales by construction or by per-project effort | FR-38 (P-10, P-38) |

---

## 10. Out of scope for this document

- **Record-immutability mechanism for project documents.** Owned by the companion workstream (TODO.md goal directive, slice 2). This document only constrains new artifacts to be compatible with it (NFR-05, FR-40) and tracks its scoping question (OQ-10).
- **Actual component installs or configuration changes.** This is a requirements and roadmap document; no cluster state changes.
- **Per-plane design and technical specifications.** Required per governance (POLICIES.md triad rule) before any implementation work approved from this umbrella; they follow this document, one triad per approved increment.
- **M5 messaging (RabbitMQ or Kafka + OpenResty front-door) and M7 MCP/agent integration**, except where their draft rows intersect the proposals above (P-24, OQ-09 agent tokens). They remain deferred TODO.md rows.
- **Cilium as CNI.** Remains the documented fresh-cluster rebuild path (`docs/specs/2026-06-30-M4-Networking-Technical-Specification.md`), unaffected by this roadmap.
- **Production-grade HA, multi-cluster, real cloud billing, and non-guest production deployment of Backstage itself** beyond the already-live production config (`app-config.production.yaml`).
- **The deferred M6 security stack at full scope** (runtime policies, TheHive/Cortex, Velociraptor, Cloud Custodian, full TIP). Only the explicitly gated Phase 4 slice is in this roadmap; the rest stays M6-owned pending its own triads.

---

**End of Requirements and Roadmap**

This document was created per the governing persona rules for new functional increments of the platform, consolidating the five verified plane maps dated 2026-08-18. All proposals herein are PROPOSED and all phasing is RECOMMENDED, pending explicit user approval per NFR-02.
