#!/usr/bin/env bash
# scripts/update-local-hosts.sh
# Discover the Envoy Gateway NodePort endpoint and print /etc/hosts entries.
set -euo pipefail

RUNTIME_DIR="${HOME}/.rational-reserve"
CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"

echo "# Discovering Envoy Gateway NodePort..."
NODE_PORT=$(kubectl --context "${CONTEXT}" -n envoy-gateway get svc \
  -l gateway.envoyproxy.io/owning-gateway-name=eg \
  -o jsonpath='{.items[0].spec.ports[?(@.port==80)].nodePort}' 2>/dev/null || true)

if [ -z "${NODE_PORT}" ]; then
    echo "ERROR: Could not find Envoy Gateway NodePort" >&2
    exit 1
fi

echo "# Envoy Gateway is reachable at http://127.0.0.1:${NODE_PORT}"
echo "# Because /etc/hosts cannot carry a port, access services as:"
echo "#   http://gitea.local:${NODE_PORT}"
echo "#   http://signoz.local:${NODE_PORT}"
echo "#   http://opencost.local:${NODE_PORT}"
echo ""
echo "# Add this line to /etc/hosts (sudo required):"
echo "127.0.0.1 gitea.local signoz.local opencost.local"

mkdir -p "${RUNTIME_DIR}"
echo "http://127.0.0.1:${NODE_PORT}" > "${RUNTIME_DIR}/envoy-gateway-url"
