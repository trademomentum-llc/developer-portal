#!/usr/bin/env bash
set -uo pipefail
PID_FILE="$HOME/.rational-reserve/m1-backstage-dev.pid"
if [ -f "$PID_FILE" ]; then
    kill "$(cat "$PID_FILE")" 2>/dev/null || true
    rm -f "$PID_FILE"
fi
pkill -f "backstage-cli repo start" 2>/dev/null || true
echo "Backstage stopped"
