#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GITEA_URL=${GITEA_URL:-http://localhost:3002}
ADMIN_USER=gitea_admin
ADMIN_PASS=$(cat ~/.rational-reserve/m1-gitea-admin-password)

api() {
    local method=$1; shift
    local path=$1; shift
    curl -fsS -u "${ADMIN_USER}:${ADMIN_PASS}" \
        -X "$method" "${GITEA_URL}${path}" \
        -H 'Content-Type: application/json' "$@"
}

# Create organization (idempotent)
if ! api GET /api/v1/orgs/openchoreo >/dev/null 2>&1; then
    api POST /api/v1/orgs -d '{"username":"openchoreo","full_name":"OpenChoreo Platform"}'
fi

# Create three repos
for repo in platform-addons platform-config hello-m2; do
    if ! api GET "/api/v1/repos/openchoreo/$repo" >/dev/null 2>&1; then
        api POST /api/v1/orgs/openchoreo/repos -d "{
          \"name\":\"$repo\",
          \"private\":false,
          \"auto_init\":true,
          \"default_branch\":\"main\"
        }" >/dev/null
        echo "created openchoreo/$repo"
    fi
done

# Branch protection on platform-addons and platform-config
for repo in platform-addons platform-config; do
    api POST "/api/v1/repos/openchoreo/$repo/branch_protections" -d '{
      "branch_name":"main",
      "enable_push":false,
      "required_approvals":1,
      "enable_merge_whitelist":true,
      "merge_whitelist_usernames":["gitea_admin"]
    }' >/dev/null 2>&1 || true
done

# Org-level Actions secret for the paved-path deploy loop (FR-13). Verified
# live on Gitea 1.25.4: org secrets exist (PUT/GET/DELETE
# /api/v1/orgs/openchoreo/actions/secrets/*) and are injected into every
# workflow of every repo in the org, so a scaffolded repo inherits
# PLATFORM_CONFIG_TOKEN without a per-repo seeding step. Secrets are
# write-only over the API, so this is create-if-absent: to rotate, DELETE the
# org secret first, then re-run this script.
if api GET /api/v1/orgs/openchoreo/actions/secrets 2>/dev/null \
    | jq -e '.[] | select(.name=="PLATFORM_CONFIG_TOKEN")' >/dev/null 2>&1; then
    echo "org secret PLATFORM_CONFIG_TOKEN present (write-only; skipping)"
else
    # Fresh least-privilege gitea_admin token scoped to repository write --
    # exactly what scripts/ci/commit-to-platform-config.sh needs. Drop any
    # stale token of the same name first so re-seeds do not accumulate.
    TOKEN_NAME=platform-config-org-secret
    api GET /api/v1/users/gitea_admin/tokens | jq -r \
        ".[] | select(.name==\"$TOKEN_NAME\") | .id" \
        | while read -r id; do
            api DELETE "/api/v1/users/gitea_admin/tokens/$id" >/dev/null
            echo "deleted stale token $TOKEN_NAME id=$id"
        done
    ORG_TOKEN=$(api POST /api/v1/users/gitea_admin/tokens -d "{
      \"name\":\"$TOKEN_NAME\",
      \"scopes\":[\"write:repository\"]
    }" | jq -r '.sha1')
    if [ -z "$ORG_TOKEN" ] || [ "$ORG_TOKEN" = "null" ]; then
        echo "ERROR: failed to create $TOKEN_NAME token" >&2
        exit 1
    fi
    api PUT /api/v1/orgs/openchoreo/actions/secrets/PLATFORM_CONFIG_TOKEN \
        -d "{\"data\":\"$ORG_TOKEN\"}" >/dev/null
    echo "created org secret PLATFORM_CONFIG_TOKEN (token $TOKEN_NAME, scope write:repository)"
fi

# Push seed content
for repo in platform-addons platform-config hello-m2; do
    if [ -d "$ROOT/seed-repos/$repo" ]; then
        "$ROOT/scripts/push-seed-content.sh" "$repo" "$ROOT/seed-repos/$repo"
    fi
done

echo "gitea seeded"
