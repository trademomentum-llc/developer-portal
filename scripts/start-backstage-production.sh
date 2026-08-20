#!/usr/bin/env bash
# scripts/start-backstage-production.sh
# Start Backstage locally using the production config and PostgreSQL backend.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
RUNTIME_DIR="${HOME}/.rational-reserve"
CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"
NODE24_BIN="/opt/homebrew/opt/node@24/bin"

if [ -d "$NODE24_BIN" ]; then
    export PATH="$NODE24_BIN:$PATH"
fi

if ! kubectl --context "$CONTEXT" get ns backstage >/dev/null 2>&1; then
    echo "ERROR: PostgreSQL namespace not found; run scripts/install-backstage-production.sh first" >&2
    exit 1
fi

POSTGRES_PASSWORD=$(kubectl --context "$CONTEXT" -n backstage get secret postgres-backstage -o jsonpath='{.data.password}' | base64 -d)

if [ -z "${POSTGRES_PASSWORD}" ]; then
    echo "ERROR: Could not read PostgreSQL password" >&2
    exit 1
fi

# Forward PostgreSQL to a stable local port so Backstage can connect from the host.
LOCAL_POSTGRES_PORT=30432
nohup kubectl --context "$CONTEXT" -n backstage port-forward svc/postgres-postgresql "${LOCAL_POSTGRES_PORT}:5432" \
    > "${RUNTIME_DIR}/postgres-portforward.log" 2>&1 &
PF_PID=$!
disown "$PF_PID"
echo "$PF_PID" > "${RUNTIME_DIR}/postgres-portforward.pid"

for i in $(seq 1 30); do
    if curl -fsS "http://127.0.0.1:${LOCAL_POSTGRES_PORT}/" >/dev/null 2>&1 || [ -f "${RUNTIME_DIR}/postgres-ready" ]; then
        break
    fi
    sleep 1
done

if [ ! -f "${RUNTIME_DIR}/backstage-backend-secret" ]; then
    openssl rand -base64 32 > "${RUNTIME_DIR}/backstage-backend-secret"
    chmod 600 "${RUNTIME_DIR}/backstage-backend-secret"
fi

if [ ! -f "${RUNTIME_DIR}/backstage-oauth-client-id" ]; then
    echo "ERROR: Gitea OAuth client id not found; run scripts/setup-gitea-oauth.sh" >&2
    exit 1
fi

export BACKEND_SECRET="$(cat "${RUNTIME_DIR}/backstage-backend-secret")"
export POSTGRES_HOST="127.0.0.1"
export POSTGRES_PORT="${LOCAL_POSTGRES_PORT}"
export POSTGRES_USER="backstage"
export POSTGRES_PASSWORD
export POSTGRES_DATABASE="backstage"
export GITEA_OAUTH_CLIENT_ID="$(cat "${RUNTIME_DIR}/backstage-oauth-client-id")"
export GITEA_OAUTH_CLIENT_SECRET="$(cat "${RUNTIME_DIR}/backstage-oauth-client-secret")"

# app-config.yaml's integrations.gitea block (loaded alongside the production
# config) references ${GITEA_ADMIN_PASSWORD}; mirror start-backstage.sh.
if [ ! -f "${RUNTIME_DIR}/m1-gitea-admin-password" ]; then
    echo "ERROR: Gitea admin password not found at ${RUNTIME_DIR}/m1-gitea-admin-password" >&2
    exit 1
fi
export GITEA_ADMIN_PASSWORD="$(cat "${RUNTIME_DIR}/m1-gitea-admin-password")"

export GITEA_HOSTNAME="localhost:3333"
export APP_BASE_URL="http://localhost:3002"
export BACKEND_BASE_URL="http://localhost:7009"
export APP_CONFIG_backend_listen_port="7009"

# Keep CORS permissive for local production validation.
export APP_CONFIG_backend_cors_origin="http://localhost:3002"

cd "${ROOT_DIR}/backstage"

if [ ! -f "packages/app/dist/index.html" ] || [ ! -d "packages/backend/dist" ]; then
    echo "Building Backstage production bundles..."
    yarn build:all
fi

LOG_FILE="${RUNTIME_DIR}/backstage-production.log"
nohup yarn --cwd packages/backend start \
    --config ../../app-config.yaml \
    --config ../../app-config.production.yaml \
    > "$LOG_FILE" 2>&1 &
PID=$!
disown "$PID"
echo "$PID" > "${RUNTIME_DIR}/backstage-production.pid"
echo "Backstage production starting with PID $PID"
echo "Backend: ${BACKEND_BASE_URL}"
echo "Logs: ${LOG_FILE}"
