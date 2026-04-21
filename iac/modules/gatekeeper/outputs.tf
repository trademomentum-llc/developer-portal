output "namespace" { value = kubernetes_namespace.gatekeeper_system.metadata[0].name }
