variable "chart_version" {
  description = "Bitnami PostgreSQL Helm chart version"
  type        = string
  default     = "16.4.5"
}

variable "node_port" {
  description = "Host-mapped NodePort for PostgreSQL"
  type        = number
  default     = 30432
}
