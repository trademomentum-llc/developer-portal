# Dev-mode auth for OpenBao's kv/ mount. The root token is acceptable here
# because openbao runs in dev mode with inmem storage (m2i-6 tracks the
# persistent-backing follow-up). For prod: swap to kubernetes auth with a
# service account, mirroring openchoreo/install/prerequisites/openbao/setup.sh.
resource "kubernetes_secret" "openbao_root_token" {
  metadata {
    name      = "openbao-root-token"
    namespace = "external-secrets"
  }
  data = {
    token = "root"
  }
  type = "Opaque"
}

resource "kubectl_manifest" "openbao_kv_store" {
  depends_on = [kubernetes_secret.openbao_root_token]
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1"
    kind       = "ClusterSecretStore"
    metadata   = { name = "openbao-kv" }
    spec = {
      provider = {
        vault = {
          server  = "http://openbao.openbao.svc:8200"
          path    = "kv"
          version = "v2"
          auth = {
            tokenSecretRef = {
              name      = "openbao-root-token"
              namespace = "external-secrets"
              key       = "token"
            }
          }
        }
      }
    }
  })
}

resource "kubectl_manifest" "flux_deploy_key" {
  depends_on = [kubectl_manifest.openbao_kv_store]
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1"
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
  depends_on = [kubectl_manifest.openbao_kv_store]
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1"
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
