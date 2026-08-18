# Technical Specification: Record Immutability for Project Documents

**Document ID:** RECORD-IMMUTABILITY-TECH-001
**Version:** 0.1
**Date:** 2026-08-18
**Status:** Draft for review -- nothing in this document is approved or decided yet
**Predecessors:** RECORD-IMMUTABILITY-REQ-001 (`2026-08-18-Record-Immutability-Requirements.md`), RECORD-IMMUTABILITY-DES-001 (`2026-08-18-Record-Immutability-Design-Specification.md`)

---

## 1. Purpose, scope, and traceability

This document is the implementation-grade companion to the Requirements (REQ-001) and Design Specification (DES-001). It specifies exact change points, file contents, commands, and tests for the record-immutability mechanism. It changes no code by itself; implementation follows the rollout checklist in section 11.

**Decision neutrality.** Open questions OQ-01 through OQ-07 (REQ-001 section 7) remain open, and this document introduces one more: **OQ-08** (emergency rewrite procedure, section 8.2) -- a PROPOSED amendment to REQ-001, NOT APPROVED, excluded from the rollout checklist until the user decides; to be adopted into REQ-001 section 7 when that document next revises. Where a pending decision changes the implementation, both branches are specified (notably signing: SSH and GPG in section 5) and the gate is named. Nothing here pre-decides an OQ.

**Evidence discipline.** Every code-level claim below was verified against the live source on 2026-08-18 and carries a `file:line` citation; section 12 lists the commands that re-check them. Verified environment facts used below: git version 2.55.0 (SSH signing requires >= 2.34); guard code at `plugins/rr-policy-guards/tools/commit-guard/`; installed hooks at `.git/hooks/`.

**Traceability map (section -> implements):**

| Section | Implements (FR) | Design element |
|---|---|---|
| 2. Guard change point A (amend block) | FR-010 | D-08 |
| 3. Guard change point B (pre-push block) | FR-011 | D-09 |
| 4. Checkpoint script | FR-005 | D-03 |
| 5. Signing configuration | FR-004 | D-02 |
| 6. ADR system | FR-007 | D-05 |
| 7. Engineering journal | FR-008 | D-06 |
| 8. Exception process | FR-003 as written: 8.1 corrections-as-new-commits is the approved procedure; 8.2 is a PROPOSED amendment (OQ-08), NOT APPROVED, excluded from rollout | DES-001 section 4 |
| 9. Guard-log hash chaining sketch | FR-012 | D-10 |
| 10. Test plan | NFR-003, NFR-007 | all |
| 11. Rollout checklist | FR-001 (Phase 0), phases per DES-001 section 5 | D-01 |

---

## 2. Guard change point A -- commit-guard amend block (FR-010 / D-08)

### 2.1 Where the flag exists today (verified)

- `plugins/rr-policy-guards/tools/commit-guard/extract.go:90-91` parses `--amend` into `inv.Amend`; the field is defined at `types.go:73` (`Amend bool` in `CommitInvocation`, types.go:69-76).
- `inv.Amend` is never consulted anywhere outside the parser and its test (`extract_test.go:63-68`); amends pass the guard today.
- `ExtractCommit` only ever sees a Bash command string in PreToolUse mode (`main.go:74-81`). `--amend` therefore surfaces **only in PreToolUse mode**; the `--scan-staged` and `--validate-msg` modes (main.go:46-49) never parse an invocation and cannot observe an amend (see 2.5).

### 2.2 Rule ID (following the existing naming convention)

The guard's rule-code taxonomy, quoted from source:

- Path rules, block severity (`scanner.go:19-57`): `NV-S-001..003` (security), `NV-R-001..002` (reproducibility), `NV-P-001` (provenance), `NV-N-001..003` (signal), `NV-X-001` (reversibility).
- Path rules, warn severity (`scanner.go:59-66`): `GR-P-001`, `GR-N-001`, `GR-Y-001`.
- Message rules (`validator.go`, header at `validator.go:1`: "commit message validation (IN-M-* rules)"): `IN-M-001` (empty subject, validator.go:38-39 and 44-45; subject > 72 chars, validator.go:55-58), `IN-M-002` (Conventional Commits form, validator.go:60-64), `IN-M-003` (missing WHY body, validator.go:66-70), `IN-M-004` (merge/revert exemption, named in comments at validator.go:28 and 48-50).

The convention is `<FAMILY>-<PRINCIPLE>-<SEQ>`. The amend rule is an **invocation-level** rule, not a message rule (validator.go:1 scopes `IN-M-*` to message validation) and not a staged-path rule. This specification therefore assigns a new intent family:

- **`IN-H-001`** -- history rewrite via `git commit --amend` (this section).
- **`IN-H-002`** -- history rewrite via force-update or deletion of `refs/heads/main` (section 3).

Fallback, if the maintainer prefers to keep a single `IN-*` intent family: `IN-M-005` / `IN-M-006`. This document recommends `IN-H-*` because it keeps the family-per-domain pattern the scanner and validator already use.

### 2.3 Exact change point

File: `plugins/rr-policy-guards/tools/commit-guard/main.go`, function `runPreToolUse` (main.go:57-149). Insert immediately after the `IsCommit` gate (main.go:78-81) and **before** step "1. Scan staged paths" (main.go:83-94) and before the bypass check (main.go:111-118). Placement before the bypass check is what makes the rule bypass-free.

```go
	// 0. Record immutability (FR-002/FR-010): --amend rewrites published
	// history. No bypass: this check precedes bypassActive() on purpose.
	if inv.Amend {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePreToolUse.String(), Rule: "IN-H-001",
			Session: input.SessionID, Command: command})
		fmt.Fprintln(stderr, "rr-commit-guard: BLOCKED -- commit invocation fails record-immutability rules:")
		fmt.Fprintln(stderr, "  [IN-H-001] --amend rewrites published history; corrections must be new commits referencing the mistaken commit (RECORD-IMMUTABILITY-REQ-001 FR-002, FR-010)")
		return exitBlock
	}
```

Notes on the contract this preserves:

