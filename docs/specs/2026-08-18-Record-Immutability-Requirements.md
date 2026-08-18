# Requirements Specification: Record Immutability for Project Documents

**Document ID:** RECORD-IMMUTABILITY-REQ-001
**Version:** 0.1
**Date:** 2026-08-18
**Status:** Draft for review -- nothing in this document is approved or decided yet
**Relationship:** Companion to `2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md`, whose OQ-10 deferred record-immutability scope to this workstream and whose NFR-05 / FR-40 / CTL-G6 point here. A Design Specification (RECORD-IMMUTABILITY-DES-001) accompanies this document; the Technical Specification follows to complete the governance triad per `~/Projects/Sovereign/Structure/POLICIES.md`.

---

## 1. Purpose

This Requirements Specification defines the mandatory properties of the record-immutability mechanism for project documents in the developer-portal repository: every change to the documents that describe and govern the project is captured from the beginning, in a form that is tamper-evident and that preserves the reasoning behind each change.

The mechanism serves two ends, both stated in the governing directive (section 2): it makes mistakes harder to hide, and it produces a training log of the mindframe and process that were taken -- including how and why that process shifted as understanding changed.

**Evidence discipline:** every current-state claim in section 4 cites the repo path or configuration fact verified on 2026-08-18 (orchestrator handoff, sections 1 and 2; key claims spot-checked against the repo during authoring). Items the inputs marked UNVERIFIED are carried forward as UNVERIFIED. Nothing in this document asserts a property of git that git does not have (NFR-006).

---

## 2. Goal and rationale

**Goal (verbatim from the directive, 2026-08-18):**

> "All project documents should capture all changes that are made from the beginning. This makes it more difficult to hide mistakes, but also becomes a training log for understanding the mindframe and process that was taken and why it may have shifted due to a new understanding. Very important to capture for the development of AGI."

**Rationale (factual):**

- **Mistake-hiding resistance.** Git history is already the primary record of this repository, but history only resists concealment if it is never rewritten and if replacement of the published branch is detectable. Neither holds today: commit amends pass the existing guard unexamined, force-push is undifferentiated from ordinary push, and the repository ships two scripts whose entire function is amend-plus-force-push (section 4). Git's hash-chained DAG provides tamper-evidence; rewrites are what break it, and git alone cannot detect a force-push (research, handoff section 1, PRIMARY RECORD / GIT LIMITS).
- **Training log.** A commit history answers "what changed"; the directive also demands "the mindframe and process that was taken and why it may have shifted." That rationale layer does not exist in durable form today: there is no ADR directory and no engineering journal, and the only machine-enforced rationale capture is the mandatory WHY body in commit messages (IN-M-003, `plugins/rr-policy-guards/tools/commit-guard/validator.go:66-70`). Decision records (Nygard-format ADRs) and an append-only engineering journal are the two complementary artifacts the research identifies for this (handoff section 1, RATIONALE CAPTURE).
- **AGI development.** The corpus produced by the rationale layer -- ADR supersession chains, journal entries, and WHY-explaining commit bodies -- is precisely the process record the directive calls "very important to capture for the development of AGI." This document makes no claims beyond producing that corpus faithfully.

---

## 3. Scope

### 3.1 In scope: the project documents that matter

The mechanism covers all project documents whose silent loss or retroactive edit would damage the record, identified in the verified repo map (handoff section 2, PROJECT DOCUMENTS THAT MATTER):

- `docs/specs/` -- the governance specification triads (28 dated specification files directly in docs/specs/, 34 Markdown files in total, plus 9 more in the m1-substrate/, m2-iac-cd/, and m3-observability/ milestone subdirectories; verified 2026-08-18)
- `TODO.md`, `SESSION_HANDOFF.md`, `PROJECT_SUMMARY.md` -- the session-state documents AGENTS.md mandates keeping current
- `CHANGELOG.md`, `AGENTS.md`, `catalog-info.yaml`
- `policies/` -- the Gatekeeper constraint triads (C1-C3)
- `iac/` -- the OpenTofu root composition and modules
- `seed-repos/` -- the Gitea seed content (platform-addons, platform-config, hello-m2)
- `provenance/` + `THIRD-PARTY-LICENSES.md` -- the attribution triple
- `scripts/ci/` -- the CI commit-automation scripts
- `.github/workflows/sync-from-gitea.yml` -- the mirror automation

