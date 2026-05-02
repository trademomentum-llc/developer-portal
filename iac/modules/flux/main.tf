resource "kubernetes_namespace" "flux_system" {
  metadata { name = "flux-system" }
}

resource "helm_release" "flux" {
  name       = "flux2"
  namespace  = kubernetes_namespace.flux_system.metadata[0].name
  repository = "https://fluxcd-community.github.io/helm-charts"
  chart      = "flux2"
  version    = "2.13.0"
  wait       = true
  timeout    = 600
}

resource "kubectl_manifest" "platform_addons_source" {
  depends_on = [helm_release.flux]
  yaml_body = yamlencode({
    apiVersion = "source.toolkit.fluxcd.io/v1"
    kind       = "GitRepository"
    metadata = {
      name      = "platform-addons"
      namespace = "flux-system"
    }
    spec = {
      interval  = "1m"
      url       = "${var.gitea_url}/openchoreo/platform-addons"
      ref       = { branch = "main" }
      secretRef = { name = "gitea-deploy-key" }
    }
  })
}

resource "kubectl_manifest" "platform_addons_kustomization" {
  depends_on = [kubectl_manifest.platform_addons_source]
  yaml_body = yamlencode({
    apiVersion = "kustomize.toolkit.fluxcd.io/v1"
    kind       = "Kustomization"
    metadata = {
      name      = "platform-addons"
      namespace = "flux-system"
    }
    spec = {
      interval  = "1m"
      path      = "./clusters/default"
      prune     = true
      sourceRef = { kind = "GitRepository", name = "platform-addons" }
    }
  })
}

resource "kubectl_manifest" "platform_config_source" {
  depends_on = [helm_release.flux]
  yaml_body = yamlencode({
    apiVersion = "source.toolkit.fluxcd.io/v1"
    kind       = "GitRepository"
    metadata = {
      name      = "platform-config"
      namespace = "flux-system"
    }
    spec = {
      interval  = "1m"
      url       = "${var.gitea_url}/openchoreo/platform-config"
      ref       = { branch = "main" }
      secretRef = { name = "gitea-deploy-key" }
    }
  })
}

resource "kubectl_manifest" "platform_config_dev_kustomization" {
  depends_on = [kubectl_manifest.platform_config_source]
  yaml_body = yamlencode({
    apiVersion = "kustomize.toolkit.fluxcd.io/v1"
    kind       = "Kustomization"
    metadata = {
      name      = "platform-config-dev"
      namespace = "flux-system"
    }
    spec = {
      interval  = "1m"
      path      = "./environments/dev"
      prune     = true
      sourceRef = { kind = "GitRepository", name = "platform-config" }
    }
  })
}

resource "kubectl_manifest" "platform_config_staging_kustomization" {
  depends_on = [kubectl_manifest.platform_config_source]
  yaml_body = yamlencode({
    apiVersion = "kustomize.toolkit.fluxcd.io/v1"
    kind       = "Kustomization"
    metadata = {
      name      = "platform-config-staging"
      namespace = "flux-system"
    }
    spec = {
      interval  = "1m"
      path      = "./environments/staging"
      prune     = true
      sourceRef = { kind = "GitRepository", name = "platform-config" }
    }
  })
}
