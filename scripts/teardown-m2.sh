#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

info() { printf "\033[1;36m[m2-teardown]\033[0m %s\n" "$*"; }

"$ROOT/scripts/remove-tofu-hook-from-settings.sh" 2>/dev/null || true

if [ -d "$ROOT/iac/.terraform" ]; then
    (cd "$ROOT/iac" && RR_TOFU_GUARD_BYPASS=1 tofu destroy -auto-approve) || true
fi

for ns in flux-system gatekeeper-system gitea-runners tofu-state; do
    kubectl delete namespace "$ns" --ignore-not-found --timeout=2m || true
done

"$ROOT/scripts/delete-m2-gitea-repos.sh" 2>/dev/null || true

info "M2 torn down. M1 preserved."
