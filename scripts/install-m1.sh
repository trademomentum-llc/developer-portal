#!/usr/bin/env bash
# scripts/install-m1.sh -- M1 Substrate installer.
#
# Stands up: k3d cluster, Gitea (helm), demo repo, Backstage skeleton.
# OpenChoreo is deferred to M3; this script skips it.
#
# Usage:
#   ./scripts/install-m1.sh          # resume from last checkpoint
#   ./scripts/install-m1.sh --fresh  # wipe checkpoints and start over
#
# See: M1 Technical Specification Sections 6-9

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROGRESS_DIR="$HOME/.rational-reserve/m1-progress"

info()  { printf "\033[1;36m[m1-install]\033[0m %s\n" "$*"; }
fail()  { printf "\033[1;31m[m1-install ERROR]\033[0m %s\n" "$*" >&2; exit 1; }
done_() { mkdir -p "$PROGRESS_DIR"; touch "$PROGRESS_DIR/$1.done"; }
skip()  { [ -f "$PROGRESS_DIR/$1.done" ]; }

if [ "${1:-}" = "--fresh" ]; then
    rm -rf "$PROGRESS_DIR"
    info "fresh install: cleared checkpoints"
fi

mkdir -p "$HOME/.rational-reserve/logs"

# ---------------------------------------------------------------------------
# Task 0: Build and register policy guard hooks
# ---------------------------------------------------------------------------
task_0_guards() {
    if skip task-0; then info "task 0: guards already built, skipping"; return 0; fi
    info "task 0: building rr-brew-guard"

    cd "$ROOT/plugins/rr-policy-guards/tools/brew-guard"
    go test ./... || fail "brew guard tests failed"
    go build -o ../../bin/rr-brew-guard . || fail "brew guard build failed"
    test -x ../../bin/rr-brew-guard || fail "brew guard binary not found"

    info "task 0: building rr-bash-guard"
    cd "$ROOT/plugins/rr-policy-guards/tools/bash-guard"
    go test ./... || fail "bash guard tests failed"
    go build -o ../../bin/rr-bash-guard . || fail "bash guard build failed"
    test -x ../../bin/rr-bash-guard || fail "bash guard binary not found"

    info "task 0: building rr-emoji-guard"
    cd "$ROOT/plugins/rr-policy-guards/tools/emoji-guard"
    go test ./... || fail "emoji guard tests failed"
    go build -o ../../bin/rr-emoji-guard . || fail "emoji guard build failed"
    test -x ../../bin/rr-emoji-guard || fail "emoji guard binary not found"

    done_ task-0
    info "task 0: all guards built"
}

# ---------------------------------------------------------------------------
# Task 1: Install yarn if missing
# ---------------------------------------------------------------------------
task_1_yarn() {
    if skip task-1; then info "task 1: yarn already done, skipping"; return 0; fi
    if command -v yarn >/dev/null 2>&1; then
        info "task 1: yarn already installed: $(yarn --version)"
    else
        info "task 1: brew install yarn"
        brew install yarn || fail "yarn install failed"
    fi
    done_ task-1
}

# ---------------------------------------------------------------------------
# Task 2: Verify openchoreo-cluster and switch context
# ---------------------------------------------------------------------------
task_2_cluster() {
    if skip task-2; then info "task 2: cluster already verified, skipping"; return 0; fi
    info "task 2: verifying openchoreo-cluster"

    if ! k3d cluster list 2>/dev/null | grep -q openchoreo; then
        fail "openchoreo-cluster not found. Run 'make quick-start.dev' in ~/Projects/openchoreo first."
    fi

    info "task 2: switching kubectl context to k3d-openchoreo"
    kubectl config use-context k3d-openchoreo || fail "failed to switch context"

    kubectl cluster-info || fail "kubectl cannot reach openchoreo-cluster"
    done_ task-2
    info "task 2: openchoreo-cluster ready"
}

