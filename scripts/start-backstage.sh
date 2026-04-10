#!/usr/bin/env bash
# scripts/start-backstage.sh -- Start Backstage dev server.
set -euo pipefail

cd /Users/nnos/Projects/developer-portal/backstage
GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")
export GITEA_ADMIN_PASSWORD

yarn start &
BACKSTAGE_PID=$!
echo "$BACKSTAGE_PID" > "$HOME/.rational-reserve/m1-backstage-dev.pid"
echo "Backstage starting with PID $BACKSTAGE_PID"
echo "Waiting for http://localhost:3000 ..."

for i in $(seq 1 90); do
    if curl -s -o /dev/null http://localhost:3000; then
        echo "Backstage is up at http://localhost:3000"
        exit 0
    fi
    sleep 2
done

echo "Backstage did not come up within 3 minutes"
exit 1
