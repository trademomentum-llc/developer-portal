# Requirements Specification: M3 Production Multi-Angle Project Visibility Platform

**Document ID:** M3-PRODUCTION-VISIBILITY-REQ-001  
**Version:** 0.1  
**Date:** 2026-05-28  
**Predecessors:** m3-observability/requirements.md (kickoff), 2026-05-28-OptionC-OpenChoreo-Cohesion-Extension-Requirements.md

---

## 1. Purpose

This document defines the requirements for evolving the M3 observability increment into a **production-like multi-angle visibility platform**. The goal is to turn the local developer-portal + OpenChoreo IDP (with the Option C cohesion improvements) into a realistic, demonstrable "working production model" where the full lifecycle and health of any project/component can be observed from multiple complementary angles.

This directly addresses the user's request to develop the platform so that "the development of a project and all its working parts" can be seen from multiple angles, while the long-term sovereign NeuroDiOS path continues to mature in parallel.

---

## 2. Vision

A developer or AI agent should be able to select any Component (e.g., hello-m2) in the Backstage catalog and immediately see coherent, correlated views across:

- **Delivery angle**: Code changes, PR gates (policy, cost, lint via Gatekeeper/Infracost), build, image promotion.
- **Deployment & Reconciliation angle**: OpenChoreo Component → ReleaseBinding → Deployment status, promotion history, predicted vs actual runtime namespace (using the deterministic predictor from Option C work).
- **Runtime health angle**: Traces, metrics, logs (SigNoz), pod/service health, error rates.
- **Cost angle**: Pre-deploy Infracost + post-deploy reference tied to the actual deployment.
- **Policy & Compliance angle**: Gatekeeper violations, secret provenance (openbao), audit events.
- **Agent/AI angle**: Rational-reserve suggestions or analysis related to the component (future M7, stubbed in M3/M4).
- **Platform angle**: Health of dependencies (Flux drift correction, local-registry, Gitea runners, etc.).

All angles must feel cohesive because developer-portal is operating as a well-integrated, first-class extension of the OpenChoreo platform (per Option C).

---

## 3. Scope

### In Scope (M3 Production Increment)
- Completion of core M3 observability (SigNoz + OTEL for workloads + key platform components, with explicit coexistence strategy against openchoreo-observability-plane).
- Backstage catalog and plugin enhancements for multi-angle entity pages (leveraging Option C annotations and namespace predictor).
- Instrumentation of at least the hello-m2 workload and critical platform services.
- Post-deploy cost visibility tied to actual OpenChoreo deployments.
- Basic policy/compliance signals surfaced alongside technical signals.
- Script-driven, repeatable, production-like local environment (preflight, install, teardown, smoke tests).
- Clear documentation of namespace/ownership boundaries (incorporating Option C namespace predictor).

### Out of Scope (future)
- Full rational-reserve (AI swarm) deep integration (M7).
- Production-grade HA, multi-cluster, or real cloud billing.
- Complete replacement of OpenChoreo (long-term Option D).

---

## 4. Functional Requirements

### 4.1 Observability Foundation (Building on M3 Kickoff)
- FR-OBS-1: SigNoz installed in dedicated namespace with zero conflicts against existing openchoreo-observability-plane resources.
- FR-OBS-2: OpenTelemetry collection for application workloads (hello-m2) and selected platform components (Gitea, Flux, etc.).
- FR-OBS-3: Automatic, high-fidelity resource attribution (service.name, service.namespace, deployment.environment, git SHA/commit, openchoreo project/component identifiers).
- FR-OBS-4: Backstage surfaces SigNoz deep links and (where feasible) embedded views for any catalog Component.

### 4.2 Multi-Angle Catalog Experience (Core Production Model)
- FR-VIS-1: Every Component entity page has dedicated, well-organized sections or tabs for Delivery, Deployment, Runtime, Cost, Policy, and Platform angles.
- FR-VIS-2: All signals are correlated by component + environment + git SHA / deployment identifier where possible (using Option C openchoreo.dev annotations and runtime-namespace prediction).
- FR-VIS-3: Cost data (Infracost) is available both pre-deploy (PR gates) and post-deploy (tied to the actual ReleaseBinding/Deployment).

### 4.3 Cohesion with OpenChoreo (Option C Foundation)
- FR-EXT-1: Namespace and placement strategy is deterministic, documented, and enforced via the namespace predictor (from Option C work) and annotations.
- FR-EXT-2: The score2openchoreo + platform-config + Flux + OpenChoreo handoff is reliable, observable, and has clear ownership boundaries.
- FR-EXT-3: Backstage catalog entities accurately model OpenChoreo concepts (Project, Component, Environment, Workload, ReleaseBinding) with deep links and status where possible.

### 4.4 Production Model Characteristics (Local but Realistic)
- FR-PROD-1: Full environment lifecycle (install, operate, teardown) is script-driven, idempotent where possible, and documented.
- FR-PROD-2: All critical paths have health checks, basic alerting rules in SigNoz, and smoke tests.
- FR-PROD-3: A new developer or agent can trace a complete change end-to-end from git commit through all angles using only the Backstage catalog + linked tools.
- FR-PROD-4: The model is stable and extensible enough to serve as the base for M4+ (deeper agent integration, advanced analytics, production hardening).

---

## 5. Non-Functional Requirements

- The visibility layer must not destabilize the local k3d-openchoreo cluster under normal demo/development load.
- Queries and dashboards must feel fast and responsive for a single developer or small team.
- All new instrumentation and configuration must be maintainable and follow the existing M2 patterns (repo-driven, version-pinned where possible).
- Must preserve the M2 delivery contract (no changes to how code flows to a running pod).

---

## 6. Success Criteria

- A full change to hello-m2 can be traced from commit → CI gates → image → OpenChoreo reconciliation → running pod → traces/metrics/logs in SigNoz, all visible and correlated from the Backstage Component page.
- Cost, policy, and platform dependency health are visible alongside technical signals.
- The local environment is stable, script-driven, and feels like a credible "production model" of the full IDP for demonstration, development, and agent experimentation.
- The foundation (including Option C cohesion improvements) is solid enough that M4 work can begin without major rework.

---

**End of Requirements Specification**

This builds directly on the M3 kickoff triad and the 2026-05-28 Option C cohesion extension triad. Design and Technical Specifications (including specific Backstage plugin work, SigNoz configuration, hello-m2 instrumentation details, and namespace strategy) will follow in subsequent documents.

This document was created per the governing persona rules for new functional increments of the platform. It incorporates the sub-agent's Option C deliverables as the interim cohesion foundation.