// auditlock_test.go -- concurrency tests for the audit append lock.
//
// The production incident these tests guard against: several agent
// harnesses (Claude Code, Codex, Kimi) run the guards concurrently on one
// machine, and the unsynchronized read-tail-hash -> append cycle let two
// writers chain from the same tail (or both write a genesis line into a
// fresh log), breaking the hash chain rr-audit-chain verifies.
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

// TestAudit_ConcurrentWriters hammers one log from N goroutines; the
// flock in withAuditLock must serialize the appends so the chain verifies.
func TestAudit_ConcurrentWriters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_EMOJI_GUARD_AUDIT_LOG", path)

	const writers = 16
	const appendsEach = 25
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < appendsEach; i++ {
				logAudit("allow", "concurrency-probe", "Write", fmt.Sprintf("w%d", w))
			}
		}(w)
	}
	wg.Wait()
	verifyChain(t, path, writers*appendsEach)
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
			logAudit("allow", "process-probe", "Write", "worker")
		}
		return
	}

	path := filepath.Join(t.TempDir(), "audit.jsonl")
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
				"RR_EMOJI_GUARD_AUDIT_LOG="+path,
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
	verifyChain(t, path, procs*linesEach)
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
