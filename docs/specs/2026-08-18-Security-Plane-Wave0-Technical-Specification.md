# Technical Specification: Security Plane Wave 0

**Document ID:** SEC-PLANE-WAVE0-TECH-001
**Version:** 0.1
**Date:** 2026-08-18
**Status:** Approved-basis implementation specification -- implements the approved Wave 0 scope of SEC-PLANE-PULLFORWARD-REQ-001 under the 2026-08-18 Tier-3 directive
**Predecessors:** SEC-PLANE-PULLFORWARD-REQ-001 (`docs/specs/2026-08-18-Security-Plane-Pull-Forward-Requirements.md`, approved-by-directive), RECORD-IMMUTABILITY-TECH-001 (`docs/specs/2026-08-18-Record-Immutability-Technical-Specification.md`, section 9 chaining sketch), M1-M4 milestone triads

---

## Purpose, evidence discipline, and traceability

This document is the implementation-grade companion to the Security Plane Pull-Forward Requirements. It specifies exact change points, file contents, commands, and tests for Wave 0 (FR-01 through FR-11). It changes no code by itself; implementation follows the lanes in section 12.

**Evidence discipline.** Every code-level claim below was verified against the live source or the live cluster on 2026-08-18 and carries a `file:line` citation or a `verified live` note with the command used. Where a value could not be resolved reliably, this document specifies the exact resolution procedure instead of inventing the value (procedure-over-fabrication).

**Verified environment facts used below.**

- Host: Colima VM 2 vCPU / 3.9 GB RAM, kernel 6.8.0-100-generic aarch64; k3d single node, k3s v1.32.9+k3s1. Wave 0 budget: no new standing cluster workloads.
- Gatekeeper (deployed): pods `gatekeeper-audit` x1 and `gatekeeper-controller-manager` x3 in `gatekeeper-system`, containers expose ports 8888 and 9090 (`kubectl --context k3d-openchoreo -n gatekeeper-system get deploy -o jsonpath=...`, run 2026-08-18); the only Service is `gatekeeper-webhook-service` 443 -- no metrics Service exists. Constraint `c1-enforce` reports `.status.totalViolations: 0` (live). Audit pod logs are JSON with `process:"audit"` and `event_type` fields (`audit_started`/`audit_finished` observed; `violation_audited` is the violation event type per Gatekeeper docs -- labeled docs-sourced, not observed, because no violations exist to observe).
- Trivy image digest resolved via Docker Hub registry API on 2026-08-18 (procedure in section 1.5): `aquasec/trivy:0.74.0` -> `sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969`; the index includes `linux/arm64` (runner-safe).
- OSV-Scanner image digest resolved via ghcr.io registry API on 2026-08-18 (procedure in section 2.4): `ghcr.io/google/osv-scanner:v2.5.1` -> `sha256:8108ae94eadea5a02c9bec6e646909d5b790b44bd62d7f5b7f0b1d6d0ffc7734`; the index includes `linux/arm64`.
- `github/codeql-action` v4.37.7 peeled commit `ff2f1c621b7f889edc0d3c761ac2e6a3f8cdb0dd` resolved via `git ls-remote` on 2026-08-18 (procedure in section 10.3).
- Backstage kubernetes proxy mechanics verified in the installed dependency tree: `LocalKubectlProxyLocator` exists (`backstage/node_modules/@backstage/plugin-kubernetes-backend/dist/cluster-locator/LocalKubectlProxyLocator.cjs.js`, default URL `localhost:8001`); the proxy selects clusters via the `Backstage-Kubernetes-Cluster` header and authorizes against `kubernetes.proxy` (`dist/service/KubernetesProxy.cjs.js`); permission names `kubernetes.proxy` / `kubernetes.resources.read` / `kubernetes.clusters.read` and export names `kubernetesProxyPermission` / `kubernetesResourcesReadPermission` / `kubernetesClustersReadPermission` verified in `@backstage/plugin-kubernetes-common/dist/index.d.ts`.
- FR-08 import paths verified: `createConditionalDecision`, `PermissionPolicy`, `PolicyQuery`, `PolicyQueryUser` are root exports of `@backstage/plugin-permission-node` (`dist/index.d.ts`); `policyExtensionPoint` is alpha-only -- declared at `@backstage/plugin-permission-node/dist/alpha.d.ts:17-19` and served through the package's `./alpha` export subpath (verified in the installed `package.json` exports); `catalogEntityReadPermission` / `Create` / `Delete` / `Refresh` in `@backstage/plugin-catalog-common/dist/alpha.cjs.js`; `catalogConditions.isEntityOwner` in `@backstage/plugin-catalog-backend/dist/alpha.d.ts`.

**Traceability map (section -> implements):**

| Section | Implements (FR) | Requirements decisions / NFRs |
|---|---|---|
| 1. Trivy in CI | FR-01 | D1, D4; NFR-04 |
| 2. OSV-Scanner in CI | FR-02 | D1, D4; NFR-04 |
| 3. Scan artifacts to platform-config | FR-03 | D1; NFR-03, NFR-08 |
| 4. Security tab | FR-04 | D1; NFR-03, NFR-06, NFR-07 |
| 5. PolicyCard live wiring | FR-05 | D1; NFR-03 |
| 6. gatekeeper_violations metric | FR-06 | D1 |
| 7. Gatekeeper audit logs to SigNoz | FR-07 | D1, D3; NFR-03 |
| 8. RBAC permission policy | FR-08 | D1; NFR-03; resolves OQ-25 |
| 9. TLS on .local gateways | FR-09 | D1; resolves OQ-24 |
| 10. dependabot + code scanning | FR-10 | D1, D4 |
| 11. Guard audit-log hash chaining | FR-11 | D1; NFR-08; TECH-001 section 9 |
| 12. Implementation lanes | all | NFR-06, NFR-10 |
| 13. Provenance duty | all | NFR-01, NFR-05 |
| 14. Test and acceptance plan | all | NFR-03, NFR-06 |

---

## 1. FR-01 -- Trivy vulnerability scanning in CI

### 1.1 Goal

The canonical app pipeline scans the repository filesystem (pre-build) and the built image (post-build, pre-push) with Trivy, fails at the HIGH,CRITICAL severity threshold, and emits JSON + SARIF reports for the artifact step (section 3).

### 1.2 Current state (verified)

- `seed-repos/hello-m2/.gitea/workflows/ci.yaml` has no scanning of any kind: checkout at `:13`, Score validation at `:26-27`, image build at `:70-72`, image push (push-only) at `:74-76`. All third-party actions are SHA-pinned today (`:13`, `:16`, `:31`).
- `iac/templates/ci.yaml` is the drifted template copy (uses `--environment dev` at `:68`, lacks the OTEL/cost steps; five-plane ORC-G9). This specification adds the scan steps to it at the same logical positions and does NOT otherwise reconcile the template drift (out of Wave 0 scope, recorded here so the implementer does not "fix" it opportunistically).
- The act-runner environment provides a working Docker daemon (the existing `docker build`/`docker push` steps at `:70-76` prove it), so digest-pinned image consumption via `docker run` needs no new tooling.

### 1.3 Consumption method (pinned by digest)

Trivy runs as the pinned image `aquasec/trivy@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969` (tag `0.74.0`, index digest resolved 2026-08-18; index includes `linux/arm64`, required because the act-runner is aarch64).

Pin rationale: the March 2026 Trivy supply-chain compromise (CVE-2026-33634, force-pushed malicious `trivy-action` tags) is the cited reason mutable tags are prohibited (REQ-001 D4). The `trivy-action` GitHub Action is deliberately NOT used; `docker run` against the digest-pinned image is the consumption method.

### 1.4 Exact change points

**File 1: `seed-repos/hello-m2/.gitea/workflows/ci.yaml`**

Insert two steps immediately AFTER `Validate Score schema` (`:26-27`) and BEFORE `Set up OpenTofu` (`:29`) -- the fs scan is pre-build:

```yaml
      - name: Trivy filesystem scan (gate)
        run: |
          docker run --rm -v "$PWD:/src" aquasec/trivy@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969 \
            fs --severity HIGH,CRITICAL --exit-code 1 \
            --format json -o /src/trivy-fs.json /src

      - name: Trivy filesystem report (SARIF artifact)
        if: always()
        run: |
          docker run --rm -v "$PWD:/src" aquasec/trivy@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969 \
            fs --severity HIGH,CRITICAL --exit-code 0 \
            --format sarif -o /src/trivy-fs.sarif /src
```

Insert two steps immediately AFTER `Build image` (`:70-72`) and BEFORE `Push image` (`:74-76`) -- the image scan is post-build, pre-push, and therefore runs on pull requests too (the build step has no event filter):

```yaml
      - name: Trivy image scan (gate)
        run: |
          docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD:/src" \
            aquasec/trivy@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969 \
            image --severity HIGH,CRITICAL --exit-code 1 --ignorefile /src/.trivyignore \
            --format json -o /src/trivy-image.json \
            registry.local-registry.svc.cluster.local:5000/hello-m2:${GITHUB_SHA::7}

      - name: Trivy image report (SARIF artifact)
        if: always()
        run: |
          docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD:/src" \
            aquasec/trivy@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969 \
            image --severity HIGH,CRITICAL --exit-code 0 --ignorefile /src/.trivyignore \
            --format sarif -o /src/trivy-image.sarif \
            registry.local-registry.svc.cluster.local:5000/hello-m2:${GITHUB_SHA::7}
```

Behavior notes:

- The gate is Trivy's native `--exit-code 1` with `--severity HIGH,CRITICAL`. `--ignore-unfixed` is deliberately NOT set: REQ-001 FR-01 names the reviewed suppressions list as the only bypass, and unfixed findings must remain visible in the committed JSON report (section 3).
- Each gate step writes its JSON report regardless of exit code (Trivy writes the report before applying the exit code), so the artifact step (section 3) has its inputs even on failure; the SARIF steps are `if: always()` for the same reason.
- The docker socket mount is required for `image` mode because the image under test exists only in the runner's daemon (it is built at `:70-72` and not yet pushed).

