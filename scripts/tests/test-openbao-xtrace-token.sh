#!/usr/bin/env bash
# scripts/tests/test-openbao-xtrace-token.sh
#
# Prove that set -x / BASH_XTRACEFD cannot leak the OpenBao token through
# _openbao_exec_prepare or the public helpers, and that the caller's xtrace
# flag is restored on every exit path. Uses a runtime-generated fake token
# (not a real credential). The token value is never printed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=scripts/lib/openbao-exec-bao.sh
source "$ROOT/scripts/lib/openbao-exec-bao.sh"

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { printf 'PASS: %s\n' "$*"; }

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Runtime-generated fake; not committed, never printed.
FAKE_TOKEN="fake-openbao-xtrace-test-$(openssl rand -hex 16)"
RECORD="$TMP/records"
BIN="$TMP/bin"
mkdir -p "$RECORD" "$BIN"

cat > "$BIN/kubectl" << 'MOCK'
#!/usr/bin/env bash
set -euo pipefail
dir=${KUBECTL_RECORD_DIR:?}
n=1
if [ -f "$dir/counter" ]; then
    n=$(($(cat "$dir/counter") + 1))
fi
printf '%s\n' "$n" > "$dir/counter"
# Swallow stdin; never echo it.
cat >/dev/null
exit 0
MOCK
chmod +x "$BIN/kubectl"

export PATH="$BIN:$PATH"
export KUBECTL_RECORD_DIR="$RECORD"
export OPENBAO_NS=openbao
export OPENBAO_POD=openbao-0
export ROOT_TOKEN="$FAKE_TOKEN"
export OPENBAO_TOKEN="$FAKE_TOKEN"

# Compare a file to the fake token without printing either value.
# $1 = file, $2 = absent
python_check_absent() {
    local file="$1"
    [ -f "$file" ] || fail "missing capture file"
    FAKE_TOKEN="$FAKE_TOKEN" RECORD_FILE="$file" python3 - << 'PY'
import os, sys
from pathlib import Path
path = Path(os.environ["RECORD_FILE"])
token = os.environ["FAKE_TOKEN"].encode()
if token in path.read_bytes():
    sys.exit("fake token found in %s" % path.name)
PY
}

# Run a command with set -x, capturing stderr and optionally BASH_XTRACEFD.
# Leaves xtrace off and restores the caller's set -e. Sets _XTRACE_FLAGS_AFTER
# and _RUN_RC. Never prints the token.
# usage: run_with_xtrace <stderr-out> <xtracefd-out or ''> <cmd> [args...]
run_with_xtrace() {
    local stderr_out="$1" fd_out="$2"
    shift 2
    local xtfd="" flags_after="" rc=0 had_e=0
    case "$-" in
        *e*) had_e=1 ;;
    esac
    set +e
    if [ -n "$fd_out" ]; then
        : > "$fd_out"
        exec {xtfd}>"$fd_out"
        BASH_XTRACEFD=$xtfd
        export BASH_XTRACEFD
    else
        unset BASH_XTRACEFD || true
    fi
    {
        set -x
        "$@" >/dev/null
        rc=$?
        flags_after=$-
        set +x
    } 2>"$stderr_out"
    if [ -n "$fd_out" ]; then
        exec {xtfd}>&-
        unset BASH_XTRACEFD || true
    fi
    [ "$had_e" -eq 1 ] && set -e
    _XTRACE_FLAGS_AFTER=$flags_after
    _RUN_RC=$rc
}

assert_xtrace_restored() {
    case "${_XTRACE_FLAGS_AFTER:-}" in
        *x*) ;;
        *) fail "xtrace was not restored after the call (flags='${_XTRACE_FLAGS_AFTER:-}')" ;;
    esac
}

assert_xtrace_still_off() {
    case "$-" in
        *x*) fail "xtrace is on but the caller had it off (flags=$-)" ;;
    esac
}

# --- 1. _openbao_exec_prepare under set -x: no token on stderr; xtrace restored ---
run_with_xtrace "$TMP/prep.stderr" "" _openbao_exec_prepare
[ "$_RUN_RC" = 0 ] || fail "prepare should succeed with ns/pod/token set"
assert_xtrace_restored
python_check_absent "$TMP/prep.stderr"
pass "_openbao_exec_prepare does not leak the token under set -x and restores xtrace"

# --- 2. Same, via BASH_XTRACEFD ---
run_with_xtrace "$TMP/prep-fd.stderr" "$TMP/prep-fd.xtrace" _openbao_exec_prepare
[ "$_RUN_RC" = 0 ] || fail "prepare should succeed under BASH_XTRACEFD"
assert_xtrace_restored
python_check_absent "$TMP/prep-fd.stderr"
python_check_absent "$TMP/prep-fd.xtrace"
pass "_openbao_exec_prepare does not leak the token on BASH_XTRACEFD"

