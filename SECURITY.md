# Security Policy

This document is the security policy for the **Developer Portal** repository
(`developer-portal`): the self-hosted Internal Developer Platform (IDP) for
Trade Momentum LLC.

It applies to source under this repository, including Backstage, lifecycle
scripts, policy guards, Score tooling, seed applications, and IaC modules.

## Supported surfaces

| Surface | Support |
|---------|---------|
| Current `main` branch | Security fixes accepted |
| Tagged releases (when published) | Security fixes for the latest tag |
| Historical commits / forks | Best-effort only |

Local k3d / laptop demos are not internet-facing production. Production
hardening still requires the publication and deployment controls described
below.

## Reporting a vulnerability

**Do not open a public GitHub Issue or pull request that discloses an
exploitable vulnerability.**

Report privately instead:

1. **Preferred:** GitHub Security Advisory for
   [trademomentum-llc/developer-portal](https://github.com/trademomentum-llc/developer-portal/security/advisories/new)
   (private report).
2. **Alternate:** email `security@trademomentumllc.com` with subject
   `SECURITY developer-portal` (if that mailbox is not yet provisioned, use the
   private GitHub advisory channel).

Include:

- Affected component and path (for example `backstage/`, `plugins/rr-policy-guards/`)
- Description and impact
- Reproduction steps or proof of concept
- Whether the issue is already public elsewhere
- Your preferred credit line (optional)

You should receive an acknowledgement within **5 business days**. We aim to
provide a remediation plan or status update within **14 days** for High and
Critical issues.

## Disclosure

We follow coordinated disclosure:

- Private report until a fix is available or the risk is accepted in writing
- Public advisory and changelog entry when a fix ships
- CVE assignment when appropriate for independently published packages

Do not discuss unfixed High/Critical issues in public channels.

## What we fix first

Priority order for inbound reports and scanner findings:

1. Authentication / authorization bypass (Backstage, Gitea OAuth, production
   config)
2. Remote code execution, path traversal, secret exposure
3. High/Critical dependency advisories in **production** dependency trees
4. Denial of service that is reachable from exposed network surfaces
5. Development-only toolchain issues (webpack-dev-server, test helpers)

Material UI v4 deprecation notices and similar non-CVE deprecations are tracked
as upgrade debt, not as security emergencies.

## Controls in this repository

### Publication gate (`rr-verify-guard`)

Every clean push MUST pass:

| Control | Tool |
|---------|------|
| Static security rules | Semgrep `p/security-audit` |
| Secret scanning | Gitleaks |
| Node SCA | `yarn npm audit --all --recursive --severity high --no-deprecations` (or `npm audit --audit-level=high`) |
| Go SCA | `govulncheck ./...` per `go.mod` root |
| Quality | Toolchain lint / test / build as detected |

There is no environment-variable or commit-message bypass for these guards.
Missing scanners when marker files exist is a blocking defect.

### Local git hooks

`plugins/rr-policy-guards/scripts/install-git-hooks.sh` installs `pre-commit`
and `commit-msg` hooks that run `rr-commit-guard` on staged content.

### Dependency governance

- Prefer fixing High/Critical findings via direct upgrades or Yarn
  `resolutions` with documented pins
- Re-run `yarn npm audit --all --recursive --severity high --no-deprecations`
  under `backstage/` after dependency changes
- Re-run `govulncheck ./...` under each Go module after module changes

### Secrets

- Never commit credentials, tokens, or private keys
- Local secrets live outside the repo (for example
  `~/.rational-reserve/` and gitignored `app-config.local.yaml`)
- Scripts MUST NOT place secrets in process arguments; prefer files with
  mode `0600` or environment injection by trusted lifecycle scripts

### Temporary files

Diagnostic artifacts from tests and smokes MUST use private temporary
directories (`fs.mkdtempSync` / `tempfile.mkdtemp` with mode `0700`), not
predictable world-writable paths such as fixed names under `/tmp`.

## Source of truth and mirrors

| Role | Location |
|------|----------|
| Primary source | Gitea (`gitea.com/trademomentum.net/developer-portal` and local org mirror) |
| GitHub mirror | `github.com/trademomentum-llc/developer-portal` (Vercel Git + Dependabot/CodeQL) |
| Vercel project | `trademomentumllcs-projects/developer-portal` (`devplatform.link`) |

Security fixes MUST land on the Gitea source of truth and propagate to GitHub
(via mirror sync) so automated GitHub scanning reflects the fixed tree.

## Related governance

Portfolio-level mandatory controls:

- `~/Projects/Sovereign/Structure/POLICIES.md`
- `~/Projects/Sovereign/Structure/Requirements.md` (RQ-006, RQ-013, RQ-016)
- `~/Projects/Sovereign/Structure/Tech-Spec.md` (publication security scanners)
- `~/Projects/Sovereign/procedural-mandates/02_SECURITY_AND_CRYPTO.md`
- `~/Projects/Sovereign/procedural-mandates/06_VERIFICATION_AND_RELEASE.md`

## Safe harbor

Good-faith research conducted without privacy violation, service disruption, or
data destruction is welcome under this policy when reported privately as above.
