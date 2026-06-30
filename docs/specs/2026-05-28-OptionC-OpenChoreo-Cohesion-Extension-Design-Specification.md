# Design Specification: Option C — OpenChoreo Cohesion Extension Layer (Interim IDP Glue)

**Document ID:** OPTC-DES-001  
**Version:** 0.1  
**Date:** 2026-05-28  
**Status:** Proposed  
**Traceability:** OPTC-REQ-001

## 1. Architecture Overview

The cohesion layer is a set of thin, deterministic adapters and metadata conventions that sit between:

- Developer authoring surface (Score + Gitea + Backstage catalog)
- Platform orchestrator (OpenChoreo CRDs + controllers + Flux handoff to platform-config)

It does not introduce new runtime control loops. It makes the existing boundary explicit, predictable, and observable.

Core components (all buildable from existing artifacts):
- Namespace Predictor (pure function, duplicated temporarily for validation, target: shared module)
- Translator Enhancements (score2openchoreo flags + annotations)
- Catalog Glue (annotations + entity relations on existing catalog-info.yaml)
- Configuration Alignment (single source for endpoints)
- Observability Links (from catalog to ReleaseBinding / runtime ns)

Data flow (updated M2 path):
1. Score commit -> Gitea Actions (consistent 3333 Gitea)
2. score2openchoreo (now emits runtime-ns annotation + accepts explicit placement)
3. Commit to platform-config (Flux applies to control ns)
4. OpenChoreo (deterministic projection to dp- ns using identical hash)
5. Backstage catalog (static + annotation-enriched views) surfaces 1-4 with links and predicted names

## 2. Namespace Strategy Design (Highest Friction Area)

**Current (problematic):**
- Implicit contract via hard-coded defaults and hand-written samples.
- Runtime name = H("dp", controlNs, project, env) where H is sha256[:8] + truncation (see openchoreo/internal/dataplane/kubernetes/name.go:43).

**Target Design:**
Adopt explicit "Control Plane Namespace" and "Runtime Namespace" as first-class concepts in both translator output and Backstage entities.

- Control Plane Ns: location of Component/Workload/Project CRs (must == Project metadata.namespace).
- Runtime Ns: auto-generated, read-only from developer perspective, predicted by H.

Decision: Duplicate the (tiny, pure, 130-line) name generation function into tools/score2openchoreo/namespace.go for M2+ with exact-match test against OpenChoreo golden vectors. Long-term: propose extraction to a neutral `openchoreo-naming` Go module.

Weighing of alternatives (MOST LOGICAL per efficiency/UX/features):
- Option A (ad-hoc strings in CI): rejected — violates determinism, high human error (evidenced by 3 iterations in SESSION_HANDOFF).
- Option B (duplicate pure func + golden tests): selected. Zero runtime change, full testability, 100% automation safe.
- Option C (call out to OpenChoreo binary at render time): rejected for M2 — adds network/dependency, slower, less hermetic.

## 3. Translator (score2openchoreo) Design Changes

Input extensions (backward compatible):
- New optional flag --control-plane-namespace (defaults to current --namespace for compatibility).
- Score metadata.annotations["openchoreo.dev/control-plane-namespace"] and ["openchoreo.dev/project"] take precedence if present.

Output additions (on Component):
- metadata.annotations["openchoreo.dev/runtime-namespace"] = predicted value
- metadata.annotations["openchoreo.dev/prediction-hash-input"] = "default-default-dev" (for audit)

This satisfies FR-2 and NFR-5 (provenance).

The Convert function remains pure; CLI wiring only.

## 4. Catalog & Entity Model Design

Use Backstage annotations as the wire format (no new plugin yet — quick win).

New annotation vocabulary (prefix openchoreo.dev/):
- control-plane-namespace
- project
- environment
- runtime-namespace-template (e.g. "dp-default-default-dev-*")
- api-base
- component (for leaf workloads)
- is-platform (for the orchestrator entity)

Add synthetic relations (dependsOn) from developer-portal components to openchoreo-platform.

This gives immediate value in catalog UI and TechDocs without waiting for a full entity provider.

Future structural (when justified): a @backstage/plugin-openchoreo-catalog-backend that implements EntityProvider and reads from OpenChoreo API or k8s informers.

## 5. Integration & Configuration Design

Single source of truth for local dev Gitea surface: 3333 (post-migration).

Update strategy: edit-driven (already executed for app-config.yaml). Follow-up script or sed in install-m2.sh to enforce consistency across remaining 3002 references in scripts/ until a config centralization (e.g. .env or values.yaml) is introduced in M3.

Proxy and integration blocks in app-config.yaml now point at the documented port-forward target.

## 6. Validation & Automation Design

- Unit: extend existing convert_test.go + new namespace_test.go with vectors that exactly match OpenChoreo controller output.
- Integration: extend smoke-score.sh or add validate-namespaces.sh that renders, predicts, and (when cluster present) kubectl get releasebinding and checks owner labels + namespace existence.
- Collision math (see Requirements): documented in README; test asserts no collision in 10k random short names (practical proof).

Automation safety: 100% for name prediction because H is deterministic and side-effect free. Script can be invoked in any environment that has the inputs.

## 7. Trade-off Log (Weighed Logically)

- Speed vs Quality: Never chose "easier" string concatenation. Duplicating  the proven namegen + test is the most effective (correctness + auditability + future shared lib path).
- Plugin now vs later: Annotation glue first (days) > full plugin (weeks + triad). Annotations are sufficient for catalog cohesion today and are forward-compatible.
- Upstream vs local: Local duplication + docs with exact paths is effective interim; upstream PR for naming lib is the logical long-term once Option C proves value.

## 8. Risks & Mitigations

Risk: Hash algorithm or truncation changes in OpenChoreo break prediction.  
Mitigation: Golden vector tests + CI step that fails on mismatch; version the predictor with OpenChoreo version.

Risk: Port drift recurs.  
Mitigation: Central .gitea-port or env var in install scripts + guard in start-backstage.sh.

## 9. Implementation Phases

Phase 0 (done 2026-05-28): analysis + small edits + this triad outline + SUMMARY/TODO updates.

Phase 1 (quick win): namespace predictor + tests + CI integration.

Phase 2: catalog annotation completeness + docs.

Phase 3 (structural decision gate): evaluate full entity provider vs continued annotation model.

**End of Design Specification**