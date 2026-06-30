terraform {
  required_version = ">= 1.9.0, < 1.12.0"
  required_providers {
    kubectl = { source = "alekc/kubectl", version = "~> 2.1" }
  }
}
