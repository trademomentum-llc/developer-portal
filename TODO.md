# TODO

> Action list ordered by priority and dependency.

**Snapshot date:** 2026-06-30

---

## Production Multi-Angle Visibility Model (new phase initiated 2026-05-28)

Building on M2 validation and the M3 observability kickoff specs, the focus is now evolving the local IDP into a realistic "working production model." The objective is comprehensive visibility into a project's full lifecycle and health from multiple angles (delivery pipeline, runtime, cost, policy, AI/agent activity), with developer-portal serving as a strong, cohesive extension of OpenChoreo (Option C interim work).

New dedicated requirements spec created: `docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md`.

Immediate priorities for this phase:
- Complete M3 core (SigNoz + OTEL instrumentation for hello-m2 and key platform components, with proper coexistence against openchoreo-observability-plane).
- Implement Backstage catalog enhancements for multi-angle views (tabs/sections for traces, metrics, logs, cost, policy).
- Strengthen Option C cohesion improvements (namespace strategy, entity modeling, translation layer reliability) — sub-agent is actively working this.
- Script-driven, repeatable production-like local environment (install/teardown, preflight, smoke tests).
- Post-deploy cost visibility tied to actual OpenChoreo deployments.
- Foundation for M4+ (deeper rational-reserve integration, advanced analytics, production hardening).

---

## M2 IaC + CD Loop -- where we actually are

M2 is **validated end-to-end locally** through a fresh `hello-m2` pipeline run:

**2026-05-28 Update:** Option C sub-agent completed successfully (general-purpose, 577s, 99 tool calls). Delivered:
- Top 5 cohesion problems with evidence.
- Quick-win edits + full Option C cohesion extension triad.
- Deterministic namespace predictor (with math proof).
- Automation quantification.

These form the interim Option C foundation.

**Production Multi-Angle Visibility Model — All Six Workstreams Launched (2026-05-28):**
1. Preflight execution — advanced (predictor verification section live)
2. Script suite — substantially implemented (real install-m3.sh + teardown-m3.sh + smoke-m3.sh as executable full-spectrum test harness)
3. hello-m2 OTEL instrumentation — skeleton ready; hardening contract defined in M3 Tech Spec
4. M3 Design + Technical (production extensions) — COMPLETE (full Technical Specification created 2026-05-28)
5. Backstage multi-angle entity pages — COMPLETE (five cards + full cards triad)
6. Namespace predictor integration + cost/policy surfaces — COMPLETE across cards + scripts (full cross-language verification in harness)

See new requirements: docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md

All work is being executed in tandem.

**2026-05-28 Cards + Predictor Completion:** The three requested cards (Cost, Policy, Deployment) plus full wiring of the namespace predictor into all five cards and the M3 Design Spec update are done. New OpenChoreo Entity Cards triad created per governance mandate. Existing cards refactored to eliminate wildcard placeholders. Module index updated; App.tsx already correct.
- Top 5 cohesion problems with deterministic evidence across developer-portal + openchoreo trees.
- Quick-win edits (app-config port consistency, enriched catalog-info with openchoreo.dev annotations, namespace predictor sketch, README).
- Full Option C cohesion extension triad created (Requirements/Design/Technical).
- Mathematical validation for deterministic namespace prediction.
- Automation quantification and prioritized roadmap.

These are now the foundation for the interim cohesive extension layer while the sovereign Option D path matures.

**New Phase — Production Multi-Angle Visibility Model (initiated 2026-05-28):** Evolving the local IDP into a realistic working production model for comprehensive visibility into any project's development and working parts from multiple angles (delivery, runtime health, cost, policy, agents). Builds directly on M3 kickoff + the new Option C cohesion triad. See new dedicated requirements: docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md.

**M3 Implementation + Full Spectrum Tests Status (2026-05-29):** 
- Both mandated Technical Specifications created (M3 Production Multi-Angle Visibility TS + OpenChoreo Entity Cards TS with full source + proof procedure).
- Values files + real executable install/teardown/smoke scripts delivered.
- Full-spectrum test harness (`smoke-m3.sh --offline`) executed successfully: 8/8 checks passed, canonical predictor vector + additional vectors + angle coverage + cohesion validated.
- All new artifacts pass syntax, emoji compliance, and persona rules (triads complete, deterministic emphasis, no emojis, cross-refs to Option C/D).

