output "envoy_gateway_namespace" {
  description = "Namespace where Envoy Gateway is installed"
  value       = module.envoy_gateway.namespace
}

output "gateway_name" {
  description = "Name of the Envoy Gateway Gateway resource"
  value       = "eg"
}
