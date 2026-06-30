resource "helm_release" "cilium" {
  name       = "cilium"
  namespace  = "kube-system"
  repository = "https://helm.cilium.io/"
  chart      = "cilium"
  version    = var.chart_version

  values = [
    <<-EOF
    cluster:
      name: ${var.cluster_name}
      id: ${var.cluster_id}
    hubble:
      enabled: true
      relay:
        enabled: true
      ui:
        enabled: true
    operator:
      replicas: 1
    EOF
  ]
}
