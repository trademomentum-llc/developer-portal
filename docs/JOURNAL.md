# Engineering Journal -- developer-portal

Append-only contemporaneous log of what was tried, what failed, and what
was learned (RECORD-IMMUTABILITY-REQ-001 FR-008; DES-001 D-06).

Rules: entries are appended at the end of this file. Existing entries are
never edited -- corrections are new entries referencing the old. Entries
marked [seed] are retrospective reconstructions written once at adoption
time; every entry after the seed block is contemporaneous, written during
the work, not reconstructed after it.

---

## 2026-08-18 -- Project origin (retrospective)

- Author/context: [seed] retrospective reconstruction written at journal adoption; sources PROJECT_SUMMARY.md and AGENTS.md.
- Tried: stand up a self-hosted Internal Developer Platform as an umbrella repo, decomposed into seven milestones (M1-M7), modeled on the Platform Engineering community reference architecture, on top of the sibling openchoreo cluster and the deferred rational-reserve swarm.
- Failed: n/a (origin entry).
- Learned: n/a (origin entry).
- Links: PROJECT_SUMMARY.md, AGENTS.md, docs/specs/2026-05-28-IDP-Policy-Guard-Layer-Requirements.md.

## 2026-08-18 -- M1 substrate (retrospective)

- Author/context: [seed] retrospective; sources docs/specs/m1-substrate/ and PROJECT_SUMMARY.md.
- Tried: M1 substrate -- host tooling, the M1 policy guards (emoji, bash, brew), Backstage scaffold, and the local Gitea/OpenBao substrate on k3d-openchoreo.
- Failed: no failure detail retained in the state docs (detail lost; that loss is part of why this journal exists).
- Learned: "M1 substrate complete and healthy" (PROJECT_SUMMARY.md); install-m1.sh made resumable via checkpoints.
- Links: docs/specs/m1-substrate/ (requirements, design-specification, technical-specification), scripts/install-m1.sh.

## 2026-08-18 -- M2 IaC + CD loop validated (retrospective)

- Author/context: [seed] retrospective; sources TODO.md M2 section and docs/specs/m2-iac-cd/.
- Tried: M2 push-to-deploy loop -- Score validated and rendered by score2openchoreo, built by Gitea Actions, reconciled by OpenChoreo, with OpenTofu, Flux drift correction, and Gatekeeper constraints C1-C3.
- Failed: early blockers were tracked in TODO.md's remediation list rather than in a contemporaneous log (detail compressed).
- Learned: "M2 is validated end-to-end locally" through a fresh hello-m2 pipeline run (TODO.md, 2026-05-28 update); later live runs #24/#27/#30 repeated the loop.
- Links: docs/specs/m2-iac-cd/, seed-repos/hello-m2/, scripts/smoke-m2.sh.

## 2026-08-18 -- M3 observability live (retrospective)

- Author/context: [seed] retrospective; sources SESSION_HANDOFF.md section 1 and the M3 triad.
- Tried: M3 observability -- SigNoz v0.130.1 plus standalone OTEL collector v0.155.0, hello-m2 trace instrumentation with openchoreo.* and git.commit.sha attributes, Backstage entity tabs.
- Failed: the SigNoz collector's OpAMP-only manager arguments kept OTLP ports 4317/4318 closed until the Deployment was patched (SESSION_HANDOFF.md section 1).
- Learned: scripts/smoke-m3.sh passes 22/22 including a live trace-ingestion assertion; run #27 (a6eaf5a) put the first end-to-end trace into ClickHouse with the pod 1/1 Running.
- Links: docs/specs/2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md (+ Design/Technical), docs/specs/m3-observability/, scripts/smoke-m3.sh.

## 2026-08-18 -- M4 cost visibility and networking (retrospective)

- Author/context: [seed] retrospective; sources the M4 triads and SESSION_HANDOFF.md.
- Tried: M4 cost visibility -- OpenCost + Prometheus in the opencost namespace feeding the Backstage CostCard -- and M4 networking via Envoy Gateway with .local routes for gitea/signoz/opencost.
- Failed: no failure detail retained in the state docs.
- Learned: smoke-m4.sh and smoke-m4-networking.sh (3/3 routes) pass, and smoke-all.sh reported ALL SMOKE SUITES PASSED (M2, M3, M4) per SESSION_HANDOFF.md.
- Links: docs/specs/2026-06-30-M4-Cost-Visibility-Requirements.md (+ Design/Technical), docs/specs/2026-06-30-M4-Networking-Requirements.md (+ Design/Technical), scripts/smoke-m4.sh, scripts/smoke-m4-networking.sh.

