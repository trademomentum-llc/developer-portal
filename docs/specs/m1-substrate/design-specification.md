# M1 Substrate -- Design Specification

> **Milestone:** M1 -- Substrate
> **Version:** 1.0
> **Date:** 2026-04-09
> **Status:** Draft, awaiting user approval
> **Companion docs:** [requirements.md](./requirements.md), [technical-specification.md](./technical-specification.md)

---

## 1. Purpose

This document describes *how* the M1 substrate is shaped. It sits one level above the Technical Specification: it answers "why is it structured this way" and "what are the moving parts and how do they relate," but does not list every file path, every API route, or every dependency version. Those belong in the Technical Specification.

Read this document to understand the architectural decisions, the component boundaries, and the rationale. Read the Technical Specification to know what to type.

## 2. Context diagram

```
                          +--------------------------------------+
                          |       Operator's macOS laptop         |
                          |  (Apple Silicon, Darwin 25+, Colima) |
                          +----------------+---------------------+
                                           |
         +---------------------------------+---------------------------------+
         |                                                                   |
+--------v---------+   +-------------------+   +-------------------+   +-----v---------+
|  Claude Code     |   |    Backstage       |   |    brew guard      |   |   Source tree  |
|  (or other host) |   |    (host process)  |   |    (Go binary)     |   |   (git repo)   |
|                  |   |    yarn dev        |   |    PreToolUse      |   |                |
|  Bash tool use --+-->|    :3000           |   |    hook            |   |  developer-    |
|                  |   |                    |   |                    |   |  portal/       |
+------------------+   +---------+----------+   +--------^-----------+   +----------------+
                                 |                       |
                                 | catalog discovery     | validates every
                                 | + API proxy           | brew command
                                 |                       |
         +-----------------------v-----------------------+------------------+
         |                    Colima VM (docker daemon)                     |
         |                                                                  |
         |    +----------------------------------------------------------+  |
         |    |                k3d cluster (k8s 1.33)                   |  |
         |    |                                                          |  |
         |    |   namespace: openchoreo-system                          |  |
         |    |   +-- controller                                         |  |
         |    |   +-- openchoreo-api       <-- proxied by Backstage     |  |
         |    |   +-- observer                                           |  |
         |    |   +-- cluster-gateway                                    |  |
         |    |   +-- cluster-agent                                      |  |
         |    |                                                          |  |
         |    |   namespace: gitea                                       |  |
         |    |   +-- gitea (server)       <-- queried by Backstage     |  |
         |    |   +-- postgres (built-in)                                |  |
         |    |   +-- memcached (built-in)                               |  |
         |    |                                                          |  |
         |    +----------------------------------------------------------+  |
         |                                                                  |
         +------------------------------------------------------------------+
```

## 3. Component boundaries

M1 has five components, each with a single clearly-owned responsibility:

| Component | Responsibility | Runs where | Language |
|---|---|---|---|
| **brew guard** | Validate `brew` commands before execution | Host, invoked per Bash tool use | **Go** |
| **k3d cluster** | Provide Kubernetes to everything else | Colima VM | -- (infra) |
| **OpenChoreo** | Platform orchestrator (component CRDs, workload runtime, policy) | In cluster | Go (upstream) |
| **Gitea** | Self-hosted git server with API | In cluster | Go (upstream) |
| **Backstage skeleton** | Developer portal front door (catalog + proxy for now) | Host, `yarn dev` | **TypeScript** (necessary -- see Section 5) |

No component owns the responsibilities of another. No component has a dependency on another component's internal data structures -- they communicate only via documented APIs (OpenChoreo's REST API, Gitea's API, Backstage's proxy).

## 4. Design decisions and rationale

### 4.1 Why Backstage runs on host in M1, not in the cluster

Backstage in M1 is a scaffold. The operator will iterate on `app-config.yaml`, install plugins, and tweak catalog discovery. Containerizing Backstage for M1 would kill the `yarn dev` hot-reload loop and add a docker build cycle to every single edit. The cost of "Backstage is in the cluster" is high, the benefit (parity with prod) is low at this stage because there is no prod yet.

