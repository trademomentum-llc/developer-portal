# iac/modules/observability

OpenTofu module for the M3 Production Multi-Angle Visibility observability plane.

## What it installs

- **SigNoz** in namespace `signoz` (chart `signoz/signoz`)
- **Standalone OpenTelemetry Collector** in namespace `otel-system` (chart `opentelemetry-collector`)
- A post-install patch on the `signoz-otel-collector` Deployment to remove the
  enterprise-only OpAMP manager arguments so OTLP ports `4317/4318` open in the
  local/community configuration.

## Inputs

| Variable | Default | Description |
|---|---|---|
| `signoz_chart_version` | `0.130.1` | SigNoz Helm chart version |
| `otel_collector_chart_version` | `0.155.0` | OpenTelemetry Collector Helm chart version |
| `signoz_values_file` | `observability/signoz/values.local.yaml` | Path to SigNoz values (relative to repo root) |
| `otel_collector_values_file` | `observability/otel/collector-values.local.yaml` | Path to collector values (relative to repo root) |
| `kube_context` | `k3d-openchoreo` | kubectl context used by the patch provisioner |

## Usage

From the repo root:

```bash
./scripts/install-m3.sh
```

Or manually from `iac/`:

```bash
tofu init -reconfigure
tofu apply -target=module.observability
```

## Coexistence

This module deliberately does not touch `openchoreo-observability-plane`.
The SigNoz stack lives in its own `signoz` namespace and the collector in
`otel-system`, per the M3 technical specification.
