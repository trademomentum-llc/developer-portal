#!/usr/bin/env bash
#
# test-smoke-codeql-contract.sh
# Inverse-proof tests for the Lane D advanced CodeQL workflow contract in
# scripts/smoke-security.sh (AGENTS.md "Inverse-proof testing").
#
# Usage:
#   ./scripts/tests/test-smoke-codeql-contract.sh
#
# Case A: the committed .github/workflows/codeql.yml satisfies the contract.
# Cases B+: tampered copies in a scratch dir (mktemp -d, removed on exit)
# must each produce the expected FAIL line, so a regression in any clause
# is observed to fail. The real workflow file is never modified.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SMOKE="${ROOT_DIR}/scripts/smoke-security.sh"
WF="${ROOT_DIR}/.github/workflows/codeql.yml"

pass() { echo " PASS $*"; }
fail() { echo " FAIL $*"; exit 1; }

SCRATCH="$(mktemp -d)"
trap 'rm -rf "${SCRATCH}"' EXIT

# contract_fails <workflow-path>: prints the smoke suite's FAIL lines that
# belong to the CodeQL contract (other lanes are environment-dependent).
contract_fails() {
    CODEQL_WORKFLOW="$1" bash "${SMOKE}" 2>&1 \
        | grep -E '^FAIL: (advanced CodeQL|codeql\.yml|job name|job permissions|matrix entry|init queries|analyze category|github/codeql-action|actions/checkout)' || true
}

out=$(contract_fails "${WF}")
[ -z "${out}" ] && pass "A: committed codeql.yml satisfies the contract" || fail "A: committed codeql.yml rejected: ${out}"

# case <id> <expected FAIL substring> <sed expression or "ABSENT">
case_() {
    local id=$1 want=$2 expr=$3 p="${SCRATCH}/$1.yml"
    if [ "${expr}" = ABSENT ]; then p="${SCRATCH}/absent.yml"; else sed -E "${expr}" "${WF}" > "${p}"; fi
    cmp -s "${p}" "${WF}" && fail "${id}: tamper expression did not change the file"
    out=$(contract_fails "${p}")
    if grep -qF -- "${want}" <<< "${out}"; then pass "${id}: fails with '${want}'"; else fail "${id}: expected '${want}', got: ${out:-<no contract FAIL>}"; fi
}

case_ B-absent       "advanced CodeQL workflow missing" ABSENT
case_ C-tagpin       "github/codeql-action/init missing or not full-SHA pinned" 's/init@[0-9a-f]{40}/init@v4/'
case_ D-checkout     "actions/checkout missing or not full-SHA pinned" 's/checkout@[0-9a-f]{40}/checkout@v7/'
case_ E-jobname      "job name does not render" 's/name: Analyze \(\$\{\{ matrix\.language \}\}\)/name: CodeQL-${{ matrix.language }}/'
case_ F-queries      "init queries is not security-extended" 's/security-extended/security-and-quality/'
case_ G-gomode       "matrix entry go with build-mode autobuild missing" '/language: go/{n;s/autobuild/none/}'
case_ H-nopython     "matrix entry python with build-mode none missing" '/language: python/,+1d'
case_ I-nojs         "matrix entry javascript-typescript with build-mode none missing" '/language: javascript-typescript/,+1d'
case_ J-secevents    "job permissions missing 'security-events: write'" 's/security-events: write/security-events: read/'
case_ K-extrawrite   "grants a write permission other than security-events" 's/actions: read/actions: write/'
case_ L-upload       "disables SARIF upload" 's|(category: .*)|\1\n          upload: never|'
case_ M-noschedule   "missing trigger schedule" '/schedule:/,+1d'
case_ N-nopr         "missing trigger pull_request" '/pull_request:/,+1d'
case_ O-category     "analyze category is not" 's|/language:\$\{\{ matrix\.language \}\}|/language:all|'

echo "All CodeQL contract cases passed."
