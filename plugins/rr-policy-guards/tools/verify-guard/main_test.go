// main_test.go -- end-to-end tests of run() that drive the binary
// entry point with synthesized PreToolUse JSON. Each subtest sets up
// a temp directory (sometimes a temp git repo), points HOME at a
// scratch dir so caches/audits go there, and calls run().

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// makeRunInput synthesizes a PreToolUse stdin payload for `tool` with
// the given Bash command.
func makeRunInput(toolName, command string) []byte {
	in := map[string]any{
		"tool_name": toolName,
		"tool_input": map[string]any{
			"command": command,
		},
		"session_id": "test",
	}
	b, _ := json.Marshal(in)
	return b
}

// withScratchHome routes HOME and audit log into a per-test dir so
// state never leaks between tests or into the user's real cache.
func withScratchHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", filepath.Join(home, "audit.jsonl"))
	t.Setenv("RR_VERIFY_GUARD_BYPASS", "")
	return home
}

func TestRun_NonBashToolAllowed(t *testing.T) {
	withScratchHome(t)
	in := makeRunInput("Edit", "")
	rc := run(bytes.NewReader(in), &bytes.Buffer{}, &bytes.Buffer{})
	if rc != exitAllow {
		t.Errorf("non-Bash tool should allow, got rc=%d", rc)
	}
}

func TestRun_NonTargetCommandAllowed(t *testing.T) {
	withScratchHome(t)
	in := makeRunInput("Bash", "ls -la")
	rc := run(bytes.NewReader(in), &bytes.Buffer{}, &bytes.Buffer{})
	if rc != exitAllow {
		t.Errorf("non-target command should allow, got rc=%d", rc)
	}
}

