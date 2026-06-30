# M4 Cost Visibility Plane
#
# Deploys a minimal Prometheus + OpenCost stack in the k3d-openchoreo cluster
# to provide actual cluster cost allocation for OpenChoreo workloads.
# The stack is intentionally separate from SigNoz/openchoreo-observability-plane
# so it can be installed and torn down independently.

locals {
  prometheus_values = fileexists("${path.module}/../../../${var.prometheus_values_file}") ? file("${path.module}/../../../${var.prometheus_values_file}") : ""
  opencost_values   = fileexists("${path.module}/../../../${var.opencost_values_file}") ? file("${path.module}/../../../${var.opencost_values_file}") : ""
}

resource "helm_release" "prometheus" {
  name             = "prometheus"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "prometheus"
  version          = var.prometheus_chart_version
  namespace        = "opencost"
  create_namespace = true

  values = local.prometheus_values != "" ? [local.prometheus_values] : []

  timeout = 300
}

resource "helm_release" "opencost" {
  name             = "opencost"
  repository       = "https://opencost.github.io/opencost-helm-chart"
  chart            = "opencost"
  version          = var.opencost_chart_version
  namespace        = "opencost"
  create_namespace = true

  values = local.opencost_values != "" ? [local.opencost_values] : []

  timeout = 300

  depends_on = [helm_release.prometheus]
}