The canonical branch is `main` on the `origin` remote (`https://gitea.com/trademomentum.net/developer-portal.git`); per AGENTS.md, all commits land on `main`.

### 3.2 Considered and rejected (out of scope)

Each of the following was evaluated in the research (handoff section 1) and is **CONSIDERED AND REJECTED** for this mechanism; the reasoning is recorded here so the rejection itself is part of the immutable record:

- **Merkle / transparency-log infrastructure (Trillian-style).** For a single sequential writer with a small log, a linear hash chain (Haber-Stornetta 1991) gives the same tamper-evidence as a Merkle tree; Merkle logs earn their complexity only for many writers, O(log n) proofs, or untrusted proof servers. Git already provides the linear chain.
- **Rekor integration.** A public transparency log is a notary for third-party auditability; this mechanism's auditability is first-party (author + future readers of one repository). The public-good default also stays Rekor v1 per the Sigstore 2026-06-28 blog "Rekor evolution," and URL/version interop for v2 is unverified -- not a foundation to build on now.
- **SHA-256 object-format migration.** Git supports sha256 repositories since 2.29 and Git 3.0 plans to change the default, but Gitea sha256 support is **UNVERIFIED**; migrating the object format would jeopardize the forge, mirror, and CI tooling for a collision-risk reduction the single-writer setting does not yet require. The repository stays SHA-1 for now; the risk is acknowledged in NFR-006.
- **in-toto / SLSA attestation pipelines.** These are artifact-centric supply-chain standards (build provenance for software outputs), not document-history capture; adopting them here would add a pipeline of attestations for a problem git plus signing already covers, and their interop with this repository's Gitea Actions CI is unverified.

### 3.3 Out of scope (other)

- **Implementation.** This document and its Design Specification change no code, hooks, or git state; the Technical Specification and its phases do.
- **Server-side enforcement configuration on gitea.com** (branch protection). Recorded as a user action (OQ-06); current state UNVERIFIED.
- **The five-plane portal initiative itself.** This mechanism is a dependency of that initiative (its FR-40 requires new artifacts to be compatible with this mechanism), not a part of it.
- **Local Gitea deployment target** (`localhost:3333/openchoreo/developer-portal`). It exists but is not a configured remote today (handoff section 2, REMOTES); this mechanism anchors to the two distinct remote targets only (gitea.com and github.com -- three remotes are configured, but `gitea-com` duplicates the `origin` URL; see section 4).

---

## 4. Current state (verified 2026-08-18)

