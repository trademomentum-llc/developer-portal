#!/usr/bin/env bash
# scripts/setup-gitea-secret.sh -- Create gitea namespace and admin secret.
set -euo pipefail

mkdir -p "$HOME/.rational-reserve"

kubectl create namespace gitea 2>/dev/null || true

GITEA_ADMIN_PASSWORD=$(openssl rand -base64 24 | tr -d '=+/' | head -c 24)
printf '%s' "$GITEA_ADMIN_PASSWORD" > "$HOME/.rational-reserve/m1-gitea-admin-password"
chmod 600 "$HOME/.rational-reserve/m1-gitea-admin-password"

kubectl create secret generic gitea-admin-secret \
    --namespace gitea \
    --from-literal=username=gitea_admin \
    --from-literal=password="$GITEA_ADMIN_PASSWORD" \
    --from-literal=email=admin@local.dev

echo "Gitea admin secret created. Password saved to ~/.rational-reserve/m1-gitea-admin-password"
