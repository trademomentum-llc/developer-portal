# Developer Portal

The Developer Portal is a self-hosted Internal Developer Platform (IDP) for
this organization. It sits on top of a local Kubernetes cluster and turns
infrastructure into a product that engineers can use without being Kubernetes
experts.

## What you can do here

- **Find services and components** in the catalog.
- **Deploy workloads** from a Score file through Gitea Actions and OpenChoreo.
- **Use paved-path templates** to create new services with the right defaults.
- **Track ownership** so you know who to ask about any piece of the platform.

## How it is built

| Layer | Technology | Purpose |
|-------|------------|---------|
| Cluster | k3d | Local Kubernetes substrate |
| Git hosting | Gitea | Repos, Actions runner, and OCI registry |
| Portal | Backstage | Catalog, scaffolder, and UI |
| IaC | OpenTofu | Declarative cluster and add-on provisioning |
| Workload orchestrator | OpenChoreo | Turns Score files into running pods |
| Policies | Gatekeeper / OPA | Guardrails for code and cost |
| Secrets | OpenBao + external-secrets | Per-application secret delivery |

## Main parts of the catalog

- **Developer Portal** -- the Backstage instance you are using now.
- **Backstage App** -- the frontend and backend code behind the portal.
- **Rational Reserve Policy Guards** -- workflow safety hooks written in Go.
- **OpenChoreo Platform** -- the orchestrator that actually runs the workloads.

## Current status

- M1 (substrate) is complete and healthy.
- M2 (IaC + CD loop) is validated end-to-end locally.
- M3 (observability) is in kickoff/specification.

## Useful links

- Local Gitea: <http://localhost:3333>
- Backstage dev: <http://localhost:3001>
- OpenChoreo API: <http://localhost:9090>
