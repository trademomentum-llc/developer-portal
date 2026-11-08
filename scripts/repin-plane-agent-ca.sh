#!/usr/bin/env bash
#
# repin-plane-agent-ca.sh
#
# The OpenChoreo cluster-plane agents (data/observability/workflow) use
# self-signed client certs (CN=default, ~90-day lifetime) that cert-manager
# re-issues on renewal. The ClusterDataPlane / ClusterObservabilityPlane /
# ClusterWorkflowPlane CRs pin the agent cert as spec.clusterAgent.clientCA.value
# at install time, so every renewal silently breaks the gateway channel
# (websocket bad handshake; first observed 2026-06-30, repaired 2026-08-20).
#
# This script re-pins each CR to the current cert from the plane's
# cluster-agent-tls Secret. It is idempotent: no drift, no patch.
#
# Modes:
#   (default)   re-pin any CR whose pinned CA differs from the live cert
#   --check     exit 1 if any plane is drifted (no mutation; for smoke suites)
#
# The structural fix (real CA or installer-side re-pin) remains upstream debt;
# until then this script should run on any plane-agent cert renewal, and the
# smoke check fails loudly when the pins go stale.

set -euo pipefail

CONTEXT="k3d-openchoreo"
MODE="apply"
[ "${1:-}" = "--check" ] && MODE="check"

DRIFTED=0

while read -r NS KIND; do

    LIVE=$(kubectl --context "${CONTEXT}" -n "${NS}" get secret cluster-agent-tls \
        -o jsonpath='{.data.tls\.crt}' | base64 -d)
    PINNED=$(kubectl --context "${CONTEXT}" get "${KIND}" default \
        -o jsonpath='{.spec.clusterAgent.clientCA.value}')

    LIVE_FP=$(printf '%s' "${LIVE}" | openssl x509 -noout -fingerprint -sha256)
    PINNED_FP=$(printf '%s' "${PINNED}" | openssl x509 -noout -fingerprint -sha256 2>/dev/null || echo "unparseable")

    if [ "${LIVE_FP}" = "${PINNED_FP}" ]; then
        echo "OK: ${KIND}/default pin matches live agent cert (${LIVE_FP##*=})"
        continue
    fi

    DRIFTED=1
    echo "DRIFT: ${KIND}/default pinned ${PINNED_FP##*=} != live ${LIVE_FP##*=}"

    if [ "${MODE}" = "apply" ]; then
        PATCH=$(python3 -c "import json,sys; print(json.dumps({'spec':{'clusterAgent':{'clientCA':{'value': sys.argv[1]}}}}))" "${LIVE}")
        kubectl --context "${CONTEXT}" patch "${KIND}" default --type merge -p "${PATCH}"
        echo "REPINNED: ${KIND}/default -> ${LIVE_FP##*=}"
    fi
done < <(printf '%s\n' \
    "openchoreo-data-plane clusterdataplane" \
    "openchoreo-observability-plane clusterobservabilityplane" \
    "openchoreo-workflow-plane clusterworkflowplane")

if [ "${MODE}" = "check" ] && [ "${DRIFTED}" = "1" ]; then
    echo "FAIL: plane agent clientCA drift detected (run scripts/repin-plane-agent-ca.sh)" >&2
    exit 1
fi

[ "${MODE}" = "check" ] && echo "PASS: all plane agent clientCA pins match live certs"
exit 0