# --- 3. prepare missing ns (token still set): no leak; xtrace restored ---
unset OPENBAO_NS
run_with_xtrace "$TMP/prep-ns.stderr" "$TMP/prep-ns.xtrace" _openbao_exec_prepare
export OPENBAO_NS=openbao
[ "$_RUN_RC" = 1 ] || fail "prepare should fail when OPENBAO_NS is unset"
assert_xtrace_restored
python_check_absent "$TMP/prep-ns.stderr"
python_check_absent "$TMP/prep-ns.xtrace"
pass "_openbao_exec_prepare failure (missing ns) restores xtrace and does not leak"

# --- 4. prepare missing token: xtrace restored ---
unset ROOT_TOKEN OPENBAO_TOKEN
run_with_xtrace "$TMP/prep-tok.stderr" "" _openbao_exec_prepare
export ROOT_TOKEN="$FAKE_TOKEN" OPENBAO_TOKEN="$FAKE_TOKEN"
[ "$_RUN_RC" = 1 ] || fail "prepare should fail when no token is set"
assert_xtrace_restored
pass "_openbao_exec_prepare failure (missing token) restores xtrace"

# --- 5. openbao_exec_bao under set -x (stderr and BASH_XTRACEFD) ---
run_with_xtrace "$TMP/bao.stderr" "$TMP/bao.xtrace" openbao_exec_bao token lookup
[ "$_RUN_RC" = 0 ] || fail "openbao_exec_bao should succeed against the mock"
assert_xtrace_restored
python_check_absent "$TMP/bao.stderr"
python_check_absent "$TMP/bao.xtrace"
pass "openbao_exec_bao does not leak the token under set -x and restores xtrace"

# --- 6. openbao_exec_bao_stdin under set -x ---
run_with_xtrace "$TMP/stdin.stderr" "$TMP/stdin.xtrace" \
    openbao_exec_bao_stdin policy write example-policy - <<<"payload-line"
[ "$_RUN_RC" = 0 ] || fail "openbao_exec_bao_stdin should succeed against the mock"
assert_xtrace_restored
python_check_absent "$TMP/stdin.stderr"
python_check_absent "$TMP/stdin.xtrace"
pass "openbao_exec_bao_stdin does not leak the token under set -x and restores xtrace"

# --- 7. openbao_exec_pod_sh under set -x ---
# shellcheck disable=SC2016
run_with_xtrace "$TMP/podsh.stderr" "$TMP/podsh.xtrace" \
    openbao_exec_pod_sh 'echo "$KUBERNETES_PORT_443_TCP_ADDR"'
[ "$_RUN_RC" = 0 ] || fail "openbao_exec_pod_sh should succeed against the mock"
assert_xtrace_restored
python_check_absent "$TMP/podsh.stderr"
python_check_absent "$TMP/podsh.xtrace"
pass "openbao_exec_pod_sh does not leak the token under set -x and restores xtrace"

# --- 8. openbao_exec_pod_sh early return (empty script): restore xtrace, no leak ---
run_with_xtrace "$TMP/podsh-empty.stderr" "$TMP/podsh-empty.xtrace" \
    openbao_exec_pod_sh ""
[ "$_RUN_RC" = 1 ] || fail "openbao_exec_pod_sh should fail with an empty script"
assert_xtrace_restored
python_check_absent "$TMP/podsh-empty.stderr"
python_check_absent "$TMP/podsh-empty.xtrace"
pass "openbao_exec_pod_sh empty-script path restores xtrace and does not leak"

# --- 9. OPENBAO_TOKEN only (ROOT_TOKEN unset) under set -x ---
unset ROOT_TOKEN
run_with_xtrace "$TMP/otoken.stderr" "$TMP/otoken.xtrace" openbao_exec_bao token lookup
export ROOT_TOKEN="$FAKE_TOKEN"
[ "$_RUN_RC" = 0 ] || fail "openbao_exec_bao should succeed with OPENBAO_TOKEN only"
assert_xtrace_restored
python_check_absent "$TMP/otoken.stderr"
python_check_absent "$TMP/otoken.xtrace"
pass "OPENBAO_TOKEN-only path does not leak under set -x and restores xtrace"

# --- 10. xtrace stays off when the caller had it off ---
case "$-" in
    *x*) fail "precondition failed: xtrace should be off before stay-off tests" ;;
esac
_openbao_exec_prepare || fail "prepare should succeed in stay-off test"
assert_xtrace_still_off
openbao_exec_bao token lookup >/dev/null
assert_xtrace_still_off
printf 'payload\n' | openbao_exec_bao_stdin policy write example-policy - >/dev/null
assert_xtrace_still_off
# shellcheck disable=SC2016
openbao_exec_pod_sh 'true' >/dev/null
assert_xtrace_still_off
# failure paths must also leave it off
unset OPENBAO_NS
_openbao_exec_prepare >/dev/null 2>&1 || true
export OPENBAO_NS=openbao
assert_xtrace_still_off
openbao_exec_pod_sh "" >/dev/null 2>&1 || true
assert_xtrace_still_off
pass "xtrace stays off after each helper when the caller had it off"

printf 'All tests passed\n'
