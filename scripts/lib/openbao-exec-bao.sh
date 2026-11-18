# shellcheck shell=bash
# scripts/lib/openbao-exec-bao.sh
# Authenticated `bao` inside the OpenBao pod, without placing the root token
# on kubectl's (or bao's) argv.
#
# Required environment:
#   OPENBAO_NS, OPENBAO_POD
#   ROOT_TOKEN or OPENBAO_TOKEN — first non-empty value is used
#
# openbao_exec_bao [bao-args...]
#   The token is written to `kubectl exec -i` stdin and read in the pod with
#   BAO_TOKEN=$(cat). No other stdin is forwarded.
#
# openbao_exec_bao_stdin [bao-args...]
#   Stdin protocol: first line is the token; remaining bytes are bao's stdin
#   (policy documents, `value=-` payloads, and similar).
#
# openbao_exec_pod_sh 'shell script'
#   Same token handoff as openbao_exec_bao, then `eval` of the script inside
#   the pod so pod-only variables (e.g. $KUBERNETES_PORT_443_TCP_ADDR) expand
#   there. The script is a constant provided by our scripts, never caller
#   input.
#
# Inside the pod the token is exported as BAO_TOKEN and VAULT_TOKEN. It never
# appears in kubectl or bao argv. xtrace is disabled at the start of each
# public helper and of _openbao_exec_prepare, before any token reference, so
# `set -x` cannot leak it to logs. The caller's xtrace flag is restored on
# every exit path.

_openbao_exec_prepare() {
    { local _openbao_x=$-; set +x; } 2>/dev/null
    trap '[[ ${_openbao_x:-} == *x* ]] && set -x' RETURN
    if [ -z "${OPENBAO_NS:-}" ] || [ -z "${OPENBAO_POD:-}" ]; then
        printf 'openbao-exec-bao: OPENBAO_NS and OPENBAO_POD must be set\n' >&2
        return 1
    fi
    if [ -z "${ROOT_TOKEN:-}" ] && [ -z "${OPENBAO_TOKEN:-}" ]; then
        printf 'openbao-exec-bao: missing ROOT_TOKEN or OPENBAO_TOKEN\n' >&2
        return 1
    fi
    return 0
}

# Write the token to stdout (callers must not log or echo this stream).
_openbao_exec_write_token() {
    if [ -n "${ROOT_TOKEN:-}" ]; then
        printf '%s' "$ROOT_TOKEN"
    else
        printf '%s' "$OPENBAO_TOKEN"
    fi
}

openbao_exec_bao() {
    { local _openbao_x=$-; set +x; } 2>/dev/null
    trap '[[ ${_openbao_x:-} == *x* ]] && set -x' RETURN
    _openbao_exec_prepare || return 1
    local _rc=0
    # Token is read inside the pod; host must not expand $(cat) / $BAO_TOKEN.
    # shellcheck disable=SC2016
    _openbao_exec_write_token | kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- \
        sh -c 'BAO_TOKEN=$(cat); export BAO_TOKEN VAULT_TOKEN="$BAO_TOKEN"; exec bao "$@"' \
        sh "$@" || _rc=$?
    return "$_rc"
}

openbao_exec_bao_stdin() {
    { local _openbao_x=$-; set +x; } 2>/dev/null
    trap '[[ ${_openbao_x:-} == *x* ]] && set -x' RETURN
    _openbao_exec_prepare || return 1
    local _rc=0
    # Token is read inside the pod; host must not expand $BAO_TOKEN.
    # shellcheck disable=SC2016
    {
        _openbao_exec_write_token
        printf '\n'
        cat
    } | kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- \
        sh -c 'IFS= read -r BAO_TOKEN; export BAO_TOKEN VAULT_TOKEN="$BAO_TOKEN"; exec bao "$@"' \
        sh "$@" || _rc=$?
    return "$_rc"
}

openbao_exec_pod_sh() {
    { local _openbao_x=$-; set +x; } 2>/dev/null
    trap '[[ ${_openbao_x:-} == *x* ]] && set -x' RETURN
    local _script="${1:-}"
    if [ -z "$_script" ]; then
        printf 'openbao-exec-bao: openbao_exec_pod_sh requires a script\n' >&2
        return 1
    fi
    _openbao_exec_prepare || return 1
    local _rc=0
    # shellcheck disable=SC2016
    _openbao_exec_write_token | kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- \
        sh -c 'BAO_TOKEN=$(cat); export BAO_TOKEN VAULT_TOKEN="$BAO_TOKEN"; eval "$1"' \
        sh "$_script" || _rc=$?
    return "$_rc"
}
