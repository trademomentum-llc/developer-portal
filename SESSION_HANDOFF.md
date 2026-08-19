# SESSION HANDOFF

> Read this file FIRST in the next session. It tells you where we stopped,
> what is now committed that was not before, what is still outstanding, and
> exactly what to do first.

**Last updated:** 2026-08-18
**Reason for handoff:** Attribution/provenance package landed (see section 0). M3 Production Multi-Angle Visibility and M4 cost visibility remain live; `./scripts/smoke-all.sh` previously reported `ALL SMOKE SUITES PASSED (M2, M3, M4)`.

---

## 0. 2026-08-18 addendum -- provenance package (this session)

Landed under a goal-mode directive (five-plane collaborative portal +
record immutability + mandatory attribution practice). The attribution
triple now exists and passed an adversarial critic review:

- `THIRD-PARTY-LICENSES.md` -- expanded from 5 entries to the full
  third-party inventory in 8 groups.
- `provenance/PROVENANCE.md` -- 189 evidenced entries (version/pin,
  upstream URL, SPDX license, copyright holder, usage mode, repo evidence
  path) plus 25 openly recorded UNVERIFIED gaps (U1-U25).
- `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md` -- certificate
  `PRC-developer-portal-2026-08-18-r2`, self-attested, SHA-256 integrity
  digests of the two files above embedded; supersede/revocation rule
  recorded.
- `AGENTS.md` Conventions -- the attribution triple is now recorded as
  standing portfolio practice.

State notes:
- All of the above is UNCOMMITTED (working tree only), along with a
  pre-existing uncommitted edit to
  `plugins/rr-policy-guards/tools/verify-guard/main.go` and untracked
  `.claude/` that predate this session.
- Git log moved since the 2026-06-30 handoff below: HEAD is now
  `67a17f9 fix(security): patch Dependabot moderate/high dependency
  alerts` (security sync commits `3c8cd30`, `24c33fb` precede it).
  Section 2's git-state block reflects 2026-06-30.
- 2026-08-18 (same session, second slice): five-plane portal roadmap
  requirements doc landed at
  `docs/specs/2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md`
  -- evidence-backed current state for all five planes (observation,
  control, orchestration, security, engagement), 53 gap registers,
  5x5 traversal matrix with 12 breakdowns, 40 FRs + 10 NFRs, 45 PROPOSED
  candidate components (nothing decided outside the locked stack), 4
  RECOMMENDED phases, 31 open questions (OQ-19 security vertical slice
  flagged as an explicit user decision). Critic-approved after a
  correction pass. Uncommitted like the rest.
- 2026-08-18 (same session, third slice -- user-directed): resolved the
  25 UNVERIFIED gap rows (U1-U25) in `provenance/PROVENANCE.md` via five
  evidence-forced resolver bundles. Outcome: 15 fully resolved and
  removed, 4 narrowed (U7, U11, U19, U25), 6 blocked by an unreachable
  cluster (U8, U9, U12, U13, U16, U17 -- Colima stopped; re-run when
  the cluster is up). Corrections applied: uuid holder (Robert Kieffer
  and other contributors), alpine:3.20 full 14-package license
  enumeration, golang:1.26-alpine now Alpine 3.24.1-based (stale claim
  fixed), Semgrep Inc. (no comma), Cilium eBPF SPDX wording, Score
  schema upstream commit pinned (3ecb17d430c2..., byte-identical).
  Certificate re-issued as `PRC-developer-portal-2026-08-18-r3`
  (digests: TPL 1c5ee689..., PROVENANCE 66173422...). Critic round 3:
  APPROVE, zero defects. Cross-document anomalies surfaced to the user:
  sibling openchoreo checkout is ABSENT on this machine (contradicts
  AGENTS.md/PROJECT_SUMMARY.md); .claude-plugin/marketplace.json is
  stale (says four guards + bypass vars; reality is six, no bypass);
  minimatch descriptor pins at backstage/package.json:67-70 are dead
  config shadowed by the bare pin at :93; tools/namespace-predictor/
  main.go:18 comment cites a wrong upstream path.
