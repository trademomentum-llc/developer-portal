# PROJECT SUMMARY

> A snapshot of the three projects this session touched and their current
> state. For "what to do next" see TODO.md. For "where we stopped and what
> changed mid-session" see SESSION_HANDOFF.md.

**Snapshot date:** 2026-04-21 (late -- reflects 35 commits ahead of origin)

---

## Overview

The user is building a self-hosted Internal Developer Platform (IDP) on
their local machine, using the Platform Engineering community reference
architecture in `~/Downloads/platform.pptx` as the blueprint. The platform
is decomposed into seven milestones (M1 through M7); M1 substrate is
complete and M2 (IaC + CD loop) is functionally code-complete with the
two live-execution tasks pending operator action.

Three projects on disk are involved, in dependency order:

1. `~/Projects/openchoreo/`        -- the platform orchestrator (upstream OSS)
2. `~/Projects/rational-reserve/`  -- AI swarm orchestration layer (custom)
3. `~/Projects/developer-portal/`  -- the umbrella IDP build (this repo)

---

## 1. openchoreo

**Source:** https://github.com/openchoreo/openchoreo (cloned at `~/Projects/openchoreo/`)

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

**Source:** `~/Projects/rational-reserve/` (built earlier from requirement
docs on an external volume).

**Role in the IDP:** Developer Control Plane, Copilots/Agents/LLM slot.
Military-hierarchy AI swarm orchestration with 12 ranks, 14 MOS, SITREP
/ FRAGO / CASREP / AAR protocols.

**Build status:** v0.1 (Spine) + v0.2 (Adapter layer) COMPLETE. Unchanged
this session. 65 tests pass. Daemon + MCP shim + doctrine loader + SQLite
state store all working. Integration into the portal is deferred to M7.

---

## 3. developer-portal (current focus)

**Source:** `~/Projects/developer-portal/`. git repository, branch `main`,
35 commits ahead of `origin/main` and not pushed (2 M1 follow-ons + 25 M2
plan commits + 8 post-handoff commits for tech debt fixes + doc updates).

**Role:** The umbrella for the full self-hosted IDP build.

**Build status:** M1 substrate complete and healthy. **M2 codegen complete**
as of this session (22 of 24 plan tasks landed; Tasks 21-22 are live
cluster runs the operator drives manually).

### Repository layout (current)

```
~/Projects/developer-portal/
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
|       +-- hooks/hooks.json         emoji-guard + new Bash matcher for tofu-guard
|       +-- plugin.json, README.md, .gitignore
|       +-- tools/
|       |   +-- emoji-guard/         M1
|       |   +-- bash-guard/          M1
|       |   +-- brew-guard/          M1
|       |   +-- tofu-guard/          NEW in M2: go.mod, parser.go, parser_test.go,
|       |                             audit.go, audit_test.go, main.go, main_test.go
|       +-- bin/
|           +-- rr-emoji-guard, rr-bash-guard, rr-brew-guard, rr-tofu-guard
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
| score2openchoreo Convert | 6 | PASS |
| score2openchoreo schema | 2 | PASS |
| score2openchoreo golden + validate-only | 3 | PASS |
| Gatekeeper Rego | 6 | PASS (via `opa test --v0-compatible`) |
| OpenTofu modules | n/a | not yet init'd; awaits tofu binary install in Task 21 |
| End-to-end pipeline | n/a | Task 22, pending Task 21 |

### M2 delta from locked-in tool list (canonical)

User's locked-in Integration & Delivery stack:
**Gitea + Backstage Software Catalog & Score + OpenTofu + Gitea Actions + Gitea OCI Registry**

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

- Task 21 live install run (operator-driven)
- Task 22 live pipeline verification (operator-driven, after 21)
- Tech debt list: 5 guard issues, 5 score2openchoreo issues, 1 Gatekeeper
  doc issue, 2 install-script issues
- Push of the 27 local commits to a remote

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
- **Score rendering:** a new Go `score2openchoreo` converter reads Score
  YAML and emits OpenChoreo Component CRDs. Score tools do not render
  raw Deployments in M2.
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
