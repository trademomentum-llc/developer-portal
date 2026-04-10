#!/usr/bin/env bash
# scripts/setup-devportal-repo.sh -- Create developer-portal repo in Gitea org.
set -euo pipefail

GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")
ORG="trademomentum.net"

curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST http://localhost:3002/api/v1/orgs/${ORG}/repos \
    -H "Content-Type: application/json" \
    -d '{
      "name": "developer-portal",
      "description": "Internal Developer Platform -- M1 Substrate",
      "private": false,
      "default_branch": "main"
    }'

echo ""
echo "developer-portal repo created in ${ORG}."
