output "namespace" {
  description = "Namespace where PostgreSQL is installed"
  value       = kubernetes_namespace.backstage.metadata[0].name
}

output "secret_name" {
  description = "Name of the Secret holding the PostgreSQL password"
  value       = kubernetes_secret.postgres.metadata[0].name
}

output "node_port" {
  description = "Host-mapped NodePort for PostgreSQL"
  value       = var.node_port
}
