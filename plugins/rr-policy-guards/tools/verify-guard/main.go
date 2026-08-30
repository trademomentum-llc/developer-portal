// main.go -- PreToolUse hook entrypoint for rr-verify-guard.
//
// Reads Claude Code PreToolUse JSON from stdin, decides whether the
// underlying `git commit` / `git push` should be allowed by running the
// full verification pipeline:
//
//   Phase 1 - direct toolchain checks (go vet, go test, etc.)
//   Phase 2 - forge intelligence (workflow grammar, runner inventory,
//             action ref resolution, optional act run)
//
// Exits 0 (allow), 1 (degraded fail-closed), or 2 (block). Block path
// emits a permissionDecision deny payload to stdout. Every invocation writes
// one JSONL line to the audit log. Verification has no bypass path.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

// run is the testable entrypoint. main() is a one-line shim around it.
func run(stdin io.Reader, stdout, stderr io.Writer) int {
	start := time.Now()
	raw, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "rr-verify-guard: failed to read stdin")
		return exitBlock
	}
	// Hosts that do not speak the Claude Code hook protocol (empty stdin)
	// must not fail-closed every shell call. Real PreToolUse always sends JSON.
	if len(bytes.TrimSpace(raw)) == 0 {
		return exitAllow
	}
	var input ToolInput
	if jerr := json.Unmarshal(raw, &input); jerr != nil {
		auditQuiet(AuditLine{Action: "block", Reason: "internal-error / unparseable-input"})
		fmt.Fprintln(stderr, "rr-verify-guard: unable to parse PreToolUse input")
		return exitBlock
	}

	if input.ToolName != "Bash" {
		auditQuiet(AuditLine{Action: "allow", Reason: "out-of-scope", Tool: input.ToolName, Session: input.SessionID})
		return exitAllow
	}
	command, _ := input.ToolInput["command"].(string)
	targets := detectTargetCommands(command)
	if !targets.Any() {
		auditQuiet(AuditLine{Action: "allow", Reason: "non-target-command", Command: command, Session: input.SessionID})
		return exitAllow
	}

	if targets.Commit && targets.Push {
		reason := "commit-and-push-must-be-separate"
		emitDeny(stdout, "rr-verify-guard: "+reason+"; commit first, then invoke push separately so the exact committed state can be verified")
		auditQuiet(AuditLine{Action: "block", Reason: reason, Command: command, Session: input.SessionID, DurationMS: time.Since(start).Milliseconds()})
		return exitBlock
	}

	cwd, _ := os.Getwd()
	repoRoot, ok, repoReason := resolveTargetRepoRoot(cwd, targets)
	if repoReason != "" {
		emitDeny(stdout, "rr-verify-guard: "+repoReason+"; use the Bash tool working directory or one explicit git -C PATH target")
		auditQuiet(AuditLine{Action: "block", Reason: repoReason, Command: command, Session: input.SessionID, DurationMS: time.Since(start).Milliseconds()})
		return exitBlock
	}
	if !ok {
		auditQuiet(AuditLine{Action: "allow", Reason: "not-git", Command: command, Session: input.SessionID})
		return exitAllow
	}
	if targets.Push {
		dirty, status, err := publishStateDirty(repoRoot)
		if err != nil {
			reason := "publish-state-unreadable"
			emitDeny(stdout, "rr-verify-guard: "+reason+"; unable to prove the Git state is clean")
			auditQuiet(AuditLine{Action: "block", Reason: reason, Command: command, Session: input.SessionID, Repo: repoRoot, DurationMS: time.Since(start).Milliseconds()})
			return exitBlock
		}
		if dirty {
			reason := "publish-state-dirty"
			emitDeny(stdout, "rr-verify-guard: "+reason+"; commit, stash, or remove all staged, unstaged, and untracked changes before push\n\n"+status)
			auditQuiet(AuditLine{Action: "block", Reason: reason, Command: command, Session: input.SessionID, Repo: repoRoot, DurationMS: time.Since(start).Milliseconds()})
			return exitBlock
		}
	}

	cfg, _, _ := LoadConfig(repoRoot)

	toolchains, _ := DetectToolchains(repoRoot)
	workflows, _ := DiscoverWorkflows(repoRoot)
	if len(toolchains) == 0 && len(workflows) == 0 && !targets.Push {
		auditQuiet(AuditLine{
			Action: "allow", Reason: "no-toolchain-no-workflow",
			Command: command, Session: input.SessionID, Repo: repoRoot,
		})
		return exitAllow
	}

	forgeInfo, _ := DetectForge(repoRoot)
	verificationToolchains := append([]Toolchain(nil), toolchains...)
	if targets.Push {
		verificationToolchains = append(verificationToolchains, ToolchainSecurity)
	}

	// Pipeline cache lookup.
	hexKey, kerr := PipelineKey(repoRoot, verificationToolchains)
	if !targets.Push && kerr == nil && LookupPipeline(hexKey, 30) {
		auditQuiet(AuditLine{
			Action: "allow", Reason: "cache-hit",
			Command: command, Session: input.SessionID, Repo: repoRoot,
			Forge: forgeInfo.Forge, Toolchains: verificationToolchains,
			DurationMS: time.Since(start).Milliseconds(),
		})
		return exitAllow
	}

	runID := NewRunID()
	var degraded []string

	// ---------- Phase 1: direct toolchain checks ----------
	if len(toolchains) > 0 {
		_, fail, tcDegraded := RunToolchains(repoRoot, toolchains, runID, cfg)
		degraded = append(degraded, tcDegraded...)
		if fail != nil {
			reason := fmt.Sprintf("toolchain-fail / %s-%s", fail.Toolchain, fail.StepName)
			emitDeny(stdout, fmt.Sprintf("rr-verify-guard: %s failed (exit %d)\nCommand: %s %s\n\n%s",
				reason, fail.ExitCode, fail.Cmd, strings.Join(fail.Args, " "), fail.Truncated))
			auditQuiet(AuditLine{
				Action: "block", Reason: reason, Command: command,
				Session: input.SessionID, Repo: repoRoot,
				Forge: forgeInfo.Forge, Toolchains: toolchains, RunID: runID,
				DurationMS: time.Since(start).Milliseconds(),
			})
			return exitBlock
		}
	}

	if targets.Push {
		_, securityFailure := RunSecurityScans(repoRoot, runID)
		if securityFailure != nil {
			reason := "security-fail / " + securityFailure.StepName
			emitDeny(stdout, fmt.Sprintf("rr-verify-guard: %s failed (exit %d)\nCommand: %s %s\n\n%s",
				reason, securityFailure.ExitCode, securityFailure.Cmd, strings.Join(securityFailure.Args, " "), securityFailure.Truncated))
			auditQuiet(AuditLine{
				Action: "block", Reason: reason, Command: command,
				Session: input.SessionID, Repo: repoRoot,
				Forge: forgeInfo.Forge, Toolchains: verificationToolchains, RunID: runID,
				DurationMS: time.Since(start).Milliseconds(),
			})
			return exitBlock
		}
	}

	// ---------- Phase 2: forge intelligence (workflow-driven) ----------
	if len(workflows) > 0 {
		// 2a. Grammar validation via `act --list` (non-blocking when act missing).
		if ActAvailable() {
			_, gFail := ActListWorkflows(workflows)
			if gFail != nil {
				reason := "workflow-invalid / " + lastPathSegment(gFail.Args[len(gFail.Args)-1])
				emitDeny(stdout, fmt.Sprintf("rr-verify-guard: %s\n\n%s", reason, gFail.Truncated))
				auditQuiet(AuditLine{
					Action: "block", Reason: reason, Command: command,
					Session: input.SessionID, Repo: repoRoot,
					Forge: forgeInfo.Forge, RunID: runID,
					DurationMS: time.Since(start).Milliseconds(),
				})
				return exitBlock
			}
		} else {
			degraded = append(degraded, "act-missing / grammar")
		}

		// 2b. Runner-inventory check via forge API.
		client := NewForgeClient(forgeInfo.Forge)
		if forgeInfo.Forge != ForgeNone && forgeInfo.Forge != ForgeUnknown && client != nil && client.HasCredentials() && os.Getenv("RR_VERIFY_GUARD_SKIP_RUNNERS") != "1" {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			res, err := CheckRunnerAvailability(ctx, client, forgeInfo.Owner, forgeInfo.Repo, workflows)
			cancel()
			if err != nil {
				degraded = append(degraded, fmt.Sprintf("forge-unreachable / %s", forgeInfo.Forge))
			} else if len(res.Unsatisfied) > 0 {
				reason := "runner-unavailable / " + strings.Join(res.Unsatisfied, ",")
				emitDeny(stdout, fmt.Sprintf(
					"rr-verify-guard: no online %s runner satisfies labels: %s\nrequired=%s\nsatisfied=%s",
					forgeInfo.Forge, strings.Join(res.Unsatisfied, ","),
					strings.Join(res.Required, ","), strings.Join(res.Satisfied, ",")))
				auditQuiet(AuditLine{
					Action: "block", Reason: reason, Command: command,
					Session: input.SessionID, Repo: repoRoot,
					Forge: forgeInfo.Forge, RunID: runID,
					DurationMS: time.Since(start).Milliseconds(),
				})
				return exitBlock
			}
		} else if forgeInfo.Forge != ForgeNone && (client == nil || !client.HasCredentials()) {
			degraded = append(degraded, "forge-credentials-missing / "+string(forgeInfo.Forge))
		}

		// 2c. Action reference resolution.
		if os.Getenv("RR_VERIFY_GUARD_SKIP_ACTIONS") != "1" {
			clientByForge := map[Forge]ForgeClient{}
			for _, f := range []Forge{ForgeGitHub, ForgeGitea, ForgeForgejo} {
				if c := NewForgeClient(f); c != nil {
					clientByForge[f] = c
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_, fail, aDegraded, _ := ResolveAllActions(ctx, clientByForge, workflows)
			cancel()
			degraded = append(degraded, aDegraded...)
			if fail != nil {
				var reason string
				if !fail.RepoExists {
					reason = "action-not-found / " + fail.Ref.Raw
				} else if !fail.RefExists {
					reason = "action-ref-not-found / " + fail.Ref.Raw
				} else if fail.Error != nil {
					reason = "action-resolve-error / " + fail.Ref.Raw
				}
				emitDeny(stdout, fmt.Sprintf("rr-verify-guard: %s", reason))
				auditQuiet(AuditLine{
					Action: "block", Reason: reason, Command: command,
					Session: input.SessionID, Repo: repoRoot,
					Forge: forgeInfo.Forge, RunID: runID,
					DurationMS: time.Since(start).Milliseconds(),
				})
				return exitBlock
			}
		} else {
			degraded = append(degraded, "actions-skip / env")
		}

		// 2d. Optional `act --rm` full run. Gated by the repo config's
		// act.enabled (the schema field previously declared but never
		// honored): workflows that need live forge credentials (mirror
		// pushes, deploys) cannot pass under local emulation, so a repo
		// may opt out of full runs while keeping `act --list` grammar
		// validation in 2a.
		actFullRun := ActAvailable() && DockerAvailable()
		if cfg.Act.Enabled != nil && !*cfg.Act.Enabled {
			actFullRun = false
		}
		if actFullRun {
			_, aFail := ActRunWorkflows(workflows)
			if aFail != nil {
				reason := "act-failure / " + lastPathSegment(aFail.Args[1])
				emitDeny(stdout, fmt.Sprintf("rr-verify-guard: %s\n\n%s", reason, aFail.Truncated))
				auditQuiet(AuditLine{
					Action: "block", Reason: reason, Command: command,
					Session: input.SessionID, Repo: repoRoot,
					Forge: forgeInfo.Forge, RunID: runID,
					ActUsed:    true,
					DurationMS: time.Since(start).Milliseconds(),
				})
				return exitBlock
			}
		} else {
			switch {
			case cfg.Act.Enabled != nil && !*cfg.Act.Enabled:
				// Repo opted out of full act runs via .rr-verify-guard.json;
				// not a degradation.
			case !ActAvailable():
				degraded = append(degraded, "act-missing")
			default:
				degraded = append(degraded, "docker-unavailable")
			}
		}
	}

	if targets.Push && len(degraded) > 0 {
		reason := "verification-degraded / " + strings.Join(degraded, "; ")
		emitDeny(stdout, "rr-verify-guard: "+reason)
		auditQuiet(AuditLine{
			Action: "block", Reason: reason, Command: command,
			Session: input.SessionID, Repo: repoRoot,
			Forge: forgeInfo.Forge, Toolchains: verificationToolchains, RunID: runID,
			DurationMS: time.Since(start).Milliseconds(),
		})
		return exitBlock
	}

	// All checks passed.
	if hexKey != "" && !targets.Push {
		_ = StorePipeline(hexKey, PipelineCacheMeta{
			Repo:       repoRoot,
			Toolchains: verificationToolchains,
			Forge:      forgeInfo.Forge,
			DurationMS: time.Since(start).Milliseconds(),
			ActUsed:    ActAvailable() && DockerAvailable(),
		})
	}
	auditQuiet(AuditLine{
		Action: "allow", Reason: "verified",
		Command: command, Session: input.SessionID, Repo: repoRoot,
		Forge: forgeInfo.Forge, Toolchains: verificationToolchains, RunID: runID,
		DurationMS: time.Since(start).Milliseconds(),
	})
	for _, d := range degraded {
		auditQuiet(AuditLine{Action: "degraded", Reason: d, Command: command, Session: input.SessionID, Repo: repoRoot, Forge: forgeInfo.Forge})
	}
	return exitAllow
}

// isTargetCommand returns true when any executable shell segment contains a
// direct git commit or git push invocation.
func isTargetCommand(cmd string) bool {
	return detectTargetCommands(cmd).Any()
}

// resolveRepoRoot calls `git rev-parse --show-toplevel` from cwd.
func resolveRepoRoot(cwd string) (string, bool) {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", false
	}
	return root, true
}

