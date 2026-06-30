# M3 Observability -- Design Specification

> **Milestone:** M3 -- Observability
> **Version:** 0.1
> **Date:** 2026-05-23
>
> **Status:** Superseded by the Production Multi-Angle Visibility triad in
> `docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-*`. This kickoff
> draft remains for historical context; use the 2026-05-28 documents for current
> implementation contracts.
> **Status:** Kickoff draft
> **Companion docs:** [requirements.md](./requirements.md), [technical-specification.md](./technical-specification.md)

---

## 1. Purpose

This document describes the shape of M3. It explains the component boundaries and the trade-offs for adding OpenTelemetry, SigNoz, and post-deploy Infracost visibility to the M1/M2 local IDP.

## 2. Context diagram

```
developer / agent
    |
    | git push
    v
Gitea -> Gitea Actions -> score2openchoreo -> platform-config
                                         |             |
                                         |             v
                                         |       OpenChoreo -> hello-m2 pod
                                         |                          |
                                         |                          | OTLP traces/metrics/logs
                                         v                          v
                                  Infracost JSON              OpenTelemetry collector
                                         |                          |
                                         v                          v
                                  cost artifact               SigNoz backend/UI
                                         |                          |
                                         +------------+-------------+
                                                      |
                                                      v
                                                Backstage links
```

M3 is observability around the existing delivery path. It must not replace the delivery path.

## 3. Component boundaries

| Component | Responsibility | Runs where | New in M3 |
|---|---|---|---|
| SigNoz | Local UI/backend for traces, metrics, logs | k3d cluster | yes |
| OpenTelemetry collector | Receives, enriches, and exports telemetry | k3d cluster | yes |
| hello-m2 instrumentation | Emits OTLP telemetry from the demo workload | workload pod | yes |
| Backstage observability links | Exposes SigNoz and cost surfaces from catalog entities | host or container | yes |
| Infracost artifact path | Stores latest accepted estimate for deployed components | Gitea / repo path | yes |
| M3 smokes | Proves health and signal flow | host scripts | yes |

Existing components remain in their current roles. OpenChoreo still deploys workloads. Flux still reconciles add-ons. Gitea Actions still runs CI. Gatekeeper remains pipeline-scoped until M6.

## 4. Design decisions

### 4.1 Start with preflight, not install

SigNoz self-hosted Kubernetes docs list minimum local resources and storage needs. The current local environment is also carrying OpenChoreo, Gitea, openbao, external-secrets, Flux, Gatekeeper, local-registry, Backstage dev, and demo workloads. M3 therefore starts with a preflight script that reports:

- Kubernetes version and node readiness.
- Colima CPU, memory, and disk state when available.
- Storage classes and default storage class.
- Existing `openchoreo-observability-plane` resources.
- Current pod resource requests in M1/M2 namespaces.

**Trade-off accepted:** Slower kickoff. This avoids deploying ClickHouse-backed SigNoz into an undersized local cluster and then debugging resource starvation as if it were an observability bug.

### 4.2 Use SigNoz as the primary UI, but preserve OpenChoreo's plane

OpenChoreo already includes an observability plane namespace. M3 does not assume it owns that namespace. SigNoz should be installed in its own namespace and integrated through OTLP endpoints and Backstage links. If the OpenChoreo observability plane already exposes compatible collectors or dashboards, M3 should document and reuse what is safe, but not rely on undocumented internals.

**Trade-off accepted:** One more namespace. The boundary keeps upstream OpenChoreo components replaceable.

### 4.3 Use OpenTelemetry Collector as the gateway

Applications should not export directly to SigNoz-specific endpoints. The collector gives M3 a stable ingestion contract, Kubernetes metadata enrichment, filtering, batching, and future exporter flexibility.

For the single-node k3d cluster:

- A deployment-style collector is the default for OTLP app traces.
- A daemonset collector is optional for host/pod log and node metrics collection.
- Cluster metrics should be single-replica to avoid duplicates.

This matches the OpenTelemetry chart guidance that deployment, daemonset, and statefulset modes serve different collection needs, and that Kubernetes presets can duplicate data if used carelessly.

### 4.4 Prefer catalog links before a Backstage plugin

Backstage should first surface SigNoz and cost views through `catalog-info.yaml` links and annotations. A custom plugin is justified only after the URL and data model prove stable. This keeps M3 aligned with M2's conservative Backstage approach.

**Trade-off accepted:** Less embedded UI in Backstage at first. The catalog still becomes the front door by linking entities to operational evidence.

### 4.5 Treat Infracost as post-deploy estimate evidence, not live spend

M2 already runs Infracost before deploy. M3 should persist the last accepted estimate and connect it to the deployed component. On local k3d, there is no meaningful cloud bill to reconcile, so the dashboard must label values as estimates.

OpenCost remains the runtime cost allocation tool for M4. M3 should not pretend Infracost can do pod-level runtime allocation.

### 4.6 Containerize Backstage only if telemetry requires it

The current host `yarn dev` path works for development and recently passed the catalog smoke. M3 may need Backstage in-cluster or containerized to emit consistent OTLP metadata and link service identity with workload telemetry. That should be a measured decision:

1. Try host Backstage backend OTLP export with local collector port-forward.
2. If metadata or endpoint stability is weak, add a Backstage image and OpenChoreo/Helm deployment path.
3. Keep host dev as a supported mode for quick iteration.

## 5. Data flow

### 5.1 Runtime telemetry

```
hello-m2 pod
    -> OTLP HTTP or gRPC
    -> collector service
    -> resource processors add k8s metadata
    -> batch processor
    -> SigNoz collector/backend
    -> SigNoz UI
```

### 5.2 Backstage telemetry

```
Backstage backend
    -> OTLP exporter when M3 env vars are set
    -> collector port-forward or in-cluster collector service
    -> SigNoz
```

### 5.3 Cost evidence

```
Gitea Actions Infracost step
    -> JSON artifact
    -> deterministic component/env path
    -> Backstage catalog link or SigNoz/dashboard panel
```

## 6. Failure modes

| Failure | Symptom | Mitigation |
|---|---|---|
| Local cluster undersized | SigNoz or ClickHouse pods pending/crashloop | Preflight blocks install until resource gap is explicit |
| Duplicate metrics | SigNoz shows double-counted node or pod metrics | Single-replica cluster collectors; daemonset only for node-local signals |
| Collector endpoint mismatch | App runs but no traces appear | Smoke sends synthetic OTLP payload and checks collector logs/health |
| Backstage host mode cannot export reliably | No Backstage traces | Containerize Backstage only after host mode fails |
| Infracost artifact missing | Cost dashboard blank | CI marks artifact absence and smoke fails with component/env path |

## 7. Acceptance shape

M3 is accepted when a user can open Backstage, select `hello-m2`, and follow links to:

- Live or recent runtime telemetry in SigNoz.
- At least one trace or metric emitted by `hello-m2`.
- The latest accepted Infracost estimate for the deployed component.

The acceptance demo must run locally and must not require external SaaS.
