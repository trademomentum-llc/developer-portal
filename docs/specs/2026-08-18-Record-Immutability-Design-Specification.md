# Design Specification: Record Immutability for Project Documents

**Document ID:** RECORD-IMMUTABILITY-DES-001
**Version:** 0.1
**Date:** 2026-08-18
**Status:** Draft for review -- nothing in this document is approved or decided yet
**Predecessor:** RECORD-IMMUTABILITY-REQ-001 (`2026-08-18-Record-Immutability-Requirements.md`)

---

## 1. Design philosophy

The mechanism is deliberately thin: the record already exists (git history), the enforcement framework already exists (the rr-policy-guards), and the rationale conventions partially exist (IN-M-003 commit bodies, the provenance-certificate supersede rule). The design adds only what is missing -- a no-rewrite rule with teeth, signatures, an anchor outside the primary forge, and the two rationale artifacts (ADRs, journal) -- and wires them together so that each layer reinforces the others.

Three principles govern every design element:

- **Reuse before build.** The single-writer linear hash chain that git already provides is sufficient tamper-evidence (Haber-Stornetta 1991; handoff section 1, LINEAR CHAIN SUFFICES); no log infrastructure is built. The existing guard framework, audit pattern, and hook installer are extended, not replaced.
- **Policy is encoded where it is enforced** (handoff section 1, PROCESS GLUE): the no-rewrite rule lives in the guard and a hook, not only in these documents.
- **Honest limits.** The design does not claim properties git does not have (NFR-006): commit dates are attacker-controlled (trusted time is the optional anchoring layer's job, not git's); git alone cannot detect a force-push (the second-remote anchor and server-side branch protection are separate controls); the existing sync mirror propagates rewrites and is redundancy, not evidence.

The design is stated at the level of mechanics and contracts; exact code, rule codes, flag parsing, and file templates belong to the Technical Specification.

---

## 2. Architecture in layers

The mechanism is five layers. Lower layers provide the substrate; upper layers make tampering visible. The rationale and enforcement layers flank the record: one explains it, the other protects it.

### 2.1 Layer 1 -- Record: git history

The canonical record is the commit DAG on `main` of the `origin` remote (`https://gitea.com/trademomentum.net/developer-portal.git`). Git's hash chaining makes any retroactive edit detectable by anyone holding a prior head -- provided history is never rewritten. The no-rewrite policy (FR-002) and its exception model (FR-003, section 4) are therefore the load-bearing rules of the entire mechanism; everything else either signs this record, explains it, enforces the rule, or anchors it.

### 2.2 Layer 2 -- Integrity: commit signing and signed checkpoint tags

Signatures bind authorship to specific record states. Every commit on `main` is signed by its author (FR-004), and periodic signed annotated checkpoint tags bind the signer to the exact head commit and tree (FR-005). The signature scheme is decided by OQ-01; this document recommends SSH signing (locally verifiable, no PKI, supported since git 2.34; handoff section 1, COMMIT SIGNING) with GPG as the alternative if expiry/revocation semantics are wanted, and does not pre-decide the question. Signatures are the layer a history rewriter cannot forge: rewritten commits would need new signatures, and old signed checkpoint tags would no longer match.

### 2.3 Layer 3 -- Rationale: ADRs, journal, and WHY bodies

The training log proper. Architecture Decision Records in Nygard format (`docs/adr/`, sequential numbers never reused, superseded-not-deleted; FR-007) capture *decisions*; the append-only `docs/JOURNAL.md` (FR-008) captures the *contemporaneous process* -- what was tried, what failed, what was learned; and the already-enforced IN-M-003 commit bodies capture the *why of each change*. Cross-linked (D-07), the three form the mindframe/process corpus the directive asks for (FR-009).

### 2.4 Layer 4 -- Enforcement: guards and hooks

