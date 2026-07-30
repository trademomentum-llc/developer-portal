// audit_test.go -- writeAudit appends well-formed JSONL and respects
// the override env var.

package main

import (
	"bufio"
	"bytes"
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
