# M1 Substrate -- Requirements Document

> **Milestone:** M1 -- Substrate (first of seven)
> **Version:** 1.0
> **Date:** 2026-04-09
> **Status:** Draft, awaiting user approval

---

## 1. Purpose

M1 establishes the foundational layer of a self-hosted Internal Developer Platform (IDP) for local development. It exists so that every subsequent milestone (M2 through M7) has a running platform to extend. Nothing else in the roadmap can start until M1 is operational.

This document captures what M1 must do. It does not describe how -- that is the Design Specification's job.

## 2. Context

The user is building an IDP modeled on the Platform Engineering community reference architecture in `/Users/nnos/Downloads/platform.pptx`. The user selected a self-hosted, open-source-first stack spanning five planes (Observability, Developer Control, Integration & Delivery, Platform Orchestration, Security). The full build has been decomposed into seven milestones (M1-M7); M1 is the substrate on which the rest lands.

## 3. Stakeholders

| Stakeholder | Role | Interest in M1 |
|---|---|---|
| User (platform owner) | Sole decision maker | Fast time-to-running platform, then iterates with team |
| Team members (future) | Developers consuming the IDP | Frictionless "clone developer-portal, run install script, have platform" |
| AI coding agents (Claude Code, Codex CLI, Qwen-Code, OpenCode, Mistral Vibe) | First-class consumers (per user directive) | Same frictionless path as human team; future MCP integration in M7 |
| Rational Reserve (swarm orchestration, built earlier this session) | Future consumer | Will integrate in M7 via MCP |

## 4. Functional Requirements

### 4.1 Runtime environment

- **FR-1** The system SHALL run entirely on a local k3d Kubernetes cluster backed by the user's existing Colima docker daemon.
- **FR-2** The system SHALL NOT require any cloud account, public DNS, or external paid service.
- **FR-3** The cluster SHALL be provisioned in a way that does not conflict with the user's existing OpenChoreo development workflow at `/Users/nnos/Projects/openchoreo/`.

### 4.2 OpenChoreo

- **FR-4** OpenChoreo SHALL be deployed into the cluster with its controller, API, observer, cluster-gateway, and cluster-agent pods reaching `Ready` state.
- **FR-5** OpenChoreo's API SHALL be reachable from host processes (specifically Backstage running on host) via a stable port or in-cluster service name.
- **FR-6** OpenChoreo deployment SHALL use the project's own published deployment path (its `make` targets and helm charts), not a custom fork.

### 4.3 Gitea

- **FR-7** Gitea SHALL be deployed into the cluster with a web UI reachable at a stable local URL.
- **FR-8** Gitea SHALL be provisioned with at least one admin user via an initial configuration step.
- **FR-9** The operator SHALL be able to create at least one demo repository containing a `catalog-info.yaml` file that Backstage can ingest.
- **FR-10** Gitea SHALL expose an API that Backstage can query for repository discovery.

### 4.4 Backstage

- **FR-11** A Backstage skeleton application SHALL be scaffolded into the repository at `/Users/nnos/Projects/developer-portal/backstage/`.
- **FR-12** The Backstage app SHALL run via `yarn dev` on the host (not containerized in M1) and serve its UI at `http://localhost:3000`.
- **FR-13** Backstage's software catalog SHALL discover and list components from Gitea via the official `@backstage/plugin-catalog-backend-module-gitea` plugin.
- **FR-14** Backstage SHALL have a proxy configuration entry pointing at OpenChoreo's API, such that `GET /api/proxy/openchoreo/health` (or equivalent) returns a non-error response.

### 4.5 Brew security pre-command hook

- **FR-15** A PreToolUse Bash hook SHALL be registered in `~/.claude/settings.json` that validates every `brew` command before execution.
- **FR-16** The hook SHALL reject brew commands matching any of the following patterns:
  - URL-based installs (argument matches `^https?://` or `^git://`)
  - Dangerous flags not on an allow-list (specifically: `--force`, `--HEAD`, `--debug-symbols`, `--build-from-source`)
  - Tap additions from untrusted sources (any `brew tap` invocation pointing at a URL or a tap not on an explicit allow-list)
  - Shell metacharacters suggesting command chaining (`;`, `&&`, `||`, `|`, backtick, `$(...)`, `>`, `<`, `&`)
  - Package names not matching `^[a-z0-9][a-z0-9_-]*(@[0-9.]+)?$`
