# SESSION HANDOFF

> Read this file FIRST in the next session. It tells you where we stopped,
> what is now committed that was not before, what is still outstanding, and
> exactly what to do first.

**Last updated:** 2026-06-30
**Reason for handoff:** M3 Production Multi-Angle Visibility and M4 cost visibility are live. `./scripts/smoke-all.sh` reports `ALL SMOKE SUITES PASSED (M2, M3, M4)`. `AGENTS.md` has been refreshed with current commands, module list, and port-forwards.

---

## 1. The single most important thing

M3 and M4 cost visibility are now live and validated end-to-end:

- SigNoz v0.130.1 installed in namespace `signoz`.
- Standalone OpenTelemetry Collector v0.155.0 installed in namespace `otel-system` and forwarding OTLP/gRPC to SigNoz.
- The SigNoz `signoz-otel-collector` Deployment was patched to remove the OpAMP-only manager arguments so that OTLP ports 4317/4318 are exposed.
- `hello-m2` run #27 (commit `a6eaf5a`) succeeded in Gitea Actions, built/pushed image `registry.local-registry.svc.cluster.local:5000/hello-m2:a6eaf5a`, and rendered OpenChoreo resources to `platform-config`.
- `hello-m2` is `1/1 Running` in namespace `dp-default-default-development-f8e58905` with injected env vars:
  - `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector-opentelemetry-collector.otel-system.svc.cluster.local:4318`
  - `OPENCHOREO_RUNTIME_NAMESPACE=dp-default-default-development-f8e58905`
  - `OPENCHOREO_ENVIRONMENT=development`
  - `GIT_SHA=a6eaf5a`
- Live trace verified in ClickHouse `signoz_traces.signoz_index_v3` with `serviceName='hello-m2'`, `resources_string['openchoreo.runtime_namespace']='dp-default-default-development-f8e58905'`, and `resources_string['git.commit.sha']='a6eaf5a'`.
- `./scripts/smoke-all.sh` reports `ALL SMOKE SUITES PASSED (AUTH, M2, M3, M4, BACKSTAGE-PRODUCTION)`.
- Backstage production runtime is validated: PostgreSQL-backed, `NODE_ENV=production`, guest disabled, Gitea auth enabled.
- Backstage Gitea OAuth provider is implemented: backend module `packages/backend/src/modules/giteaAuth.ts`, frontend sign-in module `packages/app/src/modules/giteaSignIn.tsx`, and `scripts/smoke-auth.sh` verifies `/api/auth/gitea/start` redirects to Gitea.
- `AGENTS.md` was refreshed to list M3/M4/auth/production scripts, the current root `iac/main.tf` modules, required port-forwards, and the Node 24 / guest-auth / production config notes.
- `./scripts/smoke-m3.sh` now passes 22/22 checks, including live Backstage catalog entity import, a live trace-ingestion assertion, and the post-deploy cost artifact.
- Backstage `yarn tsc` passes and the five OpenChoreo entity cards render on the live `hello-m2` catalog page after converting them to `EntityCardBlueprint.make` extension definitions (the initial `convertLegacyEntityCardExtension` attempt failed because the plain card components lacked legacy extension metadata).
- Backstage catalog provider auto-imports `hello-m2` and `developer-portal` from the local Gitea `openchoreo` org via `@backstage/plugin-catalog-backend-module-gitea`; Gitea integrations are configured for both `localhost:3333` (API) and `localhost:3002` (raw file URLs).
- Backstage dev ports moved from `3000/7007` to `3001/7008` in `app-config.yaml` and `playwright.config.ts` to avoid the Gitea service on port 3000.
- `catalog-info.yaml` root System description folded to a `>-` block scalar to avoid the `Option C:` YAML parse error, and `openchoreo.dev/system` annotation quoted as a string to satisfy the Backstage envelope policy.
- `iac/modules/observability/` created for repeatable SigNoz + OTEL Collector installs; `install-m3.sh` now applies it via OpenTofu; `tofu plan -target=module.observability` shows a clean 3-to-add plan.
- `./scripts/teardown-m3.sh` updated to destroy the observability module via OpenTofu.
- `./scripts/smoke-m3.sh` passes 13/13 live checks.

The namespace predictor (Go + TypeScript) is now a byte-for-byte semantic replica of OpenChoreo's `GenerateK8sNameWithLengthLimit(63, "dp", ...)` algorithm, with the canonical vector `dp-default-default-development-f8e58905` verified against the live cluster.

