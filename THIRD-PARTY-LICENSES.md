# Third-Party Licenses

This project incorporates or depends on the following third-party software.
Each retains its original license and copyright. This project claims no
authorship over any of them.

Companion artifacts:

- `provenance/PROVENANCE.md` -- per-component provenance listing: versions,
  usage mode, and the repo evidence paths proving both.
- `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md` -- credentialised
  recognition certificate carrying integrity digests of this file and the
  listing.

Authoritative dependency closure: this file is a human-readable summary of
direct and significant components. The machine-readable lockfiles are the
authoritative complete dependency closure for their toolchains:

- `backstage/yarn.lock` -- full Node.js transitive closure (lockfile
  metadata version 8).
- `tools/score2openchoreo/go.sum` and `seed-repos/hello-m2/go.sum` -- Go
  module closures.
- `iac/.terraform.lock.hcl` -- OpenTofu provider closure (hash-pinned).

Anything that could not be fully verified is marked UNVERIFIED rather than
guessed; see the summary at the end and the UNVERIFIED subsection of
`provenance/PROVENANCE.md`.

---

## Platform and Infrastructure

### OpenTofu

- **Source:** https://github.com/opentofu/opentofu
- **License:** MPL-2.0
- **Copyright:** The OpenTofu Authors / Linux Foundation; upstream LICENSE
  records Copyright (c) 2014 HashiCorp, Inc.
- **Usage:** IaC tool for everything under `iac/`. Constraint
  `>= 1.9.0, < 1.12.0` (iac/versions.tf:2); CI template pins
  `tofu_version: 1.9.0` (iac/templates/ci.yaml:31-33); host brew install is
  unpinned (scripts/install-m2.sh:20).

### OpenChoreo

- **Source:** https://github.com/openchoreo/openchoreo
- **License:** Apache-2.0
- **Copyright:** OpenChoreo authors
- **Usage:** Environment CRDs (`openchoreo.dev/v1alpha1`) applied by
  iac/modules/openchoreo-environments/main.tf:18-24. The platform planes
  themselves are deployed by the sibling openchoreo checkout, not by this
  repo. Unpinned here.

### OpenBao

- **Source:** https://github.com/openbao/openbao
- **License:** MPL-2.0
- **Copyright:** OpenBao contributors / Linux Foundation
- **Usage:** Secret store consumed via ClusterSecretStore
  (iac/modules/external-secrets-wiring/main.tf:25; scripts/smoke-openbao.sh;
  scripts/seed-openbao-m2-paths.sh). Deployed in dev mode by the sibling
  openchoreo checkout; unpinned here.

### External Secrets Operator

- **Source:** https://github.com/external-secrets/external-secrets
- **License:** Apache-2.0
- **Copyright:** External Secrets Authors / CNCF
- **Usage:** `external-secrets.io/v1` CRDs used by
  iac/modules/external-secrets-wiring/main.tf:18-22 and
  iac/modules/gitea-runner/main.tf:8. The operator is deployed by the
  sibling openchoreo checkout; unpinned here.

### k3d

- **Source:** https://github.com/k3d-io/k3d
- **License:** MIT
- **Copyright:** k3d authors (Rancher)
- **Usage:** Cluster bootstrap tool. The `k3d-openchoreo` cluster is created
  by the sibling openchoreo checkout and must pre-exist
  (scripts/install-m1.sh:79-84). Unpinned.

### k3s

- **Source:** https://github.com/k3s-io/k3s
- **License:** Apache-2.0
- **Copyright:** Rancher / SUSE / CNCF
- **Usage:** Kubernetes distribution inside the k3d cluster, deployed by the
  sibling openchoreo checkout. Referenced via `/etc/rancher/k3s/registries.yaml`
  handling in scripts/install-m1.sh:97,111,117. No k3s source or images are
  included in this repository. Deployed server version v1.32.9+k3s1
  (kubectl version, 2026-08-18) -- explicitly distinct from the k3d 5.9.0
  default v1.35.5-k3s1, which is NOT the deployed version.

### kubectl

- **Source:** https://github.com/kubernetes/kubernetes
- **License:** Apache-2.0
- **Copyright:** The Kubernetes Authors / CNCF
- **Usage:** CLI used throughout the scripts (scripts/install-m1.sh:84-123)
  and as the transport for the tofu `kubernetes` state backend
  (iac/backend.tf:1-9). Unpinned.

### Helm

- **Source:** https://github.com/helm/helm
- **License:** Apache-2.0
- **Copyright:** Helm Authors / CNCF
- **Usage:** CLI used by install scripts (scripts/install-m1.sh:134-157;
  scripts/install-m3.sh:29-31) and by the tofu helm provider. Unpinned.

### Flux CLI

- **Source:** https://github.com/fluxcd/flux2
- **License:** Apache-2.0
- **Copyright:** The Flux authors / CNCF
- **Usage:** Host CLI installed via brew (scripts/install-m2.sh:21) for the
  Flux deployment managed in iac/modules/flux. Unpinned.

### Infracost CLI

- **Source:** https://github.com/infracost/infracost
- **License:** Apache-2.0
- **Copyright:** Infracost Inc. (2021)
- **Usage:** Host tool (scripts/install-m2.sh:22; scripts/smoke-infracost.sh:14)
  and CI step (iac/templates/ci.yaml:45; the seed workflow fetches
  `releases/latest`, seed-repos/hello-m2/.gitea/workflows/ci.yaml:51).
  Unpinned.

### score-k8s CLI

- **Source:** https://github.com/score-spec/score-k8s
- **License:** Apache-2.0
- **Copyright:** Score authors / Humanitec
- **Usage:** Host CLI installed via brew (scripts/install-m2.sh:23).
  Unpinned.

### Yarn classic (host tool)

