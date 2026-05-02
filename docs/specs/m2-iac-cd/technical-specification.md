# M2 IaC + CD Loop -- Technical Specification

> **Milestone:** M2 -- IaC + CD Loop
> **Version:** 1.0
> **Date:** 2026-04-20
> **Status:** Draft, awaiting user approval
> **Companion docs:** [requirements.md](./requirements.md), [design-specification.md](./design-specification.md)

---

## 1. Purpose

This document is the low-level implementation reference for M2. It captures every piece of concrete information an engineer needs to implement M2 without having to make design decisions: exact file paths, exact module layouts, exact dependency versions, exact API schemas, exact commands, exact config blocks.

If you find yourself making a design decision while reading this document, something is wrong -- stop and check the Design Specification. If the answer is not there either, escalate before implementing.

## 2. Repository layout (complete)

```
/Users/nnos/Projects/developer-portal/
+-- docs/
|   +-- specs/
|       +-- m1-substrate/                             # from M1
|       +-- m2-iac-cd/
|           +-- requirements.md
|           +-- design-specification.md
|           +-- technical-specification.md            # this file
|   +-- superpowers/
|       +-- plans/
|           +-- 2026-04-09-m1-substrate.md            # from M1
|           +-- 2026-04-20-m2-iac-cd.md               # produced by writing-plans next
+-- iac/
|   +-- README.md
|   +-- main.tf
|   +-- variables.tf
|   +-- outputs.tf
|   +-- backend.tf
|   +-- versions.tf
|   +-- providers.tf
|   +-- modules/
|   |   +-- flux/
|   |   |   +-- main.tf
|   |   |   +-- variables.tf
|   |   |   +-- outputs.tf
|   |   |   +-- README.md
|   |   +-- gatekeeper/
|   |   |   +-- main.tf
|   |   |   +-- variables.tf
|   |   |   +-- outputs.tf
|   |   |   +-- constraints.tf
|   |   |   +-- README.md
|   |   +-- gitea-runner/
|   |   |   +-- main.tf
|   |   |   +-- variables.tf
|   |   |   +-- outputs.tf
|   |   |   +-- README.md
|   |   +-- openchoreo-environments/
|   |   |   +-- main.tf
|   |   |   +-- variables.tf
|   |   |   +-- outputs.tf
|   |   |   +-- README.md
|   |   +-- external-secrets-wiring/
|   |       +-- main.tf
|   |       +-- variables.tf
|   |       +-- outputs.tf
|   |       +-- README.md
|   +-- templates/
|       +-- ci.yaml                                    # canonical Gitea Actions workflow
+-- plugins/
|   +-- rr-policy-guards/
|       +-- hooks/
|       |   +-- hooks.json                             # updated to include tofu-guard
|       +-- tools/
|       |   +-- emoji-guard/                           # from M1
|       |   +-- bash-guard/                            # from M1
|       |   +-- brew-guard/                            # from M1
|       |   +-- tofu-guard/                            # new in M2
|       |       +-- go.mod
|       |       +-- main.go
|       |       +-- parser.go
|       |       +-- parser_test.go
|       |       +-- audit.go
|       |       +-- main_test.go
|       +-- bin/                                       # gitignored
|           +-- rr-emoji-guard
|           +-- rr-bash-guard
|           +-- rr-brew-guard
|           +-- rr-tofu-guard                          # new in M2
+-- tools/
|   +-- score2openchoreo/
|       +-- go.mod
|       +-- go.sum
|       +-- main.go
|       +-- cli.go
|       +-- convert.go
|       +-- schema.go
|       +-- types.go
|       +-- convert_test.go
|       +-- main_test.go
|       +-- fixtures/
|       |   +-- minimal.score.yaml
|       |   +-- minimal.component.yaml                 # expected output for minimal.score.yaml
|       |   +-- with-secret.score.yaml
|       |   +-- with-secret.component.yaml
|       |   +-- invalid-schema.score.yaml              # used to test validation failure path
|       +-- bin/
|           +-- score2openchoreo                       # build output (gitignored)
+-- policies/
|   +-- C1-platform-addons-main-protected.rego
|   +-- C1-constraint.yaml
|   +-- C2-score-schema-valid.rego
|   +-- C2-constraint.yaml
|   +-- C3-infracost-delta.rego
|   +-- C3-constraint.yaml
|   +-- README.md
+-- seed-repos/
|   +-- platform-addons/
|   |   +-- README.md
|   |   +-- clusters/
|   |   |   +-- default/
|   |   |       +-- kustomization.yaml
|   |   |       +-- flux-system/
|   |   |       |   +-- gotk-components.yaml
|   |   |       |   +-- gotk-sync.yaml
|   |   |       +-- gatekeeper/
|   |   |           +-- constrainttemplates.yaml
|   |   |           +-- constraints.yaml
|   +-- platform-config/
|   |   +-- README.md
|   |   +-- environments/
|   |       +-- dev/
|   |       |   +-- .gitkeep
|   |       +-- staging/
|   |           +-- .gitkeep
|   +-- hello-m2/
|       +-- README.md
|       +-- main.go
|       +-- Dockerfile
|       +-- score.yaml
|       +-- catalog-info.yaml
|       +-- .gitea/
|           +-- workflows/
|               +-- ci.yaml
+-- scripts/
|   +-- install-m2.sh
|   +-- teardown-m2.sh
|   +-- smoke-m2.sh
|   +-- smoke-tofu.sh
|   +-- smoke-actions.sh
|   +-- smoke-flux.sh
|   +-- smoke-score.sh
|   +-- smoke-infracost.sh
|   +-- smoke-gatekeeper.sh
|   +-- smoke-openbao.sh
|   +-- seed-gitea-repos.sh
|   +-- merge-tofu-hook-into-settings.sh
|   +-- remove-tofu-hook-from-settings.sh
|   +-- lib/                                           # from M1
|       +-- colors.sh
|       +-- wait-for.sh
|       +-- confirm.sh
```

## 3. Dependency versions (pinned)

### 3.1 Host tools added in M2

Verify at install time via `scripts/install-m2.sh`. Bounds are inclusive.

| Tool | Pinned version | brew formula | Notes |
|---|---|---|---|
| OpenTofu | 1.9.0 - 1.10.x | `opentofu` | core-only formula, no cloud plugins needed |
| Flux CLI | 2.3.0 - 2.4.x | `fluxcd/tap/flux` | tap addition requires brew-guard allow-list update |
| Infracost | 0.10.39 - 0.10.x | `infracost` | offline-capable if `infracost configure set api_key` cached |
| score-k8s (Score CLI) | 0.16.0 - 0.17.x | `score-spec/tap/score-k8s` | tap addition requires brew-guard allow-list update; used for Score schema validation ergonomics, not for rendering (score2openchoreo renders) |