- Exit code `exitBlock` = 2 (types.go:79-83), matching the existing contract documented at main.go:10.
- **`hintBypass` is NOT called** (contrast main.go:125 and 134, which call it after staged-file and message blocks; the helper at main.go:294-297 prints the `RR_COMMIT_GUARD_BYPASS=1` hint). The new rule has no bypass variable, per the no-bypass rule for new logic (DES-001 D-09, REQ-001 section 4: five legacy bypasses exist -- `RR_BASH_/RR_BREW_/RR_COMMIT_/RR_EMOJI_/RR_TOFU_GUARD_BYPASS` -- and contradict AGENTS.md; verify-guard is the documented exception, `verify-guard/main.go:13` "Verification has no bypass path").
- Placement ahead of `StagedPaths` (main.go:84) also means the block fires without touching the filesystem; the unit test needs no scratch repo (section 10).

### 2.4 Audit event shape

The existing `AuditRecord` JSON shape (types.go:57-66) is used unchanged; `appendAudit` (audit.go:28-30) fails open and auto-fills `ts` as UTC RFC3339Nano (audit.go:32-35). An amend block emits exactly one line:

```json
{"ts":"2026-08-18T00:00:00.000000000Z","decision":"block","mode":"pretooluse","rule":"IN-H-001","session":"<session-id>","command":"git commit --amend --no-edit"}
```

Field-for-field this matches the existing block records emitted at main.go:121-123 (staged-file block) and main.go:130-132 (message block): `decision`, `mode`, `rule`, `session`, `command`; `paths`/`subject` omitted as empty per the `omitempty` tags (types.go:61-65).

### 2.5 Behavior in all three modes

| Mode (main.go:46-52) | Behavior after change | Why |
|---|---|---|
| PreToolUse (stdin JSON, main.go:57-149) | `--amend` in the parsed Bash command blocks with `IN-H-001` | Only mode that parses a Bash invocation (main.go:74-81) |
| `--scan-staged` (main.go:153-189) | Unchanged | Staged files of an amend are indistinguishable from a normal commit's; nothing to key on |
| `--validate-msg` (main.go:193-223) | Unchanged | An amended commit's message file is still just a message |

Honest consequence, stated so the spec does not overclaim (NFR-006): the amend block only covers agent-context commits. A raw terminal `git commit --amend` never passes through PreToolUse. That residual path is exactly what change point B (section 3) and server-side branch protection (OQ-06) exist to catch: an amend that cannot be force-pushed is a local-only rewrite.

### 2.6 No-bypass statement

The new rule consults no environment variable. The legacy `RR_COMMIT_GUARD_BYPASS=1` (`bypass.go:9`, checked at main.go:111) is evaluated only **after** the new check and therefore cannot waive it. This follows the AGENTS.md no-bypass rule for mandatory guards, recorded against the five-guard legacy discrepancy in REQ-001 section 4.

---

## 3. Guard change point B -- pre-push force-push block (FR-011 / D-09)

### 3.1 Architecture decision: a fourth mode in rr-commit-guard

Two options were weighed: (a) a fourth mode `--pre-push` in `rr-commit-guard`; (b) a small sibling binary. **Decision: (a), a fourth mode.** Reasoning, grounded in the code:

1. `run()` already dispatches on flags (main.go:38-53); a new mode is one `case` line: `case hasFlag(args, "--pre-push"): return runPrePush(stdin, stderr)`.
2. The mode enum is a three-line extension (`types.go:10-14` consts, `types.go:16-27` `String()`), and the audit shape (types.go:57-66) and writer (audit.go) carry over unchanged.
3. The binary already shells out to git with the patterns the ancestry check needs: `exec.Command("git", args...)` with `cmd.Dir` (`git.go:15-34`, StagedPaths) and with `-C` arguments (`git.go:38-53`, StagedSize).
4. The hook wrappers and installer are already built around this one binary, including the `RR_COMMIT_GUARD_BIN` override (installed `.git/hooks/pre-commit` and `.git/hooks/commit-msg`; `plugins/rr-policy-guards/scripts/install-git-hooks.sh`).
5. The policy domain is identical -- the binary's own header (main.go:1) is "mechanical enforcement of commit-discipline principles"; push discipline is the same domain.
6. A sibling tool would duplicate the audit writer, the exit-code contract (types.go:79-83), binary discovery, and installer logic for ~150 lines of parser.

### 3.2 What the pre-push hook provides (githooks(5) contract)

Arguments: `$1` = remote name, `$2` = remote URL. **No flags**: the hook never sees `--force` (DES-001 D-09 correction pass). Ref updates arrive on stdin, one per line:

```text
<local ref> SP <local sha> SP <remote ref> SP <remote sha> LF
```

Zero SHA (40 zeros, SHA-1 object format; REQ-001 section 3.2 keeps SHA-1) means "ref does not exist on that side": zero local sha = deletion; zero remote sha = creation. Refs cannot contain spaces, so splitting each line on single spaces into exactly 4 fields is safe.

### 3.3 Parser and decision logic (sketch; new file `prepush.go`)

```go
// prepush.go -- record-immutability pre-push gate (FR-011 / D-09).
package main

const zeroSHA = "0000000000000000000000000000000000000000"

// refUpdate is one githooks(5) pre-push stdin line.
type refUpdate struct {
	localRef, localSHA, remoteRef, remoteSHA string
}

// parseRefUpdates parses stdin lines "<local ref> <local sha> <remote ref>
// <remote sha>". A trailing empty line is ignored; any other malformed line
// is an error (caller maps it to exitInternal per the exit-code contract).
func parseRefUpdates(raw string) ([]refUpdate, error)

// isAncestor reports whether old is an ancestor of new, via
// `git merge-base --is-ancestor old new`. Exit 0 = true. Exit 1 (not an
// ancestor) and any higher exit (error) both yield false: fail-closed on
// the protected ref, because an unverifiable ancestry check must not pass.
// The remote sha is the remote's advertised value learned during this
// push's negotiation (githooks(5)); in the common case it equals the
// remote-tracking ref and its object is local, but in a lost-race window
// (another push landed, was advertised, and was never fetched here) the
// object may not exist locally. merge-base then errors, and the
// fail-closed mapping above turns that window into a safe block: the
// operator fetches and retries, and a genuine fast-forward passes.
func isAncestor(old, new string) bool

// rewritesMain reports whether one update rewrites the protected ref, with
// the reason string for the block message.
func rewritesMain(u refUpdate) (bool, string) {
	if u.remoteRef != "refs/heads/main" {
		return false, ""
	}
	if u.localSHA == zeroSHA {
		return true, "deletion of refs/heads/main"
	}
	if u.remoteSHA == zeroSHA {
		return false, "" // first push of main to this remote: creation, not rewrite
	}
	if !isAncestor(u.remoteSHA, u.localSHA) {
		return true, "non-fast-forward update of refs/heads/main"
	}
	return false, ""
}
```