**Session Pause (2026-05-29):** User has directed a pause after outstanding results. The M3 core implementation and full-spectrum testing objectives for this phase are complete. Documents are current. Clean resumption point established.

**Current Concrete Next Items (in priority order, for resumption):**
1. Harden hello-m2 OTEL with full M3 resource attributes (openchoreo.*, predicted ns, git SHA) per the Technical Specification contract. **DONE 2026-06-30** -- implemented in `seed-repos/hello-m2/main.go`; `score2openchoreo` gained `--extra-env KEY=VALUE` for deployment-time injection; CI workflow computes the predicted namespace and injects all required variables.
2. Execute the full install + smoke-m3.sh --cluster cycle on a live k3d-openchoreo environment (when available) to generate real traces and validate cards against live data. **DONE 2026-06-30** -- SigNoz + standalone OTEL collector installed; hello-m2 run #27 succeeded; trace verified in ClickHouse with `openchoreo.runtime_namespace=dp-default-default-development-f8e58905` and `git.commit.sha=a6eaf5a`; `smoke-m3.sh` passes 22/22.
3. Add a live trace-ingestion assertion to `smoke-m3.sh` (query ClickHouse `signoz_traces.signoz_index_v3` for `service.name='hello-m2'` after generating an HTTP request). **DONE 2026-06-30** -- `smoke-m3.sh` now queries ClickHouse via the `signoz-clickhouse` pod and asserts at least one trace exists for the latest `git.commit.sha`.
4. Begin iac/modules/observability/ after the first successful end-to-end run. **DONE 2026-06-30** -- `iac/modules/observability/` created with Helm releases for SigNoz and the standalone OTEL collector plus a patch to disable the bundled SigNoz collector OpAMP manager; `install-m3.sh` and `teardown-m3.sh` now flow through OpenTofu; `tofu plan -target=module.observability` is clean.
5. Update the older m3-observability/ kickoff docs to reference the production M3 triad. **DONE 2026-06-30** -- all three kickoff docs now carry a superseded notice pointing to `docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-*`.
6. Wire Backstage app-config to register `hello-m2` from the local Gitea catalog-info.yaml and verify the five OpenChoreo cards render live data. **DONE 2026-06-30** -- `hello-m2` catalog imported from local Gitea; all five OpenChoreo entity cards render on the Component page using `EntityCardBlueprint.make`.

**New Production-Model TODOs (not yet created until now):**
7. Persist post-deploy Infracost cost artifact on every `hello-m2` push and wire the CostCard to the real artifact (FR-VIS-3). **DONE 2026-06-30** -- CI run #30 succeeded; artifact committed to `platform-config/cost-artifacts/hello-m2/development/latest.json`; CostCard links to the real artifact; `smoke-m3.sh` validates it (22/22 pass).
8. Create a dedicated Backstage multi-angle entity page layout that groups the existing OpenChoreo cards into deliberate tabs/sections (Overview, Deployment, Observability, Cost, Policy, Platform) rather than rendering all cards on the default overview grid (FR-VIS-1). **DONE 2026-06-30** -- new `openchoreo-entity-page` module adds Deployment, Policy, Observability, Cost, and Platform tabs; verified via Playwright on the `hello-m2` Component page. The dedicated-tab cards are no longer duplicated on the Overview grid (only the OpenChoreo Context overview card remains there). Also fixed the recurring guest sign-in / catalog failure by allowing both `localhost:3001` and `127.0.0.1:3001` in backend CORS and by hardening `start-backstage.sh` with `nohup`/`disown` and Node 24 path pinning. Added the missing `group:default/openchoreo` catalog entity to eliminate the entity-relations warning.
9. Enable Backstage catalog discovery from the local Gitea `openchoreo` org and ensure the required port-forwards are managed automatically. **DONE 2026-06-30** -- configured `@backstage/plugin-catalog-backend-module-gitea` provider in `app-config.yaml`; added Gitea integrations for both `localhost:3333` (API/listing) and `localhost:3002` (raw file URLs returned by Gitea); `scripts/start-backstage.sh` now ensures both port-forwards are active before the dev server starts; `hello-m2` and `developer-portal` components are imported automatically and `smoke-m3.sh` passes 22/22.

