# external-secrets wiring module

Creates ExternalSecret objects for M2-owned secret materialization:
- Flux git auth (`flux-system/gitea-deploy-key`)
- the legacy hello-m2 demo secret mirror (`openchoreo-data-plane/example-secret`)

The Gitea Actions runner's ExternalSecret lives in the `gitea-runner` module.

Assumes a `ClusterSecretStore` named `openbao-kv` already exists. Live
OpenChoreo workload secrets are generated from `SecretReference` resources and
the data plane's `default` ClusterSecretStore instead.