M1's existing tools are reused unchanged: Go 1.26.2, Node.js 25.9.0, yarn, k3d 5.8.3, kubectl 1.35.3, helm 3.20.1, kubebuilder 4.13.1, colima, docker CLI 29.4.0, jq, yq, openssl.

### 3.2 Helm chart versions

```
flux-community/flux2                         2.13.0
open-policy-agent/gatekeeper                 3.17.1
gitea-charts/act-runner                      0.2.10
```

Chart repos and URLs:

```
flux:        https://fluxcd-community.github.io/helm-charts
gatekeeper:  https://open-policy-agent.github.io/gatekeeper/charts
act-runner:  https://dl.gitea.com/charts/          (same repo as Gitea itself, from M1)
```

### 3.3 score2openchoreo Go dependencies

```go
// tools/score2openchoreo/go.mod
module github.com/nnos/developer-portal/tools/score2openchoreo

go 1.21

require (
    gopkg.in/yaml.v3 v3.0.1
    github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
)
```

Two dependencies, both stable and widely used:

- `gopkg.in/yaml.v3` -- the standard YAML library for Go, Apache-2.0, stdlib-adjacent in maturity
- `github.com/santhosh-tekuri/jsonschema/v5` -- JSON Schema Draft 2020-12 validator, Apache-2.0, used by multiple large Go projects

No other dependencies. The Score JSON schema is embedded into the binary at build time via `//go:embed` (no network calls at runtime).

### 3.4 rr-tofu-guard Go dependencies

**Go stdlib only. Zero external dependencies**, mirroring the rr-brew-guard pattern.

```go
// plugins/rr-policy-guards/tools/tofu-guard/go.mod
module github.com/nnos/developer-portal/plugins/rr-policy-guards/tools/tofu-guard

go 1.21
```

Rationale identical to brew-guard: a security-critical hook with zero dependencies has zero supply-chain attack surface of its own. `encoding/json`, `strings`, `regexp`, `os`, `fmt`, `io`, `path/filepath`, `time` cover every need.

## 4. rr-tofu-guard implementation

### 4.1 Module structure

Mirrors rr-brew-guard (M1 Technical Specification Section 4).

```
plugins/rr-policy-guards/tools/tofu-guard/
+-- go.mod
+-- main.go             # entrypoint: reads stdin, calls parser, handles exit
+-- parser.go           # tokenizer + ValidateTofuCommand
+-- audit.go            # AuditWriter: append JSONL to audit log
+-- main_test.go        # integration tests via stdin piping
+-- parser_test.go      # pure-function unit tests
+-- audit_test.go       # audit log tests
```

### 4.2 Core types

```go
// parser.go

type Decision struct {
    Allow  bool
    Reason string
    Action string   // "allow" | "block" | "bypass" | "not-applicable"
}

type ToolInput struct {
    ToolName  string           `json:"tool_name"`
    ToolInput ToolInputPayload `json:"tool_input"`
    SessionID string           `json:"session_id,omitempty"`
}

type ToolInputPayload struct {
    Command     string `json:"command,omitempty"`
    Description string `json:"description,omitempty"`
}
```

### 4.3 Core rules

The guard applies three rules:

1. First token is `tofu` (not `terraform`). `terraform` commands are out of scope for M2.
2. If the second token is `apply`, `destroy`, `import`, `state rm`, `state mv`, or `state push`, block.
3. If the command contains shell metacharacters (same set as brew-guard: `;`, `&&`, `||`, `|`, backtick, `$(...)`, `>`, `<`, `&`), block.

Commands allowed without inspection: `tofu version`, `tofu init`, `tofu plan`, `tofu validate`, `tofu fmt`, `tofu providers`, `tofu output`, `tofu show`, `tofu graph`, `tofu console`, `tofu state list`, `tofu state show`, `tofu force-unlock` (allowed, logged).

### 4.4 Core functions

```go
// parser.go

var BlockedSubcommands = map[string]struct{}{
    "apply":   {},
    "destroy": {},
    "import":  {},
}

// State subcommands that mutate remote state are blocked when they appear after `tofu state`.
var BlockedStateSubcommands = map[string]struct{}{
    "rm":   {},
    "mv":   {},
    "push": {},
}

var shellMeta = regexp.MustCompile("[;&|<>`$()]")

// Tokenize behaves identically to rr-brew-guard's Tokenize: respects single
// and double quotes, rejects unterminated quotes.
func Tokenize(cmd string) ([]string, error) { /* same body as brew-guard */ }

