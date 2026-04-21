# C-2: every Component admitted to the cluster must carry an annotation
# pipeline.m2/score-valid: "true". The annotation is written by the
# Gitea Actions pipeline only if score2openchoreo --validate-only exited 0.
package constraints.m2.c2

violation[{"msg": msg}] {
  input.review.kind.kind == "Component"
  anns := object.get(input.review.object.metadata, "annotations", {})
  object.get(anns, "pipeline.m2/score-valid", "") != "true"
  msg := "Component missing pipeline.m2/score-valid=true annotation"
}
