// audit.go -- append-only JSONL audit log writer.
//
// Each line carries prev_hash, the SHA-256 of the previous raw line
// (trailing newline included), making the log tamper-evident per
// RECORD-IMMUTABILITY-TECH-001 section 9.1.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// auditPath returns the audit log destination, honouring the env override.
func auditPath() string {
	if override := os.Getenv("RR_COMMIT_GUARD_AUDIT_LOG"); override != "" {
		return override
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Last resort if home dir is unreadable.
		return "/tmp/rr-commit-guard.jsonl"
	}
	return filepath.Join(home, ".rational-reserve", "logs", "commit-guard.jsonl")
}

// appendAudit writes one JSON line to the audit log. The function fails open
// (returns silently) when the log is unwritable so it never blocks a commit
// for an unrelated I/O failure. Callers that need fail-closed behaviour should
// invoke appendAuditStrict.
func appendAudit(r AuditRecord) {
	_ = appendAuditStrict(r)
}

// genesisPrevHash is the prev_hash of a chain's first line (TECH-001 9.1)
// and the fail-open fallback when the tail cannot be read (TECH-001 9.2):
// chaining must never alter or block the enforcement decision.
const genesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"

// prevHashForPath returns the prev_hash for the next appended line:
// SHA-256 hex of the raw bytes of the current last line, trailing newline
// included. Genesis and any tail-read error both yield 64 zeros.
func prevHashForPath(path string) string {
	last, err := lastLineBytes(path)
	if err != nil || len(last) == 0 {
		return genesisPrevHash
	}
	sum := sha256.Sum256(last)
	return hex.EncodeToString(sum[:])
}

// lastLineBytes returns the raw bytes of the last line of the log at path,
// including its trailing newline. If the active log is missing or empty
// (fresh log, or a rotation that just moved it aside) the last line of
// path+".1" is used instead. No log at all returns nil, nil (genesis).
func lastLineBytes(path string) ([]byte, error) {
	line, err := tailLine(path)
	if os.IsNotExist(err) || (err == nil && len(line) == 0) {
		return tailLine(path + ".1")
	}
	return line, err
}

// tailLine reads the final line of one file, trailing newline included.
// A missing file surfaces its open error; an empty file returns nil, nil.
// Only the last 1 MiB is scanned -- audit lines are small.
func tailLine(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()
	if size == 0 {
		return nil, nil
	}
	const window = 1 << 20
	start := int64(0)
	if size > window {
		start = size - window
	}
	buf := make([]byte, size-start)
	if _, err := f.ReadAt(buf, start); err != nil {
		return nil, err
	}
	body := buf
	if body[len(body)-1] == '\n' {
		body = body[:len(body)-1]
	}
	idx := bytes.LastIndexByte(body, '\n')
	return buf[idx+1:], nil
}

func appendAuditStrict(r AuditRecord) error {
	if r.TS == "" {
		r.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}
	path := auditPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// The tail read and the append are one critical section: they must run
	// under the inter-process lock or concurrent guards break the chain
	// (see auditlock.go).
	return withAuditLock(path, func() error {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		defer f.Close()
		r.PrevHash = prevHashForPath(path)
		line, err := json.Marshal(r)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
		return nil
	})
}
