# score.schema.json -- provenance and pin

The embedded Score schema is vendored, not fetched at build or run time.
Treat this file as the pin.

## Current pin (the file in this directory)

- **File:** `assets/score.schema.json`
- **SHA256:** `633c4394dfc03977c86932f5d10c77e2c356a38b63f162c281104fade5c9863c`
- **Bytes:** 16701
- **Pinned on:** 2026-04-23 (commit that adds this file into the pin log)
- **Upstream source:** https://github.com/score-spec/spec
- **Upstream commit:** `3ecb17d430c2bbf46d2dfc161fabc7d432d6d1f5` (2026-04-17)
- **Upstream path:** `score-v1b1.json` (repository root); byte-identical
  match against the raw file at that commit verified 2026-08-18
- **Schema apiVersion covered:** `score.dev/v1b1`

`schema_pin_test.go` fails if the embedded file's SHA256 drifts from the
value above. Any deliberate update must bump BOTH the file and this doc.

## Why pinned, not fetched

Score-5 (TODO): earlier versions of this repo referenced
`score-spec/spec@main`, a moving branch, which would let an upstream
silent change flip validator behavior between pipeline runs. Vendoring
with an explicit SHA pin is deterministic, offline-capable, and
reviewable. See CLAUDE.md "Deterministic first" project rule.

## Update procedure

1. Pick a specific upstream commit SHA from score-spec/spec (e.g. a
   tagged release). Record it in a PR description.
2. Fetch that commit's Score schema file (upstream path:
   `score-v1b1.json` at the repository root, as of the 2026-08-18
   verification).
3. Replace `assets/score.schema.json` with the fetched bytes exactly.
4. Compute the new SHA256: `shasum -a 256 assets/score.schema.json`.
5. Update this file: set **SHA256**, **Bytes**, **Pinned on**, and add a
   note pointing at the upstream commit SHA.
6. Run `go test ./...` -- `TestScoreSchemaPin` must pass, and no
   existing conversion test should regress. If the schema now rejects a
   previously-accepted Score document, decide whether to update the
   fixture or reject the upstream change.
7. Commit both files in the same commit with a message like
   `chore(score): bump score schema pin to <upstream SHA>`.
