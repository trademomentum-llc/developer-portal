resource "kubectl_manifest" "gateway_class" {
  yaml_body = <<-EOF
    apiVersion: gateway.networking.k8s.io/v1
    kind: GatewayClass
    metadata:
      name: eg
    spec:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
  EOF
}

resource "kubectl_manifest" "envoy_proxy" {
  yaml_body = <<-EOF
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyProxy
    metadata:
      name: eg
      namespace: envoy-gateway
    spec:
      provider:
        type: Kubernetes
        kubernetes:
          envoyService:
            type: NodePort
  EOF
}

resource "kubectl_manifest" "gateway" {
  depends_on = [kubectl_manifest.gateway_class, kubectl_manifest.envoy_proxy]

  yaml_body = <<-EOF
    apiVersion: gateway.networking.k8s.io/v1
    kind: Gateway
    metadata:
      name: eg
      namespace: envoy-gateway
    spec:
      gatewayClassName: eg
      infrastructure:
        parametersRef:
          group: gateway.envoyproxy.io
          kind: EnvoyProxy
          name: eg
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
  EOF
}