**Trade-off accepted:** the operator runs two process managers -- k3d manages in-cluster components, and the operator manually runs `yarn dev` for Backstage. Documented clearly in README.

**Reversal cost:** low. In a later milestone (likely M3 when OpenTelemetry lands and we need Backstage traced), we containerize Backstage with a Dockerfile and a helm chart. The catalog config and the plugin set carry over unchanged.

### 4.2 Why OpenChoreo deploys via its own quick-start, not raw helm

OpenChoreo's `make quick-start.dev` target is the project's official path. It wraps:

- k3d cluster creation with the right flags
- OpenChoreo controller installation
- OpenChoreo-API, observer, cluster-gateway, and cluster-agent deployment
- Default component types and workflows

Replicating this via raw `helm install` is possible but would (a) diverge from the upstream documented path, making upgrades fragile, and (b) require duplicating the logic openchoreo's own Makefile already encodes. The cost of the quick-start path is zero; the cost of reinventing it is high.

**Trade-off accepted:** M1 is tightly coupled to openchoreo's quick-start evolving. If openchoreo changes its quick-start target between major versions, M1's install script may need adjustment. Mitigation: pin openchoreo to a specific git SHA in the Technical Specification (Section 8) and test against that SHA.

### 4.3 Why Gitea is in-cluster via helm, not outside

Three options were considered:

1. **In-cluster via helm** (chosen) -- Gitea lives alongside OpenChoreo, reachable by in-cluster DNS (`gitea.gitea.svc.cluster.local`). Backstage (on host) uses a NodePort or `kubectl port-forward` to reach it.
2. **docker-compose outside the cluster** -- Gitea runs directly on the Colima docker daemon. Simpler infra but creates a second "surface" (two places that host components: cluster and docker-compose). Not wanted.
3. **Native binary on host** -- Gitea is a single Go binary. Simplest install but puts Gitea outside any isolation layer, complicating teardown.

Option 1 wins because it keeps everything in one place, lets teardown be a single `k3d cluster delete + helm uninstall`, and makes the eventual production path (Gitea in k8s, somewhere else) nearly identical.

### 4.4 Why brew guard is Go, not Bash or Python

The user's standing preference (memory): deterministic languages when possible, interpreted only when necessary. The brew guard is greenfield, has no language dependency from the ecosystem, and runs on every Bash tool invocation -- so cold-start latency matters. Measurements on Apple Silicon:

| Language | Typical cold-start | Binary size |
|---|---|---|
| Bash script | ~5 ms | -- |
| Python 3 script | 60-100 ms | -- |
| Go binary | 2-5 ms | ~3 MB |

Bash is fastest but error-prone for argument parsing (shell quoting is a nightmare). Python is slowest. Go is nearly as fast as bash for cold start, has proper argument parsing, proper JSON decoding, and produces a single artifact. Go wins on all axes except "zero setup" (Go needs to be compiled, which adds a build step to M1).

**Trade-off accepted:** a build step is added to Task 0 (compile the brew guard binary). Cost is ~2 seconds on Apple Silicon. Benefit is correctness and speed for every subsequent invocation.

### 4.5 Why deny-list parsing for brew guard, not CVE lookup

Three approaches considered:

1. **Deny-list command parser** (chosen) -- parse the command line, reject known-dangerous forms. Fast, deterministic, no network, no external state.
2. **Allow-list command parser** -- only permit a hardcoded set of formulas. Safest, but every new install requires editing the allow-list. High friction.
3. **CVE database lookup** -- per invocation, query an external vulnerability DB (GitHub Advisory, osv.dev). Most thorough, but network-dependent (fails offline), slow (network RTT), and gives false positives for unrelated CVE metadata.

The deny-list approach catches the things that actually introduce supply-chain risk (untrusted taps, URL installs, force flags, shell injection) without blocking normal use. It is the right shape for "validate brew commands" given that the user's threat model is "don't let a typo or a malicious command slip through a Claude-Code tool invocation," not "audit every formula against the CVE database."

