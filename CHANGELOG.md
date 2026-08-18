# Changelog

All notable changes to this repository are tracked here.

## Unreleased

### Added

- Added `docs/specs/m3-observability/` kickoff requirements, design, and technical specs for OpenTelemetry, SigNoz, and post-deploy Infracost visibility.
- Added a Backstage `Component` catalog entity for `developer-portal`.
- Added Backstage catalog smoke coverage for the local `developer-portal` and `hello-m2` component links.

### Changed

- Updated Backstage app title to `Developer Portal`.
- Wired Backstage catalog locations to local `catalog-info.yaml` and `seed-repos/hello-m2/catalog-info.yaml`.
- Updated `scripts/start-backstage.sh` to default to `127.0.0.1:3001` with backend `127.0.0.1:7008` and to prefer Homebrew `node@24` when present.
- Updated project status docs to mark M2 locally validated and M3 active at the specification/preflight stage.

### Known Issues

- `git push gitea-com main` reaches `gitea.com` but fails authentication until the cloud Gitea credential/PAT is refreshed. **Stale as of 2026-08-18:** `origin` and `gitea-com` now point at `https://gitea.com/trademomentum.net/developer-portal.git` (gitea.com SaaS), `github` is the trademomentum-llc mirror, local Gitea `localhost:3333` is not a configured remote, and HEAD `67a17f9` is fetch-verified in sync with `origin` (`git ls-remote`); push itself remains UNVERIFIED.
- `yarn npm audit --all --recursive` reports existing critical/high Backstage dependency advisories; no dependency files changed in the Backstage catalog work.
