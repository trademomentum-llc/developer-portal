package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudit_RoundTrip(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", tmp)

	if err := appendAuditStrict(AuditRecord{
		Decision: DecisionBlock,
		Mode:     "scan-staged",
		Rule:     "NV-S-001",
		Paths:    []string{".env"},
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := appendAuditStrict(AuditRecord{
		Decision: DecisionAllow,
		Mode:     "scan-staged",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	f, err := os.Open(tmp)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	var got []AuditRecord
	for sc.Scan() {
		var rec AuditRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			t.Fatalf("invalid JSON line: %s -- %v", sc.Text(), err)
		}
		got = append(got, rec)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0].Decision != DecisionBlock || got[0].Rule != "NV-S-001" {
		t.Errorf("first line wrong: %+v", got[0])
	}
	if got[1].Decision != DecisionAllow {
		t.Errorf("second line wrong: %+v", got[1])
	}
}

func TestAudit_TimestampAutoFilled(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", tmp)

	if err := appendAuditStrict(AuditRecord{
		Decision: DecisionAllow, Mode: "validate-msg",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	raw, _ := os.ReadFile(tmp)
	if !strings.Contains(string(raw), `"ts":`) {
		t.Fatal("expected ts field in audit line")
	}
}

func TestBypass_EnvVar(t *testing.T) {
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "1")
	if !bypassActive() {
		t.Error("expected bypassActive to be true")
	}
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "")
	if bypassActive() {
		t.Error("expected bypassActive to be false when unset")
	}
}
