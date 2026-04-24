locals {
  # OpenChoreo Environment CRD (openchoreo.dev/v1alpha1). Two envs for M2
  # per the locked-in design: dev, staging. Neither is production.
  environments = {
    dev = {
      display_name  = "Development"
      description   = "M2 development environment"
      is_production = false
    }
    staging = {
      display_name  = "Staging"
      description   = "M2 staging environment"
      is_production = false
    }
  }
}

resource "kubectl_manifest" "environment" {
  for_each = local.environments
  yaml_body = yamlencode({
    apiVersion = "openchoreo.dev/v1alpha1"
    kind       = "Environment"
    metadata = {
      name      = each.key
      namespace = "default"
      labels = {
        "openchoreo.dev/name" = each.key
      }
      annotations = {
        "openchoreo.dev/display-name" = each.value.display_name
        "openchoreo.dev/description"  = each.value.description
      }
    }
    spec = {
      dataPlaneRef = {
        kind = "ClusterDataPlane"
        name = "default"
      }
      isProduction = each.value.is_production
    }
  })
}
