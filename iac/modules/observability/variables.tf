variable "signoz_chart_version" {
  description = "SigNoz Helm chart version"
  type        = string
  default     = "0.130.1"
}

variable "otel_collector_chart_version" {
  description = "OpenTelemetry Collector Helm chart version"
  type        = string
  default     = "0.155.0"
}

variable "signoz_values_file" {
  description = "Path to SigNoz Helm values file"
  type        = string
  default     = "observability/signoz/values.local.yaml"
}

variable "otel_collector_values_file" {
  description = "Path to OTEL Collector Helm values file"
  type        = string
  default     = "observability/otel/collector-values.local.yaml"
}

variable "kube_context" {
  description = "kubectl context for the k3d-openchoreo cluster"
  type        = string
  default     = "k3d-openchoreo"
}