**File 2: `iac/templates/ci.yaml`**

Same four steps at the same logical positions: fs steps after its `Validate Score schema` (`:26-27`); image steps between `Build image` (`:55-57`) and `Push image` (`:59-61`).

**File 3: `seed-repos/hello-m2/.trivyignore` (new)** -- the suppression file, one CVE or advisory ID per line with a dated review comment above each entry:

```
# Suppressions require: reviewer, date, reason, and re-review date.
# reviewed-by: <name>  reviewed: 2026-MM-DD  re-review: 2026-MM-DD+N
# CVE-2026-XXXXX  reason: <why this does not affect the artifact>
```

Trivy reads `.trivyignore` from the scan root automatically for `fs`; the `image` steps pass it explicitly via `--ignorefile` (above). Entries without the dated comment header are a review failure, not a scanner failure -- the convention is enforced in code review and documented in the file header.

### 1.5 Digest resolution and re-verification procedure

The digests in this document were resolved on 2026-08-18 with:

```sh
TOKEN=$(curl -fsS "https://auth.docker.io/token?service=registry.docker.io&scope=repository:aquasec/trivy:pull" | jq -r .token)
curl -fsSI -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json" \
  https://registry-1.docker.io/v2/aquasec/trivy/manifests/0.74.0 | grep -i docker-content-digest
```

At implementation time the implementer re-runs this procedure and records the result in the commit message. If the tag resolves to a DIFFERENT digest than the one in this document, the implementer stops and re-decides with the user which digest to pin -- a moved tag is exactly the CVE-2026-33634 attack shape, and silently following it is prohibited.

### 1.6 Resource footprint

Zero standing workloads. Two to four short `docker run` invocations inside the existing CI job pod per run (ephemeral, burst-only; within the Wave 0 budget, REQ-001 section 5).

### 1.7 Verification and smoke hook

- Negative test (gate works): on a scratch branch, add a dependency with a known HIGH/CRITICAL advisory to `seed-repos/hello-m2/go.mod` (or temporarily pin a known-vulnerable base image in the `Dockerfile`) and push; the run must fail at the Trivy gate step and the SARIF step must still produce reports. Revert the scratch change.
- Positive test: a normal push passes the gate and produces `trivy-fs.json`, `trivy-fs.sarif`, `trivy-image.json`, `trivy-image.sarif` in the job workspace.
- Smoke hook (section 14): `scripts/smoke-security.sh` asserts (a) the four scan steps exist in both workflow files, (b) every `aquasec/trivy` reference is digest-pinned (no bare tags), and (c) after any push, `security-artifacts/hello-m2/development/latest.json` exists in platform-config (section 3).

**Traces to:** FR-01; SEC-G2; D1, D4; NFR-04.

---

## 2. FR-02 -- OSV-Scanner dependency scanning in CI

### 2.1 Goal

Advisory-precise dependency scanning of the repo's lockfiles (Go modules; `yarn.lock` when a repo has one) in the same pipeline, same threshold policy as FR-01, JSON output feeding the section 3 artifact.

### 2.2 Current state (verified)

- No dependency scanning exists in either workflow file (section 1.2). The local publication gate runs `yarn/npm audit` and `govulncheck` inside verify-guard (`plugins/rr-policy-guards/tools/verify-guard/exec.go:96-169`) -- that is an agent-side gate, not CI; this FR adds the forge-side equivalent for app repos.
- `seed-repos/hello-m2` contains a `go.mod` (verified in the repo-wide `go.mod` enumeration: 8 roots -- the six guards, hello-m2, score2openchoreo).

### 2.3 Consumption method (pinned by digest)

`ghcr.io/google/osv-scanner@sha256:8108ae94eadea5a02c9bec6e646909d5b790b44bd62d7f5b7f0b1d6d0ffc7734` (tag `v2.5.1`, index digest resolved 2026-08-18; index includes `linux/arm64`). Resolution procedure:

```sh
TOKEN=$(curl -fsS "https://ghcr.io/token?scope=repository:google/osv-scanner:pull" | jq -r .token)
curl -fsSI -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json" \
  https://ghcr.io/v2/google/osv-scanner/manifests/v2.5.1 | grep -i docker-content-digest
```

The same re-verification rule as section 1.5 applies at implementation time.

### 2.4 Exact change points

**`seed-repos/hello-m2/.gitea/workflows/ci.yaml`** and **`iac/templates/ci.yaml`**: insert immediately AFTER the Trivy fs steps (section 1.4), still pre-build:

```yaml
      - name: OSV dependency scan (gate)
        run: |
          docker run --rm -v "$PWD:/src" ghcr.io/google/osv-scanner@sha256:8108ae94eadea5a02c9bec6e646909d5b790b44bd62d7f5b7f0b1d6d0ffc7734 \
            scan source -r /src --format json --output /src/osv.json
```

Gate semantics (a deliberate strengthening, recorded as a spec decision): OSV-Scanner exits non-zero when any unignored vulnerability is found, and the step gates on that exit code. This is STRICTER than FR-01's HIGH,CRITICAL threshold. Reasoning: OSV severity metadata is advisory-dependent and not uniformly populated across ecosystems, so a severity-filtered jq gate would be a heuristic over an unstable schema; a deterministic any-vulnerability gate with explicit dated suppressions is both stronger and more honest. If this proves too noisy in practice, the documented fallback is the FR-01-threshold variant: run with `--exit-code 0`-equivalent (`|| true`) and gate on a jq count of HIGH/CRITICAL severities from the report -- the fallback needs the report schema pinned to osv-scanner v2.5.1 output and is NOT the default.

**Suppression file: `seed-repos/hello-m2/osv-scanner.toml` (new)** -- same dated-review convention as `.trivyignore`:

```toml
# Suppressions require: reviewer, date, reason, and re-review date (comment above each entry).
# reviewed-by: <name>  reviewed: 2026-MM-DD  re-review: 2026-MM-DD+N
# [[IgnoredVulns]]
# id = "GHSA-xxxx-yyyy-zzzz"
# reason = "<why this does not affect the artifact>"
# expiry = 2026-MM-DD
```

### 2.5 Relationship to Trivy (dedupe note)

Overlap is expected and accepted: both scanners report and both appear in the section 3 artifact. OSV-Scanner owns language-ecosystem advisory precision (Go modules, npm/yarn); Trivy owns OS/distro packages, IaC, secrets, and the built image. No deduplication is performed in Wave 0; the Security tab (section 4) labels each finding with its source scanner.

### 2.6 Resource footprint

Zero standing workloads; one short `docker run` per CI run (burst-only).

### 2.7 Verification and smoke hook

- Negative test: the scratch-branch vulnerable dependency from section 1.7 must also fail this gate.
- Positive test: normal push produces `osv.json`; a suppressed ID in `osv-scanner.toml` is honored (add a real past finding's ID, observe the gate pass, remove it).
- Smoke hook: `smoke-security.sh` asserts the step exists in both workflow files and the image reference is digest-pinned.

**Traces to:** FR-02; SEC-G2; D1, D4; NFR-04.

---

## 3. FR-03 -- Security scan artifacts committed to platform-config

### 3.1 Goal

Every push commits a combined, portal-readable security artifact into `platform-config`, following the cost-artifact precedent exactly.

### 3.2 Current state (verified)

- The precedent is `scripts/ci/commit-cost-artifact.sh`: base64 the payload (`:9`), compute the repo path (`:10`), GET the existing file SHA tolerating 404 (`:12-15`), POST on create / PUT on update (`:24-31`), authenticating as `gitea_admin` with `${GITEA_TOKEN}` (`:13`, `:27`) against `http://gitea-http.gitea.svc.cluster.local:3000/api/v1`. The CI invokes it at `ci.yaml:127-129` with `GITEA_TOKEN: ${{ secrets.PLATFORM_CONFIG_TOKEN }}` (`:130-131`). The sibling `scripts/ci/commit-to-platform-config.sh` uses the identical mechanics (`:9-33`).
- Honest note (carried from the precedent, accepted existing pattern): artifacts are committed AS `gitea_admin`. This is recorded, not silently inherited -- per-artifact authorship metadata lives inside the artifact JSON (`git_sha`, `run_id`).

### 3.3 Exact change points

**New file: `scripts/ci/commit-security-artifacts.sh`** -- same mechanics as the precedent, with two writes per invocation (per-run file + latest pointer). Contract:

```
commit-security-artifacts.sh <environment> <app> <artifact.json> <run-id>
```

- Path layout: `security-artifacts/<app>/<env>/<run-id>.json` and `security-artifacts/<app>/<env>/latest.json` (identical content; `latest.json` is what the portal reads, the per-run file is the append-only history, NFR-08).
- Per path: the tolerate-404 GET-then-POST/PUT pattern of `commit-cost-artifact.sh:12-31`.
- `set -euo pipefail`; commit message `ci: security artifact <app> <env> -> <sha7> (run <run-id>)`.

**`seed-repos/hello-m2/.gitea/workflows/ci.yaml`**: insert a step AFTER `Commit cost artifact to platform-config` (`:127-129`) and BEFORE `Commit to platform-config` (`:132-134`), push-only, assembling and committing the combined artifact:

```yaml
      - name: Assemble security scan artifact
        if: github.event_name == 'push'
        run: |
          jq -n \
            --arg component "hello-m2" \
            --arg environment "development" \
            --arg git_sha "${GITHUB_SHA::7}" \
            --arg run_id "${GITHUB_RUN_ID}" \
            --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
            --slurpfile trivy_fs trivy-fs.json \
            --slurpfile trivy_image trivy-image.json \
            --slurpfile osv osv.json \
            '{
              artifact_type: "security-scan",
              component: $component, environment: $environment,
              git_sha: $git_sha, run_id: $run_id, generated_at: $generated_at,
              gate: { severity_threshold: ["HIGH","CRITICAL"], status: "pass",
                      suppressions: [".trivyignore", "osv-scanner.toml"] },
              results: {
                trivy_fs:    { report: $trivy_fs[0] },
                trivy_image: { report: $trivy_image[0] },
                osv:         { report: $osv[0] }
              }
            }' > /tmp/security-artifact.json

      - name: Commit security artifact to platform-config
        if: github.event_name == 'push'
        run: /tmp/dp/scripts/ci/commit-security-artifacts.sh development hello-m2 /tmp/security-artifact.json "${GITHUB_RUN_ID}"
        env:
          GITEA_TOKEN: ${{ secrets.PLATFORM_CONFIG_TOKEN }}
```

- `gate.status` is `"pass"` by construction: the gate steps (sections 1-2) stop the job before this step on failure. Failed-scan evidence lives in the CI run logs; `latest.json` therefore always describes the last PASSING push, and the Security tab (section 4) says exactly that rather than implying a live gate state. This is the honest-status rule (NFR-03) applied to the artifact model.
- The same two steps go into `iac/templates/ci.yaml` at the same logical position (after its `Commit to platform-config` step at `:72-74` -- the template has no cost-artifact step, ORC-G9 drift noted in section 1.2).
- Size note: hello-m2's reports are small. Repos with large monorepo reports may need truncation or external storage later; that is a future decision, recorded here, not speculated into Wave 0.

### 3.4 Resource footprint

None (two Contents API calls per push).

### 3.5 Verification and smoke hook

- After a push: `curl -u gitea_admin:$TOKEN http://localhost:3333/api/v1/repos/openchoreo/platform-config/contents/security-artifacts/hello-m2/development/latest.json` returns 200; `<run-id>.json` exists beside it; the JSON parses and carries `artifact_type: security-scan`.
- Smoke hook: `smoke-security.sh` performs exactly this API read (token from `~/.rational-reserve/m1-gitea-token`, the same runtime secret `scripts/start-backstage.sh:92-96` uses).

**Traces to:** FR-03; SEC-G2, SEC-G10; D1; NFR-03, NFR-08; honors five-plane FR-40/NFR-05.

---

## 4. FR-04 -- Security tab on entity pages

### 4.1 Goal

A sixth entity-page tab, `Security`, rendering the component's latest scan artifact (section 3) and its Gatekeeper violation state (section 5's data path) -- live data or an explicit "not wired" state.

