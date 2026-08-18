package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// makeRepo initialises a throwaway git repo at t.TempDir() and returns the
// path. The repo has a default user identity so commits work.
func makeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-C", dir, "config", "user.email", "test@example.com"},
		{"-C", dir, "config", "user.name", "Test"},
		{"-C", dir, "config", "commit.gpgsign", "false"},
	} {
		if args[0] == "init" {
			args = append([]string{"init"}, args[1:]...)
			args = append(args, dir)
		}
		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func stage(t *testing.T, repo, rel, body string) {
	t.Helper()
	full := filepath.Join(repo, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", repo, "add", rel).CombinedOutput(); err != nil {
		t.Fatalf("git add %s: %v: %s", rel, err, out)
	}
}

func runWithEnv(t *testing.T, args []string, stdin string) (int, string) {
	t.Helper()
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", auditPath)
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "")
	var stderr bytes.Buffer
	code := run(args, strings.NewReader(stdin), &stderr)
	return code, stderr.String()
}

// ---- --scan-staged integration -----------------------------------------

func TestScanStaged_CleanRepo(t *testing.T) {
	repo := makeRepo(t)
	stage(t, repo, "src/main.go", "package main\n")
	code, stderr := runWithEnv(t, []string{"--scan-staged", "--repo", repo}, "")
	if code != exitAllow {
		t.Fatalf("got exit %d, stderr=%s", code, stderr)
	}
}

func TestScanStaged_BlocksEnvFile(t *testing.T) {
	repo := makeRepo(t)
	stage(t, repo, ".env", "SECRET=abc\n")
	code, stderr := runWithEnv(t, []string{"--scan-staged", "--repo", repo}, "")
	if code != exitBlock {
		t.Fatalf("got exit %d, want block; stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "NV-S-001") {
		t.Errorf("expected NV-S-001 in stderr; got %s", stderr)
	}
}

func TestScanStaged_BlocksHugeFile(t *testing.T) {
	repo := makeRepo(t)
	huge := strings.Repeat("x", 6*1024*1024+1)
	stage(t, repo, "huge.bin", huge)
	code, stderr := runWithEnv(t, []string{"--scan-staged", "--repo", repo}, "")
	if code != exitBlock {
		t.Fatalf("got exit %d, want block", code)
	}
	if !strings.Contains(stderr, "NV-N-003") {
		t.Errorf("expected NV-N-003 in stderr; got %s", stderr)
	}
}

func TestScanStaged_WarnsLargeFile(t *testing.T) {
	repo := makeRepo(t)
	large := strings.Repeat("y", 2*1024*1024)
	stage(t, repo, "big.bin", large)
	code, stderr := runWithEnv(t, []string{"--scan-staged", "--repo", repo}, "")
	if code != exitAllow {
		t.Fatalf("got exit %d, want allow (with warn); stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "GR-N-001") {
		t.Errorf("expected GR-N-001 warning; got %s", stderr)
	}
}

func TestScanStaged_BypassAllows(t *testing.T) {
	repo := makeRepo(t)
	stage(t, repo, ".env", "SECRET=abc\n")
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "1")
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", auditPath)
	var stderr bytes.Buffer
	code := run([]string{"--scan-staged", "--repo", repo}, strings.NewReader(""), &stderr)
	if code != exitAllow {
		t.Fatalf("bypass should allow; got %d", code)
	}
	// Audit should record the bypass.
	raw, _ := os.ReadFile(auditPath)
	if !strings.Contains(string(raw), `"decision":"bypass"`) {
		t.Errorf("expected bypass in audit log; got %s", raw)
	}
}

// ---- --validate-msg integration ----------------------------------------

func TestValidateMsg_HappyPath(t *testing.T) {
	dir := t.TempDir()
	msg := filepath.Join(dir, "MSG")
	os.WriteFile(msg, []byte("feat(scope): add thing\n\nWhy: tests need this.\n"), 0o644)
	code, stderr := runWithEnv(t, []string{"--validate-msg", msg}, "")
	if code != exitAllow {
		t.Fatalf("got %d, stderr=%s", code, stderr)
	}
}

func TestValidateMsg_BlocksMissingBody(t *testing.T) {
	dir := t.TempDir()
	msg := filepath.Join(dir, "MSG")
	os.WriteFile(msg, []byte("feat: add\n"), 0o644)
	code, stderr := runWithEnv(t, []string{"--validate-msg", msg}, "")
	if code != exitBlock {
		t.Fatalf("got %d, want block", code)
	}
	if !strings.Contains(stderr, "IN-M-003") {
		t.Errorf("expected IN-M-003; got %s", stderr)
	}
}

// ---- PreToolUse integration --------------------------------------------

func TestPreToolUse_IgnoresNonBash(t *testing.T) {
	input, _ := json.Marshal(ToolInput{ToolName: "Write"})
	code, _ := runWithEnv(t, nil, string(input))
	if code != exitAllow {
		t.Errorf("non-Bash should allow; got %d", code)
	}
}

func TestPreToolUse_IgnoresNonCommit(t *testing.T) {
	input, _ := json.Marshal(ToolInput{
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "git status"},
	})
	code, _ := runWithEnv(t, nil, string(input))
	if code != exitAllow {
		t.Errorf("git status should allow; got %d", code)
	}
}

