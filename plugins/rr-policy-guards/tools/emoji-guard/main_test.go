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
	old := os.Getenv("RR_EMOJI_GUARD_AUDIT_LOG")
	os.Setenv("RR_EMOJI_GUARD_AUDIT_LOG", path)
	return path, func() {
		if old == "" {
			os.Unsetenv("RR_EMOJI_GUARD_AUDIT_LOG")
		} else {
			os.Setenv("RR_EMOJI_GUARD_AUDIT_LOG", old)
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

func TestRun_WriteCleanAscii_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Write", map[string]any{
		"file_path": "/tmp/x.go",
		"content":   "package main\n\nfunc main() {}\n",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d. stderr: %s", code, exitAllow, stderr.String())
	}
}

func TestRun_WriteWithEmDash_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Write", map[string]any{
		"file_path": "/tmp/x.md",
		"content":   "hello \u2014 world", // em dash
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
	if !strings.Contains(stderr.String(), "non-ASCII") {
		t.Errorf("stderr does not mention non-ASCII: %s", stderr.String())
	}
}

func TestRun_EditWithBoxDrawing_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Edit", map[string]any{
		"file_path":  "/tmp/x.md",
		"old_string": "old",
		"new_string": "\u2500\u2500\u2500 new \u2500\u2500\u2500",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
}

func TestRun_MultiEditAllClean_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("MultiEdit", map[string]any{
		"file_path": "/tmp/x.go",
		"edits": []any{
			map[string]any{"old_string": "a", "new_string": "alpha"},
			map[string]any{"old_string": "b", "new_string": "beta"},
		},
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d", code, exitAllow)
	}
}

func TestRun_MultiEditOneBad_Blocks(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("MultiEdit", map[string]any{
		"file_path": "/tmp/x.go",
		"edits": []any{
			map[string]any{"old_string": "a", "new_string": "alpha"},
			map[string]any{"old_string": "b", "new_string": "beta \u2014 extra"},
		},
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitBlock {
		t.Errorf("exit code = %d, want %d", code, exitBlock)
	}
}

func TestRun_BashTool_IgnoredAndAllowed(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	// Bash tools are out of scope; the hook must not care what command it contains.
	in := bytes.NewReader(makeInput("Bash", map[string]any{
		"command":     "brew install \u2014 fake",
		"description": "pointless test",
	}))
	var stderr bytes.Buffer
	code := run(in, &stderr)
	if code != exitAllow {
		t.Errorf("exit code = %d, want %d", code, exitAllow)
	}
}

func TestRun_Bypass_Allows(t *testing.T) {
	_, cleanup := tempAuditLog(t)
	defer cleanup()

	os.Setenv("RR_EMOJI_GUARD_BYPASS", "1")
	defer os.Unsetenv("RR_EMOJI_GUARD_BYPASS")

	in := bytes.NewReader(makeInput("Write", map[string]any{
		"file_path": "/tmp/x.md",
		"content":   "hello \u2014 world",
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

func TestRun_AuditLogWritten(t *testing.T) {
	path, cleanup := tempAuditLog(t)
	defer cleanup()

	in := bytes.NewReader(makeInput("Write", map[string]any{
		"file_path": "/tmp/x.go",
		"content":   "clean\n",
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
