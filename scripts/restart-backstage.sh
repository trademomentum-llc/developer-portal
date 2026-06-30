#!/usr/bin/env bash
# scripts/restart-backstage.sh -- Stop and start the Backstage dev server.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/stop-backstage.sh"
sleep 2
exec "$SCRIPT_DIR/start-backstage.sh"
