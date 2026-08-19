// audit_test.go -- writeAudit appends well-formed JSONL and respects
// the override env var.

package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAudit_OverridePath(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "verify-guard.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)

	if err := writeAudit(AuditLine{Action: "allow", Reason: "test", Tool: "Bash"}); err != nil {
		t.Fatal(err)
	}
	if err := writeAudit(AuditLine{Action: "block", Reason: "test2"}); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, l := range lines {
		var rec AuditLine
		if err := json.Unmarshal([]byte(l), &rec); err != nil {
			t.Fatalf("line %d malformed: %v", i, err)
		}
		if rec.Ts == "" {
			t.Errorf("line %d missing ts", i)
		}
		if rec.GuardVersion == "" {
			t.Errorf("line %d missing guard_version", i)
		}
		if rec.Tool == "" {
			t.Errorf("line %d defaults Tool to Bash but got empty", i)
		}
	}
}

func TestWriteAudit_AppendsNewLineEach(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "v.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)

	for i := 0; i < 5; i++ {
		if err := writeAudit(AuditLine{Action: "allow", Reason: "n"}); err != nil {
			t.Fatal(err)
		}
	}
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	count := strings.Count(string(b), "\n")
	if count != 5 {
		t.Errorf("expected 5 newlines (one per line), got %d", count)
	}
}

func TestWriteAudit_NeverWritesRawTokens(t *testing.T) {
	// AuditLine has no Token field by design -- this test enforces that
	// the struct never grows one without an audit-test update.
	dir := t.TempDir()
	logPath := filepath.Join(dir, "v.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)
	t.Setenv("RR_VERIFY_GUARD_GITHUB_TOKEN", "ghp_secrettoken_should_not_appear_in_log")

	if err := writeAudit(AuditLine{Action: "allow", Reason: "verified"}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(logPath)
	if strings.Contains(string(b), "ghp_secrettoken_should_not_appear_in_log") {
		t.Fatal("token leaked into audit log")
	}
}

func TestWriteAudit_RotatesOversizedLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "verify-guard.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)
	t.Setenv("RR_VERIFY_GUARD_AUDIT_MAX_BYTES", "128")
	if err := os.WriteFile(logPath, bytes.Repeat([]byte("x"), 256), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := writeAudit(AuditLine{Action: "allow", Reason: "test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatalf("rotated audit backup missing: %v", err)
	}
}

const genesisZeros = "0000000000000000000000000000000000000000000000000000000000000000"

// rawAuditLines splits the log into its raw lines (trailing newlines
// stripped by the split).
func rawAuditLines(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
}

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
	logPath := filepath.Join(dir, "verify-guard.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)

	if err := writeAudit(AuditLine{Action: "allow", Reason: "one"}); err != nil {
		t.Fatal(err)
	}
	if err := writeAudit(AuditLine{Action: "block", Reason: "two"}); err != nil {
		t.Fatal(err)
	}

	lines := rawAuditLines(t, logPath)
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
	dir := t.TempDir()
	logPath := filepath.Join(dir, "verify-guard.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)

	if err := writeAudit(AuditLine{Action: "allow", Reason: "one"}); err != nil {
		t.Fatal(err)
	}
	// A write-only log defeats the tail read; the append must still happen
	// and the line must carry the 64-zero fallback (TECH-001 9.2).
	if err := os.Chmod(logPath, 0o200); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if err := writeAudit(AuditLine{Action: "block", Reason: "two"}); err != nil {
		t.Fatalf("write over unreadable tail: %v", err)
	}
	if err := os.Chmod(logPath, 0o600); err != nil {
		t.Fatalf("chmod back: %v", err)
	}

	lines := rawAuditLines(t, logPath)
	if len(lines) != 2 {
		t.Fatalf("fail-open broken: want 2 lines, got %d", len(lines))
	}
	if got := prevHashOf(t, lines[1]); got != genesisZeros {
		t.Errorf("tail-read failure prev_hash = %q, want 64 zeros", got)
	}
}

func TestAudit_PrevHashRotation(t *testing.T) {
	// Rotation must not break the chain: the first line of the fresh active
	// log chains to the last line of the segment just rotated aside.
	dir := t.TempDir()
	logPath := filepath.Join(dir, "verify-guard.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)
	t.Setenv("RR_VERIFY_GUARD_AUDIT_MAX_BYTES", "100")

	if err := writeAudit(AuditLine{Action: "allow", Reason: "one"}); err != nil {
		t.Fatal(err)
	}
	if err := writeAudit(AuditLine{Action: "allow", Reason: "two"}); err != nil {
		t.Fatal(err)
	}

	rotated := rawAuditLines(t, logPath+".1")
	active := rawAuditLines(t, logPath)
	if len(rotated) != 1 || len(active) != 1 {
		t.Fatalf("rotation shape wrong: .1=%d lines, active=%d lines", len(rotated), len(active))
	}
	sum := sha256.Sum256([]byte(rotated[0] + "\n"))
	want := hex.EncodeToString(sum[:])
	if got := prevHashOf(t, active[0]); got != want {
		t.Errorf("post-rotation prev_hash = %q, want %q (sha256 of rotated last line)", got, want)
	}
}
