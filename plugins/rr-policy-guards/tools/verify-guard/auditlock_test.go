// auditlock_test.go -- concurrency tests for the audit append lock.
//
// The production incident these tests guard against: several agent
// harnesses (Claude Code, Codex, Kimi) run the guards concurrently on one
// machine, and the unsynchronized read-tail-hash -> rotate -> append cycle
// let two writers chain from the same tail (or rotate under each other),
// breaking the hash chain rr-audit-chain verifies.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// verifyChain re-walks the active log and recomputes every prev_hash
// link, failing on the first mismatch (a test-local rr-audit-chain).
func verifyChain(t *testing.T, path string, wantLines int) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) != wantLines {
		t.Fatalf("want %d lines, got %d", wantLines, len(lines))
	}
	for i, line := range lines {
		want := genesisZeros
		if i > 0 {
			sum := sha256.Sum256([]byte(lines[i-1] + "\n"))
			want = hex.EncodeToString(sum[:])
		}
		if got := prevHashOf(t, line); got != want {
			t.Fatalf("line %d: prev_hash = %q, want %q", i+1, got, want)
		}
	}
}

// verifySegmentedChain re-walks the rotated segments (.3, .2, .1, active)
// in chain order and recomputes every prev_hash link across segment
// boundaries. Mirrors rr-audit-chain: if .3 exists, rotation may have
// dropped older history, so the first walked line is exempt from the
// genesis-zeros requirement.
func verifySegmentedChain(t *testing.T, path string) {
	t.Helper()
	var segments []string
	dropped := false
	for i := 3; i >= 1; i-- {
		seg := path + "." + strconv.Itoa(i)
		if _, err := os.Stat(seg); err == nil {
			segments = append(segments, seg)
			if i == 3 {
				dropped = true
			}
		}
	}
	segments = append(segments, path)

	var prevRaw string
	first := true
	total := 0
	for _, seg := range segments {
		raw, err := os.ReadFile(seg)
		if err != nil {
			t.Fatalf("read %s: %v", seg, err)
		}
		for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
			if line == "" {
				continue
			}
			total++
			got := prevHashOf(t, line)
			if first {
				first = false
				if !dropped && got != genesisZeros {
					t.Fatalf("%s: first line prev_hash = %q, want genesis zeros", seg, got)
				}
			} else {
				sum := sha256.Sum256([]byte(prevRaw + "\n"))
				if want := hex.EncodeToString(sum[:]); got != want {
					t.Fatalf("%s: prev_hash = %q, want %q (sha256 of previous raw line)", seg, got, want)
				}
			}
			prevRaw = line
		}
	}
	if total == 0 {
		t.Fatal("no audit lines found in any segment")
	}
}

// TestAudit_ConcurrentWriters hammers one log from N goroutines; the
// flock in withAuditLock must serialize the appends so the chain verifies.
func TestAudit_ConcurrentWriters(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)

	const writers = 16
	const appendsEach = 25
	var wg sync.WaitGroup
	var failures int64
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < appendsEach; i++ {
				if err := writeAudit(AuditLine{Action: "allow", Reason: fmt.Sprintf("concurrency-probe w%d", w)}); err != nil {
					atomic.AddInt64(&failures, 1)
				}
			}
		}(w)
	}
	wg.Wait()
	if failures > 0 {
		t.Fatalf("%d writeAudit calls failed", failures)
	}
	verifyChain(t, logPath, writers*appendsEach)
}

// TestAudit_ConcurrentWritersRotation forces a rotation on nearly every
// append with a tiny size budget; under the lock the chain must stay
// intact across the segment boundaries.
func TestAudit_ConcurrentWritersRotation(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_VERIFY_GUARD_AUDIT_LOG", logPath)
	t.Setenv("RR_VERIFY_GUARD_AUDIT_MAX_BYTES", "300")

	const writers = 8
	const appendsEach = 10
	var wg sync.WaitGroup
	var failures int64
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < appendsEach; i++ {
				if err := writeAudit(AuditLine{Action: "allow", Reason: fmt.Sprintf("rotation-probe w%d", w)}); err != nil {
					atomic.AddInt64(&failures, 1)
				}
			}
		}(w)
	}
	wg.Wait()
	if failures > 0 {
		t.Fatalf("%d writeAudit calls failed", failures)
	}
	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatalf("expected rotation to produce %s.1: %v", logPath, err)
	}
	verifySegmentedChain(t, logPath)
}

// TestAudit_ConcurrentProcesses re-executes the test binary as N separate
// OS processes appending to one log -- the exact multi-harness race that
// broke the production logs -- then verifies the chain.
func TestAudit_ConcurrentProcesses(t *testing.T) {
	if n := os.Getenv("RR_AUDIT_WORKER_LINES"); n != "" {
		// Worker mode: append n lines to the log named by the env override.
		count, err := strconv.Atoi(n)
		if err != nil {
			t.Fatalf("RR_AUDIT_WORKER_LINES: %v", err)
		}
		for i := 0; i < count; i++ {
			if err := writeAudit(AuditLine{Action: "allow", Reason: "process-probe"}); err != nil {
				t.Fatalf("worker append: %v", err)
			}
		}
		return
	}

	logPath := filepath.Join(t.TempDir(), "audit.jsonl")
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	const procs = 8
	const linesEach = 10
	var wg sync.WaitGroup
	var failures int64
	for p := 0; p < procs; p++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			cmd := exec.Command(exe, "-test.run", "^TestAudit_ConcurrentProcesses$", "-test.count=1")
			cmd.Env = append(os.Environ(),
				"RR_VERIFY_GUARD_AUDIT_LOG="+logPath,
				"RR_AUDIT_WORKER_LINES="+strconv.Itoa(linesEach),
			)
			if out, err := cmd.CombinedOutput(); err != nil {
				atomic.AddInt64(&failures, 1)
				t.Logf("worker %d failed: %v\n%s", p, err, out)
			}
		}(p)
	}
	wg.Wait()
	if failures > 0 {
		t.Fatalf("%d of %d worker processes failed", failures, procs)
	}
	verifyChain(t, logPath, procs*linesEach)
}

// TestWithAuditLock_Serializes proves mutual exclusion with a lost-update
// counter pattern: without the lock the paired load/store loses updates;
// under the lock the final count must be exact. The counter is atomic
// only to keep the race detector quiet -- the test's power comes from the
// load/store split, not from atomicity.
func TestWithAuditLock_Serializes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	const writers = 8
	const iters = 50
	var counter atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				err := withAuditLock(path, func() error {
					v := counter.Load()
					counter.Store(v + 1)
					return nil
				})
				if err != nil {
					t.Errorf("withAuditLock: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if got := counter.Load(); got != writers*iters {
		t.Errorf("counter = %d, want %d (lock did not serialize)", got, writers*iters)
	}
	// The lock file lives alongside the log and is never deleted.
	if _, err := os.Stat(path + ".lock"); err != nil {
		t.Errorf("lock file missing: %v", err)
	}
}
