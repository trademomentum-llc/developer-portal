#!/usr/bin/env bash
# scripts/tests/test-openbao-root-token-argv.sh
#
# Prove that OpenBao secret values never appear in kubectl argv. Covers
# kubectl exec stdin handoff, handoff_root_token, and sync_unseal_secret
# --from-file temp files (mode 0600, removed on success and failure). Uses a
# mock kubectl that records argv and stdin, and runtime-generated fakes (not
# real credentials). Secret values are never printed.
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
source "${OPENBAO_BOOTSTRAP_SCRIPT:-$ROOT/scripts/bootstrap-openbao-persistent.sh}"
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

if grep -n -- '--from-literal=' "$ROOT/scripts/bootstrap-openbao-persistent.sh"; then
    fail "bootstrap still uses --from-literal= for a secret"
fi
pass "no --from-literal= remains in bootstrap-openbao-persistent.sh"

# --- 9. handoff_root_token: --from-file temp, mode 600, deleted, content ok ---
install_handoff_mock() {
    cat > "$BIN/kubectl" << 'MOCK'
#!/usr/bin/env bash
# Record argv/stdin. For `create secret --from-file=token=PATH`, record the
# path, mode, and file content (never printed). Emit a dry-run Secret yaml
# on stdout for the apply pipe. For `apply`, honor KUBECTL_APPLY_FAIL.
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

from_file=""
ns=""
name=""
prev=""
for a in "$@"; do
    case "$a" in
        --from-file=token=*)
            from_file="${a#--from-file=token=}"
            ;;
        --from-literal=*)
            printf 'from-literal\n' >> "$dir/handoff-from-literal"
            ;;
        -n|--namespace)
            prev=ns
            ;;
        --namespace=*)
            ns="${a#--namespace=}"
            ;;
        *)
            if [ "$prev" = "ns" ]; then
                ns="$a"
                prev=""
            elif [ "$prev" = "generic" ]; then
                name="$a"
                prev=""
            fi
            if [ "$a" = "generic" ]; then
                prev=generic
            fi
            ;;
    esac
done

is_create=0
is_apply=0
for a in "$@"; do
    [ "$a" = "create" ] && is_create=1
    [ "$a" = "apply" ] && is_apply=1
done

if [ "$is_create" = 1 ]; then
    if [ -n "$from_file" ]; then
        printf '%s\n' "$from_file" > "$dir/handoff-token-path"
        if [ -e "$from_file" ]; then
            stat -c '%a' "$from_file" > "$dir/handoff-token-mode"
            cp -a "$from_file" "$dir/handoff-token-content"
        else
            printf 'missing\n' > "$dir/handoff-token-mode"
        fi
    fi
    if [ "${KUBECTL_CREATE_FAIL:-0}" = "1" ]; then
        exit 1
    fi
    [ -n "$from_file" ] && [ -e "$from_file" ] || exit 1
    b64=$(base64 -w0 < "$from_file")
    printf '%s\n' "apiVersion: v1" "kind: Secret"
    printf '%s\n' "metadata:"
    printf '  name: %s\n' "${name:-openbao-root-token}"
    printf '  namespace: %s\n' "${ns:-external-secrets}"
    printf '%s\n' "type: Opaque" "data:"
    printf '  token: %s\n' "$b64"
    exit 0
fi

if [ "$is_apply" = 1 ]; then
    cp "$dir/call-$n.stdin" "$dir/handoff-apply-yaml"
    if [ "${KUBECTL_APPLY_FAIL:-0}" = "1" ]; then
        exit 1
    fi
    exit 0
fi

exit 0
MOCK
    chmod +x "$BIN/kubectl"
}

assert_handoff_file_gone() {
    local path
    [ -s "$RECORD/handoff-token-path" ] || fail "kubectl create did not record a --from-file path"
    path=$(cat "$RECORD/handoff-token-path")
    if [ -e "$path" ]; then
        fail "temp token file was not removed"
    fi
}

assert_handoff_mode_600() {
    [ -s "$RECORD/handoff-token-mode" ] || fail "kubectl create did not record the temp file mode"
    [ "$(cat "$RECORD/handoff-token-mode")" = "600" ] || fail "temp token file was not mode 600 when kubectl ran"
}

assert_handoff_content() {
    [ -f "$RECORD/handoff-token-content" ] || fail "kubectl create did not record the temp file content"
    python_check "$RECORD/handoff-token-content" exact
}

