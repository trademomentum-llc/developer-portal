#!/usr/bin/env bash
# scripts/smoke-all.sh
# Run the full M2/M3/M4 smoke suite and report a combined result.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin all

WITH_OPENBAO_RESTART=0
for arg in ${SMOKE_JSON_ARGS[@]+"${SMOKE_JSON_ARGS[@]}"}; do
    case "$arg" in
        --with-openbao-restart) WITH_OPENBAO_RESTART=1 ;;
        *) echo "smoke-all: unknown argument: $arg" >&2; exit 2 ;;
    esac
done

# FR-34: child suites append their own records to the same JSONL file;
# this suite adds one aggregate record over the suite results.
if [ -n "${SMOKE_JSON_OUT}" ]; then
    export SMOKE_JSON_OUT
fi

SUITES=(auth m2 m3 m4 security)
FAILED=()

for suite in "${SUITES[@]}"; do
    script="${ROOT_DIR}/scripts/smoke-${suite}.sh"
    echo "=== Running smoke-${suite}.sh ==="
    if "$script"; then
        echo "=== smoke-${suite}.sh PASSED ==="
        smoke_json_count pass
    else
        echo "=== smoke-${suite}.sh FAILED ===" >&2
        FAILED+=("$suite")
        smoke_json_count fail
    fi
    echo
done

# Production Backstage smoke is separate because it requires PostgreSQL.
script="${ROOT_DIR}/scripts/smoke-backstage-production.sh"
echo "=== Running smoke-backstage-production.sh ==="
if "$script"; then
    echo "=== smoke-backstage-production.sh PASSED ==="
    smoke_json_count pass
else
    echo "=== smoke-backstage-production.sh FAILED ===" >&2
    FAILED+=("backstage-production")
    smoke_json_count fail
fi
echo

# G5 FR-4 lane (opt-in, heavy, serialized last): deletes openbao-0 and
# asserts the four M2 keys survive with no reseed.
if [ "$WITH_OPENBAO_RESTART" = "1" ]; then
    echo "=== Running smoke-openbao.sh --with-restart (FR-4 lane) ==="
    if "${ROOT_DIR}/scripts/smoke-openbao.sh" --with-restart; then
        echo "=== smoke-openbao.sh --with-restart PASSED ==="
        smoke_json_count pass
    else
        echo "=== smoke-openbao.sh --with-restart FAILED ===" >&2
        FAILED+=("openbao-restart")
        smoke_json_count fail
    fi
    echo
fi

if [ ${#FAILED[@]} -eq 0 ]; then
    echo "ALL SMOKE SUITES PASSED (AUTH, M2, M3, M4, SECURITY, BACKSTAGE-PRODUCTION)"
    exit 0
else
    echo "SMOKE FAILURES: ${FAILED[*]}" >&2
    exit 1
fi
