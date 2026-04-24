resource "kubernetes_namespace" "gitea_runners" {
  metadata { name = "gitea-runners" }
}

resource "kubectl_manifest" "runner_token" {
  depends_on = [kubernetes_namespace.gitea_runners]
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1"
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
  # Upstream "act-runner" chart was renamed to "actions" (same repo). The
  # v0.1.0 chart exposes enabled=false by default; we flip it on and
  # wire in the OpenBao-backed runner token via existingSecret.
  name       = "act-runner"
  namespace  = kubernetes_namespace.gitea_runners.metadata[0].name
  repository = "https://dl.gitea.com/charts/"
  chart      = "actions"
  version    = "0.1.0"
  wait       = true
  timeout    = 600

  set {
    name  = "enabled"
    value = "true"
  }
  set {
    name  = "giteaRootURL"
    value = var.gitea_url
  }
  set {
    name  = "existingSecret"
    value = "gitea-runner-token"
  }
  set {
    name  = "existingSecretKey"
    value = "token"
  }

  # The in-cluster local-registry (distribution/distribution) serves plain
  # HTTP at registry.local-registry.svc.cluster.local:5000. Tell dind's
  # dockerd to treat it as an insecure registry so `docker push` works.
  set {
    name  = "statefulset.dind.extraArgs[0]"
    value = "--insecure-registry=registry.local-registry.svc.cluster.local:5000"
  }

  depends_on = [kubectl_manifest.runner_token]
}
