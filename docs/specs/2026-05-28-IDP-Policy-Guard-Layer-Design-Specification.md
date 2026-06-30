# Design Specification: Internal Developer Platform — Policy Guard Layer + Milestone System

**Document ID:** IDP-DS-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Predecessor:** IDP-REQ-001

---

## 1. Design Philosophy

Policy as code is the primary control plane for the IDP. The guard layer must be:

- Auditable (every decision has a trail)
- Progressive (audit → warn → enforce)
- Composable (small, focused constraints that can be combined)
- AI-augmented but not AI-controlled (rational-reserve can propose, never mutate without human + policy approval)

The Milestone System provides the narrative and dependency tracking that pure Git history cannot convey.

---

## 2. Policy Guard Layer Architecture

### 2.1 Technology Stack

- Open Policy Agent (OPA) / Gatekeeper for admission control
- Flux CD for GitOps reconciliation of policies
- Custom Go static binaries in plugins/rr-policy-guards/tools/*-guard (stdlib only, for lightweight checks outside Kubernetes)

### 2.2 Constraint Design Guidelines

- One logical rule per ConstraintTemplate (C1, C2, C3...).
- Use Rego for Kubernetes objects.
- Use the Go guards for host-level or non-Kubernetes checks (emoji naming, brew, tofu, bash hygiene).
- Every constraint must carry a clear rationale comment and a link to the governing spec or risk register entry.

### 2.3 Progressive Enforcement Model

1. Audit mode: violations logged only.
2. Warn mode: violations logged + event emitted to rational-reserve for swarm review.
3. Enforce mode: hard rejection at admission.

Promotion between modes requires evidence (no open violations for N days + review).

---

## 3. Milestone System Design

Each milestone (M1–M7) is tracked via:

- A dedicated section in PROJECT_SUMMARY.md and TODO.md
- Entry/exit criteria documented in specs/
- Artifacts (run logs, policy versions, tofu plans) stored with provenance
- Owner and review cadence

M2 (IaC + CD) is the current validated baseline. M3 (observability) is the next focus.

Cross-milestone dependencies are explicit: M7 (AI swarm integration) cannot proceed meaningfully until M2 is stable in production.

---

## 4. Integration with Sibling Projects

- openchoreo: Provides the cluster substrate (5 planes). Policies must protect the openchoreo core resources.
- rational-reserve: AI swarm can generate policy proposals or detect drift, but all mutations require human + policy guard approval.

---

## 5. Evidence & Auditability

All policy decisions (admission reviews, guard binary runs) must be exportable in a format consumable by compliance tools. Flux + Gatekeeper audit logs + the Go guard outputs form the primary evidence set.

---

**End of Design Specification**

Technical Specification (Rego module structure, Flux kustomization layout, Go guard build & distribution contract, evidence export schema, Backstage plugin wiring for policy dashboards) completes the triad.