assert_apply_secret_content() {
    [ -f "$RECORD/handoff-apply-yaml" ] || fail "kubectl apply did not receive yaml stdin"
    FAKE_TOKEN="$FAKE_TOKEN" APPLY_YAML="$RECORD/handoff-apply-yaml" python3 - << 'PYCHECK'
import os, sys, base64, re
from pathlib import Path
token = os.environ["FAKE_TOKEN"].encode()
raw = Path(os.environ["APPLY_YAML"]).read_bytes()
text = raw.decode()
m = re.search(r"(?m)^\s*token:\s*([A-Za-z0-9+/=]+)\s*$", text)
if not m:
    sys.exit("apply yaml has no data.token")
try:
    decoded = base64.b64decode(m.group(1), validate=True)
except Exception:
    sys.exit("apply yaml token was not valid base64")
if decoded != token:
    sys.exit("apply yaml token does not match the fake token (%d vs %d bytes)" % (len(decoded), len(token)))
if token in raw:
    sys.exit("apply yaml contains the raw token")
PYCHECK
}

# Restore helpers (bootstrap's fail/info overwrite the test versions when sourced).
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { printf 'PASS: %s\n' "$*"; }

export ROOT_TOKEN="$FAKE_TOKEN"
export ESO_NS="${ESO_NS:-external-secrets}"
export ESO_SECRET="${ESO_SECRET:-openbao-root-token}"

# 9a. success: mode 600, content correct, file deleted, token off argv
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
handoff_root_token >/dev/null
assert_no_token_in_argv
if [ -f "$RECORD/handoff-from-literal" ]; then
    fail "kubectl was invoked with --from-literal"
fi
assert_handoff_mode_600
assert_handoff_content
assert_apply_secret_content
assert_handoff_file_gone
# create must have used --from-file=token= (and -o yaml / --dry-run=client)
if ! grep -F -q -- '--from-file=token=' "$RECORD"/call-*.argv; then
    fail "kubectl create was not invoked with --from-file=token="
fi
if ! grep -F -q -- '--dry-run=client' "$RECORD"/call-*.argv; then
    fail "kubectl create dropped --dry-run=client"
fi
pass "handoff_root_token success: mode 600, secret content correct, file deleted, token off argv"

# 9b. apply failure: temp file still deleted
# Run in a subshell: bootstrap fail() exits, which must not abort this test script.
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
if (KUBECTL_APPLY_FAIL=1 handoff_root_token >/dev/null 2>"$TMP/handoff-apply-fail.err"); then
    fail "handoff_root_token should have failed when kubectl apply failed"
fi
assert_no_token_in_argv
assert_handoff_mode_600
assert_handoff_content
assert_handoff_file_gone
python_check "$TMP/handoff-apply-fail.err" absent
pass "handoff_root_token apply failure: temp file deleted, token off argv and stderr"

# 9c. create failure: temp file still deleted
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
if (KUBECTL_CREATE_FAIL=1 handoff_root_token >/dev/null 2>"$TMP/handoff-create-fail.err"); then
    fail "handoff_root_token should have failed when kubectl create failed"
fi
assert_no_token_in_argv
assert_handoff_mode_600
assert_handoff_content
assert_handoff_file_gone
python_check "$TMP/handoff-create-fail.err" absent
pass "handoff_root_token create failure: temp file deleted, token off argv and stderr"

# 9d. xtrace does not leak the token
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
# Run in a subshell with xtrace; handoff_root_token must disable it before writing the token.
set +e
(
    set -x
    handoff_root_token
) >"$TMP/handoff-xtrace.out" 2>"$TMP/handoff-xtrace.err"
xrc=$?
set -e
[ "$xrc" -eq 0 ] || fail "handoff_root_token failed under xtrace"
python_check "$TMP/handoff-xtrace.out" absent
python_check "$TMP/handoff-xtrace.err" absent
assert_no_token_in_argv
assert_handoff_file_gone
pass "handoff_root_token does not leak the token under xtrace"

# --- 9e-g. write/chmod/empty-file faults must not reach kubectl ---
# Faults are injected with PATH shims (and enable -n for the printf builtin).
# Injection is mechanism-based and UID-independent; do not use file modes.
# Record the mktemp path so we can assert the trap deleted the file even
# when kubectl (which is what normally records --from-file=) is never called.
install_recording_mktemp() {
    cat > "$BIN/mktemp" << 'MOCK'
#!/usr/bin/env bash
set -euo pipefail
dir=${KUBECTL_RECORD_DIR:?}
real=$(PATH=/usr/bin:/bin:/usr/sbin:/sbin type -P mktemp)
f=$("$real" "$@")
printf '%s\n' "$f" > "$dir/handoff-token-path"
printf '%s\n' "$f"
MOCK
    /usr/bin/chmod +x "$BIN/mktemp"
}

