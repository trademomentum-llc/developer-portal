// main.go -- PreToolUse hook entrypoint for the brew security guard.
//
// Reads Claude Code PreToolUse JSON from stdin, extracts the command from
// a Bash tool invocation, and if it is a brew command, validates it against
// a security policy that blocks supply-chain-risky operations.
//
// Exit codes:
//   - 0: allow (not a brew command, or brew command passes validation)
//   - 2: block (brew command fails validation and no bypass active)
//   - 1: internal error (malformed input, fail closed as block -- exits 2)
//
// Bypass: setting RR_BREW_GUARD_BYPASS=1 allows any brew command through,
// logged as a bypass event.
//
// Audit: every decision is appended to
// ~/.rational-reserve/logs/brew-guard.jsonl unless RR_BREW_GUARD_AUDIT_LOG
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
	ToolName  string          `json:"tool_name"`
	ToolInput ToolInputPayload `json:"tool_input"`
	SessionID string          `json:"session_id,omitempty"`
}

// ToolInputPayload holds the Bash tool's parameters.
type ToolInputPayload struct {
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
}

const (
	exitAllow = 0
	exitBlock = 2
)

func run(r io.Reader, stderr io.Writer) int {
	raw, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintln(stderr, "brew-guard: failed to read stdin")
		return exitBlock
	}

	var input ToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		logAudit("block", "unparseable-input: "+err.Error(), "", "")
		fmt.Fprintln(stderr, "brew-guard: unable to parse PreToolUse input")
		return exitBlock
	}

	if input.ToolName != "Bash" {
		return exitAllow
	}

	cmd := input.ToolInput.Command
	if cmd == "" {
		return exitAllow
	}

	tokens, err := Tokenize(cmd)
	if err != nil {
		logAudit("block", "tokenize-error: "+err.Error(), cmd, input.SessionID)
		fmt.Fprintln(stderr, "brew-guard: malformed command, refusing to evaluate")
		return exitBlock
	}

	if len(tokens) == 0 || tokens[0] != "brew" {
		return exitAllow
	}

	decision := ValidateBrewCommand(tokens)

	if !decision.Allow && os.Getenv("RR_BREW_GUARD_BYPASS") == "1" {
		logAudit("bypass", decision.Reason, cmd, input.SessionID)
		fmt.Fprintln(stderr, "brew-guard: bypass in effect (RR_BREW_GUARD_BYPASS=1)")
		return exitAllow
	}

	if decision.Allow {
		logAudit("allow", "", cmd, input.SessionID)
		return exitAllow
	}

	logAudit("block", decision.Reason, cmd, input.SessionID)
	fmt.Fprintf(stderr, "brew-guard: blocked -- %s\n", decision.Reason)
	return exitBlock
}

func main() {
	os.Exit(run(os.Stdin, os.Stderr))
}