func TestPreToolUse_BlocksMissingBody(t *testing.T) {
	repo := makeRepo(t)
	stage(t, repo, "src/main.go", "package main\n")
	input, _ := json.Marshal(ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]any{
			"command": "git -C " + repo + " commit -m \"fix bug\"",
		},
	})
	code, stderr := runWithEnv(t, nil, string(input))
	if code != exitBlock {
		t.Fatalf("got %d, want block; stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "IN-M-002") && !strings.Contains(stderr, "IN-M-003") {
		t.Errorf("expected IN-M-* in stderr; got %s", stderr)
	}
}

func TestPreToolUse_BlocksStagedEnv(t *testing.T) {
	repo := makeRepo(t)
	stage(t, repo, ".env", "SECRET=abc\n")
	input, _ := json.Marshal(ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]any{
			"command": "git -C " + repo + " commit -m \"chore: env\" -m \"why\"",
		},
	})
	code, stderr := runWithEnv(t, nil, string(input))
	if code != exitBlock {
		t.Fatalf("got %d, want block", code)
	}
	if !strings.Contains(stderr, "NV-S-001") {
		t.Errorf("expected NV-S-001; got %s", stderr)
	}
}

func TestPreToolUse_HappyPath(t *testing.T) {
	repo := makeRepo(t)
	stage(t, repo, "src/main.go", "package main\n")
	input, _ := json.Marshal(ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]any{
			"command": "git -C " + repo + " commit -m \"feat(scope): add main\" -m \"Sets up the entry point.\"",
		},
	})
	code, stderr := runWithEnv(t, nil, string(input))
	if code != exitAllow {
		t.Fatalf("got %d, stderr=%s", code, stderr)
	}
}

// ---- IN-H-001 amend block (record immutability) -------------------------

func TestPreToolUse_BlocksAmend(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", auditPath)
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "")
	input, _ := json.Marshal(ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]any{
			"command": "git commit --amend --no-edit",
		},
	})
	var stderr bytes.Buffer
	code := run(nil, strings.NewReader(string(input)), &stderr)
	if code != exitBlock {
		t.Fatalf("got %d, want block; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "IN-H-001") {
		t.Errorf("expected IN-H-001 in stderr; got %s", stderr.String())
	}
	if strings.Contains(stderr.String(), "RR_COMMIT_GUARD_BYPASS") {
		t.Errorf("amend block must not print the bypass hint; got %s", stderr.String())
	}
	raw, _ := os.ReadFile(auditPath)
	if !strings.Contains(string(raw), `"rule":"IN-H-001"`) {
		t.Errorf("expected IN-H-001 in audit log; got %s", raw)
	}
	if !strings.Contains(string(raw), `"mode":"pretooluse"`) {
		t.Errorf("expected pretooluse mode in audit log; got %s", raw)
	}
}

func TestPreToolUse_BlocksAmend_BypassIgnored(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", auditPath)
	// The bypass var is evaluated only after the amend check, so it cannot
	// waive IN-H-001. Pins the no-bypass rule.
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "1")
	input, _ := json.Marshal(ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]any{
			"command": "git commit --amend --no-edit",
		},
	})
	var stderr bytes.Buffer
	code := run(nil, strings.NewReader(string(input)), &stderr)
	if code != exitBlock {
		t.Fatalf("bypass must not waive IN-H-001; got %d, stderr=%s", code, stderr.String())
	}
}

func TestPreToolUse_BlocksAmend_Compound(t *testing.T) {
	input, _ := json.Marshal(ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]any{
			"command": "git commit --amend --no-edit && git push origin main",
		},
	})
	code, stderr := runWithEnv(t, nil, string(input))
	if code != exitBlock {
		t.Fatalf("got %d, want block; stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "IN-H-001") {
		t.Errorf("expected IN-H-001 in stderr; got %s", stderr)
	}
}
