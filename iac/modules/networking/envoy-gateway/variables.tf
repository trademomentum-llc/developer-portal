variable "chart_version" {
  description = "Envoy Gateway Helm chart version"
  type        = string
}

variable "replicas" {
  description = "Number of Envoy Gateway controller replicas"
  type        = number
  default     = 1
}
