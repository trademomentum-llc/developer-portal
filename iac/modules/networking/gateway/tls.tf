# FR-09: TLS on the .local gateways (SEC-PLANE-WAVE0 spec section 9).
# SelfSigned bootstrap issuer -> local CA keypair -> local-ca ClusterIssuer.
# cert-manager is sibling-managed (openchoreo repo); this module only declares
# issuers and Certificate resources. Route-certificate Secrets land in the
# envoy-gateway namespace so Gateway listener certificateRefs need no
# cross-namespace ReferenceGrant.
resource "kubectl_manifest" "selfsigned_bootstrap_issuer" {
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
      name: selfsigned-bootstrap
    spec:
      selfSigned: {}
  EOF
}

resource "kubectl_manifest" "local_ca_certificate" {
  depends_on = [kubectl_manifest.selfsigned_bootstrap_issuer]
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: local-ca-cert
      namespace: cert-manager
    spec:
      isCA: true
      commonName: sovereign-local-ca
      secretName: local-ca-key-pair
      privateKey:
        algorithm: ECDSA
        size: 256
      issuerRef:
        name: selfsigned-bootstrap
        kind: ClusterIssuer
  EOF
}

resource "kubectl_manifest" "local_ca_issuer" {
  depends_on = [kubectl_manifest.local_ca_certificate]
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
      name: local-ca
    spec:
      ca:
        secretName: local-ca-key-pair
  EOF
}

resource "kubectl_manifest" "route_certificates" {
  depends_on = [kubectl_manifest.local_ca_issuer]
  for_each   = var.routes
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: ${each.key}-tls
      namespace: envoy-gateway
    spec:
      secretName: ${each.key}-tls
      dnsNames:
        - ${each.value.hostname}
      issuerRef:
        name: local-ca
        kind: ClusterIssuer
  EOF
}
