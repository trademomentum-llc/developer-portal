# Simulation Report: OpenBao Production-Storage Migration (Rehearsal)

**Document ID:** BAO-STORAGE-SIM-001
**Date:** 2026-08-26
**Subject spec:** BAO-STORAGE-TECH-001 (`2026-08-26-OpenBao-Production-Storage-Technical-Specification.md`)
**Method:** full rehearsal against a throwaway instance -- helm release `openbao-sim`, namespace `openbao-sim`, chart openbao-0.25.6, custody under `~/.rational-reserve/openbao-sim/`, on the shared single-node `k3d-openchoreo` cluster. The live OpenBao release was never touched.
**Evidence:** all logs under `/tmp/openbao-sim/logs/` (36 files); corrected scripts under `/tmp/openbao-sim/scripts/`.

---

## 1. Lane results

| Lane | Result | Measured evidence |
|---|---|---|
| Static (bash -n, helm template) | PASS | all scripts parse; rendered template carries sidecar, `storageClassName: local-path`, 1Gi VCT, Retain/Retain, zero postStart |
| Fixture install (dev mode + live postStart values) | PASS | 13 platform secrets landed (`fixture-13secrets.log`) |
| FR-4 pre-change inverse proof | PASS (failed as spec'd) | `FAIL (post-restart kv/gitea/runners/token absent -- no reseed was run)`, exit 1 (`fr4-pre.log`, `migrate-1.log` step 1) |
| Migration (s5 runbook) | PASS after one environmental retry | first `helm upgrade` hit a cluster API timeout (`migrate-1.log`); re-run completed; bootstrap PASS (`migrate-4.log`) |
| FR-4 post-change persistence | PASS | secrets survive pod deletion with no reseed |
| NFR-2 recovery bound (<120s) | FAIL first run, PASS after corrections | first run 142s deletion-to-secrets on the loaded shared node (`fr4-post.log`); final re-run 77s (`remigrate-fixed.log`) |
| smoke-openbao `--with-restart` JSON verdicts | as spec'd | pre: 0/1 fail (expected); post: 4/1 (the 1 = NFR-2 breach); post-rollback: 0/1 fail (expected) (`smoke-*.jsonl`) |
| Backup (raft snapshot) | PASS | `snapshot written: .../openbao-sim/openbao-20260826-153101.snap` (`backup.log`) |
| Rollback | PASS | release returns to dev-mode semantics; post-rollback restart loses secrets again -- `FAIL (...token absent -- no reseed...)` (`fr4-post-rollback2.log`) proves rollback restored the old behavior |
| Re-migration after rollback | FAIL then PASS | first attempt: `custody root token rejected by the backend (custody/cluster mismatch)`, exit 1 (`remigrate-defect.log` = defect D5); after correction C9: full PASS, 77s recovery (`remigrate-fixed.log`) |
| Teardown | PASS | namespace, PVCs, custody and backup dirs removed; `kubectl get pvc -A` and `get ns` show no sim residue (measured 2026-08-26) |

## 2. Defects found in the spec listings, and corrections applied in sim

- **D1 (environmental):** first `helm upgrade` failed on a cluster API
  timeout under load. Not a spec defect; consistent with the repo's
  serialized-heavy-ops rule. Correction: retry succeeded; no spec change.
- **D2:** NFR-2 recovery bound breached on first run (142s > 120s) on the
  loaded shared node; 77s on the quiet re-run. Correction **C8 v2**:
  wait budgets overridable via env (`OPENBAO_RESTART_WAIT_SECONDS` et
  al.), spec default 120s kept.
- **D3 / C6:** `grep -q` exits on first match; `kubectl exec` streams
  frames, so under node load the producer is SIGPIPEd and `pipefail`
  misreads a present mount as absent (observed 3x: 400 "path is already
  in use"). Correction: plain `grep` reading to EOF. Sites:
  `seed-openbao-m2-paths.sh`, `bootstrap-openbao-persistent.sh`.
- **D4 / C5, C7, C8:** bare `kubectl wait pod/<name>` races pod
  recreation after `--wait=true` deletion -- it errors NotFound and
  `set -e` kills the script (observed in fr4-pre, migrate-2, rollback).
  Correction: retry/poll loop within the same time budget. Sites:
  `smoke-openbao.sh`, `install-openbao-storage.sh`,
  `rollback-openbao-storage.sh`.
- **D5 / C9:** a retained PVC alone does not mean the Raft template is
  live -- after rollback the release is dev-mode again while the PVC
  persists, and the spec's re-run path then failed against the dev
  backend ("custody root token rejected"). Correction: the orchestrator
  skips migration steps 1-2 only when the current StatefulSet carries
  the unseal sidecar (the raft-template marker); otherwise it falls
  through to the full path, whose upgrade re-attaches the retained PVC
  and whose bootstrap recovers the persisted Raft state.

## 3. Divergences from spec expectations

- Recovery time is load-sensitive on the shared node: 142s under load,
  77s quiet. The 120s NFR-2 default is achievable but not guaranteed
  during concurrent heavy operations -- one more argument for the
  serialized-ops rule.
- One `kubectl apply` annotation warning on `openbao-unseal-key`
  (cosmetic; the placeholder Secret is created imperatively by design).

## 4. GO/NO-GO

**GO, conditional:** all mechanics are proven end-to-end, including
rollback and re-migration. Before live implementation, corrections
C5-C9 must be folded into the BAO-STORAGE-TECH-001 listings (the
corrected reference copies live at `/tmp/openbao-sim/scripts/` until
implementation). With those folded, no known blocker remains.

## 5. Honest limits

- The sim ran on the same node that hosts the live stack; it proves
  procedure and semantics, not performance isolation.
- The live instance carries real consumer traffic (ExternalSecrets
  sync); the sim covered the same Secret topology but not concurrent
  consumer load during migration.
- Log evidence is in `/tmp` and will not survive a reboot; the
  corrections are duplicated into the tech spec (spec v0.2) so the
  durable record does not depend on `/tmp`.
