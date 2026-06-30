#!/usr/bin/env bash
# Commit a post-deploy Infracost artifact into platform-config/cost-artifacts/<app>/<env>/latest.json
set -euo pipefail

ENVIRONMENT=$1
APP=$2
ARTIFACT_JSON=$3

CONTENT_B64=$(base64 -w 0 < "$ARTIFACT_JSON" 2>/dev/null || base64 < "$ARTIFACT_JSON")
PATH_IN_REPO="cost-artifacts/${APP}/${ENVIRONMENT}/latest.json"

# Tolerate 404 when the file does not exist yet.
SHA=$(curl -sS -u "gitea_admin:${GITEA_TOKEN}" \
    "http://gitea-http.gitea.svc.cluster.local:3000/api/v1/repos/openchoreo/platform-config/contents/${PATH_IN_REPO}" \
    2>/dev/null | jq -r '.sha // empty' 2>/dev/null || true)

PAYLOAD=$(jq -n \
    --arg msg "ci: cost artifact ${APP} ${ENVIRONMENT} -> ${GITHUB_SHA::7}" \
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
