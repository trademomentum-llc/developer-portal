# PROJECT SUMMARY

> A snapshot of the three projects this session touched and their current
> state. For "what to do next" see TODO.md. For "where we stopped and what
> changed mid-session" see SESSION_HANDOFF.md.

**Snapshot date:** 2026-08-20 (Wave-0 security plane accepted on the resized 6c/12GiB Colima cluster: smoke-all reports ALL SMOKE SUITES PASSED (AUTH, M2, M3, M4, SECURITY, BACKSTAGE-PRODUCTION); hello-m2 CI run #46 green through the digest-pinned Trivy/OSV gates with security artifacts committed to platform-config and the pod rolled to :59b8c8d; Engagement-plane slice landed -- CiRunsCard + entity tab; record-immutability chain live with signed commits and checkpoint tag checkpoint-2026-08. Earlier milestone state: M1 substrate, M2 IaC+CD loop, M3 observability, M4 cost visibility and networking all deployed and smoke-validated; see SESSION_HANDOFF.md section 0a for the 2026-08-20 slice and TODO.md for debt.)

---

## Overview

The user is building a self-hosted Internal Developer Platform (IDP) on
their local machine, using the Platform Engineering community reference
architecture in `~/Downloads/platform.pptx` as the blueprint. The platform
is decomposed into seven milestones (M1 through M7); M1 substrate is
complete and M2 (IaC + CD loop) is validated end-to-end locally.

Three projects on disk are involved, in dependency order:

1. `~/Projects/Sovereign/openchoreo/`        -- the platform orchestrator (upstream OSS)
2. `~/Projects/Sovereign/rational-reserve/`  -- AI swarm orchestration layer (custom)
3. `~/Projects/Sovereign/developer-portal/`  -- the umbrella IDP build (this repo)

**2026-05-28 Specifications Update:** Full Requirements + Design + Technical Specification triad created for the Policy Guard Layer + IDP Milestone System (the load-bearing enforcement and tracking mechanism):

**2026-05-28 Option C Cohesion Pass (this subagent task):** Completed architecture analysis of developer-portal + OpenChoreo stack. Identified top 5 cohesion problems (namespace impedance, translation tax, catalog disconnect, ownership drift, config inconsistency) with cross-file deterministic evidence. Implemented 3 targeted edits for immediate glue (Gitea port alignment in app-config.yaml, enriched catalog-info.yaml entities with openchoreo.dev annotations + runtime ns templates, extended score2openchoreo/README with hash determinism proof + automation guidance). Updated this SUMMARY + TODO per mandatory rule. Full prioritized recommendations + code examples + automation notes in subagent final report. No new functional modules created in this pass (edits only); any future adapter layer will carry full triad.

**Dual-Track Strategy Note (2026-05-28):** While maturing the sovereign Jasterish/NeuroDiOS path (Option D), the project is also actively pursuing Option C in parallel: making developer-portal a stronger, first-class, cohesive extension of OpenChoreo as the interim solution to current cohesion pain. A sub-agent has been spawned to analyze and implement improvements in this area. A detailed gap analysis between current sovereign assets and OpenChoreo's load-bearing responsibilities has been created.

**New Phase — Production Multi-Angle Visibility Model (initiated 2026-05-28):** Moving beyond M3 kickoff specs, the focus is now on evolving the local IDP into a realistic working production model. The goal is comprehensive, multi-angle visibility (delivery, runtime, cost, policy, agents) into any project's full lifecycle, with developer-portal acting as a cohesive extension of the OpenChoreo platform.

Recent progress:
- New Requirements + Design specs for the production model.
- Scaffolding (observability/ dirs + full set of M3 scripts).
- hello-m2 OTEL instrumentation hardened (2026-06-30) with openchoreo.* resource attributes, git commit SHA, and deterministic runtime namespace.
- `score2openchoreo` extended with `--extra-env KEY=VALUE` so the CI pipeline can inject deployment-time telemetry variables without Score schema changes.
- hello-m2 CI workflow updated to compute the predicted runtime namespace and inject OTEL endpoint + OpenChoreo context + GIT_SHA.
- Namespace predictor corrected to a byte-for-byte replica of OpenChoreo's `GenerateK8sNameWithLengthLimit`; canonical vector updated to the live cluster value `dp-default-default-development-f8e58905`.
- SigNoz and standalone OpenTelemetry Collector installed on the live `k3d-openchoreo` cluster; the SigNoz collector's OTLP receiver was enabled by removing the OpAMP-only manager arguments.
- First end-to-end trace flow verified: test span accepted by the standalone collector, forwarded to the SigNoz collector, and stored in ClickHouse.
- New robust Backstage entity cards module (`modules/openchoreo-cards`) — five cards total (Overview, Observability, + new Cost, Policy, Deployment). All cards now wired to the deterministic namespace predictor (exact Go port). Module registered via createFrontendModule.
- Full mandated triad created for the cards module (Requirements + Design + Technical Specification, the latter quoting full sources + equivalence proof procedure).
- Complete M3 Production Multi-Angle Visibility Technical Specification created (2026-05-28), including coexistence rules, Helm contracts, OTEL hardening spec, full-spectrum test matrix, and automation safety notes. This supersedes the earlier m3-observability kickoff drafts for the production model phase.
- M3 Production Design Specification updated with realized cards + predictor integration.
- Namespace predictor (Go CLI + pure TS port) is now the single source of truth across UI cards and operational scripts.
- M3 implementation phase advanced: minimal values files created, install-m3.sh and teardown-m3.sh made executable with real Helm steps, smoke-m3.sh rewritten as the executable full-spectrum test harness (10/10 checks passed in offline mode on 2026-06-30, including canonical, truncation, underscore-normalization, and additional predictor vectors plus angle coverage). Both required Technical Specifications delivered.

**Session pause note (2026-05-29):** User directed pause after exceptional progress. All core M3 implementation and full-spectrum testing objectives met for this phase. Ready for clean resumption. See TODO.md for exact next items.

See updated TODO for the six parallel workstreams and the concrete next item now in progress.

Cross-references:
- Option C Cohesion triad (2026-05-28)
- M3 Production Multi-Angle Visibility triad
- IDP Policy Guard Layer triad (C1/C2/C3 constraints referenced by the new PolicyCard)

- `docs/specs/2026-05-28-IDP-Policy-Guard-Layer-Requirements.md`
- `docs/specs/2026-05-28-IDP-Policy-Guard-Layer-Design-Specification.md`
- `docs/specs/2026-05-28-IDP-Policy-Guard-Layer-Technical-Specification.md`

These cover progressive enforcement, Rego/Go guard architecture, milestone entry/exit criteria, evidence export, and integration with openchoreo + rational-reserve. They close the formal spec gap for the highest-value module in this project.

---

## 1. openchoreo

**Source:** https://github.com/openchoreo/openchoreo (cloned at `~/Projects/Sovereign/openchoreo/`)

**Role in the IDP:** Platform orchestrator. Spans three planes in the
architecture: Developer Control Plane, Platform Orchestration Plane, and
Security Plane. Load-bearing by design.

**State on disk:** Unchanged from M1. `check-tools.sh` still exits 0 with
all installed tool versions green.

**Cluster:** k3d cluster name `k3d-openchoreo` running on Colima. Five
OpenChoreo planes are in: `openchoreo-control-plane`,
`openchoreo-data-plane`, `openchoreo-observability-plane`,
`openchoreo-workflow-plane` (Argo Workflows bundled, NOT Argo CD), plus
the supporting namespaces `cert-manager`, `external-secrets`, `gitea`,
`openbao`, `kube-system`. All pods Ready at session close.

---

## 2. rational-reserve (RR)

**Source:** `~/Projects/Sovereign/rational-reserve/` (built earlier from requirement
docs on an external volume).

**Role in the IDP:** Developer Control Plane, Copilots/Agents/LLM slot.
Military-hierarchy AI swarm orchestration with 12 ranks, 14 MOS, SITREP
/ FRAGO / CASREP / AAR protocols.

**Build status:** v0.1 (Spine) + v0.2 (Adapter layer) COMPLETE. Unchanged
this session. 65 tests pass. Daemon + MCP shim + doctrine loader + SQLite
state store all working. Integration into the portal is deferred to M7.

---

## 3. developer-portal (current focus)

**Source:** `~/Projects/Sovereign/developer-portal/`. git repository, branch `main`.
The 2026-05-22 M2 closeout work is implemented locally. Local Gitea origin
uses the localhost:3333 port-forward and has received `main`; the latest
gitea-com push reached `gitea.com` but failed authentication, so a fresh
valid cloud credential/PAT is needed.

**Role:** The umbrella for the full self-hosted IDP build.

**Build status:** M1 substrate complete and healthy. **M2 is validated
end-to-end locally.** Push -> CI -> render -> commit -> Flux pulls -> Flux
applies -> OpenChoreo reconciles -> ComponentRelease -> Deployment -> Pod
was validated through a fresh `hello-m2` Gitea Actions run on 2026-05-22.
The live data-plane deployment is available and the `hello-m2` pod is
`1/1 Running`. External `gitea-com` push remains blocked by cloud
authentication, not by current network reachability.

Backstage local development is also wired to the repo catalog now. The app
title is `Developer Portal`; the catalog loads the root `catalog-info.yaml`
and `seed-repos/hello-m2/catalog-info.yaml`; and the Playwright smoke test
signs in as Guest and verifies both `developer-portal` and `hello-m2`
component links.

Dependency SCA: Backstage `package.json` resolutions pin High/Critical
transitive advisories; `yarn npm audit --all --recursive --severity high
--no-deprecations` is clean. `rr-verify-guard` now blocks push on High+
yarn/npm audit findings and on `govulncheck` hits for every Go module root.
Moderate deprecations (e.g. Material UI v4) remain until a coordinated
Backstage UI upgrade.

M3 is at the implementation/preflight stage. The spec package lives at
`docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-*`. Values files
are in `observability/`, scripts are executable, and `preflight-m3.sh` runs
cleanly against the live `k3d-openchoreo` cluster. hello-m2 OTEL hardening
and the `--extra-env` injection path are complete; the next step is the
first live `install-m3.sh` + `smoke-m3.sh --cluster` cycle to generate real
traces.

### Repository layout (current)

```
~/Projects/Sovereign/developer-portal/
+-- PROJECT_SUMMARY.md              this file
+-- SESSION_HANDOFF.md
+-- TODO.md
+-- README.md                       M1 + M2 sections
+-- backstage/                      M1 scaffold; M2 added /gitea-actions proxy
+-- docs/
|   +-- specs/
|   |   +-- m1-substrate/           three M1 spec docs
|   |   +-- m2-iac-cd/              three M2 spec docs (requirements, design,
|   |                               technical), committed 41db8f0
|   +-- superpowers/
|       +-- plans/
|           +-- 2026-04-10-rr-system-integration.md     (earlier)
|           +-- 2026-04-20-m2-iac-cd.md                 (M2 plan, 3677 lines,
|                                                      committed 3392c3e)
+-- iac/                             NEW in M2: OpenTofu
|   +-- versions.tf, backend.tf, providers.tf, variables.tf
|   +-- main.tf, outputs.tf, README.md
|   +-- modules/
|   |   +-- flux/
|   |   +-- gatekeeper/
|   |   +-- gitea-runner/
|   |   +-- openchoreo-environments/
|   |   +-- external-secrets-wiring/
|   +-- templates/
|       +-- ci.yaml                  canonical Gitea Actions workflow
+-- plugins/
|   +-- rr-policy-guards/            M1 plugin, extended in M2
|       +-- hooks/hooks.json         six mandatory no-bypass pretool guards
|       +-- plugin.json, README.md, .gitignore
|       +-- tools/
|       |   +-- emoji-guard/         M1
|       |   +-- bash-guard/          M1
|       |   +-- brew-guard/          M1
|       |   +-- tofu-guard/          OpenTofu lifecycle policy
|       |   +-- commit-guard/        staged-file and commit-message policy
|       |   +-- verify-guard/        commit and publication verification gate
|       +-- bin/
|           +-- rr-emoji-guard, rr-bash-guard, rr-brew-guard, rr-tofu-guard,
|               rr-commit-guard, rr-verify-guard
+-- tools/
|   +-- score2openchoreo/            NEW in M2
|       +-- go.mod (yaml.v3 + jsonschema/v5)
|       +-- types.go, convert.go, schema.go, cli.go, main.go
|       +-- convert_test.go, schema_test.go, main_test.go
|       +-- assets/score.schema.json (embedded)
|       +-- fixtures/                minimal + with-secret + invalid (Score)
|                                    + minimal + with-secret (golden Components)
|       +-- .gitignore (bin/)
+-- policies/                        NEW in M2
|   +-- C1-platform-addons-main-protected.rego + C1-constraint.yaml + C1-test.rego
|   +-- C2-score-schema-valid.rego + C2-constraint.yaml + C2-test.rego
|   +-- C3-infracost-delta.rego + C3-constraint.yaml + C3-test.rego
|   +-- README.md
+-- seed-repos/                      NEW in M2: pre-built content for Gitea repos
|   +-- platform-addons/             Flux-watched (kustomization + constraints)
|   +-- platform-config/             empty dev + staging env dirs
|   +-- hello-m2/                    demo app: main.go, Dockerfile, score.yaml,
|                                    catalog-info.yaml, .gitea/workflows/ci.yaml
+-- scripts/
    +-- install-m1.sh, teardown-m1.sh                       (M1)
    +-- install-m2.sh, teardown-m2.sh                       (M2)
    +-- smoke-m2.sh + 7 per-tool smokes                     (M2)
    +-- merge-tofu-hook-into-settings.sh, remove-tofu-...   (M2)
    +-- seed-openbao-m2-paths.sh, seed-gitea-repos.sh       (M2)
    +-- push-seed-content.sh, delete-m2-gitea-repos.sh      (M2)
    +-- ci/
        +-- post-infracost-comment.sh
        +-- commit-to-platform-config.sh
    +-- lib/                          M1 sourcing helpers (confirm colors.sh
                                     exists before running install-m2.sh)
```

### M2 test status at session close

| Suite | Count | Result |
|---|---|---|
| rr-tofu-guard unit + integration | 28 sub-tests | all PASS |
| score2openchoreo Convert | expanded | PASS |
| score2openchoreo schema | 2 | PASS |
| score2openchoreo golden + validate-only | 3 | PASS |
| Gatekeeper Rego | 6 | PASS (via `opa test --v0-compatible`) |
| OpenTofu modules | initialized + applied | Task 21 done; Flux watching both platform-addons AND platform-config (commit 42b2231 added the platform-config watch) |
| score2openchoreo live CRD dry-run | OpenChoreo resources | PASS 2026-05-17 via `kubectl apply --dry-run=server`; SecretReference emission added after CI exposed the missing-resource gap |
| End-to-end pipeline | DONE 2026-05-22 | CI run #24 (sha `5d88625`) succeeded; image pushed to local registry; platform-config commit applied by Flux; ReleaseBinding Ready=True; data-plane pod 1/1 Running |
| Backstage catalog smoke | PASS 2026-06-30 | Gitea catalog provider auto-discovers `openchoreo` org repos; `hello-m2` and `developer-portal` entities load; Playwright verifies guest sign-in and component links |
| Backstage dependency audit | PASS 2026-06-30 | High/critical advisories resolved via Yarn resolutions (`@grpc/grpc-js`, `ws`, `axios`, `undici`, `react-router`); only the moderate `@material-ui/core` v4 deprecation warning remains |

### M2 delta from locked-in tool list (canonical)

User's locked-in Integration & Delivery stack:
**Gitea + Backstage Software Catalog & Score + OpenTofu + Gitea Actions + in-cluster OCI registry**

The initial draft named Gitea OCI Registry. The implemented M2 image path uses
the dedicated `local-registry` module so Gitea Actions and k3s containerd have
stable in-cluster HTTP push/pull endpoints.

Added in M2 with explicit approval 2026-04-20:
- **Flux** (cluster add-ons drift correction only; OpenChoreo stays the
  workload deployer)

Pulled forward from M6 for M2 pipeline-scoped constraints (not runtime
policy at large):
- **OPA/Gatekeeper** (three ConstraintTemplates: C-1 main-protected, C-2
  Score schema valid, C-3 Infracost delta threshold)

Observability placement callout: **Infracost** is in the user's
Observability category alongside OpenCost and Cloud Custodian. M2 uses it
for pre-deploy cost estimation in Gitea Actions PR comments. Its
post-deploy dashboard role stays M3/M4.

### Outstanding (see TODO.md and SESSION_HANDOFF.md)

- **m2i-6: OpenBao dev-mode `inmem` storage** -- production-readiness item, not M2 closeout.
- **gitea-com push** -- 2026-05-23 retry reached `gitea.com` but failed authentication; refresh cloud credentials before retry.
- **M3 kickoff** -- specs added; next step is read-only preflight before any cluster mutation.

---

## Key cross-cutting decisions (recap)

- **Backstage role:** Front door. Runs on host via `yarn dev` for now.
  Containerization deferred to M3.
- **OpenChoreo role:** Load-bearing across three planes. Confirmed.
- **M2 GitOps controller:** Flux (not Argo CD). Confirmed 2026-04-20.
- **OpenTofu state backend:** native `kubernetes` backend, Secret in
  `tofu-state` namespace. openbao is for secrets, not state. Snapshotting
  state to openbao is a future milestone.
- **Environment model:** two OpenChoreo Environments (`dev`, `staging`)
  in the single k3d cluster. Promotion is a commit to `platform-config`.
- **Score rendering:** the Go `score2openchoreo` converter reads Score YAML
  and emits OpenChoreo `Component` + `SecretReference` + `Workload` CRDs as multi-document YAML.
  Score tools do not render raw Deployments in M2. The renderer defaults
  namespace/project to `default`, infers `deployment/service` when Score has
  service ports, infers `deployment/worker` otherwise, and accepts
  `pipeline.m2/component-type` as an override.
- **Pipeline runner:** Gitea Actions in-cluster (`act-runner` helm),
  label `ubuntu-latest` per operator convention.
- **Docker daemon:** still Colima at `~/.colima/default/docker.sock`.

---

## Skills used during this session

| Skill | Used for |
|---|---|
| superpowers:brainstorming | Shaping M2 scope before spec writing |
| superpowers:writing-plans | Producing the 24-task M2 implementation plan |
| superpowers:subagent-driven-development | Executing Tasks 1-20, 23, 24 |

`finishing-a-development-branch` will be invoked in the next session after
Tasks 21-22 complete.
