// types.go -- shared types for rr-verify-guard.
//
// All cross-module types live here so each *.go file can pull from a
// single, version-stable surface. Keep this file small; if a type is
// only used in one module, define it next to that code.

package main

const (
	exitAllow    = 0
	exitDegraded = 1
	exitBlock    = 2

	guardVersion = "0.2.0"

	cacheSchema = "v2"
)

// ToolInput mirrors the shape Claude Code's PreToolUse hook passes on
// stdin. Identical to the struct used in sibling guards.
type ToolInput struct {
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
	SessionID string         `json:"session_id,omitempty"`
}

// Toolchain identifies a language/build stack detected in a repo.
type Toolchain string

const (
	ToolchainGo        Toolchain = "go"
	ToolchainNode      Toolchain = "node"
	ToolchainRust      Toolchain = "rust"
	ToolchainPython    Toolchain = "python"
	ToolchainCpp       Toolchain = "cpp"
	ToolchainMake      Toolchain = "make"
	ToolchainElixir    Toolchain = "elixir"
	ToolchainRuby      Toolchain = "ruby"
	ToolchainJVMMaven  Toolchain = "jvm-mvn"
	ToolchainJVMGradle Toolchain = "jvm-gradle"
	ToolchainSecurity  Toolchain = "security"
)

// Forge identifies the SCM platform backing a repo.
type Forge string

const (
	ForgeGitHub  Forge = "github"
	ForgeGitea   Forge = "gitea"
	ForgeForgejo Forge = "forgejo"
	ForgeUnknown Forge = "unknown"
	ForgeNone    Forge = "none"
)

// Step is one verification command to run.
type Step struct {
	Toolchain Toolchain
	Name      string
	Cmd       string
	Args      []string
	Required  bool
	// WorkDir, when non-empty, overrides the repository root as the
	// process working directory. Used for per-module SCA (yarn/go).
	WorkDir string
}

// Result captures the outcome of running a Step.
type Result struct {
	Toolchain  Toolchain
	StepName   string
	Cmd        string
	Args       []string
	ExitCode   int
	DurationMS int64
	LogPath    string
	Truncated  string
	Err        error
}

// Workflow is one parsed CI workflow file.
type Workflow struct {
	Path          string
	Source        string
	InferredForge Forge
	RunsOn        []string
	Uses          []string
}

// Runner is one runner reported by a forge API.
type Runner struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
	Status string   `json:"status"`
}
