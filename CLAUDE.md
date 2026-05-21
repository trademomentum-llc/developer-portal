# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Umbrella for a self-hosted Internal Developer Platform (IDP) broken into seven milestones (M1-M7), modeled on the Platform Engineering community reference architecture. M1 (substrate) is complete and healthy; M2 (IaC + CD loop) is code-complete but the live `tofu apply` run is blocked (see `TODO.md` "M2 install blockers" and memory `project_m2_install_blockers`).

The repo depends on two sibling checkouts -- do not expect to operate without them:
- `~/Projects/openchoreo/` -- upstream platform orchestrator. Provides the shared `k3d-openchoreo` cluster. Five planes live there: control-plane, data-plane, observability-plane, workflow-plane (Argo Workflows, NOT Argo CD), plus cert-manager / external-secrets / gitea / openbao.
- `~/Projects/rational-reserve/` -- AI swarm orchestration layer. Integration into the portal is deferred to M7.

## First thing to do in a new session

Read these three files in order -- they carry the real state that git alone does not convey:
1. `SESSION_HANDOFF.md` -- what the previous session ended on, what is different.
2. `PROJECT_SUMMARY.md` -- snapshot of all three sibling projects and cross-cutting decisions.
3. `TODO.md` -- prioritized action list with dependency order and tech-debt backlog.

The user actively maintains these. Keep them current whenever you land changes to scope, blockers, or milestone state (this is a hard project rule in `~/Projects/CLAUDE.md`).

## Commands

### Policy guards (Go, `plugins/rr-policy-guards/tools/<guard>/`)

Each guard is a single static binary, stdlib-only. Build outputs live in `plugins/rr-policy-guards/bin/` (gitignored).

```
# Per-guard test + build (run from the guard's own dir)
cd plugins/rr-policy-guards/tools/<emoji|bash|brew|tofu>-guard
go test ./...
go build -o ../../bin/rr-<name>-guard .
```

`scripts/install-m1.sh` rebuilds the M1 guards (emoji, bash, brew); `scripts/install-m2.sh` rebuilds `rr-tofu-guard`.

### score2openchoreo (Go, `tools/score2openchoreo/`)

```
cd tools/score2openchoreo
go test ./...
go build -o bin/score2openchoreo .
```

Dependencies: `yaml.v3`, `jsonschema/v5`. The Score JSON schema is embedded via `//go:embed assets/score.schema.json`.

### Gatekeeper Rego policies (`policies/`)

```
opa test --v0-compatible policies/*.rego -v
```

The `--v0-compatible` flag is required for OPA 1.x because these policies use Rego v0 syntax (what Gatekeeper 3.17 still expects). Scope the glob to `*.rego` only -- the constraint YAMLs break OPA's bundle loader. Expected: 6/6 PASS.

### Backstage (`backstage/`, yarn 4.4.1 via Corepack, Node 22/24)

```
cd backstage
yarn install
yarn dev                        # run frontend + backend on host
yarn tsc                        # typecheck
yarn test                       # backstage-cli repo test
yarn lint                       # lints since origin/master
yarn test:e2e                   # playwright
```

Dev server runs on host (not containerized in M1/M2). Use `scripts/start-backstage.sh` / `scripts/stop-backstage.sh` for the managed-process variant.

### OpenTofu (`iac/`)

Do NOT run `tofu apply` / `destroy` / `import` directly from a Bash tool use -- `rr-tofu-guard` blocks it. Use `scripts/install-m2.sh` instead, which sets `RR_TOFU_GUARD_BYPASS=1` for its own invocation. Version constraints: OpenTofu `>= 1.9.0, < 1.12.0` (see `iac/versions.tf`). Backend is the native `kubernetes` backend with a Secret in the `tofu-state` namespace on the `k3d-openchoreo` context.

Plan-only and init are allowed:
```
cd iac
tofu init -reconfigure
tofu plan
```

### M1 / M2 lifecycle scripts

```
./scripts/install-m1.sh           # resumable (checkpointed in ~/.rational-reserve/m1-progress/)
./scripts/install-m1.sh --fresh   # wipe checkpoints, start over
./scripts/teardown-m1.sh

./scripts/install-m2.sh           # linear, not checkpointed (short script)
./scripts/smoke-m2.sh             # wraps 7 per-tool smokes: tofu, actions, flux, score,
                                  #                          infracost, gatekeeper, openbao
./scripts/teardown-m2.sh
```

Do not invent a "clean slate" path. If an install script fails mid-way, the blockers documented in `TODO.md` are the canonical remediation list.

## Architecture

### Policy guards as a live constraint

`plugins/rr-policy-guards/` is a Claude Code plugin registered in `~/.claude/settings.json`. Four PreToolUse hooks are currently active:

| Guard | Matcher | Blocks |
|---|---|---|
| rr-emoji-guard | Write \| Edit \| MultiEdit | Any non-ASCII byte (> 0x7F) or invalid UTF-8 in file content |
| rr-bash-guard  | Bash | Bare `$VAR` expansion without quoting; suggests safe syntax |
| rr-brew-guard  | Bash | Dangerous brew flags, URL installs, untrusted taps |
| rr-tofu-guard  | Bash | `tofu apply \| destroy \| import` only; a fast `IsTofuCommandPrefix` gate skips the quote-aware tokenizer for non-tofu commands so heredocs with apostrophes or the word "tofu" in data no longer false-positive |

