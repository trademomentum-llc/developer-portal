#!/usr/bin/env bash
# Seed the three openbao kv v2 paths M2 relies on. Idempotent.
#
# Authenticates as the dev-mode root token (VAULT_TOKEN=root by default;
# override via OPENBAO_TOKEN env). openbao's dev-mode inmem storage loses
# both the kv mount and its contents whenever the pod restarts, so we
# (re-)enable the kv-v2 mount here before writing.
set -euo pipefail

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$OPENBAO_TOKEN" bao "$@"
}

# Ensure kv v2 mount exists at kv/ (idempotent; silent no-op if present).
if ! exec_bao secrets list 2>/dev/null | grep -q '^kv/'; then
    exec_bao secrets enable -path=kv -version=2 kv >/dev/null
fi

# Runner registration token -- operator must fetch from Gitea and paste.
# Override non-interactively via GITEA_RUNNER_TOKEN env var.
if ! exec_bao kv get kv/gitea/runners/token >/dev/null 2>&1; then
    if [ -n "${GITEA_RUNNER_TOKEN:-}" ]; then
        TOKEN="$GITEA_RUNNER_TOKEN"
    else
        printf "Gitea runner registration token: "
        read -rs TOKEN
        echo
    fi
    exec_bao kv put kv/gitea/runners/token token="$TOKEN"
fi

# Flux git auth -- gitea_admin + its token.
if ! exec_bao kv get kv/flux/gitea-deploy-key >/dev/null 2>&1; then
    FLUX_USER=gitea_admin
    FLUX_PASS=$(cat ~/.rational-reserve/m1-gitea-admin-password)
    exec_bao kv put kv/flux/gitea-deploy-key username="$FLUX_USER" password="$FLUX_PASS"
fi

# Demo app secret.
if ! exec_bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null 2>&1; then
    exec_bao kv put kv/apps/hello-m2/dev/example-secret password="demo-$(openssl rand -hex 8)"
fi

echo "openbao seeded"
