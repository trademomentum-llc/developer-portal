// main.go -- PreToolUse hook entrypoint for the bash safety guard.
//
// Reads Claude Code PreToolUse JSON from stdin, extracts the command from
// a Bash tool invocation, scans it for unsafe shell patterns (bare $VAR
// expansion, etc), and exits:
//
//   - 0 on allow (no unsafe pattern found, or tool not in scope)
//   - 2 on block (unsafe pattern found and no bypass active); stderr
//     contains a warning explaining the violation AND a corrected command
//     using safe syntax that Claude should re-issue
//   - 1 on internal error (malformed input, fail closed as block)
//
// Bypass: setting RR_BASH_GUARD_BYPASS=1 in the environment allows any
// command through, logged as a bypass event in the audit log.
//
// Audit: every allow / block / bypass is appended as a single JSON line
// to ~/.rational-reserve/logs/bash-guard.jsonl unless RR_BASH_GUARD_AUDIT_LOG
// overrides the path.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ToolInput mirrors the shape Claude Code's PreToolUse hook passes on stdin.
type ToolInput struct {
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
	SessionID string         `json:"session_id,omitempty"`
}

const (
	exitAllow    = 0
	exitBlock    = 2
	exitInternal = 1
)

func run(r io.Reader, stderr io.Writer) int {
	raw, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintln(stderr, "rr-bash-guard: failed to read stdin")
		return exitInternal
	}

	var input ToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		logAudit("block", "unparseable-input: "+err.Error(), "", "")
		fmt.Fprintln(stderr, "rr-bash-guard: unable to parse PreToolUse input")
		return exitBlock
	}

	if input.ToolName != "Bash" {
		return exitAllow
	}

	command, ok := input.ToolInput["command"].(string)
	if !ok || command == "" {
		logAudit("allow", "empty-command", input.ToolName, input.SessionID)
		return exitAllow
	}

	violations := ScanCommand(command)
	if len(violations) == 0 {
		logAudit("allow", "", input.ToolName, input.SessionID)
		return exitAllow
	}

	// Bypass check.
	if os.Getenv("RR_BASH_GUARD_BYPASS") == "1" {
		logAudit("bypass", violations[0].Description, input.ToolName, input.SessionID)
		fmt.Fprintln(stderr, "rr-bash-guard: bypass in effect (RR_BASH_GUARD_BYPASS=1)")
		for _, v := range violations {
			fmt.Fprintf(stderr, "rr-bash-guard: would have blocked: %s\n", v.Description)
		}
		return exitAllow
	}

	// Build the corrected command.
	corrected := ApplyFixes(command, violations)

	logAudit("block", violations[0].Description, input.ToolName, input.SessionID)

	fmt.Fprintln(stderr, "rr-bash-guard: BLOCKED -- unsafe shell syntax detected:")
	for _, v := range violations {
		fmt.Fprintf(stderr, "  [%s] %s\n", v.Rule, v.Description)
	}
	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "rr-bash-guard: corrected command using safe syntax:")
	fmt.Fprintf(stderr, "  %s\n", corrected)
	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "rr-bash-guard: re-issue with the corrected command above.")
	fmt.Fprintln(stderr, "rr-bash-guard: to override set RR_BASH_GUARD_BYPASS=1 (logged to audit file).")
	return exitBlock
}

func main() {
	os.Exit(run(os.Stdin, os.Stderr))
}
