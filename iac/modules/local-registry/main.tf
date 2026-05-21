# In-cluster OCI registry for M2. Runs distribution/distribution (aka
# registry:2) as a single-replica Deployment with an emptyDir backing
# (images vanish on pod restart, acceptable for a dev cluster). Reachable
# from inside the cluster at registry.local-registry.svc.cluster.local:5000
# over plain HTTP; push/pull configured as an insecure registry on the
# runner's dind and on k3s containerd.

resource "kubernetes_namespace" "local_registry" {
  metadata { name = "local-registry" }
}

resource "kubernetes_deployment" "registry" {
  metadata {
    name      = "registry"
    namespace = kubernetes_namespace.local_registry.metadata[0].name
    labels    = { app = "registry" }
  }
  spec {
    replicas = 1
    selector { match_labels = { app = "registry" } }
    template {
      metadata { labels = { app = "registry" } }
      spec {
        container {
          name  = "registry"
          image = "registry:2.8"
          port { container_port = 5000 }
          env {
            name  = "REGISTRY_HTTP_ADDR"
            value = "0.0.0.0:5000"
          }
          env {
            name  = "REGISTRY_STORAGE_DELETE_ENABLED"
            value = "true"
          }
          readiness_probe {
            http_get {
              path = "/v2/"
              port = "5000"
            }
            initial_delay_seconds = 2
            period_seconds        = 5
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "registry" {
  metadata {
    name      = "registry"
    namespace = kubernetes_namespace.local_registry.metadata[0].name
  }
  spec {
    type     = "NodePort"
    selector = { app = "registry" }
    port {
      name        = "http"
      port        = 5000
      target_port = 5000
      node_port   = 30082
    }
  }
}