Zero-SHA edge cases, specified explicitly:

| local sha | remote sha | Meaning | Decision |
|---|---|---|---|
| non-zero | non-zero, ancestor of local | fast-forward update of main | allow |
| non-zero | non-zero, NOT ancestor of local | non-fast-forward (rewrite) | **block IN-H-002** |
| zero | non-zero | deletion of main | **block IN-H-002** |
| non-zero | zero | main does not exist on this remote yet (creation) | allow |
| (any) | (any), remoteRef != refs/heads/main | other branches, tags | allow (policy scope is main per FR-002; widening is a future decision) |

`runPrePush(stdin io.Reader, stderr io.Writer) int` flow: read all of stdin; parse (malformed -> audit + `exitInternal` = 1, matching the guard's exit-1-on-bad-input precedents: stdin read failure at main.go:58-61, staged-path enumeration failure at main.go:155-158, missing or unreadable message file at main.go:194-202; deliberately NOT main.go:63-69, where an unparseable PreToolUse JSON payload returns exitBlock -- a different failure, because there the guard cannot see what it is gating; the distinction is moot for a hook anyway, since githooks(5) aborts the push on ANY non-zero exit); evaluate every update; if any blocks, emit the block below and return `exitBlock` = 2; otherwise audit allow with `Mode: "pre-push"` and return `exitAllow` = 0. No `bypassActive()` call anywhere in this path.

Block message (stderr, verbatim):

```text
rr-commit-guard: BLOCKED -- push fails record-immutability rules:
  [IN-H-002] non-fast-forward update of refs/heads/main -- history rewrite prohibited (RECORD-IMMUTABILITY-REQ-001 FR-002, FR-011)
```

(or the deletion variant of the reason line). Audit record, one per invocation:

```json
{"ts":"...","decision":"block","mode":"pre-push","rule":"IN-H-002","command":"push origin https://gitea.com/trademomentum.net/developer-portal.git"}
```

(`command` assembled from argv `$1`/`$2`; `session` is empty in hook mode.)

### 3.4 Hook wrapper (`plugins/rr-policy-guards/git-hooks/pre-push`, new file)

Mirrors the installed wrappers at `.git/hooks/pre-commit` and `.git/hooks/commit-msg` verbatim in structure (shebang, `set -eu`, `RR_COMMIT_GUARD_BIN` override, `command -v` guard with heredoc error, `exec`). stdin flows through to the binary unmodified.

```sh
#!/bin/sh
# rr-commit-guard pre-push hook.
#
# Delegates to the rr-commit-guard binary in --pre-push mode. The binary
# reads the proposed ref updates from stdin and blocks any non-fast-forward
# update or deletion of refs/heads/main (record immutability, FR-011).
# Exits 0 on allow, 2 on block, 1 on internal error.
#
# Override binary path: set RR_COMMIT_GUARD_BIN=/abs/path/to/rr-commit-guard.

set -eu

GUARD="${RR_COMMIT_GUARD_BIN:-rr-commit-guard}"

if ! command -v "${GUARD}" >/dev/null 2>&1; then
  cat >&2 <<EOF
rr-commit-guard: binary not found on PATH (${GUARD}).
Install or set RR_COMMIT_GUARD_BIN to the absolute path.
EOF
  exit 1
fi

exec "${GUARD}" --pre-push "$1" "$2"
```

Noted drift (verified 2026-08-18): the installed `.git/hooks/pre-commit` and `.git/hooks/commit-msg` differ from the sources in `plugins/rr-policy-guards/git-hooks/` by three bypass-comment lines present only in the sources (source line 8 `# Bypass: set RR_COMMIT_GUARD_BYPASS=1 ...` and source lines 19-21 `To skip this check for one commit: ...`). Re-running the installer syncs them; the new `pre-push` wrapper intentionally carries no bypass comment because the new mode has no bypass.

### 3.5 Installer change

File: `plugins/rr-policy-guards/scripts/install-git-hooks.sh`. One-line change at line 42:

```sh
for hook in pre-commit commit-msg pre-push; do
```

Everything else is reused: the non-rr-commit-guard backup behavior (install-git-hooks.sh:45-49, moves a foreign hook to `.bak.<timestamp>`), `install -m 0755` (line 50), and the `--git-path hooks` resolution (line 26). The usage comment (lines 10-12) and the README contract (`plugins/rr-policy-guards/README.md:98`, "installs `pre-commit` and `commit-msg` hooks") must be updated to name all three hooks.

### 3.6 Residual gaps (honesty, NFR-006)

- Local hooks gate only this machine's pushes; another machine, a fresh clone, or a web-UI operation bypasses them. Server-side branch protection on gitea.com (OQ-06, user action, current state UNVERIFIED) is the machine-independent control.
- A forced push that happens to be fast-forward is not a rewrite and is correctly allowed.
- Other branches are unprotected (FR-002 scopes the policy to `main`).

---

## 4. Checkpoint script (FR-005 / D-03)

New file: `scripts/checkpoint-immutability.sh`. Full sketch:

```sh
#!/bin/sh
# checkpoint-immutability.sh -- monthly signed checkpoint tag (FR-005 / D-03).
#
# Creates an annotated, signed tag checkpoint-YYYY-MM binding the signer to
# the exact head commit and tree of the current HEAD, chaining to the
# previous checkpoint, and pushes it to origin AND github.
#
# The script REFUSES to create an unsigned tag (NFR-006): an unsigned
# checkpoint claims an anchor it does not provide. Tag signing requires
# OQ-01 resolved (signing key generated and git configured, section 5);
# until then this script exits 1 without creating anything.
#
# Usage:  scripts/checkpoint-immutability.sh
# Dry run (no tag, no push; testability without mutation):
#         RECORD_CHECKPOINT_DRY_RUN=1 scripts/checkpoint-immutability.sh

set -euo pipefail

PREFIX="checkpoint"
BASE="${PREFIX}-$(date -u +%Y-%m)"

HEAD="$(git rev-parse HEAD)"
TREE="$(git rev-parse 'HEAD^{tree}')"
DATE_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
# Previous checkpoint for the chain. Sort key verified empirically on git
# 2.55.0: -version:refname compares digit runs numerically, so with base,
# -r2, and -r10 tags present it orders r10 > r2 > base > earlier zero-padded
# YYYY-MM -- head -n 1 is always the latest checkpoint. (Plain -refname
# would misrank r10 below r2 lexicographically; -creatordate ties when the
# tags share one commit.)
PREV="$(git tag -l "${PREFIX}-*" --sort=-version:refname | head -n 1 || true)"

# Append-only naming: never move an existing tag. A rerun inside the same
# month takes the next free -rN suffix (DES-001 D-03).
TAG="${BASE}"
N=2
while git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null 2>&1; do
  TAG="${BASE}-r${N}"
  N=$((N + 1))
done

# Signing is mandatory. git tag -s would also fail without a usable key,
# but fail here with an explicit reason instead of a gpg/ssh error.
if ! git config --get user.signingKey >/dev/null 2>&1; then
  echo "checkpoint: user.signingKey is not set (OQ-01 unresolved)." >&2
  echo "checkpoint: refusing to create an unsigned tag (NFR-006)." >&2
  exit 1
fi

MSG_FILE="$(mktemp "${TMPDIR:-/tmp}/checkpoint-msg.XXXXXX")"
trap 'rm -f "${MSG_FILE}"' EXIT
{
  echo "record checkpoint ${TAG}"
  echo
  echo "head:     ${HEAD}"
  echo "tree:     ${TREE}"
  echo "date-utc: ${DATE_UTC}"
  echo "prev:     ${PREV:-none}"
  echo "spec:     RECORD-IMMUTABILITY-TECH-001 (FR-005 / D-03)"
} > "${MSG_FILE}"

if [ "${RECORD_CHECKPOINT_DRY_RUN:-0}" = "1" ]; then
  echo "checkpoint: DRY RUN -- would create signed tag ${TAG} with message:"
  cat "${MSG_FILE}"
  exit 0
fi

git tag -s -F "${MSG_FILE}" "${TAG}" "${HEAD}"

# Verify before publishing: an unverifiable tag is not an anchor.
# SSH signatures require gpg.ssh.allowedSignersFile, otherwise
# git verify-tag fails with trust level undefined (man git-config).
git tag -v "${TAG}"

git push origin "refs/tags/${TAG}"
git push github "refs/tags/${TAG}"

echo "checkpoint: ${TAG} signed, verified, and pushed to origin and github."
```

Design notes:

- **Linear chain of checkpoints.** Each tag message carries `prev:` naming the previous checkpoint tag (or `none` for the first). Tag names + signed messages form a Haber-Stornetta-style linear chain (REQ-001 section 3.2): a missing or re-created checkpoint is visible in the chain.
- **`tree:` alongside `head:`.** `git rev-parse 'HEAD^{tree}'` binds the checkpoint to content, not merely lineage (DES-001 D-03).
- **Unsigned fallback is rejected.** If `user.signingKey` is unset the script exits 1 before any tag exists; there is no `--unsigned` flag and no env override. The dry-run variable is a testing affordance only: it creates nothing and pushes nothing.
- **Verify-then-push.** `git tag -v` runs before any push; under `set -euo pipefail` a failed verification aborts the script before anything leaves the machine. For SSH signatures, verification requires `gpg.ssh.allowedSignersFile` to be configured -- per `man git-config` (git 2.55.0), without an allowed-signers file the trust level is undefined and `git verify-commit`/`git verify-tag` fail (section 5 configures it).

Post-run verification commands:

```sh
git tag -v checkpoint-YYYY-MM                 # signature valid (needs allowed signers for SSH)
git log --show-signature -1                   # head commit signature status
git ls-remote origin 'refs/tags/checkpoint-*' # tag present on primary forge
git ls-remote github 'refs/tags/checkpoint-*' # tag present on second remote
```

Cadence (OQ-02, monthly recommended) and any automation of it (manual calendar reminder vs launchd/cron entry) are decided in Phase 2; both options are decision-neutral here. Honesty note carried from DES-001 2.5: the github mirror sync (`.github/workflows/sync-from-gitea.yml:42,51`) prunes deletions via `push --mirror`, so the second remote is redundancy, not evidence; the anchor's value is the signed tag held by every clone.

---

## 5. Signing configuration (FR-004 / D-02)

Decision-neutral: both OQ-01 branches are specified. Key generation and registration are **user actions**. All commands are repo-local (run inside the repository, no `--global`) so no other repository is affected. Config key names verified against `man git-config` on git 2.55.0; git config keys are case-insensitive, canonical casing shown.

### 5.1 SSH branch (recommended per REQ-001 OQ-01)

User actions (once, outside the repo):

```sh
ssh-keygen -t ed25519 -C "record-immutability signing key" -f ~/.ssh/id_ed25519_signing
ssh-add ~/.ssh/id_ed25519_signing
# (load the key into ssh-agent; start one first if needed: eval "$(ssh-agent -s)")
# Register the PUBLIC key where verification matters (e.g. GitHub account
# signing keys, gitea.com account settings) -- user action.
```

Repo-local configuration:

```sh
git config gpg.format ssh
git config user.signingKey ~/.ssh/id_ed25519_signing.pub
git config commit.gpgSign true
git config tag.gpgSign true
git config gpg.ssh.allowedSignersFile ~/.ssh/allowed_signers
```

Notes grounded in `man git-config` (git 2.55.0):

- `gpg.format`: "openpgp" default; "x509" and "ssh" are the other values.
- `user.signingKey` (canonical casing per man git-config; git config keys are case-insensitive) with `gpg.format ssh`: per the manual the value may be the private key, or the public key **when ssh-agent is used** -- which is why the `ssh-add` step above is a required user action; without the agent the section 5.3 scratch commit cannot find the key. This spec uses the public-key path so the config never names private material.
- `gpg.ssh.allowedSignersFile`: without it, `git tag -v` / `git log --show-signature` fail with trust undefined. File line format: `<principal> ssh-ed25519 AAAA... <comment>`; the principal should be the signer's identity string (e.g. the email used in commits).

### 5.2 GPG branch (alternative, if expiry/revocation semantics are wanted)

User actions (once):

```sh
gpg --full-generate-key                       # choose expiry per policy -- user action
gpg --list-secret-keys --keyid-format=long    # note the KEYID -- user action
# Register/upload the public key where verification matters -- user action.
```

Repo-local configuration:

```sh
git config user.signingKey <KEYID>
git config commit.gpgSign true
git config tag.gpgSign true
# gpg.format stays unset: "openpgp" is the default (man git-config).
```

### 5.3 Verification (either branch)

```sh
git config --get-regexp '^(gpg|user\.signingKey|commit\.gpgSign|tag\.gpgSign)' # keys present (matching is case-insensitive)
git commit --allow-empty -m "chore: verify signing"     # scratch commit, then:
git log --format='%G?' -1                               # expect G (good signature)
git log --show-signature -1                             # human-readable verification
```

(`%G?` = N today, REQ-001 section 4; after Phase 1 it must read G on new commits.)

---

## 6. ADR system (FR-007 / D-05)

### 6.1 Layout

```text
docs/adr/
  README.md                               index (format in 6.3)
  0001-record-architecture-decisions.md   the canonical first ADR (6.4)
  NNNN-kebab-title.md                     one file per decision
```

Numbers are four digits, sequential, never reused. Filenames are `NNNN-kebab-title.md`.

### 6.2 ADR template (Nygard format: Title, Status, Context, Decision, Consequences)

```markdown
# ADR-NNNN: <short imperative title>

**Status:** proposed | accepted | deprecated | superseded by ADR-MMMM
**Date:** YYYY-MM-DD
**Supersedes:** ADR-KKKK | none
**Superseded by:** ADR-MMMM | none

## Context

<The forces at play: technical, policy, environmental. What makes this
decision necessary now. Include evidence links.>

## Decision

<The decision itself, stated in full sentences, active voice.>

## Consequences

<What becomes easier or harder; follow-up actions; what would reverse this.>
```

### 6.3 Index format (`docs/adr/README.md`)

```markdown
# Architecture Decision Records

Nygard-format ADRs for developer-portal (RECORD-IMMUTABILITY-REQ-001 FR-007).
Numbers are never reused; superseded ADRs are kept, never deleted.

| ADR | Title | Status | Date | Supersedes | Superseded by |
|---|---|---|---|---|---|
| [ADR-0001](0001-record-architecture-decisions.md) | Record architecture decisions | accepted | 2026-08-18 | -- | -- |
```

The index row is updated in the same commit as any ADR addition or status change.

### 6.4 ADR-0001 -- full draft content (status: accepted)

```markdown
# ADR-0001: Record architecture decisions

**Status:** accepted
**Date:** 2026-08-18
**Supersedes:** none
**Superseded by:** none

## Context

The governing directive of 2026-08-18 requires all project documents to
capture all changes from the beginning -- both to make mistakes harder to
hide and to produce a training log of "the mindframe and process that was
taken and why it may have shifted due to a new understanding"
(RECORD-IMMUTABILITY-REQ-001 section 2). Until now the repository has had
no decision record: rationale lived only in commit bodies (IN-M-003, the
one machine-enforced rationale capture) and in session-state documents
(TODO.md, SESSION_HANDOFF.md). Research condensed in the record-immutability
triad (Nygard 2011; a 2026 arXiv study favoring Nygard for concise records;
MADR 4.0.0 heavier, Y-statements lighter) identifies lightweight
Architecture Decision Records as the decision half of the rationale layer.

## Decision

Adopt Nygard-format Architecture Decision Records in docs/adr/. Every
significant architectural decision gets one file
docs/adr/NNNN-kebab-title.md with the sections Title, Status, Context,
Decision, Consequences. Numbers are sequential and never reused. A reversed
decision is never edited or deleted: a new ADR supersedes it, and a new
commit marks the old ADR's Status line "superseded by ADR-MMMM".
docs/adr/README.md is the index and is updated in the same commit as any
ADR change. This ADR is the first entry.

## Consequences

- Every significant decision gains a durable, greppable home; the
  supersession chain is itself part of the training-log corpus (FR-009).
- Commit bodies (IN-M-003) cite ADR-NNNN when a commit implements a
  recorded decision (DES-001 D-07).
- One more artifact to maintain per decision; accepted deliberately.
- Journal entries (docs/JOURNAL.md) link to ADRs where a decision emerges
  from recorded trial and failure.
```

### 6.5 Supersession procedure (exact edits)

When ADR-NNNN is reversed, ONE new commit performs exactly these edits and no others:

1. In the old file `docs/adr/NNNN-....md`: change only the `**Status:**` line to `superseded by ADR-MMMM` and set the `**Superseded by:**` metadata field to `ADR-MMMM`. **Context, Decision, and Consequences sections are never altered.**
2. Create the new file `docs/adr/MMMM-....md` from the template with `**Supersedes:** ADR-NNNN`.
3. Update `docs/adr/README.md`: the old row's Status/Superseded-by cells, and a new row for ADR-MMMM.

Commit message cites both numbers, e.g. `docs(adr): supersede ADR-NNNN with ADR-MMMM` plus a WHY body (IN-M-003).

---

## 7. Engineering journal (FR-008 / D-06)

### 7.1 File header (`docs/JOURNAL.md`, created once)

```markdown
# Engineering Journal -- developer-portal

Append-only contemporaneous log of what was tried, what failed, and what
was learned (RECORD-IMMUTABILITY-REQ-001 FR-008; DES-001 D-06).

Rules: entries are appended at the end of this file. Existing entries are
never edited -- corrections are new entries referencing the old. Entries
marked [seed] are retrospective reconstructions written once at adoption
time; every entry after the seed block is contemporaneous, written during
the work, not reconstructed after it.

---
```

### 7.2 Entry template (verbatim)

```markdown
## YYYY-MM-DD -- <short title>

- Author/context: <who, and what session or task prompted the entry>
- Tried: <what was attempted>
- Failed: <what did not work, with the evidence observed>
- Learned: <what understanding changed>
- Links: <commit SHAs, ADR-NNNN, spec documents>
```

Seed entries use the same template with `- Author/context: [seed] retrospective ...` and `Tried/Failed/Learned` compressed to one line each where the detail is lost.

### 7.3 Seed entries guidance

The seed block reconstructs origins only, one entry each, all marked `[seed]`:

1. Project origin -- why the IDP umbrella exists (source: `PROJECT_SUMMARY.md`).
2. M1 substrate completion (source: M1 specs + `SESSION_HANDOFF.md`).
3. M2 IaC + CD loop completion (source: `docs/specs/m2-iac-cd/`).
4. M3 observability completion (source: M3 triad + `scripts/smoke-m3.sh`).
5. M4 cost visibility and networking (source: M4 triads).
6. The 2026-08-18 goal-mode session, one entry per slice of the goal directive (source: `TODO.md` goal directive; slice 2 is this record-immutability workstream; the five-plane requirements document is a sibling product of the same session).

Everything after the seed block is contemporaneous by rule (7.1). The distinction between `[seed]` and later entries is what keeps the journal honest: the seed is a labeled reconstruction, not a fake of contemporaneity (NFR-006).

---

## 8. Exception process (FR-003 / DES-001 section 4)

### 8.1 Corrections-as-new-commits (the normal path)

To correct a mistaken committed record:

1. Identify the mistaken record: commit SHA, ADR number, or journal entry date/title.
2. Commit the correction as a NEW commit. Message convention (satisfies IN-M-001/002/003):

```text
fix(docs): correct <short description>

Corrects <full-sha> ("<subject of the mistaken commit>").

What was wrong: <...>
Why it was wrong: <...>
What this correction changes: <...>
Supersedes: <full-sha>
```

Subject stays within 72 chars (IN-M-001, `validator.go:55-58`); the `Supersedes:` reference lives in the body always, and may additionally appear in the subject as a short SHA only if the 72-char budget allows. For code corrections use the appropriate type (`fix(code): ...`); the `Supersedes:` body line is the constant element.
3. If the mistake is a decision, also follow the ADR supersession procedure (section 6.5).
4. If the mistake is a journal entry, add a new entry whose `Links:` field names the corrected entry; the old entry is not edited.

### 8.2 PROPOSED AMENDMENT -- emergency rewrite procedure (OQ-08; NOT APPROVED)

**Status: PROPOSAL ONLY -- NOT APPROVED.** REQ-001 FR-003 states "There is no procedure by which history is legitimately rewritten," and DES-001 section 4 states "The rule has no escape hatch." This subsection was originally drafted as an approved escape procedure and contradicts those approved positions; it is therefore demoted from procedure to proposal. It introduces open question **OQ-08**: "Emergency rewrite procedure: adopt the five-step deliberately-heavy procedure below as the single sanctioned exception, or keep the absolute no-rewrite rule?" OQ-08 is proposed here and should be adopted into REQ-001 section 7 when that document next revises. Until the user decides, the absolute no-rewrite rule remains in force, this procedure is **not available**, and it is **excluded from the rollout checklist** (section 11) -- no phase below contains, enables, or depends on it.

The analysis is kept because it is the strongest form of the amendment case. If OQ-08 adopts the proposal, all five steps become mandatory and ordered:

1. **Justification commit BEFORE the rewrite.** A commit on `main` adding a `docs/JOURNAL.md` entry (and an ADR if the reason is systemic) recording: what will be rewritten, why a correction commit is insufficient, who approved, and the date. This commit is itself protected by the no-rewrite rule.
2. **Signed archive tag of the pre-rewrite HEAD.** `git tag -s rewrite-archive-YYYY-MM-DD <pre-rewrite-head>` with a message containing the head SHA, tree SHA, UTC date, and a reference to the justification commit; pushed to origin AND github **before** any rewrite command runs.
3. **The rewrite.** The guards will block it: `IN-H-001` covers amend in agent context and `IN-H-002` covers the force-push. There is no bypass variable (sections 2.6, 3.3). The procedure therefore requires physically removing `.git/hooks/pre-push` for the duration -- a loud, local, filesystem-level act that must itself be named in the justification entry -- and reinstalling it immediately after via `plugins/rr-policy-guards/scripts/install-git-hooks.sh`.
4. **Closure commit and checkpoint.** After the rewrite: a journal closure entry referencing the `rewrite-archive-YYYY-MM-DD` tag, the old head SHA, and the new head SHA; then a new checkpoint tag (section 4) so the chain resumes from the new head.
5. **The archive tag is never deleted.** The rewritten-out history stays fetchable by anyone who cloned, and the divergence between the archive tag and the new history is permanently visible. The procedure changes what `main` points at; it never erases the record of what `main` used to point at.

If OQ-08 adopts the proposal and any step cannot be performed (e.g. the archive tag cannot be pushed), the rewrite does not proceed. If OQ-08 is answered "keep the absolute rule," this subsection becomes a rejected-proposal record: it stays in document history, superseded not deleted (per the exception model itself), and no emergency path exists.

---

## 9. Phase-2 guard audit-log hash chaining sketch (FR-012 / D-10)

Sketch only; scope is OQ-04. Verified current state: `commit-guard/audit.go:32-53` (`appendAuditStrict`) marshals one `AuditRecord` (types.go:57-66) to one JSON line and appends it (mode 0600); `appendAudit` (audit.go:28-30) fails open. Other guards have their own audit writers with their own schemas.

### 9.1 Chain format

One new field per JSONL line: `"prev_hash":"<64 lowercase hex chars>"` = SHA-256 of the **raw bytes of the previous line including its trailing newline**. Genesis line: `prev_hash` of 64 zeros. Hashing raw line bytes -- not a canonicalized record -- keeps the scheme shape-agnostic: it works identically for every guard's JSONL regardless of schema, which is what makes one shared mechanism possible across six different audit writers.

### 9.2 Write-side change (per guard, ~15 lines each)

In the guard's append function (for commit-guard: inside `appendAuditStrict`, audit.go:32-53): before writing, read the last line of the existing log (if any), compute its SHA-256, set `prev_hash`, marshal, append. Fail-open is preserved (NFR-003): if the tail read fails, write the line with `prev_hash` of zeros and never block the enforcement decision -- a broken chain link is a verification finding, not a commit failure.

### 9.3 Verify-side (shared small tool)

`plugins/rr-policy-guards/tools/audit-chain/` (stdlib only, same build contract as the other guards): `rr-audit-chain verify <log-path>` re-walks the file line by line, recomputes each `prev_hash`, reports the first mismatching line number, and prints the chain head hash (SHA-256 of the final line). `rr-audit-chain head <log-path>` prints just the head hash. One shared verifier is correct for every guard precisely because the chain is over raw bytes (9.1).

### 9.4 Honest limits

Chaining proves internal consistency only. Deletion or truncation of the log's tail is invisible unless the head hash is anchored elsewhere. Optional anchoring, decided with OQ-04: the checkpoint script (section 4) appends one `audit-head: <guard>=<hash>` line per in-scope log to the tag message, binding the log head into a signed, pushed tag. Even then the logs remain local-only and user-owned (REQ-001 section 4): chaining detects tampering, it cannot prevent it.

---

## 10. Test plan

Layout follows the existing convention: `<source>_test.go` beside the source (e.g. `extract_test.go`, `main_test.go`, `audit_test.go`, `validator_test.go`, `scanner_test.go`). Helpers reused: `makeRepo` (main_test.go:15-34, scratch git repo in `t.TempDir()`), `stage` (main_test.go:36-48), `runWithEnv` (main_test.go:50-58, routes `RR_COMMIT_GUARD_AUDIT_LOG` into a temp dir and blanks `RR_COMMIT_GUARD_BYPASS`). PreToolUse tests build input with `json.Marshal(ToolInput{ToolName: "Bash", ToolInput: map[string]any{"command": ...}})` (pattern at main_test.go:154-171).

### 10.1 Change point A (amend block)

- `extract_test.go`: `TestExtract_Amend` already exists (extract_test.go:63-68). Add table variants: `git -C /x commit --amend` (RepoDir preserved + Amend), `git commit --amend -m "x"` (Amend with message args).
- `main_test.go`: `TestPreToolUse_BlocksAmend` -- command `git commit --amend --no-edit`; expect exit 2; stderr contains `IN-H-001`; audit line has `"rule":"IN-H-001"` and `"mode":"pretooluse"`; stderr does **not** contain `RR_COMMIT_GUARD_BYPASS` (no bypass hint printed). Because the check precedes `StagedPaths` (section 2.3), this test needs no scratch repo.
- `main_test.go`: `TestPreToolUse_BlocksAmend_BypassIgnored` -- same command with `t.Setenv("RR_COMMIT_GUARD_BYPASS", "1")` (after `runWithEnv`'s blanking); still expect exit 2. Pins the no-bypass rule.
- Regression (NFR-007): the existing 50 commit-guard test functions (extract_test.go 10, main_test.go 12, audit_test.go 3, validator_test.go 10, scanner_test.go 15; count via `grep -c "^func Test"`) must pass unchanged.

### 10.2 Change point B (pre-push)

New `prepush_test.go`:

- `TestParseRefUpdates`: two well-formed lines parse; single trailing newline tolerated; a 3-field line returns an error.
- `TestRewritesMain` (table): fast-forward -> allow; non-fast-forward -> block with "non-fast-forward" reason; local sha zeros -> block with "deletion" reason; remote sha zeros -> allow (creation); `refs/heads/main-x` -> allow; `refs/tags/main` -> allow.
- `TestIsAncestor` (real git, via `makeRepo`): linear history -> true; diverged history -> false.
- `TestRunPrePush_*`: allow path -> exit 0, audit `"mode":"pre-push"`; block path -> exit 2, `"rule":"IN-H-002"`, stderr without `RR_COMMIT_GUARD_BYPASS`; malformed stdin -> exit 1 (internal-error contract, types.go:79-83); `RR_COMMIT_GUARD_BYPASS=1` set -> block path still exits 2.

Integration:

- Hook wiring: in a scratch bare remote + clone, install the wrapper, create divergent history, `git push` non-fast-forward of `main` -> push rejected, stderr shows `IN-H-002`; fast-forward push succeeds; push of a feature branch succeeds.
- Installer: run `plugins/rr-policy-guards/scripts/install-git-hooks.sh` into a `mktemp -d` repo; assert `pre-commit`, `commit-msg`, `pre-push` exist, mode 0755; run twice (idempotent); plant a foreign `pre-push` first and assert the `.bak.<timestamp>` backup behavior (install-git-hooks.sh:45-49).
- `bash -n` on the new hook and on `scripts/checkpoint-immutability.sh`; `shellcheck` where available (optional, noted if absent).

### 10.3 Checkpoint script

- `RECORD_CHECKPOINT_DRY_RUN=1` in a scratch repo with `user.signingKey` set: prints the message, creates no tag (`git tag -l` empty), pushes nothing.
- Refusal path: scratch repo **without** `user.signingKey`: exit 1, no tag created (assert `git tag -l 'checkpoint-*'` empty).
- Suffix logic: pre-create `checkpoint-YYYY-MM`, dry-run -> selects `checkpoint-YYYY-MM-r2`.
- PREV ordering: create `checkpoint-YYYY-MM`, `checkpoint-YYYY-MM-r2`, and `checkpoint-YYYY-MM-r10` in a scratch repo; assert the script's PREV resolves to `checkpoint-YYYY-MM-r10`. `-version:refname` compares digit runs numerically, so r10 ranks above r2 above the base tag (verified empirically on git 2.55.0); plain `-refname` would misrank r10 below r2 lexicographically, and `-creatordate` ties when tags share one commit.

### 10.4 Phase-2 chaining (if OQ-04)

- `audit_test.go`: `TestAudit_PrevHashChain` -- two appends; line 2's `prev_hash` equals SHA-256 of line 1's raw bytes + `\n`; genesis line carries 64 zeros.
- Verifier: `rr-audit-chain verify` passes a live log; a hand-edited middle line is detected at the edited line number; truncation is detected when compared against an anchored head (section 9.4).

### 10.5 Manual verification checklist per phase

- Phase 0: `git status --short` clean for in-scope documents; `docs/adr/0001-*.md`, `docs/adr/README.md`, `docs/JOURNAL.md` committed; `provenance/` tracked.
- Phase 1: `git commit --amend` attempt (agent context) blocks with `IN-H-001`; scratch-remote non-ff push of `main` blocks with `IN-H-002`; `git log --format='%G?' -1` = `G`; `go test ./...` green in commit-guard; `opa test --v0-compatible policies/*.rego` still 6/6 (NFR-007).
- Phase 2: `git tag -v checkpoint-YYYY-MM` good; `git ls-remote origin` and `git ls-remote github` both list the tag; verifier passes on in-scope logs (if OQ-04).

---

## 11. Rollout checklist

Ordered steps; each names its gate and its verification. Gates are the REQ-001 open questions; nothing proceeds past a gate without the user decision. OQ-08 (section 8.2) is deliberately NOT a gate for any step: the proposal is excluded from this checklist until decided, and phases 0-2 are complete without it because the default in force is the absolute no-rewrite rule.

### Phase 0 -- Baseline (gate: OQ-07, user commit approval)

1. Write `docs/adr/0001-record-architecture-decisions.md` (section 6.4 verbatim) + `docs/adr/README.md` (6.3). Verification: files present and committed.
2. Write `docs/JOURNAL.md` header + seed block (section 7). Verification: committed; seed entries marked `[seed]`.
3. Baseline commits per DES-001 D-01 (session-state docs incl. stale-claim corrections; attribution triple incl. first commit of `provenance/`; pre-existing code edits; the five-plane requirements doc and this triad). Verification: `git status --short` clean for in-scope paths; retained provenance-certificate chain started.

### Phase 1 -- Enforcement and signing (gate: OQ-01, key generated and registered as a user action)

4. Apply signing configuration, chosen branch of section 5. Verification: `git config --get-regexp` shows the keys; a scratch commit reports `%G?` = `G`.
5. Implement change point A (section 2) + tests (10.1); `go test ./...` in `plugins/rr-policy-guards/tools/commit-guard`; rebuild `plugins/rr-policy-guards/bin/rr-commit-guard`. Verification: amend attempt blocks with `IN-H-001` and no bypass hint.
6. Implement change point B (section 3) + tests (10.2); add `git-hooks/pre-push` source; update installer line 42 and the README/usage texts (3.5); run the installer. Verification: `.git/hooks/pre-push` mode 0755; scratch-remote non-ff push of `main` blocks; ff push passes.
7. OQ-05 disposition: delete `scripts/git-fix-email.sh` and `scripts/git-fix-author.sh` in one commit (recommended branch). Verification: files gone from worktree, deletion commit present.
8. OQ-06 (user action): enable force-push rejection for `main` on gitea.com. Verification: outcome recorded in `docs/JOURNAL.md`; a force-push attempt is rejected server-side.

### Phase 2 -- Anchoring and automation (gates: OQ-02 cadence, OQ-03 OTS, OQ-04 chaining)

9. Add `scripts/checkpoint-immutability.sh` (section 4); dry-run test (10.3); create and push the first checkpoint tag. Verification: `git tag -v` good; tag listed by `git ls-remote` on both remotes.
10. OTS branch, only if OQ-03 enables it: `ots stamp` the checkpoint tag's target digest; commit the `.ots` proof under `provenance/ots/`; `git ots`-style integration optional. Verification: `ots upgrade`/`ots verify` succeeds once the Bitcoin attestation confirms (attestation is pending until then -- record pending state honestly; OTS calendar liveness remains UNVERIFIED, REQ-001 OQ-03).
11. Chaining branch, only if OQ-04 enables it: `prev_hash` write-side in in-scope guards + `rr-audit-chain` verifier (section 9); optional `audit-head:` lines in the checkpoint message. Verification: verifier passes on live logs; deliberate-edit test detected.
12. Cadence per OQ-02 (monthly recommended): manual reminder or launchd/cron entry -- decision-neutral. Verification: second checkpoint exists one month later with `prev:` naming the first.

---

## 12. Verification of this specification

Every code-level claim above re-checks with these commands (run 2026-08-18):

```sh
# Change point A
grep -n 'case t == "--amend"' plugins/rr-policy-guards/tools/commit-guard/extract.go   # lines 90-91
grep -n 'Amend' plugins/rr-policy-guards/tools/commit-guard/types.go                   # line 73
grep -n 'Amend' plugins/rr-policy-guards/tools/commit-guard/*.go | grep -v _test       # parser + type only
grep -n 'IN-M-00' plugins/rr-policy-guards/tools/commit-guard/validator.go             # 001/002/003 (+004 in comments)
grep -n 'Code: "' plugins/rr-policy-guards/tools/commit-guard/scanner.go               # NV-*/GR-* taxonomy
sed -n '57,149p' plugins/rr-policy-guards/tools/commit-guard/main.go                   # runPreToolUse + bypass at 111
sed -n '294,297p' plugins/rr-policy-guards/tools/commit-guard/main.go                  # hintBypass
sed -n '1,13p' plugins/rr-policy-guards/tools/verify-guard/main.go                     # "no bypass path" line 13

# Change point B
cat .git/hooks/pre-commit .git/hooks/commit-msg                                        # wrapper shape mirrored in 3.4
diff plugins/rr-policy-guards/git-hooks/pre-commit .git/hooks/pre-commit               # 3-line bypass-comment drift
sed -n '42,52p' plugins/rr-policy-guards/scripts/install-git-hooks.sh                  # hook list line 42

# Signing
man git-config | col -b | grep -n -i 'gpg.format\|gpgsign\|allowedsignersfile\|signingkey'
git --version                                                                          # 2.55.0 (>= 2.34)
```

Traceability: sections 2-9 implement FR-003/FR-004/FR-005/FR-007/FR-008/FR-010/FR-011/FR-012 per the map in section 1; FR-001 (baseline), FR-002 (policy), FR-006 (OTS), FR-009 (corpus) are realized through sections 8 and 11 and the documents themselves (NFR-005 meta-consistency).

---

**End of Technical Specification**

This completes the record-immutability triad: RECORD-IMMUTABILITY-REQ-001 (what and why), RECORD-IMMUTABILITY-DES-001 (architecture), RECORD-IMMUTABILITY-TECH-001 (implementation contract). Implementation begins only after the REQ-001 open questions OQ-01..OQ-07 are decided, per the phase gates in section 11. OQ-08 may stay open without blocking any phase: its default is the absolute no-rewrite rule already approved in REQ-001 FR-003.
