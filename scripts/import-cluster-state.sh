#!/usr/bin/env bash
#
# import-cluster-state.sh
# State-heal: adopt live Helm releases into the tofu kubernetes-backend state.
#
# Background (drift finding, 2026-08-18, Wave-0 acceptance):
#   The M3 observability releases (signoz, otel-collector) were installed
#   raw-helm on 2026-06-03/2026-06-30 and the M3/M4 tofu modules were never
#   applied over them (a "3-to-add" plan for module.observability was on
#   record). install-m3.sh therefore failed with:
#     "cannot re-use a name that is still in use"
#   This script imports the live releases into state so the lifecycle
#   scripts (install-m3.sh et al.) manage them going forward.
#
# Scope (per the 2026-08-18 drift matrix):
#   IMPORT  module.observability.helm_release.signoz         (signoz/signoz)
#   IMPORT  module.observability.helm_release.otel_collector (otel-system/otel-collector)
#   NONE    module.cost.*         -- already in state (applied 2026-06-30)
#   NONE    module.networking.*   -- helm_release + gateway manifests already
#           in state; tls.tf issuers/Certificates are Wave-0 adds that do not
#           exist live yet, so there is nothing to import (apply creates them).
#   SKIP    module.observability.null_resource.patch_signoz_collector --
#           bookkeeping-only local-exec with no importable cluster object; its
#           patch effect is already live (signoz-otel-collector args verified
#           2026-08-18). Importing it would only turn a clean "add" into a
#           no-benefit "replace". The next install-m3.sh apply creates it and
#           re-runs the idempotent kubectl patch.
#
# Idempotent: each (address, ID) pair is skipped when `tofu state show`
# already succeeds. Safe to re-run.
#
# Import ID format: the hashicorp/helm provider (2.17.0, .terraform.lock.hcl)
# imports a release as "<namespace>/<name>".

set -euo pipefail

echo "=== Import cluster state - $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTEXT="k3d-openchoreo"

if ! kubectl config current-context 2>/dev/null | grep -q "${CONTEXT}"; then
    echo "ERROR: Current kubectl context is not ${CONTEXT}. Aborting."
    exit 1
fi

cd "${ROOT_DIR}/iac"
export RR_TOFU_GUARD_BYPASS=1

tofu init -reconfigure >/dev/null

import_if_missing() {
    local address="$1"
    local id="$2"
    if tofu state show "${address}" >/dev/null 2>&1; then
        echo "SKIP  ${address} (already in state)"
    else
        echo "IMPORT ${address} <- ${id}"
        tofu import "${address}" "${id}"
    fi
}

import_if_missing "module.observability.helm_release.signoz" "signoz/signoz"
import_if_missing "module.observability.helm_release.otel_collector" "otel-system/otel-collector"

echo
echo "=== Import complete. Verify with: cd iac && tofu state list ==="
