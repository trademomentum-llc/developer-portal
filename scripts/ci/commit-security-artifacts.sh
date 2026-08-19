#!/usr/bin/env bash
# Commit a combined security scan artifact into platform-config:
#   security-artifacts/<app>/<env>/<run-id>.json  (append-only per-run history, NFR-08)
#   security-artifacts/<app>/<env>/latest.json    (portal read path)
# Both files carry identical content. This step is reached only when every
# scan gate in the job has PASSED (the Trivy/OSV gate steps stop the job
# earlier on failure), so latest.json always describes the last passing push
# (honest-status rule, SEC-PLANE-WAVE0-TECH-001 section 3.3).
#
# Honest note (carried from the commit-cost-artifact.sh precedent, accepted
# existing pattern): artifacts are committed AS gitea_admin. Per-artifact
# authorship metadata lives inside the artifact JSON (git_sha, run_id).
set -euo pipefail

ENVIRONMENT=$1
APP=$2
ARTIFACT_JSON=$3
RUN_ID=$4

API="http://gitea-http.gitea.svc.cluster.local:3000/api/v1/repos/openchoreo/platform-config/contents"

CONTENT_B64=$(base64 -w 0 < "$ARTIFACT_JSON" 2>/dev/null || base64 < "$ARTIFACT_JSON")

# Tolerate-404 GET-then-POST/PUT against one repo path, per the
# commit-cost-artifact.sh precedent: POST on create, PUT on update.
write_path() {
    local path_in_repo=$1

    local sha
    sha=$(curl -sS -u "gitea_admin:${GITEA_TOKEN}" \
        "${API}/${path_in_repo}" \
        2>/dev/null | jq -r '.sha // empty' 2>/dev/null || true)

    local payload
    payload=$(jq -n \
        --arg msg "ci: security artifact ${APP} ${ENVIRONMENT} -> ${GITHUB_SHA::7} (run ${RUN_ID})" \
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

write_path "security-artifacts/${APP}/${ENVIRONMENT}/${RUN_ID}.json"
write_path "security-artifacts/${APP}/${ENVIRONMENT}/latest.json"
