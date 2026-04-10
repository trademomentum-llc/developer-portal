# SESSION HANDOFF

> Read this file FIRST in the next session. It tells you where we stopped,
> what is now active that was not before, what the user's standing
> preferences are, and exactly what to do first.

**Last session ended:** 2026-04-09 ~20:20
**Reason for handoff:** User asked for handoff files and a session restart so
that the new emoji guard PreToolUse hook becomes active.

---

## 1. The single most important thing

A new PreToolUse hook is registered in `~/.claude/settings.json`. It became
active at session start (this session). It is called `rr-emoji-guard` and it
will reject any `Write`, `Edit`, or `MultiEdit` tool invocation whose content
contains a single non-ASCII byte.

The user's rule is absolute and not negotiable: **plain ASCII in all files.**
No emojis. No box-drawing characters (do not use anything in U+2500 family).
No em dashes (use `--`). No smart quotes (use straight `'` and `"`). No
Unicode arrows (use `->`, `<-`, `<->`). No section signs (write `Section 4.2`).
No check marks or cross marks (use words: `passes`, `fails`, `ok`, `blocked`,
`missing`).

The previous session violated this rule by accident roughly 2,564 times across
68 files, was caught by the user, and had to do a bulk Python fix to make
everything ASCII again. Do not repeat that. If you find yourself reaching for
a typographic dash or a fancy arrow, stop and use ASCII.

The hook lives at:
`/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/bin/rr-emoji-guard`

Source is at:
`/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/tools/emoji-guard/`

The hook source is Go, stdlib only, with full unit and integration tests
(all passing). It blocks on exit code 2 with a clear stderr message naming
the line, column, code point, and category. There is an emergency bypass:
`RR_EMOJI_GUARD_BYPASS=1` allows one command through, logged.

Audit log: `~/.rational-reserve/logs/emoji-guard.jsonl` (JSONL, append-only).

---

## 2. User's standing preferences (all in persistent memory)

These are saved at `~/.claude/projects/-Users-nnos-Projects/memory/` and the
agent will see them automatically. Brief recap:

1. **Plain ASCII in all files.** Already covered above.
2. **Deterministic / compiled languages preferred.** Use Go for new tools,
   scripts, hooks, and services. Reach for an interpreted language only when
   the ecosystem forces it (e.g., Backstage is TypeScript because Backstage is
   TypeScript). Document any interpretation decision explicitly.
3. **Three-document plan format.** Every non-trivial plan must be written as
   three separate markdown documents BEFORE any implementation: a Requirements
   Document, a Design Specification, and a Technical Specification. Implementation
   Plan (TDD bite-sized tasks) is a fourth document, produced after the user
   approves the first three.
4. **brew install requires the security pre-check hook.** This hook is NOT
   built yet. The plan calls it `rr-brew-guard` and it should be added to
   the same `plugins/rr-policy-guards/` plugin as a sibling of `rr-emoji-guard`.
   The `security-guidance` plugin the user has installed does NOT do this
   despite the user's earlier assumption -- it only checks file content
   patterns, not Bash commands.

---

## 3. Where we stopped

The session was working on three layered things:

**Layer 0: Check-tools.sh fix in openchoreo (DONE)**
The user has openchoreo at `/Users/nnos/Projects/openchoreo/`. We unblocked
its preflight script `check-tools.sh` so that all installed tool versions
now pass. The script exits 0 with all green checks.

**Layer 1: Rational Reserve v0.1 + v0.2 build (DONE in earlier turns)**
The user had us build a "Rational Reserve" military-hierarchy AI swarm
orchestration system from scratch in `/Users/nnos/Projects/rational-reserve/`.
We built it in Python (interpreted -- BEFORE the deterministic preference
was stated). It is functionally complete: 65 tests pass, end-to-end smoke
test works. State persists across daemon restarts.

NOTE: Now that the deterministic preference is in memory, a Go rewrite of
RR is a legitimate future candidate. Do not initiate it; surface the option
if the user asks.

**Layer 2: developer-portal IDP build (BLOCKED on user spec review)**
This is the umbrella platform engineering project. The user wants a full
self-hosted IDP based on the Platform Engineering reference architecture in
`/Users/nnos/Downloads/platform.pptx`. Decomposed into 7 milestones M1-M7.
M1 is "the Substrate": k3d cluster + OpenChoreo + Gitea + Backstage skeleton.

We wrote all three M1 spec documents to:
`/Users/nnos/Projects/developer-portal/docs/specs/m1-substrate/`
- `requirements.md`
- `design-specification.md`
- `technical-specification.md`