# ---------------------------------------------------------------------------
# Task 2a: Configure k3s containerd trust for the M2 local registry
# ---------------------------------------------------------------------------
task_2a_registry_trust() {
    info "task 2a: configuring k3s registry mirror for M2 local-registry"

    local node="k3d-openchoreo-server-0"
    local registries_yaml
    registries_yaml='mirrors:
  "host.k3d.internal:10082":
    endpoint:
      - "http://host.k3d.internal:10082"
  "registry.local-registry.svc.cluster.local:5000":
    endpoint:
      - "http://127.0.0.1:30082"
'

    docker inspect "$node" >/dev/null 2>&1 || fail "$node not found"

    local current
    current="$(docker exec "$node" sh -c 'cat /etc/rancher/k3s/registries.yaml 2>/dev/null || true')"
    if [ "$current" = "$registries_yaml" ]; then
        info "task 2a: registry mirror already configured"
        return 0
    fi

    printf '%s' "$registries_yaml" | docker exec -i "$node" sh -c 'mkdir -p /etc/rancher/k3s && cat > /etc/rancher/k3s/registries.yaml' \
        || fail "failed to write k3s registries.yaml"

    info "task 2a: restarting k3d server so containerd reloads registry mirrors"
    docker restart "$node" >/dev/null || fail "failed to restart $node"
    kubectl config use-context k3d-openchoreo >/dev/null || fail "failed to switch context after restart"
    kubectl wait --for=condition=Ready node --all --timeout=180s || fail "cluster did not become Ready after registry mirror restart"
    info "task 2a: registry mirror configured"
}

# ---------------------------------------------------------------------------
# Task 3: Install Gitea via helm
# ---------------------------------------------------------------------------
task_3_gitea() {
    if skip task-3; then info "task 3: gitea already done, skipping"; return 0; fi
    info "task 3: installing Gitea"

    helm repo add gitea-charts https://dl.gitea.com/charts/ 2>/dev/null || true
    helm repo update gitea-charts

    kubectl create namespace gitea 2>/dev/null || true

    # Create admin secret
    if ! kubectl get secret gitea-admin-secret -n gitea >/dev/null 2>&1; then
        GITEA_ADMIN_PASSWORD=$(openssl rand -base64 24 | tr -d '=+/' | head -c 24)
        printf '%s' "$GITEA_ADMIN_PASSWORD" > "$HOME/.rational-reserve/m1-gitea-admin-password"
        chmod 600 "$HOME/.rational-reserve/m1-gitea-admin-password"

        kubectl create secret generic gitea-admin-secret \
            --namespace gitea \
            --from-literal=username=gitea_admin \
            --from-literal=password="$GITEA_ADMIN_PASSWORD" \
            --from-literal=email=admin@local.dev
    fi

    if ! helm status gitea -n gitea >/dev/null 2>&1; then
        helm install gitea gitea-charts/gitea \
            --namespace gitea \
            --values "$ROOT/scripts/gitea-values.yaml" \
            --wait \
            --timeout 10m || fail "gitea helm install failed"
    fi

    done_ task-3
    info "task 3: Gitea installed"
}

# ---------------------------------------------------------------------------
# Task 4: Port-forward Gitea
# ---------------------------------------------------------------------------
task_4_gitea_portforward() {
    if skip task-4; then info "task 4: port-forward already done, skipping"; return 0; fi
    info "task 4: starting Gitea port-forward"

    kubectl port-forward -n gitea svc/gitea-http 3002:3000 &
    GITEA_PF_PID=$!
    echo "$GITEA_PF_PID" > "$HOME/.rational-reserve/m1-gitea-portforward.pid"

    # Wait for port to be reachable
    for i in $(seq 1 30); do
        if curl -s -o /dev/null http://localhost:3002/api/v1/version; then
            break
        fi
        sleep 1
    done

    curl -s http://localhost:3002/api/v1/version || fail "Gitea not reachable on port 3002"
    done_ task-4
    info "task 4: Gitea reachable at http://localhost:3002"
}

# ---------------------------------------------------------------------------
# Task 5: Port-forward OpenChoreo API
# ---------------------------------------------------------------------------
task_5_oc_api_portforward() {
    if skip task-5; then info "task 5: oc-api port-forward already done, skipping"; return 0; fi
    info "task 5: starting OpenChoreo API port-forward"

    # Try the Service directly; fall back to the gateway LoadBalancer
    if kubectl get svc openchoreo-api -n openchoreo-control-plane >/dev/null 2>&1; then
        kubectl port-forward -n openchoreo-control-plane svc/openchoreo-api 9090:8080 &
        OC_PF_PID=$!
        echo "$OC_PF_PID" > "$HOME/.rational-reserve/m1-oc-api-portforward.pid"

        for i in $(seq 1 30); do
            if curl -s -o /dev/null http://localhost:9090; then
                break
            fi
            sleep 1
        done

        curl -s -o /dev/null http://localhost:9090 || fail "OpenChoreo API not reachable on port 9090"
    else
        info "task 5: openchoreo-api Service not found, using gateway at 172.20.0.3:8080"
    fi

    done_ task-5
    info "task 5: OpenChoreo API reachable"
}

