#!/usr/bin/env bash
# scripts/teardown-backstage-production.sh
# Tear down PostgreSQL for Backstage production mode.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../iac"

export RR_TOFU_GUARD_BYPASS=1
tofu destroy -target=module.postgres -auto-approve

echo "Backstage production PostgreSQL torn down."
