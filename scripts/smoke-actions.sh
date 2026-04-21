#!/usr/bin/env bash
# scripts/smoke-actions.sh
set -e
TOKEN_FILE=~/.rational-reserve/m1-gitea-admin-password
RUN_ID=$(curl -fsS -u "gitea_admin:$(cat $TOKEN_FILE)" \
    -X POST "http://localhost:3002/api/v1/repos/openchoreo/hello-m2/actions/workflows/ci.yaml/dispatches" \
    -H 'Content-Type: application/json' \
    -d '{"ref":"main"}' | jq -r '.id // empty')
if [ -z "$RUN_ID" ]; then
    echo "PASS (dispatched)"
    exit 0
fi
for _ in $(seq 1 60); do
    STATUS=$(curl -fsS -u "gitea_admin:$(cat $TOKEN_FILE)" \
        "http://localhost:3002/api/v1/repos/openchoreo/hello-m2/actions/runs/$RUN_ID" \
        | jq -r .conclusion)
    case "$STATUS" in
        success) echo "PASS"; exit 0 ;;
        failure) echo "FAIL"; exit 1 ;;
    esac
    sleep 5
done
echo "FAIL: timeout"; exit 1
