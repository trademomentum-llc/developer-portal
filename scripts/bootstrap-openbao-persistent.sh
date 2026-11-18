#!/usr/bin/env bash
# scripts/bootstrap-openbao-persistent.sh
#
# One-time bootstrap for the Raft-backed OpenBao server (BAO-STORAGE-DES-001
# D-06). Replaces the dev-mode helm postStart script (content preserved
# verbatim below, measured 2026-08-26). Run once per storage lifetime by
# scripts/install-openbao-storage.sh; safe to re-run (idempotent).
#
# Steps: wait for API -> init or recover from custody -> sync the unseal
# Secret -> unseal + verify token -> enable mounts -> auth/policies/roles ->
# thirteen platform secrets -> M2 seed -> root-token handoff to
# ExternalSecrets.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=scripts/lib/openbao-exec-bao.sh
source "$ROOT/scripts/lib/openbao-exec-bao.sh"

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
ESO_NS=${ESO_NS:-external-secrets}
ESO_SECRET=${ESO_SECRET:-openbao-root-token}
UNSEAL_SECRET=${UNSEAL_SECRET:-openbao-unseal-key}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}

# OpenSearch admin password (secret/opensearch-password). Never hard-coded;
# resolved by resolve_opensearch_password() in this order:
#   1. $OPENSEARCH_ADMIN_PASSWORD (non-empty)
#   2. the file at $OPENSEARCH_ADMIN_PASSWORD_FILE (default: custody dir)
#   3. a freshly generated random value written to that file (mode 600)
# OPENBAO_REQUIRE_EXPLICIT_SECRETS=1 disables step 3 (fail-fast): the script
# exits before touching OpenBao unless 1 or 2 supplies a value.
# See .env.example. The value is never echoed or logged.
OPENSEARCH_ADMIN_PASSWORD_FILE=${OPENSEARCH_ADMIN_PASSWORD_FILE:-"$CUSTODY_DIR/opensearch-admin-password"}
OPENBAO_REQUIRE_EXPLICIT_SECRETS=${OPENBAO_REQUIRE_EXPLICIT_SECRETS:-0}
OPENSEARCH_PW=""

info() { printf "\033[1;36m[openbao-bootstrap]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-bootstrap ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

UNSEAL_KEY=""
ROOT_TOKEN=""

# bao invocations that need no auth (init, init -status, status, unseal).
bao_raw() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- bao "$@"
}

# bao invocations authenticated with the current root token.
# The token is delivered over kubectl exec stdin, never argv (see
# scripts/lib/openbao-exec-bao.sh).
bao_auth() {
    openbao_exec_bao "$@"
}

# bao operator init -status: exit 0 initialized, 2 uninitialized; both mean
# the API is answering. Any other exit means the server is still coming up.
# (CLI contract carried from vault-compatible bao; spec-time assertion,
# exercised at Phase 2.)
wait_for_api() {
    local tries=0 rc=0
    while [ "$tries" -lt 60 ]; do
        rc=0
        bao_raw operator init -status >/dev/null 2>&1 || rc=$?
        if [ "$rc" -eq 0 ] || [ "$rc" -eq 2 ]; then
            return 0
        fi
        tries=$((tries + 1))
        sleep 3
    done
    fail "openbao API did not answer within 180s"
}

init_or_recover() {
    if bao_raw operator init -status >/dev/null 2>&1; then
        info "Raft backend already initialized; recovering key material from custody"
        [ -r "$CUSTODY_DIR/unseal-key" ] || fail "initialized backend but $CUSTODY_DIR/unseal-key is missing"
        [ -r "$CUSTODY_DIR/root-token" ] || fail "initialized backend but $CUSTODY_DIR/root-token is missing"
        UNSEAL_KEY=$(cat "$CUSTODY_DIR/unseal-key")
        ROOT_TOKEN=$(cat "$CUSTODY_DIR/root-token")
    else
        info "Initializing Raft backend (Shamir key-shares=1 key-threshold=1)"
        local out
        out=$(bao_raw operator init -key-shares=1 -key-threshold=1) || fail "bao operator init failed"
        UNSEAL_KEY=$(printf '%s\n' "$out" | sed -n 's/^Unseal Key 1: //p')
        ROOT_TOKEN=$(printf '%s\n' "$out" | sed -n 's/^Initial Root Token: //p')
        [ -n "$UNSEAL_KEY" ] || fail "could not parse unseal key from init output"
        [ -n "$ROOT_TOKEN" ] || fail "could not parse root token from init output"
        umask 077
        mkdir -p "$CUSTODY_DIR"
        chmod 700 "$CUSTODY_DIR"
        printf '%s\n' "$UNSEAL_KEY" > "$CUSTODY_DIR/unseal-key"
        printf '%s\n' "$ROOT_TOKEN" > "$CUSTODY_DIR/root-token"
        chmod 600 "$CUSTODY_DIR/unseal-key" "$CUSTODY_DIR/root-token"
        info "custody written: $CUSTODY_DIR (dir 700, files 600)"
    fi
}

