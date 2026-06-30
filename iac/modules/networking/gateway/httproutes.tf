resource "kubectl_manifest" "httproutes" {
  depends_on = [kubectl_manifest.gateway]

  for_each = var.routes

  yaml_body = <<-EOF
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    metadata:
      name: ${each.key}
      namespace: ${each.value.service.namespace}
    spec:
      parentRefs:
        - name: eg
          namespace: envoy-gateway
      hostnames:
        - ${each.value.hostname}
      rules:
        - backendRefs:
            - name: ${each.value.service.name}
              port: ${each.value.service.port}
  EOF
}
