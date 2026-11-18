#!/usr/bin/env bash
# scripts/tests/test-openbao-root-token-argv.sh
#
# Prove that the OpenBao root token is delivered over kubectl exec stdin and
# never appears in kubectl argv. Uses a mock kubectl that records argv and
# stdin, and a runtime-generated fake token (not a real credential). The
# token value is never printed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=scripts/lib/openbao-exec-bao.sh
source "$ROOT/scripts/lib/openbao-exec-bao.sh"

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { printf 'PASS: %s\n' "$*"; }

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Runtime-generated fake; not committed, never printed.
FAKE_TOKEN="fake-openbao-argv-test-$(openssl rand -hex 16)"
RECORD="$TMP/records"
BIN="$TMP/bin"
mkdir -p "$RECORD" "$BIN"

cat > "$BIN/kubectl" << 'MOCK'
#!/usr/bin/env bash
# Mock kubectl: record argv and stdin, emit stub bao/kubectl output.
# Never writes recorded stdin to stdout/stderr.
set -euo pipefail
dir=${KUBECTL_RECORD_DIR:?}
n=1
if [ -f "$dir/counter" ]; then
    n=$(($(cat "$dir/counter") + 1))
fi
printf '%s\n' "$n" > "$dir/counter"

{
    printf '%s\n' "$#"
    for a in "$@"; do
        printf '%s\n' "${#a}"
        printf '%s\n' "$a"
    done
} > "$dir/call-$n.argv"

: > "$dir/call-$n.argv-joined"
for a in "$@"; do
    printf '%s\0' "$a" >> "$dir/call-$n.argv-joined"
done

cat > "$dir/call-$n.stdin"

remote=()
seen_dd=0
for a in "$@"; do
    if [ "$seen_dd" = 1 ]; then
        remote+=("$a")
    elif [ "$a" = "--" ]; then
        seen_dd=1
    fi
done

if [ "${remote[0]:-}" = "cat" ]; then
    printf 'SNAPDATA'
    exit 0
fi

joined=" ${remote[*]} "
case "$joined" in
    *" secrets list"*)
        printf '%s\n' "kv/" "secret/" "cubbyhole/" "identity/" "sys/"
        ;;
    *" kv get"*)
        printf '%s\n' "password=mock"
        ;;
esac
exit 0
MOCK
chmod +x "$BIN/kubectl"

export PATH="$BIN:$PATH"
export KUBECTL_RECORD_DIR="$RECORD"
export OPENBAO_NS=openbao
export OPENBAO_POD=openbao-0
export ROOT_TOKEN="$FAKE_TOKEN"
export OPENBAO_TOKEN="$FAKE_TOKEN"

reset_records() {
    rm -f "$RECORD"/call-* "$RECORD"/counter
}

# Compare files to the fake token without printing either value.
# $1 = file, $2 = exact|contains|absent, $3 = stdin|bytes
python_check() {
    local file="$1" mode="$2"
    FAKE_TOKEN="$FAKE_TOKEN" RECORD_FILE="$file" CHECK_MODE="$mode" python3 - << 'PY'
import os, sys
from pathlib import Path
path = Path(os.environ["RECORD_FILE"])
mode = os.environ["CHECK_MODE"]
token = os.environ["FAKE_TOKEN"].encode()
data = path.read_bytes()
if mode == "exact":
    if data != token:
        sys.exit("stdin did not match the fake token exactly (got %d bytes, expected %d)" % (len(data), len(token)))
elif mode == "firstline":
    nl = data.find(b"\n")
    if nl < 0:
        sys.exit("stdin has no newline; expected first-line token framing")
    if data[:nl] != token:
        sys.exit("first stdin line did not match the fake token")
    leftover = data[nl + 1 :]
    Path(os.environ.get("PAYLOAD_OUT", "/dev/null")).write_bytes(leftover)
elif mode == "absent":
    if token in data:
        sys.exit("fake token found in %s" % path.name)
elif mode == "contains":
    if token not in data:
        sys.exit("fake token missing from %s" % path.name)
else:
    sys.exit("unknown check mode")
PY
}

