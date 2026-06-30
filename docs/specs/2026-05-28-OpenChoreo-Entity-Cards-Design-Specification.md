# Design Specification: OpenChoreo Backstage Entity Cards Module

**Document ID:** OPENCHOREO-CARDS-DS-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Predecessor:** 2026-05-28-OpenChoreo-Entity-Cards-Requirements.md

---

## 1. Design Philosophy

The cards module exists to make the abstract "multi-angle visibility" concrete and immediately usable inside the Backstage catalog. Every card is deliberately simple, annotation-driven, and side-effect free. The namespace predictor is elevated to a first-class shared utility because namespace correlation is the only reliable glue between OpenChoreo data-plane reality and developer mental models (Option C cohesion axiom).

Key decisions:
- Strict fidelity to the two existing cards' implementation pattern (InfoCard + useEntity + resilient defaults).
- One pure function for namespace prediction (no duplication).
- Cost / Policy / Deployment cards directly implement the corresponding angles from M3 Requirements section 2 and 4.2.
- The module remains a thin presentation layer; it never owns data or performs reconciliation.

This design is the Option C surface. When the sovereign NeuroDiOS kernel (Option D) matures, these cards become the migration target for a native control-plane UI.

---

## 2. Module Structure

```
backstage/packages/app/src/modules/openchoreo-cards/
 index.ts                 # createFrontendModule + exports all five cards
 OpenChoreoOverviewCard.tsx
 ObservabilityLinksCard.tsx
 CostCard.tsx             # NEW
 PolicyCard.tsx           # NEW
 DeploymentCard.tsx       # NEW
 namespace-predictor.ts   # NEW — single source of truth (TS port)
```

The predictor is co-located for simplicity and to guarantee that any consumer inside the module uses the verified implementation. Later extraction to a shared backstage package is possible but not required for M3.

---

## 3. Annotation Ontology (OpenChoreo Context)

All cards consume the following keys (established by Option C sub-agent):

| Annotation Key                              | Default          | Purpose                              | Used By                  |
|---------------------------------------------|------------------|--------------------------------------|--------------------------|
| openchoreo.dev/control-plane-namespace      | "default"        | Control-plane ns prefix for hashing  | All + predictor          |
| openchoreo.dev/project                      | "unknown"        | Logical project grouping             | All + predictor          |
| openchoreo.dev/component                    | entity.name      | Workload identifier                  | Overview, Deployment     |
| openchoreo.dev/environment                  | "dev"            | Environment (dev/staging/prod)       | All + predictor          |
| openchoreo.dev/api-base                     | localhost:9090   | OpenChoreo API for deep links        | Overview                 |
| openchoreo.dev/runtime-namespace-template   | computed         | Template or observed value           | Observability, Deployment|
| openchoreo.dev/cost-center (future)         | derived          | Budget attribution                   | CostCard                 |

The predictor is always invoked with (control-plane-ns, project, environment).

---

## 4. Deterministic Namespace Prediction — Mathematical Definition

The reference algorithm (Go `PredictRuntimeNamespace`) is:

```
input  := controlPlaneNs + "-" + projectName + "-" + environmentName
digest := SHA-256(input)                    // 32 bytes
short  := hex(digest[0..7])                 // 8 lowercase hex characters
name   := "dp-" + controlPlaneNs + "-" + projectName + "-" + environmentName + "-" + short
name   := lowercase(name)
name   := replace(name, "_", "-")
if len(name) > 63 { name := name[0..62] }
return name
```

This is a pure mathematical function with no external state. Collision resistance follows from SHA-256 preimage resistance. Truncation to 8 hex characters (32 bits) yields 2^32 possible suffixes per (c,p,e) triple — sufficient for local and moderate-scale multi-tenant use while keeping names human-readable.

**Primary test vector (canonical):**
- Input: control="default", project="default", env="development"
- Output: "dp-default-default-development-f8e58905"

All TS and Go implementations must reproduce this vector exactly. Additional vectors (long names, underscores, mixed case) are required in the Technical Specification.

The TS port must be a line-by-line semantic transliteration of the above.

---

## 5. Card Designs

### 5.1 CostCard

- Title: "Cost (Infracost + Budget)"
- Displays predicted runtime namespace (authoritative for cost attribution).
- Surfaces pre-deploy Infracost deltas (from PR gates) and post-deploy reference (future collector).
- Primary links: local infracost reports, C3 cost-delta Rego policies, future budget dashboard scoped to predicted ns.
- Annotation extension: openchoreo.dev/cost-center (optional, falls back to project).
- Graceful state: "Cost signals pending M3 collector integration" caption.

### 5.2 PolicyCard

- Title: "Policy & Compliance"
- Displays policy profile / bundle applicable to the component (derived or annotated).
- Lists or links to active Gatekeeper constraints (C1-platform, C2-score, C3-infracost) and recent violation count for the predicted namespace (stub until collector).
- Deep links into the repository policies/ directory and rr-policy-guards binaries.
- Cross-reference to the full Policy Guard Layer triad.
- Secret provenance note (OpenBao path for the predicted ns).

### 5.3 DeploymentCard

- Title: "Deployment & Reconciliation"
- Most important consumer of the exact predicted namespace string.
- Shows:
  - Predicted runtime namespace (always computed live via the util)
  - Runtime namespace template (from annotation)
  - Reconciliation status stub (Ready / Reconciling / Error)
  - Links: OpenChoreo ReleaseBinding, ComponentRelease history, Flux kustomization for the environment, pod listing filtered by predicted ns.
- This card directly realizes the "Deployment & Reconciliation angle" of the M3 vision.

### 5.4 Refactoring of Existing Cards

Both OpenChoreoOverviewCard and ObservabilityLinksCard shall be updated to import and call the shared `predictRuntimeNamespace` instead of constructing wildcard strings. This enforces the "single source of truth" invariant.

---

## 6. Resilience Contract

Every card implements the same defensive pattern:

```ts
const annotations = entity.metadata.annotations ?? {};
const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
// ... all other fields have explicit fallbacks
const predicted = predictRuntimeNamespace(controlNs, project, env);
// render never throws
```

No useEffect or data fetching in the initial M3 cards (keeps render pure and fast).

---

## 7. Wiring and Extension Model

- `index.ts` creates the frontend module and lists all five card components in the `extensions` array.
- `App.tsx` already imports the module; no change required for feature registration.
- Future: when custom EntityPage layouts are introduced, the cards can be explicitly placed in grid sections.

---

## 8. Dual-Track Considerations

**Option C (current):** These cards + the namespace predictor + M3 scripts give a credible, usable production visibility model on top of OpenChoreo today.

**Option D (sovereign):** The identical information architecture (angles, deterministic naming, policy-as-code) will be re-implemented natively inside the NeuroDiOS control plane once the Jasterish kernel + compiler + Phase-2 subsystems are complete. The Backstage cards then become a reference implementation and migration target.

The design deliberately minimizes lock-in to OpenChoreo-specific UI primitives.

---

## 9. Verification Strategy

- Go binary vs TS function: identical output on the five canonical vectors (documented in Technical Spec).
- Visual: manual inspection of `hello-m2` entity page in local Backstage.
- Resilience: unit tests (or manual) with entities that have zero, partial, and full annotation sets.
- No new dependencies introduced (verified via package.json diff).

---

**End of Design Specification**

The design prioritizes determinism, minimalism, and strict adherence to existing code patterns. It directly enables the six tandem workstreams of the production model phase while preserving the long-term path to a sovereign NeuroDiOS foundation.