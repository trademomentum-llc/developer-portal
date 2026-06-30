# Design Specification: M3 Production Multi-Angle Visibility Platform

**Document ID:** M3-PRODUCTION-VISIBILITY-DS-001  
**Version:** 0.1  
**Date:** 2026-05-28  
**Predecessor:** 2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md

---

## 1. Design Philosophy

The production model must feel like a real, operating IDP — even if running locally on k3d. All visibility must be **correlated by component + environment + git SHA** wherever possible, using the deterministic namespace predictor delivered by the Option C sub-agent.

Key design decisions:
- Build on top of the existing M2 delivery contract (do not change how code becomes a running pod).
- Leverage Option C cohesion work (annotations, namespace predictor) as the glue.
- Keep SigNoz as the source of truth for signals, Backstage as the unified developer view.
- Make everything script-driven and observable.

---

## 2. Component Model

- **SigNoz + OTEL Collector**: Core signal store and ingestion.
- **hello-m2 + platform services**: Instrumented workloads emitting OTLP.
- **Backstage**: The multi-angle presentation layer (catalog entities + plugins).
- **Cost & Policy surfaces**: Infracost artifacts + Gatekeeper signals exposed via Backstage.
- **Namespace predictor**: Pure function (from Option C) used everywhere for prediction and correlation.

---

## 3. Backstage Multi-Angle Entity Design (Realized)

The initial implementation uses the annotation-driven grid card pattern contributed via a dedicated frontend module (`openchoreoCardsModule`). Five cards now exist:

### 3.1 The Five Cards

- **OpenChoreoOverviewCard** — Core context (project, component, environment, control-plane ns) + authoritative predicted runtime namespace.
- **ObservabilityLinksCard** — SigNoz deep links (traces, metrics, logs) correlated by service + env + predicted namespace.
- **CostCard** (new) — Infracost pre/post + budget links, cost-center attribution, scoped to the deterministic namespace.
- **PolicyCard** (new) — Gatekeeper / Rego constraints (C1/C2/C3), progressive enforcement references, secret provenance, scoped to predicted ns. Cross-references the full Policy Guard Layer triad.
- **DeploymentCard** (new) — The primary consumer of the namespace predictor. Shows exact predicted runtime namespace, template vs. computed, links to ReleaseBinding, Flux kustomizations, and data-plane views filtered by that namespace. Directly implements the "Deployment & Reconciliation" angle.

All five cards are implemented in:
`backstage/packages/app/src/modules/openchoreo-cards/`

They are registered in `index.ts` via `createFrontendModule` and pulled into the app in `App.tsx`.

### 3.2 Namespace Predictor Integration (Single Source of Truth)

Every card imports and calls:

```ts
import { predictRuntimeNamespace } from './namespace-predictor';

const predicted = predictRuntimeNamespace(controlNs, project, env);
```

The pure function in `namespace-predictor.ts` is a line-for-line semantic port of `tools/namespace-predictor/main.go:PredictRuntimeNamespace`.

**Canonical test vector (verified on both Go binary and Node crypto reference):**

```
predictRuntimeNamespace("default", "default", "development")
→ "dp-default-default-development-f8e58905"
```

Mathematical definition (repeated for emphasis; matches `GenerateK8sNameWithLengthLimit`):

```
names   := ["dp", c, p, e]
input   := join(names, "-")
hash    := hex( SHA-256(input) )[0:8]
base    := truncate-per-part(join(lowercase(sanitize(names)), "-"), 63 - 8 - len(separators))
name    := ensure-dns-compliance(base + "-" + hash)
```

This guarantees that the namespace string a developer sees in any card is identical to the namespace OpenChoreo will create in the data plane and to the value any script (preflight, smoke, CI) will compute when given the same (c, p, e) triple.

Existing wildcard patterns (`dp-*-...-*`) have been removed from the original two cards as part of this wiring pass.

### 3.3 Data Flow (Annotation → Predictor → Card)

1. Catalog entity (catalog-info.yaml or API) carries `openchoreo.dev/*` annotations (injected by Option C work and score2openchoreo).
2. `useEntity()` supplies the annotations to every card.
3. The shared predictor produces the canonical runtime namespace string.
4. Cards render links and values using that string + angle-specific surfaces (SigNoz, Infracost, policies/, OpenChoreo API, Flux manifests).
5. When future M3 collectors write actual state back (ReleaseBinding status, real cost, policy violations), the cards will display predicted vs. observed side-by-side for drift detection.

### 3.4 Future Evolution

- Custom EntityLayout / tabbed pages will later group the five cards into the six M3 angles (Delivery, Deployment, Runtime, Cost, Policy, Platform) with a dedicated Agent angle stub.
- The cards module itself is the Option C presentation layer. Under Option D (NeuroDiOS sovereign kernel) the same information architecture will be native to the control plane; these cards become a reference implementation and migration surface.

See the dedicated triad for exhaustive requirements, design rationale, and technical details:
- `2026-05-28-OpenChoreo-Entity-Cards-Requirements.md`
- `2026-05-28-OpenChoreo-Entity-Cards-Design-Specification.md`
- `2026-05-28-OpenChoreo-Entity-Cards-Technical-Specification.md` (to be completed with full source and cross-validation procedure after environment execution).

---

## 4. Namespace Strategy (Leveraging Option C)

Use the deterministic `PredictRuntimeNamespace` function everywhere (CI, Backstage, docs, scripts). This removes the previous stringly-typed magic and makes the model predictable and testable.

---

## 5. Rollout Phases

- Phase 1: Core M3 (SigNoz + basic hello-m2 instrumentation + Backstage links)
- Phase 2: Multi-angle entity pages + namespace predictor integration
- Phase 3: Cost/policy surfaces + platform component instrumentation
- Phase 4: Agent angle stubs + production hardening

---

**End of Initial Design Specification**

This pairs with the Requirements. The Technical Specification will contain the concrete implementation details (values files, exact OTEL resource attributes, Backstage plugin structure, etc.).