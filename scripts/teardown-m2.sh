#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

info() { printf "\033[1;36m[m2-teardown]\033[0m %s\n" "$*"; }

WIPE_SECRETS=0
for arg in "$@"; do
    case "$arg" in
        --wipe-secrets) WIPE_SECRETS=1 ;;
        *) echo "teardown-m2: unknown argument: $arg" >&2; exit 2 ;;
    esac
done

"$ROOT/scripts/remove-tofu-hook-from-settings.sh" 2>/dev/null || true

if [ -d "$ROOT/iac/.terraform" ]; then
    (cd "$ROOT/iac" && RR_TOFU_GUARD_BYPASS=1 tofu destroy -auto-approve) || true
fi

for ns in flux-system gatekeeper-system gitea-runners tofu-state; do
    kubectl delete namespace "$ns" --ignore-not-found --timeout=2m || true
done

"$ROOT/scripts/delete-m2-gitea-repos.sh" 2>/dev/null || true

# G5 / FR-5 (BAO-STORAGE-DES-001 D-09): OpenBao persistent state survives M2
# teardown by default -- the data-openbao-0 PVC, the openbao-unseal-key
# Secret, and the host custody files. --wipe-secrets deletes all three.
CUSTODY_DIR="$HOME/.rational-reserve/openbao"
if [ "$WIPE_SECRETS" = "1" ]; then
    if [ -t 0 ]; then
        printf "Wipe OpenBao PVC, unseal Secret, and custody files? [y/N] "
        read -r answer
        if [ "$answer" != "y" ]; then
            info "wipe aborted; OpenBao state preserved."
            exit 0
        fi
    fi
    kubectl -n openbao delete pvc data-openbao-0 --ignore-not-found --timeout=2m || true
    kubectl -n openbao delete secret openbao-unseal-key --ignore-not-found || true
    rm -rf "$CUSTODY_DIR"
    info "OpenBao secrets wiped."
else
    info "OpenBao PVC data-openbao-0, secret openbao-unseal-key, and $CUSTODY_DIR preserved (use --wipe-secrets to delete)."
fi

info "M2 torn down. M1 preserved."
