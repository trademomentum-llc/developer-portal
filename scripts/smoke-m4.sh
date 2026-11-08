#!/usr/bin/env bash
# scripts/smoke-m4.sh
# Verify the M4 cost visibility plane is reachable and returns allocation data.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"

PORT=29003
PF_PID=""
cleanup() {
    if [ -n "${PF_PID}" ]; then
        kill "${PF_PID}" 2>/dev/null || true
        wait "${PF_PID}" 2>/dev/null || true
    fi
    # FR-34: emit the JSON record from the suite's own trap (see
    # scripts/lib/smoke-json.sh, no-trap mode).
    if [ "${SMOKE_JSON_FAILED}" -eq 0 ] && [ "${SMOKE_JSON_PASSED}" -eq 0 ]; then
        SMOKE_JSON_FAILED=1
    fi
    smoke_json_emit || true
}
smoke_json_begin m4 no-trap
trap cleanup EXIT

PF_PID=$(kubectl --context k3d-openchoreo -n opencost port-forward svc/opencost "${PORT}:9090" >/tmp/smoke-m4-pf.log 2>&1 & echo $!)
sleep 3

echo "Checking OpenCost allocation endpoint ..."
if curl -fsS "http://localhost:${PORT}/model/allocation?window=today&aggregate=namespace" >/tmp/smoke-m4-cost.json 2>/dev/null; then
    echo "PASS: OpenCost returned allocation data"
    smoke_json_count pass
    head -c 200 /tmp/smoke-m4-cost.json
    echo
else
    echo "FAIL: OpenCost did not return allocation data (see /tmp/smoke-m4-pf.log)" >&2
    smoke_json_count fail
    exit 1
fi
