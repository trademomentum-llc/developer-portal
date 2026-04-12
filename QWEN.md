# developer-portal -- QWEN.md

## Project Overview

Self-hosted Internal Developer Platform (IDP) built on a k3d Kubernetes cluster, Gitea, and Backstage. This is an umbrella project for a seven-milestone roadmap (M1-M7) modeled on the Platform Engineering community reference architecture. The platform spans five planes: Observability, Developer Control Plane, Integration & Delivery, Platform Orchestration, and Security.

**Current milestone:** M1 (Substrate) -- specs written, implementation in progress.

**Key relationships:**
- `~/Projects/openchoreo/` -- upstream platform orchestrator. M1 reuses its `openchoreo-cluster` (already running).
- `~/Projects/rational-reserve/` -- AI swarm orchestration layer (built separately, v0.1/v0.2 complete)

## Directory Structure

```
developer-portal/
+-- README.md                       Quick start guide
+-- PROJECT_SUMMARY.md              Three-project snapshot
+-- SESSION_HANDOFF.md              Session state handoff
+-- TODO.md                         Prioritized action list
+-- catalog-info.yaml               Backstage catalog descriptors
+-- backstage/                      Backstage app (scaffolded with create-app)
+-- docs/
|   +-- specs/m1-substrate/         M1 Requirements, Design Spec, Technical Spec
|   +-- superpowers/plans/          Implementation plans go here
+-- plugins/
|   +-- rr-policy-guards/           Claude Code PreToolUse policy hooks (Go)
|       +-- bin/                    Compiled guard binaries
|       +-- tools/emoji-guard/      ASCII enforcement guard
|       +-- tools/bash-guard/       Bash $VAR expansion guard
|       +-- tools/brew-guard/       brew install security guard
+-- scripts/                        Install, teardown, and helper scripts
|   +-- install-m1.sh               M1 substrate installer (checkpointed)
|   +-- teardown-m1.sh              M1 teardown
|   +-- gitea-values.yaml           Gitea helm values
|   +-- setup-*.sh                  Gitea and Backstage helpers
```

## Technologies

- **Runtime:** k3d (Kubernetes), Colima (Docker daemon)
- **Git:** Gitea (helm-deployed into k3d)
- **Portal:** Backstage (TypeScript/Node.js, runs on host via `yarn dev`)
- **Policy guards:** Go (stdlib only, static binaries)
- **Package manager:** yarn (v4.4.1 via Corepack)
- **Build tooling:** Go 1.21+, Node 22/24, helm v3

## Building and Running

### Prerequisites

- Colima running: `colima start --cpu 4 --memory 8`
- Installed: k3d, kubectl, helm, node, yarn, Go 1.21+

### Install M1 Substrate

```bash
./scripts/install-m1.sh          # resume from last checkpoint
./scripts/install-m1.sh --fresh  # wipe checkpoints and start over
```

The installer is checkpointed (uses `~/.rational-reserve/m1-progress/*.done` files). It runs these tasks in order:

0. Build and register policy guard hooks (emoji-guard, bash-guard, brew-guard)
1. Install yarn if missing
2. Verify openchoreo-cluster exists and switch kubectl context to it (requires OpenChoreo running at `~/Projects/openchoreo/`)
3. Install Gitea via helm into the openchoreo-cluster
4. Port-forward Gitea to localhost:3002
5. Port-forward OpenChoreo API to localhost:9090
6. Create demo repo in Gitea with catalog-info.yaml
7. Scaffold Backstage and wire to Gitea
8. Start Backstage on localhost:3000

### Teardown

```bash
./scripts/teardown-m1.sh
```

Stops all services, deletes the k3d cluster, removes credentials. Preserves source code and audit logs.

### Services (when running)

| Service | URL | Notes |
|---------|-----|-------|
| Backstage | http://localhost:3000 | Frontend + backend on host |
| Gitea | http://localhost:3002 | Git hosting, admin: gitea_admin |
| OpenChoreo API | http://localhost:9090 | Port-forwarded from openchoreo-cluster |
| k3d cluster | -- | Name: openchoreo-cluster (shared with ~/Projects/openchoreo/) |

