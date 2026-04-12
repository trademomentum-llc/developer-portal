#!/usr/bin/env bash
# scripts/teardown-m1.sh -- M1 Substrate teardown.
#
# Stops all running services, removes the k3d cluster, cleans up PID files.
# Does NOT delete source code, specs, audit logs, or the backstage directory.
#
# See: M1 Technical Specification Section 9

set -uo pipefail  # no -e: best-effort cleanup even if some steps fail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

info() { printf "\033[1;36m[m1-teardown]\033[0m %s\n" "$*"; }

# Kill Backstage dev server
if [ -f "$HOME/.rational-reserve/m1-backstage-dev.pid" ]; then
    kill "$(cat "$HOME/.rational-reserve/m1-backstage-dev.pid")" 2>/dev/null || true
    rm -f "$HOME/.rational-reserve/m1-backstage-dev.pid"
    info "backstage dev server stopped"
fi

# Kill Gitea port-forward
if [ -f "$HOME/.rational-reserve/m1-gitea-portforward.pid" ]; then
    kill "$(cat "$HOME/.rational-reserve/m1-gitea-portforward.pid")" 2>/dev/null || true
    rm -f "$HOME/.rational-reserve/m1-gitea-portforward.pid"
    info "gitea port-forward stopped"
fi

# Kill any remaining port-forwards
pkill -f "kubectl port-forward.*gitea-http" 2>/dev/null || true
pkill -f "kubectl port-forward.*openchoreo-api" 2>/dev/null || true
info "stray port-forwards cleaned"

# Uninstall Gitea helm release
if helm status gitea -n gitea >/dev/null 2>&1; then
    helm uninstall gitea -n gitea 2>/dev/null || true
    info "gitea helm release uninstalled"
fi

# Delete gitea namespace
kubectl delete namespace gitea --ignore-not-found 2>/dev/null || true
info "gitea namespace deleted"

# NOTE: We do NOT delete the openchoreo-cluster. It is shared with
# ~/Projects/openchoreo/ and managed independently.
# If you want to tear it down, do so from the openchoreo project.

# Clear progress checkpoints (so reinstall starts fresh)
rm -rf "$HOME/.rational-reserve/m1-progress"
info "progress checkpoints cleared"

# Remove Gitea admin password and token (but NOT audit logs)
rm -f "$HOME/.rational-reserve/m1-gitea-admin-password"
rm -f "$HOME/.rational-reserve/m1-gitea-token"
info "gitea credentials removed"

info "M1 torn down. Source tree, specs, and audit logs preserved."