| Area | Verified state | Evidence |
|---|---|---|
| Canonical repo and remotes | Three configured remotes, two distinct targets: `origin` and `gitea-com` both point at `https://gitea.com/trademomentum.net/developer-portal.git` (duplicate URL; origin in sync at `67a17f9`); `github` = `github.com/trademomentum-llc/developer-portal` (Vercel-facing mirror) | `git remote -v`, `git log` (verified 2026-08-18, `gitea-com` duplicate included); handoff section 2, REMOTES |
| Mirror automation | Pull-side mirror on a 5-minute cron: `git clone --mirror` then `git push --mirror`; propagates rewritten history; adds redundancy, not tamper-evidence | `.github/workflows/sync-from-gitea.yml:42,51` |
| Commit signing | None: `commit.gpgsign`, `user.signingkey`, `gpg.format` unset repo and global; `%G?` = N on recent commits | handoff section 2, NO SIGNING (spot-verified via `git log --format='%G?'`) |
| Rewrite protection | None active: the only active hooks in `.git/hooks/` are `pre-commit` and `commit-msg` (rr-commit-guard wrappers) -- no active `pre-push` or `pre-receive` hook; the remaining 14 entries are inert git `.sample` files (including `pre-push.sample` and `pre-receive.sample`); no pre-receive logic anywhere; commit-guard parses `--amend` into `inv.Amend` but the flag is never consulted, so amends pass; verify-guard classifies `git push` generically with force-push undifferentiated | `.git/hooks/` listing (verified 2026-08-18); `plugins/rr-policy-guards/tools/commit-guard/extract.go:90-91`, `types.go:73`; `plugins/rr-policy-guards/tools/verify-guard/shell.go:163-164` |
| Rewrite scripts in-repo | `scripts/git-fix-email.sh` and `scripts/git-fix-author.sh` each run `git commit --amend` + `git push --force origin main` | `scripts/git-fix-email.sh:9-10`, `scripts/git-fix-author.sh:9-10` |
| Commit-guard today | Enforces staged-file Never/Grey lists, subject <= 72 chars (IN-M-001, `validator.go:55-58`), Conventional Commits form (IN-M-002), and a mandatory body explaining the WHY (IN-M-003 -- the only machine-enforced rationale capture). Five of the six guards carry a live, audited bypass variable: `RR_BASH_GUARD_BYPASS` (`bash-guard/main.go:73`), `RR_BREW_GUARD_BYPASS` (`brew-guard/main.go:82`), `RR_COMMIT_GUARD_BYPASS` (`commit-guard/bypass.go:9`), `RR_EMOJI_GUARD_BYPASS` (`emoji-guard/main.go:86`), `RR_TOFU_GUARD_BYPASS` (`tofu-guard/main.go:57`). verify-guard is the documented exception (`verify-guard/main.go:13`: "Verification has no bypass path"; its tests set `RR_VERIFY_GUARD_BYPASS` only to assert it does NOT bypass). The five live bypasses contradict the AGENTS.md "no bypass variables" claim | `plugins/rr-policy-guards/tools/commit-guard/validator.go:55-70`; the five guard `main.go`/`bypass.go` files cited in-line; `plugins/rr-policy-guards/tools/verify-guard/main.go:13` + `main_test.go` (bypass extent re-verified 2026-08-18); handoff section 2, COMMIT-GUARD ENFORCES TODAY |
| Guard audit logs | Append-only JSONL at `~/.rational-reserve/logs/<guard>-guard.jsonl`, mode 0600, fail-open by design (enforcement never depends on audit writes); no hash chaining or anchoring; rotation only in verify-guard (8 MiB x3); local-only, deletable by the same user they record | handoff section 2, GUARD AUDIT LOGS |
| Provenance certificate chain | Current certificate `PRC-developer-portal-2026-08-18-r3` embeds SHA-256 digests of the two listing files; the supersede rule says old certificates stay in git history -- but `provenance/` is untracked, so the retained chain has not actually started (r1/r2 exist in no retained history) | `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md`; `git status` 2026-08-18; handoff section 2, PROVENANCE CERT CHAIN |
| CI commit automation | `scripts/ci/commit-to-platform-config.sh` commits rendered Component YAML to `openchoreo/platform-config` main via the Gitea Contents API as `gitea_admin` (no PR/review); `commit-cost-artifact.sh` is similar | handoff section 2, CI COMMIT AUTOMATION |
| Rationale conventions | `docs/specs/` triads mandated and present (28 dated specification files, 34 Markdown files in total, plus 9 in the milestone subdirectories); `SESSION_HANDOFF.md`/`TODO.md`/`PROJECT_SUMMARY.md` update mandate in AGENTS.md; `CHANGELOG.md` sparse (Keep-a-Changelog, one Unreleased section); no ADR directory; no engineering journal | `ls docs/specs/` (verified 2026-08-18); handoff section 2, RATIONALE CONVENTIONS |
| Uncommitted baseline | Modified: `AGENTS.md`, `SESSION_HANDOFF.md`, `THIRD-PARTY-LICENSES.md`, `TODO.md`, `plugins/rr-policy-guards/tools/verify-guard/main.go`, `tools/score2openchoreo/assets/SCHEMA_PROVENANCE.md`. Untracked: `.claude/`, `docs/specs/2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md`, `provenance/` (2 files) | `git status` 2026-08-18; handoff section 2, CURRENTLY UNCOMMITTED |
| Stale claims in docs | `TODO.md` claims origin = `localhost:3333` and gitea-com push BLOCKED; both false today; `CHANGELOG.md` repeats the blocker | handoff section 2, STALE DOCS NOTED |