assert_no_token_in_argv() {
    local f
    shopt -s nullglob
    for f in "$RECORD"/call-*.argv "$RECORD"/call-*.argv-joined; do
        python_check "$f" absent
    done
    shopt -u nullglob
}

assert_has_exec_dash_i() {
    local f found=0
    shopt -s nullglob
    for f in "$RECORD"/call-*.argv; do
        if awk 'BEGIN{n=""}
            NR==1 {argc=$0; next}
            {len=$0; getline val;
             if (val=="-i") found=1}
            END{exit found?0:1}' "$f"; then
            found=1
        fi
    done
    shopt -u nullglob
    [ "$found" = 1 ] || fail "kubectl exec was invoked without -i"
}

# --- 1. openbao_exec_bao: token only on stdin, never argv ---
reset_records
openbao_exec_bao token lookup >/dev/null
[ -f "$RECORD/call-1.stdin" ] || fail "mock kubectl was not invoked"
python_check "$RECORD/call-1.stdin" exact
assert_no_token_in_argv
assert_has_exec_dash_i
# The pod command must use BAO_TOKEN=$(cat) and must not mention a literal token.
if ! awk 'BEGIN{ok=0}
    NR==1 {next}
    {len=$0; getline val;
     if (val ~ /BAO_TOKEN=\$\(cat\)/) ok=1}
    END{exit ok?0:1}' "$RECORD/call-1.argv"; then
    fail "pod command does not read the token with BAO_TOKEN=\$(cat)"
fi
pass "openbao_exec_bao delivers the token only on stdin"

# --- 2. openbao_exec_bao_stdin: first line token, rest payload ---
reset_records
payload=$'policy-body-line-1\npath "secret/data/*" { capabilities = ["read"] }\n'
PAYLOAD_OUT="$TMP/payload.out"
export PAYLOAD_OUT
printf '%s' "$payload" | openbao_exec_bao_stdin policy write example-policy - >/dev/null
python_check "$RECORD/call-1.stdin" firstline
printf '%s' "$payload" > "$TMP/expected-payload"
EXPECT_PAYLOAD="$TMP/expected-payload" PAYLOAD_OUT="$TMP/payload.out" python3 -c \
    'import os,sys; p=open(os.environ["PAYLOAD_OUT"],"rb").read(); e=open(os.environ["EXPECT_PAYLOAD"],"rb").read(); sys.exit(0 if p==e else "payload mismatch (%d vs %d bytes)"%(len(p),len(e)))'
assert_no_token_in_argv
assert_has_exec_dash_i
pass "openbao_exec_bao_stdin frames token as first line; payload follows"

# --- 3. openbao_exec_pod_sh: token on stdin; script (not token) in argv ---
reset_records
# shellcheck disable=SC2016
openbao_exec_pod_sh 'bao write auth/kubernetes/config kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"' >/dev/null
python_check "$RECORD/call-1.stdin" exact
assert_no_token_in_argv
assert_has_exec_dash_i
if ! awk 'BEGIN{ok=0}
    NR==1 {next}
    {len=$0; getline val;
     if (index(val, "KUBERNETES_PORT_443_TCP_ADDR")>0) ok=1}
    END{exit ok?0:1}' "$RECORD/call-1.argv"; then
    fail "pod script with KUBERNETES_PORT_443_TCP_ADDR was not passed through"
fi
pass "openbao_exec_pod_sh delivers the token only on stdin"

# --- 4. bootstrap helpers (sourced; main must not run) ---
reset_records
# shellcheck source=scripts/bootstrap-openbao-persistent.sh
source "$ROOT/scripts/bootstrap-openbao-persistent.sh"
# If main ran, it would have called wait_for_api and more kubectl; we reset
# after source in case of accidental calls -- main is guarded.
reset_records
bao_auth token lookup >/dev/null
python_check "$RECORD/call-1.stdin" exact
assert_no_token_in_argv
# put_if_absent: mock kv get succeeds, so no put should run
printf '%s' "fake-dev-value" | put_if_absent secret/example >/dev/null
# only the kv get from bao_auth should have been recorded (call-2)
if [ -f "$RECORD/call-3.argv" ]; then
    fail "put_if_absent wrote despite kv get succeeding"
