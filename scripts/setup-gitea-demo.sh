#!/usr/bin/env bash
# scripts/setup-gitea-demo.sh -- Create demo repo with catalog-info.yaml.
set -euo pipefail

GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")

# Create repo
curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST http://localhost:3002/api/v1/user/repos \
    -H "Content-Type: application/json" \
    -d '{
      "name": "demo-service",
      "description": "M1 demo component for Backstage catalog discovery",
      "private": false,
      "auto_init": true,
      "default_branch": "main"
    }'

echo ""
echo "Demo repo created."

# Push catalog-info.yaml
CATALOG_YAML='apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: demo-service
  description: M1 smoke-test component
spec:
  type: service
  lifecycle: experimental
  owner: gitea_admin'

CATALOG_B64=$(printf '%s' "$CATALOG_YAML" | base64)

curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST "http://localhost:3002/api/v1/repos/gitea_admin/demo-service/contents/catalog-info.yaml" \
    -H "Content-Type: application/json" \
    -d "{\"message\": \"M1 seed catalog entry\", \"content\": \"${CATALOG_B64}\", \"branch\": \"main\"}"

echo ""
echo "catalog-info.yaml pushed to demo-service repo."
