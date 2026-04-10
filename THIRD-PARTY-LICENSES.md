# Third-Party Licenses

This project incorporates or depends on the following third-party software.
Each retains its original license and copyright.

---

## Backstage

- **Source:** https://github.com/backstage/backstage
- **License:** Apache License 2.0
- **Copyright:** Spotify AB and contributors
- **Usage:** The `backstage/` directory was scaffolded using `@backstage/create-app`.
  All files under `backstage/` originating from the scaffold are Apache 2.0 licensed.
  Custom configuration (app-config.yaml patches, backend index.ts modifications)
  is original work by the project authors.

## Backstage Gitea Plugin

- **Package:** `@backstage/plugin-catalog-backend-module-gitea`
- **License:** Apache License 2.0
- **Copyright:** Spotify AB and contributors
- **Usage:** Installed as a dependency for Gitea catalog integration.

## Gitea Helm Chart

- **Source:** https://gitea.com/gitea/helm-chart
- **License:** MIT License
- **Copyright:** The Gitea Authors
- **Usage:** Referenced via `gitea-charts/gitea` helm repository.
  The chart is not vendored; it is fetched at install time by helm.
  The values file (`scripts/gitea-values.yaml`) is original work.

## k3s (via k3d)

- **Source:** https://github.com/k3s-io/k3s
- **License:** Apache License 2.0
- **Copyright:** Rancher Labs / SUSE
- **Usage:** k3d pulls k3s container images at cluster creation time.
  No k3s source is included in this repository.

## Go Standard Library

- **License:** BSD 3-Clause
- **Copyright:** The Go Authors
- **Usage:** All Go binaries in `plugins/rr-policy-guards/tools/` are
  built with stdlib only (no third-party Go modules).
