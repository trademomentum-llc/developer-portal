// rr-commit-guard -- mechanical enforcement of commit-discipline principles.
//
// Four invocation modes, one binary. See TEC-RR-COMMIT-GUARD-001 for the
// full contract.
//
//   rr-commit-guard                       PreToolUse mode (read stdin)
//   rr-commit-guard --scan-staged         pre-commit hook mode
//   rr-commit-guard --validate-msg FILE   commit-msg hook mode
//   rr-commit-guard --pre-push REMOTE URL pre-push hook mode (read stdin)
//
// Exit codes: 0 allow / 2 block / 1 internal error.
// Bypass:     RR_COMMIT_GUARD_BYPASS=1 (still audit-logged). The IN-H-*
// record-immutability rules have no bypass path.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// version is overridable at build time via -ldflags="-X main.version=...".
var version = "0.1.0"

// ToolInput mirrors the PreToolUse JSON schema used by rr-bash-guard.
type ToolInput struct {
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
	SessionID string         `json:"session_id,omitempty"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stderr))
}

// run is the testable entrypoint. It returns an exit code rather than calling
// os.Exit so tests can drive every mode without process teardown.
func run(args []string, stdin io.Reader, stderr io.Writer) int {
	switch {
	case hasFlag(args, "--version"):
		fmt.Fprintln(stderr, "rr-commit-guard", version)
		return exitAllow
	case hasFlag(args, "--help") || hasFlag(args, "-h"):
		printHelp(stderr)
		return exitAllow
	case hasFlag(args, "--scan-staged"):
		return runScanStaged(args, stderr)
	case hasFlag(args, "--validate-msg"):
		return runValidateMsg(args, stderr)
	case hasFlag(args, "--pre-push"):
		return runPrePush(args, stdin, stderr)
	default:
		return runPreToolUse(stdin, stderr)
	}
}

// ---- PreToolUse mode ----------------------------------------------------

func runPreToolUse(stdin io.Reader, stderr io.Writer) int {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "rr-commit-guard: failed to read stdin")
		return exitInternal
	}
	var input ToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePreToolUse.String(), Rule: "INPUT",
			Subject: "unparseable PreToolUse input"})
		fmt.Fprintln(stderr, "rr-commit-guard: unable to parse PreToolUse input")
		return exitBlock
	}
	if input.ToolName != "Bash" {
		return exitAllow
	}
	command, _ := input.ToolInput["command"].(string)
	if command == "" {
		return exitAllow
	}
	inv := ExtractCommit(command)
	if !inv.IsCommit {
		return exitAllow
	}

	// 0. Record immutability (FR-002/FR-010): --amend rewrites published
	// history. No bypass: this check precedes bypassActive() on purpose.
	if inv.Amend {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePreToolUse.String(), Rule: "IN-H-001",
			Session: input.SessionID, Command: command})
		fmt.Fprintln(stderr, "rr-commit-guard: BLOCKED -- commit invocation fails record-immutability rules:")
		fmt.Fprintln(stderr, "  [IN-H-001] --amend rewrites published history; corrections must be new commits referencing the mistaken commit (RECORD-IMMUTABILITY-REQ-001 FR-002, FR-010)")
		return exitBlock
	}

	// 1. Scan staged paths.
	paths, err := StagedPaths(inv.RepoDir)
	if err != nil {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePreToolUse.String(), Rule: "GIT",
			Subject: err.Error(), Session: input.SessionID, Command: command})
		fmt.Fprintln(stderr, "rr-commit-guard: cannot read staged paths:", err)
		return exitBlock
	}
	sizes := sizesFor(inv.RepoDir, paths)
	findings := Scan(paths, sizes, Rules())
	blocking, warnings := SplitFindings(findings)

	// 2. Validate the message if one was supplied inline (-m or -F).
	var msgVerdict MsgVerdict
	switch {
	case inv.FileArg != "":
		raw, err := os.ReadFile(inv.FileArg)
		if err == nil {
			msgVerdict = VerdictFromMessage(raw)
		}
	case len(inv.MessageArgs) > 0:
		// Multiple -m flags are concatenated with blank lines, matching git.
		joined := strings.Join(inv.MessageArgs, "\n\n")
		msgVerdict = VerdictFromMessage([]byte(joined))
	}

	// 3. Decide.
	if bypassActive() {
		appendAudit(AuditRecord{Decision: DecisionBypass,
			Mode: ModePreToolUse.String(),
			Paths: findingPaths(blocking), Subject: msgVerdict.Subject,
			Session: input.SessionID, Command: command})
		fmt.Fprintln(stderr, "rr-commit-guard: bypass in effect (RR_COMMIT_GUARD_BYPASS=1)")
		return exitAllow
	}

	if len(blocking) > 0 {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePreToolUse.String(), Rule: blocking[0].Rule.Code,
			Paths: findingPaths(blocking), Session: input.SessionID, Command: command})
		printBlocking(stderr, blocking)
		hintBypass(stderr)
		return exitBlock
	}

	if msgVerdict.Decision == DecisionBlock {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePreToolUse.String(), Rule: msgVerdict.RuleCode,
			Subject: msgVerdict.Subject, Session: input.SessionID, Command: command})
		printMsgFailure(stderr, msgVerdict)
		hintBypass(stderr)
		return exitBlock
	}

	if len(warnings) > 0 {
		appendAudit(AuditRecord{Decision: DecisionWarn,
			Mode: ModePreToolUse.String(), Rule: warnings[0].Rule.Code,
			Paths: findingPaths(warnings), Session: input.SessionID})
		printWarnings(stderr, warnings)
		return exitAllow
	}

	appendAudit(AuditRecord{Decision: DecisionAllow,
		Mode: ModePreToolUse.String(), Session: input.SessionID})
	return exitAllow
}

// ---- --scan-staged ------------------------------------------------------

func runScanStaged(args []string, stderr io.Writer) int {
	repo := flagValue(args, "--repo")
	paths, err := StagedPaths(repo)
	if err != nil {
		fmt.Fprintln(stderr, "rr-commit-guard: cannot read staged paths:", err)
		return exitInternal
	}
	sizes := sizesFor(repo, paths)
	findings := Scan(paths, sizes, Rules())
	blocking, warnings := SplitFindings(findings)

	if bypassActive() {
		appendAudit(AuditRecord{Decision: DecisionBypass,
			Mode: ModeScanStaged.String(), Paths: findingPaths(blocking)})
		fmt.Fprintln(stderr, "rr-commit-guard: bypass in effect (RR_COMMIT_GUARD_BYPASS=1)")
		return exitAllow
	}

	if len(blocking) > 0 {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModeScanStaged.String(), Rule: blocking[0].Rule.Code,
			Paths: findingPaths(blocking)})
		printBlocking(stderr, blocking)
		hintBypass(stderr)
		return exitBlock
	}

	if len(warnings) > 0 {
		appendAudit(AuditRecord{Decision: DecisionWarn,
			Mode: ModeScanStaged.String(), Rule: warnings[0].Rule.Code,
			Paths: findingPaths(warnings)})
		printWarnings(stderr, warnings)
	}

	appendAudit(AuditRecord{Decision: DecisionAllow, Mode: ModeScanStaged.String()})
	return exitAllow
}

// ---- --validate-msg -----------------------------------------------------

func runValidateMsg(args []string, stderr io.Writer) int {
	path := flagValue(args, "--validate-msg")
	if path == "" {
		fmt.Fprintln(stderr, "rr-commit-guard: --validate-msg requires a path")
		return exitInternal
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(stderr, "rr-commit-guard: cannot read message file:", err)
		return exitInternal
	}
	verdict := VerdictFromMessage(raw)

	if bypassActive() {
		appendAudit(AuditRecord{Decision: DecisionBypass,
			Mode: ModeValidateMsg.String(), Subject: verdict.Subject})
		fmt.Fprintln(stderr, "rr-commit-guard: bypass in effect (RR_COMMIT_GUARD_BYPASS=1)")
		return exitAllow
	}
	if verdict.Decision == DecisionBlock {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModeValidateMsg.String(), Rule: verdict.RuleCode,
			Subject: verdict.Subject})
		printMsgFailure(stderr, verdict)
		hintBypass(stderr)
		return exitBlock
	}
	appendAudit(AuditRecord{Decision: DecisionAllow,
		Mode: ModeValidateMsg.String(), Subject: verdict.Subject})
	return exitAllow
}

// ---- helpers -----------------------------------------------------------

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
		if strings.HasPrefix(a, name+"=") {
			return true
		}
	}
	return false
}

func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"=")
		}
	}
	return ""
}

func sizesFor(repo string, paths []string) map[string]int64 {
	out := make(map[string]int64, len(paths))
	for _, p := range paths {
		out[p] = StagedSize(repo, p)
	}
	return out
}

func findingPaths(f []Finding) []string {
	out := make([]string, 0, len(f))
	for _, x := range f {
		out = append(out, x.Path)
	}
	return out
}

func printBlocking(w io.Writer, findings []Finding) {
	fmt.Fprintln(w, "rr-commit-guard: BLOCKED -- the following staged paths violate commit-discipline rules:")
	for _, f := range findings {
		fmt.Fprintf(w, "  [%s] %s -- %s\n", f.Rule.Code, f.Path, f.Rule.Reason)
	}
}

func printWarnings(w io.Writer, findings []Finding) {
	fmt.Fprintln(w, "rr-commit-guard: warnings (commit allowed, please review):")
	for _, f := range findings {
		fmt.Fprintf(w, "  [%s] %s -- %s\n", f.Rule.Code, f.Path, f.Rule.Reason)
	}
}

func printMsgFailure(w io.Writer, v MsgVerdict) {
	fmt.Fprintln(w, "rr-commit-guard: BLOCKED -- commit message fails intent rules:")
	fmt.Fprintf(w, "  [%s] %s\n", v.RuleCode, v.Reason)
	if v.RuleCode == "IN-M-003" {
		fmt.Fprintln(w, "  hint: add a blank line after the subject, then a paragraph explaining WHY.")
	}
	if v.RuleCode == "IN-M-002" {
		fmt.Fprintln(w, "  hint: subject must be 'feat|fix|chore|docs|refactor|test|build|ci|perf|style|revert: description'.")
	}
}

func hintBypass(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "rr-commit-guard: to override set RR_COMMIT_GUARD_BYPASS=1 (still audit-logged).")
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `rr-commit-guard -- mechanical enforcement of commit-discipline principles.

Usage:
  rr-commit-guard                       PreToolUse mode (read stdin)
  rr-commit-guard --scan-staged [--repo PATH]
                                        pre-commit hook mode
  rr-commit-guard --validate-msg FILE   commit-msg hook mode
  rr-commit-guard --pre-push REMOTE URL pre-push hook mode (read stdin)
  rr-commit-guard --version
  rr-commit-guard --help

Exit codes:
  0  allow (may include warnings on stderr)
  1  internal error (malformed input, git unreachable, ...)
  2  block

Env:
  RR_COMMIT_GUARD_BYPASS=1     allow this commit through (audit-logged)
  RR_COMMIT_GUARD_AUDIT_LOG    override audit log path
                               (default ~/.rational-reserve/logs/commit-guard.jsonl)`)
}
