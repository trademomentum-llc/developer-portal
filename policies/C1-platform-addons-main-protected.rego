# C-1: the Flux GitRepository referencing platform-addons must reference the
# protected main branch, with Gitea-side branch protection enforcing PR merges.
package constraints.m2.c1

violation[{"msg": msg}] {
  input.review.kind.kind == "GitRepository"
  endswith(input.review.object.spec.url, "/openchoreo/platform-addons")
  input.review.object.spec.ref.branch != "main"
  msg := "platform-addons GitRepository must reference the main branch"
}