### 4.6 Why Backstage <-> OpenChoreo is a proxy entry for M1, not a custom plugin

M1's scope for the Backstage <-> OpenChoreo integration is exactly "Backstage can reach OpenChoreo's API." That is one line in `app-config.yaml`:

```yaml
proxy:
  '/openchoreo':
    target: http://openchoreo-api.openchoreo-system.svc.cluster.local:8080
    changeOrigin: true
```

A custom Backstage plugin (which is what surfaces OpenChoreo components in the catalog as first-class entities) is M2 work. It requires designing the Backstage entity model mapping, writing frontend components, and handling auth. None of that is needed to prove "the substrate is wired up."

**Trade-off accepted:** M1 end users can browse the catalog and see Gitea repos, but they can't click a component and see its OpenChoreo deployment status. That UX lands in M2.

### 4.7 Why demo repo in Gitea uses `catalog-info.yaml` (Backstage format), not Score

Backstage's catalog discovery plugin reads `catalog-info.yaml` files -- that's Backstage's native format for describing components. Score specs are a different format (workload specification) designed for platform orchestrators. In M2, when we wire the commit -> deploy flow, the demo repo will grow a `score.yaml` file alongside its `catalog-info.yaml`. For M1, only the Backstage-native format is needed because M1's only job is "prove the catalog discovery works."

## 5. Language exception: Backstage is TypeScript

Per the Requirements (NFR-5), any use of an interpreted language must be documented. Backstage is TypeScript/Node.js. This is an ecosystem-forced choice -- Backstage is a TypeScript project and there is no production-ready equivalent in a compiled language. The Backstage scaffolder (`@backstage/create-app`) produces a Node.js + TypeScript app that must be run under Node.js and built with `yarn`. Porting Backstage to Go or Rust is not feasible within the scope of this platform build.

**Documented exception:** all Backstage code in M1 is TypeScript. Any NEW code we write (the brew guard, install scripts, helpers) defaults to Go unless the ecosystem demands otherwise.

## 6. Data flow

### 6.1 Install flow

```
operator runs scripts/install-m1.sh
    |
    +-- [Task 0] build rr-brew-guard binary
    |     +-- register in ~/.claude/settings.json as PreToolUse hook
    |
    +-- [Task 0.5] brew install yarn     <- validated by the hook
    |
    +-- [Task 1] cd into /Users/nnos/Projects/openchoreo
    |     +-- make quick-start.dev
    |           +-- k3d cluster create
    |           +-- helm install openchoreo charts
    |           +-- wait for pods Ready
    |
    +-- [Task 2] helm repo add gitea-charts
    |     +-- helm install gitea gitea-charts/gitea
    |         +-- wait for gitea pod Ready
    |
    +-- [Task 3] create Gitea admin + demo repo (via Gitea API)
    |     +-- push catalog-info.yaml to demo repo
    |
    +-- [Task 4] cd into developer-portal
    |     +-- npx @backstage/create-app  (one-time scaffold, idempotent check)
    |
    +-- [Task 5] patch backstage/app-config.yaml
    |     +-- add gitea plugin integration
    |     +-- add openchoreo proxy entry
    |
    +-- [Task 6] run yarn dev (backgrounded)
          +-- verify http://localhost:3000 returns 200
```

### 6.2 Runtime flow -- Backstage catalog discovery

```
Backstage (on host) -- Gitea API (via kubectl port-forward or NodePort) --> list repos
                          |
                          +-- for each repo: GET /raw/main/catalog-info.yaml
                                |
                                v
                          parse YAML --> upsert component in Backstage DB
```

### 6.3 Runtime flow -- Backstage proxy to OpenChoreo

```
curl http://localhost:3000/api/proxy/openchoreo/health
    |
    v
Backstage proxy handler
    |
    +-- rewrites to http://openchoreo-api.openchoreo-system.svc.cluster.local:8080/health
          |
          v
     (via kubectl port-forward or NodePort -- decided in Technical Specification Section 4)
          |
          v
     OpenChoreo-API responds with 200 { "status": "ok" }
    |
    v
Backstage returns the response to the caller
```

