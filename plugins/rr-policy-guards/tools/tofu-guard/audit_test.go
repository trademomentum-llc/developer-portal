package main

import (
	"crypto/sha256"
	"encoding/hex"
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

const genesisZeros = "0000000000000000000000000000000000000000000000000000000000000000"

// prevHashOf extracts the prev_hash field from one raw audit line.
func prevHashOf(t *testing.T, line string) string {
	t.Helper()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(line), &parsed); err != nil {
		t.Fatalf("unparseable audit line: %v", err)
	}
	value, _ := parsed["prev_hash"].(string)
	return value
}

func TestAudit_PrevHashChain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tofu-guard.jsonl")
	t.Setenv("RR_TOFU_GUARD_AUDIT_LOG", path)

	logAudit("block", "apply", "tofu apply", "sess-1")
	logAudit("allow", "", "tofu plan", "sess-2")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d", len(lines))
	}
	if got := prevHashOf(t, lines[0]); got != genesisZeros {
		t.Errorf("genesis prev_hash = %q, want 64 zeros", got)
	}
	sum := sha256.Sum256([]byte(lines[0] + "\n"))
	if want := hex.EncodeToString(sum[:]); prevHashOf(t, lines[1]) != want {
		t.Errorf("line 2 prev_hash = %q, want %q", prevHashOf(t, lines[1]), want)
	}
}

func TestAudit_PrevHashFailOpen(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the chmod-based read denial this test simulates")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "tofu-guard.jsonl")
	t.Setenv("RR_TOFU_GUARD_AUDIT_LOG", path)

	logAudit("allow", "", "tofu version", "sess-1")
	// A write-only log defeats the tail read; the append must still happen
	// and the line must carry the 64-zero fallback (TECH-001 9.2).
	if err := os.Chmod(path, 0o200); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	logAudit("allow", "", "tofu plan", "sess-2")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("chmod back: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("fail-open broken: want 2 lines, got %d", len(lines))
	}
	if got := prevHashOf(t, lines[1]); got != genesisZeros {
		t.Errorf("tail-read failure prev_hash = %q, want 64 zeros", got)
	}
}