func TestRun_FailingVerificationCannotBeBypassed(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	withScratchHome(t)
	repo := initTestRepo(t, `package main

import "testing"

func TestForcedFail(t *testing.T) {
	t.Fatal("intentional failure")
}
`)
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		bypass  string
		command string
	}{
		{name: "environment variable", bypass: "1", command: `git commit -m "fix: test"`},
		{name: "commit message tag", command: `git commit -m "fix: test [skip-verify]"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("RR_VERIFY_GUARD_BYPASS", test.bypass)
			out := &bytes.Buffer{}
			code := run(bytes.NewReader(makeRunInput("Bash", test.command)), out, &bytes.Buffer{})
			if code != exitBlock {
				t.Fatalf("verification bypass returned code %d, want %d; output=%q", code, exitBlock, out.String())
			}
		})
	}
}

func TestRun_NotGitDirAllowed(t *testing.T) {
	withScratchHome(t)
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	defer os.Chdir(prevWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	in := makeRunInput("Bash", "git commit -m x")
	rc := run(bytes.NewReader(in), &bytes.Buffer{}, &bytes.Buffer{})
	if rc != exitAllow {
		t.Errorf("not-git dir should allow, got rc=%d", rc)
	}
}

// initTestRepo creates a minimal git repo + go module and returns the
// path. testBody is written into main_test.go.
func initTestRepo(t *testing.T, testBody string) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "config", "user.email", "t@t"},
		{"-c", "user.name=t", "config", "user.name", "t"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testfail\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(testBody), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-q", "-m", "seed"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func TestRun_PassingTestAllows(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	withScratchHome(t)
	repo := initTestRepo(t, `package main

import "testing"

func TestPasses(t *testing.T) {}
`)
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	in := makeRunInput("Bash", "git commit -m feat")
	out := &bytes.Buffer{}
	rc := run(bytes.NewReader(in), out, &bytes.Buffer{})
	if rc != exitAllow {
		t.Fatalf("passing-test repo should allow, got rc=%d, out=%q", rc, out.String())
	}
}

func TestRun_FailingTestBlocks(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	withScratchHome(t)
	repo := initTestRepo(t, `package main

import "testing"

func TestForcedFail(t *testing.T) {
	t.Fatal("intentional failure")
}
`)
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	in := makeRunInput("Bash", "git commit -m feat")
	out := &bytes.Buffer{}
	rc := run(bytes.NewReader(in), out, &bytes.Buffer{})
	if rc != exitBlock {
		t.Fatalf("failing-test repo should block, got rc=%d", rc)
	}
	var deny struct {
		HookSpecificOutput struct {
			HookEventName            string `json:"hookEventName"`
			PermissionDecision       string `json:"permissionDecision"`
			PermissionDecisionReason string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &deny); err != nil {
		t.Fatalf("deny JSON malformed: %v\nout=%q", err, out.String())
	}
	if deny.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Errorf("hookEventName = %q, want PreToolUse", deny.HookSpecificOutput.HookEventName)
	}
	if deny.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("permissionDecision = %q, want deny", deny.HookSpecificOutput.PermissionDecision)
	}
	if !strings.Contains(deny.HookSpecificOutput.PermissionDecisionReason, "toolchain-fail / go-test") {
		t.Errorf("reason should name go-test failure, got %q", deny.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestRun_MalformedStdinBlocks(t *testing.T) {
	withScratchHome(t)
	rc := run(bytes.NewReader([]byte("not json")), &bytes.Buffer{}, &bytes.Buffer{})
	if rc != exitBlock {
		t.Errorf("malformed stdin should block, got rc=%d", rc)
	}
}

func TestRun_PushBlocksDirtyPublishState(t *testing.T) {
	withScratchHome(t)
	repo := initRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "seed"), []byte("dirty\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	code := run(bytes.NewReader(makeRunInput("Bash", "git push origin main")), out, &bytes.Buffer{})
	if code != exitBlock {
		t.Fatalf("dirty push returned code %d, want %d; output=%q", code, exitBlock, out.String())
	}
	if !strings.Contains(out.String(), "publish-state-dirty") {
		t.Fatalf("dirty push output = %q, want publish-state-dirty", out.String())
	}
}

func TestRun_PushRunsSecurityScanners(t *testing.T) {
	withScratchHome(t)
	repo := initRepo(t)
	fakeBin := t.TempDir()
	writeFakeExecutable(t, fakeBin, "semgrep", 1)
	writeFakeExecutable(t, fakeBin, "gitleaks", 0)
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	code := run(bytes.NewReader(makeRunInput("Bash", "git push origin main")), out, &bytes.Buffer{})
	if code != exitBlock {
		t.Fatalf("security scanner failure returned code %d, want %d; output=%q", code, exitBlock, out.String())
	}
	if !strings.Contains(out.String(), "security-fail / semgrep") {
		t.Fatalf("security scanner output = %q, want semgrep failure", out.String())
	}
}

func TestRun_CommitAndPushMustBeSeparate(t *testing.T) {
	withScratchHome(t)
	repo := initRepo(t)
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	command := "git commit -m 'fix: test' && git push origin main"
	code := run(bytes.NewReader(makeRunInput("Bash", command)), out, &bytes.Buffer{})
	if code != exitBlock {
		t.Fatalf("compound commit and push returned code %d, want %d; output=%q", code, exitBlock, out.String())
	}
	if !strings.Contains(out.String(), "commit-and-push-must-be-separate") {
		t.Fatalf("compound output = %q, want separation reason", out.String())
	}
}

func TestRun_GitDashCTargetsReferencedRepo(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	withScratchHome(t)
	repo := initTestRepo(t, `package main

import "testing"

func TestForcedFail(t *testing.T) {
	t.Fatal("intentional failure")
}
`)
	outside := t.TempDir()
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(outside); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	command := fmt.Sprintf("git -C %q commit -m 'fix: test'", repo)
	code := run(bytes.NewReader(makeRunInput("Bash", command)), out, &bytes.Buffer{})
	if code != exitBlock {
		t.Fatalf("git -C verification returned code %d, want %d; output=%q", code, exitBlock, out.String())
	}
	if !strings.Contains(out.String(), "toolchain-fail / go-test") {
		t.Fatalf("git -C output = %q, want referenced repository test failure", out.String())
	}
}

func TestRun_DirectoryChangingTargetBlocks(t *testing.T) {
	withScratchHome(t)
	repo := initRepo(t)
	outside := t.TempDir()
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(outside); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	command := fmt.Sprintf("cd %q && git push origin main", repo)
	code := run(bytes.NewReader(makeRunInput("Bash", command)), out, &bytes.Buffer{})
	if code != exitBlock {
		t.Fatalf("directory-changing push returned code %d, want %d; output=%q", code, exitBlock, out.String())
	}
	if !strings.Contains(out.String(), "target-repo-context-unsupported") {
		t.Fatalf("directory-changing output = %q, want explicit context failure", out.String())
	}
}

func TestRun_PushRescansSecurityOnEveryAttempt(t *testing.T) {
	withScratchHome(t)
	repo := initRepo(t)
	fakeBin := t.TempDir()
	writeFakeExecutable(t, fakeBin, "semgrep", 0)
	writeFakeExecutable(t, fakeBin, "gitleaks", 0)
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	previous, _ := os.Getwd()
	defer os.Chdir(previous)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	input := makeRunInput("Bash", "git push origin main")
	firstOut := &bytes.Buffer{}
	if code := run(bytes.NewReader(input), firstOut, &bytes.Buffer{}); code != exitAllow {
		t.Fatalf("first clean scan returned code %d, want %d; output=%q", code, exitAllow, firstOut.String())
	}

	writeFakeExecutable(t, fakeBin, "semgrep", 1)
	secondOut := &bytes.Buffer{}
	code := run(bytes.NewReader(input), secondOut, &bytes.Buffer{})
	if code != exitBlock {
		t.Fatalf("second push returned code %d, want %d; output=%q", code, exitBlock, secondOut.String())
	}
	if !strings.Contains(secondOut.String(), "security-fail / semgrep") {
		t.Fatalf("second push output = %q, want fresh semgrep failure", secondOut.String())
	}
}

func writeFakeExecutable(t *testing.T, directory, name string, exitCode int) {
	t.Helper()
	path := filepath.Join(directory, name)
	body := fmt.Sprintf("#!/bin/sh\nexit %d\n", exitCode)
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
}

func TestIsTargetCommand(t *testing.T) {
	cases := map[string]bool{
		"git commit -m x":                     true,
		"git push origin main":                true,
		"git status":                          false,
		"  git commit ":                       true,
		"git add .":                           false,
		"echo git commit":                     false,
		"git commit":                          true, // bare commit
		"git push":                            true,
		"echo ready && git push origin main":  true,
		"git status; git commit -m x":         true,
		"printf ready | git push origin main": true,
		`echo "$(git push origin main)"`:      true,
		"NAME=value git push origin main":     true,
		"env NAME=value git commit -m x":      true,
		"echo 'git push origin main'":         false,
		`printf '%s' "git commit -m x"`:       false,
	}
	for in, want := range cases {
		got := isTargetCommand(in)
		if got != want {
			t.Errorf("isTargetCommand(%q) = %v, want %v", in, got, want)
		}
	}
}
