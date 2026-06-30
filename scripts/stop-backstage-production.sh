#!/usr/bin/env bash
# scripts/stop-backstage-production.sh
set -uo pipefail
PID_FILE="$HOME/.rational-reserve/backstage-production.pid"
PF_FILE="$HOME/.rational-reserve/postgres-portforward.pid"
if [ -f "$PID_FILE" ]; then
    kill "$(cat "$PID_FILE")" 2>/dev/null || true
    rm -f "$PID_FILE"
fi
if [ -f "$PF_FILE" ]; then
    kill "$(cat "$PF_FILE")" 2>/dev/null || true
    rm -f "$PF_FILE"
fi
pkill -f "packages/backend.*app-config.production" 2>/dev/null || true
echo "Backstage production stopped"
