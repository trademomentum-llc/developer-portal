variable "prometheus_chart_version" {
  description = "Prometheus Helm chart version"
  type        = string
  default     = "29.13.0"
}

variable "opencost_chart_version" {
  description = "OpenCost Helm chart version"
  type        = string
  default     = "2.5.25"
}

variable "prometheus_values_file" {
  description = "Path to Prometheus Helm values file"
  type        = string
  default     = "observability/cost/prometheus-values.local.yaml"
}

variable "opencost_values_file" {
  description = "Path to OpenCost Helm values file"
  type        = string
  default     = "observability/cost/opencost-values.local.yaml"
}

variable "kube_context" {
  description = "kubectl context for the k3d-openchoreo cluster"
  type        = string
  default     = "k3d-openchoreo"
}
