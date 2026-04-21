resource "kubectl_manifest" "flux_deploy_key" {
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1beta1"
    kind       = "ExternalSecret"
    metadata = {
      name      = "gitea-deploy-key"
      namespace = "flux-system"
    }
    spec = {
      refreshInterval = "1h"
      secretStoreRef  = { name = "openbao-kv", kind = "ClusterSecretStore" }
      target          = { name = "gitea-deploy-key", creationPolicy = "Owner" }
      data = [
        { secretKey = "username", remoteRef = { key = "flux/gitea-deploy-key", property = "username" } },
        { secretKey = "password", remoteRef = { key = "flux/gitea-deploy-key", property = "password" } },
      ]
    }
  })
}

resource "kubectl_manifest" "hello_m2_dev_secret" {
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1beta1"
    kind       = "ExternalSecret"
    metadata = {
      name      = "hello-m2-example-secret"
      namespace = "openchoreo-data-plane"
    }
    spec = {
      refreshInterval = "1h"
      secretStoreRef  = { name = "openbao-kv", kind = "ClusterSecretStore" }
      target          = { name = "example-secret", creationPolicy = "Owner" }
      data = [{
        secretKey = "password"
        remoteRef = { key = "apps/hello-m2/dev/example-secret", property = "password" }
      }]
    }
  })
}
