#!/usr/bin/env bash
# scripts/install-m4.sh
# Deploy the M4 cost visibility plane (Prometheus + OpenCost).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../iac"

tofu init -reconfigure
tofu apply -target=module.cost -auto-approve

echo "M4 cost plane installed. Use scripts/smoke-m4.sh to verify."