### 4.2 Current state (verified)

- Entity tabs are EntityContentBlueprints in `backstage/packages/app/src/modules/openchoreo-entity-page/index.tsx`: five tabs registered in the `extensions` array at `:113-119` (Deployment/Observability/Cost/Policy/Platform); `SecurityIcon` is already imported at `:7`; `componentFilter = 'kind:component'` at `:19`; the Grid layout idiom is at `:29-35`.
- The fetch idiom is CostCard's: `useApi(fetchApiRef)` (`openchoreo-cards/CostCard.tsx:17`) with `useEffect` + cancellation (`:36-63`); component/env annotations are read at `CostCard.tsx:19-24`.
- The Gitea API is reachable from the frontend through the existing proxy `/api/proxy/gitea-actions`, which rewrites to `/api/v1` and injects the admin token (`app-config.yaml:63-69`). CostCard links the cost artifact via the Gitea raw URL pattern (`CostCard.tsx:29-30`); the Security tab goes one step further and FETCHES the artifact JSON through the proxy (rendering, not just linking).
- The anti-pattern being replaced is PolicyCard's static scaffold text (`PolicyCard.tsx:50-53`): "violations ... will appear here once the M3 policy collector is wired."

### 4.3 Exact change points

**New file: `backstage/packages/app/src/modules/openchoreo-cards/gatekeeper.ts`** -- shared fetch helper used by both the Security tab and PolicyCard (written once, avoids two divergent clients):

```ts
// Fetches one Gatekeeper constraint object through the kubernetes proxy.
// Cluster name must match the clusterLocatorMethods entry in app-config.yaml (section 5.3).
export async function fetchConstraint(
  fetchFn: typeof fetch,
  kind: string,
  name: string,
): Promise<Response> {
  return fetchFn(
    `/api/kubernetes/proxy/apis/constraints.gatekeeper.sh/v1beta1/${kind}/${name}`,
    { headers: { 'Backstage-Kubernetes-Cluster': 'k3d-openchoreo-local' } },
  );
}
```

(The header name is verified against the installed proxy: `HEADER_KUBERNETES_CLUSTER = "Backstage-Kubernetes-Cluster"` in `plugin-kubernetes-backend/dist/service/KubernetesProxy.cjs.js`.)

**New file: `backstage/packages/app/src/modules/openchoreo-cards/SecurityCard.tsx`** -- follows the CostCard structure (annotations at `CostCard.tsx:19-24`; fetch+cancel idiom at `:36-63`):

