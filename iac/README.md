# M2 OpenTofu root module

Applies Flux, Gatekeeper, the Gitea Actions runner, OpenChoreo Environment
CRDs, and external-secrets wiring on top of the M1 substrate.

## Inputs

| Variable | Default | Purpose |
|---|---|---|
| kubeconfig_path | ~/.kube/config | Path to kubeconfig |
| kube_context | k3d-openchoreo | kube context name |
| gitea_url | http://gitea-http.gitea.svc.cluster.local:3000 | In-cluster Gitea API URL |
| openchoreo_project | openchoreo | OpenChoreo project name for Environments |
| infracost_threshold_monthly_usd | 50 | Merge gate threshold for C-3 |

## Outputs

| Output | Description |
|---|---|
| flux_namespace | Namespace where Flux runs |
| gatekeeper_namespace | Namespace where Gatekeeper runs |
| runner_namespace | Namespace where the Gitea Actions runner runs |
| environments | List of OpenChoreo Environments created |

Apply with: `tofu init && tofu apply -auto-approve`