Bypass env vars exist per-guard (`RR_EMOJI_GUARD_BYPASS=1`, `RR_TOFU_GUARD_BYPASS=1`, ...); every use is still audit-logged to `~/.rational-reserve/logs/<guard>.jsonl`. Hooks load at session start, so changes require a session restart to take effect.

Two consequences when working in this repo:
1. All file writes must be pure ASCII. No em dashes, smart quotes, box-drawing characters, arrows, check marks, or emoji -- use `--` / `'` / `+-|` / `->` / ASCII text.
2. Infrastructure state changes flow through install scripts, not ad-hoc `tofu` invocations.

### M2 GitOps and CD loop (why the modules exist)

M2 wires a push-to-deploy loop on top of the M1 substrate. The developer path:

1. Push to `openchoreo/hello-m2` (Gitea, org `openchoreo`).
2. Gitea Actions runner (`act-runner` helm, label `ubuntu-latest`) picks up `.gitea/workflows/ci.yaml`.
3. CI validates Score YAML against its schema, runs `tofu plan` + Infracost, builds and pushes image to the in-cluster local registry.
4. `score2openchoreo` renders OpenChoreo Component and Workload CRDs from the Score file (it does NOT emit raw Deployments; that is the locked-in design).
5. CI commits the rendered Component and Workload into `openchoreo/platform-config/environments/dev/`.
6. OpenChoreo reconciles the Component and Workload into a running pod.
7. Promotion is a commit that copies the same rendered resources into `environments/staging/`.

Only two OpenChoreo `Environment`s exist in the single k3d cluster: `dev` and `staging`.

Root `iac/main.tf` composes five modules:
- `flux/` -- cluster add-ons drift correction only; OpenChoreo remains the workload deployer.
- `gatekeeper/` -- pulls forward three pipeline-scoped constraints from M6 (C-1 main-protected, C-2 Score schema valid, C-3 Infracost delta threshold). Broader runtime policies stay M6.
- `gitea-runner/` -- the Actions runner; `depends_on = [module.flux]`.
- `openchoreo-environments/` -- the two Environment CRDs.
- `external-secrets-wiring/` -- wires OpenBao into `external-secrets` for per-app secret delivery.

### Locked-in tool list (canonical) vs M2 additions

The user's Integration & Delivery stack is fixed: **Gitea + Backstage Software Catalog & Score + OpenTofu + Gitea Actions + an OCI registry**. Roadmap tables in `PROJECT_SUMMARY.md` are drafts -- never treat their rows as decided. The implemented M2 image path uses the dedicated in-cluster `local-registry` module. Two explicit additions for M2 are:
- Flux (cluster add-ons drift only; explicit approval 2026-04-20)
- OPA/Gatekeeper (pipeline-scoped only, pulled forward from M6)

Infracost lives in Observability; in M2 it is used pre-deploy for PR cost comments. Its post-deploy dashboard role is deferred to M3/M4.

### Seed repos (`seed-repos/`)

Pre-built content pushed into three Gitea repos under org `openchoreo`:
- `platform-addons/` -- Flux-watched kustomization + Gatekeeper constraints.
- `platform-config/` -- empty `environments/{dev,staging}/` directories that CI commits into.
- `hello-m2/` -- the demo app: `main.go`, `Dockerfile`, `score.yaml`, `catalog-info.yaml`, `.gitea/workflows/ci.yaml`.

`scripts/seed-gitea-repos.sh` creates the repos; `scripts/push-seed-content.sh` pushes the content; `scripts/delete-m2-gitea-repos.sh` tears them down. Gitea normally runs at `http://localhost:3333` via port-forward on the `k3d-openchoreo` cluster.

### Score -> OpenChoreo conversion

`tools/score2openchoreo/` reads a Score YAML and emits OpenChoreo Component and Workload CRDs as multi-document YAML. Layout:
- `types.go` -- Score and OpenChoreo resource shared structs.
- `convert.go` -- `Convert(doc ScoreDocument, opts ConvertOptions) ([]OpenChoreoResource, error)`; deterministic sort order for variables and endpoints.
- `schema.go` -- `ValidateScore([]byte) error` using the embedded JSON schema in `assets/score.schema.json`.
- `cli.go` / `main.go` -- flags, stdin/file input, stdout YAML output, `--validate-only` short-circuit.
- `fixtures/` -- Score inputs (minimal, with-secret, invalid) and golden Component outputs.

Known gaps: `${resources.X.Y}` inline substitution is not implemented (see score-1 in TODO.md); secret-name fallback "X-secret" convention is undocumented (score-3); schema pin is currently `main` instead of a git SHA (score-5).

## Conventions

- **Docs per module.** The project rule (`~/Projects/CLAUDE.md`) requires a Requirements, Design Spec, and Technical Spec for every module/system/tool/script/plugin. M1 and M2 both have three-doc packages in `docs/specs/`.
- **Deterministic first.** Prefer deterministic logic; fall back to interpreted/scripted only when necessary.
- **No emojis anywhere.** Enforced by `rr-emoji-guard`. Blank out any temptation to use them.
- **No direct `tofu apply` from a Bash tool use.** Use install scripts. Plan and init are fine.
- **Runner label convention:** workflows use `runs-on: ubuntu-latest` even on the self-hosted `act-runner` (see memory `project_runner_labels`).
- **Commit discipline:** all commits land on `main`; many commits may be unpushed pending remote-resolution decisions in TODO.md. Check `git log origin/main..HEAD` before assuming anything is published.