This work directly implements the user's request to "implement M3 and test it with full spectrum tests" while maintaining the dual-track (Option C cohesion surface while Option D sovereign kernel matures).

- `score2openchoreo` renders schema-valid OpenChoreo `Component` + `SecretReference` + `Workload` multi-document YAML.
- k3d/containerd registry trust is configured through a NodePort-backed local registry mirror.
- Gitea Actions run #24 on `hello-m2` commit `5d88625` completed successfully on 2026-05-22.
- CI built and pushed `registry.local-registry.svc.cluster.local:5000/hello-m2:5d88625`.
- CI committed the rendered OpenChoreo resources to `platform-config`, Flux applied them, and OpenChoreo reconciled `releasebinding.openchoreo.dev/hello-m2-development` to `Ready=True`.
- The live data-plane deployment `hello-m2-development-95297084` is `1/1` available and pod `hello-m2-development-95297084-85c9fd7bcc-hd5qt` is `1/1 Running`.

The final live issue during closeout was OpenChoreo's generated runtime ExternalSecret reading the data plane's `default` ClusterSecretStore (`secret/apps/hello-m2/dev/example-secret`) while the original M2 seed helper only populated the `kv/` mount. `scripts/seed-openbao-m2-paths.sh` now seeds the live `secret/` path and keeps `kv/apps/hello-m2/dev/example-secret` as a compatibility mirror. `scripts/smoke-openbao.sh` passes again, and the runner ExternalSecret is back to `SecretSynced=True`.

What was proved on 2026-05-02: a push to `openchoreo/hello-m2` triggers CI, CI builds + pushes image, CI renders + commits the Component to `platform-config/environments/dev/`, **Flux pulls platform-config**, **Flux applies the (manually-corrected) Component+Workload+Project triplet**, **OpenChoreo creates a ComponentRelease, then a Deployment, then a Pod** in an auto-named per-environment data-plane namespace (`dp-default-default-development-<hash>`). On 2026-05-17, the renderer output and registry-trust configuration were fixed. On 2026-05-22, run #24 proved the full automated path without hand-corrected manifests.

### Outstanding M2 work

| Task | Status | Notes |
|------|--------|-------|
| T21 install-m2.sh end-to-end | DONE 2026-05-02 | Cluster healthy with all M2 namespaces; tofu apply ran successfully; m2i-1..m2i-6 closed |
| T22 first pipeline run on hello-m2 | DONE 2026-05-22 | Run #24 succeeded from push through image build, platform-config commit, Flux apply, OpenChoreo Ready ReleaseBinding, and Running data-plane pod |
| Score2openchoreo renderer rewrite | DONE 2026-05-17 | Emits Component + SecretReference + Workload multi-doc YAML; Go tests, score smoke, and live server-side dry-run passed |
| Push to gitea-com | BLOCKED 2026-05-23 | 2026-05-21 push could not connect to `gitea.com:443`; 2026-05-23 retry reached `gitea.com` but failed authentication. Refresh the cloud Gitea credential/PAT before retrying. |
| Push to local Gitea origin | DONE 2026-05-21 | Created `openchoreo/developer-portal` in local Gitea and pushed `main` through the localhost:3333 port-forward |

---

## M2 install blockers (status as of 2026-05-02)