## 2026-08-18 -- Goal-mode slice: provenance attribution triple (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session, first package (mandatory attribution practice); source SESSION_HANDOFF.md section 0.
- Tried: build the full attribution triple -- THIRD-PARTY-LICENSES.md (full inventory in 8 groups), provenance/PROVENANCE.md (189 evidenced entries), and the recognition certificate (r2), with AGENTS.md recording the practice.
- Failed: 25 entries could not be evidenced on the first pass; they were recorded openly as UNVERIFIED rows U1-U25 instead of being smoothed over.
- Learned: evidence-first listing with openly recorded gaps survived an adversarial critic review; attribution is recorded, never claimed.
- Links: THIRD-PARTY-LICENSES.md, provenance/PROVENANCE.md, provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md.

## 2026-08-18 -- Goal-mode slice: five-plane portal roadmap (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session slice per TODO.md ("Next slices" item 1).
- Tried: consolidate five evidence-backed plane maps (observation, control, orchestration, security, engagement) into one umbrella requirements and roadmap document.
- Failed: first pass required a critic correction round before approval.
- Learned: the portal is mostly a link farm today; the doc fixes a baseline of 53 gap registers, a 5x5 traversal matrix with 12 breakdowns, 40 FRs + 10 NFRs, 45 PROPOSED components (nothing decided outside the locked stack), 4 RECOMMENDED phases, and 31 open questions -- and records that threat intelligence, the security plane's namesake, has zero implementation.
- Links: docs/specs/2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md.

## 2026-08-18 -- Goal-mode slice: U1-U25 resolution (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session slice (user-directed); source SESSION_HANDOFF.md section 0.
- Tried: resolve the 25 UNVERIFIED provenance rows with hard evidence in five resolver bundles.
- Failed: 6 rows stayed blocked by the stopped Colima/k3d cluster (U8, U9, U12, U13, U16, U17 -- re-run when the cluster is up).
- Learned: 15 rows fully resolved and 4 narrowed (U7, U11, U19, U25); several stale upstream claims corrected on evidence (uuid holder, alpine:3.20 package enumeration, golang:1.26-alpine base, Score schema pinned to 3ecb17d430c2); certificate re-issued as r3; critic round 3 APPROVE with zero defects.
- Links: provenance/PROVENANCE.md, provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md.

## 2026-08-18 -- Goal-mode slice: record-immutability triad (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session slices four and five per SESSION_HANDOFF.md; source TODO.md item 2.
- Tried: specify the record-immutability mechanism as a full governance triad -- REQ-001 (12 FRs, 7 NFRs), DES-001 (10 design elements, 5 layers, 3 phases), TECH-001 (12 implementation-grade sections).
- Failed: TECH-001 critic round 1 found one BLOCKER -- the drafted emergency-rewrite escape procedure contradicted approved REQ-001 FR-003 and DES-001 section 4 -- and one critic sort-order claim was itself empirically wrong.
- Learned: the escape procedure was reframed as PROPOSED amendment OQ-08 (NOT APPROVED, excluded from rollout); the sort claim was adjudicated against the critic on evidence (git tag --sort=-version:refname orders r10 > r2 > base); the triad is critic-approved and implementation waits on user decisions OQ-01..OQ-08.
- Links: docs/specs/2026-08-18-Record-Immutability-Requirements.md, docs/specs/2026-08-18-Record-Immutability-Design-Specification.md, docs/specs/2026-08-18-Record-Immutability-Technical-Specification.md.

## 2026-08-18 -- Goal-mode slice: anomaly cleanup (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session slice six per SESSION_HANDOFF.md; source TODO.md item 3.
- Tried: fix the cross-document anomalies surfaced by the provenance and immutability passes (guard counts, phantom files, dead pins, stale remote-topology claims).
- Failed: no failed step recorded; convergence took three critic rounds.
- Learned: .claude-plugin/marketplace.json and the guard README now state the six-guard / five-of-six-bypass reality (verify-guard the test-pinned exception); four dead minimatch descriptor pins removed; namespace-predictor comment path corrected; stale gitea-com-blocked/origin entries annotated with dates (history preserved, push status honestly UNVERIFIED); certificate re-issued as r4.
- Links: .claude-plugin/marketplace.json, plugins/rr-policy-guards/README.md, tools/namespace-predictor/main.go, provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md.

## 2026-08-18 -- Goal-mode slice: guard enforcement implementation (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session slice seven per SESSION_HANDOFF.md; source TODO.md item 4.
- Tried: implement RECORD-IMMUTABILITY-TECH-001 change points in rr-commit-guard -- the IN-H-001 amend block placed before the bypass check (so it cannot be bypassed) and a fourth --pre-push mode applying IN-H-002 to deletion and non-fast-forward updates of main (githooks(5) stdin parsing, merge-base --is-ancestor, fail-closed including the lost-race case).
- Failed: no test failures; one residual accepted and documented -- a compound-hidden amend (e.g. git add x && git commit --amend) passes the PreToolUse gate, the same risk class as a raw-terminal amend, with IN-H-002 still blocking publication to main.
- Learned: 63/63 tests pass (50 pre-existing + 13 new) with e2e against real git; the installer now covers three hooks but was NOT run -- hook activation stays with the user and .git/hooks is untouched.
- Links: plugins/rr-policy-guards/tools/commit-guard/main.go, plugins/rr-policy-guards/tools/commit-guard/prepush.go, plugins/rr-policy-guards/git-hooks/pre-push, docs/specs/2026-08-18-Record-Immutability-Technical-Specification.md (sections 2-3).