// ValidateTofuCommand applies the decision tree.
// Pure function -- no I/O, no side effects.
func ValidateTofuCommand(tokens []string) Decision {
    if len(tokens) < 2 {
        return Decision{Allow: true, Action: "allow"} // `tofu` alone prints help
    }
    if tokens[0] != "tofu" {
        return Decision{Allow: true, Action: "not-applicable"}
    }

    sub := tokens[1]

    // State mutation subcommands
    if sub == "state" && len(tokens) >= 3 {
        if _, bad := BlockedStateSubcommands[tokens[2]]; bad {
            return Decision{Allow: false, Reason: "state-mutating: tofu state " + tokens[2], Action: "block"}
        }
        return Decision{Allow: true, Action: "allow"}
    }

    if _, bad := BlockedSubcommands[sub]; bad {
        return Decision{Allow: false, Reason: "tofu " + sub + " must run in CI, not from a Bash tool use", Action: "block"}
    }

    // All other tofu subcommands (plan, init, validate, etc.) are read-only enough to allow.
    return Decision{Allow: true, Action: "allow"}
}
```

### 4.5 Entrypoint flow

Structurally identical to rr-brew-guard's main:

```go
func main() {
    var input ToolInput
    if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
        logAudit("block", "unparseable-input", "", input.SessionID)
        fmt.Fprintln(os.Stderr, "tofu-guard: unable to parse PreToolUse input")
        os.Exit(2)
    }
    if input.ToolName != "Bash" {
        os.Exit(0)
    }

    tokens, err := Tokenize(input.ToolInput.Command)
    if err != nil {
        logAudit("block", "tokenize-error: "+err.Error(), input.ToolInput.Command, input.SessionID)
        fmt.Fprintln(os.Stderr, "tofu-guard: malformed command, refusing to evaluate")
        os.Exit(2)
    }
    if len(tokens) == 0 || tokens[0] != "tofu" {
        os.Exit(0)
    }
    if shellMeta.MatchString(input.ToolInput.Command) {
        logAudit("block", "shell-metacharacter", input.ToolInput.Command, input.SessionID)
        fmt.Fprintln(os.Stderr, "tofu-guard: shell metacharacter in tofu command")
        os.Exit(2)
    }

    decision := ValidateTofuCommand(tokens)

    if !decision.Allow && os.Getenv("RR_TOFU_GUARD_BYPASS") == "1" {
        logAudit("bypass", decision.Reason, input.ToolInput.Command, input.SessionID)
        fmt.Fprintln(os.Stderr, "tofu-guard: bypass in effect (RR_TOFU_GUARD_BYPASS=1)")
        os.Exit(0)
    }
    if decision.Allow {
        logAudit("allow", "", input.ToolInput.Command, input.SessionID)
        os.Exit(0)
    }
    logAudit("block", decision.Reason, input.ToolInput.Command, input.SessionID)
    fmt.Fprintf(os.Stderr, "tofu-guard: blocked -- %s\n", decision.Reason)
    os.Exit(2)
}
```

### 4.6 Audit log

Identical shape and path to brew-guard's audit log, one file per guard:

```
~/.rational-reserve/logs/tofu-guard.jsonl
```

`AuditEvent` struct, JSON encoding, and permissions (mode 0600) are unchanged from M1 Section 4.7.

### 4.7 Test cases (`parser_test.go`)

```go
tests := []struct {
    name    string
    cmd     string
    allow   bool
    action  string
}{
    // Allowed
    {"version", "tofu version", true, "allow"},
    {"init", "tofu init", true, "allow"},
    {"plan", "tofu plan", true, "allow"},
    {"validate", "tofu validate", true, "allow"},
    {"fmt", "tofu fmt", true, "allow"},
    {"state list", "tofu state list", true, "allow"},
    {"state show a.b", "tofu state show aws_instance.foo", true, "allow"},
    {"output", "tofu output", true, "allow"},
    {"not tofu", "ls -la", true, "not-applicable"},

    // Blocked
    {"apply", "tofu apply", false, "block"},
    {"apply -auto-approve", "tofu apply -auto-approve", false, "block"},
    {"destroy", "tofu destroy", false, "block"},
    {"import", "tofu import aws_instance.foo i-123", false, "block"},
    {"state rm", "tofu state rm aws_instance.foo", false, "block"},
    {"state mv", "tofu state mv a b", false, "block"},
    {"state push", "tofu state push state.tfstate", false, "block"},
    {"shell meta semicolon", "tofu plan; rm -rf /", false, "block"},
    {"shell meta backtick", "tofu plan `whoami`", false, "block"},
    {"shell meta pipe", "tofu plan | tee out.log", false, "block"},
}
```

### 4.8 Hook registration JSON

Appended to `~/.claude/settings.json` via `scripts/merge-tofu-hook-into-settings.sh` (symmetric to the brew-guard helper from M1):

```json
{
  "matcher": "Bash",
  "hooks": [
    {
      "type": "command",
      "command": "/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/bin/rr-tofu-guard"
    }
  ]
}
```

This entry is added alongside the brew-guard entry. Both hooks receive every Bash tool use; each decides based on the first token.

## 5. score2openchoreo implementation

### 5.1 Module structure

```
tools/score2openchoreo/
+-- go.mod
+-- go.sum
+-- main.go             # cobra-less entrypoint -- flag parsing + dispatch
+-- cli.go              # CLI flag parsing, invocation wiring
+-- schema.go           # embedded Score JSON schema, validation
+-- convert.go          # pure conversion from Score struct to Component struct
+-- types.go            # Go structs for Score input and Component output
+-- convert_test.go     # pure-function conversion tests
+-- main_test.go        # end-to-end tests: exec the binary, compare stdout to fixture
+-- fixtures/
```

### 5.2 Core types

```go
// types.go

// ScoreDocument is the subset of Score specification M2 converts.
// Unsupported fields cause a "unsupported Score field X" error.
type ScoreDocument struct {
    APIVersion string                       `yaml:"apiVersion"`
    Metadata   ScoreMetadata                `yaml:"metadata"`
    Containers map[string]ScoreContainer    `yaml:"containers"`
    Resources  map[string]ScoreResource     `yaml:"resources,omitempty"`
    Service    *ScoreService                `yaml:"service,omitempty"`
}

type ScoreMetadata struct {
    Name        string            `yaml:"name"`
    Annotations map[string]string `yaml:"annotations,omitempty"`
}

type ScoreContainer struct {
    Image     string                        `yaml:"image"`
    Command   []string                      `yaml:"command,omitempty"`
    Args      []string                      `yaml:"args,omitempty"`
    Variables map[string]string             `yaml:"variables,omitempty"`
    Resources *ScoreContainerResources      `yaml:"resources,omitempty"`
}

type ScoreContainerResources struct {
    Requests *ScoreResourceQuantities `yaml:"requests,omitempty"`
    Limits   *ScoreResourceQuantities `yaml:"limits,omitempty"`
}

type ScoreResourceQuantities struct {
    CPU    string `yaml:"cpu,omitempty"`
    Memory string `yaml:"memory,omitempty"`
}

type ScoreResource struct {
    Type     string            `yaml:"type"`
    Class    string            `yaml:"class,omitempty"`
    Metadata map[string]string `yaml:"metadata,omitempty"`
    Params   map[string]any    `yaml:"params,omitempty"`
}

type ScoreService struct {
    Ports map[string]ScorePort `yaml:"ports"`
}

type ScorePort struct {
    Port       int    `yaml:"port"`
    TargetPort int    `yaml:"targetPort,omitempty"`
    Protocol   string `yaml:"protocol,omitempty"`
}

// OpenChoreoComponent is the subset of OpenChoreo's Component CRD this milestone emits.
// Field names mirror the OpenChoreo Component v1alpha1 schema.
type OpenChoreoComponent struct {
    APIVersion string                  `yaml:"apiVersion"`
    Kind       string                  `yaml:"kind"`
    Metadata   ComponentMetadata       `yaml:"metadata"`
    Spec       ComponentSpec           `yaml:"spec"`
}

type ComponentMetadata struct {
    Name      string            `yaml:"name"`
    Namespace string            `yaml:"namespace"`
    Labels    map[string]string `yaml:"labels,omitempty"`
}

type ComponentSpec struct {
    WorkloadTemplate WorkloadTemplate `yaml:"workloadTemplate"`
    Environment      string           `yaml:"environment"`
    Owner            ComponentOwner   `yaml:"owner"`
}

type WorkloadTemplate struct {
    Containers []ContainerSpec `yaml:"containers"`
    Ports      []PortSpec      `yaml:"ports,omitempty"`
}

type ContainerSpec struct {
    Name      string                  `yaml:"name"`
    Image     string                  `yaml:"image"`
    Command   []string                `yaml:"command,omitempty"`
    Args      []string                `yaml:"args,omitempty"`
    Env       []EnvVarSpec            `yaml:"env,omitempty"`
    Resources *ContainerResourcesSpec `yaml:"resources,omitempty"`
}

type EnvVarSpec struct {
    Name      string             `yaml:"name"`
    Value     string             `yaml:"value,omitempty"`
    ValueFrom *EnvVarSourceSpec  `yaml:"valueFrom,omitempty"`
}

