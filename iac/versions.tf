terraform {
  required_version = ">= 1.9.0, < 1.12.0"
  required_providers {
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.33" }
    helm       = { source = "hashicorp/helm", version = "~> 2.17" }
    kubectl    = { source = "alekc/kubectl", version = "~> 2.1" }
  }
}
