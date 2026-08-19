#!/usr/bin/env bash
set -uo pipefail
PID_FILE="$HOME/.rational-reserve/m1-backstage-dev.pid"
if [ -f "$PID_FILE" ]; then
    kill "$(cat "$PID_FILE")" 2>/dev/null || true
    rm -f "$PID_FILE"
fi
pkill -f "backstage-cli repo start" 2>/dev/null || true
# Wave 0: reap the host-side kubectl proxy that start-backstage.sh manages
# for the Backstage kubernetes plugin (SEC-PLANE-WAVE0-TECH-001 section 5).
KUBECTL_PROXY_PID_FILE="$HOME/.rational-reserve/kubectl-proxy-8001.pid"
if [ -f "$KUBECTL_PROXY_PID_FILE" ]; then
    kill "$(cat "$KUBECTL_PROXY_PID_FILE")" 2>/dev/null || true
    rm -f "$KUBECTL_PROXY_PID_FILE"
fi
pkill -f "kubectl.*proxy.*--port=8001" 2>/dev/null || true
echo "Backstage stopped"
