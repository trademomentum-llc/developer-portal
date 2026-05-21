# developer-portal -- QWEN.md

This file provides guidance to Qwen Code when working with code in this
repository.

## Current project state

This repository is the umbrella for a self-hosted Internal Developer Platform
(IDP) split into seven milestones, M1 through M7. M1 substrate is complete and
healthy. M2, the IaC and CD loop, is architecturally validated through Pod
creation on the live `k3d-openchoreo` cluster.

The remaining M2 closeout work is narrow: trigger a fresh `hello-m2` CI run so
the rewritten `score2openchoreo` output is committed by automation and the
new image tag is pushed into the in-cluster local registry.

Always read these files first in a new session:

1. `SESSION_HANDOFF.md`
2. `PROJECT_SUMMARY.md`
3. `TODO.md`

They are more authoritative than git history alone.

## Related checkouts

- `~/Projects/openchoreo/` -- upstream platform orchestrator. Provides the
  shared `k3d-openchoreo` cluster and the canonical OpenChoreo sample manifests.
- `~/Projects/rational-reserve/` -- AI swarm orchestration layer. Integration
  into this portal is deferred to M7.

## Core constraints

- Keep all file writes pure ASCII. No emoji, smart quotes, em dashes,
  box-drawing characters, arrows, or other non-ASCII bytes.
- Do not run `tofu apply`, `tofu destroy`, or `tofu import` directly. Use
  `scripts/install-m2.sh` for M2 infrastructure mutation. `tofu init` and
  `tofu plan` are allowed.
- Keep `SESSION_HANDOFF.md`, `PROJECT_SUMMARY.md`, and `TODO.md` current when
  scope, blockers, milestone state, or closeout status changes.
- New modules, tools, scripts, plugins, or systems require Requirements,
  Design Specification, and Technical Specification docs per the parent
  project rule.
- Prefer deterministic logic and compiled tools. Use Go for local utilities
  unless an existing ecosystem forces another language.

## Common commands

Policy guards:

```bash
cd plugins/rr-policy-guards/tools/<emoji|bash|brew|tofu>-guard
go test ./...
go build -o ../../bin/rr-<name>-guard .
```

score2openchoreo:

```bash
cd tools/score2openchoreo
go test ./...
go build -o bin/score2openchoreo .
```

Gatekeeper policies:

```bash
opa test --v0-compatible policies/*.rego -v
```

Backstage:

```bash
cd backstage
yarn install
yarn dev
yarn tsc
yarn test
yarn lint
yarn test:e2e
```

Lifecycle scripts:

```bash
./scripts/install-m1.sh
./scripts/install-m1.sh --fresh
./scripts/teardown-m1.sh

./scripts/install-m2.sh
./scripts/smoke-m2.sh
./scripts/teardown-m2.sh
```

## Architecture notes

The locked-in Integration and Delivery stack is:

**Gitea + Backstage Software Catalog and Score + OpenTofu + Gitea Actions +
an OCI registry**

The implemented M2 image path uses the dedicated in-cluster `local-registry`
module so the runner and k3s containerd can push and pull over stable
cluster-local HTTP endpoints.

M2 also uses Flux for cluster add-on drift correction and OPA/Gatekeeper for
pipeline-scoped constraints. OpenChoreo remains the workload deployer.

The M2 developer path is:

1. Push to `openchoreo/hello-m2`.
2. Gitea Actions runner executes `.gitea/workflows/ci.yaml`.
3. CI validates Score, runs plan/cost checks, builds and pushes the image.
4. `score2openchoreo` renders OpenChoreo resources.
5. CI commits rendered YAML to `openchoreo/platform-config`.
6. Flux applies platform-config.
7. OpenChoreo reconciles to ComponentRelease, Deployment, and Pod.

## Local Gitea remote

The local `origin` remote should point at the 3333 port-forward:

```bash
git remote set-url origin http://localhost:3333/openchoreo/developer-portal.git
```
