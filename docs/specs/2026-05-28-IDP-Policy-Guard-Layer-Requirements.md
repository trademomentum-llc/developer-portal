# Requirements Specification: Internal Developer Platform — Policy Guard Layer + Milestone System

**Document ID:** IDP-REQ-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Highest-Value Project:** developer-portal (umbrella IDP on Backstage + openchoreo + rational-reserve)

---

## 1. Purpose

The developer-portal implements a self-hosted Internal Developer Platform (IDP) decomposed into seven milestones (M1–M7). The Policy Guard Layer (Gatekeeper/OPA-style constraints + Flux policies) and the Milestone System are the load-bearing mechanisms that enforce platform invariants and track progress.

This Requirements Specification defines the mandatory properties of the Policy Guard Layer and Milestone tracking.

---

## 2. Stakeholders

- Platform engineers (policy authors, IaC maintainers)
- Application teams (consumers of the IDP)
- Security/compliance (policy enforcement as code)
- Auditors (immutable evidence of policy decisions)

---

## 3. Functional Requirements — Policy Guard Layer

FR-001: All critical GitRepository resources (especially platform-addons and core components) must be protected by versioned ConstraintTemplates + Constraints that enforce branch, signature, and provenance rules.

FR-002: Policy violations must produce machine-readable + human-auditable events that can be correlated to specific commits and workloads.

FR-003: Policy evaluation must be deterministic given the same input object and policy version.

FR-004: The layer must support progressive rollout: new policies start in audit mode before enforcement.

FR-005: Integration with openchoreo (orchestrator) and rational-reserve (AI swarm) must allow policies to reference AI-generated recommendations without granting them mutation rights.

---

## 4. Functional Requirements — Milestone System (M1–M7)

The platform is decomposed into seven milestones. Each milestone must have:

- Clear entry criteria (what must be true before work on the milestone begins)
- Exit criteria (objective, measurable completion definition)
- Traceable artifacts (specs, run logs, policy versions)
- Owner and review cadence

M2 (IaC + CD Loop) is currently the validated baseline.

---

## 5. Non-Functional Requirements

- Policy evaluation latency must not materially impact developer inner loop (< 2s p99 for typical objects in local k3d).
- All policy decisions must be exportable for compliance reporting.
- The system must remain usable when the AI swarm (rational-reserve) is degraded or offline.

---

## 6. Compliance Alignment

Policy-as-code is the primary mechanism for security and platform governance. All changes to critical constraints must go through the same review and evidence trail as application code.

---

**End of Requirements Specification**

Design Specification (Backstage plugin model, Flux policy wiring, constraint authoring guidelines, milestone definition template) and Technical Specification (exact Rego structure, admission webhook configuration, evidence export format, integration contracts with openchoreo and rational-reserve) complete the triad.