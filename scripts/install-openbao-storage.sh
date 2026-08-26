#!/usr/bin/env bash
# scripts/install-openbao-storage.sh
#
# G5 migration orchestrator: dev-mode inmem -> Raft on a local-path PVC
# (BAO-STORAGE-DES-001 section 5). Steps:
#   1. FR-4 inverse-proof lane, pre-change (FAIL expected; PASS aborts)
#   2. helm upgrade to scripts/openbao-values.yaml + pod restart
#   3. bootstrap (scripts/bootstrap-openbao-persistent.sh)
#   4. FR-4 inverse-proof lane, post-change (PASS required)
#   5. smoke-openbao.sh + smoke-all.sh (serialized)
# Set OPENBAO_STORAGE_SKIP_FULL_SMOKE=1 when invoked from install-m2.sh
# (task 6 owns the smoke run there).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OPENBAO_NS=${OPENBAO_NS:-openbao}
RELEASE=${RELEASE:-openbao}
CHART_VERSION=0.25.6
VALUES_FILE="$ROOT/scripts/openbao-values.yaml"
UNSEAL_SECRET=${UNSEAL_SECRET:-openbao-unseal-key}

info() { printf "\033[1;36m[openbao-storage]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-storage ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

already_persistent() {
    kubectl -n "$OPENBAO_NS" get pvc data-openbao-0 >/dev/null 2>&1 || return 1
    # C9 (BAO-STORAGE-SIM-001 D5): a retained PVC alone does not mean the Raft
    # template is live -- after a rollback the release is dev-mode again while
    # the PVC persists, and the re-run path then fails against the dev backend
    # ("custody root token rejected"). Skip migration steps 1-2 only when the
    # current StatefulSet carries the unseal sidecar (the raft-template
    # marker); otherwise fall through to the full path, whose upgrade
    # re-attaches the retained PVC and whose bootstrap recovers the persisted
    # Raft state.
    kubectl -n "$OPENBAO_NS" get statefulset "$RELEASE" \
        -o jsonpath='{.spec.template.spec.containers[*].name}' 2>/dev/null \
        | tr ' ' '\n' | grep -qx openbao-unseal
}

step_1_inverse_proof() {
    info "Step 1: FR-4 inverse-proof lane (pre-change; FAIL expected)"
    if "$ROOT/scripts/smoke-openbao.sh" --with-restart; then
        fail "restart lane PASSED against dev-mode storage; the check cannot fail and is unverified (inverse-proof convention). Investigate before proceeding."
    fi
    info "Step 1 recorded: lane FAILED as expected (inmem data loss on pod deletion)"
}

step_2_upgrade() {
    info "Step 2: helm upgrade to Raft values (chart $CHART_VERSION)"
    helm repo add openbao https://openbao.github.io/openbao-helm >/dev/null 2>&1 || true
    # The unseal Secret must exist before the new pod template starts: the
    # sidecar mounts it as a volume, and a missing Secret blocks container
    # creation. Placeholder value; the bootstrap (step 3) writes the real key
    # and unseals directly, so first boot does not depend on secret-volume
    # refresh.
    if ! kubectl -n "$OPENBAO_NS" get secret "$UNSEAL_SECRET" >/dev/null 2>&1; then
        kubectl -n "$OPENBAO_NS" create secret generic "$UNSEAL_SECRET" \
            --from-literal=key=PENDING-BOOTSTRAP >/dev/null
    fi
    # volumeClaimTemplates is immutable on a StatefulSet: the Raft template
    # cannot be patched onto the existing StatefulSet. Orphan-delete it (the
    # pod keeps running); helm upgrade then recreates the StatefulSet with
    # the data volume. (Kubernetes API behavior; verified constraint, see
    # section 11.)
    kubectl -n "$OPENBAO_NS" delete statefulset "$RELEASE" --cascade=orphan --ignore-not-found >/dev/null
    # No --reuse-values: scripts/openbao-values.yaml is the complete set, so
    # dev.devRootToken and postStart return to chart defaults (removed).
    helm upgrade "$RELEASE" openbao/openbao \
        --version "$CHART_VERSION" \
        --namespace "$OPENBAO_NS" \
        --values "$VALUES_FILE"
    # updateStrategyType OnDelete: the running pod still has the old template
    # until it is deleted.
    kubectl -n "$OPENBAO_NS" delete pod openbao-0 --ignore-not-found >/dev/null
    info "waiting for openbao-0 to run with the Raft template"
    # C7 (BAO-STORAGE-SIM-001 D4): a bare kubectl wait errors NotFound when
    # the recreated pod does not exist yet and set -e kills the orchestrator.
    # Poll within the same 180s budget instead.
    local wstart=$SECONDS
    until kubectl -n "$OPENBAO_NS" wait --for=jsonpath='{.status.phase}'=Running pod/openbao-0 --timeout=15s >/dev/null 2>&1; do
        if [ $((SECONDS - wstart)) -ge 180 ]; then
            fail "openbao-0 did not reach Running within 180s; see rollback (BAO-STORAGE-TECH-001 section 7)"
        fi
        sleep 2
    done
}

step_3_bootstrap() {
    info "Step 3: bootstrap (init/custody/unseal/auth/policies/secrets/seed/handoff)"
    "$ROOT/scripts/bootstrap-openbao-persistent.sh"
}

step_4_inverse_proof_pass() {
    info "Step 4: FR-4 inverse-proof lane (post-change; PASS required)"
    "$ROOT/scripts/smoke-openbao.sh" --with-restart \
        || fail "post-change restart lane FAILED; do not claim acceptance (section 6)"
}

step_5_smokes() {
    if [ "${OPENBAO_STORAGE_SKIP_FULL_SMOKE:-0}" = "1" ]; then
        info "Step 5: deferred to the invoking orchestrator (install-m2 task 6)"
        return 0
    fi
    info "Step 5: serialized smoke suites"
    "$ROOT/scripts/smoke-openbao.sh" || fail "smoke-openbao failed after migration"
    "$ROOT/scripts/smoke-all.sh" || fail "smoke-all failed after migration"
}

main() {
    [ -f "$VALUES_FILE" ] || fail "missing $VALUES_FILE"
    if already_persistent; then
        # Greenfield-after-migration / re-run path (DES-001 section 5 ordering
        # note): no dev-mode phase exists, so steps 1-2 are skipped.
        info "PVC data-openbao-0 already present; skipping migration steps 1-2"
        step_3_bootstrap
        "$ROOT/scripts/smoke-openbao.sh" || fail "smoke-openbao failed"
        info "OpenBao persistent storage verified (re-run path)"
        return 0
    fi
    step_1_inverse_proof
    step_2_upgrade
    step_3_bootstrap
    step_4_inverse_proof_pass
    step_5_smokes
    info "OpenBao persistent storage migration complete."
}

main "$@"
