#!/usr/bin/env bash
# Seed the OpenBao kv v2 paths M2 relies on. Idempotent.
#
# Post-G5 this is a one-time bootstrap component invoked by
# scripts/bootstrap-openbao-persistent.sh, not a recovery tool (BAO-STORAGE-
# DES-001 D-06). Authenticates with the root token from the custody file
# ~/.rational-reserve/openbao/root-token by default; OPENBAO_TOKEN overrides;
# the literal "root" fallback only matches the pre-migration dev-mode
# backend. Mounts are (re-)enabled before writing so the script is also safe
# against a fresh backend.
set -euo pipefail

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}
if [ -z "${OPENBAO_TOKEN:-}" ] && [ -r "$CUSTODY_DIR/root-token" ]; then
    OPENBAO_TOKEN=$(cat "$CUSTODY_DIR/root-token")
fi
OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}
GITEA_RUNNER_SECRET_NS=${GITEA_RUNNER_SECRET_NS:-gitea-runners}
GITEA_RUNNER_SECRET_NAME=${GITEA_RUNNER_SECRET_NAME:-gitea-runner-token}

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$OPENBAO_TOKEN" bao "$@"
}

read_existing_runner_token() {
    local encoded

    encoded=$(kubectl -n "$GITEA_RUNNER_SECRET_NS" get secret "$GITEA_RUNNER_SECRET_NAME" \
        -o jsonpath='{.data.token}' 2>/dev/null || true)
    if [ -n "$encoded" ]; then
        printf '%s' "$encoded" | base64 --decode
    fi
}

ensure_kv_v2_mount() {
    local mount="$1"

    # C6 (BAO-STORAGE-SIM-001 D3): grep -q exits on first match; kubectl
    # exec streams frames, so under node load the producer is SIGPIPEd and
    # pipefail misreads a present mount as absent. Plain grep reads to EOF.
    if ! exec_bao secrets list 2>/dev/null | grep "^${mount}/" >/dev/null; then
        exec_bao secrets enable -path="$mount" -version=2 kv >/dev/null
    fi
}

# kv/ is used by M2-owned ExternalSecrets. secret/ is used by OpenChoreo's
# default ClusterSecretStore for generated application runtime secrets.
ensure_kv_v2_mount kv
ensure_kv_v2_mount secret

# Runner registration token. Prefer GITEA_RUNNER_TOKEN, recover from the
# existing Kubernetes Secret when OpenBao dev storage lost state, then prompt
# as the final interactive fallback.
if ! exec_bao kv get kv/gitea/runners/token >/dev/null 2>&1; then
    TOKEN=
    if [ -n "${GITEA_RUNNER_TOKEN:-}" ]; then
        TOKEN="$GITEA_RUNNER_TOKEN"
    else
        TOKEN=$(read_existing_runner_token)
    fi

    if [ -n "$TOKEN" ]; then
        exec_bao kv put kv/gitea/runners/token token="$TOKEN" >/dev/null
    elif [ "${SEED_OPENBAO_SKIP_RUNNER_TOKEN:-0}" = "1" ]; then
        echo "Skipping missing Gitea runner registration token"
    elif [ -t 0 ]; then
        printf "Gitea runner registration token: "
        read -rs TOKEN
        echo
        exec_bao kv put kv/gitea/runners/token token="$TOKEN" >/dev/null
    else
        echo "GITEA_RUNNER_TOKEN is required when stdin is not interactive" >&2
        exit 1
    fi
fi

# Flux git auth -- gitea_admin + its token.
if ! exec_bao kv get kv/flux/gitea-deploy-key >/dev/null 2>&1; then
    FLUX_USER=gitea_admin
    FLUX_PASS=$(cat ~/.rational-reserve/m1-gitea-admin-password)
    exec_bao kv put kv/flux/gitea-deploy-key username="$FLUX_USER" password="$FLUX_PASS" >/dev/null
fi

# Demo app secret. The secret/ mount is the live OpenChoreo runtime path.
# The kv/ mirror keeps the older M2 ExternalSecret smoke path populated.
APP_SECRET_VALUE=
if [ "${SEED_OPENBAO_ROTATE_APP_SECRET:-0}" = "1" ]; then
    APP_SECRET_VALUE="demo-$(openssl rand -hex 8)"
elif exec_bao kv get secret/apps/hello-m2/dev/example-secret >/dev/null 2>&1; then
    APP_SECRET_VALUE=$(exec_bao kv get -field=password secret/apps/hello-m2/dev/example-secret)
elif exec_bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null 2>&1; then
    APP_SECRET_VALUE=$(exec_bao kv get -field=password kv/apps/hello-m2/dev/example-secret)
else
    APP_SECRET_VALUE="demo-$(openssl rand -hex 8)"
fi

if [ "${SEED_OPENBAO_ROTATE_APP_SECRET:-0}" = "1" ] ||
    ! exec_bao kv get secret/apps/hello-m2/dev/example-secret >/dev/null 2>&1; then
    exec_bao kv put secret/apps/hello-m2/dev/example-secret password="$APP_SECRET_VALUE" >/dev/null
fi

if [ "${SEED_OPENBAO_ROTATE_APP_SECRET:-0}" = "1" ] ||
    ! exec_bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null 2>&1; then
    exec_bao kv put kv/apps/hello-m2/dev/example-secret password="$APP_SECRET_VALUE" >/dev/null
fi

echo "openbao seeded"
