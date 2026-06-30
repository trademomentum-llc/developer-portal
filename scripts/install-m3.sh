#!/usr/bin/env bash
#
# install-m3.sh
# M3 Production Multi-Angle Visibility — Script-driven install
#
# Single source of truth for bringing up the M3 observability and visibility stack.
# Must be preceded by a successful ./scripts/preflight-m3.sh.
#
# This script is idempotent where possible (helm upgrade --install).

set -euo pipefail

echo "=== M3 Install — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTEXT="k3d-openchoreo"

if ! kubectl config current-context 2>/dev/null | grep -q "${CONTEXT}"; then
    echo "ERROR: Current kubectl context is not ${CONTEXT}. Run preflight first."
    exit 1
fi

if ! command -v tofu >/dev/null 2>&1; then
    echo "ERROR: OpenTofu (tofu) is required. Install it or use the helm path manually."
    exit 1
fi

echo "1. Adding Helm repositories (idempotent)"
helm repo add signoz https://charts.signoz.io 2>/dev/null || true
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts 2>/dev/null || true
helm repo update

echo "2. Applying M3 observability module via OpenTofu"
cd "${ROOT_DIR}/iac"
export RR_TOFU_GUARD_BYPASS=1
tofu init -reconfigure
tofu apply -auto-approve -target=module.observability

echo
echo "=== M3 Install complete ==="
echo "Run: ./scripts/smoke-m3.sh --cluster   (after port-forwards or NodePorts are ready)"
echo "Then verify Backstage cards and trace flow for hello-m2."