---

## 2. Git state at handoff

- **Branch:** `main`
- **Local HEAD:** `2078f6e` -- `security(backstage): move permission.enabled=false to app-config.local.yaml`
- **origin (local Gitea):** `http://localhost:3333/openchoreo/developer-portal.git` is up-to-date with `main`.
- **hello-m2 (local Gitea):** `http://localhost:3333/openchoreo/hello-m2.git` is up-to-date with `main` at commit `a6eaf5a`.
- **Working tree:** clean.
- **gitea-com:** push remains blocked by cloud authentication; not relevant to local M3 validation.

Recent commits on `main`:
```
2078f6e security(backstage): move permission.enabled=false to app-config.local.yaml
3052dc5 security(backstage): move dev-only auth flags to app-config.local.yaml
6985be3 security(backstage): add resolutions for axios and undici
15a40cd security(backstage): force @grpc/grpc-js ^1.14.4 and ws ^8.21.0 via resolutions
e252515 chore(backstage): default BACKSTAGE_APP_HOST to localhost
1b4ba50 chore(backstage): add restart-backstage.sh convenience script
ebec46c feat(backstage): add Platform angle tab to Component entity page
fcaab53 refactor(backstage): avoid card duplication by keeping only overview card on Overview
14dcfcf fix(backstage): add openchoreo group to catalog
0b6211e fix(backstage): repair guest sign-in and add entity-page tabs
d25139c fix(backstage): use EntityCardBlueprint.make for openchoreo cards; verify cards render
79bf4f2 feat(m3): add live trace-ingestion assertion to smoke-m3.sh; update TODO
2655ed1 fix(m3): align namespace predictor with OpenChoreo, fix Backstage card types, default env to development
164b20e feat(m3): OTEL hardening, namespace predictor, score2openchoreo extra-env, live SigNoz install
```

---

## 3. What was built / proved this session

### Namespace predictor alignment

- `tools/namespace-predictor/main.go` rewritten to mirror `openchoreo/internal/dataplane/kubernetes/name.go` + `namespace_handler.go`.
- `backstage/packages/app/src/modules/openchoreo-cards/namespace-predictor.ts` updated to the same algorithm and verified against the Go binary.
- Updated `scripts/smoke-m3.sh`, `scripts/preflight-m3.sh`, and docs to use environment `development` (the live cluster value) instead of `dev`.

### hello-m2 OTEL hardening

- `seed-repos/hello-m2/main.go` now sets resource attributes: `service.name`, `service.version`, `openchoreo.project`, `openchoreo.component`, `openchoreo.environment`, `openchoreo.runtime_namespace`, `git.commit.sha`.
- `seed-repos/hello-m2/.gitea/workflows/ci.yaml` computes the predicted namespace via the Go predictor and passes all telemetry/OpenChoreo variables via `score2openchoreo --extra-env`.

### score2openchoreo extension

- Added `--extra-env KEY=VALUE` flag to `tools/score2openchoreo/cli.go` for deployment-time environment injection without Score schema changes.

### Backstage cards fix

- Removed unused `React` imports and `MAX_NAME_LENGTH`.
- Converted raw component exports to `convertLegacyEntityCardExtension(...)` extension definitions in `index.ts`.
- Changed default environment fallback from `dev` to `development` in all cards.
- `yarn tsc` passes.

### SigNoz + OTEL Collector install

- Used `observability/signoz/values.local.yaml` and `observability/otel/collector-values.local.yaml`.
- Worked around SigNoz enterprise collector OpAMP issue by patching the Deployment to remove the manager config argument.
- Verified the standalone collector forwards to `signoz-otel-collector.signoz.svc.cluster.local:4317`.

### Post-deploy cost artifact

- `scripts/ci/commit-cost-artifact.sh` commits the rendered artifact to `platform-config`.
- `seed-repos/hello-m2/.gitea/workflows/ci.yaml` generates the artifact on every push.
- `CostCard.tsx` links to the real artifact in `platform-config`.
- `smoke-m3.sh` validates artifact presence via the Gitea API.
- Live run #30 succeeded; artifact exists at `cost-artifacts/hello-m2/development/latest.json`.

### Multi-angle entity page layout

