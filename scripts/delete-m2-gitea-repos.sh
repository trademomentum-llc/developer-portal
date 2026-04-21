#!/usr/bin/env bash
set -uo pipefail

GITEA_URL=${GITEA_URL:-http://localhost:3002}
ADMIN_USER=gitea_admin
ADMIN_PASS=$(cat ~/.rational-reserve/m1-gitea-admin-password 2>/dev/null || echo)

for repo in platform-addons platform-config hello-m2; do
    curl -fsS -u "${ADMIN_USER}:${ADMIN_PASS}" \
        -X DELETE "${GITEA_URL}/api/v1/repos/openchoreo/${repo}" >/dev/null 2>&1 || true
done
echo "m2 gitea repos deleted"
