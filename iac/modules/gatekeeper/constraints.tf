resource "kubectl_manifest" "c1_template" {
  depends_on = [helm_release.gatekeeper]
  yaml_body  = file("${path.root}/../policies/C1-constraint.yaml")
}

resource "kubectl_manifest" "c2_template" {
  depends_on = [helm_release.gatekeeper]
  yaml_body  = file("${path.root}/../policies/C2-constraint.yaml")
}

resource "kubectl_manifest" "c3_template" {
  depends_on = [helm_release.gatekeeper]
  yaml_body  = file("${path.root}/../policies/C3-constraint.yaml")
}
