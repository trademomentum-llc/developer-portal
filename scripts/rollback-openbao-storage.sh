#!/usr/bin/env bash
# scripts/rollback-openbao-storage.sh
#
# Roll back to the pre-G5 dev-mode release (BAO-STORAGE-DES-001 section 9).
# Preserves the data-openbao-0 PVC, the openbao-unseal-key Secret, the host
# custody files, and all snapshots. Usage:
#   ./scripts/rollback-openbao-storage.sh [revision]   # default 1
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OPENBAO_NS=${OPENBAO_NS:-openbao}
RELEASE=${RELEASE:-openbao}
REVISION=${1:-1}
ESO_NS=${ESO_NS:-external-secrets}
ESO_SECRET=${ESO_SECRET:-openbao-root-token}
# C8 v2 (BAO-STORAGE-SIM-001 D2): wait budget overridable; default 120s kept.
ROLLBACK_WAIT=${OPENBAO_ROLLBACK_WAIT_SECONDS:-120}

info() { printf "\033[1;36m[openbao-rollback]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-rollback ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

info "orphaning the Raft StatefulSet (volumeClaimTemplates is immutable; pod keeps running)"
kubectl -n "$OPENBAO_NS" delete statefulset "$RELEASE" --cascade=orphan --ignore-not-found >/dev/null

info "helm rollback $RELEASE to revision $REVISION (dev mode + postStart)"
helm rollback "$RELEASE" "$REVISION" --namespace "$OPENBAO_NS"

info "restarting the pod onto the dev-mode template (updateStrategy OnDelete)"
kubectl -n "$OPENBAO_NS" delete pod openbao-0 --ignore-not-found >/dev/null
# C8 (BAO-STORAGE-SIM-001 D4, same class as C7): a bare kubectl wait races
# pod recreation (NotFound kills the script under set -e). Poll within the
# same budget instead.
wstart=$SECONDS
until kubectl -n "$OPENBAO_NS" wait --for=condition=Ready pod/openbao-0 --timeout=15s >/dev/null 2>&1; do
    if [ $((SECONDS - wstart)) -ge "$ROLLBACK_WAIT" ]; then
        fail "openbao-0 did not become Ready on the rolled-back template"
    fi
    sleep 2
done

info "restoring literal dev root token for ExternalSecrets"
kubectl -n "$ESO_NS" create secret generic "$ESO_SECRET" \
    --from-literal=token=root --dry-run=client -o yaml \
    | kubectl apply -f - >/dev/null

info "reseeding the fresh inmem backend (the seeder's pre-G5 recovery role)"
OPENBAO_TOKEN=root "$ROOT/scripts/seed-openbao-m2-paths.sh"

info "rollback complete. PVC data-openbao-0, secret openbao-unseal-key,"
info "custody ~/.rational-reserve/openbao/, and snapshots are preserved;"
info "a re-run of install-openbao-storage.sh resumes from persisted Raft state."
