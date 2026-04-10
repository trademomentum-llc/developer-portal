# Developer Portal -- M1 Substrate

Self-hosted Internal Developer Platform (IDP) built on k3d, Gitea, and Backstage. OpenChoreo integration is deferred to M3.

## Prerequisites

- Docker (via Colima): `colima start --cpu 4 --memory 8`
- k3d, kubectl, helm, node, yarn (brew-installable)
- Go 1.21+ (for building policy guard hooks)

## Install

```
./scripts/install-m1.sh
```

Resumes from the last completed task. Use `--fresh` to start over.

## Services

| Service | URL | Notes |
|---------|-----|-------|
| Backstage | http://localhost:3000 | Frontend + backend |
| Gitea | http://localhost:3002 | Git hosting, admin: gitea_admin |
| k3d cluster | -- | Name: m1-substrate |

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
