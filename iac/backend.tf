terraform {
  backend "kubernetes" {
    secret_suffix    = "state-m2"
    namespace        = "tofu-state"
    load_config_file = true
    config_path      = "~/.kube/config"
    config_context   = "k3d-openchoreo"
  }
}
