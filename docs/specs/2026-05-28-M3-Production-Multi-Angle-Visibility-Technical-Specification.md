# Technical Specification: M3 Production Multi-Angle Visibility Platform

**Document ID:** M3-PRODUCTION-VISIBILITY-TS-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Predecessors:** 2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md, 2026-05-28-M3-Production-Multi-Angle-Visibility-Design-Specification.md (updated), 2026-05-28-OpenChoreo-Entity-Cards-Design-Specification.md, m3-observability/* (kickoff drafts)

---

## 1. Purpose and Scope

This document is the authoritative, executable implementation contract for the M3 Production Multi-Angle Visibility increment. It translates the vision (comprehensive visibility across Delivery, Deployment & Reconciliation, Runtime, Cost, Policy, Platform, and future Agent angles) into concrete artifacts, version pins, script contracts, resource attributes, coexistence rules, and verification procedures.

It supersedes the earlier m3-observability/ kickoff triad (2026-05-23) for the production-model phase while preserving its emphasis on safe preflight-first, script-driven, non-destructive evolution of the existing M2 delivery contract.

All implementation must flow through the scripts in `developer-portal/scripts/` and the values in `observability/`. No ad-hoc `helm install` or `kubectl apply` outside these paths is permitted.

---

## 2. Reference Architecture (Implemented Components)

- **Namespace Predictor** (complete): `tools/namespace-predictor/main.go` (reference) + `backstage/packages/app/src/modules/openchoreo-cards/namespace-predictor.ts` (UI port). Deterministic SHA-256 truncation.
- **Backstage Cards** (complete): Five annotation-driven cards in `modules/openchoreo-cards/` (Overview, Observability, Cost, Policy, Deployment). All consume the shared predictor.
- **hello-m2 Workload** (skeleton ready for hardening): OTEL-instrumented Go HTTP service in `seed-repos/hello-m2/`.
- **Script Suite** (advanced skeletons + one wired section): preflight-m3.sh (predictor verification live), install-m3.sh, teardown-m3.sh, smoke-m3.sh.
- **Observability Layout** (scaffolded): Empty `observability/{signoz,otel,dashboards}/` directories + proposed IaC module.
- **Specifications** (complete for this phase): M3 Requirements + Design (updated with cards), full OpenChoreo Entity Cards triad.

---

## 3. Coexistence and Namespace Strategy (Non-Negotiable)

**Critical rule:** M3 must never collide with `openchoreo-observability-plane` (the upstream OpenChoreo observability stack).

Recommended layout (validated via preflight):

| Purpose                    | Namespace          | Notes |
|---------------------------|--------------------|-------|
| SigNoz UI + backend       | `signoz`           | Dedicated. UI on 8080 inside cluster. |
| OTEL Collector (M3)       | `otel-system`      | Separate from any OpenChoreo collector. |
| hello-m2 + workloads      | Predicted by tool  | e.g. `dp-default-default-dev-3a594436` |
| Existing OpenChoreo       | `openchoreo-*`     | Never touch |

Port strategy (local k3d + Colima):
- SigNoz UI: NodePort or port-forward to host 18080 (avoid 8080 collision with other services).
- OTLP/HTTP: 4318 (standard).
- Backstage: existing dev port.
- Gitea: existing.

The deterministic namespace predictor is the **only** mechanism used to compute workload data-plane namespaces in M3 scripts, cards, and instrumentation.

---

## 4. Concrete Version Pins and Helm Values (Post-Preflight)

After running `scripts/preflight-m3.sh`, the following pins (or newer compatible) are recommended for a local k3d environment. Update `observability/` values and this table before any `install-m3.sh` mutation.

Recommended starting pins (as of 2026-05 knowledge; preflight must re-confirm):

- SigNoz: `signoz/signoz` chart v0.38+ (or latest that supports k8s 1.28+ in k3d)
- OpenTelemetry Collector: `open-telemetry/opentelemetry-collector` v0.90+ (standalone collector deployment)

Values files to create:
- `observability/signoz/values.local.yaml` — minimal footprint (single replica, emptyDir or hostPath storage for demo, reduced sampling, no ClickHouse high-availability).
- `observability/otel/collector-values.local.yaml` — OTLP receivers (gRPC + HTTP), exporters to SigNoz, k8sattributes processor, resource detection.

**Coexistence annotations / labels** must be used on all M3 resources to make them obviously distinct from OpenChoreo planes.

---

## 5. hello-m2 OTEL Instrumentation Contract (Hardening Required)

Current skeleton (`seed-repos/hello-m2/main.go`) provides:
- Graceful no-op when `OTEL_EXPORTER_OTLP_ENDPOINT` unset.
- Basic resource (service.name, deployment.environment, service.version).
- Simple HTTP span on `/`.

**M3 Hardening Requirements (to be implemented in next script pass):**

Add these resource attributes on every span / metric (via `resource.WithAttributes`):

```go
attribute.String("openchoreo.project", os.Getenv("OPENCHOREO_PROJECT")),
attribute.String("openchoreo.component", os.Getenv("OPENCHOREO_COMPONENT")),
attribute.String("openchoreo.environment", os.Getenv("OPENCHOREO_ENVIRONMENT")),
attribute.String("openchoreo.runtime_namespace", predictedOrEnvNs),
attribute.String("git.commit.sha", os.Getenv("GIT_SHA") or ldflags),
attribute.String("git.repository", "openchoreo/hello-m2"),
```

Deployment-time injection (Gitea Actions + score2openchoreo or kustomize):
- `OTEL_EXPORTER_OTLP_ENDPOINT=http://signoz-otel-collector.otel-system.svc.cluster.local:4318`
- The four `OPENCHOREO_*` variables + `GIT_SHA` from the triggering commit.

The predicted namespace can be injected at deploy time by calling the Go predictor binary inside the pipeline (or by embedding the tiny logic).

---

## 6. Script Suite Contracts (Current State + Required Evolution)

### 6.1 preflight-m3.sh (Advanced)
- Read-only.
- Now includes section 9: Namespace Predictor Verification (executes the canonical vector and prints the authoritative ns for hello-m2).
- Must be extended to also compute expected ns for every catalog entity found and list current pods in those namespaces (post-install assertion).

### 6.2 install-m3.sh (Skeleton → Implementation Target)
Must:
1. Source or re-execute preflight checks.
2. Add Helm repos (signoz, open-telemetry).
3. Create `signoz` and `otel-system` namespaces with labels.
4. `helm install` / `helm upgrade --install` using the local values files.
5. Apply any additional ConfigMaps / Secrets for OTEL (e.g., SigNoz API token if needed for future auth).
6. Wait for key deployments (SigNoz frontend + collector ready).
7. Seed any demo dashboards from `observability/dashboards/`.

Idempotent and safe to re-run.

### 6.3 teardown-m3.sh
- `helm uninstall` (with `|| true`).
- `kubectl delete ns` (with ignore-not-found).
- Optional: cleanup of any port-forwards or host state.

### 6.4 smoke-m3.sh (Full Spectrum Test Harness Target)
This is the primary "full spectrum tests" entry point.

Required checks (progressive, some cluster-dependent):

1. **Predictor Contract** (always runnable): Re-execute Go binary + (future) shell one-liner + cross-check against known vectors.
2. **Namespace Presence**: For hello-m2, compute expected ns and confirm at least one pod is Running in it (or document why not yet).
3. **SigNoz Health**: HTTP 200 on the frontend (via port-forward or NodePort).
4. **OTEL Collector Ready**: Pod in `otel-system` accepting OTLP.
5. **Trace Flow**: (When instrumented) hello-m2 emits spans visible in SigNoz API or UI (query by service.name + openchoreo.* attributes).
6. **Multi-Angle Cohesion**: 
   - Backstage catalog entity for hello-m2 contains correct openchoreo.dev/* annotations.
   - The five cards would render the identical predicted ns (manual or scripted check when Backstage is running).
7. **Cost/Policy Surfaces**: Presence of Infracost artifacts for the component + policy constraints in the repo (static).
8. **Platform Angle**: Flux and Gitea runner health (re-use existing smoke-flux.sh / smoke-gatekeeper.sh where possible).

All checks must be clearly labeled with the M3 angle they validate.

---

## 7. Full-Spectrum Test Harness Definition

"Full spectrum" means exercising every visibility angle + the deterministic foundation + Option C cohesion in one repeatable command.

Proposed invocation after `install-m3.sh`:

```bash
./scripts/smoke-m3.sh --full-spectrum --predictor-vectors=5
```

Or a dedicated:

```bash
./scripts/verify-m3-spectrum.sh
```

Minimum vector set for predictor (to be encoded in the harness):

- (default, default, dev) → dp-default-default-dev-3a594436
- (default, hello-m2, dev)
- (openchoreo-control, prod-api, prod)
- Long name + underscore normalization case
- 63-char truncation edge case

The harness must be runnable in three modes:
- Offline / dry (predictor + static file checks only)
- Cluster-present (adds pod + service queries)
- Full (requires SigNoz + instrumented workload emitting data)

---

## 8. IaC and Values File Requirements

After the first successful manual preflight + install cycle, a new `iac/modules/observability/` module must be created (following the pattern of existing modules such as flux/ and gatekeeper/).

The module will own:
- Namespace creation
- Helm releases (or HelmRelease CRs if Flux is used for M3 addons)
- Any required RBAC for the collector

Until that module exists, the scripts in `scripts/` are the source of truth.

---

## 9. Verification and Proof Procedures

### 9.1 Predictor Equivalence (Mathematical + Empirical)

Run in any environment with Go:

```bash
go run tools/namespace-predictor/main.go default default dev
```

Expected: `dp-default-default-dev-3a594436`

Run the TS implementation (via node after extracting or future test command) and confirm identical output on the same inputs.

The implementation in `namespace-predictor.ts` (sha256 + predict function) is a direct transliteration; any divergence is a defect.

### 9.2 End-to-End Trace

1. Deploy hello-m2 with OTEL_ENDPOINT and OPENCHOREO_* vars set.
2. Generate traffic.
3. In SigNoz, filter by `service.name="hello-m2"` + `openchoreo.project="default"`.
4. Confirm the span contains the exact runtime namespace attribute computed by the predictor.

### 9.3 Card Correctness (when Backstage running)

Visit the hello-m2 Component page. All five cards must display the same `dp-default-default-dev-3a594436` (or the equivalent for the entity's annotations) as the authoritative runtime namespace.

---

## 10. Automation and Safety Notes

- The entire preflight → install → smoke cycle for a **demo** environment can be fully automated once the cluster is up and the values files are stable.
- Production-like or long-lived environments require human review of resource headroom and storage class behavior (as called out in the original m3-observability technical kickoff).
- The predictor binary is safe to call from any pipeline or script; it has no side effects.

See the separate "automation opportunities" section in the project TODO and the final validation report for quantified safe automation level.

---

## 11. Open Items and Sequencing

1. Create minimal `observability/signoz/values.local.yaml` and `otel/collector-values.local.yaml`. **DONE**.
2. Harden hello-m2 OTEL with full openchoreo attributes + git SHA. **DONE 2026-06-30** (`seed-repos/hello-m2/main.go` + `score2openchoreo --extra-env` + CI injection).
3. Implement install-m3.sh and expand smoke-m3.sh with the full spectrum matrix.
4. Produce the first real end-to-end trace + Backstage card screenshot (when cluster available).
5. Create `iac/modules/observability/` after the first successful run.
6. Update the older m3-observability/ docs to point to this production M3 Technical Specification.

---

**End of Technical Specification**

This document, together with the Requirements, the updated Design Specification, and the OpenChoreo Entity Cards triad, constitutes the complete specification set for the M3 Production Multi-Angle Visibility Platform as of 2026-05-28. All future implementation, testing, and automation must reference these documents. The deterministic namespace predictor and the five annotation-driven cards are the concrete, already-delivered foundation.