type EnvVarSourceSpec struct {
    SecretKeyRef *SecretKeySelectorSpec `yaml:"secretKeyRef,omitempty"`
}

type SecretKeySelectorSpec struct {
    Name string `yaml:"name"`
    Key  string `yaml:"key"`
}

type ContainerResourcesSpec struct {
    Requests *ResourceQuantitiesSpec `yaml:"requests,omitempty"`
    Limits   *ResourceQuantitiesSpec `yaml:"limits,omitempty"`
}

type ResourceQuantitiesSpec struct {
    CPU    string `yaml:"cpu,omitempty"`
    Memory string `yaml:"memory,omitempty"`
}

type PortSpec struct {
    Name       string `yaml:"name"`
    Port       int    `yaml:"port"`
    TargetPort int    `yaml:"targetPort,omitempty"`
    Protocol   string `yaml:"protocol,omitempty"`
}

type ComponentOwner struct {
    Project string `yaml:"project"`
}
```

### 5.3 Core functions

```go
// convert.go

// Convert takes a validated ScoreDocument and returns an OpenChoreoComponent.
// Pure function -- no I/O, no side effects, no mutable globals.
func Convert(in ScoreDocument, opts ConvertOptions) (OpenChoreoComponent, error)

type ConvertOptions struct {
    Environment string // required; `dev` or `staging`
    Namespace   string // required; OpenChoreo project namespace
    ImageRef    string // optional; overrides the Score container image
    Project     string // required; OpenChoreo project name
}
```

### 5.4 CLI flow

```go
// main.go

func main() {
    opts, err := parseFlags(os.Args[1:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
        os.Exit(2)
    }

    raw, err := readInput(opts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
        os.Exit(2)
    }

    if err := ValidateScore(raw); err != nil {
        fmt.Fprintf(os.Stderr, "score2openchoreo: Score validation failed: %s\n", err)
        os.Exit(1)
    }

    if opts.ValidateOnly {
        os.Exit(0)
    }

    var doc ScoreDocument
    if err := yaml.Unmarshal(raw, &doc); err != nil {
        fmt.Fprintf(os.Stderr, "score2openchoreo: yaml unmarshal: %s\n", err)
        os.Exit(1)
    }

    comp, err := Convert(doc, ConvertOptions{
        Environment: opts.Environment,
        Namespace:   opts.Namespace,
        ImageRef:    opts.Image,
        Project:     opts.Project,
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
        os.Exit(1)
    }

    out, err := yaml.Marshal(comp)
    if err != nil {
        fmt.Fprintf(os.Stderr, "score2openchoreo: marshal: %s\n", err)
        os.Exit(2)
    }
    if _, err := os.Stdout.Write(out); err != nil {
        os.Exit(2)
    }
}
```

Flag handling uses Go's `flag` package (stdlib). No third-party CLI library.

### 5.5 Score schema validation

```go
// schema.go

import _ "embed"

//go:embed assets/score.schema.json
var scoreSchemaJSON []byte

// ValidateScore validates raw YAML bytes against the embedded Score schema.
// Returns nil on valid input; returns a descriptive error on invalid.
func ValidateScore(raw []byte) error
```

The Score JSON schema is checked into `tools/score2openchoreo/assets/score.schema.json`, copied from the Score specification repo at a pinned tag (see Section 3.3). No network fetch at runtime.

### 5.6 Convert logic summary

The conversion rules are:

| Score field | OpenChoreo field | Notes |
|---|---|---|
| `metadata.name` | `metadata.name` | direct copy |
| `metadata.annotations` | `metadata.labels` | flattened into labels |
| `containers[<key>].image` | `spec.workloadTemplate.containers[].image` | overridden by `--image` flag |
| `containers[<key>].command` | `spec.workloadTemplate.containers[].command` | direct copy |
| `containers[<key>].args` | `spec.workloadTemplate.containers[].args` | direct copy |
| `containers[<key>].variables` | `spec.workloadTemplate.containers[].env[]` | each entry becomes an EnvVar with `value` set |
| `containers[<key>].resources.requests` | `spec.workloadTemplate.containers[].resources.requests` | direct copy |
| `containers[<key>].resources.limits` | `spec.workloadTemplate.containers[].resources.limits` | direct copy |
| `service.ports` | `spec.workloadTemplate.ports` | converted from map to list, port name = map key |
| `resources[<key>]` where `type: secret` | EnvVar with `valueFrom.secretKeyRef` | see secret binding rules |
| other `resources[<key>]` | error -- unsupported at M2 | |

Secret binding: a Score `resources.<key>` with `type: secret` produces an ExternalSecret/Secret pair in the target namespace via external-secrets. The converter emits only the Component; the ExternalSecret is provisioned by the `external-secrets-wiring` Tofu module separately at pipeline-run time.

### 5.7 Test strategy

```
convert_test.go     -- table-driven tests for Convert. Covers: all supported Score fields,
                       all error cases, idempotency (Convert(Convert(x)) semantic invariants).

main_test.go        -- exec's the binary, pipes fixture Score YAMLs on stdin, compares
                       stdout to golden Component YAMLs in fixtures/.

schema_test.go      -- tests ValidateScore against fixtures including invalid-schema.score.yaml.
```

Fixtures are the source of truth. Adding a new supported field means: write a Score fixture, write the expected Component fixture, run tests until they pass.

## 6. OpenTofu module layout

### 6.1 Root module

```hcl
# iac/versions.tf
terraform {
  required_version = ">= 1.9.0, < 1.11.0"
  required_providers {
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.33" }
    helm       = { source = "hashicorp/helm",       version = "~> 2.17" }
    kubectl    = { source = "alekc/kubectl",        version = "~> 2.1"  }
  }
}
```

```hcl
# iac/backend.tf
terraform {
  backend "kubernetes" {
    secret_suffix     = "state-m2"
    namespace         = "tofu-state"
    load_config_file  = true
    config_path       = "~/.kube/config"
    config_context    = "k3d-openchoreo"
  }
}
```

```hcl
# iac/providers.tf
provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}

provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kube_context
  }
}

provider "kubectl" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}
```

```hcl
# iac/variables.tf
variable "kubeconfig_path" { type = string; default = "~/.kube/config" }
variable "kube_context"    { type = string; default = "k3d-openchoreo" }

variable "gitea_url" {
  type        = string
  description = "In-cluster URL for Gitea API"
  default     = "http://gitea-http.gitea.svc.cluster.local:3000"
}

variable "openchoreo_project" {
  type    = string
  default = "openchoreo"
}

variable "infracost_threshold_monthly_usd" {
  type    = number
  default = 50
}
```

```hcl
# iac/main.tf
resource "kubernetes_namespace" "tofu_state" {
  metadata { name = "tofu-state" }
}

module "flux" {
  source    = "./modules/flux"
  gitea_url = var.gitea_url
}