The no-rewrite policy is enforced where the operations occur (FR-010, FR-011): the existing commit-guard blocks `git commit --amend` using the invocation flag it already parses, and a new pre-push control rejects force-push to `main`. Both follow the existing guard contract -- append-only JSONL audit, fail-open audit writes, no bypass variable for new logic (NFR-003; see D-09 for the recorded bypass discrepancy). These are local controls on the pusher's machine; server-side branch protection on gitea.com (OQ-06) is the complementary control that does not depend on the pusher's machine.

### 2.5 Layer 5 -- Anchoring: second remote and optional trusted time

Checkpoint tags are pushed to a second remote (`github`, the Vercel-facing mirror) so unilateral replacement of the published history becomes visible beyond the primary forge (FR-005). The existing pull-side sync mirror (`git clone --mirror` / `git push --mirror` on a 5-minute cron, `.github/workflows/sync-from-gitea.yml:42,51`) propagates rewrites and therefore provides redundancy, not evidence: `push --mirror` also prunes deletions, so persistence of a tag on the github mirror against a gitea-side attacker is not assured. The anchor's evidentiary value is the signed tag itself -- unforgeable without the signer's key and held by every clone; the second remote adds detection convenience, not proof. An optional OpenTimestamps layer (FR-006) adds Bitcoin-anchored trusted time that git commit dates cannot provide; OTS calendar liveness is UNVERIFIED, so the layer is off by default (OQ-03).

```text
            +---------------------------------------------------+
            | L5 ANCHORING                                      |
            | checkpoint tags on second remote (github);        |
            | optional OpenTimestamps proofs                    |
            +------------------------^--------------------------+
                                     | anchors
            +------------------------|--------------------------+
            | L2 INTEGRITY           |                          |
            | SSH-signed commits; signed checkpoint tags        |
            +------------------------^--------------------------+
                                     | signs
            +------------------------|--------------------------+
            | L1 RECORD              |                          |
            | git DAG on main (origin, gitea.com);              |
            | never rewritten                                   |
            +---^-----------------^----------------------------+
                |                 |
     narrates   |                 |   gates
  +-------------|-----+   +-------|------------------------------+
  | L3 RATIONALE      |   | L4 ENFORCEMENT                       |
  | docs/adr/ (Nygard)|   | commit-guard: --amend block;         |
  | docs/JOURNAL.md   |   | pre-push hook: force-push block;     |
  | IN-M-003 bodies   |   | audit JSONL, fail-open, no bypass    |
  +-------------------+   +--------------------------------------+
```

---

## 3. Layer mechanics (design elements D-01 through D-10)

Each element names the requirement it implements and the evidence behind it. Exact code belongs to the Technical Specification.

### D-01 -- Baseline commit procedure (Phase 0)

A short, user-approved sequence of logical commits brings the currently uncommitted record under history (verified uncommitted set: `git status` 2026-08-18; handoff section 2, CURRENTLY UNCOMMITTED): (a) session-state documents (`AGENTS.md`, `SESSION_HANDOFF.md`, `TODO.md` -- including correction of the stale origin/blocked claims noted in handoff section 2, STALE DOCS NOTED); (b) the attribution triple (`THIRD-PARTY-LICENSES.md` + the untracked `provenance/` directory), whose first commit actually starts the retained provenance-certificate chain (r1/r2 exist in no retained history today); (c) the pre-existing `verify-guard/main.go` edit and the `SCHEMA_PROVENANCE.md` edit; (d) the five-plane requirements document and this specification pair. `.claude/` is local agent state and stays untracked. Every commit carries an IN-M-003-compliant WHY body referencing this specification pair. Committing is a git mutation and requires explicit user approval (OQ-07).

**Implements:** FR-001. **Evidence:** handoff section 2, CURRENTLY UNCOMMITTED / PROVENANCE CERT CHAIN.

### D-02 -- Signing configuration (Layer 2)

