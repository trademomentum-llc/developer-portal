# Phase 2+3 Closure -- Requirements and Decisions

> Companion to docs/specs/2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md.
> That roadmap owns the FR texts and gap evidence; this document owns the
> closure-scope selection and the open-question decisions for the 2026-08-21
> execution (user-directed: "close phases 2 and 3 on the Mac; the NUC is
> unavailable until later").

**Date:** 2026-08-21
**Status:** APPROVED by execution directive (user instruction 2026-08-21).

## 1. Scope

Phase 2 (self-service project lifecycle): FR-12 (done 08-19), FR-13, FR-14,
FR-17 (done 08-21), FR-18, FR-20, FR-38 (done 08-21).

Phase 3 (observation depth + engagement surfaces): FR-05, FR-06, FR-07,
FR-08, FR-09, FR-10, FR-11, FR-19, FR-21, FR-33 (test stages done 08-21),
FR-34, FR-35, FR-36, FR-37, FR-39.

Not in scope (explicitly): Phase 4 security items beyond Wave-0 (Wave-1 is
NUC-gated on capacity), FR-22 (OQ-16 documented decision only), FR-23
(cross-project view -- no gap evidence for the current single-operator
team; deferred with reasoning), FR-40 (companion record-immutability
workstream touchpoint; artifacts here are written append-only and
digest-friendly to stay compatible).

## 2. Capacity frame (binding)

Host: 16 GiB Mac; Colima VM 6 CPU / 12 GiB; measured 2026-08-21 with the
full platform up: 9120/11934 MiB memory used, CPU requests 95% of 6 cores.
Consequence: NO new standing workload above ~100 MiB memory request is
admitted in this phase. Everything below is config, portal code, scripts,
or documentation. (NFR-07 applied; recorded in SESSION_HANDOFF 0b.)

## 3. Open-question decisions (each with the rationale used)

- OQ-01 log strategy: SigNoz-native log collection via the existing
  standalone OTEL collector (filelog/k8s) into SigNoz logs. Single backend;
  no Loki. Matches the M3 specs' "traces, metrics, logs (SigNoz)".
- OQ-02 alert channel: portal pull surface for the local phase (a
  SigNoz-backed Alerts card; zero new standing workload). Push channels
  (Gitea issues via a receiver) are documented as the NUC-era upgrade.
  Alert rules are codified in-repo from day one.
- OQ-03: resolved 08-19 (3301 port-forward primary, signoz.local ingress noted).
- OQ-05 tenancy: openchoreo.project attribute is the per-project filter key;
  SigNoz community edition suffices at this scale.
- OQ-06 retention: defaults stay demo-grade (3d traces) but become
  CONFIGURABLE in the values files; ClickHouse/Prometheus persistence moves
  to local-path PVCs (2 Gi) so restarts stop wiping telemetry.
- OQ-09 team model: single-operator plus admin-provisioned members.
  scripts/provision-member.sh creates the Gitea user and the Backstage
  group mapping; org.yaml stops being hand-edited. M7 agent tokens ride
  the same mechanism later.
- OQ-14: promotion stays manual-commit ("a feature, not a bug" stands);
  FR-20 is satisfied by a runbook.
- OQ-15 portal philosophy, stated as a rule so cards stop drifting:
  read-only live data renders IN the portal (CostCard pattern);
  actuation deep-links OUT, except forge dispatch which is allowed in
  portal (authenticated, role-gated by the Wave-0 RBAC policy).
- OQ-16: Argo Workflows stays OpenChoreo-internal. Documented; no
  team-facing workflow surface in this phase.
- OQ-17 k8s plugin auth: dev uses the existing localKubectlProxy
  (localhost:8001); the production path (service account when Backstage
  is containerized) is documented, not built.
- OQ-27: forge-run self-CI -- DECIDED BY EXECUTION 2026-08-21
  (.gitea/workflows/self-ci.yaml green on the act-runner; verify-guard
  remains the local pre-push gate).
- OQ-28 test-results store: platform-config repo, test-artifacts/ tree,
  same pattern as cost-artifacts/ and security-artifacts/.
- OQ-30: namespace predictor gets real Go unit tests (go.mod + table
  tests over the canonical vectors); smoke-m3 vectors remain the e2e check.
- OQ-31: decided 08-21 -- the template ships test + security stages
  (proven by scaffold e2e); FR-13 extends the same template to the full
  Score -> deploy loop.

## 4. Acceptance criteria (closure evidence)

Each item is verified live, not by inspection: smoke-all green after each
lane; new checks accrete into the owning smoke suite (m3 for observation,
security for guard-adjacent, a new smoke-engagement.sh only if a lane
produces one); the portal renders live data or honest not-wired states
(NFR-04) verified by Playwright where a card changed.

## 5. Governance note

Per POLICIES.md this slice carries: this requirements+decisions document
and docs/specs/2026-08-21-Phase-2-3-Closure-Technical-Specification.md
(the design and technical layers are folded into one implementation-grade
document, following the Wave-0 REQ+TECH precedent).
