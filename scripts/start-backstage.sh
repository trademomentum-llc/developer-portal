#!/usr/bin/env bash
# scripts/start-backstage.sh -- Start Backstage dev server.
set -euo pipefail

# Fixed 2026-08-18 (SEC-PLANE-WAVE0-TECH-001, Lane B): this previously pointed
# at /Users/nnos/Projects/developer-portal, which does not exist; the real
# checkout lives under ~/Projects/Sovereign/.
BACKSTAGE_DIR="/Users/nnos/Projects/Sovereign/developer-portal/backstage"
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

# Load Gitea OAuth credentials for the Backstage auth provider if they exist.
if [ -f "$RUNTIME_DIR/backstage-oauth-client-id" ]; then
    export GITEA_OAUTH_CLIENT_ID=$(cat "$RUNTIME_DIR/backstage-oauth-client-id")
fi
if [ -f "$RUNTIME_DIR/backstage-oauth-client-secret" ]; then
    export GITEA_OAUTH_CLIENT_SECRET=$(cat "$RUNTIME_DIR/backstage-oauth-client-secret")
fi

ensure_gitea_port() {
    local local_port="$1"
    local pid_file="$RUNTIME_DIR/gitea-portforward-${local_port}.pid"
    if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/api/v1/version" | grep -q '^200$'; then
        return 0
    fi
    if [ -f "$pid_file" ]; then
        kill "$(cat "$pid_file")" 2>/dev/null || true
        rm -f "$pid_file"
    fi
    pkill -f "kubectl.*port-forward.*gitea-http.*:${local_port}" 2>/dev/null || true
    nohup kubectl --context k3d-openchoreo -n gitea port-forward "svc/gitea-http" "${local_port}:3000" \
        > "/tmp/gitea-portforward-${local_port}.log" 2>&1 &
    echo $! > "$pid_file"
    for i in $(seq 1 30); do
        if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/api/v1/version" | grep -q '^200$'; then
            return 0
        fi
        sleep 1
    done
    echo "Gitea port-forward on ${local_port} did not become ready" >&2
    return 1
}

# Wave 0 (SEC-PLANE-WAVE0-TECH-001 section 5): Backstage's kubernetes plugin
# reaches the cluster through this host-side proxy (authProvider
# localKubectlProxy, cluster k3d-openchoreo-local). Reaped by
# scripts/stop-backstage.sh.
ensure_kubectl_proxy() {
    local pid_file="$RUNTIME_DIR/kubectl-proxy-8001.pid"
    if curl -s -o /dev/null -w '%{http_code}' "http://localhost:8001/api" | grep -q '^200$'; then
        return 0
    fi
    if [ -f "$pid_file" ]; then
        kill "$(cat "$pid_file")" 2>/dev/null || true
        rm -f "$pid_file"
    fi
    pkill -f "kubectl.*proxy.*--port=8001" 2>/dev/null || true
    nohup kubectl --context k3d-openchoreo proxy --port=8001 \
        > "/tmp/kubectl-proxy-8001.log" 2>&1 &
    echo $! > "$pid_file"
    for i in $(seq 1 30); do
        if curl -s -o /dev/null -w '%{http_code}' "http://localhost:8001/api" | grep -q '^200$'; then
            return 0
        fi
        sleep 1
    done
    echo "kubectl proxy on 8001 did not become ready" >&2
    return 1
}

ensure_gitea_port 3333
ensure_gitea_port 3002
ensure_kubectl_proxy

ensure_opencost_port() {
    local local_port="$1"
    local pid_file="$RUNTIME_DIR/opencost-portforward-${local_port}.pid"
    if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/model/allocation?window=today&aggregate=namespace" | grep -q '^200$'; then
        return 0
    fi
    if [ -f "$pid_file" ]; then
        kill "$(cat "$pid_file")" 2>/dev/null || true
        rm -f "$pid_file"
    fi
    pkill -f "kubectl.*port-forward.*svc/opencost.*:${local_port}" 2>/dev/null || true
    nohup kubectl --context k3d-openchoreo -n opencost port-forward "svc/opencost" "${local_port}:9090" \
        > "/tmp/opencost-portforward-${local_port}.log" 2>&1 &
    echo $! > "$pid_file"
    for i in $(seq 1 30); do
        if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/model/allocation?window=today&aggregate=namespace" | grep -q '^200$'; then
            return 0
        fi
        sleep 1
    done
    echo "OpenCost port-forward on ${local_port} did not become ready" >&2
    return 1
}

ensure_opencost_port 29003

