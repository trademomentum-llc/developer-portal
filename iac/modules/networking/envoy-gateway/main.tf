resource "helm_release" "envoy_gateway" {
  name             = "envoy-gateway"
  namespace        = "envoy-gateway"
  create_namespace = true
  repository       = "oci://docker.io/envoyproxy"
  chart            = "gateway-helm"
  version          = var.chart_version

  values = [
    <<-EOF
    deployment:
      replicas: ${var.replicas}
    EOF
  ]
}
