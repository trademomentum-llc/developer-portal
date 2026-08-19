// audit_test.go -- hash-chain tests for the brew-guard audit writer
// (RECORD-IMMUTABILITY-TECH-001 section 10.4).
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

const genesisZeros = "0000000000000000000000000000000000000000000000000000000000000000"

// readAuditLines splits the log into its raw lines (trailing newlines
// stripped by the split).
func readAuditLines(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
}

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
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_BREW_GUARD_AUDIT_LOG", path)

	logAudit("allow", "", "brew list", "s1")
	logAudit("block", "dangerous flag", "brew install --force x", "s1")

	lines := readAuditLines(t, path)
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
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_BREW_GUARD_AUDIT_LOG", path)

	logAudit("allow", "", "brew list", "s1")
	// A write-only log defeats the tail read; the append must still happen
	// and the line must carry the 64-zero fallback (TECH-001 9.2).
	if err := os.Chmod(path, 0o200); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	logAudit("block", "dangerous flag", "brew install --force x", "s1")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("chmod back: %v", err)
	}

	lines := readAuditLines(t, path)
	if len(lines) != 2 {
		t.Fatalf("fail-open broken: want 2 lines, got %d", len(lines))
	}
	if got := prevHashOf(t, lines[1]); got != genesisZeros {
		t.Errorf("tail-read failure prev_hash = %q, want 64 zeros", got)
	}
}