fi
assert_no_token_in_argv
# Force a put by making kv get fail once. Easier: point at a path and make
# mock return fail for get of a sentinel. Instead, temporarily wrap kubectl
# via a flag.
pass "sourced bootstrap bao_auth keeps the token off argv"

# Force put_if_absent write path: make kv get fail for this path only.
reset_records
cat > "$BIN/kubectl" << 'MOCK'
#!/usr/bin/env bash
set -euo pipefail
dir=${KUBECTL_RECORD_DIR:?}
n=1
if [ -f "$dir/counter" ]; then
    n=$(($(cat "$dir/counter") + 1))
fi
printf '%s\n' "$n" > "$dir/counter"
{
    printf '%s\n' "$#"
    for a in "$@"; do
        printf '%s\n' "${#a}"
        printf '%s\n' "$a"
    done
} > "$dir/call-$n.argv"
: > "$dir/call-$n.argv-joined"
for a in "$@"; do
    printf '%s\0' "$a" >> "$dir/call-$n.argv-joined"
done
cat > "$dir/call-$n.stdin"
joined=" $* "
if [[ "$joined" == *" kv get "* ]]; then
    exit 1
fi
exit 0
MOCK
chmod +x "$BIN/kubectl"
printf '%s' "fake-dev-value-for-put" | put_if_absent secret/example >/dev/null
# Expect two calls: kv get (fail) then kv put (stdin framed)
[ -f "$RECORD/call-2.stdin" ] || fail "put_if_absent did not invoke kubectl for kv put"
python_check "$RECORD/call-2.stdin" firstline
PAYLOAD_OUT="$TMP/payload.out" python_check "$RECORD/call-2.stdin" firstline
printf '%s' "fake-dev-value-for-put" > "$TMP/expected-payload"
[ "$(wc -c < "$TMP/payload.out" | tr -d ' ')" = "$(wc -c < "$TMP/expected-payload" | tr -d ' ')" ] \
    || fail "put_if_absent payload length mismatch"
cmp -s "$TMP/payload.out" "$TMP/expected-payload" || fail "put_if_absent payload mismatch"
assert_no_token_in_argv
pass "put_if_absent frames token + value on stdin, never argv"

# Restore the original mock for script-level tests.
cat > "$BIN/kubectl" << 'MOCK'
#!/usr/bin/env bash
set -euo pipefail
dir=${KUBECTL_RECORD_DIR:?}
n=1
if [ -f "$dir/counter" ]; then
    n=$(($(cat "$dir/counter") + 1))
fi
printf '%s\n' "$n" > "$dir/counter"
{
    printf '%s\n' "$#"
    for a in "$@"; do
        printf '%s\n' "${#a}"
        printf '%s\n' "$a"
    done
} > "$dir/call-$n.argv"
: > "$dir/call-$n.argv-joined"
for a in "$@"; do
    printf '%s\0' "$a" >> "$dir/call-$n.argv-joined"
done
cat > "$dir/call-$n.stdin"
remote=()
seen_dd=0
for a in "$@"; do
    if [ "$seen_dd" = 1 ]; then
        remote+=("$a")
    elif [ "$a" = "--" ]; then
        seen_dd=1
    fi
done
if [ "${remote[0]:-}" = "cat" ]; then
    printf 'SNAPDATA'
    exit 0
fi
joined=" ${remote[*]} "
case "$joined" in
    *" secrets list"*)
        printf '%s\n' "kv/" "secret/" "cubbyhole/" "identity/" "sys/"
        ;;
    *" kv get"*)
        printf '%s\n' "password=mock"
        ;;
esac
exit 0
MOCK
chmod +x "$BIN/kubectl"