# The unseal key is written to a 0600 temp file (umask 077) and passed with
# --from-file, never --from-literal, so it does not appear in kubectl argv,
# xtrace, or logs. A trap on EXIT/INT/TERM removes the file.
sync_unseal_secret() {
    # Disable xtrace before any key reference; restore on every return.
    { local _openbao_x=$-; set +x; } 2>/dev/null
    trap '[[ ${_openbao_x:-} == *x* ]] && set -x' RETURN
    local _rc=0 _had_e=0
    if [ -z "${UNSEAL_KEY:-}" ]; then
        fail "sync_unseal_secret: UNSEAL_KEY is empty"
    fi
    # Do not put the subshell on the left of || — bash then ignores errexit
    # inside it, so a failed printf/chmod could still reach kubectl and
    # replace the unseal Secret with a zero-byte key. Capture status after
    # the subshell instead, and fail each step with an explicit || exit 1.
    case "$-" in
        *e*) _had_e=1 ;;
    esac
    set +e
    (
        set +x
        set -euo pipefail
        umask 077
        tmp=$(mktemp) || exit 1
        trap 'rm -f -- "$tmp"' EXIT INT TERM
        printf '%s' "$UNSEAL_KEY" > "$tmp" || exit 1
        chmod 600 "$tmp" || exit 1
        if ! [ -s "$tmp" ]; then
            printf 'sync_unseal_secret: temp unseal-key file is empty\n' >&2
            exit 1
        fi
        _got=$(wc -c < "$tmp" | tr -d '[:space:]') || exit 1
        if [ "$_got" -ne "${#UNSEAL_KEY}" ]; then
            printf 'sync_unseal_secret: temp unseal-key file size mismatch\n' >&2
            exit 1
        fi
        kubectl -n "$OPENBAO_NS" create secret generic "$UNSEAL_SECRET" \
            --from-file=key="$tmp" --dry-run=client -o yaml \
            | kubectl apply -f - >/dev/null || exit 1
    )
    _rc=$?
    [ "$_had_e" -eq 1 ] && set -e
    [ "$_rc" -eq 0 ] || fail "could not apply secret $UNSEAL_SECRET (ns $OPENBAO_NS)"
    info "secret $UNSEAL_SECRET synced (ns $OPENBAO_NS)"
}

ensure_unsealed() {
    local status
    status=$(bao_raw status -format=json 2>/dev/null || true)
    case "$status" in
        *'"sealed":true'*|*'"sealed": true'*)
            info "server sealed; unsealing (one-time bootstrap action; the sidecar covers later restarts)"
            bao_raw operator unseal "$UNSEAL_KEY" >/dev/null || fail "bao operator unseal failed"
            ;;
    esac
    kubectl -n "$OPENBAO_NS" wait --for=condition=Ready "pod/$OPENBAO_POD" --timeout=120s \
        || fail "pod/$OPENBAO_POD did not become Ready after unseal"
    # Sanity: the custody root token is valid against this backend. A mismatch
    # means custody and cluster disagree; fail loudly rather than re-init.
    bao_auth token lookup >/dev/null 2>&1 \
        || fail "custody root token rejected by the backend (custody/cluster mismatch)"
}

# Fresh Raft has NO mounts (dev mode's pre-mounted secret/ is gone). Enable
# both kv-v2 mounts before any secret write. Idempotent.
# C6 (BAO-STORAGE-SIM-001 D3): plain grep instead of grep -q -- the -q early
# exit SIGPIPEs the framed kubectl-exec producer under load and pipefail then
# misreads a present mount as absent.
ensure_mounts() {
    if ! bao_auth secrets list 2>/dev/null | grep '^kv/' >/dev/null; then
        bao_auth secrets enable -path=kv -version=2 kv >/dev/null
    fi
    if ! bao_auth secrets list 2>/dev/null | grep '^secret/' >/dev/null; then
        bao_auth secrets enable -path=secret -version=2 kv >/dev/null
    fi
}

