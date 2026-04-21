package constraints.m2.c2_test

import data.constraints.m2.c2

test_valid_annotation_allowed {
  count(c2.violation) == 0 with input as {
    "review": {
      "kind": {"kind": "Component"},
      "object": {"metadata": {"annotations": {"pipeline.m2/score-valid": "true"}}}
    }
  }
}

test_missing_annotation_blocked {
  count(c2.violation) == 1 with input as {
    "review": {
      "kind": {"kind": "Component"},
      "object": {"metadata": {"annotations": {}}}
    }
  }
}