**Reading of the current state.** The repository already has: a single canonical branch, a working guard framework with audit, a mandatory-WHY commit-message rule, and a provenance-certificate pattern that already practices "re-issued, never edited." It lacks: any rewrite prohibition, any signature, any anchor outside the primary forge, any decision record, and any journal. The requirements below close exactly those gaps, reusing what exists.

---

## 5. Functional requirements

Every requirement traces to the verified current state (section 4) and the research conclusions (handoff section 1). Where a requirement depends on a user decision, the OQ reference is given and the requirement is stated at the level of the decision, not the outcome.

### 5.1 The record itself

- FR-001: **Baseline commit.** All currently uncommitted and untracked project documents (section 4, Uncommitted baseline) are committed so the captured record starts from a complete baseline; this includes `provenance/`, whose first commit actually starts the retained provenance-certificate chain. The baseline commit requires explicit user commit approval (OQ-07) and is the mechanism's first test of itself. **Traces to:** handoff section 2, CURRENTLY UNCOMMITTED + PROVENANCE CERT CHAIN.
- FR-002: **No-rewrite policy for main.** The canonical branch is never rewritten: no `git commit --amend`, no rebase of published history, no force-push, no history-changing reset. Git's hash-chained DAG provides tamper-evidence only while this holds; a working example of such a policy is the `github.com/vladi160/preregistrations` README. **Traces to:** handoff section 1, PRIMARY RECORD; section 4, Rewrite protection.
- FR-003: **Documented never-rewrite exception model.** There is no procedure by which history is legitimately rewritten. Mistakes are corrected by NEW commits that reference the mistaken commit; reversed decisions are new records that supersede the old; superseded records are never edited or deleted. This one rule applies uniformly to commits, ADRs, journal entries, and provenance certificates (which already operate this way). **Traces to:** handoff section 1, RATIONALE CAPTURE (superseded-not-deleted); section 2, PROVENANCE CERT CHAIN.

### 5.2 Integrity and anchoring

- FR-004: **Commit signing.** Commits on the canonical branch are cryptographically signed by their author (signature scheme per OQ-01; SSH signing recommended: simplest option per the GitHub docs, locally verifiable, no PKI, git >= 2.34; GPG if expiry/revocation semantics are wanted). **Traces to:** handoff section 1, COMMIT SIGNING; section 4, Commit signing.
- FR-005: **Signed checkpoint tags pushed to a second remote.** At a periodic cadence (per OQ-02; monthly recommended), a signed annotated tag binds the signer to the exact head commit and tree of the canonical branch and is pushed to a second remote, so that unilateral replacement of the published history becomes visible. The tag is the anchor; the second remote is redundancy, not evidence, because the existing sync mirror propagates rewrites (NFR-006). **Traces to:** handoff section 1, CHECKPOINTS/ANCHORING; `.github/workflows/sync-from-gitea.yml:42,51`.
- FR-006: **Optional OpenTimestamps layer.** If enabled (OQ-03), checkpoint tags (or their digests) are additionally timestamped via OpenTimestamps, adding Bitcoin-anchored, offline-verifiable trusted time (`.ots` proofs) that git commit dates cannot provide. OTS calendar liveness is **UNVERIFIED**; the layer is optional and the mechanism is complete without it. **Traces to:** handoff section 1, CHECKPOINTS/ANCHORING + GIT LIMITS.

