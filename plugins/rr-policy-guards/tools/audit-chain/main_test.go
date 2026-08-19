// main_test.go -- table tests for rr-audit-chain.
//
// Fixtures are built by chainedLines, which computes prev_hash exactly the
// way the guards do, so tamper tests can corrupt bytes deliberately.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chainedLines renders payloads into raw audit lines (newline included)
// linked per TECH-001 9.1: line 0 carries the 64-zero genesis hash, line N
// carries SHA-256 of line N-1's raw bytes.
func chainedLines(payloads ...string) [][]byte {
	return chainedLinesFrom(genesisPrevHash, payloads...)
}

// chainedLinesFrom is chainedLines with a caller-chosen first-line
// prev_hash, for fixtures whose true genesis was rotated away.
func chainedLinesFrom(firstPrevHash string, payloads ...string) [][]byte {
	lines := make([][]byte, 0, len(payloads))
	var prev []byte
	for i, p := range payloads {
		ph := firstPrevHash
		if i > 0 {
			sum := sha256.Sum256(prev)
			ph = hex.EncodeToString(sum[:])
		}
		line := []byte(fmt.Sprintf(`{"seq":%d,"data":%q,"prev_hash":%q}`, i, p, ph) + "\n")
		lines = append(lines, line)
		prev = line
	}
	return lines
}

func writeLog(t *testing.T, path string, lines ...[]byte) {
	t.Helper()
	if err := os.WriteFile(path, bytes.Join(lines, nil), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func runCLI(argv ...string) (int, string) {
	var out bytes.Buffer
	code := run(append([]string{"rr-audit-chain"}, argv...), &out)
	return code, out.String()
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestVerify_IntactChain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ok.jsonl")
	lines := chainedLines("alpha", "beta", "gamma", "delta")
	writeLog(t, path, lines...)

	code, out := runCLI("verify", path)
	if code != exitIntact {
		t.Fatalf("exit = %d, want 0; out: %s", code, out)
	}
	if !strings.Contains(out, "chain intact") || !strings.Contains(out, "4 lines") {
		t.Errorf("unexpected OK output: %q", out)
	}
	if want := sha256Hex(lines[3]); !strings.Contains(out, "head "+want) {
		t.Errorf("head = %s not found in %q", want, out)
	}
}

func TestVerify_TamperedMiddleLineDetectedAtEditedLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tampered.jsonl")
	lines := chainedLines("alpha", "beta", "gamma", "delta", "epsilon")
	// Hand-edit line 3's content without touching any prev_hash, the way an
	// attacker with file access would.
	lines[2] = bytes.Replace(lines[2], []byte("gamma"), []byte("gamMa"), 1)
	writeLog(t, path, lines...)

	code, out := runCLI("verify", path)
	if code != exitBroken {
		t.Fatalf("exit = %d, want 1; out: %s", code, out)
	}
	if !strings.Contains(out, "line 3") {
		t.Errorf("edited line number 3 not reported: %q", out)
	}
}

func TestVerify_FinalLineRemovalVerifiesClean(t *testing.T) {
	// TECH-001 9.4 honest limit, pinned: deleting the tail is invisible to
	// internal verification; only an externally anchored head detects it.
	path := filepath.Join(t.TempDir(), "truncated.jsonl")
	lines := chainedLines("alpha", "beta", "gamma")
	writeLog(t, path, lines[0], lines[1])

	code, out := runCLI("verify", path)
	if code != exitIntact {
		t.Fatalf("exit = %d, want 0; out: %s", code, out)
	}
	if want := sha256Hex(lines[1]); !strings.Contains(out, "head "+want) {
		t.Errorf("head of truncated log = %s not found in %q", want, out)
	}
}

func TestVerify_RotatedSegmentsWalk(t *testing.T) {
	// Synthetic verify-guard log set: .1 holds the oldest retained lines,
	// the active file continues the chain across the segment boundary.
	dir := t.TempDir()
	path := filepath.Join(dir, "verify-guard.jsonl")
	lines := chainedLines("one", "two", "three", "four")
	writeLog(t, path+".1", lines[0], lines[1])
	writeLog(t, path, lines[2], lines[3])

	code, out := runCLI("verify", path)
	if code != exitIntact {
		t.Fatalf("exit = %d, want 0; out: %s", code, out)
	}
	if !strings.Contains(out, "4 lines, 2 segment(s)") {
		t.Errorf("unexpected OK output: %q", out)
	}
	if want := sha256Hex(lines[3]); !strings.Contains(out, "head "+want) {
		t.Errorf("head = %s not found in %q", want, out)
	}
}

func TestVerify_RotatedBoundaryBreak(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "verify-guard.jsonl")
	lines := chainedLines("one", "two", "three", "four")
	// Tamper with the last line of the rotated segment.
	lines[1] = bytes.Replace(lines[1], []byte("two"), []byte("tWo"), 1)
	writeLog(t, path+".1", lines[0], lines[1])
	writeLog(t, path, lines[2], lines[3])

	code, out := runCLI("verify", path)
	if code != exitBroken {
		t.Fatalf("exit = %d, want 1; out: %s", code, out)
	}
	if !strings.Contains(out, path+".1 line 2") {
		t.Errorf("break not attributed to the rotated segment's line 2: %q", out)
	}
}

