// prepush.go -- record-immutability pre-push gate (FR-011 / D-09).
//
// Implements the fourth invocation mode: the pre-push hook wrapper execs
// `rr-commit-guard --pre-push <remote> <url>` and the proposed ref updates
// arrive on stdin, one per line, in the githooks(5) format:
//
//   <local ref> SP <local sha> SP <remote ref> SP <remote sha> LF
//
// The mode blocks any non-fast-forward update or deletion of
// refs/heads/main (IN-H-002). It has no bypass path: bypassActive() is
// never consulted here.
package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

const zeroSHA = "0000000000000000000000000000000000000000"

// protectedRef is the single branch the record-immutability policy covers
// (FR-002 scopes the policy to main; widening is a future decision).
const protectedRef = "refs/heads/main"

// refUpdate is one githooks(5) pre-push stdin line.
type refUpdate struct {
	localRef, localSHA, remoteRef, remoteSHA string
}

// parseRefUpdates parses stdin lines "<local ref> <local sha> <remote ref>
// <remote sha>". A trailing empty line is ignored; any other malformed line
// is an error (caller maps it to exitInternal per the exit-code contract).
func parseRefUpdates(raw string) ([]refUpdate, error) {
	raw = strings.TrimSuffix(raw, "\n")
	if raw == "" {
		return nil, nil
	}
	var updates []refUpdate
	for _, line := range strings.Split(raw, "\n") {
		// Refs cannot contain spaces, so splitting on single spaces into
		// exactly 4 fields is safe.
		fields := strings.Split(line, " ")
		if len(fields) != 4 {
			return nil, fmt.Errorf("malformed ref-update line: %q", line)
		}
		for _, f := range fields {
			if f == "" {
				return nil, fmt.Errorf("malformed ref-update line: %q", line)
			}
		}
		updates = append(updates, refUpdate{
			localRef:  fields[0],
			localSHA:  fields[1],
			remoteRef: fields[2],
			remoteSHA: fields[3],
		})
	}
	return updates, nil
}

// isAncestor reports whether old is an ancestor of new, via
// `git merge-base --is-ancestor old new`. Exit 0 = true. Exit 1 (not an
// ancestor) and any higher exit (error) both yield false: fail-closed on
// the protected ref, because an unverifiable ancestry check must not pass.
// The remote sha is the remote's advertised value learned during this
// push's negotiation (githooks(5)); in the common case it equals the
// remote-tracking ref and its object is local, but in a lost-race window
// (another push landed, was advertised, and was never fetched here) the
// object may not exist locally. merge-base then errors, and the
// fail-closed mapping above turns that window into a safe block: the
// operator fetches and retries, and a genuine fast-forward passes.
//
// The command runs in the current working directory, which githooks(5)
// guarantees is the repository (worktree root, or $GIT_DIR when bare).
func isAncestor(old, new string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", old, new)
	return cmd.Run() == nil
}

// rewritesMain reports whether one update rewrites the protected ref, with
// the reason string for the block message.
func rewritesMain(u refUpdate) (bool, string) {
	if u.remoteRef != protectedRef {
		return false, ""
	}
	if u.localSHA == zeroSHA {
		return true, "deletion of refs/heads/main"
	}
	if u.remoteSHA == zeroSHA {
		return false, "" // first push of main to this remote: creation, not rewrite
	}
	if !isAncestor(u.remoteSHA, u.localSHA) {
		return true, "non-fast-forward update of refs/heads/main"
	}
	return false, ""
}

// runPrePush implements --pre-push mode. args carries the hook's argv (the
// wrapper passes <remote> <url> after the flag) so the audit record can name
// the push it gated; session is empty in hook mode.
func runPrePush(args []string, stdin io.Reader, stderr io.Writer) int {
	command := pushCommand(args)
	raw, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "rr-commit-guard: failed to read stdin")
		return exitInternal
	}
	updates, err := parseRefUpdates(string(raw))
	if err != nil {
		appendAudit(AuditRecord{Decision: DecisionBlock,
			Mode: ModePrePush.String(), Rule: "INPUT",
			Subject: err.Error(), Command: command})
		fmt.Fprintln(stderr, "rr-commit-guard: unable to parse pre-push stdin:", err)
		return exitInternal
	}

	for _, u := range updates {
		if rewrite, reason := rewritesMain(u); rewrite {
			appendAudit(AuditRecord{Decision: DecisionBlock,
				Mode: ModePrePush.String(), Rule: "IN-H-002",
				Command: command})
			fmt.Fprintln(stderr, "rr-commit-guard: BLOCKED -- push fails record-immutability rules:")
			fmt.Fprintf(stderr, "  [IN-H-002] %s -- history rewrite prohibited (RECORD-IMMUTABILITY-REQ-001 FR-002, FR-011)\n", reason)
			return exitBlock
		}
	}

	appendAudit(AuditRecord{Decision: DecisionAllow,
		Mode: ModePrePush.String(), Command: command})
	return exitAllow
}

// pushCommand assembles the audit command field from the hook's argv: the
// pre-push wrapper invokes the binary as `rr-commit-guard --pre-push
// <remote> <url>`, so the recorded command reads like the push it gates.
func pushCommand(args []string) string {
	for i, a := range args {
		if a == "--pre-push" {
			parts := []string{"push"}
			if i+1 < len(args) {
				parts = append(parts, args[i+1])
			}
			if i+2 < len(args) {
				parts = append(parts, args[i+2])
			}
			return strings.Join(parts, " ")
		}
	}
	return "push"
}
