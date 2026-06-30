# Requirements Specification: M4 Actual Cluster Cost Visibility

**Document ID:** M4-COST-VISIBILITY-REQ-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md

---

## 1. Purpose

This document defines the requirements for adding **actual cluster cost visibility** to the developer portal as the first concrete M4 increment. Today the Cost angle shows pre-deploy Infracost estimates and a post-deploy cost artifact committed by CI. This increment complements those estimates with real, cluster-derived cost allocation from OpenCost so that teams can compare forecast vs actual spend and reason about workload efficiency.

---

## 2. Vision

A developer or AI agent viewing any Component in Backstage should see:

- **Pre-deploy estimate:** Infracost PR comment / gate result (existing M2).
- **Post-deploy reference:** Cost artifact committed to platform-config (existing M3).
- **Actual cluster allocation:** Per-namespace or per-workload cost computed by OpenCost from Kubernetes metrics.

All three signals are correlated by component, environment, and runtime namespace.

---

## 3. Scope

### In Scope

- Deploy OpenCost in the k3d-openchoreo cluster with a local Prometheus metrics source.
- Expose OpenCost via a stable cluster service and (optionally) a port-forward helper.
- Surface OpenCost deep links or summary data in the existing Backstage CostCard.
- Provide install/teardown scripts and a smoke check.

### Out of Scope

- Real cloud provider billing integration (AWS/GCP/Azure cost and usage reports).
- Persistent long-term cost history outside the cluster.
- Advanced cost anomaly detection or chargeback workflows.

---

## 4. Functional Requirements

### 4.1 OpenCost Deployment

- FR-COST-1: OpenCost runs in a dedicated namespace (e.g., `opencost`).
- FR-COST-2: OpenCost is configured to read metrics from a Prometheus-compatible endpoint inside the cluster.
- FR-COST-3: The deployment is version-pinned, repeatable, and managed via OpenTofu.

### 4.2 Cost Correlation

- FR-COST-4: OpenCost allocations are queryable by the predicted runtime namespace of a Component/Environment pair.
- FR-COST-5: The Backstage CostCard links to the OpenCost UI for the selected Component's namespace.

### 4.3 Operational Hygiene

- FR-COST-6: Install and teardown are script-driven.
- FR-COST-7: A smoke test verifies OpenCost is reachable and returns non-error cost data.

---

## 5. Non-Functional Requirements

- The cost stack must fit within the existing local k3d cluster resource envelope or declare clear resource requirements.
- It must not destabilize SigNoz, OpenChoreo, or the M2 CD loop.
- All configuration follows the existing repo-driven, version-pinned, ASCII-only conventions.

---

## 6. Success Criteria

- `scripts/install-m4.sh` deploys OpenCost + Prometheus and `scripts/smoke-m4.sh` reports a healthy cost endpoint.
- The Backstage CostCard for `hello-m2` links to the OpenCost view filtered by its predicted runtime namespace.
- The M2/M3 smoke suites continue to pass after OpenCost is installed.

---

**End of Requirements Specification**