- 2026-08-18 (same session, fourth slice): Record Immutability spec
  pair landed and critic-approved (2 rounds, 8 minor corrections, one
  critic error adjudicated against the critic on evidence):
  - `docs/specs/2026-08-18-Record-Immutability-Requirements.md`
    (RECORD-IMMUTABILITY-REQ-001: 12 FRs, 7 NFRs, 7 OQs)
  - `docs/specs/2026-08-18-Record-Immutability-Design-Specification.md`
    (RECORD-IMMUTABILITY-DES-001: 10 design elements, 5 layers, 3
    phases, full traceability)
  Mechanism in one line: git history as the record + a no-rewrite policy
  enforced by guards (commit-guard amend block via the already-parsed
  inv.Amend; new pre-push non-fast-forward block) + commit signing and
  signed checkpoint tags anchored to a second remote + ADRs (Nygard) and
  an append-only docs/JOURNAL.md as the rationale/training-log corpus.
  Considered-and-rejected recorded with reasoning: Merkle infra, Rekor,
  sha256 repo migration, in-toto/SLSA. GATE: implementation waits on 7
  user decisions (OQ-01..OQ-07), most importantly OQ-07 (commit the
  uncommitted baseline -- needs commit approval) and OQ-01 (SSH vs GPG
  signing key -- user generates). Also recorded: five of six guards
  carry a live bypass var (verify-guard is the test-pinned exception),
  contradicting AGENTS.md's no-bypass claim.
- 2026-08-18 (same session, fifth slice): Record Immutability triad
  COMPLETE and critic-approved. Third document landed:
  `docs/specs/2026-08-18-Record-Immutability-Technical-Specification.md`
  (RECORD-IMMUTABILITY-TECH-001, 12 sections, implementation-grade:
  commit-guard amend block as new rule family IN-H-001 placed before the
  bypass check so it cannot be bypassed; fourth `--pre-push` guard mode
  with full stdin ref-update parsing and zero-sha edge cases;
  scripts/checkpoint-immutability.sh with signed-tag chaining that
  refuses unsigned tags; decision-neutral SSH/GPG signing config;
  Nygard ADR system incl. full ADR-0001 draft; docs/JOURNAL.md templates;
  phase-2 guard-log hash chaining sketch; 50-test-grounded test plan;
  12-step gated rollout). Critic round 1 found one BLOCKER: the
  emergency-rewrite hatch (orchestrator-briefed) contradicted approved
  REQ-001 FR-003 / DES-001 s4 -- reframed as PROPOSED amendment OQ-08,
  NOT APPROVED, excluded from rollout. Critic round 2: APPROVE; the
  critic also empirically reproduced and withdrew its own M6 claim
  (git tag --sort=-version:refname handles -r2/-r10 correctly; its
  suggested replacement would have introduced a real bug).