## 7. Brew guard architecture (detailed)

```
Claude Code Bash tool invocation
    |
    | (tool use JSON with "command" field)
    v
PreToolUse hook ("matcher": "Bash")
    |
    | pipes JSON to stdin
    v
/Users/nnos/Projects/developer-portal/tools/rr-brew-guard/bin/rr-brew-guard
    |
    +-- read stdin as JSON
    +-- extract tool_input.command
    +-- lex into tokens (respecting quoted strings)
    |
    +-- is first token "brew"?
    |     no -> exit 0 (not our concern)
    |
    +-- is second token in {install, reinstall, upgrade}?
    |     no -> exit 0 (brew info, list, etc. are harmless)
    |
    +-- scan remaining tokens:
    |     +-- URL pattern?               -> exit 2 "url-based install"
    |     +-- flag not in allow-list?    -> exit 2 "disallowed flag: X"
    |     +-- "tap" anywhere?            -> exit 2 "untrusted tap"
    |     +-- shell metacharacter?       -> exit 2 "shell metacharacter"
    |     +-- package name regex fail?   -> exit 2 "suspicious package name"
    |
    +-- check env RR_BREW_GUARD_BYPASS=1?
    |     yes -> append bypass event to ~/.rational-reserve/logs/brew-guard.jsonl, exit 0
    |
    +-- exit 0 (allow)
```

### 7.1 Allow-list of safe flags

```
--quiet
--no-auto-update
--formula
--cask
```

Any flag not in this set causes a block. Extending the list is a code change, reviewable in a PR.

### 7.2 Audit log format

Every block and every bypass writes a single line of JSON to `~/.rational-reserve/logs/brew-guard.jsonl`:

```json
{"ts":"2026-04-09T19:50:00Z","action":"block","reason":"untrusted tap","command":"brew tap evil/src","session":"<id>"}
{"ts":"2026-04-09T19:51:00Z","action":"bypass","reason":"RR_BREW_GUARD_BYPASS=1","command":"brew install --HEAD something","session":"<id>"}
```

The audit log is append-only, owned by the user's account (mode 0600), and never rotated automatically in M1. Log rotation is a future concern when volume justifies it.

## 8. File structure (top-level, fully materialized in Technical Spec)

```
/Users/nnos/Projects/developer-portal/
+-- README.md                              <- how to run
+-- docs/
|   +-- specs/
|       +-- m1-substrate/
|           +-- requirements.md            <- you are not reading this
|           +-- design-specification.md    <- you are reading this
|           +-- technical-specification.md <- read this next
|   +-- superpowers/
|       +-- plans/
|           +-- 2026-04-09-m1-substrate.md <- implementation plan (produced later)
+-- tools/
|   +-- rr-brew-guard/
|       +-- go.mod
|       +-- main.go
|       +-- main_test.go
|       +-- bin/rr-brew-guard              <- build output (gitignored)
|       +-- README.md
+-- scripts/
|   +-- install-m1.sh                      <- orchestration
|   +-- teardown-m1.sh                     <- reverse
+-- backstage/                             <- scaffolded in Task 4, not committed until after scaffold
```

Exact paths, module names, dependency versions, and implementation contents live in the Technical Specification.

## 9. Interfaces

### 9.1 brew guard interface

- **Input:** JSON on stdin, matching Claude Code's PreToolUse protocol
- **Output:** none on allow; human-readable error message on stderr on block
- **Exit codes:** 0 = allow, 2 = block, 1 = internal error (logged, defaults to block for safety)
- **Environment variables:**
  - `RR_BREW_GUARD_BYPASS=1` -- override (logged as bypass)
  - `RR_BREW_GUARD_AUDIT_LOG` -- override audit log path (default `~/.rational-reserve/logs/brew-guard.jsonl`)

### 9.2 OpenChoreo API (consumed)

M1 only needs OpenChoreo's `/health` endpoint (or equivalent readiness probe). The proxy entry in Backstage's `app-config.yaml` targets the cluster-internal service; the full API surface is consumed in later milestones.