- **FR-17** The hook SHALL allow common safe flags (`--quiet`, `--no-auto-update`, `--formula`, `--cask`).
- **FR-18** The hook SHALL exit with code 0 on allow and exit code 2 on block (the exit code Claude Code's PreToolUse protocol uses to block tool execution).
- **FR-19** The hook SHALL support an override environment variable `RR_BREW_GUARD_BYPASS=1` for emergency use, which SHALL log the bypass to a local audit file.

### 4.6 Operational commands

- **FR-20** A single `scripts/install-m1.sh` command SHALL provision the full M1 stack from an empty state, assuming the prerequisites in Section 6 are met.
- **FR-21** A single `scripts/teardown-m1.sh` command SHALL remove the k3d cluster, uninstall the Gitea helm release, and stop any host-side Backstage process, leaving the host filesystem cleaned up except for the source tree itself.
- **FR-22** A `README.md` at the repository root SHALL describe in under 300 words how to install, use, and tear down M1.

## 5. Non-Functional Requirements

### 5.1 Performance

- **NFR-1** Cold-start time from empty state to a reachable Backstage UI on the operator's machine SHALL NOT exceed 15 minutes, excluding first-run npm package downloads which may take an additional 10 minutes.
- **NFR-2** The M1 stack SHALL fit within a Colima VM sized at 8 GB RAM / 4 CPU / 40 GB disk.
- **NFR-3** The brew guard hook SHALL complete within 100 ms per invocation on an Apple Silicon laptop.

### 5.2 Determinism

- **NFR-4** Components with a real language choice SHALL use deterministic (compiled) languages. Specifically: the brew guard hook SHALL be written in Go, not Bash or an interpreted language.
- **NFR-5** Interpreted languages are permitted only where the target ecosystem forces them. Backstage is permitted to be TypeScript/Node.js because Backstage is TypeScript. This exception SHALL be documented explicitly in the Technical Specification.
- **NFR-6** All install and teardown scripts SHALL be idempotent; running them twice in succession SHALL produce the same end state.

### 5.3 Security

- **NFR-7** No credentials (API tokens, passwords, private keys) SHALL be committed to the source tree.
- **NFR-8** The Gitea admin password SHALL be generated at install time and written to a file outside the source tree (`~/.rational-reserve/m1-gitea-admin-password` or equivalent), with file permissions 0600.
- **NFR-9** The Backstage proxy configuration SHALL NOT expose cluster-internal services that bypass OpenChoreo's own auth layer.
- **NFR-10** Every `brew install` SHALL pass through the brew guard hook defined in FR-15 through FR-19.

### 5.4 Observability

- **NFR-11** M1 SHALL emit no telemetry by default (observability is M3's responsibility). Logs produced by the installed components SHALL be inspectable via `kubectl logs` and `yarn dev` stdout only.
- **NFR-12** The brew guard hook SHALL log every blocked command to a local audit file (`~/.rational-reserve/logs/brew-guard.jsonl`) as structured JSON.

### 5.5 Portability

- **NFR-13** M1 SHALL work on Apple Silicon (aarch64) macOS running Colima. Linux and Intel macOS are out of scope for M1 but SHALL NOT be actively excluded by design choices that would be hard to reverse later.

### 5.6 Documentation

- **NFR-14** Every source file in `tools/rr-brew-guard/` SHALL have a file-level comment explaining its responsibility in under 10 lines.
- **NFR-15** The three specification documents (Requirements, Design, Technical) SHALL be kept in sync during implementation; any deviation from spec during build SHALL be reflected back into the specs before M1 is declared done.

## 6. Prerequisites (required before M1 begins)

The following must already be installed and working on the target machine:

| Prerequisite | Minimum version | Verified in session | Installer if missing |
|---|---|---|---|
| macOS | 14+ (Darwin 25+) |  Darwin 25.4.0 | -- |
| Colima | 0.6+ |  running, socket at `~/.colima/default/docker.sock` | `brew install colima` |
| docker CLI | 23.0-30.0 |  29.4.0 | `brew install docker` |
| k3d | 5.0+ |  5.8.3 | `brew install k3d` |
| kubectl | 1.31-1.36 |  1.35.3 | `brew install kubectl` |
| helm | 3.16-3.21 |  3.20.1 | `brew install helm@3` |
| Go | 1.21+ |  1.26.2 | `brew install go` |
| Node.js | 20+ |  25.9.0 | `brew install node` |
| **yarn** (NEW) | 1.22+ or Yarn 4 via corepack | X missing | `brew install yarn` (gated by brew guard hook built in Task 0) |
| kubebuilder | 4.3-4.14 |  4.13.1 | `brew install kubebuilder` |
| jq | 1.6+ | verify at install time | `brew install jq` (gated by brew guard hook) |
| yq | 4.0+ | verify at install time | `brew install yq` (gated by brew guard hook) |
| openssl | (macOS default) |  pre-installed | -- |

**Note on yarn:** yarn is the only missing prerequisite. Its installation is Task 0 of M1 and is gated by the brew guard hook M1 itself builds. This is intentional -- the hook goes in first, then the hook validates the yarn install.

## 7. Acceptance Criteria

M1 is complete when **all** of the following are true:

- [ ] `rr-brew-guard` binary builds, all unit tests pass, and the hook is registered in `~/.claude/settings.json`
- [ ] `brew install yarn` succeeds (validated by the hook) and yarn is on PATH at a version >= 1.22
- [ ] `k3d cluster list` shows the developer-portal cluster running
- [ ] `kubectl get pods -n openchoreo-system` shows all pods in `Ready` state
- [ ] `helm list -n gitea` shows a deployed Gitea release
- [ ] Gitea UI is reachable at `http://localhost:3002` (or whichever NodePort is assigned), admin login works
- [ ] A demo repository exists in Gitea containing a `catalog-info.yaml` file
- [ ] Backstage UI is reachable at `http://localhost:3000`
- [ ] Backstage's catalog page lists the demo Gitea repository as a registered component
- [ ] `curl http://localhost:3000/api/proxy/openchoreo/health` (or equivalent route) returns a non-500 HTTP response
- [ ] `scripts/teardown-m1.sh` returns the system to a clean pre-install state in under 60 seconds
- [ ] `README.md` exists at the repo root and describes install/use/teardown in under 300 words
- [ ] All three spec documents (Requirements, Design Spec, Technical Spec) are present, checked in, and internally consistent

## 8. Out of Scope (explicitly deferred)

The following are **not** M1 work and SHALL NOT be implemented as part of this milestone, even if tempting:

| Item | Deferred to |
|---|---|
| Score template scaffolders in Backstage | M2 |
| Actual commit -> deploy flow through OpenChoreo | M2 |
| OpenTofu IaC integration | M2 |
| Infracost (pre-deploy cost estimation) | M2 |
| OpenTelemetry Collector + SigNoz instrumentation | M3 |
| Real auth (Keycloak, Gitea OAuth integration with Backstage) | M3+ |
| OpenCost (runtime cost attribution) | M4 |
| Cilium CNI + Envoy Gateway | M4 |
| RabbitMQ or Kafka (with OpenResty front-door) | M5 |
| Cloud Custodian + MISP + TheHive/Cortex/Velociraptor suite | M6 |
| OPA/Gatekeeper policy layer (if not delivered via OpenChoreo CRDs) | M6 |
| MCP plugin surfacing Backstage catalog and scaffolder to AI agents | M7 |
| Rational Reserve <-> OpenChoreo integration | M7 |
| Per-agent authentication tokens for Gitea | M7 |
| SSL / TLS / production hardening | Never (this is local dev) |
| Linux or Intel macOS support | Never for M1, revisit if team membership changes |

## 9. Assumptions

- **A-1** The operator has write access to `/Users/nnos/Projects/developer-portal/` and `~/.claude/settings.json`.
- **A-2** The Colima VM has sufficient resources (per NFR-2). If not, the operator will be instructed to resize it as part of Task 1.
- **A-3** The operator's Claude Code session loads `~/.claude/settings.json` -- this is Claude Code's default behavior.
- **A-4** OpenChoreo's `make quick-start.dev` target works as documented on the operator's machine. If it does not, M1 falls back to a manual helm install path described in the Technical Spec.
- **A-5** Gitea's helm chart at `gitea-charts/gitea` is compatible with k3d-provisioned Kubernetes 1.33.

## 10. Glossary

- **IDP** -- Internal Developer Platform
- **M1-M7** -- The seven milestones of the user's platform build, of which M1 is the first
- **Substrate** -- The foundational layer (k3d + OpenChoreo + Gitea + Backstage skeleton) that later milestones extend
- **PreToolUse hook** -- Claude Code's mechanism for validating tool invocations before execution; exit code 2 blocks, exit code 0 allows
- **brew guard** -- The small Go binary this milestone introduces, registered as a PreToolUse hook for Bash tools
- **Score** -- Workload specification standard from ScoreSpec; out-of-scope for M1, in-scope for M2
- **Rational Reserve / RR** -- The military-hierarchy AI swarm orchestration system built earlier in this session; out-of-scope for M1, in-scope for M7
