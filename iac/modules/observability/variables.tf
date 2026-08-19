variable "signoz_chart_version" {
  description = "SigNoz Helm chart version"
  type        = string
  default     = "0.130.1"
}

variable "otel_collector_chart_version" {
  description = "OpenTelemetry Collector Helm chart version"
  type        = string
  # Aligned to the live deployed release on 2026-08-18 (state-heal): the
  # otel-collector release was installed raw-helm at chart 0.159.2 and was
  # never tracked in tofu state. The module must describe reality. All values
  # keys we override (mode, image, command, clusterRole, resources,
  # extraVolumes, extraVolumeMounts, config) are unchanged between 0.155.0
  # and 0.159.2 chart defaults; 0.159.2 only adds optional presets.
  default     = "0.159.2"
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
