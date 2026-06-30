#!/usr/bin/env bash
# scripts/teardown-m4.sh
# Tear down the M4 cost visibility plane.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../iac"

tofu destroy -target=module.cost -auto-approve

echo "M4 cost plane torn down."
