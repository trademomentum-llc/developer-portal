#!/usr/bin/env bash
#
# teardown-m3.sh
# M3 Production Multi-Angle Visibility — Clean, safe teardown
#
# Reverses everything done by install-m3.sh. Safe to run multiple times.

set -euo pipefail

echo "=== M3 Teardown — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTEXT="k3d-openchoreo"

if ! kubectl config current-context 2>/dev/null | grep -q "${CONTEXT}"; then
    echo "WARNING: Current kubectl context is not ${CONTEXT}."
fi

echo "Destroying M3 observability module via OpenTofu"
cd "${ROOT_DIR}/iac"
export RR_TOFU_GUARD_BYPASS=1
tofu destroy -auto-approve -target=module.observability

echo "M3 resources removed (namespaces signoz + otel-system)."
echo "Run preflight-m3.sh to confirm cluster state before any future install."