Repo-local git configuration: `gpg.format=ssh`, `user.signingkey=<public key path>`, `commit.gpgsign=true`, `tag.gpgsign=true`; an SSH allowed-signers file enables local `git log --show-signature` / `git tag -v` verification without any PKI or network dependency. Key generation and registration of the public key is the user action (OQ-01); SSH is the recommended scheme (handoff section 1, COMMIT SIGNING: simplest per the GitHub docs; GPG only if expiry/revocation semantics are wanted; gitsign rejected for badge/network/verify friction).

**Implements:** FR-004. **Evidence:** handoff section 1, COMMIT SIGNING; section 2, NO SIGNING (verified: `commit.gpgsign`/`user.signingkey`/`gpg.format` unset, `%G?` = N).

### D-03 -- Checkpoint script and tag format (Layer 2 -> Layer 5)

A new script `scripts/checkpoint-record.sh`: resolves `git rev-parse HEAD` and `git rev-parse 'HEAD^{tree}'`; creates an annotated, signed tag named `checkpoint-YYYY-MM` (suffix `-rN` on rerun within a month) whose message records the head commit SHA, tree SHA, UTC date, signer identity, and this specification's document ID; pushes the tag to `origin` **and** `github`. Pushing the tree hash alongside the commit SHA binds the checkpoint to content, not merely to lineage. Cadence is monthly per OQ-02. Caveat carried from Layer 5: the sync mirror propagates rewrites, so the second remote is redundancy; the evidentiary anchor is the signed tag itself, which a rewriter cannot backfill without the signer's key.

**Implements:** FR-005. **Evidence:** handoff section 1, CHECKPOINTS/ANCHORING; `.github/workflows/sync-from-gitea.yml:42,51`; remotes verified 2026-08-18.

### D-04 -- Optional OpenTimestamps stamping (Layer 5)

If enabled (OQ-03), the checkpoint script additionally runs the OTS client (open-source; opentimestamps.org, client v0.7.2 dated 2024-12-31) to stamp the checkpoint tag's digest, and the resulting `.ots` proof is committed under `provenance/ots/`. Proofs are Bitcoin-anchored and offline-verifiable, supplying the trusted time git cannot. OTS calendar liveness is UNVERIFIED; the mechanism is complete with this layer off, and turning it on later requires no change to any other layer.

**Implements:** FR-006. **Evidence:** handoff section 1, CHECKPOINTS/ANCHORING + GIT LIMITS (commit timestamps attacker-controlled).

### D-05 -- ADR set layout (Layer 3)

`docs/adr/NNNN-kebab-title.md` per decision, plus `docs/adr/README.md` as the index (number, title, status, supersedes / superseded-by). Nygard format: Title, Status (proposed / accepted / deprecated / superseded), Context, Decision, Consequences. Numbers are sequential and never reused. A reversed decision is kept: a new commit sets its Status to `superseded by ADR-NNNN`; its Context and Decision sections are never edited. ADR-0001 ("Adopt architecture decision records") is the first entry and records the adoption decision with a reference to the directive and this specification pair. Nygard is chosen over MADR 4.0.0 (heavier) and Y-statements (lighter); a 2026 arXiv study favors Nygard for concise records (handoff section 1, RATIONALE CAPTURE).

**Implements:** FR-007. **Evidence:** handoff section 1, RATIONALE CAPTURE; section 2, RATIONALE CONVENTIONS (no ADR directory today).

### D-06 -- Engineering journal format (Layer 3)

`docs/JOURNAL.md`, append-only: entries are appended at the end of the file; existing entries are never edited (corrections are new entries referencing the old). Entry format:

```text
## YYYY-MM-DD -- short title
- Author/context: who, and what session or task prompted the entry
- Tried: what was attempted
- Failed: what did not work, with the evidence observed
- Learned: what understanding changed
- Links: commit SHAs, ADR-NNNN, spec documents
```

The journal is seeded in Phase 0 with the project's origin entries (handoff section 3: "journal seeded with the project's origin entries"). It is the lab notebook: "the record made during the work, not a reconstruction."

**Implements:** FR-008. **Evidence:** handoff section 1, RATIONALE CAPTURE; section 2, RATIONALE CONVENTIONS (no journal today).

