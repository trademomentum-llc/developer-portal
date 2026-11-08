# Requirements Specification: OpenBao Production-Grade Storage

**Document ID:** BAO-STORAGE-REQ-001
**Version:** 0.1
**Date:** 2026-08-24
**Status:** Draft -- pending user approval (G5 is the top-ranked open gate in the TODO.md register; per phase-gate discipline this triad precedes any cluster change)
**Relationship:** Closes register item G5 / m2i-6 (carried since 2026-05-02). Design Specification and Technical Specification follow per `~/Projects/Sovereign/Structure/POLICIES.md`.

---

## 1. Purpose

OpenBao runs in dev mode with `inmem` storage. Every pod restart or node
flap wipes all secrets: the Gitea runner registration token, the Flux
deploy key, and the per-app runtime secrets that ExternalSecrets delivers
to workloads. Recovery is a manual reseed via
`scripts/seed-openbao-m2-paths.sh`. This document states the requirements
for moving OpenBao to persistent storage so that restarts lose nothing.

## 2. Evidence of the failure mode (measured)

| Date | Event | Evidence |
|---|---|---|
| 2026-08-21 | Colima stop x2; inmem wiped; manual reseed | SESSION_HANDOFF 0b |
| 2026-08-24 ~10:00 | server-0 rolling restart; wipe; reseeded | this repo, smoke-openbao PASS after reseed |
| 2026-08-24 ~14:30 | node flap (containerd down); wipe; smoke-m2 403 on `kv/gitea/runners/token`; reseeded | smoke-all output, openbao-0 RESTARTS=2 |

Four incidents in four days. Each incident is only detected when a
downstream consumer fails (runner cannot register, smoke 403s), so the
failure mode is silent between checks.

## 3. Goal

OpenBao state survives pod restarts and cluster churn with zero manual
reseed steps. After this work, deleting `openbao-0` is a non-event.

## 4. Functional requirements

- FR-1: OpenBao storage backend is persistent (Raft integrated storage
  or file) on a local-path PVC in the `openbao` namespace.
- FR-2: Seal/unseal strategy is defined and implemented for local
  operation (dev mode's no-seal behavior ends). Unseal key custody is
  decided in OQ-1; keys never enter git.
- FR-3: `scripts/install-m2.sh` / a new install step bootstraps the
  persistent backend; `scripts/seed-openbao-m2-paths.sh` becomes a
  one-time bootstrap, not a recovery tool.
- FR-4: `scripts/smoke-openbao.sh` gains an inverse-proof lane: delete
  `openbao-0`, wait for readiness, assert all four M2 keys are present
  WITHOUT any reseed. This lane fails today (verified 2026-08-24, twice).
- FR-5: Teardown preserves the PVC unless an explicit `--wipe-secrets`
  flag is given.
- FR-6: The Gitea runner token path (`kv/gitea/runners/token`) and the
  OpenChoreo runtime app-secret path (`secret/apps/...`) are covered by
  the persistence claim.

## 5. Non-functional requirements

- NFR-1: No new standing cluster workload beyond the existing openbao-0
  pod; memory footprint increase bounded to the PVC + Raft overhead
  (<100 MiB measured at acceptance).
- NFR-2: Recovery time from pod deletion to secret availability < 2 min
  on the measured host envelope.
- NFR-3: Unseal keys and root token stored under `~/.rational-reserve/`
  with mode 600, excluded from all remotes; provenance updated.
- NFR-4: The change flows through the install scripts (tofu-guard
  compliant); no ad-hoc kubectl mutations remain as the steady state.

## 6. Options considered

| Option | Verdict | Reason |
|---|---|---|
| A. Raft integrated storage + local-path PVC | PROPOSED | Self-contained, no external dependency, HA-ready later |
| B. File backend + local-path PVC | Fallback | Simpler, but no path to HA; Raft preferred for parity with production topology |
| C. PostgreSQL backend (reuse existing postgres) | Rejected | Secrets infrastructure must not share fate with an app database |
| D. Stay dev-mode + CronJob auto-reseed | Rejected | Automates the symptom; token rotation and ExternalSecrets drift remain |
| E. Replace ExternalSecrets with plain k8s Secrets | Rejected | Architecture regression; M2's secret-delivery design stands |

## 7. Open questions

- OQ-1: Unseal key custody -- Shamir keys in `~/.rational-reserve/`
  (mode 600) with a documented manual-unseal runbook, vs auto-unseal via
  a local file-based key (convenience vs honest security posture).
- OQ-2: Backup cadence for the PVC (none / weekly snapshot to
  `~/.rational-reserve/backups/` / leave to record-immutability
  checkpoint tags).
- OQ-3: Does CI (act-runner dind) keep any dev-mode dependency that a
  persistent backend would break? (Expected: no; verify in the Technical
  Specification.)

## 8. Acceptance criteria

1. The FR-4 inverse-proof lane passes: secrets survive pod deletion with
   no reseed.
2. `smoke-openbao.sh` green immediately after a full cluster restart
   (Colima stop/start), with no manual steps.
3. `smoke-all.sh` green after the change.
4. Design Specification + Technical Specification approved before any
   cluster mutation.
