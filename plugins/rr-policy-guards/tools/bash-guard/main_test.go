// main_test.go -- integration tests for the PreToolUse entrypoint.
//
// These tests exercise run() directly with in-memory stdin/stderr, covering
// the end-to-end decision paths: allow, block, bypass, out-of-scope,
// malformed input.

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tempAuditLog returns a temp audit log path and sets the env var so the
// hook writes to it. Returns a cleanup function.
func tempAuditLog(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	old := os.Getenv("RR_BASH_GUARD_AUDIT_LOG")
	os.Setenv("RR_BASH_GUARD_AUDIT_LOG", path)
	return path, func() {
		if old == "" {
			os.Unsetenv("RR_BASH_GUARD_AUDIT_LOG")
		} else {
			os.Setenv("RR_BASH_GUARD_AUDIT_LOG", old)
		}
	}
}

// makeInput builds a JSON byte slice simulating a PreToolUse hook input.
func makeInput(tool string, toolInput map[string]any) []byte {
	out, _ := json.Marshal(map[string]any{
		"session_id": "test-session",
		"tool_name":  tool,
		"tool_input": toolInput,
	})
	return out
}

func TestRun_SafeBashCommand_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "ls -la /tmp",
		"description": "list temp files",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d. stderr: %s", code, exitAllow, stderr.String())
	}
}

func TestRun_BareVar_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "ls $HOME/projects",
		"description": "list projects",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
	if !strings.Contains(stderr.String(), "BLOCKED") {
		t.Errorf("stderr does not contain BLOCKED: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "corrected command") {
		t.Errorf("stderr does not contain corrected command: %s", stderr.String())
	}
}

func TestRun_WriteTool_IgnoredAndAllowed(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Write", map[string]any{
		"file_path": "/tmp/x.go",
		"content":   "package main with $HOME in it",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d (Write is out of scope)", code, exitAllow)
	}
}

func TestRun_Bypass_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	os.Setenv("RR_BASH_GUARD_BYPASS", "1")
	defer os.Unsetenv("RR_BASH_GUARD_BYPASS")

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "cd $HOME",
		"description": "go home",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("bypass: exit code = %d, want %d", code, exitAllow)
	}
	if !strings.Contains(stderr.String(), "bypass") {
		t.Errorf("bypass: stderr does not mention bypass: %s", stderr.String())
	}
}

func TestRun_MalformedJson_BlocksFailClosed(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader([]byte("{not json at all"))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("malformed: exit code = %d, want %d (fail closed)", code, exitBlock)
	}
}

func TestRun_ExitStatus_Allowed(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "some_cmd; echo $?",
		"description": "check exit code",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d ($? is safe)", code, exitAllow)
	}
}

func TestRun_CommandSubstitution_Allowed(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "echo $(date)",
		"description": "print date",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d ($() is safe)", code, exitAllow)
	}
}

func TestRun_AuditLogWritten(t *testing.T) {
	path, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "ls /tmp",
		"description": "safe command",
	}))
	var stderr bytes.Buffer
	run(in, &stderr)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("audit log not written: %v", err)
	}
	if !strings.Contains(string(data), `"action":"allow"`) {
		t.Errorf("audit log missing allow event: %s", string(data))
	}
}

func TestRun_BlockMessageIncludesCorrectedCommand(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "cp $HOME/file /tmp/dest",
		"description": "copy file",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
	output := stderr.String()
	if !strings.Contains(output, "<HOME_VALUE>") {
		t.Errorf("corrected command should contain placeholder: %s", output)
	}
}
