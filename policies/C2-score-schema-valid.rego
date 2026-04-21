# C-2: every Component admitted to the cluster must carry an annotation
# `pipeline.m2/score-valid: "true"`. This annotation is written by the
# Gitea Actions pipeline only if score2openchoreo --validate-only exited 0.
package constraints.m2.c2

violation[{"msg": msg}] {
  input.review.kind.kind == "Component"
  input.review.object.metadata.annotations["pipeline.m2/score-valid"] != "true"
  msg := "Component missing pipeline.m2/score-valid=true annotation"
}
