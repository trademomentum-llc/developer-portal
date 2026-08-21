# Technical Specification: M4 Networking (Cilium + Envoy Gateway)

**Document ID:** M4-NETWORKING-TECH-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-M4-Networking-Design-Specification.md

---

## 1. Implementation Plan

This document provides the concrete implementation steps. The first phase delivers Envoy Gateway ingress on the existing cluster; the second phase documents and scripts a Cilium-based cluster rebuild.

---

## 2. OpenTofu Module Layout

```
iac/modules/networking/
├── main.tf
├── variables.tf
├── versions.tf
├── outputs.tf
├── envoy-gateway/
│   ├── main.tf
│   ├── variables.tf
│   └── values.yaml.tpl
├── gateway/
│   ├── main.tf
│   ├── variables.tf
│   └── httproutes.tf
└── cilium/
    ├── main.tf
    ├── variables.tf
    └── values.yaml.tpl
```

### 2.1 `iac/modules/networking/main.tf`

```hcl
module "envoy_gateway" {
  source = "./envoy-gateway"
}

module "gateway" {
  source     = "./gateway"
  depends_on = [module.envoy_gateway]
  routes = {
    backstage = {
      hostname = "backstage.local"
      service  = { name = "backstage", namespace = "default", port = 3000 }
    }
    gitea = {
      hostname = "gitea.local"
      service  = { name = "gitea-http", namespace = "gitea", port = 3000 }
    }
    signoz = {
      hostname = "signoz.local"
      service  = { name = "signoz", namespace = "signoz", port = 8080 }
    }
    opencost = {
      hostname = "opencost.local"
      service  = { name = "opencost", namespace = "opencost", port = 9090 }
    }
  }
}

module "cilium" {
  count  = var.enable_cilium ? 1 : 0
  source = "./cilium"
}
```

### 2.2 `envoy-gateway/main.tf`

```hcl
resource "helm_release" "envoy_gateway" {
  name             = "envoy-gateway"
  namespace        = "envoy-gateway"
  create_namespace = true
  repository       = "oci://docker.io/envoyproxy"
  chart            = "gateway-helm"
  version          = var.chart_version
  values = [templatefile("${path.module}/values.yaml.tpl", {
    replica_count = var.replica_count
  })]
}
```

`values.yaml.tpl`:

```yaml
deployment:
  replicas: ${replica_count}
envoyGateway:
  gateway:
    controllerName: gateway.envoyproxy.io/gatewayclass-controller
```

### 2.3 `gateway/main.tf`

```hcl
resource "kubectl_manifest" "gateway_class" {
  yaml_body = <<-EOF
    apiVersion: gateway.networking.k8s.io/v1
    kind: GatewayClass
    metadata:
      name: eg
    spec:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
  EOF
}

resource "kubectl_manifest" "gateway" {
  depends_on = [kubectl_manifest.gateway_class]
  yaml_body  = <<-EOF
    apiVersion: gateway.networking.k8s.io/v1
    kind: Gateway
    metadata:
      name: eg
      namespace: envoy-gateway
    spec:
      gatewayClassName: eg
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
  EOF
}
```

`gateway/httproutes.tf`:

```hcl
resource "kubectl_manifest" "httproutes" {
  depends_on = [kubectl_manifest.gateway]
  for_each   = var.routes
  yaml_body  = <<-EOF
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    metadata:
      name: ${each.key}
      namespace: ${each.value.service.namespace}
    spec:
      parentRefs:
        - name: eg
          namespace: envoy-gateway
      hostnames:
        - ${each.value.hostname}
      rules:
        - backendRefs:
            - name: ${each.value.service.name}
              port: ${each.value.service.port}
  EOF
}
```

### 2.4 `cilium/main.tf`

```hcl
resource "helm_release" "cilium" {
  name       = "cilium"
  namespace  = "kube-system"
  repository = "https://helm.cilium.io/"
  chart      = "cilium"
  version    = var.chart_version
  values = [templatefile("${path.module}/values.yaml.tpl", {
    cluster_name = var.cluster_name
    cluster_id   = var.cluster_id
  })]
}
```

`values.yaml.tpl`:

```yaml
cluster:
  name: ${cluster_name}
  id: ${cluster_id}
hubble:
  enabled: true
  relay:
    enabled: true
  ui:
    enabled: true
operator:
  replicas: 1
```

---

## 3. Scripts

### 3.1 `scripts/install-m4-networking.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT_DIR}/iac"
tofu init -reconfigure
tofu apply -target=module.networking -auto-approve
```

### 3.2 `scripts/teardown-m4-networking.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT_DIR}/iac"
tofu destroy -target=module.networking -auto-approve
```

### 3.3 `scripts/update-local-hosts.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
GATEWAY_IP=$(kubectl --context k3d-openchoreo -n envoy-gateway get svc eg -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)
if [ -z "$GATEWAY_IP" ]; then
    NODE_PORT=$(kubectl --context k3d-openchoreo -n envoy-gateway get svc eg -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || true)
    GATEWAY_IP="127.0.0.1:${NODE_PORT}"
fi
cat <<EOF
# M4 Envoy Gateway host entries
127.0.0.1 backstage.local gitea.local signoz.local opencost.local hubble.local
# Gateway endpoint: $GATEWAY_IP
EOF
```

If the gateway uses a NodePort on `127.0.0.1`, the hostname entries cannot map to `127.0.0.1` plus a port in `/etc/hosts`. In that case the script should instead map each hostname to `127.0.0.1` and instruct the user to access via `http://<hostname>:<node_port>`. For milestone purposes we prefer a `k3d cluster create --port 80:80@loadbalancer` mapping so that the gateway service appears on `localhost:80`, allowing clean `.local` entries.

### 3.4 `scripts/smoke-m4-networking.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
HOSTS=(backstage.local gitea.local signoz.local opencost.local)
for host in "${HOSTS[@]}"; do
    status=$(curl -s -o /dev/null -w "%{http_code}" "http://${host}" || true)
    if [ "$status" != "200" ] && [ "$status" != "302" ]; then
        echo "FAIL: $host returned $status" >&2
        exit 1
    fi
    echo "PASS: $host ($status)"
done
```

### 3.5 `scripts/rebuild-cluster-with-cilium.sh`

This script is a convenience wrapper that:

1. Prompts the user for confirmation.
2. Runs `scripts/teardown-m1.sh` (which removes the cluster).
3. Creates a new k3d cluster:
   ```
   k3d cluster create k3d-openchoreo \
     --servers 1 --agents 1 \
     --api-port 127.0.0.1:6550 \
     --k3s-arg '--flannel-backend=none@server:*' \
     --k3s-arg '--disable=traefik@server:*' \
     --port 80:80@loadbalancer \
     --port 443:443@loadbalancer
   ```
4. Installs Cilium via Helm and waits for `cilium status`.
5. Runs `install-m1.sh`, `install-m2.sh`, `install-m3.sh`, `install-m4.sh`, `install-m4-networking.sh`.
6. Re-pushes seed repos.

`--api-port 127.0.0.1:6550` is mandatory on this rebuild. k3d defaults to
publishing the API on `0.0.0.0:6550`, which exposes the kube-apiserver to
the LAN. Host-side NodePort / Envoy publishes that the loadbalancer maps
must similarly be loopback-scoped where the host forwards them; do not
reintroduce wildcard binds as part of the Cilium cutover.

---

## 4. Root `iac/main.tf` Update

Add:

```hcl
module "networking" {
  source = "./modules/networking"
}
```

This is additive on existing clusters (`enable_cilium = false`).

---

## 5. Verification Commands

```bash
./scripts/install-m4-networking.sh
./scripts/update-local-hosts.sh
./scripts/smoke-m4-networking.sh
./scripts/smoke-all.sh
```

---

## 6. Notes

- The Gateway API CRDs are bundled with the Envoy Gateway Helm chart.
- Backstage currently runs on the host, not inside Kubernetes, so the route to `backstage.local` targets the host IP or a NodePort that forwards to the host. For a fully containerized setup, Backstage would need to be deployed into the cluster (future milestone).
- Cilium and Envoy Gateway versions are pinned in `variables.tf`.