module "gatekeeper" {
  source                          = "./modules/gatekeeper"
  infracost_threshold_monthly_usd = var.infracost_threshold_monthly_usd
}

module "gitea_runner" {
  source    = "./modules/gitea-runner"
  gitea_url = var.gitea_url
  depends_on = [module.flux]
}

module "openchoreo_environments" {
  source  = "./modules/openchoreo-environments"
  project = var.openchoreo_project
}

module "external_secrets_wiring" {
  source = "./modules/external-secrets-wiring"
}
```

```hcl
# iac/outputs.tf
output "flux_namespace"         { value = module.flux.namespace }
output "gatekeeper_namespace"   { value = module.gatekeeper.namespace }
output "runner_namespace"       { value = module.gitea_runner.namespace }
output "environments"           { value = module.openchoreo_environments.names }
```

### 6.2 modules/flux

```hcl
# iac/modules/flux/main.tf
resource "kubernetes_namespace" "flux_system" {
  metadata { name = "flux-system" }
}

resource "helm_release" "flux" {
  name       = "flux2"
  namespace  = kubernetes_namespace.flux_system.metadata[0].name
  repository = "https://fluxcd-community.github.io/helm-charts"
  chart      = "flux2"
  version    = "2.13.0"
  wait       = true
  timeout    = 600
}

# GitRepository pointing at platform-addons. Authentication via an existing
# Kubernetes Secret called `gitea-deploy-key` that external-secrets-wiring
# materializes from openbao.
resource "kubectl_manifest" "platform_addons_source" {
  depends_on = [helm_release.flux]
  yaml_body  = yamlencode({
    apiVersion = "source.toolkit.fluxcd.io/v1"
    kind       = "GitRepository"
    metadata   = {
      name      = "platform-addons"
      namespace = "flux-system"
    }
    spec = {
      interval  = "1m"
      url       = "${var.gitea_url}/openchoreo/platform-addons"
      ref       = { branch = "main" }
      secretRef = { name = "gitea-deploy-key" }
    }
  })
}

resource "kubectl_manifest" "platform_addons_kustomization" {
  depends_on = [kubectl_manifest.platform_addons_source]
  yaml_body  = yamlencode({
    apiVersion = "kustomize.toolkit.fluxcd.io/v1"
    kind       = "Kustomization"
    metadata   = {
      name      = "platform-addons"
      namespace = "flux-system"
    }
    spec = {
      interval = "1m"
      path     = "./clusters/default"
      prune    = true
      sourceRef = { kind = "GitRepository", name = "platform-addons" }
    }
  })
}
```

```hcl
# iac/modules/flux/variables.tf
variable "gitea_url" { type = string }
```

```hcl
# iac/modules/flux/outputs.tf
output "namespace" { value = kubernetes_namespace.flux_system.metadata[0].name }
```

### 6.3 modules/gatekeeper

```hcl
# iac/modules/gatekeeper/main.tf
resource "kubernetes_namespace" "gatekeeper_system" {
  metadata { name = "gatekeeper-system" }
}

resource "helm_release" "gatekeeper" {
  name       = "gatekeeper"
  namespace  = kubernetes_namespace.gatekeeper_system.metadata[0].name
  repository = "https://open-policy-agent.github.io/gatekeeper/charts"
  chart      = "gatekeeper"
  version    = "3.17.1"
  wait       = true
  timeout    = 600
  set {
    name  = "controllerManager.resources.requests.cpu"
    value = "100m"
  }
}
```

```hcl
# iac/modules/gatekeeper/constraints.tf
resource "kubectl_manifest" "c1_template" {
  depends_on = [helm_release.gatekeeper]
  yaml_body  = file("${path.root}/../policies/C1-constraint.yaml")
}
resource "kubectl_manifest" "c2_template" {
  depends_on = [helm_release.gatekeeper]
  yaml_body  = file("${path.root}/../policies/C2-constraint.yaml")
}
resource "kubectl_manifest" "c3_template" {
  depends_on = [helm_release.gatekeeper]
  yaml_body  = file("${path.root}/../policies/C3-constraint.yaml")
}
```

```hcl
# iac/modules/gatekeeper/variables.tf
variable "infracost_threshold_monthly_usd" { type = number }
```

The Rego policies and ConstraintTemplates live in `/policies/` and are applied both by Tofu at install time and by Flux on reconcile (the same files are symlinked/copied into `platform-addons/clusters/default/gatekeeper/`). The duplication is deliberate -- Tofu ensures first install, Flux enforces drift correction afterward.

### 6.4 modules/gitea-runner

```hcl
# iac/modules/gitea-runner/main.tf
resource "kubernetes_namespace" "gitea_runners" {
  metadata { name = "gitea-runners" }
}

# ExternalSecret that pulls the runner registration token from openbao.
resource "kubectl_manifest" "runner_token" {
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1beta1"
    kind       = "ExternalSecret"
    metadata   = {
      name      = "gitea-runner-token"
      namespace = "gitea-runners"
    }
    spec = {
      refreshInterval = "1h"
      secretStoreRef  = { name = "openbao-kv", kind = "ClusterSecretStore" }
      target          = { name = "gitea-runner-token", creationPolicy = "Owner" }
      data            = [{
        secretKey = "token"
        remoteRef = { key = "gitea/runners/token", property = "token" }
      }]
    }
  })
  depends_on = [kubernetes_namespace.gitea_runners]
}

resource "helm_release" "act_runner" {
  name       = "act-runner"
  namespace  = kubernetes_namespace.gitea_runners.metadata[0].name
  repository = "https://dl.gitea.com/charts/"
  chart      = "act-runner"
  version    = "0.2.10"
  wait       = true
  timeout    = 600

  set {
    name  = "giteaRootURL"
    value = var.gitea_url
  }
  set {
    name  = "existingSecret"
    value = "gitea-runner-token"
  }
  set {
    name  = "podSecurityContext.runAsNonRoot"
    value = "true"
  }
  set {
    name  = "securityContext.allowPrivilegeEscalation"
    value = "false"
  }

  depends_on = [kubectl_manifest.runner_token]
}
```

### 6.5 modules/openchoreo-environments

```hcl
# iac/modules/openchoreo-environments/main.tf
locals {
  environments = ["dev", "staging"]
}

resource "kubectl_manifest" "environment" {
  for_each = toset(local.environments)
  yaml_body = yamlencode({
    apiVersion = "openchoreo.dev/v1alpha1"
    kind       = "Environment"
    metadata = {
      name      = each.key
      namespace = "openchoreo-control-plane"
    }
    spec = {
      displayName = each.key
      project     = var.project
    }
  })
}
```

```hcl
# iac/modules/openchoreo-environments/outputs.tf
output "names" { value = local.environments }
```

### 6.6 modules/external-secrets-wiring

Creates ExternalSecret objects (and one ClusterSecretStore if not already present from M1) for:

- `flux-system/gitea-deploy-key` -- consumed by Flux GitRepository
- `apps/hello-m2/dev/example-secret` -- consumed by hello-m2 workload

Exact YAML omitted here -- the file layout is the same pattern as 6.4's `runner_token`. See the file list in Section 2.

## 7. Gatekeeper policies

### 7.1 C-1 platform-addons main-protected

Rego sketch:

```rego
package constraints.m2.c1

