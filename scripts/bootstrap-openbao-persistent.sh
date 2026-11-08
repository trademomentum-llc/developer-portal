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

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
ESO_NS=${ESO_NS:-external-secrets}
ESO_SECRET=${ESO_SECRET:-openbao-root-token}
UNSEAL_SECRET=${UNSEAL_SECRET:-openbao-unseal-key}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}

info() { printf "\033[1;36m[openbao-bootstrap]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-bootstrap ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

UNSEAL_KEY=""
ROOT_TOKEN=""

# bao invocations that need no auth (init, init -status, status, unseal).
bao_raw() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- bao "$@"
}

# bao invocations authenticated with the current root token.
bao_auth() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" bao "$@"
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

sync_unseal_secret() {
    kubectl -n "$OPENBAO_NS" create secret generic "$UNSEAL_SECRET" \
        --from-literal=key="$UNSEAL_KEY" --dry-run=client -o yaml \
        | kubectl apply -f - >/dev/null
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
    # $KUBERNETES_PORT_443_TCP_ADDR expands IN THE POD (single-quoted sh -c).
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" sh -c \
        'bao write auth/kubernetes/config kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"' >/dev/null

    kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" \
        bao policy write openchoreo-secret-reader-policy - >/dev/null <<'POLICY'
path "secret/data/*" { capabilities = ["read"] }
path "secret/metadata/*" { capabilities = ["list", "read"] }
POLICY

    kubectl -n "$OPENBAO_NS" exec -i "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" \
        bao policy write openchoreo-secret-writer-policy - >/dev/null <<'POLICY'
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

put_if_absent() {
    local path="$1" value="$2"
    if ! bao_auth kv get "$path" >/dev/null 2>&1; then
        bao_auth kv put "$path" value="$value" >/dev/null
    fi
}

# The thirteen static platform secrets from the retired postStart, values
# verbatim (section 1.2). Existing values are preserved; rotation is an
# explicit future change, never a bootstrap side effect.
seed_platform_secrets() {
    put_if_absent secret/npm-token "fake-npm-token-for-development"
    put_if_absent secret/docker-username "dev-user"
    put_if_absent secret/docker-password "dev-password"
    put_if_absent secret/github-pat "fake-github-token-for-development"
    put_if_absent secret/username "dev-user"
    put_if_absent secret/password "dev-password"
    put_if_absent secret/backstage-backend-secret "local-dev-backend-secret"
    put_if_absent secret/backstage-client-secret "backstage-portal-secret"
    put_if_absent secret/backstage-jenkins-api-key "placeholder-not-in-use"
    put_if_absent secret/observer-oauth-client-secret "openchoreo-observer-resource-reader-client-secret"
    put_if_absent secret/rca-oauth-client-secret "openchoreo-rca-agent-secret"
    put_if_absent secret/opensearch-username "admin"
    put_if_absent secret/opensearch-password "ThisIsTheOpenSearchPassword1"
}

# D-07: the ClusterSecretStore openbao-kv reads its token from this Secret;
# after bootstrap it must hold the generated root token, not the literal root.
handoff_root_token() {
    kubectl -n "$ESO_NS" create secret generic "$ESO_SECRET" \
        --from-literal=token="$ROOT_TOKEN" --dry-run=client -o yaml \
        | kubectl apply -f - >/dev/null
    info "secret $ESO_SECRET (ns $ESO_NS) now holds the generated root token"
}

main() {
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

main "$@"
