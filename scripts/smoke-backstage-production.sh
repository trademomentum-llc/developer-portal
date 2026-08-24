#!/usr/bin/env bash
# scripts/smoke-backstage-production.sh
# Verify Backstage production backend is reachable.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin backstage-production

BACKEND_URL="${BACKSTAGE_PROD_BACKEND:-http://localhost:7009}"

for i in $(seq 1 60); do
    status=$(curl -s -o /dev/null -w "%{http_code}" "${BACKEND_URL}/" 2>/dev/null || true)
    if [ "${status}" = "200" ]; then
        break
    fi
    sleep 2
done

status=$(curl -s -o /dev/null -w "%{http_code}" "${BACKEND_URL}/" 2>/dev/null || true)
if [ "${status}" != "200" ]; then
    echo "FAIL: Backstage production backend returned ${status}" >&2
    smoke_json_count fail
    exit 1
fi
echo "PASS: Backstage production backend reachable (${status})"
smoke_json_count pass
