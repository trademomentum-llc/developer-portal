# Technical Specification: Internal Developer Platform — Policy Guard Layer + Milestone System

**Document ID:** IDP-TECH-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Predecessors:** IDP-REQ-001, IDP-DS-001

---

## 1. ConstraintTemplate & Constraint Layout

All new constraints follow the naming pattern:

- ConstraintTemplate: cN-<descriptive-kebab> (e.g., c1platformaddonsmainprotected)
- Constraint: cN-enforce

Rego packages live under `package constraints.m2.cN` (or equivalent milestone namespace).

---

## 2. Go Policy Guards (plugins/rr-policy-guards)

- Each guard is a single static Go binary (stdlib only).
- Build output: plugins/rr-policy-guards/bin/<name>-guard (gitignored).
- Invocation: the binary reads from stdin or flags and exits non-zero on violation.
- Used for checks that are awkward or impossible in Rego (host-level hygiene, emoji policy, brew/tofu version pinning, etc.).

---

## 3. Flux + Gatekeeper Wiring

Policies are stored in `policies/` and applied via Flux Kustomizations or HelmReleases in the openchoreo cluster.

Admission webhooks must be in the correct namespace with proper failurePolicy (Fail for enforce, Ignore for audit during rollout).

---

## 4. Evidence Export

Primary evidence sources:
- Gatekeeper audit logs
- Flux event logs
- Output of the Go guard binaries (captured in Gitea Actions or tofu runs)
- Backstage catalog + plugin UI state for milestone visibility

A standard export format (JSON Lines with trace fields) is required for compliance handoff.

---

## 5. Backstage Integration Points (Future)

- Policy dashboard panel (M3+)
- Milestone status cards
- AI swarm proposal review workflow (M7)

---

**End of Technical Specification**

This completes the initial triad for the Policy Guard Layer + Milestone System in developer-portal. The three documents now exist in developer-portal/docs/specs/ and should be referenced from PROJECT_SUMMARY.md, TODO.md, and SESSION_HANDOFF.md.