# Requirements: Option C — OpenChoreo Cohesion Extension Layer (Interim IDP Glue)

**Document ID:** OPTC-REQ-001  
**Version:** 0.1  
**Date:** 2026-05-28  
**Status:** Proposed  
**Owner:** Platform Architecture (subagent analysis)  
**Related:** GAP-OPENCHOREO-VS-NEURODIOS-001, M2 IaC+CD specs, developer-portal TODO 2026-05-28 Option C section

## 1. Purpose (Deterministic Context)

The developer-portal (Backstage) + OpenChoreo stack currently exhibits Frankenstein layering. Per dual-track strategy (PROJECT_SUMMARY 2026-05-28), Option C makes the portal a first-class, cohesive extension of OpenChoreo as the load-bearing interim while sovereign Jasterish/NeuroDiOS (Option D) matures. This document defines the minimal requirements for that cohesion layer. All requirements are traceable to observed code facts in:

- /Users/nnos/Projects/developer-portal/tools/score2openchoreo/convert.go + cli.go + main.go + ci.yaml
- /Users/nnos/Projects/developer-portal/backstage/app-config.yaml + catalog-info.yaml + seed-repos/hello-m2/catalog-info.yaml
- /Users/nnos/Projects/openchoreo/api/v1alpha1/{component,workload,project,environment}_types.go
- /Users/nnos/Projects/openchoreo/internal/{controller/releasebinding/controller.go, dataplane/kubernetes/name.go, pipeline/component/pipeline.go}
- /Users/nnos/Projects/developer-portal/iac/modules/flux/main.tf + openchoreo-environments/main.tf
- Gap analysis and SESSION_HANDOFF.md

## 2. Functional Requirements (FR)

FR-1. The cohesion layer SHALL expose a deterministic function for predicting OpenChoreo runtime (data-plane) namespace given (controlPlaneNamespace, projectName, environmentName) that is byte-for-byte equivalent to openchoreo dpkubernetes.GenerateK8sNameWithLengthLimit for all valid inputs. (Proof: pure sha256 truncation; see Tech Spec for reference implementation and collision math.)

FR-2. score2openchoreo SHALL accept explicit control-plane namespace/project parameters (or infer from Score metadata.annotations["openchoreo.dev/*"]) and SHALL emit an annotation on the generated Component documenting the predicted runtime namespace. Current hard-coded defaults ("default") SHALL be overridable without source change.

FR-3. Backstage Software Catalog SHALL contain machine-readable annotations and relations that model OpenChoreo ownership (projectName), control-plane placement, and runtime-namespace projection for every Component entity. Static file locations SHALL be augmented (not replaced) by at least annotation-driven discovery.

FR-4. Gitea/Backstage/OpenChoreo integration points (URLs, tokens, proxies) SHALL be consistent across app-config, scripts, and seed content. Port and host drift (observed 3002 vs 3333) SHALL be eliminated for the local dev surface.

FR-5. The translation path (Gitea Actions -> score2openchoreo -> platform-config commit -> Flux Kustomization -> OpenChoreo reconciliation) SHALL be observable from the Backstage catalog entry for a Component (links to ReleaseBinding status, predicted runtime ns, live pod).

FR-6. Ownership strings in Backstage (spec.owner) SHALL be mappable to OpenChoreo Project + Gitea identity without stringly-typed magic values. A minimal mapping table or annotation convention is required.

## 3. Non-Functional Requirements (NFR)

NFR-1. Determinism: All namespace, hash, and name generation logic MUST be pure functions (no clocks, random, external IO in the hot path). Validation test coverage >= 95% on equivalence with OpenChoreo implementation.

NFR-2. Performance: Namespace prediction and catalog annotation enrichment MUST complete in < 10ms per entity (local dev). No additional cluster round-trips in catalog index path.

NFR-3. Safety: Changes to cohesion layer MUST NOT alter runtime workload behavior. All new annotations/links are read-only metadata. CI must continue to pass existing smoke suites.

NFR-4. Maintainability: Cohesion improvements SHALL minimize delta to upstream OpenChoreo. Prefer annotation conventions and thin adapters over forking CRD shapes or controllers.

NFR-5. Auditability: Every generated OpenChoreo resource emitted by the translator SHALL carry provenance annotations (pipeline run, score SHA, renderer version).

## 4. Constraints

C-1. No changes to OpenChoreo CRDs or controllers in this interim Option C track (those belong to Option D sovereign work).

C-2. Backstage remains the thin presentation layer; heavy orchestration stays in OpenChoreo.

C-3. All new code/docs follow project AGENTS.md: deterministic logic first, no emojis, full triad for any new module, keep PROJECT_SUMMARY/TODO current, UTF-8 downloadable specs.

C-4. Local dev surface (k3d-openchoreo + port-forwards) remains the validation target.

## 5. Success Criteria (Measurable)

SC-1. `go test` in score2openchoreo + new namespace predictor passes with equivalence golden vectors for (default, default, dev) and (default, hello-m2, staging).

SC-2. Backstage catalog smoke (existing e2e) continues to pass and new annotations are visible in entity view.

SC-3. Manual promotion (dev -> staging commit) succeeds with correct runtime ns prediction for both environments.

SC-4. Port consistency: single source of truth (3333) for local Gitea in all active M2 paths; zero 3002 references in active docs/scripts/configs for Gitea surface.

SC-5. A validator script (or CI step) can, given a rendered Component manifest, compute the exact dp- ns that OpenChoreo will create and confirm the ReleaseBinding owner labels match.

## 6. Out of Scope (for this iteration)

- Full dynamic Backstage entity provider plugin (structural; requires dedicated triad + spike).
- Upstream contribution of Score adapter to OpenChoreo.
- Identity federation (Gitea <-> OpenChoreo RBAC).
- Multi-cluster or production tenancy model.

**End of Requirements**