# $1 = exit status. Stdout is always empty; never prints a secret value.
# Callers must `enable -n printf` so the builtin is not used.
install_printf_stub() {
    local rc="$1"
    cat > "$BIN/printf" << MOCK
#!/usr/bin/env bash
exit ${rc}
MOCK
    /usr/bin/chmod +x "$BIN/printf"
}

remove_command_stubs() {
    rm -f "$BIN/mktemp" "$BIN/chmod" "$BIN/printf"
}

assert_kubectl_not_invoked() {
    local matches
    shopt -s nullglob
    matches=("$RECORD"/call-*)
    shopt -u nullglob
    if [ -f "$RECORD/counter" ] || [ "${#matches[@]}" -gt 0 ]; then
        fail "kubectl was invoked after a temp-file fault"
    fi
}

assert_fault_cleanup() {
    local err="$1"
    assert_kubectl_not_invoked
    assert_no_token_in_argv
    python_check "$err" absent
    assert_handoff_file_gone
}

# 9e. write failure (PATH printf stub exits 1): no kubectl, non-zero, file gone
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
install_recording_mktemp
install_printf_stub 1
hash -r
if (
    enable -n printf
    hash -r
    handoff_root_token >/dev/null 2>"$TMP/handoff-write-fail.err"
); then
    remove_command_stubs
    fail "handoff_root_token should have failed when writing the temp token file failed"
fi
remove_command_stubs
assert_fault_cleanup "$TMP/handoff-write-fail.err"
pass "handoff_root_token write failure: kubectl not invoked, non-zero, temp file deleted, token off argv and stderr"

# 9f. chmod failure (stubbed chmod): no kubectl, non-zero, file gone
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
install_recording_mktemp
cat > "$BIN/chmod" << 'MOCK'
#!/usr/bin/env bash
exit 1
MOCK
/usr/bin/chmod +x "$BIN/chmod"
hash -r
if (handoff_root_token >/dev/null 2>"$TMP/handoff-chmod-fail.err"); then
    remove_command_stubs
    fail "handoff_root_token should have failed when chmod failed"
fi
remove_command_stubs
assert_fault_cleanup "$TMP/handoff-chmod-fail.err"
pass "handoff_root_token chmod failure: kubectl not invoked, non-zero, temp file deleted, token off argv and stderr"

# 9g. zero-byte file (PATH printf stub exits 0 with no output): no kubectl, non-zero, file gone
reset_records
rm -f "$RECORD"/handoff-*
install_handoff_mock
install_recording_mktemp
install_printf_stub 0
hash -r
if (
    enable -n printf
    hash -r
    handoff_root_token >/dev/null 2>"$TMP/handoff-empty-fail.err"
); then
    remove_command_stubs
    fail "handoff_root_token should have failed when the temp token file was empty"
fi
remove_command_stubs
assert_fault_cleanup "$TMP/handoff-empty-fail.err"
pass "handoff_root_token zero-byte file: kubectl not invoked, non-zero, temp file deleted, token off argv and stderr"

# 10. sync_unseal_secret: --from-file temp, mode 600, deleted, content ok
FAKE_UNSEAL_KEY="fake-openbao-unseal-test-$(openssl rand -hex 16)"

install_unseal_mock() {
    cat > "$BIN/kubectl" << 'MOCK'
#!/usr/bin/env bash
# Record argv/stdin. For `create secret --from-file=key=PATH`, record the
# path, mode, and file content (never printed). Emit a dry-run Secret yaml
# on stdout for the apply pipe. For `apply`, honor KUBECTL_APPLY_FAIL.
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

from_file=""
ns=""
name=""
prev=""
for a in "$@"; do
    case "$a" in
        --from-file=key=*)
            from_file="${a#--from-file=key=}"
            ;;
        --from-literal=*)
            printf 'from-literal\n' >> "$dir/unseal-from-literal"
            ;;
        -n|--namespace)
            prev=ns
            ;;
        --namespace=*)
            ns="${a#--namespace=}"
            ;;
        *)
            if [ "$prev" = "ns" ]; then
                ns="$a"
                prev=""
            elif [ "$prev" = "generic" ]; then
                name="$a"
                prev=""
            fi
            if [ "$a" = "generic" ]; then
                prev=generic
            fi
            ;;
    esac
done

is_create=0
is_apply=0
for a in "$@"; do
    [ "$a" = "create" ] && is_create=1
    [ "$a" = "apply" ] && is_apply=1
done