# --- 5. seed-openbao-m2-paths.sh (kv already present; no writes) ---
reset_records
unset ROOT_TOKEN
OPENBAO_TOKEN="$FAKE_TOKEN" SEED_OPENBAO_SKIP_RUNNER_TOKEN=1 \
    "$ROOT/scripts/seed-openbao-m2-paths.sh" >/dev/null
assert_no_token_in_argv
# at least one call should have carried the token on stdin
found_stdin=0
shopt -s nullglob
for f in "$RECORD"/call-*.stdin; do
    if FAKE_TOKEN="$FAKE_TOKEN" RECORD_FILE="$f" CHECK_MODE=exact python3 - << 'PY'
import os, sys
from pathlib import Path
token = os.environ["FAKE_TOKEN"].encode()
data = Path(os.environ["RECORD_FILE"]).read_bytes()
sys.exit(0 if data == token else 1)
PY
    then
        found_stdin=1
    fi
done
shopt -u nullglob
[ "$found_stdin" = 1 ] || fail "seed script did not send the token on exec stdin"
pass "seed-openbao-m2-paths.sh keeps the token off argv"

# --- 6. smoke-openbao.sh presence lane ---
reset_records
# Re-export token; smoke reads OPENBAO_TOKEN
OPENBAO_TOKEN="$FAKE_TOKEN" "$ROOT/scripts/smoke-openbao.sh" >/dev/null
assert_no_token_in_argv
found_stdin=0
shopt -s nullglob
for f in "$RECORD"/call-*.stdin; do
    if FAKE_TOKEN="$FAKE_TOKEN" RECORD_FILE="$f" CHECK_MODE=exact python3 - << 'PY'
import os, sys
from pathlib import Path
token = os.environ["FAKE_TOKEN"].encode()
data = Path(os.environ["RECORD_FILE"]).read_bytes()
sys.exit(0 if data == token else 1)
PY
    then
        found_stdin=1
    fi
done
shopt -u nullglob
[ "$found_stdin" = 1 ] || fail "smoke script did not send the token on exec stdin"
pass "smoke-openbao.sh keeps the token off argv"

# --- 7. backup-openbao.sh ---
reset_records
CUSTODY="$TMP/custody"
BACKUP="$TMP/backups"
mkdir -p "$CUSTODY" "$BACKUP"
umask 077
printf '%s\n' "$FAKE_TOKEN" > "$CUSTODY/root-token"
chmod 600 "$CUSTODY/root-token"
# backup reads the token from the custody file into ROOT_TOKEN
OPENBAO_CUSTODY_DIR="$CUSTODY" OPENBAO_BACKUP_DIR="$BACKUP" \
    "$ROOT/scripts/backup-openbao.sh" >/dev/null
assert_no_token_in_argv
# The snapshot-save exec should have the token on stdin; cat/rm have empty stdin
found_stdin=0
shopt -s nullglob
for f in "$RECORD"/call-*.stdin; do
    if FAKE_TOKEN="$FAKE_TOKEN" RECORD_FILE="$f" CHECK_MODE=exact python3 - << 'PY'
import os, sys
from pathlib import Path
token = os.environ["FAKE_TOKEN"].encode()
data = Path(os.environ["RECORD_FILE"]).read_bytes()
sys.exit(0 if data == token else 1)
PY
    then
        found_stdin=1
    fi
done
shopt -u nullglob
[ "$found_stdin" = 1 ] || fail "backup script did not send the token on exec stdin"
pass "backup-openbao.sh keeps the token off argv"

# --- 8. static: no env VAULT_TOKEN= left on kubectl exec in these scripts ---
if grep -n 'env VAULT_TOKEN=' \
    "$ROOT/scripts/bootstrap-openbao-persistent.sh" \
    "$ROOT/scripts/seed-openbao-m2-paths.sh" \
    "$ROOT/scripts/smoke-openbao.sh" \
    "$ROOT/scripts/backup-openbao.sh" \
    "$ROOT/scripts/lib/openbao-exec-bao.sh"; then
    fail "scripts still pass VAULT_TOKEN on a command line"
fi
pass "no env VAULT_TOKEN= remains in the updated scripts"

printf 'All tests passed\n'
