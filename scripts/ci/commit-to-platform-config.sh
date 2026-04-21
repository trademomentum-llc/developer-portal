#!/usr/bin/env bash
# Commit a rendered Component YAML into platform-config/environments/<env>/<app>.yaml
set -euo pipefail

ENVIRONMENT=$1
APP=$2
COMPONENT_YAML=$3

CONTENT_B64=$(base64 -w 0 < "$COMPONENT_YAML" 2>/dev/null || base64 < "$COMPONENT_YAML")
PATH_IN_REPO="environments/${ENVIRONMENT}/${APP}.yaml"

# Check if file exists; update vs create uses different SHAs
SHA=$(curl -fsS -u "gitea_admin:${GITEA_TOKEN}" \
    "http://gitea-http.gitea.svc.cluster.local:3000/api/v1/repos/openchoreo/platform-config/contents/${PATH_IN_REPO}" \
    2>/dev/null | jq -r '.sha // empty')

PAYLOAD=$(jq -n \
    --arg msg "ci: ${APP} ${ENVIRONMENT} -> ${GITHUB_SHA::7}" \
    --arg content "$CONTENT_B64" \
    --arg sha "$SHA" \
    --arg branch "main" \
    '{message:$msg, content:$content, branch:$branch} + (if $sha != "" then {sha:$sha} else {} end)')

METHOD=POST
if [ -n "$SHA" ]; then METHOD=PUT; fi

curl -fsS -u "gitea_admin:${GITEA_TOKEN}" \
    -X "$METHOD" \
    "http://gitea-http.gitea.svc.cluster.local:3000/api/v1/repos/openchoreo/platform-config/contents/${PATH_IN_REPO}" \
    -H 'Content-Type: application/json' \
    -d "$PAYLOAD"