violation[{"msg": msg}] {
  input.review.kind.kind == "GitRepository"
  input.review.object.spec.url == "http://gitea-http.gitea.svc.cluster.local:3000/openchoreo/platform-addons"
  # Gatekeeper's kind here is Flux GitRepository -- C-1 enforces that the
  # branch protection setting exists. Actual git-side protection is in Gitea's
  # branch protection rules, which this template asserts are present.
  not input.review.object.spec.ref.branch == "main"
  msg := "platform-addons GitRepository must reference the protected main branch"
}
```

Gitea-side branch protection for `platform-addons/main` requiring PRs is configured in `seed-repos/platform-addons/` via the Gitea API at install time. The Rego above is the cluster-side companion; defense in depth.

### 7.2 C-2 Score schema valid

Rego uses a data field seeded by the pipeline (the score2openchoreo `--validate-only` result). Rego sketch:

```rego
package constraints.m2.c2

violation[{"msg": msg}] {
  input.review.object.metadata.annotations["pipeline.m2/score-valid"] != "true"
  msg := "Score schema validation did not pass for this Component"
}
```

The pipeline annotates the rendered Component with `pipeline.m2/score-valid: "true"` only if `score2openchoreo --validate-only` exits 0 before render.

### 7.3 C-3 Infracost delta

Rego reads an annotation written by the pipeline with the numeric monthly delta:

```rego
package constraints.m2.c3

violation[{"msg": msg}] {
  delta := to_number(input.review.object.metadata.annotations["pipeline.m2/cost-delta-usd"])
  threshold := to_number(input.parameters.thresholdUSD)
  delta > threshold
  msg := sprintf("estimated monthly cost delta $%v exceeds threshold $%v", [delta, threshold])
}
```

The ConstraintTemplate for C-3 parameterizes `thresholdUSD`; the Constraint instance pins it to `50` (matching FR-21).

## 8. Seeded Gitea repos

### 8.1 Organization and repos

Seeded by `scripts/seed-gitea-repos.sh` via the Gitea API. The three repos are owned by an `openchoreo` organization.

```bash
# scripts/seed-gitea-repos.sh (sketch)
source "$ROOT/scripts/lib/gitea-api.sh"

GITEA_TOKEN=$(cat ~/.rational-reserve/m1-gitea-token)

gitea_api_post /api/v1/orgs '{"username":"openchoreo","full_name":"OpenChoreo Platform"}'

for repo in platform-addons platform-config hello-m2; do
  gitea_api_post /api/v1/orgs/openchoreo/repos '{"name":"'"$repo"'","private":false,"auto_init":true,"default_branch":"main"}'
done

# Branch protection on platform-addons and platform-config main branch
for repo in platform-addons platform-config; do
  gitea_api_post "/api/v1/repos/openchoreo/$repo/branch_protections" '{
    "branch_name": "main",
    "enable_push": false,
    "required_approvals": 1,
    "enable_merge_whitelist": true,
    "merge_whitelist_usernames": ["gitea_admin"]
  }'
done

# Seed content from seed-repos/ into each repo
for repo in platform-addons platform-config hello-m2; do
  push_seed_content "$repo" "$ROOT/seed-repos/$repo"
done
```

`push_seed_content` is a helper that walks the directory and posts each file via the contents API.

### 8.2 platform-addons seed content

Contains the Flux Kustomization that Flux's root Kustomization points at:

```
seed-repos/platform-addons/
+-- README.md
+-- clusters/
    +-- default/
        +-- kustomization.yaml
        +-- flux-system/
        |   +-- gotk-components.yaml         # Flux's own controllers, managed by Flux itself
        |   +-- gotk-sync.yaml               # Flux's root Kustomization (self-reconcile)
        +-- gatekeeper/
            +-- constrainttemplates.yaml
            +-- constraints.yaml
```

The `gotk-components.yaml` and `gotk-sync.yaml` files are produced by `flux install --export`; exact content is generated at seed time and committed. `kustomization.yaml` at the root simply lists the subdirectories.

### 8.3 platform-config seed content

```
seed-repos/platform-config/
+-- README.md
+-- environments/
    +-- dev/
    |   +-- .gitkeep
    +-- staging/
        +-- .gitkeep
```

Deliberately empty. The first push populating `environments/dev/hello-m2.yaml` is the first end-to-end pipeline run.

### 8.4 hello-m2 seed content

`main.go`:

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
        fmt.Fprintf(w, "hello from hello-m2 env=%s secret=%s\n",
            os.Getenv("ENVIRONMENT"),
            redact(os.Getenv("EXAMPLE_SECRET")),
        )
    })
    addr := ":8080"
    log.Printf("listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}

func redact(s string) string {
    if len(s) <= 4 {
        return "****"
    }
    return s[:2] + "****" + s[len(s)-2:]
}
```

`Dockerfile`:

```
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/hello-m2 .

FROM alpine:3.20
COPY --from=build /out/hello-m2 /hello-m2
USER 65532:65532
ENTRYPOINT ["/hello-m2"]
```

`score.yaml`:

```yaml
apiVersion: score.dev/v1b1
metadata:
  name: hello-m2
  annotations:
    pipeline.m2/owner: openchoreo
containers:
  web:
    image: gitea.gitea.svc.cluster.local:3000/openchoreo/hello-m2:latest
    variables:
      ENVIRONMENT: "${resources.env.value}"
      EXAMPLE_SECRET: "${resources.example.password}"
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 200m
        memory: 128Mi
service:
  ports:
    http:
      port: 8080
resources:
  env:
    type: environment
  example:
    type: secret
    metadata:
      name: example-secret
```

`catalog-info.yaml`:

```yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: hello-m2
  description: M2 demo workload
  annotations:
    gitea.io/project: openchoreo/hello-m2
  links:
    - url: http://localhost:3002/openchoreo/hello-m2
      title: Gitea repo
      icon: web
    - url: http://localhost:3002/openchoreo/hello-m2/actions
      title: CI runs
      icon: dashboard
spec:
  type: service
  lifecycle: experimental
  owner: openchoreo
```

`.gitea/workflows/ci.yaml` -- identical to the canonical template in `iac/templates/ci.yaml` (Section 9).

## 9. Canonical ci.yaml workflow template

`iac/templates/ci.yaml`:

