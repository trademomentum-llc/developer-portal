# scripts/smoke-openbao.sh
#!/usr/bin/env bash
set -e
kubectl -n openbao exec openbao-0 -- bao kv get kv/gitea/runners/token >/dev/null
kubectl -n openbao exec openbao-0 -- bao kv get kv/flux/gitea-deploy-key >/dev/null
kubectl -n openbao exec openbao-0 -- bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null
echo "PASS"
