# Design Specification: M4 Actual Cluster Cost Visibility

**Document ID:** M4-COST-VISIBILITY-DES-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-M4-Cost-Visibility-Requirements.md

---

## 1. Overview

This design adds an OpenCost-based cost allocation plane to the local IDP. The plane is deployed alongside the existing M3 observability stack but in its own namespace so it can be installed, upgraded, and torn down independently.

---

## 2. Components

| Component | Role | Namespace | Source |
|---|---|---|---|
| Prometheus | Metrics store for node, pod, and container resource usage | `opencost` | prometheus-community Helm chart |
| OpenCost | Cost allocation engine and UI | `opencost` | opencost Helm chart |
| OpenCost exporter | Sidecar/container that scrapes Prometheus and computes allocations | `opencost` | Bundled with OpenCost chart |

---

## 3. Data Flow

1. Prometheus scrapes kubelet/cAdvisor, node-exporter, and kube-state-metrics.
2. OpenCost queries Prometheus for container/resource usage.
3. OpenCost applies default or configurable pricing to produce cost allocations.
4. Backstage CostCard links to OpenCost UI filtered by predicted runtime namespace.

```text
k3d node metrics -> Prometheus -> OpenCost -> UI / API
                                      |
                                      v
                            Backstage CostCard (deep link)
```

---

## 4. Integration Points

### 4.1 Namespace Predictor

OpenCost API accepts `window` and `aggregate` parameters but can also filter by namespace. The Backstage CostCard computes the predicted runtime namespace for the current Component/Environment and opens:

```
http://localhost:<port>/opencost/ui/?namespace=<predicted-ns>
```

The exact URL format depends on OpenCost UI capabilities; the implementation will use the supported query parameter.

### 4.2 Backstage CostCard

The existing CostCard already links to Infracost and the platform-config cost artifact. It will gain an additional link to the OpenCost UI for the component's namespace. No new backend plugin is required for the first increment.

---

## 5. Operational Model

- `scripts/install-m4.sh` runs `tofu apply -target=module.cost`.
- `scripts/teardown-m4.sh` runs `tofu destroy -target=module.cost`.
- `scripts/smoke-m4.sh` port-forwards the OpenCost service and checks `/model/allocation` returns valid JSON.

---

## 6. Resource and Coexistence Considerations

- The local cluster is already running SigNoz, OpenChoreo, Gitea, and Flux. The OpenCost stack will be sized down for local use (single Prometheus replica, reduced retention).
- If cluster memory is exhausted, the install script will warn and suggest increasing the container runtime memory limit.

---

**End of Design Specification**
