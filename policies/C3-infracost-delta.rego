# C-3: when admitting a Component, the pipeline.m2/cost-delta-usd annotation
# (numeric string) must be <= the threshold parameter.
package constraints.m2.c3

violation[{"msg": msg}] {
  input.review.kind.kind == "Component"
  delta_str := input.review.object.metadata.annotations["pipeline.m2/cost-delta-usd"]
  delta := to_number(delta_str)
  threshold := to_number(input.parameters.thresholdUSD)
  delta > threshold
  msg := sprintf("estimated monthly cost delta $%v exceeds threshold $%v", [delta, threshold])
}