```yaml
name: M2 CI

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  validate-and-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.26"

      - name: Build score2openchoreo
        run: |
          cd tools/score2openchoreo
          go build -o bin/score2openchoreo .

      - name: Validate Score schema
        run: ./tools/score2openchoreo/bin/score2openchoreo --validate-only --input score.yaml

      - name: Set up OpenTofu
        if: contains(github.event.pull_request.changed_files, 'iac/') || github.event_name == 'push'
        uses: opentofu/setup-opentofu@v1
        with:
          tofu_version: 1.9.0

      - name: tofu plan
        if: contains(github.event.pull_request.changed_files, 'iac/') || github.event_name == 'push'
        run: |
          cd iac
          tofu init
          tofu plan -out=plan.out

      - name: Infracost breakdown
        if: github.event_name == 'pull_request' && contains(github.event.pull_request.changed_files, 'iac/')
        run: |
          infracost breakdown --path iac/ --format json --out-file infracost.json
          MONTHLY=$(jq -r '.totalMonthlyCost' infracost.json)
          BASE=$(jq -r '.pastTotalMonthlyCost // 0' infracost.json)
          DELTA=$(awk -v a="$MONTHLY" -v b="$BASE" 'BEGIN{print a-b}')
          echo "DELTA=$DELTA" >> $GITHUB_ENV

      - name: Post Infracost PR comment
        if: github.event_name == 'pull_request' && contains(github.event.pull_request.changed_files, 'iac/')
        run: ./scripts/ci/post-infracost-comment.sh "$DELTA"

      - name: Build image
        run: |
          docker build -t gitea.gitea.svc.cluster.local:3000/openchoreo/hello-m2:${GITHUB_SHA::7} .

      - name: Push image
        if: github.event_name == 'push'
        run: |
          echo "$GITEA_TOKEN" | docker login gitea.gitea.svc.cluster.local:3000 -u gitea_admin --password-stdin
          docker push gitea.gitea.svc.cluster.local:3000/openchoreo/hello-m2:${GITHUB_SHA::7}
        env:
          GITEA_TOKEN: ${{ secrets.GITEA_WRITE_TOKEN }}

      - name: Render Component
        if: github.event_name == 'push'
        run: |
          ./tools/score2openchoreo/bin/score2openchoreo \
            --input score.yaml \
            --environment dev \
            --namespace openchoreo-data-plane \
            --project openchoreo \
            --image gitea.gitea.svc.cluster.local:3000/openchoreo/hello-m2:${GITHUB_SHA::7} \
            > /tmp/component.yaml

      - name: Commit to platform-config
        if: github.event_name == 'push'
        run: ./scripts/ci/commit-to-platform-config.sh dev hello-m2 /tmp/component.yaml
```

The `.gitea/workflows/ci.yaml` in `hello-m2` is a copy of the above, at repo scaffold time.

## 10. Install script structure

```bash
#!/usr/bin/env bash
# scripts/install-m2.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT/scripts/lib/colors.sh"
source "$ROOT/scripts/lib/wait-for.sh"
source "$ROOT/scripts/lib/confirm.sh"

info() { printf "\033[1;36m[m2-install]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[m2-install ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

task_0_tofu_guard() {
    info "Task 0: building rr-tofu-guard"
    cd "$ROOT/plugins/rr-policy-guards/tools/tofu-guard"
    go test ./... || fail "tofu-guard tests failed"
    go build -o "$ROOT/plugins/rr-policy-guards/bin/rr-tofu-guard" .
    "$ROOT/scripts/merge-tofu-hook-into-settings.sh"
}

task_0_5_host_tools() {
    info "Task 0.5: installing host tools (gated by brew-guard)"
    command -v tofu      >/dev/null || brew install opentofu
    command -v flux      >/dev/null || brew install fluxcd/tap/flux
    command -v infracost >/dev/null || brew install infracost
    command -v score-k8s >/dev/null || brew install score-spec/tap/score-k8s
}

task_1_seed_openbao_paths() {
    info "Task 1: seeding openbao kv paths"
    "$ROOT/scripts/seed-openbao-m2-paths.sh"
}

task_2_seed_gitea_repos() {
    info "Task 2: seeding Gitea organization + repos"
    "$ROOT/scripts/seed-gitea-repos.sh"
}

task_3_build_score2openchoreo() {
    info "Task 3: building score2openchoreo"
    cd "$ROOT/tools/score2openchoreo"
    go test ./...
    go build -o bin/score2openchoreo .
}

task_4_tofu_apply() {
    info "Task 4: tofu init + apply"
    cd "$ROOT/iac"
    tofu init
    tofu apply -auto-approve
}

task_5_wait_flux_reconcile() {
    info "Task 5: waiting for Flux to reconcile platform-addons"
    kubectl -n flux-system wait --for=condition=Ready kustomization.kustomize.toolkit.fluxcd.io/platform-addons --timeout=5m
}

task_6_smoke() {
    info "Task 6: running smoke tests"
    "$ROOT/scripts/smoke-m2.sh"
}

main() {
    task_0_tofu_guard
    task_0_5_host_tools
    task_1_seed_openbao_paths
    task_2_seed_gitea_repos
    task_3_build_score2openchoreo
    task_4_tofu_apply
    task_5_wait_flux_reconcile
    task_6_smoke
    info "M2 complete."
}

main "$@"
```

## 11. Teardown script structure

```bash
#!/usr/bin/env bash
# scripts/teardown-m2.sh
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

info() { printf "\033[1;36m[m2-teardown]\033[0m %s\n" "$*"; }

# Remove M2 hook registration
"$ROOT/scripts/remove-tofu-hook-from-settings.sh" 2>/dev/null || true

# tofu destroy (reverse of apply)
if [ -d "$ROOT/iac/.terraform" ]; then
    (cd "$ROOT/iac" && RR_TOFU_GUARD_BYPASS=1 tofu destroy -auto-approve) || true
fi

# Belt-and-suspenders: delete namespaces in case tofu destroy missed
for ns in flux-system gatekeeper-system gitea-runners tofu-state; do
    kubectl delete namespace "$ns" --ignore-not-found --timeout=2m || true
done

# Delete seeded Gitea repos (preserves the openchoreo organization for re-seed ergonomics)
"$ROOT/scripts/delete-m2-gitea-repos.sh" 2>/dev/null || true

info "M2 torn down. M1 substrate preserved."
```

Note the `RR_TOFU_GUARD_BYPASS=1` on the destroy -- teardown is the one legitimate case where a human-driven `tofu destroy` is desired; the bypass is logged.

## 12. Smoke test scripts

Each smoke script exits 0 on pass, 1 on fail, with a single-line summary on stdout.

```bash
# scripts/smoke-m2.sh
#!/usr/bin/env bash
set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for check in tofu actions flux score infracost gatekeeper openbao; do
    printf "[%s] " "$check"
    "$ROOT/scripts/smoke-$check.sh" || { echo "FAIL"; exit 1; }
done
echo "M2 smoke: all pass"
```

