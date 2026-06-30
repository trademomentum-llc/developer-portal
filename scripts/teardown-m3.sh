#!/usr/bin/env bash
#
# teardown-m3.sh
# M3 Production Multi-Angle Visibility — Clean, safe teardown
#
# Reverses everything done by install-m3.sh. Safe to run multiple times.

set -euo pipefail

echo "=== M3 Teardown — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

CONTEXT="k3d-openchoreo"

helm uninstall signoz -n signoz 2>/dev/null || true
helm uninstall otel-collector -n otel-system 2>/dev/null || true

kubectl --context "${CONTEXT}" delete namespace signoz otel-system --ignore-not-found --timeout=60s || true

echo "M3 resources removed (namespaces signoz + otel-system)."
echo "Run preflight-m3.sh to confirm cluster state before any future install."