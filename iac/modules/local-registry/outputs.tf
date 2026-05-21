output "registry_host" {
  description = "In-cluster DNS name:port to push/pull images."
  value       = "${kubernetes_service.registry.metadata[0].name}.${kubernetes_namespace.local_registry.metadata[0].name}.svc.cluster.local:5000"
}

output "registry_node_port" {
  description = "NodePort used by k3s containerd mirror configuration."
  value       = 30082
}
