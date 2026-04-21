locals {
  environments = ["dev", "staging"]
}

resource "kubectl_manifest" "environment" {
  for_each = toset(local.environments)
  yaml_body = yamlencode({
    apiVersion = "core.choreo.dev/v1alpha1"
    kind       = "Environment"
    metadata = {
      name      = each.key
      namespace = "openchoreo-control-plane"
    }
    spec = {
      displayName = each.key
      project     = var.project
    }
  })
}