- New module `backstage/packages/app/src/modules/openchoreo-entity-page/index.tsx` adds Deployment, Observability, Cost, Policy, and Platform tabs for Component entities.
- `App.tsx` registers the module.
- Playwright verification confirms all tabs render on `http://localhost:3001/catalog/default/component/hello-m2`.
- The four dedicated-tab cards are no longer duplicated on the Overview grid.

### Gitea catalog provider / discovery

- Configured `@backstage/plugin-catalog-backend-module-gitea` provider in `backstage/app-config.yaml` to scan the `openchoreo` org on `localhost:3333`.
- Added a second Gitea integration for `localhost:3002` because Gitea returns raw catalog-info URLs on its internal ROOT_URL port.
- Updated `scripts/start-backstage.sh` to ensure both `3333:3000` and `3002:3000` port-forwards to `svc/gitea-http` are active before the dev server starts.
- `hello-m2` and `developer-portal` are now auto-imported; relations resolve correctly.

### Smoke harness catalog assertions

- `scripts/smoke-m3.sh` now verifies the Backstage backend API is reachable and that `component/default/hello-m2` and `component/default/developer-portal` are present in the catalog.
- Checks that `hello-m2` carries the `openchoreo.dev/*` annotations used by the entity cards and that its relations resolve to `group:default/openchoreo`.

### Backstage persistent dev database

- `backstage/app-config.local.yaml` now uses a file-backed `better-sqlite3` database directory at `~/.rational-reserve/backstage-db` instead of the in-memory database configured in `app-config.yaml`.
- Catalog, search, auth, and plugin state now survive dev-server restarts.
- `backstage/app-config.local.yaml.example` is tracked; `scripts/start-backstage.sh` copies it to `app-config.local.yaml` on first run so a fresh checkout starts with the correct local overrides.

### Backstage guest sign-in / catalog fix

- `backstage/app-config.yaml` now allows both `http://localhost:3001` and `http://127.0.0.1:3001` in `backend.cors.origin`.
- `scripts/start-backstage.sh` only overrides `backend.cors.origin` when `BACKSTAGE_APP_HOST` is explicitly set, uses `nohup`/`disown` so the backend survives SIGHUP, and pins Node 24 via PATH.
- Guest sign-in now works and the catalog loads from either `localhost:3001` or `127.0.0.1:3001`.
- Added `group:default/openchoreo` to `backstage/examples/org.yaml` to eliminate the entity-relations warning.

### Backstage auth hardening

- `scripts/smoke-m3.sh` now obtains a guest token from `/api/auth/guest/refresh` and sends it as a Bearer token for catalog API calls.
- This allowed removal of `dangerouslyDisableDefaultAuthPolicy` and `dangerouslyAllowOutsideDevelopment` from `app-config.local.yaml.example`; the default auth policy is now active in local dev.
- `yarn tsc`, `smoke-m3.sh` (22/22), and the Playwright guest-sign-in test all pass with the hardened config.

### Backstage production config template

- Added `backstage/app-config.production.yaml` with env-var-driven PostgreSQL connection, backend auth secret, disabled guest provider, and enabled permission framework.
- Keeps secrets out of git and gives a clear path for deploying Backstage beyond local dev.

### Gitea OAuth setup helper

- Added `scripts/setup-gitea-oauth.sh` to create the local Gitea OAuth app for Backstage sign-in and store `client_id`/`client_secret` under `~/.rational-reserve/backstage-oauth-client-{id,secret}` with `chmod 600`.
- The script is idempotent: it reports the existing app if one is already present.

### M4 cost visibility plane (OpenCost + Prometheus)

- Added the M4 cost visibility spec triad under `docs/specs/2026-06-30-M4-Cost-Visibility-*`.
- Added `iac/modules/cost/` with OpenTofu-managed Helm releases for Prometheus 29.13.0 and OpenCost 2.5.25 in namespace `opencost`.
- Added `scripts/install-m4.sh`, `scripts/teardown-m4.sh`, and `scripts/smoke-m4.sh`.
- Deployed the stack on k3d-openchoreo; `scripts/smoke-m4.sh` passes and `/model/allocation` returns live namespace-level cost data.
- Added `/api/proxy/opencost` to `backstage/app-config.yaml` and updated the CostCard to fetch and display the live allocation total for the predicted runtime namespace.
- `scripts/start-backstage.sh` now ensures the OpenCost port-forward (`localhost:29003 -> svc/opencost:9090`) is active before the dev server starts.
- `scripts/smoke-m3.sh` continues to pass 22/22 with OpenCost installed.

