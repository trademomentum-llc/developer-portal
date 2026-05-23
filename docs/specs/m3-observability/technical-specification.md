# M3 Observability -- Technical Specification

> **Milestone:** M3 -- Observability
> **Version:** 0.1
> **Date:** 2026-05-23
> **Status:** Kickoff draft
> **Companion docs:** [requirements.md](./requirements.md), [design-specification.md](./design-specification.md)

---

## 1. Purpose

This document is the low-level implementation reference for M3. At kickoff, exact chart versions are intentionally not pinned in code yet because M3 has not performed the required chart and cluster preflight. The first implementation task is version inventory and resource preflight; after that, this document must be updated with exact pins before any install script mutates cluster state.

## 2. Proposed repository layout

```
/Users/nnos/Projects/developer-portal/
+-- docs/
|   +-- specs/
|       +-- m3-observability/
|           +-- requirements.md
|           +-- design-specification.md
|           +-- technical-specification.md
+-- iac/
|   +-- modules/
|       +-- observability/                  # proposed, after preflight
|           +-- README.md
|           +-- main.tf
|           +-- variables.tf
|           +-- outputs.tf
|           +-- versions.tf
+-- observability/
|   +-- signoz/
|   |   +-- values.local.yaml               # proposed, no secrets
|   +-- otel/
|   |   +-- collector-values.local.yaml     # proposed, no secrets
|   +-- dashboards/
|       +-- hello-m2-cost.json              # proposed, generated or hand-authored
+-- scripts/
|   +-- preflight-m3.sh                     # proposed
|   +-- install-m3.sh                       # proposed
|   +-- teardown-m3.sh                      # proposed
|   +-- smoke-m3.sh                         # proposed
|   +-- smoke-signoz.sh                     # proposed
|   +-- smoke-otel.sh                       # proposed
|   +-- smoke-m3-cost.sh                    # proposed
```

## 3. Namespaces and service names

Proposed defaults:

| Resource | Name | Notes |
|---|---|---|
| SigNoz namespace | `signoz` | Dedicated namespace; do not reuse `openchoreo-observability-plane` |
| SigNoz UI service | `signoz` | Upstream docs expose HTTP on port 8080 by default |
| SigNoz collector service | `signoz-otel-collector` | Upstream docs use this service for OTLP routing |
| Collector namespace | `otel-system` or `signoz` | Final choice depends on whether SigNoz chart collector is reused |
| Local UI port | `8080` or first free `18080+` | Avoid collisions with existing Gitea/Backstage ports |

The namespace choice is not final until `preflight-m3.sh` inventories existing resources.

## 4. Version inventory task

Before installation, run and record:

```
helm repo add signoz https://charts.signoz.io
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update
helm search repo signoz/signoz --versions | head -n 5
helm search repo signoz/k8s-infra --versions | head -n 5
helm search repo open-telemetry/opentelemetry-collector --versions | head -n 5
```

After selection, update this section with exact pins:

| Chart | Repository | Version | App/image version | Status |
|---|---|---|---|---|
| `signoz/signoz` | `https://charts.signoz.io` | to be selected by M3 version inventory | to be recorded | not pinned yet |
| `signoz/k8s-infra` | `https://charts.signoz.io` | to be selected by M3 version inventory | to be recorded | not pinned yet |
| `open-telemetry/opentelemetry-collector` | `https://open-telemetry.github.io/opentelemetry-helm-charts` | to be selected by M3 version inventory | to be recorded | not pinned yet |

No `install-m3.sh` implementation should proceed while this table is unpinned.

## 5. Preflight script contract

`scripts/preflight-m3.sh` should be read-only and should print:

```
date
kubectl --context k3d-openchoreo version
kubectl --context k3d-openchoreo get nodes -o wide
kubectl --context k3d-openchoreo get storageclass
kubectl --context k3d-openchoreo get ns
kubectl --context k3d-openchoreo -n openchoreo-observability-plane get all
kubectl --context k3d-openchoreo get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded
kubectl --context k3d-openchoreo top nodes
kubectl --context k3d-openchoreo top pods -A
```

If `kubectl top` is unavailable, the script should report that metrics-server is unavailable and continue. It should also report Colima CPU/memory/disk settings when `colima status` is available.

## 6. SigNoz install shape

