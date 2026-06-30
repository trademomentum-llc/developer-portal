# Requirements Specification: OpenChoreo Backstage Entity Cards Module

**Document ID:** OPENCHOREO-CARDS-REQ-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Predecessors:** 2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md, 2026-05-28-OptionC-OpenChoreo-Cohesion-Extension-Requirements.md, 2026-05-28-IDP-Policy-Guard-Layer-Requirements.md

---

## 1. Purpose

This document defines the requirements for the **OpenChoreo Entity Cards** functional module. The module provides a set of resilient, annotation-driven React components that surface the multi-angle visibility model (Delivery, Deployment, Runtime, Cost, Policy, Platform) directly on Backstage Component entity pages.

These cards are the primary developer-facing realization of the M3 Production Multi-Angle Visibility Platform on the Option C interim stack. They consume the deterministic namespace predictor (from Option C) as the single source of truth for runtime namespace correlation across all angles.

---

## 2. Scope

### 2.1 In Scope

- Five cards total: OpenChoreoOverviewCard, ObservabilityLinksCard (existing), plus CostCard, PolicyCard, DeploymentCard (new).
- Shared pure TypeScript namespace prediction utility with byte-for-byte equivalence to the Go reference implementation (`tools/namespace-predictor/main.go`).
- Strict annotation-driven data flow using the `openchoreo.dev/*` ontology established in Option C work.
- Resilience contract: cards must render under all partial/missing annotation conditions without throwing.
- Integration points with existing M3 surfaces (SigNoz, Infracost, Gatekeeper, Flux, OpenChoreo API).
- Documentation of the card contract, annotation keys, and equivalence proof for the predictor.
- Updates to the two existing cards to consume the shared predictor (removal of wildcard placeholders).

### 2.2 Out of Scope (Future Increments)

- Full custom EntityLayout / tabbed pages (future M3+).
- Live status polling or embedded iframes (requires collectors).
- Rational-reserve agent angle card (M7).
- Sovereign NeuroDiOS control-plane surface replacement (Option D long-term).

---

## 3. Functional Requirements

### 3.1 Core Presentation Layer (FR-CARD)

- FR-CARD-1: The module shall export a Backstage frontend module (`openchoreoCardsModule`) that contributes all five cards via the catalog plugin extension surface.
- FR-CARD-2: Every card shall use `useEntity()` and read only from `entity.metadata.annotations` plus entity metadata (name, etc.).
- FR-CARD-3: All cards shall implement the same visual contract: `InfoCard` with `variant="gridItem"`, consistent Material-UI typography and link styling, and graceful empty-state / default-value behavior.

### 3.2 Deterministic Namespace Predictor (FR-PRED)

- FR-PRED-1: A single pure function `predictRuntimeNamespace(controlPlaneNs, projectName, environmentName): string` shall be the source of truth in the frontend.
- FR-PRED-2: The function shall produce identical output to `PredictRuntimeNamespace` in `tools/namespace-predictor/main.go` for all valid inputs (sha256 first-8-hex truncation, normalization, 63-char limit).
- FR-PRED-3: The implementation shall be zero-dependency, synchronous where possible, and include embedded test vectors with mathematical equivalence proof.
- FR-PRED-4: The Go CLI binary remains the reference for scripts, CI, and preflight/smoke validation; the TS port is for UI only.

### 3.3 Angle-Specific Cards (FR-ANGLE)

- FR-ANGLE-COST: CostCard shall surface Infracost artifacts (pre- and post-deploy), cost-center attribution, and links to C3 cost-delta policies. It shall display the predicted runtime namespace for correlation.
- FR-ANGLE-POLICY: PolicyCard shall surface Gatekeeper constraint status, Rego policy references, and secret-provenance signals (OpenBao). It shall use the predicted namespace for scoping.
- FR-ANGLE-DEPLOY: DeploymentCard shall surface reconciliation state (predicted vs. actual namespace, ReleaseBinding status, Flux sync health). It is the primary consumer of the exact namespace string for "Deployment & Reconciliation" angle (M3 REQ).

### 3.4 Resilience and Cohesion (FR-RES)

- FR-RES-1: No card shall throw on missing annotations; all required values shall have documented deterministic defaults.
- FR-RES-2: All cards shall reference the Option C annotations (`openchoreo.dev/project`, `openchoreo.dev/component`, `openchoreo.dev/environment`, `openchoreo.dev/control-plane-namespace`, `openchoreo.dev/runtime-namespace-template`).
- FR-RES-3: Cards shall gracefully degrade when the full M3 collector plane is absent (link placeholders + explanatory captions).

### 3.5 Governance and Documentation

- The module shall be accompanied by a complete Requirements + Design + Technical Specification triad (this document + siblings).
- All updates shall maintain PROJECT_SUMMARY.md and TODO.md currency.
- Cross-references to M3 Production Visibility triad and Option C Cohesion triad are mandatory.

---

## 4. Non-Functional Requirements

- NFR-PERF: Cards render synchronously or with minimal async (namespace computation must not block entity page).
- NFR-DEP: Zero new runtime dependencies added to the Backstage app.
- NFR-DET: Namespace prediction must be fully deterministic; any deviation between Go and TS ports constitutes a defect.
- NFR-MAINT: Code style shall follow existing card patterns exactly (no stylistic divergence).
- NFR-SEC: Cards perform no privileged actions; they are read-only presentation.

---

## 5. Success Criteria

1. A developer viewing the `hello-m2` Component entity sees five coherent cards, all displaying the identical predicted runtime namespace `dp-default-default-development-f8e58905` (or equivalent for the entity's annotations).
2. Cost, Policy, and Deployment angles are visible with actionable links to Infracost, Gatekeeper policies, OpenChoreo ReleaseBindings, and Flux resources.
3. The TS predictor produces byte-identical results to the Go binary across a minimum of five test vectors (including boundary cases for length and special characters).
4. Existing two cards have been refactored to use the shared util (no more `*` wildcards in predicted namespace display).
5. The full triad exists and is referenced from M3 and Option C documents.
6. PROJECT_SUMMARY.md and TODO.md accurately reflect the completed cards work and identify the concrete next tandem workstream item.

---

**End of Requirements Specification**

This specification was created in strict accordance with the governing System Architect + NeuroScience persona. It treats the cards module as a first-class functional artifact requiring complete triad coverage. All logic prioritizes determinism and cohesion with the Option C namespace foundation while the sovereign Option D (Jasterish/NeuroDiOS) path continues in parallel.