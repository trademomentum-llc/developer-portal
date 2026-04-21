# external-secrets wiring module

Creates ExternalSecret objects for the three M2 consumers:
- Flux git auth (`flux-system/gitea-deploy-key`)
- hello-m2 workload secret (`openchoreo-data-plane/example-secret`)

The Gitea Actions runner's ExternalSecret lives in the `gitea-runner` module.

Assumes a `ClusterSecretStore` named `openbao-kv` already exists (provisioned
by the external-secrets install in M1).
