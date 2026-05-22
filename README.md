# Developer Portal -- Internal Developer Platform

Self-hosted Internal Developer Platform (IDP) built on k3d, Gitea,
Backstage, OpenTofu, Flux, Gatekeeper, and OpenChoreo. M1 substrate is
complete. M2 is validated through OpenChoreo Pod creation and is in closeout.

## Prerequisites

- Docker (via Colima): `colima start --cpu 4 --memory 8`
- k3d, kubectl, helm, node, yarn (brew-installable)
- Go 1.21+ (for building policy guard hooks)
- OpenChoreo cluster must be running (see `~/Projects/openchoreo/`)

## Install

```
./scripts/install-m1.sh
```

Resumes from the last completed task. Use `--fresh` to start over.

## Services

| Service | URL | Notes |
|---------|-----|-------|
| Backstage | http://localhost:3000 | Frontend + backend |
| Gitea | http://localhost:3333 | Git hosting, admin: gitea_admin |
| OpenChoreo API | http://localhost:9090 | Port-forwarded from k3d-openchoreo |
| k3d cluster | -- | Name: k3d-openchoreo (shared) |

## Teardown

```
./scripts/teardown-m1.sh
```

Stops all services, deletes the k3d cluster, removes credentials. Preserves source code and audit logs.

## Policy Guards

Three PreToolUse hooks enforce safety in Claude Code sessions:

- **rr-emoji-guard** -- blocks non-ASCII in file writes
- **rr-bash-guard** -- blocks bare $VAR expansion, suggests safe syntax
- **rr-brew-guard** -- blocks dangerous brew flags, URL installs, untrusted taps

All hooks are Go binaries in `plugins/rr-policy-guards/bin/`.

## Project Structure

- `docs/specs/m1-substrate/` -- Requirements, Design Spec, Technical Spec
- `plugins/rr-policy-guards/` -- Policy guard hooks (Go)
- `scripts/` -- Install, teardown, and helper scripts
- `backstage/` -- Backstage app (scaffolded, wired to Gitea)

## M2 -- IaC + CD loop

M2 lands Flux (cluster add-ons drift correction), Gatekeeper (three pipeline
constraints), and the Gitea Actions runner. The developer path: push to
`openchoreo/hello-m2`, CI validates Score, runs `tofu plan` + Infracost,
builds and pushes the image to the in-cluster local registry, renders
OpenChoreo Component, SecretReference, and Workload resources via `score2openchoreo`, and
commits them to
`openchoreo/platform-config/environments/dev/`. OpenChoreo reconciles the
Component, SecretReference, and Workload into a running pod. Promote to staging by committing
the same rendered resources into `environments/staging/`. All three M2 repos
live under the
Gitea `openchoreo` organization. Run `scripts/install-m2.sh` to set up,
`scripts/teardown-m2.sh` to remove, `scripts/smoke-m2.sh` to verify.
Direct `tofu apply` from a Bash tool use is blocked by `rr-tofu-guard`;
use the install script.