func TestVerify_FullBackupSetExemptsDroppedGenesis(t *testing.T) {
	// With .3 present, rotation may already have dropped the true genesis
	// segment, so the first retained line may carry a non-zero prev_hash
	// pointing at deleted history. All remaining links must still verify.
	dir := t.TempDir()
	path := filepath.Join(dir, "verify-guard.jsonl")
	// The first retained line chains to a predecessor rotation deleted.
	lines := chainedLinesFrom(strings.Repeat("a", 64), "a", "b", "c", "d", "e", "f")
	writeLog(t, path+".3", lines[0], lines[1])
	writeLog(t, path+".2", lines[2])
	writeLog(t, path+".1", lines[3])
	writeLog(t, path, lines[4], lines[5])

	code, out := runCLI("verify", path)
	if code != exitIntact {
		t.Fatalf("exit = %d, want 0; out: %s", code, out)
	}
	if !strings.Contains(out, "6 lines, 4 segment(s)") {
		t.Errorf("unexpected OK output: %q", out)
	}
}

func TestVerify_MissingPrevHashIsBroken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.jsonl")
	lines := chainedLines("alpha", "beta", "gamma")
	// A pre-chaining (or stripped) line has no prev_hash at all.
	lines[1] = []byte(`{"seq":1,"data":"beta"}` + "\n")
	writeLog(t, path, lines...)

	code, out := runCLI("verify", path)
	if code != exitBroken {
		t.Fatalf("exit = %d, want 1; out: %s", code, out)
	}
	if !strings.Contains(out, "line 2: missing prev_hash") {
		t.Errorf("missing prev_hash not reported at line 2: %q", out)
	}
}

func TestVerify_GenesisMustBeZeros(t *testing.T) {
	path := filepath.Join(t.TempDir(), "badgenesis.jsonl")
	lines := chainedLines("alpha", "beta")
	lines[0] = bytes.Replace(lines[0], []byte(genesisPrevHash), []byte(strings.Repeat("1", 64)), 1)
	writeLog(t, path, lines...)

	code, out := runCLI("verify", path)
	if code != exitBroken {
		t.Fatalf("exit = %d, want 1; out: %s", code, out)
	}
	if !strings.Contains(out, "line 1") || !strings.Contains(out, "genesis") {
		t.Errorf("genesis violation not reported at line 1: %q", out)
	}
}

func TestVerify_MalformedPrevHashIsBroken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "badhex.jsonl")
	lines := chainedLines("alpha", "beta")
	lines[1] = bytes.Replace(lines[1], []byte(sha256Hex(lines[0])), []byte(strings.Repeat("Z", 64)), 1)
	writeLog(t, path, lines...)

	code, out := runCLI("verify", path)
	if code != exitBroken {
		t.Fatalf("exit = %d, want 1; out: %s", code, out)
	}
	if !strings.Contains(out, "line 2: malformed prev_hash") {
		t.Errorf("malformed prev_hash not reported at line 2: %q", out)
	}
}

func TestVerify_MissingFileIsError(t *testing.T) {
	code, out := runCLI("verify", filepath.Join(t.TempDir(), "nope.jsonl"))
	if code != exitError {
		t.Fatalf("exit = %d, want 2; out: %s", code, out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR output, got %q", out)
	}
}

func TestVerify_EmptyLogIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.jsonl")
	writeLog(t, path)

	code, out := runCLI("verify", path)
	if code != exitError {
		t.Fatalf("exit = %d, want 2; out: %s", code, out)
	}
	if !strings.Contains(out, "empty") {
		t.Errorf("expected empty-log error, got %q", out)
	}
}

func TestHead_PrintsChainTip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ok.jsonl")
	lines := chainedLines("alpha", "beta", "gamma")
	writeLog(t, path, lines...)

	code, out := runCLI("head", path)
	if code != exitIntact {
		t.Fatalf("exit = %d, want 0; out: %s", code, out)
	}
	if strings.TrimSpace(out) != sha256Hex(lines[2]) {
		t.Errorf("head = %q, want %q", strings.TrimSpace(out), sha256Hex(lines[2]))
	}
}

func TestHead_UsesActiveSegmentTip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "verify-guard.jsonl")
	lines := chainedLines("one", "two", "three")
	writeLog(t, path+".1", lines[0])
	writeLog(t, path, lines[1], lines[2])

	code, out := runCLI("head", path)
	if code != exitIntact {
		t.Fatalf("exit = %d, want 0; out: %s", code, out)
	}
	if strings.TrimSpace(out) != sha256Hex(lines[2]) {
		t.Errorf("head = %q, want active tip %q", strings.TrimSpace(out), sha256Hex(lines[2]))
	}
}

func TestUsage_Errors(t *testing.T) {
	for _, argv := range [][]string{
		{},
		{"verify"},
		{"bogus", "x"},
		{"verify", "a", "b"},
	} {
		code, out := runCLI(argv...)
		if code != exitError {
			t.Errorf("argv %v: exit = %d, want 2; out: %s", argv, code, out)
		}
	}
}
