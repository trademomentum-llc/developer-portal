# Phase 2+3 Closure -- Technical Specification

> Implementation-grade companion to
> docs/specs/2026-08-21-Phase-2-3-Closure-Requirements.md (scope, capacity
> frame, OQ decisions are binding there). Four lanes, each with: files,
> approach, verification. Shared invariants: NFR-04 honesty (live or
> labeled not-wired), digest-pinned third-party references, quote shell
> expansions, no emojis, signed commits by the orchestrator after review.

**Date:** 2026-08-21

## Lane A -- observation depth (FR-05, FR-07, FR-08, FR-09, FR-10, FR-11)

Files: observability/otel/collector-values.local.yaml,
iac/modules/observability/*, observability/signoz/values.local.yaml,
observability/dashboards/ (new), scripts/install-m3.sh, scripts/smoke-m3.sh,
backstage/app-config.yaml (proxy), backstage/packages/app/src/modules/
openchoreo-cards/{ObservabilityCard.tsx,AlertsCard.tsx(new)},
openchoreo-entity-page/index.tsx (mount Alerts into Observability tab).

1. FR-05 logs: add a filelog receiver (varlog pods, /var/log/pods) with
   k8s attributes to the standalone collector and a logs pipeline to the
   SigNoz otlp exporter; keep the existing gatekeeper-audit filelog.
2. FR-07 metrics: enable the spanmetrics connector (traces in -> RED
   metrics out -> prometheus/otlp exporter to SigNoz).
3. FR-08 dashboards: materialize observability/dashboards/hello-m2.json
   (RED + recent-errors board) and a platform.json; install-m3.sh seeds
   them via the SigNoz API (localhost:3301) idempotently.
4. FR-09 traces in portal: app-config proxy '/signoz' ->
   http://localhost:3301 (changeOrigin, pathRewrite to /api); extend
   ObservabilityCard to list the component's 10 most recent traces via
   the SigNoz query_range API filtered service.name=<component>; keep
   deep links. Not-wired state when the proxy/forward is down.
5. FR-10 durability: ClickHouse + Prometheus persistence to local-path
   PVCs (2 Gi each) via values files; retention stays 3d default but
   becomes a values key (observability/signoz/values.local.yaml).
6. FR-11 tenancy: document openchoreo.project as the filter key in the
   dashboards and the card queries (service.name + openchoreo.project);
   no new instrumentation this phase.
Verification: smoke-m3.sh gains: logs present in signoz_logs for
hello-m2's namespace; spanmetrics series present; dashboards seeded;
PVCs Bound; smoke-all green. Playwright: Observability tab renders live
trace list (or not-wired).

## Lane B -- orchestration (FR-19, FR-20, FR-21)

Files: backstage/app-config.yaml (kubernetes section -- already has
localKubectlProxy), backstage/packages/app/src/modules/openchoreo-cards/
DeploymentCard.tsx, openchoreo-entity-page/index.tsx,
docs/runbooks/promotion.md (new), scripts/smoke-m3.sh.

1. FR-19: mount the Backstage kubernetes plugin's workload view on the
   Deployment tab for Component entities (the plugin is already
   installed; configure the k3d-openchoreo cluster via the existing
   localKubectlProxy locator, label selector openchoreo.dev/component).
   If the plugin's card conflicts with the custom DeploymentCard, the
   custom card stays primary and the plugin view is a section below.
2. FR-21: DeploymentCard gains an observed-state block: query the
   OpenChoreo API (via /api/proxy/openchoreo, new proxy endpoint to
   localhost:9090) for the ReleaseBinding of the annotated component;
   render releaseName/state/conditions or an explicit not-wired state.
3. FR-20: docs/runbooks/promotion.md -- the manual dev->staging
   promotion procedure (copy the Component file in platform-config,
   commit, Flux applies; verification commands; rollback = git revert).
   Linked from the PlatformCard.
Verification: smoke-m3 gains DeploymentCard/proxy presence checks;
Playwright: Deployment tab shows observed objects; runbook steps
executed once against staging (hello-m2 stays dev-only -- the runbook
is verified by dry-reading its commands against the live cluster, not
by promoting the demo).

## Lane C -- control plane (FR-13, FR-14, FR-18)

Files: backstage/examples/template/content/ (score.yaml, index.js ->
HTTP server, ci.yaml full-loop extension), scripts/provision-member.sh
(new), scripts/sync-catalog-owners.sh (new or folded into
provision-member.sh), docs/, scripts/smoke-m2.sh (optional checks).

1. FR-13 full loop: the template becomes deployable -- index.js gains a
   minimal http server on :3000 (/healthz + /), content/score.yaml
   declares the service, and content/.gitea/workflows/ci.yaml gains the
   proven hello-m2 stages (score2openchoreo render via the developer-
   portal clone, platform-config commit). Repo secret question: verify
   whether local Gitea supports org-level Actions secrets; if yes, seed
   PLATFORM_CONFIG_TOKEN at org level once (scripts/seed-gitea-repos.sh
   gains the step) so every scaffolded repo inherits it; if no, the
   template's publish step documents the manual seed step and the
   pipeline skips platform-config commit when the secret is absent
   (honest degradation, not a fake success).
2. FR-14: scripts/provision-member.sh -- creates a Gitea user (API,
   temp password + must-change-password), adds them to the openchoreo
   org team, and upserts the Backstage user/group entities in
   backstage/examples/org.yaml FROM the Gitea state (Gitea is the
   source of truth; the file becomes generated). Idempotent.
3. FR-18: template content gains docs/index.md + mkdocs.yml and the
   catalog-info.yaml gains backstage.io/techdocs-ref: dir:. Verify the
   TechDocs tab renders for a scaffolded project (local generator).
Verification: full scaffold e2e -- create project, its CI goes green,
its pod lands Running in the dev dataplane namespace, TechDocs tab
renders; then the e2e repo/pod are cleaned up. Provision script tested
against a throwaway user, then the user is deleted.

## Lane D -- engagement (FR-34, FR-35, FR-36, FR-37, FR-39)

Files: scripts/smoke-*.sh (JSON result emission), scripts/ci/
commit-test-artifact.sh (new), backstage/.../CiRunsCard.tsx (dispatch
button), backstage/.../TestResultsCard.tsx (new) mounted on the
Engagement tab, tools/namespace-predictor/ (go.mod + main_test.go),
backstage/packages/app/src/modules/openchoreo-cards/namespace-predictor
(.test.ts), seed-repos/*/README.md + template content (portal backlink).

