output "flux_namespace"       { value = module.flux.namespace }
output "gatekeeper_namespace" { value = module.gatekeeper.namespace }
output "runner_namespace"     { value = module.gitea_runner.namespace }
output "environments"         { value = module.openchoreo_environments.names }
