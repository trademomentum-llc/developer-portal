package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "rr-tofu-guard")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}
	return bin
}

func runGuard(t *testing.T, bin, stdin string, env map[string]string) (int, string, string) {
	t.Helper()
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	}
	return code, out.String(), errb.String()
}

func TestBinaryAllowsPlan(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Bash","tool_input":{"command":"tofu plan"},"session_id":"s1"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, _ := runGuard(t, bin, in, map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 0 {
		t.Fatalf("exit=%d want 0", code)
	}
}

func TestBinaryBlocksApply(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Bash","tool_input":{"command":"tofu apply"},"session_id":"s2"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, stderr := runGuard(t, bin, in, map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 2 {
		t.Fatalf("exit=%d want 2; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, "blocked") {
		t.Fatalf("stderr missing 'blocked': %q", stderr)
	}
}

func TestBinaryBypass(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Bash","tool_input":{"command":"tofu apply"},"session_id":"s3"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, _ := runGuard(t, bin, in, map[string]string{
		"RR_TOFU_GUARD_AUDIT_LOG": log,
		"RR_TOFU_GUARD_BYPASS":    "1",
	})
	if code != 0 {
		t.Fatalf("exit=%d want 0 (bypass)", code)
	}
}

func TestBinaryIgnoresNonBashTool(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Write","tool_input":{"command":"tofu apply"},"session_id":"s4"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, _ := runGuard(t, bin, in, map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 0 {
		t.Fatalf("exit=%d want 0 (non-Bash ignored)", code)
	}
}