- **Source:** https://github.com/yarnpkg/yarn
- **License:** BSD-2-Clause
- **Copyright:** Yarn Contributors
- **Usage:** Host bootstrap tool installed via brew
  (scripts/install-m1.sh:66-67); the brew `yarn` formula tracks the
  classic 1.x line. Unpinned. Distinct from the vendored Yarn 4.4.1 used
  inside `backstage/` (see Backstage and Node.js).

### Docker / Moby

- **Source:** https://github.com/moby/moby
- **License:** Apache-2.0
- **Copyright:** Docker, Inc.
- **Usage:** `docker build`/`push` in CI (iac/templates/ci.yaml:57,61) and
  the dind sidecar in the act-runner (iac/modules/gitea-runner/main.tf:58-61).
  Unpinned.

### jq

- **Source:** https://github.com/jqlang/jq
- **License:** MIT (documentation CC-BY-3.0)
- **Copyright:** Stephen Dolan
- **Usage:** JSON parsing in the CI template (iac/templates/ci.yaml:46-47)
  and in scripts/ci/*.sh. Provided by the CI runner image, unpinned.

### OPA (Open Policy Agent)

- **Source:** https://github.com/open-policy-agent/opa
- **License:** Apache-2.0
- **Copyright:** OPA authors / CNCF (originally Styra)
- **Usage:** Test runtime for the Rego policies under `policies/`
  (`opa test --v0-compatible`, policies/README.md:16-20). "OPA 1.x" is a
  prose reference, not a pin.

## Helm Charts and Container Images

Nothing in this group is vendored: charts are fetched by helm/tofu at
install time from upstream chart repos or OCI registries, and images are
pulled by the cluster at runtime. Zookeeper and Alertmanager are explicitly
disabled in the values files and are not deployed.

### Flux CD (helm chart `flux2`)

- **Source:** https://github.com/fluxcd/flux2 and
  https://github.com/fluxcd-community/helm-charts
- **License:** Apache-2.0
- **Copyright:** The Flux authors / CNCF
- **Usage:** Chart 2.13.0 deployed via iac/modules/flux/main.tf:8-10.
  Drift correction for cluster add-ons only; OpenChoreo remains the
  workload deployer.

### OPA Gatekeeper (helm chart `gatekeeper`)

- **Source:** https://github.com/open-policy-agent/gatekeeper
- **License:** Apache-2.0
- **Copyright:** Gatekeeper authors / CNCF (originally Microsoft)
- **Usage:** Chart 3.17.1 deployed via iac/modules/gatekeeper/main.tf:8-10.
  Admission runtime for the constraints in `policies/` and
  `seed-repos/platform-addons/`. Deployed v3.17.1 confirmed live
  2026-08-18 (chart gatekeeper-3.17.1; images
  openpolicyagent/gatekeeper:v3.17.1 on audit and controller-manager) --
  matches the repo pin, no drift.

### Gitea Actions runner (act_runner)

- **Source:** https://gitea.com/gitea/act_runner
- **License:** MIT
- **Copyright:** The Gitea Authors (2022)
- **Usage:** Chart `actions` 0.1.0 from dl.gitea.com/charts, deployed via
  iac/modules/gitea-runner/main.tf:26-34. Deployed live (measured
  2026-08-18, namespace gitea-runners): chart actions-0.1.0 (app
  0.261.3), runner image docker.gitea.com/act_runner:0.3.1 (StatefulSet
  act-runner-actions-act-runner), dind sidecar
  docker.io/docker:29.4.0-dind. Residual caveat: the
  catthehacker/ubuntu:act-* job image exists only while a CI job is in
  flight (zero matches in a full-cluster scan 2026-08-18); its tag is
  observable only during a CI run.

### Gitea (helm chart `gitea-charts/gitea`)

- **Source:** https://gitea.com/gitea/helm-chart and
  https://github.com/go-gitea/gitea
- **License:** MIT
- **Copyright:** The Gitea Authors (chart also credits NOVUM-RGI, Charlie
  Drage, John Felten)
- **Usage:** Chart unpinned in this repo (no `--version`), installed by
  scripts/install-m1.sh:134,153-157 with values file
  `scripts/gitea-values.yaml` (original work). Deployed live (measured
  2026-08-18): chart gitea-12.5.0, app version 1.25.4, image
  docker.gitea.com/gitea:1.25.4-rootless; the repo does not pin the
  chart, so the deployed version is the live evidence. Chart sub-chart
  dependencies measured in the gitea namespace 2026-08-18:
  docker.io/bitnamilegacy/postgresql:17.6.0-debian-12-r4 and
  docker.io/bitnamilegacy/valkey-cluster:8.1.3-debian-12-r3.

### SigNoz

- **Source:** https://github.com/SigNoz/signoz and
  https://github.com/SigNoz/charts
- **License:** MIT (community edition; the `ee/` and `cmd/enterprise/`
  directories ship under a separate enterprise license)
- **Copyright:** SigNoz Inc (2020-present)
- **Usage:** Chart 0.130.1 (iac/modules/observability/main.tf:17-21;
  variables.tf:4), values in observability/signoz/values.local.yaml. A
  local workaround removes enterprise-only OpAMP args
  (iac/modules/observability/main.tf:43-59).

### ClickHouse

- **Source:** https://github.com/ClickHouse/ClickHouse
- **License:** Apache-2.0
- **Copyright:** ClickHouse, Inc.
- **Usage:** Enabled subchart of the SigNoz chart, chart-managed and
  unpinned (observability/signoz/values.local.yaml:15-26).

### OpenTelemetry Collector

- **Source:** https://github.com/open-telemetry/opentelemetry-helm-charts
  and https://github.com/open-telemetry/opentelemetry-collector-contrib
- **License:** Apache-2.0
- **Copyright:** OpenTelemetry Authors / CNCF
- **Usage:** Chart 0.155.0 with image
  `otel/opentelemetry-collector-contrib:0.155.0`
  (iac/modules/observability/main.tf:30-34; variables.tf:10;
  observability/otel/collector-values.local.yaml:12-17).

### Prometheus

- **Source:** https://github.com/prometheus/prometheus and
  https://github.com/prometheus-community/helm-charts
- **License:** Apache-2.0
- **Copyright:** The Prometheus Authors / CNCF
- **Usage:** Chart 29.13.0 (iac/modules/cost/main.tf:13-17;
  variables.tf:4), values in observability/cost/prometheus-values.local.yaml.
  Alertmanager and pushgateway are disabled.

### OpenCost

- **Source:** https://github.com/opencost/opencost and
  https://github.com/opencost/opencost-helm-chart
- **License:** Apache-2.0
- **Copyright:** The OpenCost Authors / CNCF (originated at Kubecost)
- **Usage:** Chart 2.5.25 (iac/modules/cost/main.tf:26-30; variables.tf:10),
  values in observability/cost/opencost-values.local.yaml.

### Envoy Gateway

- **Source:** https://github.com/envoyproxy/gateway
- **License:** Apache-2.0
- **Copyright:** Envoy Gateway Authors / CNCF
- **Usage:** OCI chart `oci://docker.io/envoyproxy/gateway-helm` 1.3.1
  (iac/modules/networking/envoy-gateway/main.tf:5-7). Deployed images
  measured 2026-08-18: control plane docker.io/envoyproxy/gateway:v1.3.1
  (matches the chart pin) and data plane
  docker.io/envoyproxy/envoy:distroless-v1.33.0.

### Cilium (module disabled)

- **Source:** https://github.com/cilium/cilium
- **License:** Apache-2.0 (userspace); eBPF objects under bpf/ are dual
  GPL-2.0-only OR BSD-2-Clause (upstream bpf/COPYING, alongside
  LICENSE.GPL-2.0 and LICENSE.BSD-2-Clause)
- **Copyright:** Copyright Authors of Cilium (source headers); Isovalent /
  CNCF
- **Usage:** Chart 1.16.5 referenced by
  iac/modules/networking/cilium/main.tf:4-6, but the module is disabled by
  default (`enable_cilium = false`, iac/modules/networking/variables.tf:22).
  Documented fresh-cluster rebuild path only.

### Bitnami PostgreSQL chart

- **Source:** https://github.com/bitnami/charts
- **License:** Apache-2.0
- **Copyright:** Broadcom Inc (Bitnami) (2025)
- **Usage:** OCI chart `oci://registry-1.docker.io/bitnamicharts/postgresql`
  16.4.5 (iac/modules/postgres/main.tf:23-25) for production Backstage.

### PostgreSQL (bitnamilegacy container image)

- **Source:** https://www.postgresql.org/
- **License:** PostgreSQL License
- **Copyright:** PostgreSQL Global Development Group
- **Usage:** Image `bitnamilegacy/postgresql:17.6.0-debian-12-r4`
  (iac/modules/postgres/main.tf:64-69) with
  `global.security.allowInsecureImages=true` (main.tf:76) -- a deliberate
  opt-in to legacy/unsupported Bitnami images.

### distribution/distribution (OCI registry)

- **Source:** https://github.com/distribution/distribution
- **License:** Apache-2.0
- **Copyright:** distribution authors / CNCF
- **Usage:** Image `registry:2.8` for the local registry mirror
  (iac/modules/local-registry/main.tf:26).

### memcached (Gitea chart subchart)

- **Source:** https://github.com/memcached/memcached
- **License:** BSD-3-Clause
- **Copyright:** memcached authors (Danga Interactive)
- **Usage:** Enabled as a chart-managed subchart of the Gitea chart
  (scripts/gitea-values.yaml:25-26). Unpinned.

### Bitnami PostgreSQL subchart (of the Gitea chart)

- **Source:** https://github.com/bitnami/charts
- **License:** Apache-2.0 (chart) / PostgreSQL License (database)
- **Copyright:** Broadcom Inc (Bitnami) / PostgreSQL Global Development
  Group
- **Usage:** Chart-managed subchart enabled in scripts/gitea-values.yaml:20-24
  (postgresql-ha disabled). Unpinned; a transitive dependency of the Gitea
  chart with no pin in this repo.

## IaC Providers

All providers are fetched by `tofu init` from registry.opentofu.org and are
hash-pinned in `iac/.terraform.lock.hcl`, which is the authoritative record.

### OpenTofu provider hashicorp/kubernetes

- **Source:** https://github.com/hashicorp/terraform-provider-kubernetes
- **License:** MPL-2.0
- **Copyright:** HashiCorp, Inc.
- **Usage:** Constraint `~> 2.33`, locked 2.38.0 (iac/versions.tf:4;
  iac/.terraform.lock.hcl:64-65). Manages cluster resources across `iac/`.

### OpenTofu provider hashicorp/helm

- **Source:** https://github.com/hashicorp/terraform-provider-helm
- **License:** MPL-2.0
- **Copyright:** HashiCorp, Inc.
- **Usage:** Constraint `~> 2.17`, locked 2.17.0 (iac/versions.tf:5;
  iac/.terraform.lock.hcl:37-38). `helm_release` deployments.

### OpenTofu provider hashicorp/null

- **Source:** https://github.com/hashicorp/terraform-provider-null
- **License:** MPL-2.0
- **Copyright:** HashiCorp, Inc.
- **Usage:** Constraint `~> 3.2`, locked 3.3.0 (iac/versions.tf:7;
  iac/.terraform.lock.hcl:89-90).

### OpenTofu provider hashicorp/random

- **Source:** https://github.com/hashicorp/terraform-provider-random
- **License:** MPL-2.0
- **Copyright:** HashiCorp, Inc.
- **Usage:** Constraint `~> 3.6`, locked 3.9.0 (iac/versions.tf:8;
  iac/.terraform.lock.hcl:126-127).

### OpenTofu provider alekc/kubectl

- **Source:** https://github.com/alekc/terraform-provider-kubectl
- **License:** MPL-2.0
- **Copyright:** alekc (Alexander Chernov); fork of
  gavinbunney/terraform-provider-kubectl
- **Usage:** Constraint `~> 2.1`, locked 2.2.0 (iac/versions.tf:6;
  iac/.terraform.lock.hcl:4-5). `kubectl_manifest` resources (CRDs and
  custom resources).

## Go Tooling and Modules

### Go toolchain

- **Source:** https://go.dev (https://github.com/golang/go)
- **License:** BSD-3-Clause
- **Copyright:** The Go Authors (2009) / Google
- **Usage:** Builds every Go tool in this repo. `go` directives: 1.21
  (tools/score2openchoreo and the bash/brew/emoji/tofu/verify guards), 1.23
  (commit-guard), 1.26 (seed-repos/hello-m2, go.mod:3); CI installs Go 1.26
  (iac/templates/ci.yaml:18). The
  toolchain itself is unpinned by this repo. All six policy guards under
  `plugins/rr-policy-guards/tools/` are stdlib-only and have no third-party
  module dependencies; the guards themselves are first-party MIT-licensed
  project code.

### gopkg.in/yaml.v3 (go-yaml)

- **Source:** https://github.com/go-yaml/yaml
- **License:** MIT AND Apache-2.0 (dual: libyaml-ported files MIT, the rest
  Apache-2.0)
- **Copyright:** Kirill Simonov (2006-2011, MIT files); Canonical Ltd
  (2011-2019, Apache files; NOTICE 2011-2016 Canonical Ltd.)
- **Usage:** v3.0.1, build dependency of score2openchoreo
  (tools/score2openchoreo/go.mod:6; imported at main.go:10, schema.go:13).

### github.com/santhosh-tekuri/jsonschema/v5

- **Source:** https://github.com/santhosh-tekuri/jsonschema
- **License:** Apache-2.0
- **Copyright:** Santhosh Kumar Tekuri (evidenced via
  api.github.com/users/santhosh-tekuri; the v5.3.1 artifact contains no
  copyright statement: bare Apache-2.0 LICENSE, no NOTICE/COPYING/headers)
- **Usage:** v5.3.1, build dependency of score2openchoreo
  (tools/score2openchoreo/go.mod:7; imported at schema.go:12). JSON Schema
  Draft 2020-12 validation of Score files.

### OpenChoreo (algorithm replica in namespace-predictor)

- **Source:** https://github.com/openchoreo/openchoreo
- **License:** Apache-2.0
- **Copyright:** OpenChoreo project authors
- **Usage:** `tools/namespace-predictor/main.go` replicates
  `GenerateK8sNameWithLengthLimit` byte-for-byte (main.go:18,32,43-45,83-88).
  A source-level mirror, not a module dependency; the mirrored upstream
  revision is not recorded (see UNVERIFIED items).

### Semgrep

- **Source:** https://github.com/semgrep/semgrep
- **License:** LGPL-2.1
- **Copyright:** Semgrep Inc. (upstream COPYRIGHT file: Copyright (C)
  2019, 2020, 2021, 2022, 2023, 2024 Semgrep Inc.)
- **Usage:** Invoked by rr-verify-guard at push with `p/security-audit`
  (plugins/rr-policy-guards/tools/verify-guard/exec.go:96-97). Resolved
  from PATH, unpinned.

### Gitleaks

- **Source:** https://github.com/gitleaks/gitleaks
- **License:** MIT
- **Copyright:** Zachary Rice (2019)
- **Usage:** Invoked by rr-verify-guard at push
  (plugins/rr-policy-guards/tools/verify-guard/exec.go:110-111). Resolved
  from PATH, unpinned.

### govulncheck (golang.org/x/vuln)

- **Source:** https://github.com/golang/vuln
- **License:** BSD-3-Clause
- **Copyright:** The Go Authors
- **Usage:** Invoked by rr-verify-guard for go.mod roots
  (plugins/rr-policy-guards/tools/verify-guard/exec.go:168-169). Resolved
  from PATH, unpinned.

### Seed application Go modules (seed-repos/hello-m2)

The demo app `hello-m2` has the following module dependencies, pinned in
`seed-repos/hello-m2/go.mod` with hashes in `seed-repos/hello-m2/go.sum`
(the authoritative closure). The golang.org/x module licenses were
verified per-repo from the module cache at the exact pinned versions
(2026-08-18).

| Module | Version | License | Copyright |
|---|---|---|---|
| go.opentelemetry.io/otel (+ trace, sdk, metric, otlptrace exporters) | v1.44.0 | Apache-2.0 | The OpenTelemetry Authors / CNCF (LICENSE embeds Go BSD-3-Clause for derived code) |
| go.opentelemetry.io/auto/sdk | v1.2.1 | Apache-2.0 | The OpenTelemetry Authors / CNCF |
| go.opentelemetry.io/proto/otlp | v1.10.0 | Apache-2.0 | The OpenTelemetry Authors / CNCF |
| github.com/cenkalti/backoff/v5 (indirect) | v5.0.3 | MIT | Cenk Alti (2014) |
| github.com/cespare/xxhash/v2 (indirect) | v2.3.0 | MIT | Caleb Spare (2016) |
| github.com/go-logr/logr (indirect) | v1.4.3 | Apache-2.0 | go-logr contributors |
| github.com/go-logr/stdr (indirect) | v1.2.2 | Apache-2.0 | go-logr contributors |
| github.com/google/uuid (indirect) | v1.6.0 | BSD-3-Clause | Google Inc. (2009, 2014) |
| github.com/grpc-ecosystem/grpc-gateway/v2 (indirect) | v2.29.0 | BSD-3-Clause | Gengo, Inc. (2015) |
| golang.org/x/net (indirect) | v0.55.0 | BSD-3-Clause | The Go Authors ("Copyright 2009 The Go Authors.", per-repo LICENSE at the pinned version; PATENTS file present) |
| golang.org/x/sys (indirect) | v0.45.0 | BSD-3-Clause | The Go Authors ("Copyright 2009 The Go Authors.", per-repo LICENSE at the pinned version; PATENTS file present) |
| golang.org/x/text (indirect) | v0.37.0 | BSD-3-Clause | The Go Authors ("Copyright 2009 The Go Authors.", per-repo LICENSE at the pinned version; PATENTS file present) |
| google.golang.org/genproto (googleapis/api + googleapis/rpc, indirect) | v0.0.0-20260526163538-3dc84a4a5aaa | Apache-2.0 | Google LLC |
| google.golang.org/grpc (indirect) | v1.83.0 | Apache-2.0 | gRPC authors / Google LLC (CNCF) |
| google.golang.org/protobuf (indirect) | v1.36.11 | BSD-3-Clause | The Go Authors (2018) / Google LLC |

## Backstage and Node.js

The `backstage/` directory was scaffolded using `@backstage/create-app`
(`npx @latest`, docs/specs/m1-substrate/technical-specification.md:693) and
tracks the Backstage release line 1.49.1 (backstage/backstage.json:2). All
files originating from the scaffold are Apache-2.0 licensed; custom code on
top of the scaffold (`packages/backend/src/modules/giteaAuth.ts`,
`packages/app/src/modules/*`, the app-config files) is original work by the
project authors.

All 113 installed `@backstage/*` packages were verified (2026-08-18) to
ship no LICENSE/COPYING file inside their tarballs; the authoritative
license evidence is the package.json `license` field (Apache-2.0) plus
the upstream monorepo LICENSE ("Copyright 2020 The Backstage Authors";
the project was originally Spotify AB). The tables below list direct
dependencies and the `resolutions` override pins; `backstage/yarn.lock` is
the authoritative full transitive closure.

### Framework, scaffold, and package manager

| Component | Version/pin | License | Copyright |
|---|---|---|---|
| @backstage/create-app (scaffold tool) | unpinned (`@latest` at scaffold time; produced release line 1.49.1) | Apache-2.0 | The Backstage Authors (CNCF; originally Spotify AB) |
| Backstage framework release line | 1.49.1 (backstage/backstage.json:2) | Apache-2.0 | The Backstage Authors |
| Yarn (vendored release) | 4.4.1 (`packageManager` field; backstage/.yarn/releases/yarn-4.4.1.cjs) | BSD-2-Clause | Yarn Contributors (2016-present) |

### Root workspace devDependencies (backstage/package.json)

Build, test, and lint tooling. Declared range, with installed version in
parentheses.

| Package | Range (installed) | License | Copyright |
|---|---|---|---|
| @backstage/cli | ^0.36.0 (0.36.0) | Apache-2.0 | The Backstage Authors |
| @backstage/cli-defaults | ^0.1.0 (0.1.0) | Apache-2.0 | The Backstage Authors |
| @backstage/e2e-test-utils | ^0.1.2 (0.1.2) | Apache-2.0 | The Backstage Authors |
| @jest/environment-jsdom-abstract | ^30.0.0 (30.3.0) | MIT | Meta Platforms, Inc. and affiliates |
| @playwright/test | ^1.32.3 (1.59.1) | Apache-2.0 | Microsoft Corporation |
| @types/jest | ^30.0.0 (30.0.0) | MIT | DefinitelyTyped contributors |
| jest | ^30.2.0 (30.3.0) | MIT | Meta Platforms, Inc. and affiliates |
| jsdom | ^27.1.0 (27.4.0) | MIT | Elijah Insua, Domenic Denicola, et al. |
| node-gyp | ^10.0.0 (10.3.1) | MIT | Nathan Rajlich / Node.js contributors |
| prettier | ^2.3.2 (2.8.8) | MIT | James Long and contributors |
| typescript | ~5.8.0 (5.8.3) | Apache-2.0 | Microsoft Corp. |

### App workspace dependencies (backstage/packages/app/package.json)

Runtime dependencies of the frontend, plus its test tooling.

| Package | Range (installed) | License | Copyright |
|---|---|---|---|
| @backstage/core-components | ^0.18.8 (0.18.8) | Apache-2.0 | The Backstage Authors |
| @backstage/core-plugin-api | ^1.12.4 (1.12.4) | Apache-2.0 | The Backstage Authors |
| @backstage/frontend-defaults | ^0.5.0 (0.5.0) | Apache-2.0 | The Backstage Authors |
| @backstage/frontend-plugin-api | ^0.15.1 (0.15.1) | Apache-2.0 | The Backstage Authors |
| @backstage/integration-react | ^1.2.16 (1.2.16) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-api-docs | ^0.13.5 (0.13.5) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-app-react | ^0.2.1 (0.2.1) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-app-visualizer | ^0.2.1 (0.2.1) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog | ^2.0.1 (2.0.1) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog-graph | ^0.6.0 (0.6.0) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog-import | ^0.13.11 (0.13.11) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-kubernetes | ^0.12.17 (0.12.17) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-notifications | ^0.5.15 (0.5.15) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-org | ^0.7.0 (0.7.0) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-scaffolder | ^1.36.1 (1.36.1) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-search | ^1.7.0 (1.7.0) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-signals | ^0.0.29 (0.0.29) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-techdocs | ^1.17.2 (1.17.2) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-techdocs-module-addons-contrib | ^1.1.34 (1.1.34) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-user-settings | ^0.9.1 (0.9.1) | Apache-2.0 | The Backstage Authors |
| @backstage/ui | ^0.13.1 (0.13.2) | Apache-2.0 | The Backstage Authors |
| @material-ui/core | ^4.12.2 (4.12.4) | MIT | Call-Em-All (Material-UI Team) |
| @material-ui/icons | ^4.9.1 (4.11.3) | MIT | Call-Em-All (Material-UI Team) |
| react | ^18.0.2 (18.3.1) | MIT | Facebook, Inc. and its affiliates (Meta) |
| react-dom | ^18.0.2 (18.3.1) | MIT | Facebook, Inc. and its affiliates (Meta) |
| react-router | ^6.30.4, resolution pin 6.30.4 (6.30.4) | MIT | Remix Software |
| react-router-dom | ^6.30.4, resolution pin 6.30.4 (6.30.4) | MIT | Remix Software |
| @backstage/frontend-test-utils (dev) | ^0.5.1 (0.5.1) | Apache-2.0 | The Backstage Authors |
| @testing-library/dom (dev) | ^9.0.0 (9.3.4) | MIT | Kent C. Dodds and contributors |
| @testing-library/jest-dom (dev) | ^6.0.0 (6.9.1) | MIT | Testing Library contributors |
| @testing-library/react (dev) | ^14.0.0 (14.3.1) | MIT | Kent C. Dodds and contributors |
| @testing-library/user-event (dev) | ^14.0.0 (14.6.1) | MIT | Testing Library contributors |
| @types/react-dom (dev) | `*`, resolution pin ^18 (18.3.7) | MIT | DefinitelyTyped contributors |
| cross-env (dev) | ^7.0.0 (7.0.3) | MIT | Kent C. Dodds |

### Backend workspace dependencies (backstage/packages/backend/package.json)

Runtime dependencies of the Backstage backend.

| Package | Range (installed) | License | Copyright |
|---|---|---|---|
| @backstage/backend-defaults | ^0.16.0 (0.16.0) | Apache-2.0 | The Backstage Authors |
| @backstage/config | ^1.3.6 (1.3.6) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-app-backend | ^0.5.12 (0.5.12) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-auth-backend | ^0.29.2 (0.29.2) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-auth-backend-module-github-provider | ^0.5.1 (0.5.1) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-auth-backend-module-guest-provider | ^0.2.17 (0.2.17) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-auth-node | ^0.6.14 (0.6.14) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog-backend | ^3.5.0 (3.5.0) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog-backend-module-gitea | ^0.1.10 (0.1.10) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog-backend-module-logs | ^0.1.20 (0.1.20) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-catalog-backend-module-scaffolder-entity-model | ^0.2.18 (0.2.18) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-kubernetes-backend | ^0.21.2 (0.21.2) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-notifications-backend | ^0.6.3 (0.6.3) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-permission-backend | ^0.7.10 (0.7.10) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-permission-backend-module-allow-all-policy | ^0.2.17 (0.2.17) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-permission-common | ^0.9.7 (0.9.7) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-permission-node | ^0.10.11 (0.10.11) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-proxy-backend | ^0.6.11 (0.6.11) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-scaffolder-backend | ^3.2.0 (3.3.0) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-scaffolder-backend-module-github | ^0.9.7 (0.9.7) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-scaffolder-backend-module-notifications | ^0.1.20 (0.1.20) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-search-backend | ^2.1.0 (2.1.0) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-search-backend-module-catalog | ^0.3.13 (0.3.13) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-search-backend-module-pg | ^0.5.53 (0.5.53) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-search-backend-module-techdocs | ^0.4.12 (0.4.12) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-search-backend-node | ^1.4.2 (1.4.2) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-signals-backend | ^0.3.13 (0.3.13) | Apache-2.0 | The Backstage Authors |
| @backstage/plugin-techdocs-backend | ^2.1.6 (2.1.6) | Apache-2.0 | The Backstage Authors |
| better-sqlite3 | ^12.0.0 (12.8.0) | MIT (bundles SQLite, public domain) | Joshua Wise |
| pg (node-postgres) | ^8.11.3 (8.20.0) | MIT | Brian Carlson |

### Resolution override pins (backstage/package.json, resolutions block)

The `resolutions` block is a security-override layer applying repo-wide to
all workspaces. The packages below appear only as resolution pins (they are
transitive-only dependencies); `react-router`, `react-router-dom`, and
`@types/react-dom` also carry resolution pins and are listed in the app
table above. Copyright holders are taken from the installed tarball
LICENSE headers (verified 2026-08-18).

| Package | Pin (installed) | License | Copyright |
|---|---|---|---|
| @types/react | ^18 (18.3.28) | MIT | DefinitelyTyped contributors |
| @grpc/grpc-js | ^1.14.4 (1.14.4) | Apache-2.0 | Google Inc. |
| @babel/core | ^7.29.1 (7.29.7) | MIT | The Babel Team |
| @protobufjs/utf8 | ^1.1.1 (1.1.2) | BSD-3-Clause | Daniel Wirtz (protobufjs) |
| ws | ^8.21.0 (8.21.0) | MIT | Einar Otto Stangvik |
| axios | ^1.13.5 (1.19.0) | MIT | Matt Zabriskie and collaborators |
| undici | ^7.28.0 (7.28.0) | MIT | Node.js contributors |
| lodash | 4.18.1 | MIT | John-David Dalton / Lodash contributors |
| protobufjs | 7.6.5 | BSD-3-Clause | Daniel Wirtz |
| shell-quote | 1.10.0 | MIT | James Halliday |
| websocket-driver | 0.7.5 | Apache-2.0 | James Coglan |
| form-data | 4.0.6 | MIT | Felix Geisendoerfer and contributors |
| minimatch | 9.0.9 (single installed copy; the descriptor pins at package.json:67-70 are dead config shadowed by the bare pin at :93) | ISC | Isaac Z. Schlueter |
| immutable | 4.3.9 | MIT | Lee Byron and contributors |
| tar | 7.5.22 | BlueOak-1.0.0 | Isaac Z. Schlueter |
| linkify-it | 6.1.0 | MIT | Vitaly Puzrin (2015) |
| fast-xml-builder | 1.3.0 | MIT | Natural Intelligence (2026); author Amit Gupta |
| brace-expansion | 5.0.8 | MIT | Julian Gruber; TypeScript port Isaac Z. Schlueter |
| fast-uri | 4.1.1 | BSD-3-Clause | Vincent Le Goff |
| postcss | 8.5.25 | MIT | Andrey Sitnik |
| basic-ftp | 6.0.2 | MIT | Patrick Juchli |
| js-cookie | 3.0.8 | MIT | Klaus Hartl |
| multer | 2.2.0 | MIT | Hage Yaapa, Jaret Pfluger, et al. |
| http-proxy-middleware | 2.0.10 | MIT | Steven Chim |
| dompurify | 3.4.12 | (MPL-2.0 OR Apache-2.0) | Mario Heiderich, Cure53 |
| follow-redirects | 1.16.0 | MIT | Ruben Verborgh and contributors |
| qs | 6.15.3 | BSD-3-Clause | Jordan Harband |
| uuid | 11.1.1 | MIT | Robert Kieffer and other contributors (2010-2020) |
| svgo | 3.3.4 | MIT | Kir Belevich and contributors |
| koa | 2.16.4 | MIT | Koa contributors ((c) 2019) |
| adm-zip | 0.6.0 | MIT | Nasca Iacob (cthackers) |
| js-yaml | 4.3.0 | MIT | Vladimir Zapparov and contributors |
| webpack-dev-server | 5.2.6 | MIT | Tobias Koppers (webpack contributors) |
| prismjs | 1.30.0 | MIT | Lea Verou and contributors |
| body-parser | 1.20.6 | MIT | Douglas Christopher Wilson, Jonathan Ong, et al. |
| fast-xml-parser | 5.7.0 | MIT | Amit Gupta (NaturalIntelligence) |
| markdown-it | 14.2.0 | MIT | Vitaly Puzrin, Alex Kocharin ((c) 2014) |
| launch-editor | 2.14.1 | MIT | Evan You |
| ip-address | 10.1.1 | MIT | Beau Gunderson |
| @octokit/request | 8.4.1 | MIT | Gregor Martynus (Octokit) |
| @octokit/request-error | 5.1.1 | MIT | Octokit contributors (LICENSE: Copyright (c) 2019 Octokit contributors) |
| @octokit/plugin-paginate-rest | 11.4.1 | MIT | Octokit contributors (LICENSE: Copyright (c) 2019 Octokit contributors) |
| @tootallnate/once | 2.0.1 | MIT | Nathan Rajlich |
| esbuild | 0.28.1 | MIT | Evan Wallace ((c) 2020) |

Runtime-fetched third-party artifacts adjacent to this scope (not in yarn):
Playwright downloads browser binaries at `yarn test:e2e` time (per the
browsers.json of installed @playwright/test 1.59.1): Chromium
147.0.7727.15 r1217 and chromium-headless-shell r1217 (BSD-3-Clause,
Copyright 2015 The Chromium Authors), Firefox 148.0.2 r1511 (MPL-2.0
primary; multi-license aggregate, ~80 sub-licenses including LGPL-2.1
components), WebKit 26.4 r2272 (portions LGPL and BSD-2-Clause, Apple),
and ffmpeg r1011 (binary license not separately fetched); each browser is
a multi-license aggregate. TechDocs with `generator.runIn: docker` pulls
the `spotify/techdocs` image at doc-build time (app-config.yaml:82-83) --
default tag v1.2.8, built from github.com/backstage/techdocs-container,
Apache-2.0, Copyright 2020 Spotify AB; per-component licenses inside the
image are not enumerated. `better-sqlite3` bundles SQLite (public domain)
compiled via node-gyp. The Node.js runtime itself is required at `22 || 24`
(backstage/package.json:6); measured 2026-08-18: v22.23.2 (host default,
~/.local/bin/node) and v24.19.0 (/opt/homebrew/opt/node@24/bin/node, the
AGENTS.md path), licensed MIT with bundled third-party libraries under
their own licenses.

## CI/CD Actions and Images

### actions/checkout

- **Source:** https://github.com/actions/checkout
- **License:** MIT
- **Copyright:** GitHub, Inc. and contributors (2018)
- **Usage:** `@v4` (mutable major tag) in iac/templates/ci.yaml:13 and
  seed-repos/hello-m2/.gitea/workflows/ci.yaml:13.

### actions/setup-go

- **Source:** https://github.com/actions/setup-go
- **License:** MIT
- **Copyright:** GitHub, Inc. and contributors (2018)
- **Usage:** `@v5` (mutable major tag) in iac/templates/ci.yaml:16-18 and
  seed-repos/hello-m2/.gitea/workflows/ci.yaml:16-18; installs Go 1.26
  (Go toolchain covered under Go Tooling and Modules).

### opentofu/setup-opentofu

- **Source:** https://github.com/opentofu/setup-opentofu
- **License:** MPL-2.0
- **Copyright:** OpenTofu Authors; the LICENSE records Copyright (c) 2020
  HashiCorp, Inc., inherited from the forked hashicorp/setup-terraform
- **Usage:** `@v1` (mutable major tag) in iac/templates/ci.yaml:31 and
  seed-repos/hello-m2/.gitea/workflows/ci.yaml:31; installs OpenTofu 1.9.0.

### golang (Docker Official Image)

- **Source:** https://hub.docker.com/_/golang
- **License:** BSD-3-Clause (Go toolchain)
- **Copyright:** The Go Authors / Google LLC
- **Usage:** Tag `1.26-alpine` (no digest), build stage of
  seed-repos/hello-m2/Dockerfile:1. Its Alpine layer is a different, newer
  base than alpine:3.20: as of 2026-08-18 golang:1.26-alpine is Alpine
  3.24.1-based (16 packages, including apk-tools 3.0.6/libapk, busybox
  1.37.0, musl 1.2.6-r2, libcrypto3/libssl3 3.5.7 Apache-2.0,
  ca-certificates-bundle 20260611 MPL-2.0 AND MIT, zlib Zlib).

### alpine (Docker Official Image)

- **Source:** https://hub.docker.com/_/alpine
- **License:** Multi-license aggregate. alpine:3.20 is currently Alpine
  3.20.10 (image created 2026-04-16), 14 packages enumerated 2026-08-18
  (Docker Hub registry API, sha256-verified blobs): alpine-baselayout
  3.6.5-r0 GPL-2.0-only; alpine-baselayout-data 3.6.5-r0 GPL-2.0-only;
  alpine-keys 2.4-r1 MIT; apk-tools 2.14.4-r1 GPL-2.0-only; busybox
  1.36.1-r31 GPL-2.0-only; busybox-binsh 1.36.1-r31 GPL-2.0-only;
  ca-certificates-bundle 20260413-r0 MPL-2.0 AND MIT; libcrypto3 3.3.7-r0
  Apache-2.0; libssl3 3.3.7-r0 Apache-2.0; musl 1.2.5-r3 MIT; musl-utils
  1.2.5-r3 MIT AND BSD-2-Clause AND GPL-2.0-or-later; scanelf 1.3.7-r2
  GPL-2.0-only; ssl_client 1.36.1-r31 GPL-2.0-only; zlib 1.3.2-r0 Zlib.
- **Copyright:** Alpine Linux development team
- **Usage:** Tag `3.20` (no digest), runtime stage of
  seed-repos/hello-m2/Dockerfile:6.

### node (Docker Official Image)

- **Source:** https://hub.docker.com/_/node and
  https://github.com/nodejs/docker-node
- **License:** MIT for the image build files (docker-node LICENSE,
  Copyright (c) 2015 Joyent, Inc. / Node.js contributors) and for the
  Node.js runtime itself; the runtime bundles externally maintained
  libraries under their own licenses (including Apache-2.0, Unicode-3.0,
  and BSD variants, per the nodejs/node LICENSE). The Debian trixie-slim
  base layer is debootstrap minbase built from Debian main only, so every
  package is DFSG-free per the Debian Social Contract; per-package license
  texts are extractable at /usr/share/doc/<pkg>/copyright per Debian
  Policy 12.5, but the exact per-package list of the deployed layer was
  not enumerated (PARTIAL, see UNVERIFIED items). Current trixie-slim is
  13.6-slim (snapshot trixie-20260803). No single license covers the whole
  image stack.
- **Copyright:** Joyent, Inc. and Node.js contributors (image and
  runtime); Debian Project (base layer, not enumerated)
- **Usage:** Tag `24-trixie-slim` (no digest), base image of the
  production Backstage backend image
  (backstage/packages/backend/Dockerfile:15), built via `yarn build-image`
  (backstage/packages/backend/package.json:16). Pulled at docker build
  time.

CI also uses OpenTofu, Infracost, Docker, jq, and git; these are covered in
the groups above or are runner-image-provided binaries with no repo-level
pin. `.github/workflows/sync-from-gitea.yml` uses no third-party GitHub
Actions at all -- only `run:` steps with git against gitea.com and
github.com, which are external services, not incorporated software. The
base-image tags in this group are floating (alpine:3.20,
golang:1.26-alpine, node:24-trixie-slim); the package enumerations above
are a snapshot of 2026-08-18 and drift as tags move -- digest pinning
would close this.

## Schemas and Specifications

### Score specification and JSON schema (score.dev/v1b1)

- **Source:** https://github.com/score-spec/spec
- **License:** Apache-2.0
- **Copyright:** The Score Authors (2022); CNCF sandbox (originated at
  Humanitec)
- **Usage:** Workload specification used by
  seed-repos/hello-m2/score.yaml:1. The JSON schema is vendored at
  tools/score2openchoreo/assets/score.schema.json (SHA256
  `633c4394dfc03977c86932f5d10c77e2c356a38b63f162c281104fade5c9863c`,
  16701 bytes, pinned 2026-04-23 per assets/SCHEMA_PROVENANCE.md) and
  embedded via `//go:embed` (schema.go:19-20); a pin test fails the build
  on drift. The upstream revision is commit
  3ecb17d430c2bbf46d2dfc161fabc7d432d6d1f5 (2026-04-17, score-spec/spec),
  upstream file `score-v1b1.json` at the repository root, byte-identical
  to the vendored pin (match verified 2026-08-18).

## Documentation Tooling

### MkDocs

- **Source:** https://github.com/mkdocs/mkdocs
- **License:** BSD-2-Clause
- **Copyright:** Tom Christie (2014-present) and contributors
- **Usage:** Docs engine behind mkdocs.yml:1-2, run by Backstage TechDocs
  at docs build time. Resolved at install time, unpinned; nothing
  docs-related is vendored in this repo.

### mkdocs-techdocs-core

- **Source:** https://github.com/backstage/mkdocs-techdocs-core
- **License:** Apache-2.0
- **Copyright:** The Backstage Authors (2020-21)
- **Usage:** MkDocs `techdocs-core` plugin enabled in mkdocs.yml:3-4.
  Resolved at docs build time, unpinned.

## Items not fully verified (summary)

Resolution passes on 2026-08-18 resolved 22 of the 25 earlier gaps with
hard evidence (15 in a first pass, then the 7 cluster-blocked rows in a
second pass with the cluster up; their values are folded into the entries
above). Three gaps remain; full detail is in the UNVERIFIED subsection of
`provenance/PROVENANCE.md` (original U-numbers retained):

- U7: the exact OpenChoreo commit mirrored by namespace-predictor is not
  recorded (holder and license confirmed from upstream SPDX headers).
- U19: per-component licenses inside the spotify/techdocs image (tag
  v1.2.8, source, and license are verified).
- U25: exact per-package license list of the node:24-trixie-slim Debian
  base layer (DFSG-free main-only composition and extraction path
  documented).

Residual caveat carried forward from the closed U17: the
catthehacker/ubuntu:act-* CI job image exists only while a CI job is in
flight (zero matches in a full-cluster scan 2026-08-18); its tag is
observable only during a CI run. The runner itself is measured:
act_runner 0.3.1 with dind docker.io/docker:29.4.0-dind.