| ID | Item | Status |
|---|---|---|
| m2i-1 | OpenChoreo CRD group drift -- `core.choreo.dev` -> `openchoreo.dev` | **DONE 2026-05-02 commit 42b2231** -- score2openchoreo, gatekeeper constraints (both `policies/` and the `seed-repos/` mirror), and the technical-spec doc all migrated. Earlier commit 692f200 had only fixed the Environment module |
| m2i-2 | k3d cluster CPU exhausted | **DONE** -- archive memory: colima upgrade (6 CPU / 10 GB) addressed |
| m2i-3 | Stale `tofu-state` ns from 2026-04-21 | **DONE** -- imported during the successful tofu apply |
| m2i-4 | ExternalSecret cache misses under cluster pressure | **DONE** (resolved with m2i-2) |
| m2i-5 | Deprecated `k3d-m1-substrate` cluster | **DONE** -- archive memory: torn down |
| m2i-6 | OpenBao dev-mode `inmem` storage loses kv on restart | **OPEN, low-priority** -- production-readiness item, not M2 closeout. The seed helper now restores missing M2 kv paths, can recover the runner token from the existing Kubernetes Secret, and seeds the OpenChoreo runtime app-secret path |
| **m2i-7** | **k3d/containerd does not trust in-cluster local-registry** | **DONE 2026-05-17** -- `scripts/install-m1.sh` writes `/etc/rancher/k3s/registries.yaml`; local-registry Service is NodePort `30082`; k3s mirror maps `registry.local-registry.svc.cluster.local:5000` to `http://127.0.0.1:30082`. Run #24 proved the pod can pull the fresh `5d88625` image |

---

## m2-renderer-rewrite -- the score2openchoreo redesign

**Status:** DONE 2026-05-17.

**Why it mattered:** the old `tools/score2openchoreo/convert.go` emitted a single Component CRD with `spec.workloadTemplate`, `spec.environment`, and `spec.owner.project`. The actual `openchoreo.dev/v1alpha1/Component` schema rejects all three -- it expects `spec.componentType` (a ref to a ClusterComponentType like `deployment/service`), `spec.owner.projectName`, and stores the workload definition in a separate `Workload` CRD. The renderer was written against an outdated conception of Component.

