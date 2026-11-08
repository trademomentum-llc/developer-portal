#!/usr/bin/env bash
# scripts/smoke-openbao.sh
#
# Default lane: presence of the four M2 keys (fast; part of smoke-m2).
# --with-restart: FR-4 inverse-proof lane (BAO-STORAGE-DES-001 D-08) --
# records exact values, deletes openbao-0, requires Ready (which proves the
# unseal sidecar acted, because the readiness probe fails while sealed),
# asserts value equality with no reseed. Heavy and serialized: standalone or
# via smoke-all.sh --with-openbao-restart, never concurrent with other
# smokes. Expected FAIL against dev-mode inmem storage; expected PASS
# against the Raft backend.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin openbao

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}

# Post-G5 default: the custody root token. OPENBAO_TOKEN overrides; the
# literal "root" fallback only matches the pre-migration dev-mode backend.
if [ -z "${OPENBAO_TOKEN:-}" ] && [ -r "$CUSTODY_DIR/root-token" ]; then
    OPENBAO_TOKEN=$(cat "$CUSTODY_DIR/root-token")
fi
OPENBAO_TOKEN=${OPENBAO_TOKEN:-root}

WITH_RESTART=0
for arg in ${SMOKE_JSON_ARGS[@]+"${SMOKE_JSON_ARGS[@]}"}; do
    case "$arg" in
        --with-restart) WITH_RESTART=1 ;;
        *) echo "smoke-openbao: unknown argument: $arg" >&2; exit 2 ;;
    esac
done

exec_bao() {
    kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$OPENBAO_TOKEN" bao "$@"
}

M2_KEYS="
kv/gitea/runners/token
kv/flux/gitea-deploy-key
secret/apps/hello-m2/dev/example-secret
kv/apps/hello-m2/dev/example-secret
"

check_presence() {
    local rc=0 key
    for key in $M2_KEYS; do
        if exec_bao kv get "$key" >/dev/null; then
            smoke_json_count pass
        else
            smoke_json_count fail
            rc=1
        fi
    done
    if [ "$rc" -eq 0 ]; then echo "PASS"; else echo "FAIL"; fi
    return "$rc"
}

check_restart() {
    local keys=() before=() after="" i=0 key start elapsed
    local restart_wait=${OPENBAO_RESTART_WAIT_SECONDS:-120}
    for key in $M2_KEYS; do
        keys[$i]="$key"
        if ! before[$i]=$(exec_bao kv get "$key" 2>&1); then
            echo "FAIL (pre-restart read of $key)"
            smoke_json_count fail
            return 1
        fi
        i=$((i + 1))
    done

    start=$SECONDS
    kubectl -n "$OPENBAO_NS" delete pod "$OPENBAO_POD" --wait=true --timeout=60s >/dev/null
    # Ready requires the readiness probe (bao status) to pass; the probe fails
    # while sealed, so Ready proves the sidecar unsealed the server with no
    # human step. This lane never invokes any seed script -- it is the only
    # actor, so value equality below is the no-reseed proof.
    # C5 (BAO-STORAGE-SIM-001 D4): tolerate the recreation gap (the pod is
    # briefly NotFound after deletion) by polling within the wait budget
    # instead of dying on the first error under set -e.
    until kubectl -n "$OPENBAO_NS" wait --for=condition=Ready "pod/$OPENBAO_POD" --timeout=15s >/dev/null 2>&1; do
        if [ $((SECONDS - start)) -ge "$restart_wait" ]; then
            echo "FAIL (pod/$OPENBAO_POD not Ready within ${restart_wait}s after deletion)"
            smoke_json_count fail
            return 1
        fi
        sleep 2
    done
    elapsed=$((SECONDS - start))

    i=0
    while [ "$i" -lt "${#keys[@]}" ]; do
        if ! after=$(exec_bao kv get "${keys[$i]}" 2>&1); then
            echo "FAIL (post-restart ${keys[$i]} absent -- no reseed was run)"
            smoke_json_count fail
            return 1
        fi
        if [ "$after" != "${before[$i]}" ]; then
            echo "FAIL (post-restart ${keys[$i]} value changed)"
            smoke_json_count fail
            return 1
        fi
        smoke_json_count pass
        i=$((i + 1))
    done
    echo "PASS (restart lane, ${elapsed}s deletion-to-secrets; NFR-2 bound 120s)"
    if [ "$elapsed" -ge 120 ]; then
        echo "FAIL (recovery ${elapsed}s exceeds the 120s NFR-2 bound)"
        smoke_json_count fail
        return 1
    fi
    return 0
}

if [ "$WITH_RESTART" = "1" ]; then
    check_restart
else
    check_presence
fi
