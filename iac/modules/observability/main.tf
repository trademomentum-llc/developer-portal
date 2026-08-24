# M3 Production Multi-Angle Visibility — Observability Plane
#
# Installs SigNoz and a standalone OpenTelemetry Collector on the
# k3d-openchoreo cluster. The standalone collector forwards OTLP to the
# SigNoz collector. The SigNoz collector is patched after install to remove
# the enterprise-only OpAMP manager arguments that block OTLP ports in the
# community/local chart configuration.
#
# This module intentionally does NOT install into openchoreo-observability-plane;
# it uses dedicated namespaces (signoz, otel-system) per the M3 technical spec.

locals {
  signoz_values         = fileexists("${path.module}/../../../${var.signoz_values_file}") ? file("${path.module}/../../../${var.signoz_values_file}") : ""
  otel_collector_values = fileexists("${path.module}/../../../${var.otel_collector_values_file}") ? file("${path.module}/../../../${var.otel_collector_values_file}") : ""
}

resource "helm_release" "signoz" {
  name             = "signoz"
  repository       = "https://charts.signoz.io"
  chart            = "signoz"
  version          = var.signoz_chart_version
  namespace        = "signoz"
  create_namespace = true

  values = local.signoz_values != "" ? [local.signoz_values] : []

  # Generous wait: the telemetrystore-migrator pre-upgrade hook polls
  # ClickHouse readiness and runs schema migrations; on a resource-tight
  # laptop cluster this can exceed the default window during node churn.
  timeout = 900
}

resource "helm_release" "otel_collector" {
  name             = "otel-collector"
  repository       = "https://open-telemetry.github.io/opentelemetry-helm-charts"
  chart            = "opentelemetry-collector"
  version          = var.otel_collector_chart_version
  namespace        = "otel-system"
  create_namespace = true

  values = local.otel_collector_values != "" ? [local.otel_collector_values] : []

  timeout = 300
}

# Workaround: SigNoz v0.130.1 collector Deployment ships with an OpAMP-only
# manager argument that prevents OTLP ports from opening in the local/community
# configuration. Replace the container args with the single collector config arg.
resource "null_resource" "patch_signoz_collector" {
  triggers = {
    signoz_revision = helm_release.signoz.metadata[0].revision
  }

  provisioner "local-exec" {
    command = <<-EOT
      kubectl --context "${var.kube_context}" -n signoz patch deployment signoz-otel-collector \
        --type='json' \
        -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/args", "value": ["--config=/conf/otel-collector-config.yaml"]}]'
    EOT
  }

  depends_on = [helm_release.signoz]
}
