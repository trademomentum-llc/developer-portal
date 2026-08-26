# Provenance Listing -- developer-portal

This is the per-component provenance listing for every third-party work
incorporated into, deployed by, or required by this repository. For each
component it records: name, version/pin as recorded in the repo, upstream
source URL, SPDX license, copyright/attribution holder, how this project
uses it (usage mode), and the repo evidence path(s) proving the version
and usage.

## How this listing is maintained

- Regenerated whenever dependencies change: dependency PRs, milestone
  install scripts, lockfile updates, or chart/provider bumps.
- After regenerating this file and `THIRD-PARTY-LICENSES.md`, the
  certificate `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md` is
  re-issued with fresh SHA-256 digests. Any change to either file
  invalidates the digests recorded in the current certificate.
- License and copyright verification for this generation (2026-08-18) was
  performed against upstream LICENSE files, registry metadata, the
  installed `backstage/node_modules` tree, and the local Go module cache.

## Authoritative dependency closure

This listing is the human-readable record. The machine-readable lockfiles
are the authoritative complete dependency closure:

- `backstage/yarn.lock` -- full Node.js transitive closure (lockfile
  metadata version 8).
- `tools/score2openchoreo/go.sum`, `seed-repos/hello-m2/go.sum` -- Go
  module closures.
- `iac/.terraform.lock.hcl` -- OpenTofu provider closure (hash-pinned).

Only two third-party artifacts are vendored (bytes committed in-repo): the
Score JSON schema (`tools/score2openchoreo/assets/score.schema.json`) and
the Yarn release (`backstage/.yarn/releases/yarn-4.18.0.cjs`). Everything
else is fetched at install time (helm charts, tofu providers, brew/npx
tools) or pulled at container runtime or image build time (container
images; for example the `node:24-trixie-slim` base is pulled by
`yarn build-image` at docker build time, not by the cluster).

## Usage modes

- `vendored` -- upstream file bytes are committed in this repo.
- `scaffold origin` -- file tree originates from an upstream scaffolder.
- `deployed via helm` -- chart installed by a tofu `helm_release` or an
  install script.
- `runtime image` -- container image pulled by the cluster at runtime.
- `build dependency` -- needed to compile, build, or test project code.
- `runtime dependency` -- needed by the running software.
- `fetched at install time` -- host tool installed unpinned by scripts
  (brew, npx, releases/latest) or by CI.
- `referenced` -- deployed by the sibling openchoreo checkout and only
  referenced from this repo.

---

## 1. Platform and Infrastructure (15 entries)

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | OpenTofu | >= 1.9.0, < 1.12.0; CI pins 1.9.0; brew unpinned | https://github.com/opentofu/opentofu | MPL-2.0 | The OpenTofu Authors / Linux Foundation (LICENSE records Copyright (c) 2014 HashiCorp, Inc.) | fetched at install time (host + CI) | iac/versions.tf:2; iac/templates/ci.yaml:31-33; scripts/install-m2.sh:20 |
| 2 | OpenChoreo (Environment CRDs) | unpinned | https://github.com/openchoreo/openchoreo | Apache-2.0 | OpenChoreo authors | referenced (CRDs applied by this repo; platform deployed by sibling) | iac/modules/openchoreo-environments/main.tf:18-24 |
| 3 | OpenBao | unpinned | https://github.com/openbao/openbao | MPL-2.0 | OpenBao contributors / Linux Foundation | referenced (deployed by sibling, dev mode) | iac/modules/external-secrets-wiring/main.tf:25; scripts/smoke-openbao.sh:5-7; scripts/seed-openbao-m2-paths.sh |
| 4 | External Secrets Operator | unpinned | https://github.com/external-secrets/external-secrets | Apache-2.0 | External Secrets Authors / CNCF | referenced (CRDs used; operator deployed by sibling) | iac/modules/external-secrets-wiring/main.tf:18-22; iac/modules/gitea-runner/main.tf:8 |
| 5 | k3d | host binary v5.9.0 (measured 2026-08-18); deployed cluster version managed by the sibling checkout | https://github.com/k3d-io/k3d | MIT | k3d authors (Rancher) | referenced (cluster created by sibling) | scripts/install-m1.sh:79-84 |
| 6 | k3s | deployed server v1.32.9+k3s1 (kubectl version, 2026-08-18); explicitly distinct from the k3d 5.9.0 default v1.35.5-k3s1, which is NOT the deployed version | https://github.com/k3s-io/k3s | Apache-2.0 | Rancher / SUSE / CNCF | referenced (deployed by sibling) | scripts/install-m1.sh:97,111,117 |
| 7 | kubectl | unpinned | https://github.com/kubernetes/kubernetes | Apache-2.0 | The Kubernetes Authors / CNCF | fetched at install time (host CLI; backend transport) | scripts/install-m1.sh:84-123; iac/backend.tf:1-9 |
| 8 | Helm | unpinned | https://github.com/helm/helm | Apache-2.0 | Helm Authors / CNCF | fetched at install time (host CLI; helm provider) | scripts/install-m1.sh:134-157; scripts/install-m3.sh:29-31 |
| 9 | Flux CLI | unpinned | https://github.com/fluxcd/flux2 | Apache-2.0 | The Flux authors / CNCF | fetched at install time (brew) | scripts/install-m2.sh:21 |
| 10 | Infracost CLI | unpinned (releases/latest in CI) | https://github.com/infracost/infracost | Apache-2.0 | Infracost Inc. (2021) | fetched at install time (host + CI) | scripts/install-m2.sh:22; scripts/smoke-infracost.sh:14; iac/templates/ci.yaml:45; seed-repos/hello-m2/.gitea/workflows/ci.yaml:51 |
| 11 | score-k8s CLI | unpinned | https://github.com/score-spec/score-k8s | Apache-2.0 | Score authors / Humanitec | fetched at install time (brew) | scripts/install-m2.sh:23 |
| 12 | Docker / Moby | unpinned | https://github.com/moby/moby | Apache-2.0 | Docker, Inc. | runtime image (dind sidecar) + CI tooling | iac/templates/ci.yaml:57,61; iac/modules/gitea-runner/main.tf:58-61 |
| 13 | jq | unpinned | https://github.com/jqlang/jq | MIT (docs CC-BY-3.0) | Stephen Dolan | fetched at install time (runner-image provided) | iac/templates/ci.yaml:46-47; scripts/ci/*.sh |
| 14 | OPA (Open Policy Agent) | "OPA 1.x" (prose, not pinned) | https://github.com/open-policy-agent/opa | Apache-2.0 | OPA authors / CNCF (originally Styra) | fetched at install time (policy test runtime) | policies/README.md:16-20 |
| 15 | yarn classic (host tool) | unpinned (brew yarn formula, classic 1.x line) | https://github.com/yarnpkg/yarn | BSD-2-Clause | Yarn Contributors | fetched at install time (brew) | scripts/install-m1.sh:66-67 |

