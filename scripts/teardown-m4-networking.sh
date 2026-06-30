#!/usr/bin/env bash
# scripts/teardown-m4-networking.sh
# Tear down the M4 networking plane (Envoy Gateway ingress).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../iac"

export RR_TOFU_GUARD_BYPASS=1
tofu init -reconfigure
tofu destroy -target=module.networking -auto-approve

echo "M4 networking plane torn down."
