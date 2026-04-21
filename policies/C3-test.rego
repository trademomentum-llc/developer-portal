package constraints.m2.c3_test

import data.constraints.m2.c3

test_under_threshold_allowed {
  count(c3.violation) == 0 with input as {
    "review": {
      "kind": {"kind": "Component"},
      "object": {"metadata": {"annotations": {"pipeline.m2/cost-delta-usd": "10"}}}
    },
    "parameters": {"thresholdUSD": "50"}
  }
}

test_over_threshold_blocked {
  count(c3.violation) == 1 with input as {
    "review": {
      "kind": {"kind": "Component"},
      "object": {"metadata": {"annotations": {"pipeline.m2/cost-delta-usd": "75"}}}
    },
    "parameters": {"thresholdUSD": "50"}
  }
}