### 9.3 Gitea API (consumed)

M1 uses:
- `GET /api/v1/repos/search` -- for Backstage catalog discovery (via the official Backstage Gitea plugin)
- `POST /api/v1/repos` -- for one-time demo repo creation during install
- `PUT /api/v1/repos/{owner}/{repo}/contents/{filepath}` -- to upload `catalog-info.yaml`

All calls use the admin token generated during Gitea's first-run configuration.

## 10. Error handling and failure modes

| Failure | Observable symptom | Response |
|---|---|---|
| Colima not running | `docker version` fails | install-m1.sh detects and prints clear message: "Start Colima first: `colima start`" |
| openchoreo quick-start fails mid-run | make target exits non-zero | install-m1.sh stops, prints last 50 lines of make output, does not proceed to Gitea install |
| Gitea helm install hangs (image pull timeout) | pod in ImagePullBackOff state | install-m1.sh times out after 5 minutes, prints `kubectl describe pod` output |
| Backstage `yarn dev` fails to start | port 3000 not listening after 60s | install-m1.sh prints yarn's stdout/stderr, leaves cluster running for debug |
| brew guard binary missing when Claude Code tries Bash | PreToolUse hook exits non-zero | Claude Code surfaces the hook error; operator is told to run Task 0 |
| Over-aggressive brew guard blocks a legitimate install | operator sees "blocked by brew guard: X" | operator adds the package pattern to deny-list exceptions OR sets `RR_BREW_GUARD_BYPASS=1` for one command |

## 11. Alternatives considered and rejected

| Alternative | Rejected because |
|---|---|
| Use Docker Desktop instead of Colima | Colima is already running and working. Changing tools mid-project is unnecessary. |
| Use kind instead of k3d | OpenChoreo's quick-start uses k3d; aligning with upstream is cheaper than re-wiring. |
| Use Forgejo instead of Gitea | Forgejo is a Gitea fork and is technically excellent, but Gitea has more mature Backstage plugin support in M1's timeframe. |
| Skip Backstage entirely and use OpenChoreo's built-in UI | OpenChoreo does not have a catalog/scaffolder UI comparable to Backstage. The user explicitly named Backstage in the stack. |
| Build brew guard in Bash | Argument parsing correctness is too risky in Bash; shell quoting is a known hazard. |
| Build brew guard in Python | Fails the deterministic-languages preference for the one reason that matters most (cold-start latency on hot paths). |
| Do not build a brew guard; rely on manual inspection | The user explicitly required a pre-command hook. Not optional. |

## 12. Open questions resolved in this document

- **Q1: Where does Backstage run in M1 -- host or cluster?** -> Host, see Section 4.1.
- **Q2: How does Backstage reach OpenChoreo's API -- proxy entry or custom plugin?** -> Proxy entry, see Section 4.6.
- **Q3: What language is the brew guard?** -> Go, see Section 4.4.
- **Q4: Deny-list, allow-list, or CVE lookup for brew validation?** -> Deny-list, see Section 4.5.
- **Q5: Where do the Gitea credentials live?** -> Filesystem outside source tree, mode 0600, see Requirements NFR-8.

## 13. Open questions deferred to Technical Specification

- **Q6:** Exact Gitea helm chart version and values.yaml content
- **Q7:** Exact Backstage scaffold options (`npx @backstage/create-app` flags)
- **Q8:** Exact openchoreo git SHA pinning strategy
- **Q9:** NodePort or kubectl port-forward strategy for Gitea access
- **Q10:** Exact rr-brew-guard Go module structure and dependency choices (if any beyond stdlib)

## 14. Success criteria revisited

M1 is a well-formed design when:

- Every requirement in `requirements.md` maps to at least one component or design decision in this document
- Every design decision has explicit rationale or an explicit trade-off
- Every interface between components is documented
- Every alternative considered is recorded, with rejection reasoning
- Nothing in this document contradicts `requirements.md`

These five tests are the self-review checklist. The Technical Specification adds one more layer -- it tests whether an engineer with zero context can type their way through.