### 5.3 Rationale capture (the training log)

- FR-007: **Architecture Decision Records.** A `docs/adr/` set in Nygard format (Title, Status, Context, Decision, Consequences; statuses proposed/accepted/deprecated/superseded), numbered sequentially with numbers never reused; reversed decisions are kept and marked superseded with a reference to the replacement, never altered or deleted. ADR-0001 records the decision to adopt ADRs. **Traces to:** handoff section 1, RATIONALE CAPTURE (Nygard 2011; 2026 arXiv study favors Nygard for concise records; MADR heavier, Y-statements lighter); section 4, Rationale conventions.
- FR-008: **Engineering journal.** An append-only `docs/JOURNAL.md` -- the contemporaneous chronological log of what was tried, what failed, and what was learned: "the record made during the work, not a reconstruction." Seeded with the project's origin entries at baseline. **Traces to:** handoff section 1, RATIONALE CAPTURE (lab notebook); section 4, Rationale conventions.
- FR-009: **Training-log consumability.** The ADR supersession chain, the journal, and the IN-M-003 commit bodies together form the mindframe/process corpus the directive requires; all three are plain text, committed in the repository, and greppable, so the corpus is consumable by future readers (human or machine) without special tooling. `git notes` is never the system of record (not pushed/fetched by default; mutable without changing any commit hash). **Traces to:** handoff section 1, RATIONALE CAPTURE; `validator.go:66-70`.

### 5.4 Enforcement

- FR-010: **Amend blocked at the guard.** The existing commit-guard blocks `git commit --amend`, consuming the already-parsed `inv.Amend` invocation flag (`extract.go:90-91`, defined `types.go:73`) as a block decision with an audit record. **Traces to:** FR-002; section 4, Rewrite protection.
- FR-011: **Force-push blocked pre-push.** A pre-push control (new git hook or guard extension) rejects force-push operations targeting the canonical branch, following the existing guard audit pattern (append-only JSONL, fail-open audit) and the AGENTS.md no-bypass rule for new logic. Server-side branch protection on gitea.com is documented as the complementary user action (OQ-06). **Traces to:** FR-002; section 4, Rewrite protection; handoff section 1, PROCESS GLUE (policy encoded where it is enforced, not only in docs).
- FR-012: **Guard audit-log integrity (phase-2 proposal).** The guard JSONL audit logs gain hash chaining (each entry binds the hash of its predecessor) so that deletion or editing of the local audit trail is detectable; scope per OQ-04. The logs remain local-only and user-owned; chaining makes tampering evident, not impossible. **Traces to:** section 4, Guard audit logs; handoff section 1, LINEAR CHAIN SUFFICES.

---

## 6. Non-functional requirements