# Durable configuration, verbatim from the retired postStart (section 1.2),
# minus the dev-mode token export. All writes are idempotent against Raft.
apply_durable_config() {
    bao_auth auth enable kubernetes >/dev/null 2>&1 || true
    # $KUBERNETES_PORT_443_TCP_ADDR expands IN THE POD (eval'd there).
    # shellcheck disable=SC2016
    openbao_exec_pod_sh \
        'bao write auth/kubernetes/config kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"' >/dev/null

    openbao_exec_bao_stdin \
        policy write openchoreo-secret-reader-policy - >/dev/null <<'POLICY'
path "secret/data/*" { capabilities = ["read"] }
path "secret/metadata/*" { capabilities = ["list", "read"] }
POLICY

    openbao_exec_bao_stdin \
        policy write openchoreo-secret-writer-policy - >/dev/null <<'POLICY'
path "secret/data/*" { capabilities = ["create", "read", "update", "delete"] }
path "secret/metadata/*" { capabilities = ["create", "read", "update", "delete", "list"] }
POLICY

    bao_auth write auth/kubernetes/role/openchoreo-secret-reader-role \
        bound_service_account_names=default \
        bound_service_account_namespaces="dp*" \
        policies=openchoreo-secret-reader-policy ttl=20m >/dev/null

    bao_auth write auth/kubernetes/role/openchoreo-secret-writer-role \
        bound_service_account_names="*" \
        bound_service_account_namespaces="openbao,openchoreo-workflow-plane" \
        policies=openchoreo-secret-writer-policy ttl=20m >/dev/null
}

# put_if_absent <kv-path>  -- the value is read from STDIN, never argv.
# The root token (first line) and the value (remaining bytes) reach the pod
# over the kubectl exec stdin stream; bao reads the value via `value=-`.
# Neither the token nor the value appears in kubectl or bao argv (ps, audit
# logs, shell traces). printf is a bash builtin, so feeding it does not fork
# a process either. Existing values are preserved (no overwrite); rotation
# is deliberate.
put_if_absent() {
    local path="$1" value
    value=$(cat)
    [ -n "$value" ] || fail "put_if_absent $path: empty value on stdin"
    if ! bao_auth kv get "$path" >/dev/null 2>&1; then
        printf '%s' "$value" \
            | openbao_exec_bao_stdin kv put "$path" value=- >/dev/null \
            || fail "bao kv put $path failed"
    fi
}

# Resolve the OpenSearch admin password into OPENSEARCH_PW (not exported).
# Order: env -> custody file -> generate (unless fail-fast). Logs only the
# source, never the value.
resolve_opensearch_password() {
    if [ -n "${OPENSEARCH_ADMIN_PASSWORD:-}" ]; then
        OPENSEARCH_PW="$OPENSEARCH_ADMIN_PASSWORD"
        info "opensearch admin password sourced from env (OPENSEARCH_ADMIN_PASSWORD)"
    elif [ -r "$OPENSEARCH_ADMIN_PASSWORD_FILE" ]; then
        OPENSEARCH_PW=$(cat "$OPENSEARCH_ADMIN_PASSWORD_FILE")
        [ -n "$OPENSEARCH_PW" ] || fail "$OPENSEARCH_ADMIN_PASSWORD_FILE is empty"
        info "opensearch admin password sourced from custody file ($OPENSEARCH_ADMIN_PASSWORD_FILE)"
    elif [ "$OPENBAO_REQUIRE_EXPLICIT_SECRETS" = "1" ]; then
        fail "OPENBAO_REQUIRE_EXPLICIT_SECRETS=1: set OPENSEARCH_ADMIN_PASSWORD or provide $OPENSEARCH_ADMIN_PASSWORD_FILE (see .env.example); refusing to generate or seed a default"
    else
        command -v openssl >/dev/null 2>&1 || fail "openssl not found; cannot generate the opensearch admin password (set OPENSEARCH_ADMIN_PASSWORD instead)"
        (
            umask 077
            mkdir -p "$(dirname "$OPENSEARCH_ADMIN_PASSWORD_FILE")" \
                && openssl rand -base64 48 | tr -d '/+=\n' | cut -c1-32 > "$OPENSEARCH_ADMIN_PASSWORD_FILE" \
                && chmod 600 "$OPENSEARCH_ADMIN_PASSWORD_FILE"
        ) || fail "could not write $OPENSEARCH_ADMIN_PASSWORD_FILE"
        OPENSEARCH_PW=$(cat "$OPENSEARCH_ADMIN_PASSWORD_FILE")
        [ "${#OPENSEARCH_PW}" -ge 24 ] || fail "generated opensearch admin password is too short"
        info "opensearch admin password sourced from generated value (custody file $OPENSEARCH_ADMIN_PASSWORD_FILE, mode 600)"
        info "NOTE: OpenSearch itself must be configured with this same value (see AGENTS.md)"
    fi
}

