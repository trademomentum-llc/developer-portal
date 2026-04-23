# TODO

> Action list ordered by priority and dependency.

**Snapshot date:** 2026-04-21 (late-session update)

---

## M2 IaC + CD Loop

### Done this session (committed on main, 25 M2 commits)

| Task | Commit |
|------|--------|
| M2 spec package (3 docs) | 41db8f0 |
| M2 implementation plan (24 tasks) | 3392c3e |
| T1 rr-tofu-guard failing parser tests | e86ef88 |
| T2 rr-tofu-guard parser implementation | 9a435bf |
| T3 rr-tofu-guard audit writer | aa74333 |
| T4 rr-tofu-guard main + integration tests | 7a59b01 |
| T4-fix missing test paths + audit error detail | e89194a |
| T5 register rr-tofu-guard PreToolUse hook | 074c87b |
| T5-fix remove-script defensive null handling | 0ee5acc |
| T6 score2openchoreo module + types | f7e756f |
| T7 score2openchoreo Convert + table tests | c2d1359 |
| T8 score2openchoreo schema validator + fixtures | 18d2e72 |
| T9 score2openchoreo CLI + main + golden tests | e0f067d |
| T10 Gatekeeper policies C-1 C-2 C-3 + Rego tests | ae3f073 |
| T10-fix C-2 Rego handles missing annotations | 0514a2a |
| T11 tofu root module scaffolding | aa0a0f1 |
| T12 tofu modules/flux | fecb218 |
| T13 tofu modules/gatekeeper | 700d2ef |
| T14 tofu modules/gitea-runner | 844175a |
| T15 tofu modules/openchoreo-environments | 7d71da3 |
| T16 tofu modules/external-secrets-wiring | ae49185 |
| T17 gitea + openbao seed scripts | 59fee9b |
| T18 seed-repos content | 6718753 |
| T19 canonical CI workflow + runner helpers | de3f383 |
| T20 install/teardown/smoke scripts | d3e9fb9 |
| T23 backstage proxy for /api/proxy/gitea-actions | cc56870 |
| T24 README M2 section | 45fa8a8 |

### Outstanding M2 work (for next session or operator-driven)

| Task | Status | Dependency |
|------|--------|------------|
| T21 Run scripts/install-m2.sh end-to-end (live) | PARTIAL 2026-04-22 | Tasks 0-3 succeed; task 4 (tofu apply) blocked -- see m2-install-blockers below |
| T22 First pipeline run on hello-m2 | NOT STARTED | T21 complete |
| Push local commits to remote | DONE 2026-04-22 | Pushed to origin (local Gitea openchoreo cluster) and gitea-com cloud; local Forgejo still deferred (not yet installed) |

### M2 install blockers (2026-04-22 partial run)

| ID | Item | Status |
|---|---|---|
| m2i-1 | OpenChoreo CRD group drift -- iac/modules/openchoreo-environments references `core.choreo.dev/v1alpha1/Environment`; installed CRD is `environments.openchoreo.dev`. Update API path. | open |
| m2i-2 | k3d-openchoreo single-node cluster CPU-exhausted -- Flux helm-controller stuck Pending (`Insufficient cpu`); times out tofu helm_release. Either scale Docker resources for k3d, free CPU, or switch to multi-node config. | open |
| m2i-3 | Stale `tofu-state` namespace from 2026-04-21 attempt -- run `tofu import kubernetes_namespace.tofu_state tofu-state` before retry. | open |
| m2i-4 | ExternalSecret manifests may be impacted by alekc/kubectl provider discovery cache misses under cluster pressure (likely resolves when m2i-2 fixed). | open, monitor |
| m2i-5 | Deprecated `k3d-m1-substrate` cluster (Gitea pod stuck Terminating, agents NotReady) consumes Docker resources unnecessarily. Tear down with `k3d cluster delete m1-substrate`. | open |
| m2i-6 | OpenBao runs in dev mode with `inmem` storage -- secrets and `kv/` mount die on pod restart. Plan persistent storage + auto-unseal for production-readiness. | open, low-priority |

---

## Post-M2 tech debt (discovered during subagent-driven reviews)

Track these before declaring M2 complete. None are blocking Tasks 21-22.

### Guards (rr-policy-guards)

