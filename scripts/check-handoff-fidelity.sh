#!/usr/bin/env bash
# check-handoff-fidelity.sh -- handoff freshness and state-drift gate.
#
# Enforces the standing snapshot directive in SESSION_HANDOFF.md: after any
# major state change a dated snapshot addendum must land before the session
# ends. Three lanes:
#
#   freshness    FAIL if SESSION_HANDOFF.md's "**Last updated:**" date is
#                older than HEAD's committer date (a commit landed without
#                a snapshot), or the file/marker is missing.
#   remote-sync  FAIL if any configured remote's refs/heads/main points at
#                a commit other than HEAD (publication drift). SKIP per
#                remote when it is unreachable -- unreachable is UNVERIFIED,
#                never a failure.
#   hygiene      informational only: reports extra git worktrees and stash
#                entries so cross-tool cleanup (e.g. a sanitization pass by
#                another harness) is visible. Never fails.
#
# Usage:  scripts/check-handoff-fidelity.sh [--offline]
# Testability overrides (used by scripts/tests/test-check-handoff-fidelity.sh):
#         HANDOFF_FIDELITY_ROOT   repo root to inspect (default: toplevel)
#         HANDOFF_FIDELITY_FILE   handoff file (default: $ROOT/SESSION_HANDOFF.md)
#         --offline skips the remote-sync lane entirely.

set -euo pipefail

OFFLINE=0
case "${1:-}" in
  --offline) OFFLINE=1 ;;
  "") ;;
  *) echo "usage: check-handoff-fidelity.sh [--offline]" >&2; exit 1 ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/colors.sh
. "${SCRIPT_DIR}/lib/colors.sh"

pass() { echo -e "${COLOR_GREEN} PASS${COLOR_RESET} $*"; }
fail_lane() { echo -e "${COLOR_RED} FAIL${COLOR_RESET} $*"; }
skip() { echo -e "${COLOR_YELLOW} SKIP${COLOR_RESET} $*"; }
info() { echo -e "${COLOR_BOLD}INFO${COLOR_RESET} $*"; }

ROOT="${HANDOFF_FIDELITY_ROOT:-$(git rev-parse --show-toplevel)}"
HANDOFF="${HANDOFF_FIDELITY_FILE:-${ROOT}/SESSION_HANDOFF.md}"
FAILURES=0

# ---- Lane 1: freshness -----------------------------------------------------

if [ ! -f "${HANDOFF}" ]; then
  fail_lane "freshness: ${HANDOFF} does not exist"
  FAILURES=$((FAILURES + 1))
else
  UPDATED="$(grep -m1 '^\*\*Last updated:\*\*' "${HANDOFF}" | sed -e 's/^\*\*Last updated:\*\* *//' -e 's/ .*//' || true)"
  HEAD_DATE="$(git -C "${ROOT}" log -1 --format=%cd --date=format:%Y-%m-%d)"
  if [ -z "${UPDATED}" ]; then
    fail_lane "freshness: no '**Last updated:**' marker in ${HANDOFF}"
    FAILURES=$((FAILURES + 1))
  elif [[ "${UPDATED}" < "${HEAD_DATE}" ]]; then
    fail_lane "freshness: handoff Last updated ${UPDATED} predates HEAD commit date ${HEAD_DATE} -- snapshot required"
    FAILURES=$((FAILURES + 1))
  else
    pass "freshness: handoff ${UPDATED} covers HEAD commit date ${HEAD_DATE}"
  fi
fi

# ---- Lane 2: remote-sync ----------------------------------------------------

HEAD_SHA="$(git -C "${ROOT}" rev-parse HEAD)"
REMOTES="$(git -C "${ROOT}" remote || true)"

if [ "${OFFLINE}" -eq 1 ]; then
  skip "remote-sync: --offline"
elif [ -z "${REMOTES}" ]; then
  skip "remote-sync: no remotes configured"
else
  for REMOTE in ${REMOTES}; do
    REMOTE_SHA="$(GIT_TERMINAL_PROMPT=0 git -C "${ROOT}" ls-remote "${REMOTE}" refs/heads/main 2>/dev/null | awk '{print $1}' || true)"
    if [ -z "${REMOTE_SHA}" ]; then
      skip "remote-sync: ${REMOTE} unreachable (UNVERIFIED, not a failure)"
    elif [ "${REMOTE_SHA}" != "${HEAD_SHA}" ]; then
      fail_lane "remote-sync: ${REMOTE} main is ${REMOTE_SHA}, HEAD is ${HEAD_SHA} -- publication drift"
      FAILURES=$((FAILURES + 1))
    else
      pass "remote-sync: ${REMOTE} main matches HEAD (${HEAD_SHA})"
    fi
  done
fi

# ---- Lane 3: hygiene (informational) ----------------------------------------

WT_EXTRA=$(( $(git -C "${ROOT}" worktree list | wc -l | tr -d ' ') - 1 ))
STASH_COUNT=$(git -C "${ROOT}" stash list | wc -l | tr -d ' ')
info "hygiene: ${WT_EXTRA} extra worktree(s), ${STASH_COUNT} stash entrie(s)"
pass "hygiene: informational only"

echo
if [ "${FAILURES}" -gt 0 ]; then
  fail_lane "handoff fidelity: ${FAILURES} lane(s) failed"
  exit 1
fi
pass "handoff fidelity: all lanes green"
