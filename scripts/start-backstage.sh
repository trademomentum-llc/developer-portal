#!/usr/bin/env bash
# scripts/start-backstage.sh -- Start Backstage dev server.
set -euo pipefail

BACKSTAGE_DIR="/Users/nnos/Projects/developer-portal/backstage"
RUNTIME_DIR="${HOME}/.rational-reserve"
APP_HOST="${BACKSTAGE_APP_HOST:-127.0.0.1}"
APP_PORT="${BACKSTAGE_APP_PORT:-3001}"
BACKEND_PORT="${BACKSTAGE_BACKEND_PORT:-7008}"
APP_BASE_URL="${BACKSTAGE_APP_BASE_URL:-http://${APP_HOST}:${APP_PORT}}"
BACKEND_BASE_URL="${BACKSTAGE_BACKEND_BASE_URL:-http://${APP_HOST}:${BACKEND_PORT}}"
NODE24_BIN="/opt/homebrew/opt/node@24/bin"

if [ -d "$NODE24_BIN" ]; then
    export PATH="$NODE24_BIN:$PATH"
fi

cd "$BACKSTAGE_DIR"
GITEA_ADMIN_PASSWORD=$(cat "$RUNTIME_DIR/m1-gitea-admin-password")
export GITEA_ADMIN_PASSWORD
if [ -f "$RUNTIME_DIR/m1-gitea-token" ]; then
    GITEA_ADMIN_TOKEN=$(cat "$RUNTIME_DIR/m1-gitea-token")
else
    GITEA_ADMIN_TOKEN="${GITEA_ADMIN_TOKEN:-dummy}"
fi
export GITEA_ADMIN_TOKEN
export HOST="$APP_HOST"
export PORT="$APP_PORT"
export APP_CONFIG_app_baseUrl="$APP_BASE_URL"
export APP_CONFIG_backend_baseUrl="$BACKEND_BASE_URL"
export APP_CONFIG_backend_listen_port="$BACKEND_PORT"
export APP_CONFIG_backend_cors_origin="$APP_BASE_URL"

yarn start &
BACKSTAGE_PID=$!
echo "$BACKSTAGE_PID" > "$RUNTIME_DIR/m1-backstage-dev.pid"
echo "Backstage starting with PID $BACKSTAGE_PID"
echo "Waiting for $APP_BASE_URL ..."

for i in $(seq 1 90); do
    if curl -s -o /dev/null "$APP_BASE_URL"; then
        echo "Backstage is up at $APP_BASE_URL"
        exit 0
    fi
    sleep 2
done

echo "Backstage did not come up within 3 minutes"
exit 1