These specs are awaiting the user's review and approval. Once approved, the
NEXT step is to invoke `superpowers:writing-plans` to produce the TDD
implementation plan (the fourth document), then begin building.

Two open questions in the spec review (detailed in the specs themselves):
- Backstage running on host vs in cluster -- I argued for host
- OpenChoreo crossing three planes as load-bearing -- the user already
  confirmed this is intentional

---

## 4. What to do first in the new session

In this exact order:

1. Read this file (SESSION_HANDOFF.md) first.
2. Read PROJECT_SUMMARY.md to refresh yourself on the three projects.
3. Read TODO.md to see the action list.
4. Verify the emoji guard hook is loaded by running `cat ~/.claude/settings.json`
   and confirming the PreToolUse Bash hook entry pointing at rr-emoji-guard
   is present. If it is not, the user will tell you something is wrong.
5. Greet the user and ask whether they want to:
   (a) Review the three M1 spec documents before you proceed (recommended)
   (b) Approve the specs as-is and have you produce the M1 Implementation Plan
   (c) Switch to a different task entirely
6. Wait for their direction before doing anything else.

Do NOT:
- Start implementing M1 without explicit user approval of the specs.
- Run `brew install` anything until the rr-brew-guard hook is built.
- Write any file containing a non-ASCII character; trust the hook to enforce it.

---

## 5. Critical files to load into your context early

In the new session, before doing meaningful work, read at least these:

```
/Users/nnos/Projects/developer-portal/docs/specs/m1-substrate/requirements.md
/Users/nnos/Projects/developer-portal/docs/specs/m1-substrate/design-specification.md
/Users/nnos/Projects/developer-portal/docs/specs/m1-substrate/technical-specification.md
/Users/nnos/Projects/developer-portal/PROJECT_SUMMARY.md
/Users/nnos/Projects/developer-portal/TODO.md
/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/README.md
```

The three RR memory files (deterministic-languages, three-document-plans,
brew-security-hook, plain-ASCII) load automatically as part of the user's
memory system.

---

## 6. State of the three projects in one paragraph each

**openchoreo** (`/Users/nnos/Projects/openchoreo/`): unchanged from upstream
EXCEPT for `check-tools.sh` which had its version bounds bumped to match
modern tool versions and now exits 0. Colima is running, Docker daemon is
reachable, all preflight checks pass except Kubectl Server (which gracefully
skips when no kube context is set). Ready for `make quick-start.dev` whenever
M1 implementation begins.

**rational-reserve** (`/Users/nnos/Projects/rational-reserve/`): full v0.1
spine (core Python library, models, factory, c2_router, doctrine loader,
SQLite persistence, daemon, MCP server with 3 tools) plus v0.2 adapter layer
(SKILL.md pack, primers, install.sh, MCP config snippets for 5 host agents).
65 unit + integration tests, all passing. End-to-end smoke test confirms
state persists across reconnections. Phase v0.3 (LLM execution),
v0.4 (lifecycle/AAR), Phase 2, Phase 3 are all explicit follow-up specs.

**developer-portal** (`/Users/nnos/Projects/developer-portal/`): repository
created. Contains three M1 spec documents (awaiting user approval), the
rr-policy-guards plugin (built, tested, hook registered), and these handoff
files. No implementation has begun on M1 itself. Backstage is not yet
scaffolded. No Gitea installed. No k3d cluster running. yarn is not yet
installed -- it is the only missing prerequisite for M1.

---

## 7. Things the user is paying attention to

Based on the session, the user will likely react strongly to:

- Any non-ASCII character in any file (will be caught by the hook now)
- Process drift (skipping the three-document plan format, jumping to code)
- Proceeding without explicit approval at decision gates
- Using interpreted languages where Go would have worked
- Any brew install attempted before the rr-brew-guard hook exists

The user values clear paper trails, explicit decisions, and being asked
before being told. When in doubt, stop and ask.

---

## 8. Things to remember about the user's preferred workflow

- Three-document specs (Requirements, Design, Technical) FIRST, then a
  separate Implementation Plan, then build.
- Use the superpowers:brainstorming skill at the start of any creative work.
- Use the superpowers:writing-plans skill to produce the TDD implementation
  plan AFTER the three specs are approved.
- Use the plugin-dev:hook-development skill if you need to add another hook
  to the rr-policy-guards plugin.
- The user has invoked these skills explicitly several times. They want
  process discipline, not improvisation.