if [ "$is_create" = 1 ]; then
    if [ -n "$from_file" ]; then
        printf '%s\n' "$from_file" > "$dir/unseal-key-path"
        if [ -e "$from_file" ]; then
            stat -c '%a' "$from_file" > "$dir/unseal-key-mode"
            cp -a "$from_file" "$dir/unseal-key-content"
        else
            printf 'missing\n' > "$dir/unseal-key-mode"
        fi
    fi
    if [ "${KUBECTL_CREATE_FAIL:-0}" = "1" ]; then
        exit 1
    fi
    [ -n "$from_file" ] && [ -e "$from_file" ] || exit 1
    b64=$(base64 -w0 < "$from_file")
    printf '%s\n' "apiVersion: v1" "kind: Secret"
    printf '%s\n' "metadata:"
    printf '  name: %s\n' "${name:-openbao-unseal-key}"
    printf '  namespace: %s\n' "${ns:-openbao}"
    printf '%s\n' "type: Opaque" "data:"
    printf '  key: %s\n' "$b64"
    exit 0
fi

if [ "$is_apply" = 1 ]; then
    cp "$dir/call-$n.stdin" "$dir/unseal-apply-yaml"
    if [ "${KUBECTL_APPLY_FAIL:-0}" = "1" ]; then
        exit 1
    fi
    exit 0
fi

exit 0
MOCK
    chmod +x "$BIN/kubectl"
}

assert_unseal_file_gone() {
    local path
    [ -s "$RECORD/unseal-key-path" ] || fail "kubectl create did not record a --from-file path"
    path=$(cat "$RECORD/unseal-key-path")
    if [ -e "$path" ]; then
        fail "temp unseal-key file was not removed"
    fi
}

assert_unseal_mode_600() {
    [ -s "$RECORD/unseal-key-mode" ] || fail "kubectl create did not record the temp file mode"
    [ "$(cat "$RECORD/unseal-key-mode")" = "600" ] || fail "temp unseal-key file was not mode 600 when kubectl ran"
}

assert_unseal_content() {
    [ -f "$RECORD/unseal-key-content" ] || fail "kubectl create did not record the temp file content"
    FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$RECORD/unseal-key-content" exact
}

assert_unseal_not_in_argv() {
    local f
    shopt -s nullglob
    for f in "$RECORD"/call-*.argv "$RECORD"/call-*.argv-joined; do
        [ -f "$f" ] || continue
        FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$f" absent
    done
    shopt -u nullglob
}

assert_apply_unseal_content() {
    [ -f "$RECORD/unseal-apply-yaml" ] || fail "kubectl apply did not receive yaml stdin"
    FAKE_TOKEN="$FAKE_UNSEAL_KEY" APPLY_YAML="$RECORD/unseal-apply-yaml" python3 - << 'PYCHECK'
import os, sys, base64, re
from pathlib import Path
secret = os.environ["FAKE_TOKEN"].encode()
raw = Path(os.environ["APPLY_YAML"]).read_bytes()
text = raw.decode()
m = re.search(r"(?m)^\s*key:\s*([A-Za-z0-9+/=]+)\s*$", text)
if not m:
    sys.exit("apply yaml has no data.key")
try:
    decoded = base64.b64decode(m.group(1), validate=True)
except Exception:
    sys.exit("apply yaml key was not valid base64")
if decoded != secret:
    sys.exit("apply yaml key does not match the fake unseal key (%d vs %d bytes)" % (len(decoded), len(secret)))
if secret in raw:
    sys.exit("apply yaml contains the raw unseal key")
PYCHECK
}

export UNSEAL_KEY="$FAKE_UNSEAL_KEY"
export OPENBAO_NS="${OPENBAO_NS:-openbao}"
export UNSEAL_SECRET="${UNSEAL_SECRET:-openbao-unseal-key}"

# 10a. success
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
sync_unseal_secret >/dev/null
assert_unseal_not_in_argv
if [ -f "$RECORD/unseal-from-literal" ]; then
    fail "kubectl was invoked with --from-literal"
fi
assert_unseal_mode_600
assert_unseal_content
assert_apply_unseal_content
assert_unseal_file_gone
if ! grep -F -q -- '--from-file=key=' "$RECORD"/call-*.argv; then
    fail "kubectl create was not invoked with --from-file=key="
fi
if ! grep -F -q -- '--dry-run=client' "$RECORD"/call-*.argv; then
    fail "kubectl create dropped --dry-run=client"
fi
pass "sync_unseal_secret success: mode 600, secret content correct, file deleted, key off argv"

