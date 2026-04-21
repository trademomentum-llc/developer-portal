variable "kubeconfig_path" {
  type    = string
  default = "~/.kube/config"
}

variable "kube_context" {
  type    = string
  default = "k3d-openchoreo"
}

variable "gitea_url" {
  type        = string
  description = "In-cluster URL for Gitea API"
  default     = "http://gitea-http.gitea.svc.cluster.local:3000"
}

variable "openchoreo_project" {
  type    = string
  default = "openchoreo"
}

variable "infracost_threshold_monthly_usd" {
  type    = number
  default = 50
}
