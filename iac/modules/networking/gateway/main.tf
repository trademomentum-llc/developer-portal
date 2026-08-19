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
        # FR-09: port 80 keeps serving plain HTTP (no redirect in Wave 0);
        # the HTTPS listeners below are offered alongside it. HTTPRoutes use
        # no sectionName (httproutes.tf), so they bind to both by hostname.
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
        - name: https-gitea
          hostname: gitea.local
          protocol: HTTPS
          port: 443
          tls:
            mode: Terminate
            certificateRefs:
              - kind: Secret
                name: gitea-tls
          allowedRoutes:
            namespaces:
              from: All
        - name: https-signoz
          hostname: signoz.local
          protocol: HTTPS
          port: 443
          tls:
            mode: Terminate
            certificateRefs:
              - kind: Secret
                name: signoz-tls
          allowedRoutes:
            namespaces:
              from: All
        - name: https-opencost
          hostname: opencost.local
          protocol: HTTPS
          port: 443
          tls:
            mode: Terminate
            certificateRefs:
              - kind: Secret
                name: opencost-tls
          allowedRoutes:
            namespaces:
              from: All
  EOF
}
