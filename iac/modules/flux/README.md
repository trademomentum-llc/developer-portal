# Flux module

Installs Flux v2 in `flux-system`, creates a GitRepository pointing at the
`openchoreo/platform-addons` Gitea repo (auth via `gitea-deploy-key` Secret
provisioned by `external-secrets-wiring`), and a Kustomization that watches
`./clusters/default/`.

| Input | Default | Purpose |
|---|---|---|
| gitea_url | (required) | Base URL for Gitea API |

| Output | |
|---|---|
| namespace | flux-system |