### D-07 -- Corpus cross-linking convention (Layer 3)

The training-log corpus is only consumable if its three components reference each other: ADRs cite the ADRs they supersede; journal entries cite the commits and ADRs they relate to; commit bodies (IN-M-003, already enforced) cite `ADR-NNNN` when the commit implements a recorded decision. All three are plain ASCII text committed in the repository and greppable without special tooling (NFR-004). `git notes` is explicitly not used for rationale: it is not pushed/fetched by default and is mutable without changing any commit hash (handoff section 1, RATIONALE CAPTURE).

**Implements:** FR-009. **Evidence:** handoff section 1, RATIONALE CAPTURE; `plugins/rr-policy-guards/tools/commit-guard/validator.go:66-70`.

### D-08 -- Commit-guard amend block (Layer 4)

The commit-guard already parses `--amend` into `inv.Amend` (`extract.go:90-91`; field defined at `types.go:73`) and never consults it, so amends pass today (verified 2026-08-18). The change point is a new decision rule: any parsed commit invocation with `Amend` set yields `DecisionBlock` with a new rule code (assigned in the Technical Specification), a reason string citing FR-002 and this specification, and an audit record in the existing append-only JSONL pattern. Existing staged-file (Never/Grey) and message (IN-M-001/002/003) rules are unchanged (NFR-007); new tests cover the block per NFR-003.

**Implements:** FR-010. **Evidence:** `plugins/rr-policy-guards/tools/commit-guard/extract.go:90-91`, `types.go:73`; handoff section 2, NO REWRITE PROTECTION.

### D-09 -- Pre-push force-push control (Layer 4)

A new `.git/hooks/pre-push` wrapper, installed by extending the existing installer `plugins/rr-policy-guards/scripts/install-git-hooks.sh` (today installs the `pre-commit` and `commit-msg` wrappers per `plugins/rr-policy-guards/README.md:98`; installed hooks verified at `.git/hooks/`), rejects any push that would force-update `main`. The pre-push hook receives the remote name and URL as arguments and the proposed ref updates on stdin as `<local ref> <local sha> <remote ref> <remote sha>` lines; the control parses those lines and blocks an update to `refs/heads/main` when the update deletes the ref, or when the remote-side value is not an ancestor of the new local value (non-fast-forward) -- the two shapes a history rewrite takes on the wire. Exact parsing belongs to the Technical Specification. Decisions and audit records follow the existing guard pattern (append-only JSONL under `~/.rational-reserve/logs/`, fail-open audit: enforcement never depends on audit writes). Verify-guard currently classifies `git push` generically (`shell.go:163-164`) with force-push undifferentiated; that gate is left as-is and the new hook carries the force-push decision. **Bypass posture:** the new logic has no `RR_*` bypass variable. AGENTS.md states mandatory guards have no bypass variables, yet five of the six guards carry a live, audited bypass variable (`RR_BASH_GUARD_BYPASS`, `RR_BREW_GUARD_BYPASS`, `RR_COMMIT_GUARD_BYPASS`, `RR_EMOJI_GUARD_BYPASS`, `RR_TOFU_GUARD_BYPASS`); verify-guard is the documented exception (`verify-guard/main.go:13`: "Verification has no bypass path", pinned by tests that set `RR_VERIFY_GUARD_BYPASS` and assert it does not bypass). This is a recorded discrepancy (handoff section 2, COMMIT-GUARD ENFORCES TODAY; extent re-verified 2026-08-18); this design follows the no-bypass rule for new logic and flags the existing variables' disposition to the guard layer's own specification rather than silently propagating either choice. Local hooks gate only this machine's pushes; server-side branch protection on gitea.com (OQ-06, current state UNVERIFIED) is the complementary machine-independent control.

**Implements:** FR-011. **Evidence:** handoff section 2, NO REWRITE PROTECTION / COMMIT-GUARD ENFORCES TODAY; `.git/hooks/` listing; `plugins/rr-policy-guards/tools/verify-guard/shell.go:163-164`; `plugins/rr-policy-guards/README.md:98`.

