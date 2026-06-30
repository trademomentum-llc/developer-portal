#!/usr/bin/env bash
#
# preflight-m3.sh
# M3 Production Multi-Angle Visibility — Read-only inventory & readiness check
#
# This script must remain read-only. It never mutates cluster state.
# Run it before any install-m3.sh or tofu apply for observability resources.
#
# It implements the contract from docs/specs/m3-observability/technical-specification.md
# and supports the 2026-05-28 Production Multi-Angle Visibility Requirements.

set -euo pipefail

echo "=== M3 Preflight — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="
echo "Context: k3d-openchoreo (Colima)"
echo

echo "1. Cluster & Node Basics"
kubectl --context k3d-openchoreo version --short 2>/dev/null || kubectl --context k3d-openchoreo version
echo
kubectl --context k3d-openchoreo get nodes -o wide
echo

echo "2. Storage Classes"
kubectl --context k3d-openchoreo get storageclass
echo

echo "3. All Namespaces (sorted)"
kubectl --context k3d-openchoreo get ns --sort-by=.metadata.name
echo

echo "4. Existing OpenChoreo Observability Plane (do not reuse)"
kubectl --context k3d-openchoreo -n openchoreo-observability-plane get all 2>/dev/null || echo "Namespace or resources not found (expected at early M3)"
echo

echo "5. Non-Running / Problem Pods (cluster-wide)"
kubectl --context k3d-openchoreo get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded 2>/dev/null || echo "No non-running pods or query not supported"
echo

echo "6. Resource Pressure (top nodes & pods)"
if command -v kubectl >/dev/null && kubectl --context k3d-openchoreo top nodes >/dev/null 2>&1; then
    kubectl --context k3d-openchoreo top nodes
    echo
    kubectl --context k3d-openchoreo top pods -A --sort-by=cpu | head -20
else
    echo "kubectl top not available (metrics-server missing or not ready). This is expected in early local k3d."
fi
echo

echo "7. Colima / Host Resource View (best effort)"
if command -v colima >/dev/null 2>&1; then
    colima status 2>/dev/null || echo "colima status failed or Colima not running"
else
    echo "colima CLI not found on host PATH"
fi
echo

echo "8. OpenChoreo Plane Namespaces Summary"
kubectl --context k3d-openchoreo get ns | grep -E 'openchoreo-(control|data|observability|workflow)-plane' || echo "OpenChoreo planes not yet enumerated or different naming"
echo

echo "9. Namespace Predictor Verification (deterministic, Option C)"
echo "   Using hello-m2 catalog annotations (control=default, project=default, env=dev)"
echo "   Canonical vector must produce: dp-default-default-dev-3a594436"
if command -v go >/dev/null 2>&1; then
    PREDICTOR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/tools/namespace-predictor"
    if [[ -f "$PREDICTOR_DIR/main.go" ]]; then
        COMPUTED_NS=$(go run "$PREDICTOR_DIR/main.go" default default dev 2>/dev/null || echo "PREDICTOR_FAILED")
        echo "   Go binary output for (default, default, dev): $COMPUTED_NS"
        if [[ "$COMPUTED_NS" == "dp-default-default-dev-3a594436" ]]; then
            echo "    Predictor matches canonical test vector (deterministic contract verified)"
        else
            echo "    Predictor output mismatch — investigate immediately (data integrity risk)"
        fi
    else
        echo "   Predictor source not found at $PREDICTOR_DIR/main.go (expected after Option C work)"
    fi
else
    echo "   'go' not on PATH — cannot execute namespace-predictor binary. Install Go or use pre-built binary for full preflight."
fi
echo "   (Future enhancement: also compute expected ns for every catalog entity and assert pods exist in those namespaces after install)"
echo

echo "=== Preflight complete. Review output before proceeding to install-m3.sh or observability/ IaC ==="
echo "Record storageClass, resource headroom, and any port conflicts in the M3 technical spec before pinning versions."
echo "The namespace predictor output above is now the authoritative value used by Backstage cards and all M3 scripts."