func resolveTargetRepoRoot(cwd string, targets targetCommandSet) (string, bool, string) {
	if targets.RepoContextUnsupported {
		return "", false, "target-repo-context-unsupported"
	}
	repoDirs := targets.RepoDirs
	if len(repoDirs) == 0 {
		repoDirs = []string{""}
	}
	resolvedRoot := ""
	for _, repoDir := range repoDirs {
		start := cwd
		if repoDir != "" {
			if filepath.IsAbs(repoDir) {
				start = filepath.Clean(repoDir)
			} else {
				start = filepath.Clean(filepath.Join(cwd, repoDir))
			}
		}
		root, ok := resolveRepoRoot(start)
		if !ok {
			if repoDir == "" {
				return "", false, ""
			}
			return "", false, "target-repo-unreadable"
		}
		if resolvedRoot == "" {
			resolvedRoot = root
			continue
		}
		if root != resolvedRoot {
			return "", false, "multiple-target-repositories"
		}
	}
	return resolvedRoot, true, ""
}

func publishStateDirty(repoRoot string) (bool, string, error) {
	command := exec.Command("git", "-C", repoRoot, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := command.CombinedOutput()
	if err != nil {
		return false, "", err
	}
	status := strings.TrimSpace(string(output))
	return status != "", status, nil
}

// emitDeny writes the PreToolUse permissionDecision=deny JSON to stdout.
func emitDeny(w io.Writer, reason string) {
	out := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(out)
}

// lastPathSegment returns the trailing segment of a path or argument.
func lastPathSegment(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' || s[i] == '\\' {
			return s[i+1:]
		}
	}
	return s
}