### D-10 -- Guard audit-log hash chaining (Layer 4, Phase 2 proposal)

The guard audit logs (`~/.rational-reserve/logs/<guard>-guard.jsonl`, mode 0600, append-only, fail-open) currently have no integrity protection: they are local-only and deletable by the same user they record (handoff section 2, GUARD AUDIT LOGS). The Phase 2 proposal adds a `prev` field to each JSONL entry carrying the SHA-256 of the previous entry (linear chain; sufficient for a single sequential writer per Haber-Stornetta 1991), plus an offline verifier script that re-walks a log and reports breaks. Chaining makes retroactive log edits evident, not impossible; the fail-open contract is preserved (chain computation never gates an enforcement decision). Scope -- which guards, and whether the batch lands in Phase 2 -- is OQ-04; the recommendation is commit-guard and verify-guard first, the two that gate publication.

**Implements:** FR-012. **Evidence:** handoff section 2, GUARD AUDIT LOGS; handoff section 1, LINEAR CHAIN SUFFICES.

---

## 4. The exception model: one coherent rule

The mechanism's discipline reduces to a single rule applied uniformly:

**History is never rewritten; corrections are new records that reference the records they correct.**

- **Commits:** a mistaken commit is corrected by a new commit that references the mistaken one's SHA (FR-003). There is no legitimate rewrite procedure -- no amend, no rebase of published history, no force-push, no history-changing reset (FR-002).
- **ADRs:** a reversed decision is never edited or deleted; a new ADR supersedes it, and a new commit marks the old one `superseded by ADR-NNNN` (D-05). The supersession chain is itself part of the training corpus -- it is exactly "why the mindframe shifted due to a new understanding."
- **Journal:** entries are never edited; a correction is a new entry referencing the old (D-06).
- **Provenance certificates:** already operate this way -- certificates are re-issued with new digests, and superseded certificates stay in git history (`provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md`, supersede rule; handoff section 2, PROVENANCE CERT CHAIN). The mechanism adopts the existing pattern rather than inventing a parallel one.

The rule has no escape hatch. Any future pressure to rewrite history is itself recorded -- as a journal entry and, if it ever prevailed, as an ADR -- because hiding the pressure would defeat the mechanism's stated purpose. The two in-repo rewrite scripts (`scripts/git-fix-email.sh:9-10`, `scripts/git-fix-author.sh:9-10`: amend + `push --force`) conflict with this rule; their disposition is OQ-05 (recommended: delete in a new commit -- the deletion is itself an immutable record, and the scripts remain in history).

---

## 5. Phasing

### Phase 0 -- Baseline (needs user commit approval)

- **Scope:** D-01 (baseline commits); D-05 partially (ADR-0001 + `docs/adr/` index); D-06 partially (`docs/JOURNAL.md` seeded with origin entries).
- **Entry criteria:** user approves the Requirements and this Design Specification; OQ-07 answered (commit structure and who commits).
- **Exit criteria:** `git status` clean for the documents that matter; ADR-0001 present and accepted; journal seeded; the retained provenance-certificate chain started.
- **Dependencies:** none outside the repository. Every commit requires explicit user approval.

### Phase 1 -- Enforcement and signing

- **Scope:** D-02 (signing configuration, after the OQ-01 key action); D-08 (commit-guard amend block); D-09 (pre-push force-push control); OQ-05 disposition of the rewrite scripts.
- **Entry criteria:** Phase 0 exit met; OQ-01 answered and the signing key generated and registered (user action).
- **Exit criteria:** `--amend` blocked by the guard with an audit record; force-push to `main` blocked pre-push with an audit record; new commits signed; guard Go test suites and `opa test --v0-compatible policies/*.rego` green (NFR-007).
- **Dependencies:** Phase 0; the guard framework and hook installer that already exist.

### Phase 2 -- Anchoring and automation