### Backstage Development

```bash
cd backstage
yarn install
yarn start          # starts on localhost:3000 with hot-reload
```

### Policy Guards

All three guards are Go binaries in `plugins/rr-policy-guards/bin/`:

- **rr-emoji-guard** -- blocks non-ASCII in file writes (Write/Edit/MultiEdit)
- **rr-bash-guard** -- blocks bare $VAR expansion, suggests safe syntax
- **rr-brew-guard** -- blocks dangerous brew flags, URL installs, untrusted taps

To build and test:

```bash
cd plugins/rr-policy-guards/tools/<guard>
go test ./...
go build -o ../../bin/<guard> .
```

## Development Conventions

1. **Plain ASCII in all files.** Absolute rule. No emojis, em dashes, smart quotes, box-drawing characters, or any non-ASCII byte. The emoji-guard hook enforces this at PreToolUse time (exit 2 on violation).
2. **Deterministic / compiled languages preferred.** Use Go for new tools, scripts, hooks, and services. Use interpreted languages only when the ecosystem forces it (e.g., Backstage is TypeScript).
3. **Three-document plan format.** Before any non-trivial implementation, produce: Requirements Document, Design Specification, Technical Specification. Then an Implementation Plan (TDD bite-sized tasks) after user approval.
4. **TDD practices.** Policy guards include both unit tests (table-driven) and integration tests. All tests must pass before a binary is considered ready.
5. **Checkpointed installs.** The `install-m1.sh` script uses progress files so it can be interrupted and resumed. Use `--fresh` to start over.
6. **No non-ASCII characters.** Use `--` for em dashes, `->` for arrows, straight quotes, and write words instead of symbols for check marks or cross marks.

## M1-M7 Roadmap

| Milestone | Scope |
|-----------|-------|
| M1 | Substrate -- k3d + OpenChoreo + Gitea + Backstage skeleton |
| M2 | IaC + CD loop -- OpenTofu, Gitea Actions, Argo-style GitOps, Score templates, Infracost |
| M3 | Observability -- OpenTelemetry Collector, SigNoz, instrumentation |
| M4 | Cost + mesh -- OpenCost, Cilium, Envoy Gateway |
| M5 | Messaging -- RabbitMQ or Kafka with OpenResty front-door |
| M6 | Security suite -- OPA/Gatekeeper, MISP, TheHive + Cortex + Velociraptor, Cloud Custodian |
| M7 | Agent integration -- Backstage MCP plugin, RR <-> OpenChoreo wiring, per-agent Gitea tokens |

## User's Locked-in Tool Choices

- **Observability:** OpenTelemetry + SigNoz + (Infracost + OpenCost + Cloud Custodian)
- **Dev Control Plane:** VS Code/Cursor + named agents + OpenChoreo
- **Integration & Delivery:** Gitea + Backstage Catalog & Score + OpenTofu + Gitea Actions + Gitea OCI Registry
- **Platform Orchestration:** OpenChoreo + Cilium & Envoy Gateway + (RabbitMQ/Kafka + OpenResty)
- **Security:** OpenChoreo via Backstage + Cilium & Envoy Gateway
- **SOC stack (M6):** TheHive + Cortex + Velociraptor

## Important Notes

- **Docker daemon:** Provided by Colima at `~/.colima/default/docker.sock`. Docker Desktop is NOT installed.
- **Helm:** User uses helm v3 (3.20.1). Helm v4 is installed but not on PATH.
- **Backstage runs on host** (not in cluster) during M1 to preserve hot-reload. Containerization deferred to a later milestone.
- **OpenChoreo cluster:** M1 uses the existing `openchoreo-cluster` from `~/Projects/openchoreo/`. It does NOT create its own cluster. Teardown does not delete it.
- **yarn version:** 4.4.1 via Corepack (packageManager field in package.json).
- **Node versions:** 22 or 24 (specified in backstage/package.json engines).