### M4 networking (Envoy Gateway ingress)

- Added `docs/specs/2026-06-30-M4-Networking-Requirements.md`, `docs/specs/2026-06-30-M4-Networking-Design-Specification.md`, and `docs/specs/2026-06-30-M4-Networking-Technical-Specification.md`.
- Added `iac/modules/networking/` (Envoy Gateway Helm, GatewayClass, Gateway, EnvoyProxy NodePort config, HTTPRoutes) and wired it into root `iac/main.tf`.
- Added `scripts/install-m4-networking.sh`, `scripts/teardown-m4-networking.sh`, `scripts/smoke-m4-networking.sh`, and `scripts/update-local-hosts.sh`.
- Deployed Envoy Gateway on k3d-openchoreo; `scripts/smoke-m4-networking.sh` passes HTTP 200 for `gitea.local`, `signoz.local`, and `opencost.local`.
- Cilium as the CNI remains a documented fresh-cluster rebuild path rather than an in-place Flannel replacement.

### Backstage production runtime

- Added the spec triad `docs/specs/2026-06-30-Backstage-Production-Runtime-*`.
- Added `iac/modules/postgres/` to deploy PostgreSQL in the `backstage` namespace with a NodePort service and a Terraform-generated password stored in a Kubernetes Secret.
- Added `scripts/install-backstage-production.sh`, `scripts/teardown-backstage-production.sh`, `scripts/start-backstage-production.sh`, `scripts/stop-backstage-production.sh`, and `scripts/smoke-backstage-production.sh`.
- `start-backstage-production.sh` sets `NODE_ENV=production`, loads `app-config.production.yaml`, forwards PostgreSQL to a local port, and runs the built backend on port 7009 with guest disabled and Gitea auth enabled.
- `smoke-backstage-production.sh` validates the production backend returns HTTP 200.

### Backstage Gitea authentication provider

- Added the spec triad `docs/specs/2026-06-30-Backstage-Gitea-Auth-Provider-*` per project governance.
- Implemented backend module `backstage/packages/backend/src/modules/giteaAuth.ts` using `createOAuthAuthenticator` and `createOAuthProviderFactory`; it exchanges the authorization code with Gitea, fetches `/api/v1/user`, and issues a Backstage user token mapped to `user:default/<gitea-login>` with `group:default/openchoreo` ownership.
- Implemented frontend module `backstage/packages/app/src/modules/giteaSignIn.tsx` with a custom `giteaAuthApiRef`, `ApiBlueprint`-registered `OAuth2` implementation, and a `SignInPageBlueprint` that adds a Gitea option alongside guest sign-in.
- Wired the modules into `packages/backend/src/index.ts` and `packages/app/src/App.tsx`.
- Updated `app-config.local.yaml.example` and `app-config.production.yaml` with Gitea provider blocks, and updated `scripts/start-backstage.sh` to export `GITEA_OAUTH_CLIENT_ID`/`GITEA_OAUTH_CLIENT_SECRET` from `~/.rational-reserve/backstage-oauth-client-*`.
- Added `scripts/smoke-auth.sh` and included it in `scripts/smoke-all.sh`; it validates that `/api/auth/gitea/start` redirects to the local Gitea OAuth authorize URL.

### Unified smoke validation

- Added `scripts/smoke-all.sh` to run AUTH, M2, M3, and M4 smoke suites end-to-end.
- Made `scripts/smoke-infracost.sh` skip gracefully when no local `INFRACOST_API_KEY` is configured, avoiding a false failure in local dev.
- Reseeded OpenBao so `scripts/smoke-openbao.sh` passes.
- `scripts/smoke-all.sh` now reports `ALL SMOKE SUITES PASSED (AUTH, M2, M3, M4)`.

### Entity-page tab polish

- Removed the four dedicated-tab cards from the Overview grid in `openchoreo-cards/index.tsx`; only the `OpenChoreo Overview` card remains on Overview.
- Verified via Playwright that the Deployment, Policy, Observability, Cost, and Platform cards render only inside their dedicated tabs.

### Dependency audit completion