- **Scope:** D-03 (checkpoint script + first signed checkpoint tag pushed to both remotes); D-04 (optional OTS, if OQ-03 enables it); D-10 (guard audit-log hash chaining, scope per OQ-04); OQ-06 server-side branch protection (user action on gitea.com).
- **Entry criteria:** Phase 1 exit met; OQ-02 (cadence), OQ-03 (OTS), OQ-04 (chaining scope) answered.
- **Exit criteria:** `checkpoint-YYYY-MM` signed tag present on `origin` and `github`; chaining verifier passes on the in-scope guard logs (if adopted); branch-protection state on gitea.com verified and recorded.
- **Dependencies:** Phase 1; user access to gitea.com repository settings.

---

## 6. Traceability

Every design element cites the requirement it implements and the input evidence behind it (handoff section 1 = verified web research, sources named there; handoff section 2 = verified repo map).

| Design element | Implements | Evidence |
|---|---|---|
| D-01 Baseline commit procedure | FR-001 | Handoff s2: CURRENTLY UNCOMMITTED, PROVENANCE CERT CHAIN, STALE DOCS NOTED; `git status` 2026-08-18 |
| D-02 Signing configuration | FR-004 | Handoff s1: COMMIT SIGNING (GitHub docs; git >= 2.34); s2: NO SIGNING (spot-verified) |
| D-03 Checkpoint script and tag format | FR-005 | Handoff s1: CHECKPOINTS/ANCHORING; `.github/workflows/sync-from-gitea.yml:42,51`; remotes verified |
| D-04 Optional OTS stamping | FR-006 | Handoff s1: CHECKPOINTS/ANCHORING (opentimestamps.org, client v0.7.2), GIT LIMITS (attacker-controlled dates); calendar liveness UNVERIFIED |
| D-05 ADR set layout | FR-007 | Handoff s1: RATIONALE CAPTURE (Nygard 2011, cognitect.com blog; 2026 arXiv study); s2: RATIONALE CONVENTIONS |
| D-06 Journal format | FR-008 | Handoff s1: RATIONALE CAPTURE (lab notebook); s2: RATIONALE CONVENTIONS; s3 (seed with origin entries) |
| D-07 Corpus cross-linking | FR-009 | Handoff s1: RATIONALE CAPTURE (git notes never system of record); `validator.go:66-70` (IN-M-003) |
| D-08 Commit-guard amend block | FR-010 | Handoff s2: NO REWRITE PROTECTION; `extract.go:90-91`, `types.go:73` (verified) |
| D-09 Pre-push force-push control | FR-011 | Handoff s1: PROCESS GLUE; s2: NO REWRITE PROTECTION, COMMIT-GUARD ENFORCES TODAY (bypass discrepancy); `shell.go:163-164`; `.git/hooks/` listing; `plugins/rr-policy-guards/README.md:98` |
| D-10 Audit-log hash chaining | FR-012 | Handoff s2: GUARD AUDIT LOGS; s1: LINEAR CHAIN SUFFICES (Haber-Stornetta 1991) |
| Section 4 exception model | FR-003 | Handoff s1: RATIONALE CAPTURE (superseded-not-deleted); s2: PROVENANCE CERT CHAIN; `scripts/git-fix-email.sh:9-10`, `scripts/git-fix-author.sh:9-10` |
| Layer 1 record policy (s2.1) | FR-002 | Handoff s1: PRIMARY RECORD (github.com/vladi160/preregistrations README), GIT LIMITS |
| Considered-and-rejected set (REQ s3.2) | NFR-002, NFR-006 | Handoff s1: LINEAR CHAIN SUFFICES; CHECKPOINTS/ANCHORING (Sigstore 2026-06-28 "Rekor evolution"); GIT LIMITS (Gitea sha256 UNVERIFIED) |

---

**End of Design Specification**

The Technical Specification (exact guard rule code and decision wiring, pre-push hook code, `checkpoint-record.sh`, ADR/journal templates, git signing configuration, verifier script) completes the triad.