| ID | Item | Status |
|---|---|---|
| guard-1 | Document or surface silently-swallowed audit I/O errors | DONE 2026-04-23 -- verified all 4 guards carry file- and function-level swallow-semantics comments; TODO description was stale |
| guard-2 | Pick ONE registration path for rr-tofu-guard | DONE 2026-04-23 -- finding on audit: only settings.json was ever live (plugin never installed). Pivoted from "install as Claude plugin" to "single-source-of-truth in settings.json" after finding that mid-session hook reload works -- all 4 guards registered in ~/.claude/settings.json with absolute paths; smoke-tested each binary. Plugin hooks.json + .claude-plugin/marketplace.json left in the repo as future-ready scaffolding in case a plugin-install migration is ever wanted. Backup at ~/.claude/settings.json.bak-1776979520. |
| guard-3 | Fix merge-tofu-hook-into-settings.sh jq filter to push into existing Bash matcher instead of appending a new one | DONE commit d4373e6 |
| guard-4 | Investigate overmatch: `wc -w` heredoc containing word `tofu` was blocked; narrow the match | DONE 2026-04-23 -- added IsTofuCommandPrefix gate before Tokenize in tofu-guard; 13 unit subtests + integration test reproducing the audit-log case; binary rebuilt; regressions preserved |
| guard-5 | Backport corrected remove-script filter into the plan doc (commit 074c87b vs plan lines 684-690) | DONE 2026-04-23 -- plan doc now carries the fixed filter that preserves sibling hooks on the Bash matcher |

### score2openchoreo

| ID | Item | Status |
|---|---|---|
| score-1 | Decide and implement inline `${resources.X.Y}` substitution OR error loudly on partial matches | DONE 2026-04-23 -- errors loudly per design call (deterministic-first); 4 subtests cover prefix/suffix/middle/two-refs; error names the variable |
| score-2 | Add tests for: multi-container sort order, multi-variable sort order, `environment` resource branch, missing-resource error, annotations-to-labels mapping | DONE commit 55ce5a7 |
| score-3 | Document secret-name fallback "X-secret" convention | DONE 2026-04-23 -- inline comment in convert.go at the secret-name fallback site; new tools/score2openchoreo/README.md documents all three conversion conventions |
| score-4 | Lowercase error strings (Go idiom) | DONE commit 0983b63 |
| score-5 | Pin score.schema.json to a git SHA instead of the moving `main` branch | DONE 2026-04-23 -- vendored file is the pin; SHA256 recorded in assets/SCHEMA_PROVENANCE.md with update procedure; TestScoreSchemaPin guards against drift |

### Gatekeeper

| ID | Item | Status |
|---|---|---|
| gk-1 | Update policies/README.md with correct `opa test --v0-compatible policies/*.rego -v` invocation | DONE commit 680f40d |

### Install / scripts

| ID | Item | Status |
|---|---|---|
| inst-1 | Verify or create `scripts/lib/{colors,wait-for,confirm}.sh` (referenced by install-m2.sh) | DONE commit 5e1e088 |
| inst-2 | Move shebangs to line 1 in all smoke scripts | DONE commit 5e1e088 |

---

## Push / remote resolution

| Option | Action |
|---|---|
| A | Create `trademomentum.net/developer-portal` in local Gitea; push to origin |
| B | Fix gitea.com connectivity; rotate embedded credential; push to gitea-com |
| C | Retire localhost origin; rename gitea-com to origin |

The embedded credential in the gitea-com URL is exposed in `.git/config`
and in at least one conversation transcript -- rotate regardless of which
option is chosen.

---

## M3-M7 (unchanged from prior snapshots)

| Milestone | Scope | Status |
|---|---|---|
| M3 | OpenTelemetry + SigNoz + Infracost post-deploy dashboards | deferred |
| M4 | OpenCost + Cilium + Envoy Gateway | deferred |
| M5 | RabbitMQ or Kafka + OpenResty front-door | deferred |
| M6 | OPA/Gatekeeper runtime policies + MISP + TheHive + Cortex + Velociraptor + Cloud Custodian | deferred |
| M7 | MCP plugin surfacing Backstage + RR to OpenChoreo + per-agent Gitea tokens | deferred |
