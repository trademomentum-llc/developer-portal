resource "kubernetes_namespace" "gatekeeper_system" {
  metadata { name = "gatekeeper-system" }
}

resource "helm_release" "gatekeeper" {
  name       = "gatekeeper"
  namespace  = kubernetes_namespace.gatekeeper_system.metadata[0].name
  repository = "https://open-policy-agent.github.io/gatekeeper/charts"
  chart      = "gatekeeper"
  version    = "3.17.1"
  wait       = true
  timeout    = 600
  set {
    name  = "controllerManager.resources.requests.cpu"
    value = "100m"
  }
}