- 2026-08-18 (same session, sixth slice): anomaly cleanup DONE and
  critic-approved (3 rounds). Fixed: .claude-plugin/marketplace.json (4
  guards + bypass advertising -> six guards + five-of-six bypass
  reality); plugins/rr-policy-guards/README.md (phantom
  plugin.json/hooks/hooks.json layout refs, packaged-config pointer,
  bash-rotation overstatement, and the round-2 catch at :16 "no bypass
  variables" -> five-of-six + verify-guard exception); backstage/
  package.json (four dead descriptor-scoped minimatch pins removed;
  resolution no-op, yarn.lock untouched); tools/namespace-predictor/
  main.go:18 (comment path typo -> internal/dataplane/kubernetes/
  name.go; go vet + canonical vector verified); TODO.md/CHANGELOG.md
  stale gitea-com-blocked/origin entries (dated annotations appended,
  history preserved, push status honestly UNVERIFIED); root README.md:42
  (four hooks -> six guards). Certificate re-issued r4
  (PRC-developer-portal-2026-08-18-r4; PROVENANCE.md digest
  a6c647b7..., TPL digest unchanged) after the U7 row recorded the
  main.go correction. NOT done: ObservabilityLinksCard localhost:8080
  (gated on roadmap OQ-03 canonical SigNoz path); TODO.md pre-existing
  em-dashes (cosmetic, pre-existing).
- 2026-08-18 (same session, seventh slice): guard enforcement
  IMPLEMENTED and critic-approved per RECORD-IMMUTABILITY-TECH-001.
  rr-commit-guard gained: (a) IN-H-001 amend block in PreToolUse mode,
  placed before the bypass check so it cannot be bypassed (bypass-ignored
  test-pinned); (b) fourth mode `--pre-push` with IN-H-002 blocking
  deletion of main and non-fast-forward updates of main (githooks(5)
  stdin parsing, merge-base --is-ancestor, fail-closed incl. the
  lost-race case); new git-hooks/pre-push wrapper (no bypass comments)
  + installer updated to three hooks (NOT run -- activation stays with
  the user; .git/hooks untouched). Tests: 63/63 pass (50 pre-existing +
  13 new); e2e against real git verified independently by the critic;
  binary rebuilt at plugins/rr-policy-guards/bin/rr-commit-guard.
  README three-hook text + layout tree updated. ACCEPTED RESIDUAL
  (documented): IN-H-001 fires only when `git commit --amend` is the
  leading invocation; a compound-hidden amend (e.g. `git add x && git
  commit --amend`) passes the PreToolUse gate -- same risk class as a
  raw-terminal amend; IN-H-002 blocks publishing it to main at push
  time; extractor widening deferred to the commit-guard's own spec.
  AGENTS.md rr-commit-guard row updated to match.
- 2026-08-18 (same session, eighth slice): checkpoint script
  IMPLEMENTED and critic-approved per TECH-001 s4/s10:
  `scripts/checkpoint-immutability.sh` (signed annotated
  checkpoint-YYYY-MM tags, prev:-chained via the M6-adjudicated
  --sort=-version:refname, refuses unsigned tags AND refuses when
  either origin/github remote is missing -- preflight added after the
  critic's MINOR-1; verify-before-push; dry-run via env var or
  --dry-run) plus `scripts/tests/test-checkpoint-immutability.sh`
  (10/10 PASS: refusal, signed happy path vs throwaway SSH key, -r2
  rerun chaining, base/-r2/-r10 -> chains to -r10, dual-remote push,
  missing-remote refusal, dry-run purity, bash -n + shellcheck).
  Real repo untouched (zero checkpoint tags, no config changes, .git/
  hooks untouched). AGENTS.md gained a Record immutability command note.
- 2026-08-18 (same session, ninth slice): rationale layer INSTANTIATED
  and critic-approved per TECH-001 s6/s7: docs/adr/ (TEMPLATE.md --
  consumes no decision number; 0001-record-architecture-decisions.md --
  accepted 2026-08-18; README.md index) and docs/JOURNAL.md (header +
  13 [seed]-marked retrospective entries: origin, M1-M4, and the eight
  2026-08-18 slices + this one; end-of-seed-block marker bars
  non-contemporaneous appends). Critic line-checked every seed-entry
  fact against the state docs and live artifacts: all accurate.
  Files-only; OQ-07 (baseline commit) remains genuinely open.
- UNGATED WORK IS NOW EXHAUSTED. Remaining: everything is gated on the
  user's decision batch (presented 2026-08-18, turn 2 of waiting):
  Tier 1 = OQ-07 baseline commit approval, OQ-01 signing key (SSH
  recommended; user generates), Colima start (6 provenance U-rows +
  smokes). Tier 2 = OQ-02..06, OQ-08 (recommendations recorded in TODO/
  session report). Tier 3 = roadmap OQ-19/15/20/25/03 + Phase 1
  approval. If the batch stays unanswered next turn, mark the goal
  blocked (3-turn rule).
- 2026-08-18 (same session, TIER-1/TIER-2 EXECUTED): the user approved
  the batch (Tier 1 with a backup + staged-commit + zero-medium+
  condition; Tier 2 blessed; Tier 3 = pull the security plane forward,
  "more than just functional", enterprise-class bar, no more question
  batches). Executed:
  - Backup catalogue: ~/Projects/Sovereign/backups/developer-portal/
    2026-08-18/ (repo-snapshot.tar.gz 34.3 MiB, 1452 entries, sha256 in
    MANIFEST.md; verified).
  - Cluster up (user started Colima): all 7 cluster-blocked U-rows
    resolved live (Gitea 12.5.0/1.25.4, Gatekeeper v3.17.1 = repo pin,
    k3s v1.32.9+k3s1, cert-manager v1.19.4, Argo v3.6.2, envoy
    distroless-v1.33.0, act_runner 0.3.1 + dind 29.4.0); cert r5.
  - Signing: repo-local SSH signing on the user-designated key
    ~/.ssh/id_ed25519_pqc (an orchestrator-generated key was an error;
    user corrected; removed; allowed_signers set; sign/verify proof
    passing).
  - Staged series: 10 signed commits S1-S10 (d85e568..80ae9bd, all
    %G?=G), per-stage checks green. The FINAL sweep FAILED the
    zero-medium+ gate (22 yarn advisories incl. vm2 criticals, 9
    semgrep blocking, 13 gitleaks FPs, govulncheck missing) --
    checkpoint tag correctly withheld.
  - Remediation: 19/22 yarn advisories eliminated (vm2 eradicated via
    typescript-json-schema 0.68.0; 9 pins bumped); the 3 react-router
    moderates are unfixable (no fixed 6.x; v7 outside all @backstage
    peer ranges) -> accepted residual risk in SECURITY.md + provenance.
    Yarn 4.4.1 -> 4.18.0 (required for npmMinimalAgeGate "7d"). 6 CI
    action tags pinned to SHAs. .gitleaksignore for 13 proven false
    positives. govulncheck v1.7.0 installed; all 8 Go roots clean
    (hello-m2 x/net v0.56.0, x/text v0.39.0, x/sys v0.46.0). Commits
    S11 f0c5d10 + S12 82783ee (both G). Provenance regenerated, cert
    r6 (192 entries).
  - FIRST SIGNED CHECKPOINT: tag checkpoint-2026-08 (head 82783ee,
    prev: none) pushed to origin (gitea.com) AND github -- the
    immutable record's anchoring has begun. The tag push also proves
    gitea.com push auth works (the long-standing blocker is resolved).
  - Gate caveat (honest): zero medium+ is met except the 3 documented
    react-router moderates; absolute zero needs upstream Backstage
    v7-compatible peers or a Backstage upgrade.
  - Left for the user: branch protection on gitea.com (OQ-06; needs
    gitea.com admin UI/PAT -- no credential available to agents);
    .env.local holds a live Vercel OIDC token (gitignored, local-only,
    deliberately still flagged by gitleaks dir -- rotate if ever
    shared); .claude/ stays untracked by design.
- 2026-08-18 (same session, TIER-3 KICKOFF): security plane
  pull-forward requirements landed and committed (fbbd26d, signed):
  `docs/specs/2026-08-18-Security-Plane-Pull-Forward-Requirements.md`
  (SEC-PLANE-PULLFORWARD-REQ-001; 15 FRs / 10 NFRs / 6 decisions;
  critic-approved + 3 framing fixes). Grounded in a two-lane verified
  research pass. Load-bearing facts of record:
  - HOST REALITY: the Colima VM is 2 vCPU / 3.9 GB (older 6c/10GB
    claims are stale); ~84% memory used. Wave 0 = zero-new-standing-
    workload items only.
  - Wave 0 (now): Trivy CLI + OSV-Scanner in CI pinned by digest
    (March 2026 Trivy supply-chain compromise CVE-2026-33634 is the
    cited reason); Gatekeeper violation visibility (constraint
    .status.violations + gatekeeper_violations metric + audit JSON to
    OTEL/SigNoz); custom Security tab (Roadie plugin is GitHub-only,
    useless vs Gitea); RBAC custom permission policy (admin/developer/
    viewer from Gitea group claims); TLS via Certificate resources on
    the existing Gateway; dependabot.yml + code scanning; guard-log
    hash chaining (resolves the "at all" half of OQ-04).
  - Wave 1 (after one documented Colima resize to >=6c/12GB): Falco
    0.44.1 (modern_ebpf verified working on this kernel: 6.8.0-100,
    BTF present) + Falcosidekick OTLP -> SigNoz; Trivy Operator;
    MISP 2.5.44 slim (AGPL-3.0, ~3-4GB) as the threat-intel platform
    of record with CIRCL feed + restSearch egress.
  - Wave 2 (scale-out docs only): TheHive 5 DISQUALIFIED (license
    drift to proprietary; 3/4 AGPL EOL), Wazuh/OpenCTI deferred on
    capacity, Velociraptor lab-only, Cloud Custodian deferred.
  - SigNoz pipeline is the security-event sink, honestly labeled
    "security observability, not a SIEM".
  - Environmental flags: cert-manager 1.19 EOL 2026-07-08 (upstream
    1.21.x; sibling-owned upgrade); Envoy Gateway pin 1.3.1 vs
    upstream v1.9.0.
- 2026-08-18 (same session, TIER-3 SPEC COMPLETE): Wave-0 technical
  specification landed and committed (8ff505a, signed):
  `docs/specs/2026-08-18-Security-Plane-Wave0-Technical-Specification.md`
  (SEC-PLANE-WAVE0-TECH-001, ~73KB, 14 sections). Implementation-grade
  for all 11 Wave-0 FRs; external pins re-resolved independently by the
  critic and byte-exact (trivy 0.74.0 sha256:62b1e65e..., osv-scanner
  v2.5.1 sha256:8108ae94..., codeql-action v4.37.7 peeled ff2f1c62...);
  live cluster claims verified (no gatekeeper metrics Service, 4 pods
  on 8888, zero current violations); one BLOCKER found and fixed
  (policyExtensionPoint is alpha-subpath-only) + smoke-suite ownership
  assigned (accretes per lane; smoke-all edit is acceptance-time). The
  security plane triad is now complete (REQ fbbd26d + TECH 8ff505a).
  Five implementation lanes: A CI scanning (FR-01..03), B Gatekeeper
  visibility (FR-05..07), C portal surfaces (FR-04, FR-08), D
  infra/config (FR-09 TLS, FR-10 dependabot/CodeQL), E guards (FR-11
  hash chaining).
- Also flagged this pass (not fixed, noted): scripts/
  start-backstage.sh:5 hardcodes BACKSTAGE_DIR=/Users/nnos/Projects/
  developer-portal/backstage (the non-Sovereign path) - verify whether
  that path still exists/symlinks before relying on the script.
- 2026-08-18 (same session, WAVE-0 IMPLEMENTED): all five lanes of
  SEC-PLANE-WAVE0-TECH-001 are in the working tree, critic-reviewed (2
  critics, split for depth). Lane A (CI scanning: Trivy 0.74.0 +
  OSV-Scanner v2.5.1 digest-pinned gates in the seed workflow +
  template, commit-security-artifacts.sh, smoke harness) - APPROVED,
  gate proven locally (vulnerable fixture exit 1 with the exact CVE;
  clean tree exit 0). Lane B (Gatekeeper visibility: app-config
  localKubectlProxy, gatekeeper.ts, PolicyCard live rewrite, Prometheus
  gatekeeper scrape, collector filelog) - APPROVED. Lane C (Security
  tab + SecurityCard, RBAC SecurityRbacPolicy replacing allow-all) -
  APPROVED (2 spec errors found and fixed with evidence:
  createConditionalDecision does not exist in the installed tree ->
  createCatalogConditionalDecision; scalar YAML env-substitution would
  crash -> quoted flow list). Lane D (TLS via tls.tf issuer chain +
  HTTPS listeners, dependabot.yml 9 go.mod roots, code-scanning.yml
  pinned CodeQL) - APPROVED after fix-backs. Lane E (prev_hash hash
  chaining in all six guards + tools/audit-chain verifier, 14 tests) -
  APPROVED. Assembly pass: smoke-security.sh now 40 pass / 0 fail /
  9 skip (exit 0); legacy pre-chain guard logs archived aside as
  *.prechain (nothing deleted); dependabot audit-chain entry added;
  AGENTS.md audit-chain lines added. Four spec deviations all
  adjudicated CORRECT against evidence (msg->message dual-key,
  BACKSTAGE_DIR stale-path repair, the two Lane C substitutions).
  Known live-cluster caveats: prometheus-server + M3 collector pods
  Pending (2c/4GB host pressure - Wave 1 resize prerequisite);
  FR-06/FR-07 live checks SKIP until lifecycle re-runs.
- 2026-08-18 (WAVE-0 COMMITTED): provenance r7 (197 entries; Trivy,
  OSV-Scanner, CodeQL-MIT-verified, govulncheck firmed; chain wording:
  r5 at d85e568, r6 at 82783ee in history) and the seven-commit Wave-0
  series, all signed (%G?=G): d17a06e Lane A, 67ce13f Lane B, de1ac07
  Lane C, ba40190 Lane D, c10e4cb Lane E, 429e730 smoke suite,
  f20cff8 provenance r7. Tree clean except .claude/ (untracked by
  design).
- 2026-08-18 (WAVE-0 ACCEPTANCE, PART 1): lifecycle applies attempted.
  Results: smoke-security.sh 41 pass / 0 fail / 8 skip; Gatekeeper
  C1/C2/C3 verified live at 0 violations; guard hash chains verify
  live. TWO BLOCKERS surfaced (both pre-existing, neither caused by
  Wave-0):
  1. TOFU STATE DRIFT: install-m3.sh fails - helm releases signoz
     (0.130.1) and otel-collector (live chart 0.159.2 vs pin 0.155.0 -
     pin drift too) exist but are absent from the kubernetes-backend
     state ("cannot re-use a name that is still in use"). M4's
     opencost/prometheus releases likely same. Remediation: a
     sanctioned import step (new lifecycle script wrapping tofu import
     - direct import is guard-blocked), then re-run the applies.
  2. CAPACITY: the VM is 2 vCPU / 3.9 GB with memory requests 99%
     allocated; 31 pods Pending with "Insufficient memory" (chronic,
     x161 over 13h): prometheus-server, otel-collector, envoy gateway
     pod, M3 SigNoz pods. The live Prometheus/collector/HTTPS checks
     cannot pass until the documented Colima resize (>= 6 CPU / 12 GB,
     per the security requirements' Wave-1 prerequisite) happens -
     that resize is a USER step (stopping the VM takes the platform
     down briefly; not an agent action).
  Envoy gateway pod Pending also means the .local routes are
  currently down (smoke-m4-networking 6/6 FAIL for that reason only).
- Next: state-drift remediation (sanctioned import script + imports +
  re-apply M3/M4/networking), then live CI acceptance; the remaining
  live checks then await the user's Colima resize.

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