**Reference shape** (validated by hand this session against the cluster's CRD schema):

```yaml
apiVersion: openchoreo.dev/v1alpha1
kind: Component
metadata: {name, namespace}
spec:
  owner: {projectName: <project>}
  autoDeploy: true
  componentType: {kind: ClusterComponentType, name: deployment/<service|web-application|worker|scheduled-task>}
---
apiVersion: openchoreo.dev/v1alpha1
kind: Workload
metadata: {name: <component>-workload, namespace}
spec:
  owner: {componentName: <component>, projectName: <project>}
  endpoints: {http: {type: HTTP, port: <port>, visibility: [external]}}
  container:
    image: <oci-ref>
    env: [{key, value | valueFrom.secretKeyRef}]
```

**Completion evidence:**

- `go test ./...` passed in `tools/score2openchoreo`.
- `go build -o bin/score2openchoreo .` passed.
- `./scripts/smoke-score.sh` passed.
- `kubectl --context k3d-openchoreo apply --dry-run=server -f /private/tmp/hello-m2-rendered.yaml` accepted the generated `Component` and `Workload`.

**Completed work:**

| ID | Item |
|---|---|
| score-6a | DONE -- types now include Component, Workload, optional Project, and optional SecretReference shapes matching openchoreo.dev/v1alpha1 |
| score-6b | DONE -- Convert returns `[]OpenChoreoResource` and CLI writes multi-document YAML |
| score-6c | DONE -- heuristic: Score service ports -> `deployment/service`, otherwise `deployment/worker`; `pipeline.m2/component-type` annotation overrides |
| score-6d | DONE -- default namespace and project are `default`, matching the local cluster Project and DeploymentPipeline |
| score-6e | DONE -- golden fixtures are multi-document YAML |
| score-6f | DONE -- convert and CLI/golden tests updated for new shape |
| score-6g | DONE -- README and M2 specs updated in the M2 closeout commit |
| score-6h | DONE -- hello-m2 CI render step no longer passes stale namespace/project flags |

---

## M2 done backlog (chronological closeout this session)

| Item | Commit | Date |
|---|---|---|
| Track CLAUDE.md, document Gitea port migration in plan | a99d97e | 2026-05-02 |
| **Complete CRD-group migration: score2openchoreo + gatekeeper + flux platform-config watch** | **42b2231** | **2026-05-02** |
| Implement score2openchoreo Component+Workload renderer rewrite | M2 closeout commit | 2026-05-17 |
| Implement k3d/containerd local-registry trust via NodePort mirror | M2 closeout commit | 2026-05-17 |
| Wire Backstage catalog to local developer-portal and hello-m2 entities | pending commit | 2026-05-22 |

---

## Post-M2 tech debt -- closed before this session

(Carried forward for reference; all DONE before 2026-04-30.)

### Guards (rr-policy-guards)

guard-1 through guard-6 -- DONE 2026-04-23. See git log for commit-level detail.

### score2openchoreo

score-1 through score-5 -- DONE 2026-04-23. (Note: score-6 above is a NEW item from this session, distinct from those.)

### Gatekeeper / Install

gk-1, inst-1, inst-2 -- DONE.

---

## Push / remote resolution -- update

The 2026-04-21 entry suggested gitea.com URL had embedded credentials. As of 2026-05-02:
- `.git/config` has clean URLs (no embedded creds)
- gitea-com auth uses ephemeral PATs minted per-push and immediately revoked
- osxkeychain caches credentials but expires when PATs are revoked

origin (local Gitea) now points at `http://localhost:3333/openchoreo/developer-portal.git` and has received `main`. gitea-com push is blocked by cloud authentication as of 2026-05-23; the remote is reachable but rejected the cached credential.

---

## Open dependency/security backlog

| Item | Status | Notes |
|---|---|---|
| Backstage dependency audit remediation | DONE 2026-06-30 | Resolved all high/critical advisories via Yarn resolutions (`@grpc/grpc-js ^1.14.4`, `ws ^8.21.0`, `axios ^1.18.1`, `undici ^7.28.0`, `react-router ^6.30.4`). Moved dev-only flags (`dangerouslyDisableDefaultAuthPolicy`, `dangerouslyAllowOutsideDevelopment`, `permission.enabled=false`) from `app-config.yaml` to `app-config.local.yaml` so production config stays secure. The only remaining finding is the moderate deprecation warning for `@material-ui/core` v4, which Backstage itself still depends on; resolving it requires a coordinated Backstage version upgrade and is out of scope for this pass. |

---

## M3 Observability -- kickoff

M3 core observability is **implemented and validated live on k3d-openchoreo**. The kickoff triad in `docs/specs/m3-observability/` is superseded by the Production Multi-Angle Visibility triad in `docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-*`; the kickoff docs remain for historical context only.

| Task | Status | Notes |
|---|---|---|
| M3 spec package | DONE 2026-05-23 | Added `docs/specs/m3-observability/{requirements,design-specification,technical-specification}.md`; superseded by 2026-05-28 production triad |
| M3 preflight script | DONE 2026-06-30 | `scripts/preflight-m3.sh` inventories cluster headroom, storage classes, existing OpenChoreo observability-plane resources, and verifies the namespace predictor |
| M3 chart/version inventory | DONE 2026-06-30 | Pinned versions recorded in `iac/modules/observability/variables.tf` and `observability/{signoz,otel}/values.local.yaml` |
| M3 install/teardown scripts | DONE 2026-06-30 | `scripts/install-m3.sh` and `scripts/teardown-m3.sh` flow through OpenTofu module `iac/modules/observability/` |
| M3 smoke suite | DONE 2026-06-30 | `scripts/smoke-m3.sh` validates SigNoz health, OTLP collector, Backstage cards + live catalog entities, `hello-m2` telemetry, live trace ingestion in ClickHouse, and post-deploy cost artifact (22/22 pass) |
| M3 Backstage dependency audit | DONE 2026-06-30 | All high/critical advisories resolved via Yarn resolutions; only the `@material-ui/core` v4 deprecation warning remains (requires Backstage upstream upgrade) |

**Remaining production-model work:** items 7 and 8 are DONE 2026-06-30. Next production-model priorities are TBD with the user.

---

## 2026-05-28 Option C: Cohesive Extension of OpenChoreo (Interim)

**Mission:** Make developer-portal a first-class, intentional extension of OpenChoreo (not layered Frankenstein) while sovereign Jasterish/NeuroDiOS (Option D) matures. See gap analysis in docs/specs/2026-05-28-Gap-Analysis-OpenChoreo-vs-Jasterish-NeuroDiOS-Foundation.md.

**Top 5 Cohesion Problems Identified (deterministic cross-ref):**
1. Namespace/Placement Impedance: hardcoded "default" in score2openchoreo/cli + convert; runtime dp- ns generated via sha256 in openchoreo/internal/dataplane/kubernetes/name.go + releasebinding/controller.go. Components must colocate with Project in control ns.
2. Translation Tax: score2openchoreo is ad-hoc Go binary cloned wholesale in Gitea CI (ci.yaml:21); no plugin/extension point, incomplete fidelity (no Traits/Workflows).
3. Catalog/Entity Disconnect: ~~purely static file locations in app-config.yaml; no provider for OpenChoreo CRs~~ -- resolved for Gitea-hosted entities via `@backstage/plugin-catalog-backend-module-gitea` auto-discovery of the `openchoreo` org; ownership strings now map to `group:default/openchoreo` for `hello-m2`.
4. Ownership/Tenancy Boundary Drift: Gitea orgs + Backstage owners + OpenChoreo Projects + label-injected dp- ns have no synchronized model or provenance.
5. Config/Integration Inconsistencies: Gitea port drift (3002 in some scripts/app-config vs 3333 in M2 docs/hand-off); missing Backstage OpenChoreo plugin surface; Flux only for handoff.

**Implemented Small Cohesion Changes (this pass, edits only to existing artifacts per quality rules):**
- Aligned Gitea integration/proxy in `backstage/app-config.yaml` to `localhost:3333` for API access and added a second integration for `localhost:3002` because Gitea returns raw catalog-info URLs on its internal ROOT_URL port; `start-backstage.sh` now ensures both port-forwards are active.
- Enhanced catalog-info.yaml (root + hello-m2) with openchoreo.dev/* annotations, links to API, runtime-ns template note, and openchoreo-platform entity. Catalog now surfaces the Option C model.
- Extended tools/score2openchoreo/README.md with full "Namespace and ownership placement" section + mathematical determinism proof for the dp- hash + automation note (pure function, collision bound, safe 100% automation for validators).

**Prioritization (weighed for efficiency / UX / features):**
Quick wins (done + next 1-2 days): port alignment (done), catalog annotations (done), add deterministic ns predictor helper script (pure Go, reuses hash logic, test against golden), update all remaining 3002 refs in scripts/ via follow-up pass.
Structural (M4+ or dedicated spike): Backstage custom OpenChoreo catalog entity provider plugin (would require full Req/Design/TechSpec triad per AGENTS.md); extract namegen to shared Go module consumable by both repos; contribute Score adapter surface upstream to OpenChoreo; unified identity provider mapping Gitea->OpenChoreo owners.

**Automation Opportunities (per project rule):**
- Namespace prediction validator: 100% deterministic/safe. Implement as `tools/score2openchoreo/validate-namespaces.go` or standalone; CI can assert rendered Component ns matches Project ns and predicted runtime ns exists post-ReleaseBinding Ready.
- Golden fixture expansion + CRD schema sync: can be driven by `kubectl get crd components.openchoreo.dev -o yaml` piped to generator.
- Full CI clone of developer-portal for score2openchoreo is replaceable by `go install` or OCI image once versioned -- high value, low risk.
Level: full for pure computation layers; guarded (human review on CRD schema changes) for any OpenChoreo API coupling.

See final subagent report for complete proposals, code examples, and 3-spec triad outline for any new Cohesion Adapter module. Update snapshot after review.

---

## M4-M7

| Milestone | Scope | Status |
|---|---|---|
| M4 | OpenCost + Cilium + Envoy Gateway | deferred |
| M5 | RabbitMQ or Kafka + OpenResty front-door | deferred |
| M6 | OPA/Gatekeeper runtime policies + MISP + TheHive + Cortex + Velociraptor + Cloud Custodian | deferred |
| M7 | MCP plugin surfacing Backstage + RR to OpenChoreo + per-agent Gitea tokens | deferred |
