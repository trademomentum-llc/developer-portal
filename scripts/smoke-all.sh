#!/usr/bin/env bash
# scripts/smoke-all.sh
# Run the full M2/M3/M4 smoke suite and report a combined result.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SUITES=(auth m2 m3 m4 security)
FAILED=()

for suite in "${SUITES[@]}"; do
    script="${ROOT_DIR}/scripts/smoke-${suite}.sh"
    echo "=== Running smoke-${suite}.sh ==="
    if "$script"; then
        echo "=== smoke-${suite}.sh PASSED ==="
    else
        echo "=== smoke-${suite}.sh FAILED ===" >&2
        FAILED+=("$suite")
    fi
    echo
done

# Production Backstage smoke is separate because it requires PostgreSQL.
script="${ROOT_DIR}/scripts/smoke-backstage-production.sh"
echo "=== Running smoke-backstage-production.sh ==="
if "$script"; then
    echo "=== smoke-backstage-production.sh PASSED ==="
else
    echo "=== smoke-backstage-production.sh FAILED ===" >&2
    FAILED+=("backstage-production")
fi
echo

if [ ${#FAILED[@]} -eq 0 ]; then
    echo "ALL SMOKE SUITES PASSED (AUTH, M2, M3, M4, SECURITY, BACKSTAGE-PRODUCTION)"
    exit 0
else
    echo "SMOKE FAILURES: ${FAILED[*]}" >&2
    exit 1
fi
