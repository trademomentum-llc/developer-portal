#!/usr/bin/env bash
#
# test-check-handoff-fidelity.sh
# Integration tests for scripts/check-handoff-fidelity.sh
#
# Usage:
#   ./scripts/tests/test-check-handoff-fidelity.sh
#
# Every case runs in a throwaway scratch repo (mktemp -d) with a fake
# SESSION_HANDOFF.md. The real repository and the user's git config are
# never touched; scratch is removed on exit, so the suite is re-runnable.
# Cases B and C are the inverse-proof lanes: they exist to prove the
# freshness and remote-sync checks actually FAIL when they should.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CHECKER="${ROOT_DIR}/scripts/check-handoff-fidelity.sh"

BOLD="\033[1m"
GREEN="\033[32m"
RED="\033[31m"
RESET="\033[0m"

log() { echo -e "$*"; }
pass() { echo -e "${GREEN} PASS${RESET} $*"; }
fail() { echo -e "${RED} FAIL${RESET} $*"; exit 1; }
info() { echo -e "${BOLD}INFO${RESET} $*"; }

SCRATCH="$(mktemp -d)"
trap 'rm -rf "${SCRATCH}"' EXIT

info "Testing ${CHECKER}"
info "Scratch world: ${SCRATCH} (removed on exit)"

# make_repo DIR DATE -- scratch repo with one commit at DATE (ISO).
make_repo() {
    local dir="$1" date="$2"
    git init -q -b main "${dir}"
    git -C "${dir}" config user.email "fidelity-test@example.com"
    git -C "${dir}" config user.name "Fidelity Test"
    git -C "${dir}" config commit.gpgsign false
    echo x > "${dir}/f"
    git -C "${dir}" add f
    GIT_AUTHOR_DATE="${date}" GIT_COMMITTER_DATE="${date}" \
        git -C "${dir}" commit -q -m "feat: x"
}

# make_handoff DIR DATE -- minimal handoff carrying a Last updated line.
make_handoff() {
    local dir="$1" date="$2"
    cat > "${dir}/SESSION_HANDOFF.md" <<EOF
# SESSION HANDOFF

**Last updated:** ${date}
**Reason for handoff:** test fixture
EOF
}

# run_checker DIR -- invoke the checker against DIR, capture output+status.
run_checker() {
    local dir="$1"
    OUT="$(HANDOFF_FIDELITY_ROOT="${dir}" "${CHECKER}" 2>&1)" && STATUS=0 || STATUS=$?
}

# ---- Case A: fresh handoff, no remotes -> overall PASS --------------------

info "Case A: handoff Last updated == HEAD date, no remotes"
REPO_A="${SCRATCH}/a"
make_repo "${REPO_A}" "2026-08-23T12:00:00"
make_handoff "${REPO_A}" "2026-08-23"
run_checker "${REPO_A}"
[ "${STATUS}" -eq 0 ] || fail "A: expected exit 0, got ${STATUS}: ${OUT}"
echo "${OUT}" | grep -q "PASS freshness" || fail "A: missing PASS freshness: ${OUT}"
echo "${OUT}" | grep -q "SKIP remote-sync" || fail "A: expected SKIP remote-sync with no remotes: ${OUT}"
pass "A: fresh handoff passes, remote lane skips with no remotes"

# ---- Case B: stale handoff -> freshness FAIL (inverse proof) --------------

info "Case B: handoff predates HEAD commit"
REPO_B="${SCRATCH}/b"
make_repo "${REPO_B}" "2026-08-23T12:00:00"
make_handoff "${REPO_B}" "2026-08-01"
run_checker "${REPO_B}"
[ "${STATUS}" -eq 1 ] || fail "B: expected exit 1, got ${STATUS}: ${OUT}"
echo "${OUT}" | grep -q "FAIL freshness" || fail "B: missing FAIL freshness: ${OUT}"
pass "B: stale handoff fails the freshness lane"

# ---- Case C: remote behind HEAD -> remote-sync FAIL (inverse proof) -------

info "Case C: origin behind HEAD"
REPO_C="${SCRATCH}/c"
make_repo "${REPO_C}" "2026-08-23T12:00:00"
git init -q --bare "${SCRATCH}/c-origin.git"
git -C "${REPO_C}" remote add origin "${SCRATCH}/c-origin.git"
git -C "${REPO_C}" push -q origin main
echo y > "${REPO_C}/f"
git -C "${REPO_C}" add f
GIT_AUTHOR_DATE="2026-08-23T13:00:00" GIT_COMMITTER_DATE="2026-08-23T13:00:00" \
    git -C "${REPO_C}" commit -q -m "feat: unpushed y"
make_handoff "${REPO_C}" "2026-08-23"
run_checker "${REPO_C}"
[ "${STATUS}" -eq 1 ] || fail "C: expected exit 1, got ${STATUS}: ${OUT}"
echo "${OUT}" | grep -q "FAIL remote-sync" || fail "C: missing FAIL remote-sync: ${OUT}"
pass "C: remote behind HEAD fails the remote-sync lane"

# ---- Case D: remote in sync -> overall PASS --------------------------------

info "Case D: origin in sync with HEAD"
REPO_D="${SCRATCH}/d"
make_repo "${REPO_D}" "2026-08-23T12:00:00"
git init -q --bare "${SCRATCH}/d-origin.git"
git -C "${REPO_D}" remote add origin "${SCRATCH}/d-origin.git"
git -C "${REPO_D}" push -q origin main
make_handoff "${REPO_D}" "2026-08-23"
run_checker "${REPO_D}"
[ "${STATUS}" -eq 0 ] || fail "D: expected exit 0, got ${STATUS}: ${OUT}"
echo "${OUT}" | grep -q "PASS remote-sync" || fail "D: missing PASS remote-sync: ${OUT}"
pass "D: in-sync remote passes"

# ---- Case E: unreachable remote -> SKIP, overall PASS ----------------------

info "Case E: unreachable origin"
REPO_E="${SCRATCH}/e"
make_repo "${REPO_E}" "2026-08-23T12:00:00"
git -C "${REPO_E}" remote add origin "${SCRATCH}/does-not-exist.git"
make_handoff "${REPO_E}" "2026-08-23"
run_checker "${REPO_E}"
[ "${STATUS}" -eq 0 ] || fail "E: expected exit 0 (skip, not fail), got ${STATUS}: ${OUT}"
echo "${OUT}" | grep -q "SKIP remote-sync" || fail "E: missing SKIP remote-sync: ${OUT}"
pass "E: unreachable remote skips instead of failing"

# ---- Case F: stash + extra worktree -> reported, still PASS ----------------

info "Case F: stash and second worktree present"
REPO_F="${SCRATCH}/f"
make_repo "${REPO_F}" "2026-08-23T12:00:00"
echo dirty > "${REPO_F}/f"
git -C "${REPO_F}" stash -q
git -C "${REPO_F}" worktree add -q --detach "${SCRATCH}/f-wt" >/dev/null 2>&1
make_handoff "${REPO_F}" "2026-08-23"
run_checker "${REPO_F}"
[ "${STATUS}" -eq 0 ] || fail "F: expected exit 0, got ${STATUS}: ${OUT}"
echo "${OUT}" | grep -q "hygiene" || fail "F: missing hygiene lane output: ${OUT}"
echo "${OUT}" | grep -q "stash" || fail "F: stash not reported: ${OUT}"
pass "F: stash/worktree state is reported without failing"

echo
pass "ALL HANDOFF-FIDELITY TESTS PASSED"