- NFR-001 **Open-source only.** Every tool the mechanism adopts (git itself, the existing Go guards, optionally the OpenTimestamps client) is open-source and license-compatible with the portfolio; no proprietary service enters the record path.
- NFR-002 **No new heavy infrastructure.** The mechanism is git, the existing guard framework, shell scripts, and optionally one small CLI (ots). No servers, databases, or log infrastructure are added; the single-writer linear hash chain suffices by design (section 3.2).
- NFR-003 **Fail safe and audited.** Enforcement follows the existing guard contract: block decisions are recorded in the append-only audit log, and the audit layer is fail-open -- a failed audit write never changes an allow/block decision (enforcement never depends on audit writes). New blocking logic carries the same test coverage as existing guard rules.
- NFR-004 **Plain ASCII.** All artifacts the mechanism introduces (ADRs, journal, checkpoint tag messages, these specifications) are plain ASCII, consistent with the rr-emoji-guard constraint already in force.
- NFR-005 **Meta-consistency.** Every component of the mechanism itself produces immutable records: these specifications are committed, ADR-0001 records the mechanism's own adoption, the journal records its rollout, and the checkpoint tags anchor the history that contains all of the above. The mechanism must never demand of project documents a discipline its own artifacts do not satisfy.
- NFR-006 **Honesty about git's limits.** These specifications must not claim properties git does not have. Specifically: commit author/committer dates are attacker-controlled, so trusted time requires an external notary (FR-006); git alone cannot detect a force-push, so anchoring (FR-005) and server-side protection (OQ-06) are separate controls, and the existing sync mirror propagates rewrites -- it is redundancy, not evidence; the SHA-1 object format carries collision risk, accepted for now per section 3.2. UNVERIFIED items (OTS calendar liveness, Gitea sha256 support, gitea.com branch-protection state) stay labeled UNVERIFIED.
- NFR-007 **Non-regression.** The existing guard suites (27 Go test files across the six guards), `opa test --v0-compatible policies/*.rego` (6/6), and the install/smoke script contracts remain green after each implementation phase; the commit-guard's existing staged-file and message rules are unchanged in behavior except for the new amend block.

---

## 7. Open questions (user decisions)

Each entry states why it matters and what it blocks. UNVERIFIED markers are preserved from the inputs.

| ID | Question | Why it matters / blocks | Recommendation from the research |
|---|---|---|---|
| OQ-01 | SSH or GPG signing key, and key generation (user action: generate and register the key)? | Blocks FR-004; nothing can be signed until the keypair exists and its public half is registered where verification happens | SSH: simplest per the GitHub docs, locally verifiable, no PKI, git >= 2.34. GPG only if expiry/revocation semantics are wanted. gitsign rejected: no GitHub Verified badge, needs network to Fulcio/Rekor, needs `gitsign verify` |
| OQ-02 | Checkpoint cadence? | Blocks FR-005 automation; too frequent is noise, too rare widens the unanchored window | Monthly signed tag (`checkpoint-YYYY-MM`) |
| OQ-03 | OpenTimestamps on or off? | Blocks FR-006; OTS calendar liveness is UNVERIFIED, and the mechanism is complete without it | Decide after OQ-01/OQ-02 land; off by default until calendar liveness is verified |
| OQ-04 | Guard audit-log hash-chaining scope: all guards or verify/commit only, and is chaining in the phase-2 batch at all? | Blocks FR-012 sizing; the logs are local-only and user-owned, so chaining buys tamper-evidence against casual edits, not a determined operator | Chain commit-guard and verify-guard first (the two that gate publication) |
| OQ-05 | Disposition of the pre-existing rewrite scripts `scripts/git-fix-email.sh` and `scripts/git-fix-author.sh`? | FR-002 makes their function (amend + force-push) prohibited; keeping executable rewrite scripts in the repo contradicts the policy they would now violate | Delete them in a new commit (the deletion is itself an immutable record); their history remains in git |
| OQ-06 | Enable server-side branch protection on gitea.com (user action; current state UNVERIFIED)? | Git alone cannot detect a force-push; server-side rejection of force-push on `main` is the only control that does not depend on the pusher's own machine | Enable force-push rejection for `main` on gitea.com after Phase 1 lands |
| OQ-07 | How to handle the currently uncommitted baseline (section 4): one baseline commit or several logical commits, and who commits it? | Blocks FR-001; committing is a git mutation, which requires explicit user approval -- this is the mechanism's first test of its own discipline | Several logical commits with IN-M-003-compliant WHY bodies, committed by the user or with explicit approval |

---

**End of Requirements Specification**

The Design Specification (RECORD-IMMUTABILITY-DES-001) defines the layered architecture, per-layer mechanics, the exception model, phasing, and traceability. The Technical Specification (exact guard changes, hook code, checkpoint script, ADR/journal templates, signing configuration) completes the triad.