# SigNoz frontend forward for entity-page trace links (AGENTS.md port table,
# local 3301). Cluster evidence: svc/signoz in namespace signoz exposes the
# frontend on port 8080 (M4 networking requirements map 3301 -> signoz:8080).
# Reaped by scripts/stop-backstage.sh.
ensure_signoz_port() {
    local local_port="$1"
    local pid_file="$RUNTIME_DIR/signoz-portforward-${local_port}.pid"
    if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/" | grep -q '^200$'; then
        return 0
    fi
    if [ -f "$pid_file" ]; then
        kill "$(cat "$pid_file")" 2>/dev/null || true
        rm -f "$pid_file"
    fi
    pkill -f "kubectl.*port-forward.*svc/signoz.*:${local_port}" 2>/dev/null || true
    nohup kubectl --context k3d-openchoreo -n signoz port-forward "svc/signoz" "${local_port}:8080" \
        > "/tmp/signoz-portforward-${local_port}.log" 2>&1 &
    echo $! > "$pid_file"
    for i in $(seq 1 30); do
        if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/" | grep -q '^200$'; then
            return 0
        fi
        sleep 1
    done
    echo "SigNoz port-forward on ${local_port} did not become ready" >&2
    return 1
}

ensure_signoz_port 3301

# OpenChoreo control-plane API forward for the cards' api-base links
# (OpenChoreoOverviewCard/DeploymentCard default http://localhost:9090;
# README.md and docs/index.md document it). Cluster evidence:
# svc/openchoreo-api in namespace openchoreo-control-plane exposes 8080.
# Reaped by scripts/stop-backstage.sh.
ensure_openchoreo_api_port() {
    local local_port="$1"
    local pid_file="$RUNTIME_DIR/openchoreo-api-portforward-${local_port}.pid"
    # Probe /health: the API 404s on / (verified live), and a
    # service-specific endpoint avoids false-passing on unrelated
    # processes holding the port.
    if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/health" | grep -q '^200$'; then
        return 0
    fi
    if [ -f "$pid_file" ]; then
        kill "$(cat "$pid_file")" 2>/dev/null || true
        rm -f "$pid_file"
    fi
    pkill -f "kubectl.*port-forward.*svc/openchoreo-api.*:${local_port}" 2>/dev/null || true
    nohup kubectl --context k3d-openchoreo -n openchoreo-control-plane port-forward "svc/openchoreo-api" "${local_port}:8080" \
        > "/tmp/openchoreo-api-portforward-${local_port}.log" 2>&1 &
    echo $! > "$pid_file"
    for i in $(seq 1 30); do
        if curl -s -o /dev/null -w '%{http_code}' "http://localhost:${local_port}/health" | grep -q '^200$'; then
            return 0
        fi
        sleep 1
    done
    echo "OpenChoreo API port-forward on ${local_port} did not become ready" >&2
    return 1
}

ensure_openchoreo_api_port 9090

cd "$BACKSTAGE_DIR"

# Ensure local dev overrides exist; the tracked example supplies guest auth,
# disabled permission framework, and a persistent SQLite database directory.
if [ ! -f "app-config.local.yaml" ]; then
    if [ -f "app-config.local.yaml.example" ]; then
        cp "app-config.local.yaml.example" "app-config.local.yaml"
        echo "Created app-config.local.yaml from example"
    fi
fi

GITEA_ADMIN_PASSWORD=$(cat "$RUNTIME_DIR/m1-gitea-admin-password")
export GITEA_ADMIN_PASSWORD
if [ -f "$RUNTIME_DIR/m1-gitea-token" ]; then
    GITEA_ADMIN_TOKEN=$(cat "$RUNTIME_DIR/m1-gitea-token")
else
    GITEA_ADMIN_TOKEN="${GITEA_ADMIN_TOKEN:-dummy}"
fi
export GITEA_ADMIN_TOKEN
# FR-09: SigNoz API key for the /api/proxy/signoz backend proxy header.
# Seeded by scripts/install-m3.sh; empty means the cards render not-wired.
if [ -f "$RUNTIME_DIR/signoz-api-key" ]; then
    SIGNOZ_API_KEY=$(cat "$RUNTIME_DIR/signoz-api-key")
else
    SIGNOZ_API_KEY="${SIGNOZ_API_KEY:-}"
fi
export SIGNOZ_API_KEY
export HOST="$APP_HOST"
export PORT="$APP_PORT"
export APP_CONFIG_app_baseUrl="$APP_BASE_URL"
export APP_CONFIG_app_listen_host="$APP_HOST"
export APP_CONFIG_backend_baseUrl="$BACKEND_BASE_URL"
export APP_CONFIG_backend_listen_host="127.0.0.1"
export APP_CONFIG_backend_listen_port="$BACKEND_PORT"
# Only pin CORS to a custom app host; otherwise the app-config.yaml list
# (localhost + 127.0.0.1) is used so both URLs work.
if [ -n "${BACKSTAGE_APP_HOST:-}" ]; then
    export APP_CONFIG_backend_cors_origin="$APP_BASE_URL"
fi

LOG_FILE="${BACKSTAGE_LOG_FILE:-$RUNTIME_DIR/backstage-dev.log}"
nohup yarn start > "$LOG_FILE" 2>&1 &
BACKSTAGE_PID=$!
disown "$BACKSTAGE_PID"
echo "$BACKSTAGE_PID" > "$RUNTIME_DIR/m1-backstage-dev.pid"
echo "Backstage starting with PID $BACKSTAGE_PID"
echo "Logs: $LOG_FILE"
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
