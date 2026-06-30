resource "kubernetes_namespace" "tofu_state" {
  metadata { name = "tofu-state" }
}

module "flux" {
  source    = "./modules/flux"
  gitea_url = var.gitea_url
}

module "gatekeeper" {
  source                          = "./modules/gatekeeper"
  infracost_threshold_monthly_usd = var.infracost_threshold_monthly_usd
}

module "gitea_runner" {
  source     = "./modules/gitea-runner"
  gitea_url  = var.gitea_url
  depends_on = [module.flux]
}

module "openchoreo_environments" {
  source  = "./modules/openchoreo-environments"
  project = var.openchoreo_project
}

module "external_secrets_wiring" {
  source = "./modules/external-secrets-wiring"
}

module "local_registry" {
  source = "./modules/local-registry"
}

module "observability" {
  source = "./modules/observability"
}
