#!/usr/bin/env bash
# scripts/setup-gitea-oauth.sh
# Create a Gitea OAuth application for Backstage local sign-in and store the
# credentials under ~/.rational-reserve/. This is idempotent: if an app named
# "Backstage Dev" already exists, the script reports its client id without
# creating a duplicate.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
RUNTIME_DIR="${HOME}/.rational-reserve"
GITEA_URL="${GITEA_URL:-http://localhost:3333}"
APP_NAME="Backstage Dev"
REDIRECT_URI="http://localhost:7008/api/auth/oauth2/handler/frame"

if [ ! -f "${RUNTIME_DIR}/m1-gitea-admin-password" ]; then
    echo "Gitea admin password not found at ${RUNTIME_DIR}/m1-gitea-admin-password" >&2
    exit 1
fi

GITEA_ADMIN_PASSWORD=$(cat "${RUNTIME_DIR}/m1-gitea-admin-password")

existing_id=$(curl -fsS -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    "${GITEA_URL}/api/v1/user/applications/oauth2" 2>/dev/null | \
    python3 -c "import sys,json; apps=json.load(sys.stdin); print(next((str(a['id']) for a in apps if a.get('name')=='${APP_NAME}'), ''))" 2>/dev/null || true)

if [ -n "${existing_id}" ]; then
    echo "Gitea OAuth app '${APP_NAME}' already exists (id=${existing_id})."
    client_id=$(curl -fsS -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
        "${GITEA_URL}/api/v1/user/applications/oauth2" 2>/dev/null | \
        python3 -c "import sys,json; apps=json.load(sys.stdin); print(next((a['client_id'] for a in apps if a.get('name')=='${APP_NAME}'), ''))" 2>/dev/null || true)
    echo "client_id: ${client_id}"
    echo "(If you need a fresh secret, delete the app in Gitea and rerun this script.)"
    exit 0
fi

response=$(curl -fsS -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST "${GITEA_URL}/api/v1/user/applications/oauth2" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"${APP_NAME}\",\"redirect_uris\":[\"${REDIRECT_URI}\"]}" 2>/dev/null)

client_id=$(echo "${response}" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("client_id",""))')
client_secret=$(echo "${response}" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("client_secret",""))')

if [ -z "${client_id}" ] || [ -z "${client_secret}" ]; then
    echo "Failed to create Gitea OAuth app. Response: ${response}" >&2
    exit 1
fi

printf '%s' "${client_id}" > "${RUNTIME_DIR}/backstage-oauth-client-id"
printf '%s' "${client_secret}" > "${RUNTIME_DIR}/backstage-oauth-client-secret"
chmod 600 "${RUNTIME_DIR}/backstage-oauth-client-id" "${RUNTIME_DIR}/backstage-oauth-client-secret"

echo "Created Gitea OAuth app '${APP_NAME}'."
echo "client_id saved to: ${RUNTIME_DIR}/backstage-oauth-client-id"
echo "client_secret saved to: ${RUNTIME_DIR}/backstage-oauth-client-secret"
echo "redirect_uri: ${REDIRECT_URI}"
