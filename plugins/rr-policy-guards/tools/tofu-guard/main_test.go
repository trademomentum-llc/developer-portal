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

func TestBinaryUnparseableInput(t *testing.T) {
	bin := buildBinary(t)
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, stderr := runGuard(t, bin, "not valid json", map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 2 {
		t.Fatalf("exit=%d want 2", code)
	}
	if !strings.Contains(stderr, "unable to parse") {
		t.Fatalf("stderr missing 'unable to parse': %q", stderr)
	}
	// Audit should have an unparseable-input block entry with err detail
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if !strings.Contains(string(data), "unparseable-input") {
		t.Fatalf("audit missing unparseable-input reason: %s", data)
	}
}

func TestBinaryMalformedQuoting(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Bash","tool_input":{"command":"tofu apply 'unterminated"},"session_id":"s5"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, stderr := runGuard(t, bin, in, map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 2 {
		t.Fatalf("exit=%d want 2", code)
	}
	if !strings.Contains(stderr, "malformed command") {
		t.Fatalf("stderr missing 'malformed command': %q", stderr)
	}
}

func TestBinaryShellMetacharacter(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Bash","tool_input":{"command":"tofu plan; rm -rf /tmp/test"},"session_id":"s6"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, stderr := runGuard(t, bin, in, map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 2 {
		t.Fatalf("exit=%d want 2", code)
	}
	if !strings.Contains(stderr, "shell metacharacter") {
		t.Fatalf("stderr missing 'shell metacharacter': %q", stderr)
	}
}

func TestBinaryHeredocWithApostropheAllowed(t *testing.T) {
	// Regression: tofu-guard.jsonl 2026-04-21T18:20:20 recorded a block with
	// reason "tokenize-error: unterminated quote" for a cat heredoc that
	// contained an apostrophe (Gitea's) and mentioned "tofu apply" as
	// literal text. The command was not a tofu invocation and must not be
	// blocked just because the tokenizer cannot handle heredocs.
	heredoc := "cat <<'EOF' | wc -w\n## M2 heredoc\nGitea's OCI registry references to tofu apply in text.\nEOF"
	in, err := json.Marshal(map[string]any{
		"tool_name":  "Bash",
		"tool_input": map[string]any{"command": heredoc},
		"session_id": "s-heredoc",
	})
	if err != nil {
		t.Fatal(err)
	}
	bin := buildBinary(t)
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, stderr := runGuard(t, bin, string(in), map[string]string{"RR_TOFU_GUARD_AUDIT_LOG": log})
	if code != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q", code, stderr)
	}
}

func TestBinaryBypassAuditEvent(t *testing.T) {
	bin := buildBinary(t)
	in := `{"tool_name":"Bash","tool_input":{"command":"tofu apply"},"session_id":"s7"}`
	log := filepath.Join(t.TempDir(), "a.jsonl")
	code, _, _ := runGuard(t, bin, in, map[string]string{
		"RR_TOFU_GUARD_AUDIT_LOG": log,
		"RR_TOFU_GUARD_BYPASS":    "1",
	})
	if code != 0 {
		t.Fatalf("exit=%d want 0", code)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if !strings.Contains(string(data), `"action":"bypass"`) {
		t.Fatalf("audit missing bypass action: %s", data)
	}
	if !strings.Contains(string(data), `"reason":"`) {
		t.Fatalf("audit missing reason: %s", data)
	}
}
