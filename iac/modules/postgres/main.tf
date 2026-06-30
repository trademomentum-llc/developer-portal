resource "random_password" "postgres" {
  length  = 24
  special = false
}

resource "kubernetes_namespace" "backstage" {
  metadata { name = "backstage" }
}

resource "kubernetes_secret" "postgres" {
  metadata {
    name      = "postgres-backstage"
    namespace = kubernetes_namespace.backstage.metadata[0].name
  }
  data = {
    password = random_password.postgres.result
  }
}

resource "helm_release" "postgres" {
  name       = "postgres"
  namespace  = kubernetes_namespace.backstage.metadata[0].name
  repository = "oci://registry-1.docker.io/bitnamicharts"
  chart      = "postgresql"
  version    = var.chart_version

  set {
    name  = "auth.database"
    value = "backstage"
  }
  set {
    name  = "auth.username"
    value = "backstage"
  }
  set_sensitive {
    name  = "auth.password"
    value = random_password.postgres.result
  }
  set {
    name  = "primary.service.type"
    value = "NodePort"
  }
  set {
    name  = "primary.service.nodePorts.postgresql"
    value = var.node_port
  }
  set {
    name  = "primary.resources.requests.memory"
    value = "256Mi"
  }
  set {
    name  = "primary.resources.requests.cpu"
    value = "100m"
  }
  set {
    name  = "primary.persistence.size"
    value = "1Gi"
  }
  set {
    name  = "image.registry"
    value = "docker.io"
  }
  set {
    name  = "image.repository"
    value = "bitnamilegacy/postgresql"
  }
  set {
    name  = "image.tag"
    value = "17.6.0-debian-12-r4"
  }
  set {
    name  = "image.pullPolicy"
    value = "IfNotPresent"
  }
  set {
    name  = "global.security.allowInsecureImages"
    value = "true"
  }
}
