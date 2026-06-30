# Gap Analysis: Current Jasterish Micro-Kernel + Compiler Capabilities vs. OpenChoreo Load-Bearing Responsibilities

**Document ID:** GAP-OPENCHOREO-VS-NEURODIOS-001  
**Version:** 0.1  
**Date:** 2026-05-28

---

## 1. Purpose

This document provides a detailed, evidence-based comparison between:

- What the current Jasterish Micro-Kernel (integrated under `engine/nnos/neurodios/jasterish-microkernel/`) + JStar compiler (`apps/src/jstar/`) can realistically deliver today, and
- The actual load-bearing responsibilities OpenChoreo is currently providing across the Developer Control, Platform Orchestration, and Security planes in the five-plane IDP architecture.

The goal is to give a clear, sober view of the gap that must be closed for Option D (Sovereign Platform) to become viable, while Option C (first-class extension of OpenChoreo) is pursued in parallel as an interim solution.

---

## 2. OpenChoreo's Current Load-Bearing Responsibilities (Evidence from developer-portal docs)

From the M2 IaC + CD design and requirements documents, OpenChoreo is explicitly treated as the **workload reconciler** and owns significant parts of:

- **Workload Lifecycle Management**
  - Consumes `Component` + `Workload` + `SecretReference` manifests (rendered by `score2openchoreo`).
  - Creates internal `ComponentRelease`, `Deployment`, and ultimately running Pods.
  - Manages promotion across Environments (dev → staging, etc.).
  - Auto-creates per-environment data-plane namespaces (`dp-<dataplane>-<project>-<environment>-<hash>`).

- **Environment & Tenancy**
  - Owns the Environment CRD concept.
  - Controls the actual runtime namespace strategy for workloads.
  - Provides the boundary between "platform config" and "live workload execution."

- **Secret & Configuration Projection**
  - Works with ExternalSecrets + ClusterSecretStore to materialize secrets into workloads.
  - Owns the runtime secret binding model for applications.

- **Reconciliation & Drift Correction (Workloads)**
  - Acts as the primary reconciler for application workloads (distinct from Flux, which is scoped to add-ons).

- **Platform Orchestration Plane Primitives**
  - Provides the core model for Projects, Components, Environments, and Deployments that the IDP is built around.

OpenChoreo is **not** responsible for (in the current M2 design):
- Flux-managed add-ons drift correction.
- Pipeline policy evaluation (Gatekeeper).
- CI execution (Gitea Actions).
- Backstage catalog surfacing.

---

## 3. Current Jasterish Micro-Kernel + Compiler Capabilities (as of 2026-05-28)

**Micro-Kernel (after Phase 2 integration):**
- ~12.7k lines of JStar across 11 modules.
- Core OS primitives: process management + round-robin scheduling, IPC (message passing), memory (PMM + VMM + heap), ELF loading, basic VFS + ATA disk driver with persistence, IDT/exception handling, syscalls, boot, drivers (PIT, keyboard, etc.).
- CoW fork, `sys_exec`, user-mode entry.
- Has an `EXPANSION_PLAN.md` showing systematic progress.

**JStar Compiler:**
- Rust bootstrap (jstar2) with ongoing self-hosting work (`compiler.jstr`).
- Recent critical codegen bug (data fixups) has been fixed.
- Self-hosting is progressing but still has known gaps (nested loops, data sections in some paths, parser warnings).

**What it does NOT have today (relevant to platform orchestration):**
- No workload / component declarative model.
- No environment or tenancy abstraction.
- No reconciliation loop for desired-state workload objects.
- No native support for Score or equivalent developer authoring surface.
- No catalog/inventory service.
- No secret projection with provenance at platform level.
- No environment promotion workflows.
- The kernel is still primarily a bare-metal OS kernel, not yet a platform orchestration kernel.

---

## 4. Gap Analysis (Side-by-Side)

| Responsibility Area                    | OpenChoreo Today (Load-Bearing)                          | Jasterish/NeuroDiOS Foundation Today                  | Gap Size     | Notes |
|----------------------------------------|----------------------------------------------------------|-------------------------------------------------------|--------------|-------|
| Workload Lifecycle (Component → Pod)  | Full ownership (Component → Release → Deployment → Pod in auto ns) | None (only low-level ELF loading + process primitives) | Very Large | Biggest gap |
| Environment & Namespace Strategy      | Owns Environments + auto-generates `dp-*` namespaces    | Basic process/memory isolation only                  | Large | Namespace model is a major friction point today |
| Secret/Configuration Projection       | Integrated with ExternalSecrets + ClusterSecretStore    | None at platform level                               | Large | Existing openbao usage is lower-level |
| Reconciliation / Desired State        | Primary workload reconciler                             | None (micro-kernel is imperative)                    | Very Large | Core of what makes OpenChoreo "weight-bearing" |
| Developer Authoring Surface (Score)   | Via mandatory `score2openchoreo` translation            | None                                                 | Large | Translation tax is painful |
| Catalog / Inventory                   | Internal models; Backstage consumes lightly             | None                                                 | Medium-Large | Backstage catalog is thin today |
| Policy at Runtime                     | Limited (mostly pipeline-scoped Gatekeeper today)       | None at platform level                               | Medium | Policy Guard Layer specs exist but are not yet platform-integrated |
| Determinism & Observability Primitives| Present via Kubernetes + OpenChoreo internals           | Strong intent + some primitives (IPC, scheduling)    | Medium | NeuroDiOS philosophy is an advantage here |

---

## 5. Realistic Assessment

- The current Jasterish Micro-Kernel is an excellent **sovereign OS foundation** and has made impressive Phase 2 progress.
- It is **nowhere near** ready to replace OpenChoreo's platform orchestration responsibilities.
- The compiler is maturing (critical bug fixed), but self-hosting and language features still have gaps that would affect a production platform layer.
- Closing the gaps above is a multi-year program, even with focused effort.

This confirms that **pursuing Option C (first-class extension of OpenChoreo) as a strong interim solution is pragmatic and necessary** while the sovereign foundation matures.

---

## 6. Recommendations

- Treat the gap analysis as input to both tracks:
  - **Interim (Option C)**: Focus on reducing the translation tax, improving Backstage modeling of OpenChoreo concepts, and clarifying namespace/ownership boundaries.
  - **Long-term (Option D)**: Use this document as the baseline for scoping the minimal viable orchestration primitives (see companion Tech Spec stub).

This document should be kept live and updated as both the Jasterish foundation and the OpenChoreo integration evolve.

---

**End of Gap Analysis**