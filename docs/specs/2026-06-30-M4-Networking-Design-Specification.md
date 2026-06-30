# Design Specification: M4 Networking (Cilium + Envoy Gateway)

**Document ID:** M4-NETWORKING-DES-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-M4-Networking-Requirements.md

---

## 1. Overview

This document describes the design for adding a production-like networking layer to the k3d-openchoreo cluster. The work is split into two independent tracks:

1. **Envoy Gateway ingress** on the existing cluster (low risk, additive).
2. **Cilium CNI foundation** via a fresh cluster rebuild (higher risk, documented path).

Both tracks share the same OpenTofu module layout and smoke test harness.

---

## 2. High-Level Architecture

```
+-----------------------------------------------------------+
|                        Host (macOS)                       |
|  /etc/hosts maps *.local -> Envoy Gateway endpoint          |
+-----------------------------------------------------------+
                            |
                            v
+-----------------------------------------------------------+
|  k3d-openchoreo cluster                                   |
|  +-----------------+    +-----------------------------+   |
|  |  Envoy Gateway  |--->|  GatewayClass / Gateway     |   |
|  |  (envoy-gateway)|    |  HTTPRoute resources        |   |
|  +-----------------+    +-----------------------------+   |
|           |                                               |
|           v                                               |
|  +---------------------------------------------------+    |
|  |  CNI layer (Flannel today, Cilium after rebuild)  |    |
|  +---------------------------------------------------+    |
|           |                                               |
|  +--------+---------+  +---------+  +--------+          |
|  | Backstage        |  | Gitea   |  | SigNoz |          |
|  +------------------+  +---------+  +--------+          |
|  +---------------------------------------------------+    |
|  | OpenCost                                            |    |
|  +---------------------------------------------------+    |
+-----------------------------------------------------------+
```

---

## 3. Components

### 3.1 OpenTofu Module: `iac/modules/networking/`

Sub-modules:

- `envoy-gateway/` -- Helm release for Envoy Gateway CRDs and controller.
- `gateway/` -- GatewayClass, Gateway, and HTTPRoute resources.
- `cilium/` -- Helm release for Cilium (used only during cluster rebuild).

Root `iac/modules/networking/main.tf` wires the additive Envoy Gateway path by default. Cilium is toggled via a variable and is primarily intended for fresh clusters.

### 3.2 Envoy Gateway Deployment

- Namespace: `envoy-gateway`.
- Helm chart: `oci://docker.io/envoyproxy/gateway-helm` (version pinned in `variables.tf`).
- GatewayClass: `eg`.
- Gateway: listener on port 80, binds to all namespaces.
- HTTPRoutes: one per hostname, referencing the appropriate backend service and port.

Because k3d runs inside Docker on macOS, the Envoy Gateway service is exposed via a NodePort or LoadBalancer that k3d maps to the host. The helper script discovers the exposed port or IP.

### 3.3 Cilium Deployment

- Installed via Helm in namespace `kube-system` on a fresh cluster.
- k3d cluster created with `--k3s-arg '--flannel-backend=none@server:*'` so Cilium is the sole CNI.
- Hubble UI enabled and exposed through Envoy Gateway as `hubble.local`.
- Cilium status is verified with `cilium status --wait` before proceeding.

### 3.4 DNS Helper Script

`scripts/update-local-hosts.sh`:

- Discovers the Envoy Gateway endpoint (NodePort or external IP).
- Prints the required `/etc/hosts` lines for `backstage.local`, `gitea.local`, `signoz.local`, `opencost.local`, and optionally `hubble.local`.
- Offers an `--apply` flag that appends the entries to `/etc/hosts` (requires sudo).

### 3.5 Cluster Rebuild Script

`scripts/rebuild-cluster-with-cilium.sh`:

- Tears down the existing `k3d-openchoreo` cluster.
- Creates a new cluster with Flannel disabled and the same port-forwards/agents as `install-m1.sh`.
- Installs Cilium via Helm.
- Re-runs `install-m2.sh`, `install-m3.sh`, `install-m4.sh`, and the Envoy Gateway module.
- Re-pushes seed repos and re-imports the Backstage catalog.

This script is deliberately separate and destructive; it requires explicit confirmation.

---

## 4. Routing Table

| Hostname | Destination Service | Destination Port | Notes |
|---|---|---|---|
| backstage.local | `backstage` in default namespace | 3000 | Frontend service |
| gitea.local | `gitea-http` in `gitea` namespace | 3000 | Same backend as port-forward 3333 |
| signoz.local | `signoz` in `signoz` namespace | 8080 | SigNoz query frontend |
| opencost.local | `opencost` in `opencost` namespace | 9090 | OpenCost API/UI |
| hubble.local | `hubble-ui` in `kube-system` namespace | 80 | Only when Cilium is enabled |

---

## 5. Security Considerations

- No TLS for local `.local` hostnames in this milestone.
- Envoy Gateway listener uses plain HTTP.
- Cilium network policies are not enforced by default; policy enforcement is left for M6.

---

## 6. Testing Strategy

- `scripts/smoke-m4-networking.sh`:
  1. Waits for Envoy Gateway pods to be ready.
  2. Resolves each `.local` hostname via `/etc/hosts`.
  3. Performs HTTP GET and expects 200 (or 302 for services that redirect).
  4. Checks that the response body or headers identify the correct service.
- `scripts/smoke-all.sh` continues to pass after networking is added.

---

## 7. Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Cilium replaces Flannel and breaks existing cluster | Cilium is only installed via a fresh cluster rebuild script, not on the live cluster in-place. |
| Envoy Gateway consumes too many resources | Pin a small replica count and resource footprint in Helm values. |
| Hostname/DNS resolution fails on macOS | Provide a script that updates `/etc/hosts`; document manual fallback. |
| k3d LoadBalancer not reachable from host | Use NodePort or `k3d cluster create --port` mappings for the gateway service. |

---

## 8. Acceptance Criteria

- `scripts/install-m4-networking.sh` completes without errors.
- `scripts/update-local-hosts.sh` prints correct `/etc/hosts` entries.
- `scripts/smoke-m4-networking.sh` passes.
- `scripts/smoke-all.sh` still passes.
