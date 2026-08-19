package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
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
	tmp := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", tmp)

	if err := appendAuditStrict(AuditRecord{
		Decision: DecisionBlock,
		Mode:     "scan-staged",
		Rule:     "NV-S-001",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := appendAuditStrict(AuditRecord{
		Decision: DecisionAllow,
		Mode:     "scan-staged",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	raw, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
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
	tmp := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", tmp)

	if err := appendAuditStrict(AuditRecord{Decision: DecisionAllow, Mode: "scan-staged"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	// A write-only log defeats the tail read; the append must still happen
	// and the line must carry the 64-zero fallback (TECH-001 9.2).
	if err := os.Chmod(tmp, 0o200); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if err := appendAuditStrict(AuditRecord{Decision: DecisionBlock, Mode: "scan-staged"}); err != nil {
		t.Fatalf("append over unreadable tail: %v", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		t.Fatalf("chmod back: %v", err)
	}

	raw, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("fail-open broken: want 2 lines, got %d", len(lines))
	}
	if got := prevHashOf(t, lines[1]); got != genesisZeros {
		t.Errorf("tail-read failure prev_hash = %q, want 64 zeros", got)
	}
}