# The thirteen platform secrets from the retired postStart (section 1.2).
# Twelve are verbatim dev placeholders; the OpenSearch admin password is
# resolved at runtime (resolve_opensearch_password) and is no longer a
# literal. Values are fed via stdin. Existing values are preserved; rotation
# is an explicit change, never a bootstrap side effect.
seed_platform_secrets() {
    put_if_absent secret/npm-token <<<"fake-npm-token-for-development"
    put_if_absent secret/docker-username <<<"dev-user"
    put_if_absent secret/docker-password <<<"dev-password"
    put_if_absent secret/github-pat <<<"fake-github-token-for-development"
    put_if_absent secret/username <<<"dev-user"
    put_if_absent secret/password <<<"dev-password"
    put_if_absent secret/backstage-backend-secret <<<"local-dev-backend-secret"
    put_if_absent secret/backstage-client-secret <<<"backstage-portal-secret"
    put_if_absent secret/backstage-jenkins-api-key <<<"placeholder-not-in-use"
    put_if_absent secret/observer-oauth-client-secret <<<"openchoreo-observer-resource-reader-client-secret"
    put_if_absent secret/rca-oauth-client-secret <<<"openchoreo-rca-agent-secret"
    put_if_absent secret/opensearch-username <<<"admin"
    put_if_absent secret/opensearch-password <<<"$OPENSEARCH_PW"
}

# D-07: the ClusterSecretStore openbao-kv reads its token from this Secret;
# after bootstrap it must hold the generated root token, not the literal root.
# The token is written to a 0600 temp file (umask 077) and passed with
# --from-file, never --from-literal, so it does not appear in kubectl argv,
# xtrace, or logs. A trap on EXIT/INT/TERM removes the file.
handoff_root_token() {
    # Disable xtrace before any token reference; restore on every return.
    { local _openbao_x=$-; set +x; } 2>/dev/null
    trap '[[ ${_openbao_x:-} == *x* ]] && set -x' RETURN
    local _rc=0 _had_e=0
    if [ -z "${ROOT_TOKEN:-}" ]; then
        fail "handoff_root_token: ROOT_TOKEN is empty"
    fi
    # Do not put the subshell on the left of || — bash then ignores errexit
    # inside it, so a failed printf/chmod could still reach kubectl and
    # install an empty or truncated token. Capture status after the subshell
    # instead, and fail each step with an explicit || exit 1.
    case "$-" in
        *e*) _had_e=1 ;;
    esac
    set +e
    (
        set +x
        set -euo pipefail
        umask 077
        tmp=$(mktemp) || exit 1
        trap 'rm -f -- "$tmp"' EXIT INT TERM
        printf '%s' "$ROOT_TOKEN" > "$tmp" || exit 1
        chmod 600 "$tmp" || exit 1
        if ! [ -s "$tmp" ]; then
            printf 'handoff_root_token: temp token file is empty\n' >&2
            exit 1
        fi
        _got=$(wc -c < "$tmp" | tr -d '[:space:]') || exit 1
        if [ "$_got" -ne "${#ROOT_TOKEN}" ]; then
            printf 'handoff_root_token: temp token file size mismatch\n' >&2
            exit 1
        fi
        kubectl -n "$ESO_NS" create secret generic "$ESO_SECRET" \
            --from-file=token="$tmp" --dry-run=client -o yaml \
            | kubectl apply -f - >/dev/null || exit 1
    )
    _rc=$?
    [ "$_had_e" -eq 1 ] && set -e
    [ "$_rc" -eq 0 ] || fail "could not apply secret $ESO_SECRET (ns $ESO_NS)"
    info "secret $ESO_SECRET (ns $ESO_NS) now holds the generated root token"
}

main() {
    # Fail fast (before touching the cluster) if the password is unavailable.
    resolve_opensearch_password
    wait_for_api
    init_or_recover
    sync_unseal_secret
    ensure_unsealed
    ensure_mounts
    apply_durable_config
    seed_platform_secrets
    info "seeding the four M2 kv paths"
    OPENBAO_TOKEN="$ROOT_TOKEN" "$ROOT/scripts/seed-openbao-m2-paths.sh"
    handoff_root_token
    info "bootstrap complete"
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
    main "$@"
fi
