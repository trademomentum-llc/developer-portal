#!/usr/bin/env bash
# scripts/smoke-m4-networking.sh
# Verify Envoy Gateway routes reach the expected platform services over HTTP
# (port 80) and HTTPS (port 443, FR-09). Port 80 keeps serving plain HTTP with
# no redirect in Wave 0, so both sections assert 200/302 independently.
set -euo pipefail

CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"
ENVOY_NS="envoy-gateway"
LOCAL_PORT="${ENVOY_LOCAL_PORT:-38080}"
LOCAL_PORT_HTTPS="${ENVOY_LOCAL_PORT_HTTPS:-38443}"
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

# FR-09: HTTPS section. Second port-forward to listener port 443; the local CA
# is not host-trusted by default (macOS trust is a manual user step, spec 9.3),
# so curl uses -k. An untrusted client without -k must fail TLS validation,
# which is the FR-09 negative test.
info "Forwarding ${ENVOY_NS}/${SERVICE_NAME} to localhost:${LOCAL_PORT_HTTPS} (443)"
kubectl --context "${CONTEXT}" -n "${ENVOY_NS}" port-forward "svc/${SERVICE_NAME}" "${LOCAL_PORT_HTTPS}:443" >/tmp/envoy-portforward-https.log 2>&1 &
PF_HTTPS_PID=$!
trap 'kill "${PF_PID}" "${PF_HTTPS_PID}" 2>/dev/null || true' EXIT

# Wait for the HTTPS port-forward to accept TLS connections. No -f here: an
# unknown Host header yields 404, which still proves the listener is up.
for i in $(seq 1 30); do
    if curl -ksS -o /dev/null "https://localhost:${LOCAL_PORT_HTTPS}/" 2>/dev/null; then
        break
    fi
    sleep 1
done

for host in "${HOSTS[@]}"; do
    status=$(curl -sk -o /dev/null -w "%{http_code}" -H "Host: ${host}" "https://localhost:${LOCAL_PORT_HTTPS}/" 2>/dev/null || true)
    if [ "${status}" = "200" ] || [ "${status}" = "302" ]; then
        echo "PASS: https://${host} -> ${status}"
    else
        echo "FAIL: https://${host} -> ${status}" >&2
        FAILED=1
    fi
done

if [ "${FAILED}" -ne 0 ]; then
    exit 1
fi

info "M4 networking smoke test passed."