## 2026-08-18 -- Goal-mode slice: checkpoint script implementation (retrospective)

- Author/context: [seed] retrospective; 2026-08-18 goal-mode session slice eight per SESSION_HANDOFF.md; source TODO.md item 5.
- Tried: implement scripts/checkpoint-immutability.sh per TECH-001 section 4 -- signed annotated checkpoint-YYYY-MM tags, prev:-chained via --sort=-version:refname, verify-before-push, dual-remote push, dry-run support.
- Failed: the critic's first round caught a missing-remote preflight gap (MINOR-1); fixed -- the script now also refuses when either the origin or github remote is missing.
- Learned: 10/10 tests pass (refusal paths, signed happy path against a throwaway SSH key, -r2 rerun chaining, base/-r2/-r10 chaining to -r10, dual-remote push, dry-run purity, bash -n + shellcheck); the real repo was untouched throughout.
- Links: scripts/checkpoint-immutability.sh, scripts/tests/test-checkpoint-immutability.sh, docs/specs/2026-08-18-Record-Immutability-Technical-Specification.md (section 4).

## 2026-08-18 -- Goal-mode slice: ADR and journal instantiation (retrospective)

- Author/context: [seed] retrospective written in the same act that creates this journal; final 2026-08-18 goal-mode slice.
- Tried: instantiate the rationale layer per RECORD-IMMUTABILITY-TECH-001 sections 6-7 as files only -- docs/adr/ (TEMPLATE.md, 0001-record-architecture-decisions.md, README.md index) and this docs/JOURNAL.md with its seed block.
- Failed: none recorded.
- Learned: files-only instantiation keeps OQ-07 (baseline commit approval) genuinely open -- nothing here is committed yet; the seed block reconstructs project origin through this slice with [seed] markers on every retrospective entry, no fake contemporaneity.
- Links: docs/adr/0001-record-architecture-decisions.md, docs/adr/README.md, docs/JOURNAL.md, docs/specs/2026-08-18-Record-Immutability-Technical-Specification.md (sections 6-7).

---

**End of seed block.** Every entry appended below this line must be
contemporaneous -- written during the work, per the rules in the header.

## 2026-08-20 -- Wave-0 acceptance on the resized cluster

- Author/context: goal continuation session; user completed the Colima
  resize (6 CPU / 12 GiB), unblocking the queued Wave-0 acceptance.
- Tried: re-run install-m3.sh, run the full smoke umbrella, push
  hello-m2 through the new Trivy/OSV security gates live, and land the
  in-flight Engagement-plane slice (CiRunsCard + tab).
- Failed: the first live gate runs (#39/#40) exposed a vacuous-pass
  wiring bug -- dind sibling containers made `-v "$PWD:/src"` an empty
  mount, so Trivy fs "passed" scanning nothing; a stale local-Gitea
  developer-portal mirror then 127'd the artifact-commit step; and a
  partial packages/app/dist (no index.html) had both Backstage backends
  404ing /, silently failing the auth and production smokes. The
  linkify-it 5.0.2 downgrade, first judged spurious on tsc evidence,
  proved REQUIRED once build:all actually ran. And after run #46 went
  green, the rollout itself was blocked: all three cluster-plane agents
  had been silently disconnected since 06-30 (websocket bad handshake)
  because the Cluster*Plane CRs pin the install-time agent clientCA and
  the self-signed agent certs renewed on 06-26.
- Learned: static review and tsc do not cover runner topology or
  bundler type paths -- only live execution does (third time this
  project: serviceLocatorMethod, the dind mounts, linkify-it). Time-boxed
  platform lesson: self-signed agent certs pinned into CRs are a renewal
  time bomb (next expiry 2026-09-24); re-pinned all three CRs, agents
  reconnected, pod rolled to :59b8c8d. Fixed: --volumes-from
  "$(hostname)" mounts (6964ca9), mirror synced (dbd79de), dist rebuilt
  + index.html guard (ea29c84), smoke umbrella ALL SUITES PASSED (AUTH,
  M2, M3, M4, SECURITY, BACKSTAGE-PRODUCTION), smoke-security 46/0/3
  (FR-03 flipped PASS), Engagement slice committed (e82f2bc),
  provenance cert r9.
- Links: seed-repos/hello-m2/.gitea/workflows/ci.yaml,
  backstage/packages/app/src/modules/openchoreo-cards/CiRunsCard.tsx,
  scripts/start-backstage-production.sh, provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md.

