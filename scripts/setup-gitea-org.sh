#!/usr/bin/env bash
# scripts/setup-gitea-org.sh -- Create Gitea organization and transfer demo repo.
set -euo pipefail

GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")
ORG="trademomentum.net"

# Create the organization
curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST http://localhost:3002/api/v1/orgs \
    -H "Content-Type: application/json" \
    -d "{
      \"username\": \"${ORG}\",
      \"full_name\": \"Trade Momentum\",
      \"description\": \"IDP organization\",
      \"visibility\": \"public\"
    }"

echo ""
echo "Organization '${ORG}' created."

# Transfer demo-service repo to the org
curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST http://localhost:3002/api/v1/repos/gitea_admin/demo-service/transfer \
    -H "Content-Type: application/json" \
    -d "{\"new_owner\": \"${ORG}\"}"

echo ""
echo "demo-service transferred to ${ORG}."
