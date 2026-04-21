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

# Push seed content
for repo in platform-addons platform-config hello-m2; do
    if [ -d "$ROOT/seed-repos/$repo" ]; then
        "$ROOT/scripts/push-seed-content.sh" "$repo" "$ROOT/seed-repos/$repo"
    fi
done

echo "gitea seeded"
