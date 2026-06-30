#!/usr/bin/env bash
# scripts/smoke-all.sh
# Run the full M2/M3/M4 smoke suite and report a combined result.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SUITES=(m2 m3 m4)
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

if [ ${#FAILED[@]} -eq 0 ]; then
    echo "ALL SMOKE SUITES PASSED (M2, M3, M4)"
    exit 0
else
    echo "SMOKE FAILURES: ${FAILED[*]}" >&2
    exit 1
fi
