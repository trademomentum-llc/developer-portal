#!/usr/bin/env bash
# scripts/smoke-m4-networking.sh
# Verify Envoy Gateway routes reach the expected platform services.
set -euo pipefail

CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"
ENVOY_NS="envoy-gateway"
LOCAL_PORT="${ENVOY_LOCAL_PORT:-38080}"
HOSTS=(gitea.local signoz.local opencost.local)
FAILED=0

info() { echo "[smoke-m4-networking] $*"; }

# Find the Envoy proxy service name created for our Gateway.
SERVICE_NAME=$(kubectl --context "${CONTEXT}" -n "${ENVOY_NS}" get svc \
  -l gateway.envoyproxy.io/owning-gateway-name=eg \
  -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

if [ -z "${SERVICE_NAME}" ]; then
    echo "FAIL: Could not find Envoy Gateway proxy service" >&2
    exit 1
fi

info "Forwarding ${ENVOY_NS}/${SERVICE_NAME} to localhost:${LOCAL_PORT}"
kubectl --context "${CONTEXT}" -n "${ENVOY_NS}" port-forward "svc/${SERVICE_NAME}" "${LOCAL_PORT}:80" >/tmp/envoy-portforward.log 2>&1 &
PF_PID=$!
trap 'kill "${PF_PID}" 2>/dev/null || true' EXIT

# Wait for port-forward to accept connections.
for i in $(seq 1 30); do
    if curl -fsS -o /dev/null "http://localhost:${LOCAL_PORT}/" 2>/dev/null; then
        break
    fi
    sleep 1
done

for host in "${HOSTS[@]}"; do
    status=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: ${host}" "http://localhost:${LOCAL_PORT}/" 2>/dev/null || true)
    if [ "${status}" = "200" ] || [ "${status}" = "302" ]; then
        echo "PASS: ${host} -> ${status}"
    else
        echo "FAIL: ${host} -> ${status}" >&2
        FAILED=1
    fi
done

if [ "${FAILED}" -ne 0 ]; then
    exit 1
fi

info "M4 networking smoke test passed."