1. FR-34: smoke suites accept --json <path> (or SMOKE_JSON_OUT env) and
   emit {suite, passed, failed, skipped, ts, git_sha} records;
   scripts/ci/commit-test-artifact.sh commits them to platform-config
   test-artifacts/<project>/development/ following the cost-artifact
   pattern; hello-m2's CI gains the step (smoke-score.sh output).
2. FR-36: TestResultsCard reads the latest test artifact through the
   gitea-actions proxy (raw file from platform-config) and renders
   pass/fail/skip per suite; not-wired state when absent.
3. FR-35: CiRunsCard gains a "Dispatch workflow" action calling POST
   /api/proxy/gitea-actions/repos/{o}/{r}/actions/workflows/{wf}/dispatches
   (proxied, authenticated by the backend-attached token; the card is
   already behind the Wave-0 RBAC role check on the Security tab --
   mount the dispatch control under the same permission).
4. FR-37: tools/namespace-predictor gains go.mod (module-local, stdlib
   only) + main_test.go table tests over the smoke-m3 canonical vectors
   (including truncation and underscore normalization); the TS port
   gains a backstage-cli package test asserting equality with the Go
   binary's output for the same vectors.
5. FR-39: template content README.md and seed-repos/*/README.md gain a
   "Portal: <entity page URL>" line; catalog-info links already cover
   the reverse direction.
Verification: predictor go test + backstage jest pass locally and in
self-CI (self-ci.yaml gains nothing -- predictor now has a module root
and is picked up by the go-tests loop automatically); smoke-all green;
Playwright: Engagement tab shows dispatch control + test results card;
the hello-m2 CI run after merge commits a test artifact.

## Cross-lane rules

- backstage/app-config.yaml is edited ONLY by lanes A (signoz proxy) and
  B (openchoreo proxy + kubernetes locator adjustments); distinct keys;
  the orchestrator reconciles the final diff.
- No lane commits or pushes. The orchestrator reviews each lane's diff,
  runs its verification, then signs the commit.
- provenance/PROVENANCE.md only changes if a NEW third-party dependency
  enters (none should; all pins already exist). If one does, the lane
  flags it and the orchestrator regenerates + re-issues the certificate.
