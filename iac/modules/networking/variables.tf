variable "envoy_gateway_chart_version" {
  description = "Envoy Gateway Helm chart version"
  type        = string
  default     = "1.3.1"
}

variable "envoy_gateway_replicas" {
  description = "Number of Envoy Gateway controller replicas"
  type        = number
  default     = 1
}

variable "cilium_chart_version" {
  description = "Cilium Helm chart version"
  type        = string
  default     = "1.16.5"
}

variable "enable_cilium" {
  description = "Whether to install Cilium (intended for fresh clusters without Flannel)"
  type        = bool
  default     = false
}

variable "cilium_cluster_name" {
  description = "Cilium cluster name"
  type        = string
  default     = "k3d-openchoreo"
}

variable "cilium_cluster_id" {
  description = "Cilium cluster ID"
  type        = number
  default     = 1
}

variable "routes" {
  description = "Map of hostname routes to backend services"
  type = map(object({
    hostname = string
    service = object({
      name      = string
      namespace = string
      port      = number
    })
  }))
  default = {
    gitea = {
      hostname = "gitea.local"
      service = {
        name      = "gitea-http"
        namespace = "gitea"
        port      = 3000
      }
    }
    signoz = {
      hostname = "signoz.local"
      service = {
        name      = "signoz"
        namespace = "signoz"
        port      = 8080
      }
    }
    opencost = {
      hostname = "opencost.local"
      service = {
        name      = "opencost"
        namespace = "opencost"
        port      = 9090
      }
    }
  }
}
