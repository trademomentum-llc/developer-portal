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
