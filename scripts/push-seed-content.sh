#!/usr/bin/env bash
# Push a local directory tree as file commits to a Gitea repo via the contents API.
set -euo pipefail

REPO=$1
SRC=$2
GITEA_URL=${GITEA_URL:-http://localhost:3002}
ADMIN_USER=gitea_admin
ADMIN_PASS=$(cat ~/.rational-reserve/m1-gitea-admin-password)

push_file() {
    local relpath=$1
    local fullpath=$2
    local content_b64
    content_b64=$(base64 -w 0 < "$fullpath" 2>/dev/null || base64 < "$fullpath")
    curl -fsS -u "${ADMIN_USER}:${ADMIN_PASS}" \
        -X POST "${GITEA_URL}/api/v1/repos/openchoreo/${REPO}/contents/${relpath}" \
        -H 'Content-Type: application/json' \
        -d "{\"message\":\"M2 seed\",\"content\":\"${content_b64}\",\"branch\":\"main\"}" \
        >/dev/null 2>&1 || true
    # 409 is expected on re-run when the file exists; ignore
}

cd "$SRC"
find . -type f -not -path '*/\.git/*' | while read -r f; do
    rel=${f#./}
    push_file "$rel" "$f"
done

echo "seeded content for openchoreo/${REPO}"
