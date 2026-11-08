#!/usr/bin/env bash
# Commit a smoke-suite result artifact (FR-34) into platform-config:
#   test-artifacts/<app>/<env>/<run-id>.jsonl  (append-only per-run history, NFR-08)
#   test-artifacts/<app>/<env>/latest.jsonl    (portal read path)
# Both files carry identical content: one JSON Lines record per smoke suite
# ({"suite","passed","failed","skipped","ts","git_sha"}), produced by running
# the suites with SMOKE_JSON_OUT or --json (see scripts/lib/smoke-json.sh).
# The portal's TestResultsCard reads latest.jsonl through the gitea-actions
# proxy and renders pass/fail/skip per suite.
#
# Follows the commit-cost-artifact.sh / commit-security-artifacts.sh pattern:
# tolerate-404 GET-then-POST/PUT against one repo path, POST on create, PUT
# on update. Honest note carried from those precedents: artifacts are
# committed AS gitea_admin; per-run provenance lives inside the records
# (git_sha, ts) and in the commit message (run id).
set -euo pipefail

ENVIRONMENT=$1
APP=$2
ARTIFACT_JSONL=$3
RUN_ID=$4

API="${GITEA_API_BASE:-http://gitea-http.gitea.svc.cluster.local:3000/api/v1/repos/openchoreo/platform-config/contents}"

CONTENT_B64=$(base64 -w 0 < "$ARTIFACT_JSONL" 2>/dev/null || base64 < "$ARTIFACT_JSONL")

write_path() {
    local path_in_repo=$1

    local sha
    sha=$(curl -sS -u "gitea_admin:${GITEA_TOKEN}" \
        "${API}/${path_in_repo}" \
        2>/dev/null | jq -r '.sha // empty' 2>/dev/null || true)

    local payload
    payload=$(jq -n \
        --arg msg "ci: test artifact ${APP} ${ENVIRONMENT} -> ${GITHUB_SHA::7} (run ${RUN_ID})" \
        --arg content "$CONTENT_B64" \
        --arg sha "$sha" \
        --arg branch "main" \
        '{message:$msg, content:$content, branch:$branch} + (if $sha != "" then {sha:$sha} else {} end)')

    local method=POST
    if [ -n "$sha" ]; then method=PUT; fi

    curl -fsS -u "gitea_admin:${GITEA_TOKEN}" \
        -X "$method" \
        "${API}/${path_in_repo}" \
        -H 'Content-Type: application/json' \
        -d "$payload"
}

write_path "test-artifacts/${APP}/${ENVIRONMENT}/${RUN_ID}.jsonl"
write_path "test-artifacts/${APP}/${ENVIRONMENT}/latest.jsonl"
