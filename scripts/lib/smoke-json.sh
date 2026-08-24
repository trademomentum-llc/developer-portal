#!/usr/bin/env bash
# scripts/lib/smoke-json.sh
#
# FR-34: shared machine-readable result emission for the smoke suites.
# A suite that sources this library gains:
#
#   - a --json <path> / --json=<path> flag (or the SMOKE_JSON_OUT env var)
#   - one JSON Lines record appended to that path on exit:
#       {"suite":"m3","passed":12,"failed":0,"skipped":1,"ts":"...","git_sha":"..."}
#
# Usage in a suite:
#
#   source "${ROOT}/scripts/lib/smoke-json.sh"
#   smoke_json_parse_args "$@"; set -- ${SMOKE_JSON_ARGS[@]+"${SMOKE_JSON_ARGS[@]}"}
#   smoke_json_begin <suite-name>          # installs the EXIT trap
#   ...
#   smoke_json_count pass|fail|skip        # per check, where the suite counts
#
# Suites that already install their own EXIT trap (port-forward cleanup)
# call `smoke_json_begin <name> no-trap` instead and invoke smoke_json_emit
# from inside their existing trap.
#
# Meta-suites (smoke-m2, smoke-all) additionally `export SMOKE_JSON_OUT` so
# each child suite appends its own record to the same JSONL file.
#
# The record is written with printf (no jq dependency): suite names are
# fixed tokens chosen by the suites themselves, counts are integers, ts is
# UTC ISO-8601, and git_sha is the developer-portal HEAD (or "unknown").
# The emission never changes the suite's exit status and never fails the
# suite: a broken artifact path degrades to a stderr warning.

SMOKE_JSON_OUT="${SMOKE_JSON_OUT:-}"
SMOKE_JSON_SUITE=""
SMOKE_JSON_PASSED=0
SMOKE_JSON_FAILED=0
SMOKE_JSON_SKIPPED=0
SMOKE_JSON_EMITTED=0

# Remaining (non-json) arguments after smoke_json_parse_args.
SMOKE_JSON_ARGS=()

smoke_json_parse_args() {
    SMOKE_JSON_ARGS=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --json)
                if [[ $# -lt 2 ]]; then
                    echo "smoke-json: --json requires a path argument" >&2
                    exit 2
                fi
                SMOKE_JSON_OUT="$2"
                shift 2
                ;;
            --json=*)
                SMOKE_JSON_OUT="${1#*=}"
                shift
                ;;
            *)
                SMOKE_JSON_ARGS+=("$1")
                shift
                ;;
        esac
    done
}

smoke_json_count() {
    case "${1:-}" in
        pass) SMOKE_JSON_PASSED=$((SMOKE_JSON_PASSED + 1)) ;;
        fail) SMOKE_JSON_FAILED=$((SMOKE_JSON_FAILED + 1)) ;;
        skip) SMOKE_JSON_SKIPPED=$((SMOKE_JSON_SKIPPED + 1)) ;;
    esac
}

smoke_json_begin() {
    SMOKE_JSON_SUITE="$1"
    if [ "${2:-}" != "no-trap" ]; then
        trap smoke_json_exit_trap EXIT
    fi
}

smoke_json_exit_trap() {
    # Preserve the suite's exit status; only append the record.
    local code=$?
    # Honesty guard: a suite that failed on an uncounted path (set -e abort
    # outside any smoke_json_count fail) must not record an all-pass.
    if [ "${code}" -ne 0 ] && [ "${SMOKE_JSON_FAILED}" -eq 0 ]; then
        SMOKE_JSON_FAILED=$((SMOKE_JSON_FAILED + 1))
    fi
    smoke_json_emit || true
    # shellcheck disable=SC2086
    return ${code}
}

smoke_json_emit() {
    [ -z "${SMOKE_JSON_OUT}" ] && return 0
    [ "${SMOKE_JSON_EMITTED}" -eq 1 ] && return 0
    SMOKE_JSON_EMITTED=1

    local root ts git_sha dir
    root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
    ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    git_sha="$(git -C "${root}" rev-parse HEAD 2>/dev/null || echo unknown)"
    dir="$(dirname "${SMOKE_JSON_OUT}")"
    if ! mkdir -p "${dir}" 2>/dev/null; then
        echo "smoke-json: cannot create ${dir}; skipping JSON record" >&2
        return 0
    fi

    printf '{"suite":"%s","passed":%d,"failed":%d,"skipped":%d,"ts":"%s","git_sha":"%s"}\n' \
        "${SMOKE_JSON_SUITE}" "${SMOKE_JSON_PASSED}" "${SMOKE_JSON_FAILED}" \
        "${SMOKE_JSON_SKIPPED}" "${ts}" "${git_sha}" >> "${SMOKE_JSON_OUT}" \
        || echo "smoke-json: cannot write ${SMOKE_JSON_OUT}" >&2
}
