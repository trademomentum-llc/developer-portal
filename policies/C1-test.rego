package constraints.m2.c1_test

import data.constraints.m2.c1

test_main_branch_allowed {
  count(c1.violation) == 0 with input as {
    "review": {
      "kind": {"kind": "GitRepository"},
      "object": {
        "spec": {
          "url": "http://gitea-http.gitea.svc.cluster.local:3000/openchoreo/platform-addons",
          "ref": {"branch": "main"}
        }
      }
    }
  }
}

test_non_main_branch_blocked {
  count(c1.violation) == 1 with input as {
    "review": {
      "kind": {"kind": "GitRepository"},
      "object": {
        "spec": {
          "url": "http://gitea-http.gitea.svc.cluster.local:3000/openchoreo/platform-addons",
          "ref": {"branch": "develop"}
        }
      }
    }
  }
}
