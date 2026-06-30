#!/usr/bin/env bash
# scripts/smoke-m4.sh
# Verify the M4 cost visibility plane is reachable and returns allocation data.
set -euo pipefail

PORT=29003
PF_PID=""
cleanup() {
    if [ -n "${PF_PID}" ]; then
        kill "${PF_PID}" 2>/dev/null || true
        wait "${PF_PID}" 2>/dev/null || true
    fi
}
trap cleanup EXIT

PF_PID=$(kubectl --context k3d-openchoreo -n opencost port-forward svc/opencost "${PORT}:9090" >/tmp/smoke-m4-pf.log 2>&1 & echo $!)
sleep 3

echo "Checking OpenCost allocation endpoint ..."
if curl -fsS "http://localhost:${PORT}/model/allocation?window=today&aggregate=namespace" >/tmp/smoke-m4-cost.json 2>/dev/null; then
    echo "PASS: OpenCost returned allocation data"
    head -c 200 /tmp/smoke-m4-cost.json
    echo
else
    echo "FAIL: OpenCost did not return allocation data (see /tmp/smoke-m4-pf.log)" >&2
    exit 1
fi
