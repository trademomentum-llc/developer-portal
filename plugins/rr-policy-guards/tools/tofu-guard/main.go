// Package main: rr-tofu-guard entrypoint. Reads Claude Code PreToolUse JSON
// from stdin, inspects Bash tool uses, blocks state-mutating tofu subcommands.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ToolInput struct {
	ToolName  string           `json:"tool_name"`
	ToolInput ToolInputPayload `json:"tool_input"`
	SessionID string           `json:"session_id,omitempty"`
}

type ToolInputPayload struct {
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
}

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
