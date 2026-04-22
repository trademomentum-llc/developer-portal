resource "kubernetes_namespace" "gitea_runners" {
  metadata { name = "gitea-runners" }
}

resource "kubectl_manifest" "runner_token" {
  depends_on = [kubernetes_namespace.gitea_runners]
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1beta1"
    kind       = "ExternalSecret"
    metadata = {
      name      = "gitea-runner-token"
      namespace = "gitea-runners"
    }
    spec = {
      refreshInterval = "1h"
      secretStoreRef  = { name = "openbao-kv", kind = "ClusterSecretStore" }
      target          = { name = "gitea-runner-token", creationPolicy = "Owner" }
      data = [{
        secretKey = "token"
        remoteRef = { key = "gitea/runners/token", property = "token" }
      }]
    }
  })
}

resource "helm_release" "act_runner" {
  name       = "act-runner"
  namespace  = kubernetes_namespace.gitea_runners.metadata[0].name
  repository = "https://dl.gitea.com/charts/"
  chart      = "act-runner"
  version    = "0.2.10"
  wait       = true
  timeout    = 600

  set {
    name  = "giteaRootURL"
    value = var.gitea_url
  }
  set {
    name  = "existingSecret"
    value = "gitea-runner-token"
  }
  set {
    name  = "podSecurityContext.runAsNonRoot"
    value = "true"
  }
  set {
    name  = "securityContext.allowPrivilegeEscalation"
    value = "false"
  }
  set {
    name  = "config.runner.labels[0]"
    value = "ubuntu-latest:docker://ghcr.io/catthehacker/ubuntu:act-latest"
  }

  depends_on = [kubectl_manifest.runner_token]
}