- Added Yarn resolutions in `backstage/package.json` for `@grpc/grpc-js ^1.14.4`, `ws ^8.21.0`, `axios ^1.18.1`, `undici ^7.28.0`, and `react-router ^6.30.4`, clearing all high/critical advisories.
- `yarn npm audit --all` now reports only the moderate `@material-ui/core` v4 deprecation warning, which Backstage itself still depends on; resolving it requires a coordinated Backstage version upgrade.

### Auth hardening

- Moved `backend.auth.dangerouslyDisableDefaultAuthPolicy`, `auth.providers.guest.dangerouslyAllowOutsideDevelopment`, and `permission.enabled=false` from `app-config.yaml` to a new `app-config.local.yaml`.
- `app-config.yaml` no longer contains dev-only dangerous auth/permission flags, keeping production config clean.
- Backstage dev server still loads the local overrides automatically and guest sign-in continues to work.

### Live smoke cycle

- `./scripts/smoke-m3.sh` passes 22/22 (added live Backstage catalog entity checks for `hello-m2` and `developer-portal`).
- `./scripts/preflight-m3.sh` runs successfully.
- Manual ClickHouse query confirms trace ingestion with correct resource attributes.

---

## 4. What is NOT yet done

### gitea-com push

External `gitea-com` push is still blocked by cloud authentication. Local Gitea has current state.

### Backstage catalog live render verification

Done. Guest sign-in works and all five OpenChoreo cards plus the new Deployment, Policy, Observability, and Cost entity-page tabs render on `http://localhost:3001/catalog/default/component/hello-m2`.

### iac/modules/observability/

Done 2026-06-30. `iac/modules/observability/` exists and is wired into `install-m3.sh` / `teardown-m3.sh` via OpenTofu.

### Backstage dependency audit remediation

Done 2026-06-30. All high/critical advisories are resolved; only the moderate `@material-ui/core` v4 deprecation warning remains.

---

## 5. Live state at handoff

- **k3d-openchoreo cluster:** healthy.
- **Gitea local port state:** port-forwards `localhost:3333 -> gitea-http:3000` and `localhost:3002 -> gitea-http:3000` should be running. `scripts/start-backstage.sh` ensures them automatically; if needed, recreate with `kubectl --context k3d-openchoreo -n gitea port-forward svc/gitea-http 3333:3000 &` and the same for `3002:3000`.
- **SigNoz:** namespace `signoz` exists; frontend service `signoz` exists; OTLP receiver on `signoz-otel-collector.signoz.svc.cluster.local:4317/4318`.
- **OTEL collector:** namespace `otel-system`; forwards to SigNoz.
- **hello-m2 workload:** running in `dp-default-default-development-f8e58905` at image tag `a6eaf5a`.
- **platform-config:** contains the rendered `hello-m2` Component/Workload for `development`.

---

## 6. Skills / agents to reach for in the next session

- `webapp-testing` for Backstage card rendering verification.
- Standard Go test/build loop for `tools/namespace-predictor` and `tools/score2openchoreo`.
- `./scripts/smoke-m3.sh` as the acceptance gate for any M3 change.

---

## 7. What to do first in the next session

In this exact order:

1. Read this file.
2. Read `TODO.md`.
3. Read `PROJECT_SUMMARY.md`.
4. `git status` and `git log --oneline origin/main..HEAD` to verify state.
5. Confirm cluster health: `kubectl --context k3d-openchoreo get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded`.
6. Run `./scripts/smoke-all.sh` to confirm the live AUTH + M2/M3/M4 + BACKSTAGE-PRODUCTION smoke cycle still passes.
7. Review `TODO.md` "Next candidate priorities" and ask the user which to tackle next. Remaining backlog is primarily containerizing Backstage in-cluster, adding a reverse proxy/TLS, or the Cilium fresh-cluster rebuild.

---

## 8. State of the three projects in one line each

- **openchoreo** (`/Users/nnos/Projects/openchoreo/`): unchanged, cluster healthy, used as reference for namespace algorithm and CRD shapes.
- **rational-reserve** (`/Users/nnos/Projects/rational-reserve/`): unchanged this session.
- **developer-portal** (`/Users/nnos/Projects/developer-portal/`): M3 Production Multi-Angle Visibility, M4 cost visibility, and Backstage Gitea auth provider installed and smoke-validated on k3d-openchoreo (`smoke-all.sh` passes AUTH/M2/M3/M4); next step is user-prioritized from TODO.md candidates.