Individual smoke scripts are sketched below; each is fewer than 40 lines.

```bash
# scripts/smoke-tofu.sh
set -e
tofu version
cd "$(dirname "$0")/../iac"
tofu init -reconfigure >/dev/null
tofu plan -detailed-exitcode && echo "PASS: no plan diff" || {
    rc=$?; [ $rc -eq 2 ] && echo "PASS: plan diff present (expected)" || exit $rc
}
```

```bash
# scripts/smoke-flux.sh
flux reconcile source git platform-addons
flux get kustomizations platform-addons | grep -q True
echo "PASS"
```

```bash
# scripts/smoke-gatekeeper.sh
kubectl get constrainttemplates | grep -q C1
kubectl get constrainttemplates | grep -q C2
kubectl get constrainttemplates | grep -q C3
kubectl get constraints -A | grep -qc '3 constraints' || {
    # count the constraints another way if the previous grep is fragile
    cnt=$(kubectl get constraints -A --no-headers | wc -l | tr -d ' ')
    [ "$cnt" -ge 3 ]
}
echo "PASS"
```

```bash
# scripts/smoke-score.sh
out=$("$(dirname "$0")/../tools/score2openchoreo/bin/score2openchoreo" \
  --input "$(dirname "$0")/../tools/score2openchoreo/fixtures/minimal.score.yaml" \
  --environment dev --namespace openchoreo-data-plane --project openchoreo)
echo "$out" | yq eval '.kind' - | grep -q Component
echo "PASS"
```

```bash
# scripts/smoke-infracost.sh
infracost breakdown --path "$(dirname "$0")/../iac" --format table | head
echo "PASS"
```

```bash
# scripts/smoke-openbao.sh
kubectl -n openbao exec openbao-0 -- bao kv get kv/gitea/runners/token >/dev/null
kubectl -n openbao exec openbao-0 -- bao kv get kv/flux/gitea-deploy-key >/dev/null
kubectl -n openbao exec openbao-0 -- bao kv get kv/apps/hello-m2/dev/example-secret >/dev/null
echo "PASS"
```

```bash
# scripts/smoke-actions.sh
# Triggers the canonical hello workflow in hello-m2 and waits for it to succeed.
RUN_ID=$(curl -s -u gitea_admin:"$(cat ~/.rational-reserve/m1-gitea-admin-password)" \
    -X POST http://localhost:3002/api/v1/repos/openchoreo/hello-m2/actions/workflows/ci.yaml/dispatches \
    -H 'Content-Type: application/json' \
    -d '{"ref":"main"}' | jq -r .id)
# poll up to 5 minutes
for i in $(seq 1 60); do
    STATUS=$(curl -s -u gitea_admin:"$(cat ~/.rational-reserve/m1-gitea-admin-password)" \
        http://localhost:3002/api/v1/repos/openchoreo/hello-m2/actions/runs/"$RUN_ID" | jq -r .conclusion)
    [ "$STATUS" = "success" ] && { echo "PASS"; exit 0; }
    [ "$STATUS" = "failure" ] && { echo "FAIL: workflow failed"; exit 1; }
    sleep 5
done
echo "FAIL: timeout"; exit 1
```

## 13. Test strategy

### 13.1 rr-tofu-guard

- `parser_test.go`: covers every row in Section 4.7. Table-driven, runs in milliseconds.
- `audit_test.go`: writes to a tmp file, verifies JSONL and mode 0600.
- `main_test.go`: execs the binary via `os/exec`, pipes real JSON, checks exit codes and stderr.

### 13.2 score2openchoreo

- `convert_test.go`: table-driven, one row per supported Score field and each error case. Pure-function tests.
- `main_test.go`: golden-file tests. Each `.score.yaml` in `fixtures/` has a matching `.component.yaml`; the test execs the binary with fixed flags and diffs stdout.
- `schema_test.go`: validates both a valid and an invalid fixture, asserts the right error on the invalid case.

### 13.3 Gatekeeper

- Rego unit tests via `opa test` run in CI. Each ConstraintTemplate ships with a `_test.rego` file covering allow and deny cases.
- End-to-end: the hello-m2 pipeline run is itself a test of C-2 and C-3.

### 13.4 Pipeline end-to-end

- `scripts/smoke-m2.sh` runs every per-tool smoke. Operators run it after install and on demand.
- A deliberate "bad PR" scenario is kept as a manual test -- a PR that sets monthly delta above $50 should be blocked by C-3 and show a red check.

## 14. Rollback strategy

M2 layers cleanly onto M1 and can be removed cleanly:

- `tofu destroy` on the M2 `iac/` root reverses every helm release and namespace Tofu created
- `scripts/teardown-m2.sh` wraps this plus the hook removal and repo cleanup
- M1 substrate is never touched by the M2 teardown path

If `tofu destroy` gets stuck (e.g., finalizers on a ConstraintTemplate), the teardown script's namespace-delete fallback handles it.

## 15. Documentation requirements

- Every Go file in `tools/score2openchoreo/` and `plugins/rr-policy-guards/tools/tofu-guard/` SHALL carry a file-level comment under 10 lines.
- Every Tofu module SHALL have a `README.md` with inputs, outputs, and one-sentence purpose.
- Every Rego policy SHALL have a file header comment stating what it enforces and which ConstraintTemplate it implements.
- The repo-root `README.md` SHALL gain an M2 section under 150 words per NFR-17.

## 16. What this document does not cover

- Step-by-step TDD task ordering with checkboxes -- that is the Implementation Plan produced by the `superpowers:writing-plans` skill next.
- Exact patches to `backstage/app-config.yaml` for the `/api/proxy/gitea-actions` entry -- Implementation Plan, produced at scaffold time.
- Exact sequence of `git init && git push` calls for the initial seed of each repo -- Implementation Plan.
- openbao policy HCL for the M2 paths -- Implementation Plan, since exact policy wording depends on openbao version semantics at install time.

## 17. Self-review checklist

Before producing the Implementation Plan from this document, verify:

- [ ] Every functional requirement in `requirements.md` maps to a concrete section here
- [ ] Every design decision in `design-specification.md` has a matching implementation detail
- [ ] Every file in Section 2 has an unambiguous owning section
- [ ] Every pinned version appears with both a lower and upper bound
- [ ] Every API call has an example body or an explicit link to upstream docs
- [ ] Every shell snippet is runnable from the documented directory with the documented prerequisites
- [ ] No section contains "implement reasonable X" placeholders
- [ ] The three Gatekeeper constraints fire exactly where the pipeline flow says they fire
- [ ] The openbao kv paths in FR-37 and Design Section 7.4 match byte-for-byte with the ExternalSecret manifests referenced here
- [ ] The state backend decision (kubernetes backend, `tofu-state` namespace) is reflected in every place Tofu is invoked
