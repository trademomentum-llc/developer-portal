#!/usr/bin/env bash
# scripts/smoke-openbao.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin openbao

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$OPENBAO_TOKEN" bao "$@"
}

for key in \
    kv/gitea/runners/token \
    kv/flux/gitea-deploy-key \
    secret/apps/hello-m2/dev/example-secret \
    kv/apps/hello-m2/dev/example-secret; do
    if exec_bao kv get "$key" >/dev/null; then
        smoke_json_count pass
    else
        smoke_json_count fail
        exit 1
    fi
done
echo "PASS"
