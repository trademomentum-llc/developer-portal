// types.go -- shared types for rr-commit-guard.
//
// All public types are defined here so the rest of the package stays small
// and grep-friendly. See TEC-RR-COMMIT-GUARD-001 Section 3 for the shapes.
package main

// Mode names the four invocation modes of the binary.
type Mode int

const (
	ModePreToolUse Mode = iota
	ModeScanStaged
	ModeValidateMsg
	ModePrePush
)

func (m Mode) String() string {
	switch m {
	case ModePreToolUse:
		return "pretooluse"
	case ModeScanStaged:
		return "scan-staged"
	case ModeValidateMsg:
		return "validate-msg"
	case ModePrePush:
		return "pre-push"
	default:
		return "unknown"
	}
}

// Decision is the outcome of a single guard invocation.
type Decision string

const (
	DecisionAllow  Decision = "allow"
	DecisionWarn   Decision = "warn"
	DecisionBlock  Decision = "block"
	DecisionBypass Decision = "bypass"
)

// Rule describes one path-level check. Match is invoked per staged path; size
// is the staged blob size in bytes (or worktree size as fallback).
type Rule struct {
	Code      string
	Principle string
	Severity  string // "block" | "warn"
	Reason    string
	Match     func(path string, size int64) bool
}

// Finding pairs a fired rule with the offending path.
type Finding struct {
	Rule Rule
	Path string
	Size int64
}

// AuditRecord is the JSON shape persisted to the audit log.
type AuditRecord struct {
	TS       string   `json:"ts"`
	Decision Decision `json:"decision"`
	Mode     string   `json:"mode"`
	Rule     string   `json:"rule,omitempty"`
	Paths    []string `json:"paths,omitempty"`
	Subject  string   `json:"subject,omitempty"`
	Session  string   `json:"session,omitempty"`
	Command  string   `json:"command,omitempty"`
}

// CommitInvocation describes a parsed `git commit ...` Bash command.
type CommitInvocation struct {
	IsCommit    bool
	MessageArgs []string
	FileArg     string
	Amend       bool
	NoEdit      bool
	RepoDir     string
}

// Exit codes (NFR-005).
const (
	exitAllow    = 0
	exitInternal = 1
	exitBlock    = 2
)
