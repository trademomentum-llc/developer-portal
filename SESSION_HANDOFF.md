# SESSION HANDOFF

> Read this file FIRST in the next session. It tells you where we stopped,
> what is now committed that was not before, what is still outstanding, and
> exactly what to do first.

**Last updated:** 2026-06-30
**Reason for handoff:** M3 Production Multi-Angle Visibility live install and full-spectrum smoke cycle completed successfully on k3d-openchoreo. Working tree is clean after the final commit.

---

## 1. The single most important thing

M3 is now live and validated end-to-end:

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
- `./scripts/smoke-m3.sh` now passes 13/13 checks, including a live trace-ingestion assertion.
- Backstage `yarn tsc` passes and the five OpenChoreo entity cards render on the live `hello-m2` catalog page after converting them to `EntityCardBlueprint.make` extension definitions (the initial `convertLegacyEntityCardExtension` attempt failed because the plain card components lacked legacy extension metadata).
- Backstage dev ports moved from `3000/7007` to `3001/7008` in `app-config.yaml` and `playwright.config.ts` to avoid the Gitea service on port 3000.
- `catalog-info.yaml` root System description folded to a `>-` block scalar to avoid the `Option C:` YAML parse error, and `openchoreo.dev/system` annotation quoted as a string to satisfy the Backstage envelope policy.
- `iac/modules/observability/` created for repeatable SigNoz + OTEL Collector installs; `install-m3.sh` now applies it via OpenTofu; `tofu plan -target=module.observability` shows a clean 3-to-add plan.
- `./scripts/teardown-m3.sh` updated to destroy the observability module via OpenTofu.
- `./scripts/smoke-m3.sh` passes 13/13 live checks.

The namespace predictor (Go + TypeScript) is now a byte-for-byte semantic replica of OpenChoreo's `GenerateK8sNameWithLengthLimit(63, "dp", ...)` algorithm, with the canonical vector `dp-default-default-development-f8e58905` verified against the live cluster.

---

## 2. Git state at handoff

- **Branch:** `main`
- **Local HEAD:** `79bf4f2` -- `feat(m3): add live trace-ingestion assertion to smoke-m3.sh; update TODO`
- **origin (local Gitea):** `http://localhost:3333/openchoreo/developer-portal.git` is up-to-date with `main`.
- **hello-m2 (local Gitea):** `http://localhost:3333/openchoreo/hello-m2.git` is up-to-date with `main` at commit `a6eaf5a`.
- **Working tree:** clean.
- **gitea-com:** push remains blocked by cloud authentication; not relevant to local M3 validation.

Recent commits on `main`:
```
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

### Live smoke cycle

- `./scripts/smoke-m3.sh` passes 13/13.
- `./scripts/preflight-m3.sh` runs successfully.
- Manual ClickHouse query confirms trace ingestion with correct resource attributes.

---

## 4. What is NOT yet done

### gitea-com push

External `gitea-com` push is still blocked by cloud authentication. Local Gitea has current state.

### Backstage catalog live render verification

The cards typecheck and the catalog locations are configured. A headless browser test signed in as Guest and confirmed all five card titles (`OpenChoreo Context`, `Observability`, `Cost`, `Policy`, `Deployment`) render on `http://localhost:3000/catalog/default/component/hello-m2`.

### iac/modules/observability/

SigNoz and the OTEL collector are currently installed via Helm commands, not via OpenTofu. The next repeatable-infrastructure step is to create `iac/modules/observability/` and wire it into `iac/main.tf`.

### Backstage dependency audit remediation

`yarn npm audit --all --recursive` still reports existing critical/high transitive advisories. No new dependencies were introduced.

---

## 5. Live state at handoff

- **k3d-openchoreo cluster:** healthy.
- **Gitea local port state:** port-forward `localhost:3333 -> gitea-http:3000` should be running. If not, recreate with `kubectl --context k3d-openchoreo -n gitea port-forward svc/gitea-http 3333:3000 &`.
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
6. Run `./scripts/smoke-m3.sh` to confirm the live smoke cycle still passes.
7. Pick the next production-model slice from `TODO.md` items 7 and 8:
   - Persist post-deploy Infracost cost artifact on every `hello-m2` push and wire CostCard to the real artifact.
   - Create a dedicated Backstage multi-angle entity page layout with tabs/sections.

---

## 8. State of the three projects in one line each

- **openchoreo** (`/Users/nnos/Projects/openchoreo/`): unchanged, cluster healthy, used as reference for namespace algorithm and CRD shapes.
- **rational-reserve** (`/Users/nnos/Projects/rational-reserve/`): unchanged this session.
- **developer-portal** (`/Users/nnos/Projects/developer-portal/`): M3 Production Multi-Angle Visibility installed and smoke-validated on k3d-openchoreo; Backstage cards typecheck and are wired to the live predictor; next steps are IaC module and live UI verification.