The brew-installed yarn classic (#15) is the host bootstrap tool; the Yarn
used inside `backstage/` is the vendored 4.4.1 release (group 5 #3).

## 2. Helm Charts and Container Images (16 entries)

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | Flux CD (chart flux2) | chart 2.13.0 | https://github.com/fluxcd/flux2 + https://github.com/fluxcd-community/helm-charts | Apache-2.0 | The Flux authors / CNCF | deployed via helm (tofu) | iac/modules/flux/main.tf:8-10 |
| 2 | OPA Gatekeeper (chart gatekeeper) | chart 3.17.1; deployed v3.17.1 confirmed live 2026-08-18 (chart gatekeeper-3.17.1; images openpolicyagent/gatekeeper:v3.17.1 on audit + controller-manager) -- matches the repo pin, no drift | https://github.com/open-policy-agent/gatekeeper | Apache-2.0 | Gatekeeper authors / CNCF (originally Microsoft) | deployed via helm (tofu) | iac/modules/gatekeeper/main.tf:8-10 |
| 3 | Gitea Actions runner (act_runner) | chart actions 0.1.0; deployed live 2026-08-18 (namespace gitea-runners): chart actions-0.1.0 (app 0.261.3); runner image docker.gitea.com/act_runner:0.3.1 (StatefulSet act-runner-actions-act-runner); dind sidecar docker.io/docker:29.4.0-dind | https://gitea.com/gitea/act_runner | MIT | The Gitea Authors (2022) | deployed via helm (tofu); residual sub-gap: the catthehacker/ubuntu:act-* job image exists only while a CI job is in flight (zero matches in a full-cluster scan 2026-08-18), its tag observable only during a CI run | iac/modules/gitea-runner/main.tf:26-34 |
| 4 | Gitea (chart gitea-charts/gitea) | repo does not pin the chart (no --version); deployed live 2026-08-18: chart gitea-12.5.0, app 1.25.4, image docker.gitea.com/gitea:1.25.4-rootless | https://gitea.com/gitea/helm-chart + https://github.com/go-gitea/gitea | MIT | The Gitea Authors (chart also credits NOVUM-RGI, Charlie Drage, John Felten) | deployed via helm (install script); the repo values file does not pin the chart, so the deployed version is the live evidence | scripts/install-m1.sh:134,153-157; scripts/gitea-values.yaml |
| 5 | SigNoz | chart 0.130.1 | https://github.com/SigNoz/signoz + https://github.com/SigNoz/charts | MIT (community; ee/ and cmd/enterprise/ under separate enterprise license) | SigNoz Inc (2020-present) | deployed via helm (tofu) | iac/modules/observability/main.tf:17-21; variables.tf:4; observability/signoz/values.local.yaml |
| 6 | ClickHouse (SigNoz subchart) | unpinned (chart-managed) | https://github.com/ClickHouse/ClickHouse | Apache-2.0 | ClickHouse, Inc. | deployed via helm (subchart) | observability/signoz/values.local.yaml:15-26 |
| 7 | OpenTelemetry Collector | chart 0.155.0; image otel/opentelemetry-collector-contrib:0.155.0 | https://github.com/open-telemetry/opentelemetry-helm-charts + https://github.com/open-telemetry/opentelemetry-collector-contrib | Apache-2.0 | OpenTelemetry Authors / CNCF | deployed via helm (tofu) + runtime image | iac/modules/observability/main.tf:30-34; variables.tf:10; observability/otel/collector-values.local.yaml:12-17 |
| 8 | Prometheus | chart 29.13.0 | https://github.com/prometheus/prometheus + https://github.com/prometheus-community/helm-charts | Apache-2.0 | The Prometheus Authors / CNCF | deployed via helm (tofu) | iac/modules/cost/main.tf:13-17; variables.tf:4; observability/cost/prometheus-values.local.yaml |
| 9 | OpenCost | chart 2.5.25 | https://github.com/opencost/opencost + https://github.com/opencost/opencost-helm-chart | Apache-2.0 | The OpenCost Authors / CNCF (originated at Kubecost) | deployed via helm (tofu) | iac/modules/cost/main.tf:26-30; variables.tf:10; observability/cost/opencost-values.local.yaml |
| 10 | Envoy Gateway | OCI chart 1.3.1; deployed images live 2026-08-18: control plane docker.io/envoyproxy/gateway:v1.3.1 (matches the chart pin), data plane docker.io/envoyproxy/envoy:distroless-v1.33.0 | https://github.com/envoyproxy/gateway | Apache-2.0 | Envoy Gateway Authors / CNCF | deployed via helm (tofu, OCI chart) | iac/modules/networking/envoy-gateway/main.tf:5-7; iac/modules/networking/variables.tf:4 |
| 11 | Cilium (module disabled) | chart 1.16.5; enable_cilium=false | https://github.com/cilium/cilium | Apache-2.0 (userspace); eBPF objects under bpf/ dual GPL-2.0-only OR BSD-2-Clause (upstream bpf/COPYING, with LICENSE.GPL-2.0 and LICENSE.BSD-2-Clause) | Copyright Authors of Cilium (source headers) / Isovalent / CNCF | deployed via helm (tofu) -- disabled by default | iac/modules/networking/cilium/main.tf:4-6; iac/modules/networking/variables.tf:16,19-23 |
| 12 | Bitnami PostgreSQL chart | OCI chart 16.4.5 | https://github.com/bitnami/charts | Apache-2.0 | Broadcom Inc (Bitnami) (2025) | deployed via helm (tofu, OCI chart) | iac/modules/postgres/main.tf:23-25; iac/modules/postgres/variables.tf:4 |
| 13 | PostgreSQL (bitnamilegacy image) | image 17.6.0-debian-12-r4 | https://www.postgresql.org/ | PostgreSQL | PostgreSQL Global Development Group | runtime image (deliberate legacy-image opt-in) | iac/modules/postgres/main.tf:64-69 (image.repository/tag); allowInsecureImages=true at main.tf:76 |
| 14 | distribution/distribution (OCI registry) | image registry:2.8 | https://github.com/distribution/distribution | Apache-2.0 | distribution authors / CNCF | runtime image | iac/modules/local-registry/main.tf:26 |
| 15 | memcached (Gitea chart subchart) | unpinned (chart-managed) | https://github.com/memcached/memcached | BSD-3-Clause | memcached authors (Danga Interactive) | deployed via helm (subchart) | scripts/gitea-values.yaml:25-26 |
| 16 | Bitnami PostgreSQL subchart (Gitea chart) | unpinned (chart-managed) | https://github.com/bitnami/charts | Apache-2.0 (chart) / PostgreSQL (db) | Broadcom Inc (Bitnami) / PostgreSQL Global Development Group | deployed via helm (subchart) | scripts/gitea-values.yaml:20-24 |

Chart sub-chart dependencies measured live in the gitea namespace
2026-08-18 (recorded as dependency notes under the Gitea entry #4, not as
separate listed entries): docker.io/bitnamilegacy/postgresql:17.6.0-debian-12-r4
and docker.io/bitnamilegacy/valkey-cluster:8.1.3-debian-12-r3.

## 3. IaC Providers (5 entries)

All fetched by `tofu init` from registry.opentofu.org; hashes pinned in
`iac/.terraform.lock.hcl` (authoritative).

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | hashicorp/kubernetes | constraint ~> 2.33; locked 2.38.0 | https://github.com/hashicorp/terraform-provider-kubernetes | MPL-2.0 | HashiCorp, Inc. | fetched at install time (tofu init) | iac/versions.tf:4; iac/.terraform.lock.hcl:64-65 |
| 2 | hashicorp/helm | constraint ~> 2.17; locked 2.17.0 | https://github.com/hashicorp/terraform-provider-helm | MPL-2.0 | HashiCorp, Inc. | fetched at install time (tofu init) | iac/versions.tf:5; iac/.terraform.lock.hcl:37-38 |
| 3 | hashicorp/null | constraint ~> 3.2; locked 3.3.0 | https://github.com/hashicorp/terraform-provider-null | MPL-2.0 | HashiCorp, Inc. | fetched at install time (tofu init) | iac/versions.tf:7; iac/.terraform.lock.hcl:89-90 |
| 4 | hashicorp/random | constraint ~> 3.6; locked 3.9.0 | https://github.com/hashicorp/terraform-provider-random | MPL-2.0 | HashiCorp, Inc. | fetched at install time (tofu init) | iac/versions.tf:8; iac/.terraform.lock.hcl:126-127 |
| 5 | alekc/kubectl | constraint ~> 2.1; locked 2.2.0 | https://github.com/alekc/terraform-provider-kubectl | MPL-2.0 | alekc (Alexander Chernov); fork of gavinbunney/terraform-provider-kubectl | fetched at install time (tofu init) | iac/versions.tf:6; iac/.terraform.lock.hcl:4-5 |

## 4. Go Tooling and Modules (22 entries)

Tooling (7 entries):

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | Go toolchain | go directives 1.21 / 1.23 / 1.26; CI installs 1.26; toolchain unpinned | https://go.dev | BSD-3-Clause | The Go Authors (2009) / Google | build dependency (language toolchain) | tools/score2openchoreo/go.mod:3; plugins/rr-policy-guards/tools/{bash,brew,emoji,tofu,verify}-guard/go.mod:3; plugins/rr-policy-guards/tools/commit-guard/go.mod:3; seed-repos/hello-m2/go.mod:3; iac/templates/ci.yaml:18 |
| 2 | gopkg.in/yaml.v3 (go-yaml) | v3.0.1 | https://github.com/go-yaml/yaml | MIT AND Apache-2.0 | Kirill Simonov (2006-2011, MIT files); Canonical Ltd (2011-2019, Apache files) | build dependency (score2openchoreo) | tools/score2openchoreo/go.mod:6; go.sum:3-4; main.go:10; schema.go:13 |
| 3 | github.com/santhosh-tekuri/jsonschema/v5 | v5.3.1 | https://github.com/santhosh-tekuri/jsonschema | Apache-2.0 | Santhosh Kumar Tekuri (evidenced via api.github.com/users/santhosh-tekuri; the v5.3.1 artifact contains no copyright statement: bare Apache-2.0 LICENSE, no NOTICE/COPYING/headers) | build dependency (score2openchoreo) | tools/score2openchoreo/go.mod:7; go.sum:1-2; schema.go:12 |
| 4 | OpenChoreo (algorithm replica) | unpinned; mirrored revision not recorded | https://github.com/openchoreo/openchoreo | Apache-2.0 | The OpenChoreo Authors (upstream SPDX headers: Copyright 2025 The OpenChoreo Authors) | build dependency (source-level mirror in namespace-predictor) | tools/namespace-predictor/main.go:18,32,43-45,83-88 |
| 5 | Semgrep | unpinned (PATH) | https://github.com/semgrep/semgrep | LGPL-2.1 | Semgrep Inc. (upstream COPYRIGHT file: Copyright (C) 2019, 2020, 2021, 2022, 2023, 2024 Semgrep Inc.) | fetched at install time (invoked by rr-verify-guard) | plugins/rr-policy-guards/README.md:25; plugins/rr-policy-guards/tools/verify-guard/exec.go:96-97 |
| 6 | Gitleaks | unpinned (PATH) | https://github.com/gitleaks/gitleaks | MIT | Zachary Rice (2019) | fetched at install time (invoked by rr-verify-guard) | plugins/rr-policy-guards/README.md:26; tools/verify-guard/exec.go:110-111 |
| 7 | govulncheck (golang.org/x/vuln/cmd/govulncheck) | v1.7.0 (host-installed 2026-08-18 via go install, ~/go/bin/govulncheck; previously unpinned PATH-resolved) | https://github.com/golang/vuln | BSD-3-Clause | The Go Authors (module cache LICENSE at v1.7.0, first line: Copyright 2009 The Go Authors.) | fetched at install time (go install host tool; invoked by rr-verify-guard per AGENTS.md and used in the 2026-08-18 full-spectrum sweeps) | plugins/rr-policy-guards/README.md:30; tools/verify-guard/exec.go:168-169 |

Seed application modules, seed-repos/hello-m2 (15 entries; go.sum is the
authoritative closure). The golang.org/x module licenses were verified
per-repo from the module cache at the exact pinned versions (2026-08-18).

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 8 | go.opentelemetry.io/otel (+ trace, sdk, metric, otlptrace exporters) | v1.44.0 | https://github.com/open-telemetry/opentelemetry-go | Apache-2.0 | The OpenTelemetry Authors / CNCF (LICENSE embeds Go BSD-3-Clause for derived code) | build dependency (hello-m2) | seed-repos/hello-m2/go.mod:6-9,20-21; main.go:14-20 |
| 9 | go.opentelemetry.io/auto/sdk | v1.2.1 | https://github.com/open-telemetry/opentelemetry-go | Apache-2.0 | The OpenTelemetry Authors / CNCF | build dependency (hello-m2) | seed-repos/hello-m2/go.mod:19 |
| 10 | go.opentelemetry.io/proto/otlp | v1.10.0 | https://github.com/open-telemetry/opentelemetry-proto-go | Apache-2.0 | The OpenTelemetry Authors / CNCF | build dependency (hello-m2) | seed-repos/hello-m2/go.mod:22 |
| 11 | github.com/cenkalti/backoff/v5 (indirect) | v5.0.3 | https://github.com/cenkalti/backoff | MIT | Cenk Alti (2014) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:13 |
| 12 | github.com/cespare/xxhash/v2 (indirect) | v2.3.0 | https://github.com/cespare/xxhash | MIT | Caleb Spare (2016) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:14 |
| 13 | github.com/go-logr/logr (indirect) | v1.4.3 | https://github.com/go-logr/logr | Apache-2.0 | go-logr contributors | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:15 |
| 14 | github.com/go-logr/stdr (indirect) | v1.2.2 | https://github.com/go-logr/stdr | Apache-2.0 | go-logr contributors | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:16 |
| 15 | github.com/google/uuid (indirect) | v1.6.0 | https://github.com/google/uuid | BSD-3-Clause | Google Inc. (2009, 2014) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:17 |
| 16 | github.com/grpc-ecosystem/grpc-gateway/v2 (indirect) | v2.29.0 | https://github.com/grpc-ecosystem/grpc-gateway | BSD-3-Clause | Gengo, Inc. (2015) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:18 |
| 17 | golang.org/x/net (indirect) | v0.56.0 | https://pkg.go.dev/golang.org/x/net | BSD-3-Clause (per-repo LICENSE at the pinned version; PATENTS file present) | The Go Authors (per-repo LICENSE first line: Copyright 2009 The Go Authors.) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:23 |
| 18 | golang.org/x/sys (indirect) | v0.46.0 | https://pkg.go.dev/golang.org/x/sys | BSD-3-Clause (per-repo LICENSE at the pinned version; PATENTS file present) | The Go Authors (per-repo LICENSE first line: Copyright 2009 The Go Authors.) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:24 |
| 19 | golang.org/x/text (indirect) | v0.39.0 | https://pkg.go.dev/golang.org/x/text | BSD-3-Clause (per-repo LICENSE at the pinned version; PATENTS file present) | The Go Authors (per-repo LICENSE first line: Copyright 2009 The Go Authors.) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:25 |
| 20 | google.golang.org/genproto (googleapis/api + googleapis/rpc, indirect) | v0.0.0-20260526163538-3dc84a4a5aaa | https://github.com/googleapis/go-genproto | Apache-2.0 | Google LLC | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:26-27 |
| 21 | google.golang.org/grpc (indirect) | v1.83.0 | https://github.com/grpc/grpc-go | Apache-2.0 | gRPC authors / Google LLC (CNCF) | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:28 |
| 22 | google.golang.org/protobuf (indirect) | v1.36.11 | https://github.com/protocolbuffers/protobuf-go | BSD-3-Clause | The Go Authors (2018) / Google LLC | build dependency (hello-m2, indirect) | seed-repos/hello-m2/go.mod:29 |

## 5. Backstage and Node.js (128 entries)

`backstage/` was scaffolded with `@backstage/create-app` and tracks the
Backstage release line 1.49.1. Scaffold-originating files are Apache-2.0;
custom code on top (`packages/backend/src/modules/giteaAuth.ts`,
`packages/app/src/modules/*`, app-config files) is original first-party
work. All 113 installed `@backstage/*` packages were verified (2026-08-18)
to ship no LICENSE/COPYING file in their tarballs; the authoritative
license evidence is the package.json license field (Apache-2.0) plus the
upstream monorepo LICENSE ("Copyright 2020 The Backstage Authors"; the
project was originally Spotify AB). `backstage/yarn.lock` is the
authoritative full transitive closure. The host Node.js runtime measured
2026-08-18: v22.23.2 (default, ~/.local/bin/node) and v24.19.0
(/opt/homebrew/opt/node@24/bin/node, the AGENTS.md path); both inside the
declared engines range "22 || 24" (backstage/package.json:6).

Framework, scaffold, and package manager (3 entries):

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | @backstage/create-app (scaffold tool) | unpinned (@latest at scaffold time; produced release line 1.49.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors (CNCF; originally Spotify AB) | scaffold origin | docs/specs/m1-substrate/technical-specification.md:693; scripts/install-m1.sh:270; THIRD-PARTY-LICENSES.md |
| 2 | Backstage framework release line | 1.49.1 | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | scaffold origin (release manifest) | backstage/backstage.json:2 |
| 3 | Yarn (vendored release) | 4.18.0 (bumped from 4.4.1 on 2026-08-18 to gain npmMinimalAgeGate support; npmMinimalAgeGate: "7d" set at backstage/.yarnrc.yml:5) | https://github.com/yarnpkg/berry | BSD-2-Clause | Yarn Contributors (2016-present) | vendored | backstage/package.json:113; backstage/.yarnrc.yml:3; backstage/.yarn/releases/yarn-4.18.0.cjs |

Root workspace devDependencies (11 entries; build/test tooling; installed
version in parentheses):

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 4 | @backstage/cli | ^0.36.0 (0.36.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | build dependency | backstage/package.json:39; packages/app/package.json:18; packages/backend/package.json:53 |
| 5 | @backstage/cli-defaults | ^0.1.0 (0.1.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | build dependency | backstage/package.json:40 |
| 6 | @backstage/e2e-test-utils | ^0.1.2 (0.1.2) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | build dependency (e2e) | backstage/package.json:41 |
| 7 | @jest/environment-jsdom-abstract | ^30.0.0 (30.3.0) | https://github.com/jestjs/jest | MIT | Meta Platforms, Inc. and affiliates | build dependency (test) | backstage/package.json:42 |
| 8 | @playwright/test | ^1.32.3 (1.59.1) | https://github.com/microsoft/playwright | Apache-2.0 | Microsoft Corporation | build dependency (e2e) | backstage/package.json:43; packages/app/package.json:49 |
| 9 | @types/jest | ^30.0.0 (30.0.0) | https://github.com/DefinitelyTyped/DefinitelyTyped | MIT | DefinitelyTyped contributors | build dependency (types) | backstage/package.json:44 |
| 10 | jest | ^30.2.0 (30.3.0) | https://github.com/jestjs/jest | MIT | Meta Platforms, Inc. and affiliates | build dependency (test) | backstage/package.json:45 |
| 11 | jsdom | ^27.1.0 (27.4.0) | https://github.com/jsdom/jsdom | MIT | Elijah Insua, Domenic Denicola, et al. | build dependency (test) | backstage/package.json:46 |
| 12 | node-gyp | ^10.0.0 (10.3.1) | https://github.com/nodejs/node-gyp | MIT | Nathan Rajlich / Node.js contributors | build dependency (native builds) | backstage/package.json:47; packages/backend/package.json:49 |
| 13 | prettier | ^2.3.2 (2.8.8) | https://github.com/prettier/prettier | MIT | James Long and contributors | build dependency (formatting) | backstage/package.json:48 |
| 14 | typescript | ~5.8.0 (5.8.3) | https://github.com/microsoft/TypeScript | Apache-2.0 | Microsoft Corp. | build dependency (language toolchain) | backstage/package.json:49 |

Playwright browser binaries (downloaded at `yarn test:e2e` time, per the
browsers.json of installed @playwright/test 1.59.1): Chromium
147.0.7727.15 r1217 plus chromium-headless-shell r1217 (BSD-3-Clause,
Copyright 2015 The Chromium Authors); Firefox 148.0.2 r1511 (MPL-2.0
primary; multi-license aggregate, ~80 sub-licenses including LGPL-2.1
components); WebKit 26.4 r2272 (portions LGPL and BSD-2-Clause, Apple);
ffmpeg r1011 (binary license not separately fetched). Each browser is a
multi-license aggregate.

App workspace dependencies (34 entries; frontend runtime plus its test
tooling):

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 15 | @backstage/core-components | ^0.18.8 (0.18.8) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:19 |
| 16 | @backstage/core-plugin-api | ^1.12.4 (1.12.4) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:20 |
| 17 | @backstage/frontend-defaults | ^0.5.0 (0.5.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:21 |
| 18 | @backstage/frontend-plugin-api | ^0.15.1 (0.15.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:22 |
| 19 | @backstage/integration-react | ^1.2.16 (1.2.16) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:23 |
| 20 | @backstage/plugin-api-docs | ^0.13.5 (0.13.5) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:24 |
| 21 | @backstage/plugin-app-react | ^0.2.1 (0.2.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:25 |
| 22 | @backstage/plugin-app-visualizer | ^0.2.1 (0.2.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:26 |
| 23 | @backstage/plugin-catalog | ^2.0.1 (2.0.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:27 |
| 24 | @backstage/plugin-catalog-graph | ^0.6.0 (0.6.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:28 |
| 25 | @backstage/plugin-catalog-import | ^0.13.11 (0.13.11) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:29 |
| 26 | @backstage/plugin-kubernetes | ^0.12.17 (0.12.17) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:30 |
| 27 | @backstage/plugin-notifications | ^0.5.15 (0.5.15) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:31 |
| 28 | @backstage/plugin-org | ^0.7.0 (0.7.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:32 |
| 29 | @backstage/plugin-scaffolder | ^1.36.1 (1.36.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:33 |
| 30 | @backstage/plugin-search | ^1.7.0 (1.7.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:34 |
| 31 | @backstage/plugin-signals | ^0.0.29 (0.0.29) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:35 |
| 32 | @backstage/plugin-techdocs | ^1.17.2 (1.17.2) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:36 |
| 33 | @backstage/plugin-techdocs-module-addons-contrib | ^1.1.34 (1.1.34) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:37 |
| 34 | @backstage/plugin-user-settings | ^0.9.1 (0.9.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:38 |
| 35 | @backstage/ui | ^0.13.1 (0.13.2) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/app/package.json:39 |
| 36 | @material-ui/core | ^4.12.2 (4.12.4) | https://github.com/mui/material-ui | MIT | Call-Em-All (Material-UI Team) | runtime dependency | packages/app/package.json:40 |
| 37 | @material-ui/icons | ^4.9.1 (4.11.3) | https://github.com/mui/material-ui | MIT | Call-Em-All (Material-UI Team) | runtime dependency | packages/app/package.json:41 |
| 38 | react | ^18.0.2 (18.3.1) | https://github.com/facebook/react | MIT | Facebook, Inc. and its affiliates (Meta) | runtime dependency | packages/app/package.json:42 |
| 39 | react-dom | ^18.0.2 (18.3.1) | https://github.com/facebook/react | MIT | Facebook, Inc. and its affiliates (Meta) | runtime dependency | packages/app/package.json:43 |
| 40 | react-router | ^6.30.4; resolution pin 6.30.4 (6.30.4) | https://github.com/remix-run/react-router | MIT | Remix Software | runtime dependency | packages/app/package.json:44; backstage/package.json:60 |
| 41 | react-router-dom | ^6.30.4; resolution pin 6.30.4 (6.30.4) | https://github.com/remix-run/react-router | MIT | Remix Software | runtime dependency | packages/app/package.json:45; backstage/package.json:61 |
| 42 | @backstage/frontend-test-utils (dev) | ^0.5.1 (0.5.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | build dependency (test) | packages/app/package.json:48 |
| 43 | @testing-library/dom (dev) | ^9.0.0 (9.3.4) | https://github.com/testing-library/dom-testing-library | MIT | Kent C. Dodds and contributors | build dependency (test) | packages/app/package.json:50 |
| 44 | @testing-library/jest-dom (dev) | ^6.0.0 (6.9.1) | https://github.com/testing-library/jest-dom | MIT | Testing Library contributors | build dependency (test) | packages/app/package.json:51 |
| 45 | @testing-library/react (dev) | ^14.0.0 (14.3.1) | https://github.com/testing-library/react-testing-library | MIT | Kent C. Dodds and contributors | build dependency (test) | packages/app/package.json:52 |
| 46 | @testing-library/user-event (dev) | ^14.0.0 (14.6.1) | https://github.com/testing-library/user-event | MIT | Testing Library contributors | build dependency (test) | packages/app/package.json:53 |
| 47 | @types/react-dom (dev) | *; resolution pin ^18 (18.3.7) | https://github.com/DefinitelyTyped/DefinitelyTyped | MIT | DefinitelyTyped contributors | build dependency (types) | packages/app/package.json:54; backstage/package.json:53 |
| 48 | cross-env (dev) | ^7.0.0 (7.0.3) | https://github.com/kentcdodds/cross-env | MIT | Kent C. Dodds | build dependency | packages/app/package.json:55 |

Backend workspace dependencies (30 entries; backend runtime):

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 49 | @backstage/backend-defaults | ^0.16.0 (0.16.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:19 |
| 50 | @backstage/config | ^1.3.6 (1.3.6) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:20 |
| 51 | @backstage/plugin-app-backend | ^0.5.12 (0.5.12) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:21 |
| 52 | @backstage/plugin-auth-backend | ^0.29.2 (0.29.2) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:22 |
| 53 | @backstage/plugin-auth-backend-module-github-provider | ^0.5.1 (0.5.1) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency (declared, not wired in index.ts) | packages/backend/package.json:23 |
| 54 | @backstage/plugin-auth-backend-module-guest-provider | ^0.2.17 (0.2.17) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:24 |
| 55 | @backstage/plugin-auth-node | ^0.6.14 (0.6.14) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency (basis of custom giteaAuth module) | packages/backend/package.json:25 |
| 56 | @backstage/plugin-catalog-backend | ^3.5.0 (3.5.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:26 |
| 57 | @backstage/plugin-catalog-backend-module-gitea | ^0.1.10 (0.1.10) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency (Gitea catalog integration) | packages/backend/package.json:27 |
| 58 | @backstage/plugin-catalog-backend-module-logs | ^0.1.20 (0.1.20) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:28 |
| 59 | @backstage/plugin-catalog-backend-module-scaffolder-entity-model | ^0.2.18 (0.2.18) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:29 |
| 60 | @backstage/plugin-kubernetes-backend | ^0.21.2 (0.21.2) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:30 |
| 61 | @backstage/plugin-notifications-backend | ^0.6.3 (0.6.3) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:31 |
| 62 | @backstage/plugin-permission-backend | ^0.7.10 (0.7.10) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:32 |
| 63 | @backstage/plugin-permission-backend-module-allow-all-policy | ^0.2.17 (0.2.17) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:33 |
| 64 | @backstage/plugin-permission-common | ^0.9.7 (0.9.7) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:34 |
| 65 | @backstage/plugin-permission-node | ^0.10.11 (0.10.11) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:35 |
| 66 | @backstage/plugin-proxy-backend | ^0.6.11 (0.6.11) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:36 |
| 67 | @backstage/plugin-scaffolder-backend | ^3.2.0 (3.3.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:37 |
| 68 | @backstage/plugin-scaffolder-backend-module-github (REMOVED 2026-08-19) | was ^0.9.7 (0.9.7) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | REMOVED: dead dependency -- declared but never registered since June; the Gitea module (entry #81) replaces it; gone from both package.json and yarn.lock (verified 2026-08-19) | formerly packages/backend/package.json:38 |
| 69 | @backstage/plugin-scaffolder-backend-module-notifications | ^0.1.20 (0.1.20) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:39 |
| 70 | @backstage/plugin-search-backend | ^2.1.0 (2.1.0) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:40 |
| 71 | @backstage/plugin-search-backend-module-catalog | ^0.3.13 (0.3.13) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:41 |
| 72 | @backstage/plugin-search-backend-module-pg | ^0.5.53 (0.5.53) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:42 |
| 73 | @backstage/plugin-search-backend-module-techdocs | ^0.4.12 (0.4.12) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:43 |
| 74 | @backstage/plugin-search-backend-node | ^1.4.2 (1.4.2) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:44 |
| 75 | @backstage/plugin-signals-backend | ^0.3.13 (0.3.13) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:45 |
| 76 | @backstage/plugin-techdocs-backend | ^2.1.6 (2.1.6) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency | packages/backend/package.json:46 |
| 77 | better-sqlite3 | ^12.0.0 (12.8.0) | https://github.com/WiseLibs/better-sqlite3 | MIT (bundles SQLite, public domain) | Joshua Wise | runtime dependency (dev database, app-config.yaml:43-45) | packages/backend/package.json:48 |
| 78 | pg (node-postgres) | ^8.11.3 (8.20.0) | https://github.com/brianc/node-postgres | MIT | Brian Carlson | runtime dependency (production database, app-config.production.yaml:19-26) | packages/backend/package.json:50 |
| 79 | @backstage/plugin-catalog-common | ^1.1.8 (1.1.8) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency (direct backend dependency since 2026-08-18; previously transitive/app-side) | packages/backend/package.json:30 |
| 80 | @backstage/plugin-kubernetes-common | ^0.9.10 (0.9.10) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency (direct backend dependency since 2026-08-18; previously transitive/app-side) | packages/backend/package.json:32 |
| 81 | @backstage/plugin-scaffolder-backend-module-gitea | ^0.2.19 (0.2.23) | https://github.com/backstage/backstage | Apache-2.0 | The Backstage Authors | runtime dependency (direct since 2026-08-19; provides the publish:gitea scaffolder action, registered at packages/backend/src/index.ts:17-19) | packages/backend/package.json:39; backstage/examples/template/template.yaml |

Resolution override pins (47 entries; security-override layer in
backstage/package.json:51-101 applying repo-wide to all workspaces;
transitive-only packages. react-router, react-router-dom, and
@types/react-dom also carry resolution pins and are listed in the app
table above. Copyright holders are taken from the installed tarball
LICENSE headers, verified 2026-08-18. Security remediation of
2026-08-18: bumped undici, brace-expansion, fast-uri, dompurify, js-yaml,
ip-address; added pins nanoid, file-type, typescript-json-schema (the
last drops vm2, eliminating it from the tree); removed the four dead
descriptor-scoped minimatch pins. The @octokit/* pins at
package.json:97-99 stay and are NOT dead config: @octokit/rest@19.0.13
remains a dependency of @backstage/integration, which also serves the
Gitea integration):

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 82 | @types/react | ^18 (18.3.28) | https://github.com/DefinitelyTyped/DefinitelyTyped | MIT | DefinitelyTyped contributors | build dependency (resolution override) | backstage/package.json:52 |
| 83 | @grpc/grpc-js | ^1.14.4 (1.14.4) | https://github.com/grpc/grpc-node | Apache-2.0 | Google Inc. | runtime dependency (resolution override) | backstage/package.json:54 |
| 84 | @babel/core | ^7.29.1 (7.29.7) | https://github.com/babel/babel | MIT | The Babel Team | build dependency (resolution override) | backstage/package.json:55 |
| 85 | @protobufjs/utf8 | ^1.1.1 (1.1.2) | https://github.com/protobufjs/protobuf.js | BSD-3-Clause | Daniel Wirtz (protobufjs) | runtime dependency (resolution override) | backstage/package.json:56 |
| 86 | ws | ^8.21.0 (8.21.0) | https://github.com/websockets/ws | MIT | Einar Otto Stangvik | runtime dependency (resolution override) | backstage/package.json:57 |
| 87 | axios | ^1.13.5 (1.19.0) | https://github.com/axios/axios | MIT | Matt Zabriskie and collaborators | runtime dependency (resolution override) | backstage/package.json:58 |
| 88 | undici | ^7.29.0 (7.29.0) | https://github.com/nodejs/undici | MIT | Node.js contributors | runtime dependency (resolution override) | backstage/package.json:59 |
| 89 | lodash | 4.18.1 | https://github.com/lodash/lodash | MIT | John-David Dalton / Lodash contributors | runtime dependency (resolution override) | backstage/package.json:62 |
| 90 | protobufjs | 7.6.5 | https://github.com/protobufjs/protobuf.js | BSD-3-Clause | Daniel Wirtz | runtime dependency (resolution override) | backstage/package.json:63 |
| 91 | shell-quote | 1.10.0 | https://github.com/ljharb/shell-quote | MIT | James Halliday | build dependency (resolution override) | backstage/package.json:64 |
| 92 | websocket-driver | 0.7.5 | https://github.com/faye/websocket-driver-node | Apache-2.0 | James Coglan | runtime dependency (resolution override) | backstage/package.json:65 |
| 93 | form-data | 4.0.6 | https://github.com/form-data/form-data | MIT | Felix Geisendoerfer and contributors | runtime dependency (resolution override) | backstage/package.json:66 |
| 94 | minimatch | 9.0.9 (exactly one installed copy; exactly one yarn.lock resolution) | https://github.com/isaacs/minimatch | ISC | Isaac Z. Schlueter | build dependency (resolution override) | backstage/package.json:89; the descriptor-scoped pins formerly at :67-70 were dead config and were removed in the 2026-08-18 remediation |
| 95 | immutable | 4.3.9 | https://github.com/immutable-js/immutable-js | MIT | Lee Byron and contributors | runtime dependency (resolution override) | backstage/package.json:67 |
| 96 | tar | 7.5.22 | https://github.com/isaacs/node-tar | BlueOak-1.0.0 | Isaac Z. Schlueter | build dependency (resolution override) | backstage/package.json:68 |
| 97 | linkify-it | 5.0.2 | https://github.com/markdown-it/linkify-it | MIT | Vitaly Puzrin (2015) | runtime dependency (resolution override; re-pinned from 6.1.0 on 2026-08-20 -- markdown-it 14.2.0 type declarations break yarn build:all against v6; 5.0.2 satisfies its ^5.0.1 range, audit clean) | backstage/package.json:69 |
| 98 | fast-xml-builder | 1.3.0 | https://github.com/NaturalIntelligence/fast-xml-builder | MIT | Natural Intelligence (2026); author Amit Gupta | runtime dependency (resolution override) | backstage/package.json:70 |
| 99 | brace-expansion | 5.0.9 | https://github.com/juliangruber/brace-expansion | MIT | Julian Gruber; TypeScript port Isaac Z. Schlueter | build dependency (resolution override) | backstage/package.json:71 |
| 100 | fast-uri | 4.1.2 | https://github.com/fastify/fast-uri | BSD-3-Clause | Vincent Le Goff | runtime dependency (resolution override) | backstage/package.json:72 |
| 101 | postcss | 8.5.25 | https://github.com/postcss/postcss | MIT | Andrey Sitnik | build dependency (resolution override) | backstage/package.json:73 |
| 102 | basic-ftp | 6.0.2 | https://github.com/patrickjuchli/basic-ftp | MIT | Patrick Juchli | runtime dependency (resolution override) | backstage/package.json:74 |
| 103 | js-cookie | 3.0.8 | https://github.com/js-cookie/js-cookie | MIT | Klaus Hartl | runtime dependency (resolution override) | backstage/package.json:75 |
| 104 | multer | 2.2.0 | https://github.com/expressjs/multer | MIT | Hage Yaapa, Jaret Pfluger, et al. | runtime dependency (resolution override) | backstage/package.json:76 |
| 105 | http-proxy-middleware | 2.0.10 | https://github.com/chimurai/http-proxy-middleware | MIT | Steven Chim | build dependency (resolution override) | backstage/package.json:77 |
| 106 | dompurify | 3.4.13 | https://github.com/cure53/DOMPurify | (MPL-2.0 OR Apache-2.0) | Mario Heiderich, Cure53 | runtime dependency (resolution override) | backstage/package.json:78 |
| 107 | follow-redirects | 1.16.0 | https://github.com/follow-redirects/follow-redirects | MIT | Ruben Verborgh and contributors | runtime dependency (resolution override) | backstage/package.json:79 |
| 108 | qs | 6.15.3 | https://github.com/ljharb/qs | BSD-3-Clause | Jordan Harband | runtime dependency (resolution override) | backstage/package.json:80 |
| 109 | uuid | 11.1.1 | https://github.com/uuidjs/uuid | MIT | Robert Kieffer and other contributors (2010-2020) | runtime dependency (resolution override) | backstage/package.json:81 |
| 110 | svgo | 3.3.4 | https://github.com/svg/svgo | MIT | Kir Belevich and contributors | build dependency (resolution override) | backstage/package.json:82 |
| 111 | koa | 2.16.4 | https://github.com/koajs/koa | MIT | Koa contributors (LICENSE: (c) 2019 Koa contributors) | runtime dependency (resolution override) | backstage/package.json:83 |
| 112 | adm-zip | 0.6.0 | https://github.com/cthackers/adm-zip | MIT | Nasca Iacob (cthackers) | build dependency (resolution override) | backstage/package.json:84 |
| 113 | js-yaml | 4.3.1 | https://github.com/nodeca/js-yaml | MIT | Vladimir Zapparov and contributors | runtime dependency (resolution override) | backstage/package.json:85 |
| 114 | webpack-dev-server | 5.2.6 | https://github.com/webpack/webpack-dev-server | MIT | Tobias Koppers (webpack contributors) | build dependency (resolution override) | backstage/package.json:86 |
| 115 | prismjs | 1.30.0 | https://github.com/PrismJS/prism | MIT | Lea Verou and contributors | runtime dependency (resolution override) | backstage/package.json:87 |
| 116 | body-parser | 1.20.6 | https://github.com/expressjs/body-parser | MIT | Douglas Christopher Wilson, Jonathan Ong, et al. | runtime dependency (resolution override) | backstage/package.json:88 |
| 117 | fast-xml-parser | 5.7.0 | https://github.com/NaturalIntelligence/fast-xml-parser | MIT | Amit Gupta (NaturalIntelligence) | runtime dependency (resolution override) | backstage/package.json:90 |
| 118 | markdown-it | 14.2.0 | https://github.com/markdown-it/markdown-it | MIT | Vitaly Puzrin, Alex Kocharin (LICENSE: (c) 2014) | runtime dependency (resolution override) | backstage/package.json:91 |
| 119 | launch-editor | 2.14.1 | https://github.com/yyx990803/launch-editor | MIT | Evan You | build dependency (resolution override) | backstage/package.json:92 |
| 120 | ip-address | 10.3.1 | https://github.com/beaugunderson/ip-address | MIT | Beau Gunderson | runtime dependency (resolution override) | backstage/package.json:93 |
| 121 | nanoid | 3.3.18 (pin added 2026-08-18) | https://github.com/ai/nanoid | MIT | Andrey Sitnik | runtime dependency (resolution override) | backstage/package.json:94 |
| 122 | file-type | 21.3.2 (pin added 2026-08-18; ESM-only, and the only dependent never imports it) | https://github.com/sindresorhus/file-type | MIT | Sindre Sorhus | build dependency (resolution override) | backstage/package.json:95 |
| 123 | typescript-json-schema | 0.68.0 (pin added 2026-08-18; this pin drops vm2, so vm2 is eliminated from the tree -- it was never a listed entry) | https://github.com/YousefED/typescript-json-schema | BSD-3-Clause | Yousef El-Dardiry and Dominik Moritz | build dependency (resolution override) | backstage/package.json:96 |
| 124 | @octokit/request | 8.4.1 | https://github.com/octokit/request.js | MIT | Gregor Martynus (Octokit) | runtime dependency (resolution override) | backstage/package.json:97 |
| 125 | @octokit/request-error | 5.1.1 | https://github.com/octokit/request-error.js | MIT | Octokit contributors (LICENSE-verified: Copyright (c) 2019 Octokit contributors; installed author field Gregor Martynus) | runtime dependency (resolution override) | backstage/package.json:98 |
| 126 | @octokit/plugin-paginate-rest | 11.4.1 | https://github.com/octokit/plugin-paginate-rest.js | MIT | Octokit contributors (LICENSE-verified: Copyright (c) 2019 Octokit contributors) | runtime dependency (resolution override) | backstage/package.json:99 |
| 127 | @tootallnate/once | 2.0.1 | https://github.com/TooTallNate/once | MIT | Nathan Rajlich | build dependency (resolution override) | backstage/package.json:100 |
| 128 | esbuild | 0.28.1 | https://github.com/evanw/esbuild | MIT | Evan Wallace (LICENSE: (c) 2020 Evan Wallace) | build dependency (resolution override) | backstage/package.json:101 |

## 6. CI/CD Actions and Images (9 entries)

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | actions/checkout | pinned SHA 11d5960a326750d5838078e36cf38b85af677262 (v4.4.0, lightweight tag; pinned 2026-08-18, previously the mutable @v4 tag) | https://github.com/actions/checkout | MIT | GitHub, Inc. and contributors (2018) | fetched at install time (CI action) | iac/templates/ci.yaml:13; seed-repos/hello-m2/.gitea/workflows/ci.yaml:13 |
| 2 | actions/setup-go | pinned SHA 40f1582b2485089dde7abd97c1529aa768e1baff (v5.6.0; pinned 2026-08-18, previously the mutable @v5 tag); installs Go 1.26 | https://github.com/actions/setup-go | MIT | GitHub, Inc. and contributors (2018) | fetched at install time (CI action) | iac/templates/ci.yaml:16-18; seed-repos/hello-m2/.gitea/workflows/ci.yaml:16-18 |
| 3 | opentofu/setup-opentofu | pinned SHA 9d84900f3238fab8cd84ce47d658d25dd008be2f (v1.0.8; pinned 2026-08-18, previously the mutable @v1 tag); installs OpenTofu 1.9.0 | https://github.com/opentofu/setup-opentofu | MPL-2.0 | OpenTofu Authors (LICENSE records Copyright (c) 2020 HashiCorp, Inc., inherited from forked hashicorp/setup-terraform) | fetched at install time (CI action) | iac/templates/ci.yaml:31; seed-repos/hello-m2/.gitea/workflows/ci.yaml:31 |
| 4 | golang (Docker Official Image) | tag 1.26-alpine (no digest) | https://hub.docker.com/_/golang | BSD-3-Clause (Go toolchain) | The Go Authors / Google LLC | runtime image (CI build stage) | seed-repos/hello-m2/Dockerfile:1 |
| 5 | alpine (Docker Official Image) | tag 3.24 (no digest; floating tag, snapshot 2026-08-21) | https://hub.docker.com/_/alpine | Multi-license aggregate; Alpine 3.24.1, 16 packages enumerated 2026-08-21 by pulling the image and reading apk metadata (apk info --license): GPL-2.0-only (alpine-baselayout 3.7.2-r1, alpine-baselayout-data 3.7.2-r1, apk-tools 3.0.6-r0, busybox 1.37.0-r31, busybox-binsh 1.37.0-r31, libapk 3.0.6-r0, scanelf 1.3.9-r1, ssl_client 1.37.0-r31), MIT (alpine-keys 2.6-r0, alpine-release 3.24.1-r0, musl 1.2.6-r2), MPL-2.0 AND MIT (ca-certificates-bundle 20260611-r0), Apache-2.0 (libcrypto3 3.5.7-r0, libssl3 3.5.7-r0), MIT AND BSD-2-Clause AND GPL-2.0-or-later (musl-utils 1.2.6-r2), Zlib (zlib 1.3.2-r0) | Alpine Linux development team | runtime image (hello-m2 runtime stage; bumped from EOL 3.20 on 2026-08-21) | seed-repos/hello-m2/Dockerfile:6 |
| 6 | node (Docker Official Image) | tag 24-trixie-slim (no digest; floating tag, snapshot 2026-08-18) | https://hub.docker.com/_/node + https://github.com/nodejs/docker-node | MIT (image build files and Node.js runtime; runtime bundles third-party libraries under their own licenses incl. Apache-2.0, Unicode-3.0, BSD variants); Debian trixie-slim base layer: debootstrap minbase from Debian main only (all packages DFSG-free per the Debian Social Contract), per-package list not enumerated (PARTIAL) | Joyent, Inc. and Node.js contributors; Debian Project (base layer, not enumerated) | runtime image (Backstage backend image base; pulled at docker build time via yarn build-image; upstream Dockerfile FROM debian:trixie-slim, NODE_VERSION=24.19.0) | backstage/packages/backend/Dockerfile:15; backstage/packages/backend/package.json:16 |
| 7 | Trivy | image aquasec/trivy:0.74.0@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969 (digest verified against the registry manifest for tag 0.74.0, 2026-08-18) | https://github.com/aquasecurity/trivy | Apache-2.0 | Aqua Security (upstream org; stock Apache-2.0 LICENSE, no filled copyright line) | runtime image (CI scanning gates: filesystem and image scans with SARIF artifacts); digest-pinned following the March 2026 Trivy supply-chain compromise (CVE-2026-33634) | seed-repos/hello-m2/.gitea/workflows/ci.yaml:29-45,92-107; iac/templates/ci.yaml:29-45,77-92 |
| 8 | OSV-Scanner | image ghcr.io/google/osv-scanner:v2.5.1@sha256:8108ae94eadea5a02c9bec6e646909d5b790b44bd62d7f5b7f0b1d6d0ffc7734 (digest verified against the registry manifest for tag v2.5.1, 2026-08-18) | https://github.com/google/osv-scanner | Apache-2.0 | Google | runtime image (CI dependency scan gate; suppression file seed-repos/hello-m2/osv-scanner.toml) | seed-repos/hello-m2/.gitea/workflows/ci.yaml:44; iac/templates/ci.yaml:44; seed-repos/hello-m2/osv-scanner.toml |
| 9 | CodeQL action (github/codeql-action) (REMOVED 2026-08-21) | was pinned SHA ff2f1c621b7f889edc0d3c761ac2e6a3f8cdb0dd (annotated tag v4.37.7 peeled commit; re-verified via ls-remote 2026-08-18) | https://github.com/github/codeql-action | MIT | GitHub (LICENSE: Copyright (c) 2020 GitHub) | REMOVED: the org's "GitHub recommended" code-security configuration enables CodeQL default setup, which rejects SARIF from an in-repo advanced workflow; code-scanning.yml removed, GitHub-side scanning now owned by the org default setup (GitHub-managed, MIT action remains its upstream; the CodeQL CLI terms note still applies to that managed path) | formerly .github/workflows/code-scanning.yml |

Base images in this group remain tag-only (no digests); the three actions
were pinned to commit SHAs and the two scanner images (Trivy, OSV-Scanner)
are digest-pinned as of 2026-08-18 (see the entries above).
`.github/workflows/sync-from-gitea.yml` uses no third-party GitHub
Actions. jq, curl, docker, and git in CI steps are runner-image-provided
binaries with no repo-level pin.

Base-layer detail (snapshot 2026-08-21): since the 2026-08-21 bump, both
Dockerfile stages ride the same Alpine line -- golang:1.26-alpine is
Alpine 3.24.1-based and the runtime stage is alpine:3.24 (3.24.1, 16
packages, enumerated in entry 5; supersedes the 2026-08-18 note that the
two stages were on different Alpine lines). alpine:3.24,
golang:1.26-alpine, and node:24-trixie-slim are floating tags; the
enumerations here are a snapshot and drift as tags move -- digest
pinning would close this.

## 7. Schemas and Specifications (1 entry)

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | Score specification and JSON schema (score.dev/v1b1) | schema vendored: SHA256 633c4394dfc03977c86932f5d10c77e2c356a38b63f162c281104fade5c9863c, 16701 bytes, pinned 2026-04-23; upstream commit 3ecb17d430c2bbf46d2dfc161fabc7d432d6d1f5 (2026-04-17), upstream path score-v1b1.json (repository root), byte-identical match verified 2026-08-18 | https://github.com/score-spec/spec | Apache-2.0 | The Score Authors (2022); CNCF sandbox (originated at Humanitec) | vendored (schema) + referenced (spec apiVersion in score.yaml) | tools/score2openchoreo/assets/score.schema.json; assets/SCHEMA_PROVENANCE.md:8-13; schema_pin_test.go:13; schema.go:19-20; seed-repos/hello-m2/score.yaml:1 |

## 8. Documentation Tooling (2 entries)

| # | Component | Version/pin | Upstream URL | License (SPDX) | Copyright holder | Usage mode | Evidence |
|---|---|---|---|---|---|---|---|
| 1 | MkDocs | unpinned | https://github.com/mkdocs/mkdocs | BSD-2-Clause | Tom Christie (2014-present) and contributors | fetched at install time (TechDocs docs build) | mkdocs.yml:1-2 |
| 2 | mkdocs-techdocs-core | unpinned | https://github.com/backstage/mkdocs-techdocs-core | Apache-2.0 | The Backstage Authors (2020-21) | fetched at install time (TechDocs docs build) | mkdocs.yml:3-4 |

TechDocs generator image (pulled at docs build time; recorded here as
prose, not counted as a listed entry): spotify/techdocs default tag
v1.2.8 (TechdocsGenerator.defaultDockerImage in installed
@backstage/plugin-techdocs-node 1.14.4; backstage/app-config.yaml:82-83
sets runIn: docker with no dockerImage override), built from
github.com/backstage/techdocs-container (main), Apache-2.0, Copyright
2020 Spotify AB. The image is FROM python:3.14-alpine and bundles the
PlantUML v1.2026.2 jar and pip mkdocs-techdocs-core==1.7.0; Docker Hub
tag v1.2.8 exists (pushed 2026-01-19). Per-component licenses inside the
image are not enumerated (see U19 below).

---

## UNVERIFIED items and honest gaps

An honest gap is recorded here rather than hidden. Each row states what
could not be verified, what IS verified, and where the evidence sits.
Nothing in this section is counted among the 198 listed entries above
unless explicitly noted as a caveat on a listed entry.

Resolution passes of 2026-08-18: a first pass resolved 15 rows with hard
evidence (U1-U6, U10, U14, U15, U18, U20-U24) and narrowed four (U7,
U11, U19, U25); a second pass with the cluster up resolved the seven
cluster-blocked rows with live read-only evidence (U8, U9, U11, U12,
U13, U16, U17), their values folded into the listed entries above.
Three rows remain (U7, U19, U25). Numbering convention: the original
U-numbers are retained, so the register is non-contiguous; this preserves
cross-references from earlier session documents.

Resolved-findings recap, first pass (one-liners; detail lives in the
listed entries): all resolution-override copyright holders are now
tarball-LICENSE-verified (including fast-xml-builder = Natural
Intelligence (2026), author Amit Gupta, upstream
github.com/NaturalIntelligence/fast-xml-builder); the vendored Score
schema is byte-identical to score-spec/spec commit
3ecb17d430c2bbf46d2dfc161fabc7d432d6d1f5 (score-v1b1.json at the
repository root); alpine:3.24 is fully enumerated (16 packages, Alpine
3.24.1; supersedes the 2026-08-18 alpine:3.20 enumeration after the
EOL-driven runtime bump); golang:1.26-alpine rides the same Alpine
3.24.1 line; the descriptor-scoped minimatch pins at
backstage/package.json:67-70 are dead config shadowed by the bare pin at
:93 (exactly one installed copy and one yarn.lock resolution exist); all
113 installed @backstage/* packages confirmed to ship no LICENSE file;
plugin.json/hooks.json are confirmed absent (README.md drift; actual
registration lives in .claude-plugin/marketplace.json, itself stale);
assessments/report.json files contain no embedded tool attribution.

Resolved-findings recap, second pass (cluster-live, 2026-08-18): deployed
Gitea chart gitea-12.5.0, app 1.25.4, image
docker.gitea.com/gitea:1.25.4-rootless; Envoy data-plane
docker.io/envoyproxy/envoy:distroless-v1.33.0 with control plane
docker.io/envoyproxy/gateway:v1.3.1 (matches the chart pin); act-runner
chart actions-0.1.0 (app 0.261.3) in namespace gitea-runners with runner
image docker.gitea.com/act_runner:0.3.1 and dind
docker.io/docker:29.4.0-dind; deployed k3s server v1.32.9+k3s1
(explicitly distinct from the k3d 5.9.0 default v1.35.5-k3s1, which is
NOT the deployed version); Gatekeeper deployed v3.17.1 (chart
gatekeeper-3.17.1), matching the repo pin with no drift; cert-manager
v1.19.4 (chart cert-manager-v1.19.4; controller/cainjector/webhook
images v1.19.4; namespace cert-manager; deployed 2026-05-01; Apache-2.0
per the upstream GitHub license field) and Argo Workflows v3.6.2
(namespace openchoreo-workflow-plane; quay.io/argoproj/argocli:v3.6.2 and
quay.io/argoproj/workflow-controller:v3.6.2; Apache-2.0 per the upstream
GitHub license field) confirmed live as sibling-openchoreo-managed
components -- both remain outside this repo's component scope (no repo
references), recorded here as live context only. Residual caveat carried
forward from U17: the catthehacker/ubuntu:act-* CI job image exists only
while a CI job is in flight (zero matches in a full-cluster scan
2026-08-18); its tag is observable only during a CI run.

| # | Item | What is UNVERIFIED | What IS verified | Evidence / notes |
|---|---|---|---|---|
| U7 | OpenChoreo mirror in namespace-predictor (listed entry, group 4 #4) | the exact mirrored upstream commit revision (none recorded in-file; sibling checkout verified absent on this machine, 2026-08-18) | holder and license confirmed via upstream SPDX headers (Copyright 2025 The OpenChoreo Authors, Apache-2.0) in fetched name.go and namespace_handler.go; both cited upstream paths exist; algorithm re-diffed against upstream main 2026-08-18 and still behavior-identical (upstream refactored to an extraHashInput-capable generateK8sName; the mirrored subset is unaffected) | tools/namespace-predictor/main.go:18,32,43-45,83-88; note: the main.go:18 comment "kubernetes.name.go" was a typo for internal/dataplane/kubernetes/name.go; corrected 2026-08-18 in a comment-only edit (tools/namespace-predictor/main.go:18, kubernetes.name.go -> kubernetes/name.go) |
| U19 | spotify/techdocs generator image | per-component licenses inside the image (Python, Alpine packages, OpenJDK, PlantUML, graphviz) not enumerated | default tag v1.2.8 verified (TechdocsGenerator.defaultDockerImage in installed @backstage/plugin-techdocs-node 1.14.4 and upstream; no dockerImage override in app-config); built from github.com/backstage/techdocs-container (main), Apache-2.0, Copyright 2020 Spotify AB; image is FROM python:3.14-alpine and bundles PlantUML v1.2026.2 jar and pip mkdocs-techdocs-core==1.7.0; Docker Hub tag v1.2.8 exists (pushed 2026-01-19) | backstage/app-config.yaml:82-83 |
| U25 | node:24-trixie-slim Debian base layer (listed entry, group 6 #6) | exact per-package list and license names of the deployed layer (daemon down, pull disallowed) | upstream Dockerfile is FROM debian:trixie-slim (NODE_VERSION=24.19.0); debian:trixie-slim is debootstrap minbase built from Debian main only, so every package is DFSG-free per the Debian Social Contract (clauses 1 and 5, debian.org/social_contract); per-package license texts are mechanically extractable at /usr/share/doc/<pkg>/copyright per Debian Policy 12.5; current trixie-slim = 13.6-slim (snapshot trixie-20260803) | backstage/packages/backend/Dockerfile:15 |

## Accepted residual risks (documented, not unverified)

- react-router / react-router-dom 6.30.4 (listed entries, group 5 #40-41;
  also exact resolution pins at backstage/package.json:60-61): three
  moderate advisories -- GHSA-wrjc-x8rr-h8h6 (>=6.0.0 <7.18.0),
  GHSA-337j-9hxr-rhxg (>=6.4.0 <7.18.0), GHSA-jjmj-jmhj-qwj2
  (>=6.30.2 <=6.30.4). No fixed 6.x exists (max published 6.x = 6.30.4,
  verified via npm view 2026-08-18); v7 is outside every installed
  @backstage/* peer range (^6.30.2), so forcing v7 was rejected as an
  unsupported major against all Backstage peer declarations. Remediation
  path: upstream Backstage shipping v7-compatible peer ranges. Recorded
  as an accepted residual risk 2026-08-18.

## First-party works (for completeness, not third-party)

The following are original work by this project and carry no third-party
attribution duty: the six policy guards under plugins/rr-policy-guards/
(MIT, project-owned, stdlib-only) and their sibling audit-log verifier
plugins/rr-policy-guards/tools/audit-chain/ (rr-audit-chain;
project-owned, stdlib-only, go 1.21); tools/score2openchoreo and
tools/namespace-predictor (project code; the latter embeds the OpenChoreo
algorithm replica recorded as group 4 #4); policies/*.rego (v0 syntax);
scripts/ci/*.sh; seed-repos/platform-addons/ kustomize and Gatekeeper
constraint content; seed-repos/platform-config/ (empty environment dirs);
root and hello-m2 catalog-info.yaml; index.html (self-contained);
scripts/gitea-values.yaml and observability/**/values*.yaml (values files
for third-party charts); .gitleaksignore (first-party false-positive
fingerprints for gitleaks, added 2026-08-18). External services used but
not incorporated
(gitea.com SaaS, GitHub, Vercel) are out of scope for this listing.

## Host custody artifacts (G5, 2026-08-26)

Non-repo artifacts held under ~/.rational-reserve/ (outside all remotes):
openbao/unseal-key and openbao/root-token (dir 700, files 600) -- Shamir
1-of-1 unseal key and generated root token for the Raft-backed OpenBao
(BAO-STORAGE-DES-001 D-05); backups/openbao/*.snap (mode 600) -- raft
snapshots written by scripts/backup-openbao.sh. These are secrets custody,
not third-party works; recorded here so their existence and location are
auditable.
