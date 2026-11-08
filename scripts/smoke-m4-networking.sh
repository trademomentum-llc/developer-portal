#!/usr/bin/env bash
# scripts/smoke-m4-networking.sh
# Verify Envoy Gateway routes reach the expected platform services over HTTP
# (port 80) and HTTPS (port 443, FR-09). Port 80 keeps serving plain HTTP with
# no redirect in Wave 0, so both sections assert 200/302 independently.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"

CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"
ENVOY_NS="envoy-gateway"
LOCAL_PORT="${ENVOY_LOCAL_PORT:-38080}"
LOCAL_PORT_HTTPS="${ENVOY_LOCAL_PORT_HTTPS:-38443}"
HOSTS=(gitea.local signoz.local opencost.local)
FAILED=0
PF_PID=""
PF_HTTPS_PID=""

info() { echo "[smoke-m4-networking] $*"; }

# Single EXIT trap: kills any port-forwards that were started and emits the
# FR-34 JSON record (no-trap mode; see scripts/lib/smoke-json.sh).
cleanup() {
    if [ -n "${PF_PID}" ]; then kill "${PF_PID}" 2>/dev/null || true; fi
    if [ -n "${PF_HTTPS_PID}" ]; then kill "${PF_HTTPS_PID}" 2>/dev/null || true; fi
    if [ "${SMOKE_JSON_FAILED}" -eq 0 ] && [ "${SMOKE_JSON_PASSED}" -eq 0 ]; then
        SMOKE_JSON_FAILED=1
    fi
    smoke_json_emit || true
}
smoke_json_begin m4-networking no-trap
trap cleanup EXIT

# Find the Envoy proxy service name created for our Gateway.
SERVICE_NAME=$(kubectl --context "${CONTEXT}" -n "${ENVOY_NS}" get svc \
  -l gateway.envoyproxy.io/owning-gateway-name=eg \
  -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

if [ -z "${SERVICE_NAME}" ]; then
    echo "FAIL: Could not find Envoy Gateway proxy service" >&2
    smoke_json_count fail
    exit 1
fi
smoke_json_count pass

info "Forwarding ${ENVOY_NS}/${SERVICE_NAME} to localhost:${LOCAL_PORT}"
kubectl --context "${CONTEXT}" -n "${ENVOY_NS}" port-forward "svc/${SERVICE_NAME}" "${LOCAL_PORT}:80" >/tmp/envoy-portforward.log 2>&1 &
PF_PID=$!

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
        smoke_json_count pass
    else
        echo "FAIL: ${host} -> ${status}" >&2
        smoke_json_count fail
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

# Wait for the HTTPS port-forward to accept TLS connections. Envoy builds one
# SNI-strict filter chain per hostname and has no catch-all, so every probe
# must send a real listener hostname via --resolve; an SNI of "localhost"
# matches no filter chain and Envoy RSTs the connection, which also kills the
# kubectl port-forward tunnel.
for i in $(seq 1 30); do
    if curl -ksS -o /dev/null --resolve "${HOSTS[0]}:${LOCAL_PORT_HTTPS}:127.0.0.1" "https://${HOSTS[0]}:${LOCAL_PORT_HTTPS}/" 2>/dev/null; then
        break
    fi
    sleep 1
done

for host in "${HOSTS[@]}"; do
    status=$(curl -sk -o /dev/null -w "%{http_code}" --resolve "${host}:${LOCAL_PORT_HTTPS}:127.0.0.1" "https://${host}:${LOCAL_PORT_HTTPS}/" 2>/dev/null || true)
    if [ "${status}" = "200" ] || [ "${status}" = "302" ]; then
        echo "PASS: https://${host} -> ${status}"
        smoke_json_count pass
    else
        echo "FAIL: https://${host} -> ${status}" >&2
        smoke_json_count fail
        FAILED=1
    fi
done

if [ "${FAILED}" -ne 0 ]; then
    exit 1
fi

info "M4 networking smoke test passed."
