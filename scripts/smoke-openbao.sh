#!/usr/bin/env bash
# scripts/smoke-openbao.sh
set -euo pipefail

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$OPENBAO_TOKEN" bao "$@"
}

exec_bao kv get kv/gitea/runners/token >/dev/null
exec_bao kv get kv/flux/gitea-deploy-key >/dev/null
exec_bao kv get secret/apps/hello-m2/dev/example-secret >/dev/null
exec_bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null
echo "PASS"
