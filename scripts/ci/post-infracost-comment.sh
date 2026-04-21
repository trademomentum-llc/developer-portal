#!/usr/bin/env bash
# Post Infracost delta as a PR comment on the current PR.
set -euo pipefail

DELTA=$1
COMMENT="Infracost monthly delta: \$${DELTA}"

curl -fsS -u "gitea_admin:${GITEA_TOKEN:-$(cat ~/.rational-reserve/m1-gitea-admin-password)}" \
    -X POST "http://gitea-http.gitea.svc.cluster.local:3000/api/v1/repos/${GITHUB_REPOSITORY}/issues/${GITHUB_PR_NUMBER}/comments" \
    -H 'Content-Type: application/json' \
    -d "{\"body\":\"${COMMENT}\"}"