Upstream SigNoz Kubernetes docs currently describe Helm installation from `https://charts.signoz.io`, a dedicated namespace, a values file, and a health check at:

```
curl -X GET http://localhost:8080/api/v1/health
```

They also document that the Helm chart installs SigNoz, SigNoz Collector, ClickHouse, and Zookeeper. That means M3 must handle storage class and local disk pressure before install.

`observability/signoz/values.local.yaml` should minimally define:

```yaml
global:
  storageClass: <recorded-storage-class>
clickhouse:
  installCustomStorageClass: true
```

The concrete storage class value must be filled by the M3 preflight result. This is deliberately not committed as a fake value.

## 7. OpenTelemetry collection shape

Default application ingest path:

```
OTEL_EXPORTER_OTLP_ENDPOINT=http://<collector-service>:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_SERVICE_NAME=hello-m2
OTEL_RESOURCE_ATTRIBUTES=service.namespace=openchoreo,service.version=<git-sha>,deployment.environment=dev
```

If the SigNoz chart-provided collector is sufficient, prefer it over adding a second collector. If it is not sufficient for Kubernetes metadata enrichment, add a separate OpenTelemetry Collector chart release with:

- `mode: deployment` for OTLP app gateway and cluster metrics.
- `presets.kubernetesAttributes.enabled: true`.
- `presets.clusterMetrics.enabled: true` with one replica.
- Daemonset/node-log collection only after resource preflight passes.

## 8. hello-m2 instrumentation shape

`seed-repos/hello-m2/main.go` is the first workload to instrument. The minimum acceptable implementation is:

- Add OpenTelemetry Go SDK dependencies.
- Create a tracer provider on startup when `OTEL_EXPORTER_OTLP_ENDPOINT` is set.
- Add a span around the HTTP request handler.
- Add service name, version, and environment resource attributes.
- Keep the app runnable without telemetry when env vars are absent.

The Score file should pass through the required OTEL environment variables into the rendered OpenChoreo Workload.

## 9. Backstage instrumentation shape

Backstage backend instrumentation should be gated by environment variables:

```
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:<collector-port>
OTEL_SERVICE_NAME=developer-portal-backstage
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=local
```

Host dev mode can use a port-forward to the collector. If host mode is not stable, add a containerized Backstage path in M3 and document the new runtime in `PROJECT_SUMMARY.md`, `TODO.md`, and `SESSION_HANDOFF.md`.

## 10. Infracost artifact shape

The first post-deploy cost artifact path should be deterministic:

```
cost-artifacts/
  hello-m2/
    dev/
      latest.json
      latest.md
```

Required fields in `latest.json`:

```json
{
  "component": "hello-m2",
  "environment": "dev",
  "source": "infracost",
  "estimate_type": "post-deploy-reference",
  "currency": "USD",
  "monthly_cost": 0,
  "monthly_delta": 0,
  "git_sha": "",
  "generated_at": ""
}
```

This is not live spend. It is the last accepted estimate tied to the deployed revision. OpenCost owns live/runtime allocation in M4.

## 11. Smoke checks

### 11.1 `scripts/smoke-signoz.sh`

Required checks:

- Port-forward the SigNoz UI service.
- `curl http://127.0.0.1:<port>/api/v1/health`.
- Exit non-zero if status is not healthy.

### 11.2 `scripts/smoke-otel.sh`

Required checks:

- Verify collector pods are Ready.
- Send a synthetic OTLP trace or run a tiny instrumented command against the collector.
- Verify collector logs show accepted data or SigNoz query returns the synthetic service.

### 11.3 `scripts/smoke-m3-cost.sh`

Required checks:

- Validate latest Infracost JSON exists for `hello-m2/dev`.
- Validate required fields are present.
- Validate Backstage catalog contains a link or annotation to the cost artifact or dashboard.

### 11.4 `scripts/smoke-m3.sh`

Runs all M3 smokes in this order:

1. `scripts/preflight-m3.sh`
2. `scripts/smoke-signoz.sh`
3. `scripts/smoke-otel.sh`
4. `scripts/smoke-m3-cost.sh`

## 12. Current blockers

- Cloud Gitea push needs refreshed authentication before remote publication.
- Backstage dependency audit has existing critical/high advisories and needs a dedicated dependency-alignment pass.
- M3 version pins are not selected yet.
- M3 cluster resource headroom is unknown until `preflight-m3.sh` exists and runs.