# 10b. apply failure: temp file still deleted
# Run in a subshell: bootstrap fail() exits, which must not abort this test script.
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
if (KUBECTL_APPLY_FAIL=1 sync_unseal_secret >/dev/null 2>"$TMP/unseal-apply-fail.err"); then
    fail "sync_unseal_secret should have failed when kubectl apply failed"
fi
assert_unseal_not_in_argv
assert_unseal_mode_600
assert_unseal_content
assert_unseal_file_gone
FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$TMP/unseal-apply-fail.err" absent
pass "sync_unseal_secret apply failure: temp file deleted, key off argv and stderr"

# 10c. create failure: temp file still deleted
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
if (KUBECTL_CREATE_FAIL=1 sync_unseal_secret >/dev/null 2>"$TMP/unseal-create-fail.err"); then
    fail "sync_unseal_secret should have failed when kubectl create failed"
fi
assert_unseal_not_in_argv
assert_unseal_mode_600
assert_unseal_content
assert_unseal_file_gone
FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$TMP/unseal-create-fail.err" absent
pass "sync_unseal_secret create failure: temp file deleted, key off argv and stderr"

# 10d. xtrace does not leak the unseal key
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
set +e
(
    set -x
    sync_unseal_secret
) >"$TMP/unseal-xtrace.out" 2>"$TMP/unseal-xtrace.err"
xrc=$?
set -e
[ "$xrc" -eq 0 ] || fail "sync_unseal_secret failed under xtrace"
FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$TMP/unseal-xtrace.out" absent
FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$TMP/unseal-xtrace.err" absent
assert_unseal_not_in_argv
assert_unseal_file_gone
pass "sync_unseal_secret does not leak the unseal key under xtrace"

# --- 10e-g. write/chmod/empty-file faults must not reach kubectl ---
# Same UID-independent PATH shims as 9e-g. Do not use file modes.
install_recording_mktemp_unseal() {
    cat > "$BIN/mktemp" << 'MOCK'
#!/usr/bin/env bash
set -euo pipefail
dir=${KUBECTL_RECORD_DIR:?}
real=$(PATH=/usr/bin:/bin:/usr/sbin:/sbin type -P mktemp)
f=$("$real" "$@")
printf '%s\n' "$f" > "$dir/unseal-key-path"
printf '%s\n' "$f"
MOCK
    /usr/bin/chmod +x "$BIN/mktemp"
}

assert_unseal_fault_cleanup() {
    local err="$1"
    assert_kubectl_not_invoked
    assert_unseal_not_in_argv
    FAKE_TOKEN="$FAKE_UNSEAL_KEY" python_check "$err" absent
    assert_unseal_file_gone
}

# 10e. write failure (PATH printf stub exits 1)
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
install_recording_mktemp_unseal
install_printf_stub 1
hash -r
if (
    enable -n printf
    hash -r
    sync_unseal_secret >/dev/null 2>"$TMP/unseal-write-fail.err"
); then
    remove_command_stubs
    fail "sync_unseal_secret should have failed when writing the temp unseal-key file failed"
fi
remove_command_stubs
assert_unseal_fault_cleanup "$TMP/unseal-write-fail.err"
pass "sync_unseal_secret write failure: kubectl not invoked, non-zero, temp file deleted, key off argv and stderr"

# 10f. chmod failure (stubbed chmod)
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
install_recording_mktemp_unseal
cat > "$BIN/chmod" << 'MOCK'
#!/usr/bin/env bash
exit 1
MOCK
/usr/bin/chmod +x "$BIN/chmod"
hash -r
if (sync_unseal_secret >/dev/null 2>"$TMP/unseal-chmod-fail.err"); then
    remove_command_stubs
    fail "sync_unseal_secret should have failed when chmod failed"
fi
remove_command_stubs
assert_unseal_fault_cleanup "$TMP/unseal-chmod-fail.err"
pass "sync_unseal_secret chmod failure: kubectl not invoked, non-zero, temp file deleted, key off argv and stderr"

# 10g. zero-byte file (PATH printf stub exits 0 with no output)
reset_records
rm -f "$RECORD"/unseal-*
install_unseal_mock
install_recording_mktemp_unseal
install_printf_stub 0
hash -r
if (
    enable -n printf
    hash -r
    sync_unseal_secret >/dev/null 2>"$TMP/unseal-empty-fail.err"
); then
    remove_command_stubs
    fail "sync_unseal_secret should have failed when the temp unseal-key file was empty"
fi
remove_command_stubs
assert_unseal_fault_cleanup "$TMP/unseal-empty-fail.err"
pass "sync_unseal_secret zero-byte file: kubectl not invoked, non-zero, temp file deleted, key off argv and stderr"

printf 'All tests passed\n'
