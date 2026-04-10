# TODO

> Action list ordered by priority and dependency.

**Snapshot date:** 2026-04-10

---

## Completed this session

| Task | Status |
|------|--------|
| P0.1 Verify emoji guard | DONE -- registered, matcher Write/Edit/MultiEdit |
| P0.2 ASCII enforcement test | DONE -- em dash blocked, exit 2 |
| P1.0 Build brew-guard | DONE -- 47 tests, binary at plugins/rr-policy-guards/bin/rr-brew-guard |
| P1.0b Build bash-guard | DONE -- 27 tests, blocks bare $VAR, suggests corrected command |
| P1.1 Install yarn | DONE -- 1.22.22 via brew (brew-guard allowed it) |
| P1.5 Scaffold Backstage | DONE -- npx create-app, yarn install running |
| P1.6 Wire Backstage to Gitea | DONE -- app-config.yaml patched, index.ts updated |
| P1.9 Install + teardown scripts | DONE -- scripts/install-m1.sh + teardown-m1.sh |
| P2.1 Update Tech Spec paths | DONE -- brew-guard relocated to plugins/rr-policy-guards/ |
| gitea-values.yaml | DONE -- written to scripts/ |

---

## Blocked on Colima/Docker

| Task | Status | Dependency |
|------|--------|------------|
| P1.3 k3d cluster + Gitea helm install | BLOCKED | Docker daemon (Colima) |
| P1.4 Create demo Gitea repo | BLOCKED | Gitea running |
| P1.8 Verify Backstage + Gitea integration | BLOCKED | Everything running |
| P1.10 README | BLOCKED | After verification |
| P1.11 Done-definition checklist | BLOCKED | After verification |

Colima is being killed by an integrity monitor. User is working on it.

---

## Deferred

| Task | Trigger |
|------|---------|
| OpenChoreo 3-plane integration | Deferred to M3 |
| P2.2 Initialize git in developer-portal | Ask user first |
| Knockout EDA implementation | Design spec approved, needs writing-plans |

---

## P3-P5 -- Unchanged from prior session

See previous TODO snapshot for M2-M7 roadmap, cross-cutting items,
and "things to NOT do" list. All still apply.
