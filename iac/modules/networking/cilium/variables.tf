variable "chart_version" {
  description = "Cilium Helm chart version"
  type        = string
}

variable "cluster_name" {
  description = "Cilium cluster name"
  type        = string
}

variable "cluster_id" {
  description = "Cilium cluster ID"
  type        = number
}
