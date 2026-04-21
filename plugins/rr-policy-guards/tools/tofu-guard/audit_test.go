package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogAuditWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tofu-guard.jsonl")
	t.Setenv("RR_TOFU_GUARD_AUDIT_LOG", path)

	logAudit("block", "apply", "tofu apply", "sess-1")
	logAudit("allow", "", "tofu plan", "sess-2")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d", len(lines))
	}

	var first map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("json: %v", err)
	}
	if first["action"] != "block" || first["reason"] != "apply" || first["command"] != "tofu apply" {
		t.Fatalf("bad payload: %v", first)
	}
}

func TestLogAuditPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tofu-guard.jsonl")
	t.Setenv("RR_TOFU_GUARD_AUDIT_LOG", path)

	logAudit("allow", "", "tofu version", "sess-3")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("perm=%o want 0600", perm)
	}
}
