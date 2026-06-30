# Technical Specification: Backstage Production Runtime

**Document ID:** BACKSTAGE-PROD-RUNTIME-TECH-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-Backstage-Production-Runtime-Design-Specification.md

---

## 1. Implementation Plan

This document provides the concrete implementation steps for deploying PostgreSQL and running Backstage in production mode on the host.

---

## 2. OpenTofu Module: `iac/modules/postgres/`

### 2.1 File Layout

```
iac/modules/postgres/
├── main.tf
├── variables.tf
├── versions.tf
└── outputs.tf
```

### 2.2 `main.tf`

```hcl
resource "random_password" "postgres" {
  length  = 24
  special = false
}

resource "kubernetes_namespace" "backstage" {
  metadata { name = "backstage" }
}

resource "kubernetes_secret" "postgres" {
  metadata {
    name      = "postgres-backstage"
    namespace = kubernetes_namespace.backstage.metadata[0].name
  }
  data = {
    password = random_password.postgres.result
  }
}

resource "helm_release" "postgres" {
  name       = "postgres"
  namespace  = kubernetes_namespace.backstage.metadata[0].name
  repository = "oci://registry-1.docker.io/bitnamicharts"
  chart      = "postgresql"
  version    = var.chart_version

  set {
    name  = "auth.database"
    value = "backstage"
  }
  set {
    name  = "auth.username"
    value = "backstage"
  }
  set_sensitive {
    name  = "auth.password"
    value = random_password.postgres.result
  }
  set {
    name  = "primary.service.type"
    value = "NodePort"
  }
  set {
    name  = "primary.service.nodePorts.postgresql"
    value = var.node_port
  }
  set {
    name  = "primary.resources.requests.memory"
    value = "256Mi"
  }
  set {
    name  = "primary.resources.requests.cpu"
    value = "100m"
  }
  set {
    name  = "primary.persistence.size"
    value = "1Gi"
  }
}
```

### 2.3 `variables.tf`

```hcl
variable "chart_version" {
  description = "Bitnami PostgreSQL Helm chart version"
  type        = string
  default     = "16.4.5"
}

variable "node_port" {
  description = "Host-mapped NodePort for PostgreSQL"
  type        = number
  default     = 30432
}
```

### 2.4 `outputs.tf`

```hcl
output "namespace" {
  value = kubernetes_namespace.backstage.metadata[0].name
}

output "secret_name" {
  value = kubernetes_secret.postgres.metadata[0].name
}

output "node_port" {
  value = var.node_port
}
```

### 2.5 `versions.tf`

```hcl
terraform {
  required_version = ">= 1.9.0, < 1.12.0"
  required_providers {
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.33" }
    helm       = { source = "hashicorp/helm", version = "~> 2.17" }
    random     = { source = "hashicorp/random", version = "~> 3.6" }
  }
}
```

---

## 3. Root `iac/main.tf` Update

Add:

```hcl
module "postgres" {
  source = "./modules/postgres"
}
```

---

## 4. Scripts

### 4.1 `scripts/install-backstage-production.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../iac"
export RR_TOFU_GUARD_BYPASS=1
tofu init -reconfigure
tofu apply -target=module.postgres -auto-approve
echo "Backstage production PostgreSQL installed."
```

### 4.2 `scripts/teardown-backstage-production.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../iac"
export RR_TOFU_GUARD_BYPASS=1
tofu destroy -target=module.postgres -auto-approve
echo "Backstage production PostgreSQL torn down."
```

### 4.3 `scripts/start-backstage-production.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
RUNTIME_DIR="${HOME}/.rational-reserve"
CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"
NODE24_BIN="/opt/homebrew/opt/node@24/bin"

[ -d "$NODE24_BIN" ] && export PATH="$NODE24_BIN:$PATH"

POSTGRES_PORT=$(kubectl --context "$CONTEXT" -n backstage get svc postgres-postgresql -o jsonpath='{.spec.ports[0].nodePort}')
POSTGRES_PASSWORD=$(kubectl --context "$CONTEXT" -n backstage get secret postgres-backstage -o jsonpath='{.data.password}' | base64 -d)

export BACKEND_SECRET="$(cat "${RUNTIME_DIR}/backstage-backend-secret" 2>/dev/null || openssl rand -base64 32 | tee "${RUNTIME_DIR}/backstage-backend-secret")"
export POSTGRES_HOST="127.0.0.1"
export POSTGRES_PORT
export POSTGRES_USER="backstage"
export POSTGRES_PASSWORD
export POSTGRES_DATABASE="backstage"
export GITEA_OAUTH_CLIENT_ID="$(cat "${RUNTIME_DIR}/backstage-oauth-client-id")"
export GITEA_OAUTH_CLIENT_SECRET="$(cat "${RUNTIME_DIR}/backstage-oauth-client-secret")"
export GITEA_HOSTNAME="localhost:3333"
export APP_BASE_URL="http://localhost:3002"
export BACKEND_BASE_URL="http://localhost:7009"

cd "${ROOT_DIR}/backstage"

# Run migrations before starting.
yarn --cwd packages/backend knex --config app-config.yaml --config app-config.production.yaml migrate:latest || true

nohup yarn start --config app-config.yaml --config app-config.production.yaml \
  > "${RUNTIME_DIR}/backstage-production.log" 2>&1 &
echo $! > "${RUNTIME_DIR}/backstage-production.pid"
echo "Backstage production starting on ${BACKEND_BASE_URL}"
```

### 4.4 `scripts/stop-backstage-production.sh`

```bash
#!/usr/bin/env bash
set -uo pipefail
PID_FILE="$HOME/.rational-reserve/backstage-production.pid"
if [ -f "$PID_FILE" ]; then
    kill "$(cat "$PID_FILE")" 2>/dev/null || true
    rm -f "$PID_FILE"
fi
pkill -f "backstage-cli repo start" 2>/dev/null || true
echo "Backstage production stopped"
```

### 4.5 `scripts/smoke-backstage-production.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
BACKEND_URL="${BACKSTAGE_PROD_BACKEND:-http://localhost:7009}"
for i in $(seq 1 60); do
    status=$(curl -s -o /dev/null -w "%{http_code}" "${BACKEND_URL}/" 2>/dev/null || true)
    [ "$status" = "200" ] && break
    sleep 2
done
status=$(curl -s -o /dev/null -w "%{http_code}" "${BACKEND_URL}/" 2>/dev/null || true)
if [ "$status" != "200" ]; then
    echo "FAIL: Backstage production not reachable (${status})" >&2
    exit 1
fi
echo "PASS: Backstage production reachable (${status})"
```

---

## 5. Backstage Config Notes

`app-config.production.yaml` already disables the guest provider and enables permissions. Ensure `backend.listen.port` is not hardcoded so `APP_CONFIG_backend_listen_port` can override it. The current production config uses `backend.baseUrl` from env; add an optional `backend.listen.port` override if needed.

---

## 6. Verification Commands

```bash
./scripts/install-backstage-production.sh
./scripts/start-backstage-production.sh
./scripts/smoke-backstage-production.sh
./scripts/smoke-all.sh
```

---

## 7. Notes

- The Bitnami chart creates the database and user on first boot using the `auth.*` values.
- `yarn knex migrate:latest` requires the backend package to be built first; use `yarn build` if migrations fail.
- Production mode disables hot reload; changes require a restart.
