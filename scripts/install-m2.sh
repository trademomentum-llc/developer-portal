#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT/scripts/lib/colors.sh"

info() { printf "\033[1;36m[m2-install]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[m2-install ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

task_0_tofu_guard() {
    info "Task 0: building rr-tofu-guard"
    cd "$ROOT/plugins/rr-policy-guards/tools/tofu-guard"
    go test ./... || fail "tofu-guard tests failed"
    go build -o "$ROOT/plugins/rr-policy-guards/bin/rr-tofu-guard" .
    "$ROOT/scripts/merge-tofu-hook-into-settings.sh"
}

task_0_5_host_tools() {
    info "Task 0.5: host tools"
    command -v tofu      >/dev/null || brew install opentofu
    command -v flux      >/dev/null || brew install fluxcd/tap/flux
    command -v infracost >/dev/null || brew install infracost
    command -v score-k8s >/dev/null || brew install score-spec/tap/score-k8s
}

task_1_seed_openbao() {
    info "Task 1: seeding openbao kv paths"
    "$ROOT/scripts/seed-openbao-m2-paths.sh"
}

task_2_seed_gitea() {
    info "Task 2: seeding Gitea repos"
    "$ROOT/scripts/seed-gitea-repos.sh"
}

task_3_build_score2openchoreo() {
    info "Task 3: building score2openchoreo"
    cd "$ROOT/tools/score2openchoreo"
    go test ./... || fail "score2openchoreo tests failed"
    go build -o bin/score2openchoreo .
}

task_4_tofu_apply() {
    info "Task 4: tofu apply"
    cd "$ROOT/iac"
    export RR_TOFU_GUARD_BYPASS=1
    tofu init -reconfigure
    tofu apply -auto-approve
}

task_4_5_openbao_storage() {
    info "Task 4.5: OpenBao persistent storage (G5)"
    OPENBAO_STORAGE_SKIP_FULL_SMOKE=1 "$ROOT/scripts/install-openbao-storage.sh"
}

task_5_wait_flux() {
    info "Task 5: waiting for Flux reconcile"
    kubectl -n flux-system wait --for=condition=Ready kustomization.kustomize.toolkit.fluxcd.io/platform-addons --timeout=5m
}

task_6_smoke() {
    info "Task 6: running smoke tests"
    "$ROOT/scripts/smoke-m2.sh"
}

main() {
    task_0_tofu_guard
    task_0_5_host_tools
    task_1_seed_openbao
    task_2_seed_gitea
    task_3_build_score2openchoreo
    task_4_tofu_apply
    task_4_5_openbao_storage
    task_5_wait_flux
    task_6_smoke
    info "M2 complete."
}

main "$@"
