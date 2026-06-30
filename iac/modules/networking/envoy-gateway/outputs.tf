output "namespace" {
  description = "Namespace where Envoy Gateway is installed"
  value       = helm_release.envoy_gateway.namespace
}
