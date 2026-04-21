#!/usr/bin/env bash
# Seed the three openbao kv v2 paths M2 relies on. Idempotent.
set -euo pipefail

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- bao "$@"
}

# Runner registration token -- operator must fetch from Gitea and paste
if ! exec_bao kv get kv/gitea/runners/token >/dev/null 2>&1; then
    printf "Gitea runner registration token: "
    read -rs TOKEN
    echo
    exec_bao kv put kv/gitea/runners/token token="$TOKEN"
fi

# Flux git auth -- gitea_admin + its token
if ! exec_bao kv get kv/flux/gitea-deploy-key >/dev/null 2>&1; then
    FLUX_USER=gitea_admin
    FLUX_PASS=$(cat ~/.rational-reserve/m1-gitea-admin-password)
    exec_bao kv put kv/flux/gitea-deploy-key username="$FLUX_USER" password="$FLUX_PASS"
fi

# Demo app secret
if ! exec_bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null 2>&1; then
    exec_bao kv put kv/apps/hello-m2/dev/example-secret password="demo-$(openssl rand -hex 8)"
fi

echo "openbao seeded"
