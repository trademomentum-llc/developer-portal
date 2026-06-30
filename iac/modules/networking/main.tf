module "envoy_gateway" {
  source = "./envoy-gateway"

  chart_version = var.envoy_gateway_chart_version
  replicas      = var.envoy_gateway_replicas
}

module "gateway" {
  source     = "./gateway"
  depends_on = [module.envoy_gateway]

  routes = var.routes
}

module "cilium" {
  count  = var.enable_cilium ? 1 : 0
  source = "./cilium"

  chart_version = var.cilium_chart_version
  cluster_name  = var.cilium_cluster_name
  cluster_id    = var.cilium_cluster_id
}
