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

echo "1. Adding Helm repositories (idempotent)"
helm repo add signoz https://charts.signoz.io 2>/dev/null || true
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts 2>/dev/null || true
helm repo update

echo "2. Creating M3 namespaces (idempotent)"
kubectl --context "${CONTEXT}" create namespace signoz --dry-run=client -o yaml | kubectl apply -f -
kubectl --context "${CONTEXT}" create namespace otel-system --dry-run=client -o yaml | kubectl apply -f -

echo "3. Installing / upgrading SigNoz (using local values)"
helm upgrade --install signoz signoz/signoz \
    --namespace signoz \
    --create-namespace \
    -f "${ROOT_DIR}/observability/signoz/values.local.yaml" \
    --wait --timeout 10m

echo "4. Installing / upgrading standalone OTEL collector"
helm upgrade --install otel-collector open-telemetry/opentelemetry-collector \
    --namespace otel-system \
    -f "${ROOT_DIR}/observability/otel/collector-values.local.yaml" \
    --wait --timeout 5m

echo "5. (Placeholder) Applying any additional M3 dashboards or ConfigMaps"
# kubectl apply -f observability/dashboards/...

echo
echo "=== M3 Install complete ==="
echo "Run: ./scripts/smoke-m3.sh --cluster   (after port-forwards or NodePorts are ready)"
echo "Then verify Backstage cards and trace flow for hello-m2."