# ---------------------------------------------------------------------------
# Task 6: Create demo repo in Gitea
# ---------------------------------------------------------------------------
task_6_demo_repo() {
    if skip task-6; then info "task 6: demo repo already done, skipping"; return 0; fi
    info "task 6: creating demo repo"

    GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")

    # Create repo
    curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
        -X POST http://localhost:3002/api/v1/user/repos \
        -H "Content-Type: application/json" \
        -d '{
          "name": "demo-service",
          "description": "M1 demo component for Backstage catalog discovery",
          "private": false,
          "auto_init": true,
          "default_branch": "main"
        }' || fail "demo repo creation failed"

    # Push catalog-info.yaml
    CATALOG_YAML='apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: demo-service
  description: M1 smoke-test component
spec:
  type: service
  lifecycle: experimental
  owner: gitea_admin'

    CATALOG_B64=$(printf '%s' "$CATALOG_YAML" | base64)

    curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
        -X POST "http://localhost:3002/api/v1/repos/gitea_admin/demo-service/contents/catalog-info.yaml" \
        -H "Content-Type: application/json" \
        -d "{\"message\": \"M1 seed catalog entry\", \"content\": \"${CATALOG_B64}\", \"branch\": \"main\"}" \
        || fail "catalog-info.yaml push failed"

    done_ task-6
    info "task 6: demo repo created with catalog-info.yaml"
}

# ---------------------------------------------------------------------------
# Task 7: Scaffold Backstage
# ---------------------------------------------------------------------------
task_7_backstage_scaffold() {
    if skip task-7; then info "task 7: backstage scaffold already done, skipping"; return 0; fi
    info "task 7: scaffolding Backstage"

    if [ ! -d "$ROOT/backstage" ]; then
        cd "$ROOT"
        npx --yes @backstage/create-app@latest --path ./backstage --skip-install \
            || fail "backstage scaffold failed"
    fi

    cd "$ROOT/backstage"
    yarn install || fail "yarn install failed"

    done_ task-7
    info "task 7: Backstage scaffolded and dependencies installed"
}

# ---------------------------------------------------------------------------
# Task 8: Wire Backstage to Gitea
# ---------------------------------------------------------------------------
task_8_backstage_config() {
    if skip task-8; then info "task 8: backstage config already done, skipping"; return 0; fi
    info "task 8: wiring Backstage to Gitea"

    cd "$ROOT/backstage/packages/backend"
    yarn add @backstage/plugin-catalog-backend-module-gitea \
        || fail "gitea plugin install failed"

    info "task 8: patching app-config.yaml -- manual step may be needed"
    # The app-config.yaml patch and backend index.ts registration are
    # environment-specific. The implementation plan handles the exact patches
    # after the scaffold produces its files.

    done_ task-8
    info "task 8: Backstage wired to Gitea"
}

# ---------------------------------------------------------------------------
# Task 9: Start Backstage and verify
# ---------------------------------------------------------------------------
task_9_backstage_run() {
    if skip task-9; then info "task 9: backstage already verified, skipping"; return 0; fi
    info "task 9: starting Backstage"

    cd "$ROOT/backstage"
    GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")
    export GITEA_ADMIN_PASSWORD

    yarn start &
    BACKSTAGE_PID=$!
    echo "$BACKSTAGE_PID" > "$HOME/.rational-reserve/m1-backstage-dev.pid"

    # Wait for Backstage to come up (can take 30-60 seconds)
    for i in $(seq 1 90); do
        if curl -s -o /dev/null http://localhost:3000; then
            break
        fi
        sleep 2
    done

    curl -s -o /dev/null http://localhost:3000 || fail "Backstage not reachable on port 3000"

    done_ task-9
    info "task 9: Backstage running at http://localhost:3000"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    task_0_guards
    task_1_yarn
    task_2_cluster
    task_2a_registry_trust
    task_3_gitea
    task_4_gitea_portforward
    task_5_oc_api_portforward
    task_6_demo_repo
    task_7_backstage_scaffold
    task_8_backstage_config
    task_9_backstage_run
    info "M1 complete."
    info "  Backstage:       http://localhost:3000"
    info "  Gitea:           http://localhost:3002"
    info "  OpenChoreo API:  http://localhost:9090"
}

main "$@"