- Reads `openchoreo.dev/component` and `openchoreo.dev/environment` annotations (same fallbacks as CostCard).
- Fetches the scan artifact: `GET /api/proxy/gitea-actions/repos/openchoreo/platform-config/contents/security-artifacts/${component}/${env}/latest.json`; the Gitea contents API returns base64 in `.content` -- decode with `atob` and `JSON.parse`. Renders: gate status + threshold, per-scanner finding counts by severity, `git_sha`, `generated_at`, and a link to the artifact in Gitea (raw URL pattern per `CostCard.tsx:30`).
- Fetches the three constraints via `fetchConstraint` (kinds/names in section 5.4) and renders per-constraint `totalViolations`.
- Not-wired states (NFR-03), each an explicit, labeled UI state: artifact 404 -> "No scan artifact yet -- the security pipeline has not committed one for this component"; proxy error -> the error text (CostCard's error-render idiom, `CostCard.tsx:80-86`); violations unreachable -> "Gatekeeper violations: not wired (kubernetes proxy unavailable)". No placeholder is ever styled as live data.
- Accessibility (NFR-07): the card is keyboard-navigable text/links only, uses the same MUI `InfoCard`/`Typography` structure as the sibling cards, and all status text is conveyed by words ("pass"/"fail"/"not wired"), never color alone.

**Edit: `backstage/packages/app/src/modules/openchoreo-entity-page/index.tsx`** -- add a sixth blueprint after `policyContent` (`:75-91`) and register it in `extensions` (`:113-119`):

```tsx
const securityContent = EntityContentBlueprint.make({
  name: 'security',
  params: {
    path: '/security',
    title: 'Security',
    group: 'security',
    icon: <SecurityIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <SecurityCard />
        </Grid>
      </Grid>
    ),
  },
});
```

with `import { SecurityCard } from '../openchoreo-cards/SecurityCard';` beside the existing card imports (`:13-17`).

### 4.4 Resource footprint

None (frontend code only).

### 4.5 Verification and smoke hook

- `yarn tsc` passes in `backstage/`; `yarn dev` renders the Security tab on the hello-m2 entity page with live artifact data after a push, and with the explicit not-wired state on a fresh catalog entity that has no artifact.
- Smoke hook: `smoke-security.sh` asserts the tab is registered (`grep` for `path: '/security'` in `index.tsx`) and that `SecurityCard.tsx` + `gatekeeper.ts` exist. Render-level verification stays with the existing Playwright pattern (`backstage/packages/app/e2e-tests/`), noted as a manual check in Wave 0 (the e2e suite is a one-test scaffold, five-plane ENG-G9; extending it is not Wave 0 scope).

**Traces to:** FR-04; SEC-G10, TRV-B6; D1; NFR-03, NFR-06, NFR-07.

---

## 5. FR-05 -- PolicyCard live Gatekeeper wiring

### 5.1 Goal

PolicyCard renders live constraint `.status.violations` data instead of the static scaffold.

### 5.2 Current state (verified)

- `PolicyCard.tsx:50-53` carries the scaffold text; its links (`:37-47`) are static.
- The portal reaches the cluster **nowhere** today: the kubernetes backend plugin is registered (`packages/backend/src/index.ts:64`, dependency at `packages/backend/package.json:30`) but the `kubernetes:` config block is empty (`app-config.yaml:122-123`). The frontend plugin is present too (`packages/app/package.json:30`).
- Live cluster facts (verified 2026-08-18): constraints `c1-enforce`, `c2-enforce`, `c3-enforce` exist (kinds `C1PlatformAddonsMainProtected`, `C2ScoreSchemaValid`, `C3InfracostDelta`; `policies/C1-constraint.yaml`, `C2-constraint.yaml`, `C3-constraint.yaml`); `c1-enforce.status.totalViolations` is `0` (`kubectl --context k3d-openchoreo get c1platformaddonsmainprotected c1-enforce -o jsonpath='{.status.totalViolations}'`).

### 5.3 Decision: kubernetes plugin proxy, not a raw app-config proxy

DECIDED: use the already-registered kubernetes-backend with a config cluster locator and the `localKubectlProxy` auth provider. Reasoning:

1. A raw app-config proxy (`proxy.endpoints`) cannot authenticate to the k3d API: the k3d kubeconfig uses client-certificate mTLS, which the proxy-backend cannot present; the alternative (a long-lived ServiceAccount token in config) introduces a new managed secret for no gain.
2. The kubernetes plugin is installed for exactly this purpose (five-plane ORC-G2); wiring it is config-only on the backend.
3. Its proxy already integrates the permission framework -- it authorizes every request against `kubernetes.proxy` (verified in the installed `KubernetesProxy.cjs.js`), so FR-08 (section 8) governs cluster access uniformly with everything else.
4. `localKubectlProxy` is verified present in the installed backend (`LocalKubectlProxyLocator.cjs.js`, default `localhost:8001`) and matches reality: Backstage runs on the host, where `kubectl --context k3d-openchoreo` works (all `scripts/smoke-*.sh` rely on it).

Rejected: the small app-config proxy to the k3d API (reasons 1-2 above).

### 5.4 Exact change points

**Edit: `backstage/app-config.yaml`** -- replace the empty block at `:122-123`:

```yaml
kubernetes:
  serviceLocator:
    type: 'multiTenant'
  clusterLocatorMethods:
    - type: 'config'
      clusters:
        - name: k3d-openchoreo-local
          url: http://localhost:8001
          authProvider: localKubectlProxy
```

**Edit: `scripts/start-backstage.sh`** -- add an `ensure_kubectl_proxy` function mirroring `ensure_gitea_port` (`:26-48`): pid file under `"${RUNTIME_DIR}"`, `pkill` the stale pattern, `nohup kubectl --context k3d-openchoreo proxy --port=8001 &`, readiness loop against `http://localhost:8001/api`; call it beside `ensure_gitea_port 3333`/`3002` (`:50-51`). `scripts/stop-backstage.sh` must reap this pid the same way it reaps the others (implementer reads it and mirrors the existing pattern).

**Rewrite: `backstage/packages/app/src/modules/openchoreo-cards/PolicyCard.tsx`** -- keep the existing header/links block (`:27-48`), replace the scaffold caption (`:50-53`) with live state:

- For each of `{ kind: C1PlatformAddonsMainProtected, name: c1-enforce }`, `{ kind: C2ScoreSchemaValid, name: c2-enforce }`, `{ kind: C3InfracostDelta, name: c3-enforce }`: `fetchConstraint` (section 4.3), read `.status.totalViolations` and the first 5 entries of `.status.violations` (each rendered as `kind`, `namespace`, `name`, `msg`).
- Render rule: `.status.violations` is capped by Gatekeeper (default 20) while `.status.totalViolations` is the complete count -- display "showing first N of M" when they differ (Gatekeeper audit behavior, docs-sourced).
- Unreachable proxy -> explicit "not wired" state (same rule as section 4.3).

### 5.5 Resource footprint

One host-side `kubectl proxy` process (negligible); zero cluster-side additions.

### 5.6 Verification and smoke hook

- Seed a violation: apply a `Component` in a scratch namespace WITHOUT the `pipeline.m2/score-valid=true` annotation (the C2 rego at `policies/C2-constraint.yaml` flags exactly this); wait one audit interval; `kubectl get c2scoreschemavalid c2-enforce -o jsonpath='{.status.totalViolations}'` becomes non-zero; PolicyCard renders the entry; delete the scratch namespace.
- Production-mode check: with `permission.enabled: true` (section 8), a developer-role session can read through the proxy and an unauthenticated request gets 401/403 (the proxy's `kubernetes.proxy` authorization, verified mechanism).
- Smoke hook: `smoke-security.sh` asserts the `kubernetes:` block in `app-config.yaml`, the `ensure_kubectl_proxy` function in `start-backstage.sh`, and -- in cluster mode -- that `http://localhost:8001/apis/constraints.gatekeeper.sh/v1beta1` answers.

**Traces to:** FR-05; SEC-G8, TRV-B8; D1; NFR-03. Resolves the FR-05 half of five-plane FR-24.

---

## 6. FR-06 -- gatekeeper_violations metric in the M4 Prometheus

### 6.1 Goal

The existing M4 Prometheus scrapes Gatekeeper metrics; `gatekeeper_violations` is queryable.

### 6.2 Current state (verified)

- M4 Prometheus is the prometheus-community `prometheus` chart (`iac/modules/cost/main.tf:15-17`), values `observability/cost/prometheus-values.local.yaml:1-14` -- minimal; no custom scrape config; alertmanager/pushgateway disabled (`:10-14`).
- Live cluster (verified 2026-08-18): `gatekeeper-audit` (x1) and `gatekeeper-controller-manager` (x3) containers expose port 8888; the namespace has NO metrics Service (only `gatekeeper-webhook-service` 443). Pod-role discovery is therefore required. The `gatekeeper_violations` metric is emitted by the audit pod (Gatekeeper docs; the audit pod is the only component that computes violations, consistent with section 5.2's live constraint status).

### 6.3 Exact change points

**Edit: `observability/cost/prometheus-values.local.yaml`** -- append:

```yaml
extraScrapeConfigs: |
  - job_name: gatekeeper
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [gatekeeper-system]
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_container_port_number]
        regex: "8888"
        action: keep
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
```

Apply path: the module's helm_release values change flows through the M4 lifecycle (`scripts/install-m4.sh` re-run; direct `tofu apply` is guard-blocked per repo convention).

### 6.4 Resource footprint

None (a scrape job on the existing Prometheus; Gatekeeper already exposes the port).

### 6.5 Verification and smoke hook

- Resolve the Prometheus service name (`kubectl --context k3d-openchoreo -n opencost get svc` -- procedure, not a fabricated name), port-forward it, then: `GET /api/v1/targets` shows job `gatekeeper` with 4 up endpoints (1 audit + 3 controller-manager, per the live pod count); `GET /api/v1/query?query=gatekeeper_violations` returns a result set (possibly empty-valued series while violations are zero -- after the section 5.6 seed it must report the violation).
- Smoke hook: `smoke-security.sh` asserts the `extraScrapeConfigs` block exists in the values file, and in cluster mode queries the targets endpoint for an up `gatekeeper` job.

**Traces to:** FR-06; SEC-G8; D1.

---

## 7. FR-07 -- Gatekeeper audit logs into SigNoz

### 7.1 Goal

Gatekeeper audit-pod JSON logs are collected into the existing standalone OTEL collector and forwarded to SigNoz, queryable by `event_type`.

### 7.2 Current state (verified)

- `observability/otel/collector-values.local.yaml`: `mode: deployment` (`:10`); image `otel/opentelemetry-collector-contrib:0.155.0` (`:12-14` -- the contrib distribution, which includes the `filelog` receiver); receivers are OTLP-only (`:31-40`); the logs pipeline is `receivers: [otlp]` (`:74-77`); export to SigNoz at `:56-62`. There is NO k8s log receiver today (five-plane OBS-G1).
- Live audit log shape (verified 2026-08-18, `kubectl -n gatekeeper-system logs deploy/gatekeeper-audit --tail=200`): JSON lines with `process:"audit"` and `event_type` of `audit_started`/`audit_finished`. Violation entries carry `event_type:"violation_audited"` per Gatekeeper docs (docs-sourced: no violations exist to observe live; the verification below seeds one).
- Single-node fact: k3d-openchoreo has exactly one node (capacity facts, REQ-001 section 5), so a `filelog` receiver with a hostPath mount inside the existing Deployment-mode collector sees every pod's logs. On a multi-node cluster this design must become a DaemonSet -- recorded as an explicit limit, not discovered later.

### 7.3 Exact change points

**Edit: `observability/otel/collector-values.local.yaml`** -- three additions.

1. Receiver (inside `config.receivers`, after the `otlp` block at `:32-40`):

```yaml
    filelog/gatekeeper-audit:
      include:
        - /var/log/pods/gatekeeper-system_gatekeeper-audit-*/*/*.log
      start_at: end
      operators:
        - type: json_parser
```

2. Pipeline (`:74-77`): `receivers: [otlp, filelog/gatekeeper-audit]` on the `logs` pipeline.

3. Mounts (top-level chart values, new keys):

```yaml
extraVolumes:
  - name: varlogpods
    hostPath:
      path: /var/log/pods
extraVolumeMounts:
  - name: varlogpods
    mountPath: /var/log/pods
    readOnly: true
```

Apply path: `scripts/install-m3.sh` re-run (module values change; guard-blocked direct apply per convention).

Query key in SigNoz: `process = "audit" AND event_type = "violation_audited"` (the parse note: `json_parser` promotes the JSON fields to log attributes; `event_type` is the discriminating attribute).

### 7.4 Resource footprint

One read-only hostPath mount on the existing collector pod; negligible log volume (audit emits a handful of lines per audit interval; verified shape above).

### 7.5 Verification and smoke hook

- Seed the section 5.6 violation; within two audit intervals, the SigNoz logs explorer (or the ClickHouse query pattern used by `scripts/smoke-m3.sh:337-374`) returns lines matching `event_type = "violation_audited"`; clean up the seed.
- Smoke hook: `smoke-security.sh` asserts the receiver and pipeline entries exist in the values file; the live log-query assertion rides the section 14 acceptance checklist (it needs a seeded violation, so it is a checklist step, not an unattended smoke).

**Traces to:** FR-07; SEC-G4; D1, D3; NFR-03.

---

## 8. FR-08 -- RBAC permission policy module

### 8.1 Goal

Replace the allow-all permission policy with a custom policy enforcing admin/developer/viewer roles resolved from Gitea-mapped identity claims.

### 8.2 Current state (verified)

- The framework runs the allow-all policy: import at `packages/backend/src/index.ts:48-50`, dependency at `packages/backend/package.json:33`. Enabled in production (`app-config.production.yaml:39-40`), disabled in local dev (`app-config.local.yaml.example:35-36`) so the guest flow (`:17`) stays simple.
- Identity claims: the Gitea sign-in resolver issues `sub: user:default/<login>` and `ent: [user:default/<login>, group:default/openchoreo]` (`packages/backend/src/modules/giteaAuth.ts:186-193`). Group claims therefore arrive ONLY as `group:default/openchoreo` today; the resolver has no team-membership source.
- Catalog groups in `backstage/examples/org.yaml`: `guests` (`:21-27`), `gitea_admin` (`:29-37`), `openchoreo` (`:39-47`).
- Available imports (all verified in the installed tree): `PermissionPolicy`, `PolicyQuery`, `PolicyQueryUser`, `createConditionalDecision` from `@backstage/plugin-permission-node` root, and `policyExtensionPoint` from `@backstage/plugin-permission-node/alpha` -- alpha-only, declared at `dist/alpha.d.ts:17-19` and served via the package's `./alpha` export subpath (already a direct dep, `package.json:35`); `PolicyDecision`/`AuthorizeResult` from `@backstage/plugin-permission-common` (direct dep, `:34`); `catalogEntityReadPermission` / `Create` / `Delete` / `Refresh` from `@backstage/plugin-catalog-common/alpha`; `catalogConditions.isEntityOwner` from `@backstage/plugin-catalog-backend/alpha` (direct dep, `:26`); `kubernetesProxyPermission` / `kubernetesResourcesReadPermission` / `kubernetesClustersReadPermission` from `@backstage/plugin-kubernetes-common`.

### 8.3 Decision: config-driven admin list (and why not a new catalog group)

DECIDED: admin = identities whose `user:default/<login>` appears in a config list (`permission.policy.adminUsers`, default `["gitea_admin"]`, env-overridable via `${BACKSTAGE_ADMIN_USERS}`); developer = identities carrying `group:default/openchoreo` in `ownershipEntityRefs`; viewer = any authenticated identity. The policy ALSO honors `group:default/gitea_admin` as an admin ref so a future resolver change needs no policy change.

Rejected alternative -- minting `group:default/openchoreo-admins` in `examples/org.yaml` and resolving membership in the resolver: the resolver's only data is the Gitea profile (`giteaAuth.ts:104-115`: id, login, email, full_name, avatar_url); team membership would require an extra Gitea API call on the sign-in path (latency and a new auth-path failure mode), and team-sync is the control-plane collaboration item (five-plane FR-14/OQ-09), not this wave. Adding a catalog group that nothing emits would be a fabricated status (NFR-03). The config list is deterministic, testable, and honest at single-operator scale.

### 8.4 Exact change points

**New file: `backstage/packages/backend/src/extensions/permissionsPolicyExtension.ts`**

```ts
import {
  PermissionPolicy,
  PolicyQuery,
  PolicyQueryUser,
  createConditionalDecision,
} from '@backstage/plugin-permission-node';
import {
  AuthorizeResult,
  PolicyDecision,
} from '@backstage/plugin-permission-common';
import {
  catalogEntityCreatePermission,
  catalogEntityDeletePermission,
  catalogEntityReadPermission,
  catalogEntityRefreshPermission,
} from '@backstage/plugin-catalog-common/alpha';
import { catalogConditions } from '@backstage/plugin-catalog-backend/alpha';
import {
  kubernetesClustersReadPermission,
  kubernetesProxyPermission,
  kubernetesResourcesReadPermission,
} from '@backstage/plugin-kubernetes-common';
import { Config } from '@backstage/config';

// Role model (SEC-PLANE-WAVE0-TECH-001 section 8):
//   admin     = user ref in permission.policy.adminUsers, or group:default/gitea_admin
//   developer = ownershipEntityRefs contains group:default/openchoreo
//   viewer    = any authenticated identity
// Anything not listed below is admin-only.
export class SecurityRbacPolicy implements PermissionPolicy {
  constructor(private readonly config: Config) {}

  private rolesFor(ownershipEntityRefs: string[]): {
    isAdmin: boolean;
    isDeveloper: boolean;
  } {
    const admins =
      this.config.getOptionalStringArray('permission.policy.adminUsers') ??
      ['gitea_admin'];
    const adminRefs = new Set([
      ...admins.map(u => `user:default/${u}`),
      'group:default/gitea_admin',
    ]);
    return {
      isAdmin: ownershipEntityRefs.some(r => adminRefs.has(r)),
      isDeveloper: ownershipEntityRefs.includes('group:default/openchoreo'),
    };
  }

  async handle(
    request: PolicyQuery,
    user?: PolicyQueryUser,
  ): Promise<PolicyDecision> {
    const { isAdmin, isDeveloper } = this.rolesFor(
      user?.identity.ownershipEntityRefs ?? [],
    );
    if (isAdmin) {
      return { result: AuthorizeResult.ALLOW };
    }
    if (request.permission.name === catalogEntityReadPermission.name) {
      return { result: AuthorizeResult.ALLOW }; // viewer and up
    }
    if (
      request.permission.name === catalogEntityCreatePermission.name ||
      request.permission.name === kubernetesProxyPermission.name ||
      request.permission.name === kubernetesResourcesReadPermission.name ||
      request.permission.name === kubernetesClustersReadPermission.name
    ) {
      return isDeveloper
        ? { result: AuthorizeResult.ALLOW }
        : { result: AuthorizeResult.DENY };
    }
    if (request.permission.name === catalogEntityRefreshPermission.name) {
      // Developers refresh entities they own; admins are handled above.
      return isDeveloper
        ? createConditionalDecision(request.permission, catalogConditions.isEntityOwner)
        : { result: AuthorizeResult.DENY };
    }
    // catalogEntityDeletePermission and everything unlisted: admin-only.
    return { result: AuthorizeResult.DENY };
  }
}
```

(`PolicyQueryUser` is imported from `@backstage/plugin-permission-node`; the kubernetes permissions are required so the FR-05 data path works for developers in production -- the proxy authorizes `kubernetes.proxy` per request, section 5.3.)

**New file: `backstage/packages/backend/src/modules/permissionsPolicy.ts`**

```ts
import {
  coreServices,
  createBackendModule,
} from '@backstage/backend-plugin-api';
import { policyExtensionPoint } from '@backstage/plugin-permission-node/alpha';
import { SecurityRbacPolicy } from '../extensions/permissionsPolicyExtension';

const securityRbacPolicyModule = createBackendModule({
  pluginId: 'permission',
  moduleId: 'security-rbac-policy',
  register(reg) {
    reg.registerInit({
      deps: { policy: policyExtensionPoint, config: coreServices.rootConfig },
      async init({ policy, config }) {
        policy.setPolicy(new SecurityRbacPolicy(config));
      },
    });
  },
});

export default securityRbacPolicyModule;
```

(Import paths verified in the installed tree: `coreServices.rootConfig` in `@backstage/backend-plugin-api/dist/index.d.ts`; `PolicyQueryUser` in `@backstage/plugin-permission-node/dist/index.d.ts`; `policyExtensionPoint` is alpha-only -- `@backstage/plugin-permission-node/dist/alpha.d.ts:17-19`, served via the package's `./alpha` export subpath. Default export mirrors the existing `giteaAuth.ts:218` module pattern consumed by `index.ts:31`.)

**Edit: `packages/backend/src/index.ts`** -- delete the allow-all import at `:48-50`; add after the permission backend (`:46`):

```ts
backend.add(import('./modules/permissionsPolicy'));
```

**Edit: `packages/backend/package.json`** -- remove `"@backstage/plugin-permission-backend-module-allow-all-policy"` (`:33`); add direct dependencies `"@backstage/plugin-catalog-common"` and `"@backstage/plugin-kubernetes-common"` (both already in the resolved tree transitively -- no new third-party code enters the repo; section 13 records them anyway). `yarn install` refreshes the lockfile.

**Edit: `backstage/app-config.yaml`** -- extend the permission comment block at `:126-128`:

```yaml
permission:
  enabled: true
  policy:
    adminUsers: ${BACKSTAGE_ADMIN_USERS:-gitea_admin}
```

(The base config gains `enabled: true` explicitly; local dev keeps overriding to `false` via `app-config.local.yaml` -- see risk note.)

### 8.5 Behavior-change risks (recorded)

- Local dev guest access MUST keep working: `app-config.local.yaml.example:35-36` keeps `permission.enabled: false`; with the framework disabled the policy is never consulted and the guest provider (`:17`) is unaffected. `scripts/start-backstage.sh:83-88` seeds the local file from the example, preserving this.
- Production keeps `permission.enabled: true` (`app-config.production.yaml:39-40`); the first production start with the new policy must be smoke-verified (section 14) because DENY-by-default now applies to everything unlisted -- including scaffolder and techdocs write paths, which become admin-only until a later wave revisits role granularity. This is a deliberate, recorded posture: smallest enforceable rule set first.
- `catalogConditions.isEntityOwner` matches against entity `spec.owner`; dangling owners (five-plane CTL-G5) mean some refresh requests fall to admins until that gap is fixed. Recorded, not worked around.

### 8.6 Resource footprint

None (backend code/config only).

### 8.7 Verification and smoke hook

- `yarn tsc` passes; backend boots in dev (guest flow intact, `scripts/smoke-auth.sh` green).
- Production-mode checks (part of the section 14 checklist): backend starts with the production config; a session whose refs are only `user:default/<login>` + `group:default/openchoreo` (developer) can read catalog entities and use `/api/kubernetes/proxy`, cannot delete entities; an identity outside `adminUsers` and outside `openchoreo` (viewer) can read but not create; an `adminUsers` identity is unrestricted.
- Smoke hook: `smoke-security.sh` asserts the allow-all import is gone from `index.ts`, the two new files exist, `adminUsers` is configured, and `app-config.local.yaml.example` still carries `enabled: false`.

**Traces to:** FR-08; SEC-G12; D1; NFR-03. Resolves OQ-25.

---

## 9. FR-09 -- TLS on the .local gateways

### 9.1 Goal

`gitea.local`, `signoz.local`, and `opencost.local` serve HTTPS via cert-manager Certificate resources and HTTPS listeners on the existing Gateway, without breaking existing HTTP and port-forward workflows.

### 9.2 Current state (verified)

- The Gateway `eg` (`envoy-gateway` ns) has a single HTTP listener on port 80 (`iac/modules/networking/gateway/main.tf:44-50`); HTTPRoutes are hostname-scoped with `parentRefs` name+namespace only -- no `sectionName` (`iac/modules/networking/gateway/httproutes.tf:13-17`), so a route attaches to ANY listener matching its hostname, which is exactly what lets the same routes serve both 80 and 443 without edits.
- cert-manager v1.19.4 runs in the sibling cluster (deployed, sibling-managed; the 1.19 line is EOL 2026-07-08 -- sibling-owned upgrade flagged in REQ-001 section 5.1). Nothing in this repo references it today.
- The gateway-shim annotation pattern is rejected (REQ-001 FR-09): it requires `enableGatewayAPI` on the sibling-owned controller; Certificate resources + listener `certificateRefs` need no controller change.

### 9.3 Exact change points

**New file: `iac/modules/networking/gateway/tls.tf`** -- five `kubectl_manifest` resources (same resource type the module already uses, `main.tf:1-52`):

```hcl
# SelfSigned bootstrap issuer -> local CA keypair -> local-ca ClusterIssuer.
resource "kubectl_manifest" "selfsigned_bootstrap_issuer" {
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
      name: selfsigned-bootstrap
    spec:
      selfSigned: {}
  EOF
}

resource "kubectl_manifest" "local_ca_certificate" {
  depends_on = [kubectl_manifest.selfsigned_bootstrap_issuer]
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: local-ca-cert
      namespace: cert-manager
    spec:
      isCA: true
      commonName: sovereign-local-ca
      secretName: local-ca-key-pair
      privateKey:
        algorithm: ECDSA
        size: 256
      issuerRef:
        name: selfsigned-bootstrap
        kind: ClusterIssuer
  EOF
}

resource "kubectl_manifest" "local_ca_issuer" {
  depends_on = [kubectl_manifest.local_ca_certificate]
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
      name: local-ca
    spec:
      ca:
        secretName: local-ca-key-pair
  EOF
}
```

plus one Certificate per route (loop over `var.routes`, `variables.tf:37-72`):

```hcl
resource "kubectl_manifest" "route_certificates" {
  depends_on = [kubectl_manifest.local_ca_issuer]
  for_each   = var.routes
  yaml_body = <<-EOF
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: ${each.key}-tls
      namespace: envoy-gateway
    spec:
      secretName: ${each.key}-tls
      dnsNames:
        - ${each.value.hostname}
      issuerRef:
        name: local-ca
        kind: ClusterIssuer
  EOF
}
```

(Secrets land in `envoy-gateway`, the Gateway's namespace, so listener `certificateRefs` need no cross-namespace ReferenceGrant.)

**Edit: `iac/modules/networking/gateway/main.tf`** -- extend the listener list at `:44-50` with one HTTPS listener per route hostname (the existing HTTP listener stays; see the redirect decision below):

```yaml
        - name: https-gitea
          hostname: gitea.local
          protocol: HTTPS
          port: 443
          tls:
            mode: Terminate
            certificateRefs:
              - kind: Secret
                name: gitea-tls
          allowedRoutes:
            namespaces:
              from: All
```

(and the `https-signoz` / `https-opencost` twins). Because `httproutes.tf` uses no `sectionName` (`:13-15`), the existing HTTPRoutes bind to these listeners by hostname automatically.

**HTTP->HTTPS redirect decision (decided):** port 80 keeps serving plain HTTP in Wave 0; no redirect is configured. Reasoning: the smoke suite curls HTTP through a port-forward (`scripts/smoke-m4-networking.sh:37-45` on `38080`), Backstage and scripts use HTTP port-forwards (`scripts/start-backstage.sh:50-77`), and a redirect would silently break every one of those workflows for zero gain on loopback. HTTPS is offered alongside; a redirect becomes its own documented change after clients migrate.

**macOS trust (documented user step):**

```sh
kubectl --context k3d-openchoreo -n cert-manager get secret local-ca-key-pair \
  -o jsonpath='{.data.tls\.crt}' | base64 -d > /tmp/sovereign-local-ca.crt
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain /tmp/sovereign-local-ca.crt
```

Until this step is done, clients use `curl -k` (the smoke does exactly that, below). The step is a user action on the host keychain and is never scripted by install scripts (trust is a human decision).

Apply path: module change flows through `scripts/install-m4-networking.sh` (direct `tofu apply` is guard-blocked per repo convention).

### 9.4 Resource footprint

None new (cert-manager is already deployed; Secrets and Gateway listeners are config objects).

### 9.5 Verification and smoke hook

- `kubectl --context k3d-openchoreo wait --for=condition=Ready certificate -n envoy-gateway --all --timeout=120s`.
- Extend `scripts/smoke-m4-networking.sh`: after the existing HTTP loop (`:37-45`), add a second port-forward (`38443:443`, same `svc` discovery as `:15-17`) and an HTTPS loop: `curl -sk -o /dev/null -w "%{http_code}" -H "Host: ${host}" https://localhost:38443/` expecting 200/302 for each host. The existing HTTP assertions stay unchanged (redirect decision, 9.3).
- Smoke hook: `smoke-security.sh` asserts the three HTTPS listeners in `gateway/main.tf` and the Certificate resources in `tls.tf`.

**Traces to:** FR-09; SEC-G9; D1. Resolves OQ-24.

---

## 10. FR-10 -- dependabot and code scanning as config-as-code

### 10.1 Goal

Scanner posture for the GitHub mirror is version-controlled in this repo.

### 10.2 Current state (verified)

- `.github/` contains exactly one file: `workflows/sync-from-gitea.yml`. No `dependabot.yml`, no code scanning workflow (five-plane SEC-G11).
- Go module roots (verified enumeration): `plugins/rr-policy-guards/tools/{bash,brew,commit,emoji,tofu,verify}-guard`, `seed-repos/hello-m2`, `tools/score2openchoreo` (8 roots; `tools/namespace-predictor` has no `go.mod`).
- Mirror relationship (honest note): the Gitea remotes are the source of truth and GitHub is the mirror (`sync-from-gitea.yml`); dependabot and code scanning take effect GitHub-side only, and dependabot PRs raised on the mirror must flow back through the sync discipline -- an accepted asymmetry recorded here; closing that loop is a control-plane forge-strategy item, not this FR. Gitea Actions never executes `.github/workflows` (it runs `.gitea/workflows`), so nothing double-runs on the forge.

### 10.3 Exact change points

**New file: `.github/dependabot.yml`**

```yaml
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
    schedule: { interval: weekly }
  - package-ecosystem: npm
    directory: /backstage
    schedule: { interval: weekly }
    open-pull-requests-limit: 5
  - package-ecosystem: gomod
    directory: /tools/score2openchoreo
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /seed-repos/hello-m2
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /plugins/rr-policy-guards/tools/bash-guard
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /plugins/rr-policy-guards/tools/brew-guard
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /plugins/rr-policy-guards/tools/commit-guard
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /plugins/rr-policy-guards/tools/emoji-guard
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /plugins/rr-policy-guards/tools/tofu-guard
    schedule: { interval: weekly }
  - package-ecosystem: gomod
    directory: /plugins/rr-policy-guards/tools/verify-guard
    schedule: { interval: weekly }
```

Cross-lane note: when Lane E (section 11) lands `plugins/rr-policy-guards/tools/audit-chain/go.mod`, add a matching gomod entry in the same commit (dependabot validates directories at run time; the entry and the directory must appear together).

**New file: `.github/workflows/code-scanning.yml`**

```yaml
name: Code scanning
on:
  push:
    branches: [main]
  schedule:
    - cron: "0 6 * * 1"
permissions:
  security-events: write
  contents: read
jobs:
  codeql:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        language: [go, javascript-typescript]
    steps:
      - uses: actions/checkout@11d5960a326750d5838078e36cf38b85af677262 # v4.4.0
      - uses: github/codeql-action/init@ff2f1c621b7f889edc0d3c761ac2e6a3f8cdb0dd # v4.37.7
        with:
          languages: ${{ matrix.language }}
      - uses: github/codeql-action/analyze@ff2f1c621b7f889edc0d3c761ac2e6a3f8cdb0dd # v4.37.7
```

Pin resolution (recorded, per D4): `git ls-remote --tags https://github.com/github/codeql-action 'refs/tags/v4*'` on 2026-08-18 gave `v4.37.7`; the peeled commit `refs/tags/v4.37.7^{}` = `ff2f1c621b7f889edc0d3c761ac2e6a3f8cdb0dd` is what the workflow pins (annotated-tag object SHAs must not be pinned -- only the peeled commit). The implementer re-runs the ls-remote and records the result in the commit message.

DECIDED -- CodeQL over Semgrep OSS for the mirror workflow: CodeQL is native to GitHub code scanning (results land in the mirror's Security tab with zero extra accounts or infra), Semgrep already runs locally in verify-guard (`plugins/rr-policy-guards/tools/verify-guard/exec.go:96-169`) so CodeQL diversifies rather than duplicates, and this workflow is mirror-side assurance -- it does not replace or weaken any local gate.

### 10.4 Resource footprint

Zero (GitHub-hosted execution).

### 10.5 Verification and smoke hook

- Files present and YAML-valid; the code-scanning workflow's `uses:` lines are all full-length SHAs (D4); after the next mirror sync, the workflow appears under the mirror's Actions tab (manual confirmation, recorded in the implementation journal).
- Smoke hook: `smoke-security.sh` asserts both files exist and that `dependabot.yml` covers all current `go.mod` roots (recomputed with the same `find` used in 10.2).

**Traces to:** FR-10; SEC-G11; D1, D4.

---

## 11. FR-11 -- Guard audit-log hash chaining

### 11.1 Goal

Every guard's JSONL audit trail becomes tamper-evident via a SHA-256 `prev_hash` chain, per RECORD-IMMUTABILITY-TECH-001 section 9, with a shared verifier.

### 11.2 Current state (verified, all six writers)

| Guard | Append function | Change point |
|---|---|---|
| emoji-guard | `logAudit` | `tools/emoji-guard/audit.go:35` (marshal `:61`, write `:61-65`) |
| bash-guard | `logAudit` | `tools/bash-guard/audit.go:34` (marshal `:60-65`) |
| brew-guard | `logAudit` | `tools/brew-guard/audit.go:33` (marshal `:59`) |
| tofu-guard | `logAudit` | `tools/tofu-guard/audit.go:32` (marshal `:54`) |
| commit-guard | `appendAuditStrict` (fail-open wrapper `appendAudit` `:28-30`) | `tools/commit-guard/audit.go:32-53` (marshal `:45`) |
| verify-guard | `writeAudit` | `tools/verify-guard/audit.go:61-92` (rotation `:80`, `:103-126`; write `:88`) |

All six marshal one JSON object and append it plus `'\n'` to a mode-0600 file -- the schema-agnostic chain works because the hash covers raw line bytes, not the record (TECH-001 9.1). verify-guard additionally rotates (`audit.go:103-126`), which the tail-read helper must account for.

### 11.3 Chain format (from TECH-001 9.1, restated as binding here)

One new field per line: `"prev_hash":"<64 lowercase hex>"` = SHA-256 of the raw bytes of the previous line INCLUDING its trailing newline. Genesis line: 64 zeros. Each guard adds the field to its own audit struct (schema-local field name `prev_hash`, JSON key identical across guards).

### 11.4 Write-side change points (per guard, ~15 lines each)

Identical shape in all six append paths: before marshaling, read the last raw line of the current log (see helper), set `prev_hash`, marshal, append. Fail-open is preserved absolutely (TECH-001 9.2): any tail-read error yields a line with 64 zeros and NEVER blocks or alters the enforcement decision -- a broken link is a verification finding, not a policy failure.

**Inter-process serialization (added 2026-08-21).** The tail-read -> append sequence is a read-modify-write cycle and raced when several agent harnesses (Claude Code, Codex, Kimi) ran the same guard concurrently: two writers chained from the same tail, or both wrote a genesis line into a fresh log, and the verifier correctly reported BROKEN. All six append paths now run the entire critical section (tail read, verify-guard rotation, append) under an exclusive advisory `flock(2)` on a sidecar lock file `<log>.lock` (created once, mode 0600, never deleted; a sidecar rather than the log itself because verify-guard rotation renames the log and the lock must pin a stable inode). Each guard carries an identical `auditlock.go` (`withAuditLock(path, fn)`), consistent with the no-shared-module build contract. The log format is byte-unchanged. Failure policy matches the audit path's existing contract: if the lock cannot be taken, the append is skipped (error returned for strict callers) rather than performed unlocked -- a lost line is recoverable, a broken chain is not; policy enforcement still never blocks on audit I/O.

Shared helper shape (each guard gets its own copy, stdlib only, consistent with the guards' no-shared-module build contract):

```
lastLineBytes(path) -> ([]byte, error):
  if active file exists and is non-empty: return its last line including '\n'
  else if path+".1" exists (verify-guard rotation just ran): return its last line including '\n'
  else: return nil (genesis)
```

### 11.5 Verifier (decided: standalone shared tool)

DECIDED: a new stdlib-only tool at `plugins/rr-policy-guards/tools/audit-chain/` (own `go.mod`, same build contract as the other guards; binary `plugins/rr-policy-guards/bin/rr-audit-chain`), implementing `rr-audit-chain verify <log-path>` and `rr-audit-chain head <log-path>` exactly as TECH-001 9.3 specifies: `verify` re-walks the file (and, for verify-guard, the rotated segments in order `.3` -> `.2` -> `.1` -> active), recomputes each `prev_hash`, reports the first mismatching line number, and prints the chain head; `head` prints only the head hash.

Reasoning over a per-guard `verify` subcommand: the chain is deliberately schema-agnostic over raw bytes (11.3), so ONE verifier serves all six writers; six subcommands would duplicate identical logic into six binaries and couple verification to whichever guard binary happened to be present. The standalone tool also matches the repo's one-static-binary-per-tool layout (AGENTS.md guard build commands).

### 11.6 Tests (per TECH-001 10.4)

- Per guard (`audit_test.go` beside each `audit.go`): `TestAudit_PrevHashChain` -- two appends; line 2's `prev_hash` equals SHA-256 of line 1's raw bytes + `\n`; genesis line carries 64 zeros; tail-read failure path writes zeros and still appends (fail-open pinned).
- Per guard (`auditlock_test.go`, added 2026-08-21): `TestAudit_ConcurrentWriters` (16 goroutines x 25 appends), `TestAudit_ConcurrentProcesses` (8 re-executed test-binary processes x 10 appends), and `TestWithAuditLock_Serializes` (lost-update counter under the lock) all produce logs whose chains re-verify; verify-guard adds `TestAudit_ConcurrentWritersRotation` (rotation forced by a tiny size budget, chain walked across segments).
- Verifier (`main_test.go` in audit-chain): passes a well-formed multi-line log; a hand-edited middle line is detected AT the edited line number; a log whose final line is removed verifies clean internally (11.7's limit, pinned by test so it is never mistaken for a bug); rotated-segment walk for a synthetic verify-guard log set.
- Non-regression: each guard's existing test suite passes unchanged (`go test ./...` per guard).

### 11.7 Honest limits (carried from TECH-001 9.4)

Chaining proves internal consistency only: deletion or truncation of the log's tail is invisible unless the head hash is anchored elsewhere. The optional anchor -- an `audit-head: <guard>=<hash>` line in the checkpoint tag message (TECH-001 9.4, section 4 of that document) -- is REFERENCED here and deliberately NOT built in this wave; it lands with the checkpoint cadence decision in the record-immutability workstream.

### 11.8 Resource footprint

None (local tooling only).

### 11.9 Verification and smoke hook

- `go test ./...` in all six guards and in `audit-chain`; rebuild all six guard binaries plus `bin/rr-audit-chain` (per-guard build commands per AGENTS.md).
- Live check: run any guard decision, then `rr-audit-chain verify ~/.rational-reserve/logs/<guard>.jsonl` passes; hand-edit a line in a COPY and confirm detection at the right line number.
- Smoke hook: `smoke-security.sh` runs `rr-audit-chain verify` against the live guard logs present on the host (skips guards with no log yet, reporting SKIP not FAIL -- a log that does not exist yet is not a chain failure; recorded semantics).

**Traces to:** FR-11; SEC-G4; CTL-G6 touchpoint; D1; NFR-08; TECH-001 section 9 (binding format), REQ-001 FR-11 cross-document note (the "at all" half of record-immutability OQ-04 is resolved here under D5; per-guard scope remains OQ-04's).

---

## 12. Implementation lanes and order

Five parallel-safe lanes with disjoint file sets. Cross-lane conflicts are named explicitly with their sequencing rule.

### Lane A -- CI scanning (FR-01, FR-02, FR-03)

- Steps in order: (1) create `scripts/smoke-security.sh` with the harness skeleton (`set -euo pipefail`, namespaced `info()`, PASS/FAIL accounting, SKIP semantics per section 14.2); (2) `.trivyignore` + `osv-scanner.toml` seeds; (3) scan steps into `seed-repos/hello-m2/.gitea/workflows/ci.yaml`; (4) `scripts/ci/commit-security-artifacts.sh` + artifact step; (5) same into `iac/templates/ci.yaml`; (6) push, watch a run end to end; (7) negative-test on a scratch branch, revert; (8) append the Lane A smoke checks (section 14.2) to `scripts/smoke-security.sh`.
- Files: `seed-repos/hello-m2/.gitea/workflows/ci.yaml`, `iac/templates/ci.yaml`, `scripts/ci/commit-security-artifacts.sh` (new), `seed-repos/hello-m2/.trivyignore` (new), `seed-repos/hello-m2/osv-scanner.toml` (new), `scripts/smoke-security.sh` (new; Lane A creates it).
- Exit: seeded HIGH/CRITICAL finding fails the gate; clean push passes and commits `security-artifacts/hello-m2/development/{latest,<run-id>}.json`.

### Lane B -- Gatekeeper visibility (FR-05, FR-06, FR-07)

- Steps in order: (1) `app-config.yaml` kubernetes block + `start-backstage.sh` proxy manager; (2) `gatekeeper.ts` helper + PolicyCard rewrite; (3) Prometheus `extraScrapeConfigs` via the M4 lifecycle; (4) collector filelog receiver via the M3 lifecycle; (5) seed a C2 violation, verify all three surfaces, clean up; (6) append the Lane B smoke checks (section 14.2) to `scripts/smoke-security.sh`.
- Files: `backstage/app-config.yaml`, `scripts/start-backstage.sh`, `scripts/stop-backstage.sh`, `backstage/packages/app/src/modules/openchoreo-cards/gatekeeper.ts` (new), `.../PolicyCard.tsx`, `observability/cost/prometheus-values.local.yaml`, `observability/otel/collector-values.local.yaml`, `scripts/smoke-security.sh` (shared, append-only).
- Exit: PolicyCard shows live `totalViolations`; Prometheus targets show the gatekeeper job up; seeded violation produces a queryable `violation_audited` line in SigNoz.

### Lane C -- Portal surfaces (FR-04, FR-08)

- Steps in order: (1) `SecurityCard.tsx` + entity-page registration; (2) `permissionsPolicyExtension.ts` + `permissionsPolicy.ts`; (3) `index.ts` + `package.json` edits, `yarn install`, `yarn tsc`; (4) `app-config.yaml` permission block; (5) dev boot (guest intact) + production-mode role checks; (6) append the Lane C smoke checks (section 14.2) to `scripts/smoke-security.sh`.
- Files: `backstage/packages/app/src/modules/openchoreo-cards/SecurityCard.tsx` (new), `.../openchoreo-entity-page/index.tsx`, `backstage/packages/backend/src/extensions/permissionsPolicyExtension.ts` (new), `backstage/packages/backend/src/modules/permissionsPolicy.ts` (new), `backstage/packages/backend/src/index.ts`, `backstage/packages/backend/package.json`, `backstage/app-config.yaml`, `scripts/smoke-security.sh` (shared, append-only).
- Exit: Security tab renders live artifact + violations (or explicit not-wired); roles behave per the section 8.7 checks.

### Lane D -- Infra/config (FR-09, FR-10)

- Steps in order: (1) `tls.tf` + Gateway listeners via the M4-networking lifecycle; (2) smoke-m4-networking HTTPS extension; (3) documented macOS trust step (user action); (4) `dependabot.yml` + `code-scanning.yml`; (5) append the Lane D smoke checks (section 14.2) to `scripts/smoke-security.sh`.
- Files: `iac/modules/networking/gateway/tls.tf` (new), `iac/modules/networking/gateway/main.tf`, `scripts/smoke-m4-networking.sh`, `.github/dependabot.yml` (new), `.github/workflows/code-scanning.yml` (new), `scripts/smoke-security.sh` (shared, append-only).
- Exit: Certificates Ready; HTTPS 200/302 for all three hosts with HTTP untouched; both config files present and pin-clean.

### Lane E -- Guards (FR-11)

- Steps in order: (1) `audit-chain` tool + tests; (2) per-guard write-side change + `TestAudit_PrevHashChain` (six guards); (3) rebuild all binaries; (4) live verify + tamper-detection check; (5) dependabot entry for the new module (edit owned by Lane D's file, sequenced here); (6) append the Lane E smoke checks (section 14.2) to `scripts/smoke-security.sh`.
- Files: `plugins/rr-policy-guards/tools/audit-chain/*` (new), the six `tools/*/audit.go` + `audit_test.go`, `plugins/rr-policy-guards/bin/` (gitignored build outputs), `.github/dependabot.yml` (single-entry addition), `scripts/smoke-security.sh` (shared, append-only).
- Exit: verifier passes live logs; a hand-edited line is detected at the right line number; all guard test suites green.

### Explicit cross-lane conflicts

- `backstage/app-config.yaml` is edited by Lane B (FR-05, `kubernetes:` block) and Lane C (FR-08, `permission:` block). Disjoint keys, same file: Lane B lands first; Lane C rebases. (FR-04 consumes the Gitea proxy already in the file, so Lane C has no app-config need beyond FR-08.)
- `.github/dependabot.yml` is created by Lane D and appended by Lane E (audit-chain entry, only after the module exists).
- `gatekeeper.ts` is created in Lane B and consumed by Lane C's SecurityCard -- Lane B lands first, or Lane C implements against this specification and merges after B.
- FR-04's scan-artifact section soft-depends on Lane A's artifact path; the card renders the honest not-wired state until Lane A's first artifact lands, so this is an ordering preference, not a merge dependency.
- `scripts/smoke-security.sh` is shared, append-only: Lane A creates the harness skeleton, each lane appends its own check block as it lands, and no lane edits another lane's block. The `scripts/smoke-all.sh` `SUITES`/banner edit (`:7`, `:34`) is deliberately NOT in any lane -- it is the acceptance-time orchestrator step (section 14.3), owned by whoever executes the acceptance checklist, so the umbrella never advertises a half-populated suite mid-wave.
- All other file sets are disjoint; Lanes A, B, D, E are otherwise fully parallel.

---

## 13. Provenance duty

New third-party components introduced by Wave 0, with licenses per the REQ-001 section 7 register:

| Component | Form | License | Pin |
|---|---|---|---|
| Trivy | container image `aquasec/trivy:0.74.0` | Apache-2.0 | index digest `sha256:62b1e65e...` (section 1.3) |
| OSV-Scanner | container image `ghcr.io/google/osv-scanner:v2.5.1` | Apache-2.0 | index digest `sha256:8108ae94...` (section 2.3) |
| CodeQL action | `github/codeql-action` v4.37.7 | MIT (implementer re-verifies against the action repo LICENSE at implementation time; license value per upstream repo, not independently re-verified here) | peeled commit `ff2f1c62...` (section 10.3) |
| `@backstage/plugin-catalog-common` | npm, newly-DIRECT dependency (already in the resolved tree) | Apache-2.0 | existing yarn.lock version |
| `@backstage/plugin-kubernetes-common` | npm, newly-DIRECT dependency (already in the resolved tree) | Apache-2.0 | existing yarn.lock version |

Explicitly NOT new (existing deployments/resources, no attribution change): cert-manager (sibling-managed), Gatekeeper, Prometheus, OTEL collector, SigNoz, `kubectl` (host tooling), dependabot (GitHub-native feature, no vendored code).

Duty at implementation time (NFR-05, non-optional): regenerate `THIRD-PARTY-LICENSES.md` and `provenance/PROVENANCE.md`, and re-issue `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md`; the superseded certificate stays in git history. Attribution is recorded, never claimed.

---

## 14. Test and acceptance plan

### 14.1 Per-FR verification (summary; full commands in each section)

| FR | Primary verification | Negative test |
|---|---|---|
| FR-01 | Clean push passes; reports written (section 1.7) | Scratch-branch HIGH/CRITICAL dependency fails the gate |
| FR-02 | `osv.json` produced; suppression honored (2.7) | Same scratch dependency fails the OSV gate |
| FR-03 | Contents API returns `latest.json` + `<run-id>.json` (3.5) | Malformed artifact JSON fails `jq` assembly (step fails loudly) |
| FR-04 | Tab renders live data; not-wired state on artifact-less entity (4.5) | Proxy down -> explicit not-wired, no placeholder-as-live |
| FR-05 | Seeded C2 violation visible on PolicyCard (5.6) | Unauthenticated proxy request denied in production mode |
| FR-06 | Targets up; `gatekeeper_violations` queryable (6.5) | Seed removed -> series returns to zero |
| FR-07 | `violation_audited` line queryable in SigNoz (7.5) | Non-matching query returns nothing (no false positives) |
| FR-08 | Role matrix checks (8.7) | Viewer create -> DENY; developer delete -> DENY |
| FR-09 | Certificates Ready; HTTPS 200/302 x3; HTTP unchanged (9.5) | Untrusted client without `-k` fails TLS validation (proves real TLS) |
| FR-10 | Files present; pins are full SHAs (10.5) | dependabot config missing a go.mod root -> smoke fails |
| FR-11 | Verifier passes live logs (11.9) | Hand-edited middle line detected at exact line number |

### 14.2 Smoke-suite touchpoints

- NEW `scripts/smoke-security.sh`, following the established pattern (`set -euo pipefail`, namespaced `info()` logger, PASS/FAIL lines, non-zero exit on failure -- the shape of `scripts/smoke-m4-networking.sh:1-51`). Checks, grouped by lane: workflow scan steps + digest pins (A); app-config kubernetes block, proxy manager, PolicyCard wiring, Prometheus values, collector values (B); Security tab registration, RBAC files, allow-all removal, `enabled: false` preserved in the example (C); HTTPS listeners, Certificate resources, dependabot/code-scanning files + pin check (D); `rr-audit-chain verify` on live guard logs (E). Cluster-mode checks (kubectl/API reads) degrade to SKIP with a printed reason when the cluster is unreachable, mirroring `smoke-m3.sh`'s offline/cluster modes. Ownership: Lane A creates the file (harness skeleton); each lane appends its own check block as it lands (append-only; no lane edits another lane's block).
- `scripts/smoke-all.sh`: add `"security"` to `SUITES` at `:7` (and to the final banner strings at `:34`), so the umbrella covers AUTH + M2 + M3 + M4 + BACKSTAGE-PRODUCTION + SECURITY. Ownership: an acceptance-time orchestrator step, deliberately not in any lane -- it is executed by whoever runs the section 14.3 checklist, after all lanes have landed, so the umbrella never advertises a half-populated suite mid-wave.
- Non-regression (NFR-06): `smoke-all.sh` must be green after each lane lands; the M2 delivery contract is untouched by every section of this document.

### 14.3 Wave-0 acceptance checklist (maps to REQ-001 section 13, Wave-0 exit)

- [ ] A CI run fails on a seeded HIGH/CRITICAL finding (FR-01/FR-02 negative tests executed and reverted).
- [ ] Scan artifacts are committed per push and readable via the Contents API (FR-03).
- [ ] The Security tab renders them live, with explicit not-wired states elsewhere (FR-04).
- [ ] PolicyCard shows live constraint violations from a seeded C2 violation (FR-05).
- [ ] `gatekeeper_violations` is queryable in Prometheus (FR-06).
- [ ] Audit events are queryable in SigNoz via `event_type` (FR-07).
- [ ] The three roles are enforced by the permission backend; guest dev flow intact (FR-08).
- [ ] The three `.local` routes serve HTTPS (FR-09).
- [ ] dependabot and code scanning are in-repo and pin-clean (FR-10).
- [ ] The guard hash chain verifies and detects a planted edit (FR-11).
- [ ] `smoke-all.sh` (extended with the security suite) is green (NFR-06).
- [ ] Attribution triple regenerated and certificate re-issued (section 13; NFR-05).

---

**End of Technical Specification**

This document was created per the POLICIES.md triad rule as the Wave-0 implementation specification of SEC-PLANE-PULLFORWARD-REQ-001. Every code-level claim was verified against the live source or cluster on 2026-08-18; sections 1.5, 2.3, and 10.3 record the re-verification procedures the implementer must re-run and log at implementation time.
