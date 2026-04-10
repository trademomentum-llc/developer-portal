// main.go -- PreToolUse hook entrypoint for the non-ASCII guard.
//
// Reads Claude Code PreToolUse JSON from stdin, extracts the content to be
// written by a Write / Edit / MultiEdit tool invocation, scans it for any
// non-ASCII rune, and exits:
//
//   - 0 on allow (no non-ASCII found, or tool not in scope)
//   - 2 on block (non-ASCII found and no bypass active); stderr explains
//   - 1 on internal error (malformed input, fail closed as block)
//
// Bypass: setting RR_EMOJI_GUARD_BYPASS=1 in the environment allows any
// content through, logged as a bypass event in the audit log.
//
// Audit: every allow / block / bypass is appended as a single JSON line
// to ~/.rational-reserve/logs/emoji-guard.jsonl unless RR_EMOJI_GUARD_AUDIT_LOG
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

// inScope reports whether the tool name is one whose content we inspect.
// Keeping this list explicit prevents accidental scope creep and makes
// tests easy to reason about.
func inScope(toolName string) bool {
	switch toolName {
	case "Write", "Edit", "MultiEdit":
		return true
	}
	return false
}

func run(r io.Reader, stderr io.Writer) int {
	raw, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintln(stderr, "rr-emoji-guard: failed to read stdin")
		return exitInternal
	}

	var input ToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		// Fail closed on malformed input: we do not know what we are
		// being asked to validate, so refuse rather than guess.
		logAudit("block", "unparseable-input: "+err.Error(), "", "")
		fmt.Fprintln(stderr, "rr-emoji-guard: unable to parse PreToolUse input")
		return exitBlock
	}

	if !inScope(input.ToolName) {
		// Other tool categories (Bash, Read, Glob, etc) are not our concern.
		return exitAllow
	}

	content := ExtractContent(input.ToolName, input.ToolInput)
	if content == "" {
		// Nothing to scan. Allow.
		logAudit("allow", "empty-content", input.ToolName, input.SessionID)
		return exitAllow
	}

	hit := Scan(content)
	if hit == nil {
		logAudit("allow", "", input.ToolName, input.SessionID)
		return exitAllow
	}

	// Violation found. Check bypass.
	if os.Getenv("RR_EMOJI_GUARD_BYPASS") == "1" {
		logAudit("bypass", hit.String(), input.ToolName, input.SessionID)
		fmt.Fprintln(stderr, "rr-emoji-guard: bypass in effect (RR_EMOJI_GUARD_BYPASS=1)")
		fmt.Fprintf(stderr, "rr-emoji-guard: would have blocked at %s\n", hit.String())
		return exitAllow
	}

	logAudit("block", hit.String(), input.ToolName, input.SessionID)
	fmt.Fprintf(stderr,
		"rr-emoji-guard: blocked -- file write contains non-ASCII character at %s\n",
		hit.String())
	fmt.Fprintln(stderr,
		"rr-emoji-guard: all files must be pure ASCII. Use `--` instead of em dash, "+
			"`->` instead of Unicode arrows, plain ASCII art for diagrams.")
	fmt.Fprintln(stderr,
		"rr-emoji-guard: to override for one command set RR_EMOJI_GUARD_BYPASS=1 "+
			"(logged to audit file).")
	return exitBlock
}

func main() {
	os.Exit(run(os.Stdin, os.Stderr))
}
