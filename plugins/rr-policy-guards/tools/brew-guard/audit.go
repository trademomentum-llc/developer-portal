// audit.go -- append-only JSONL audit logging for the brew guard.
//
// Mirrors the emoji-guard and bash-guard audit pattern.
//
// Each line carries prev_hash, the SHA-256 of the previous raw line
// (trailing newline included), making the log tamper-evident per
// RECORD-IMMUTABILITY-TECH-001 section 9.1.
//
// The audit log path resolves from (in order):
//
//  1. RR_BREW_GUARD_AUDIT_LOG env var (for tests and overrides)
//  2. ~/.rational-reserve/logs/brew-guard.jsonl (production default)
//
// Logging is best-effort. If the log cannot be written, the hook does not
// fail -- policy enforcement must not depend on a working audit subsystem.

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

// AuditEvent is the JSON shape written to the audit log.
type AuditEvent struct {
	Timestamp string `json:"ts"`
	Action    string `json:"action"`
	Reason    string `json:"reason,omitempty"`
	Command   string `json:"command,omitempty"`
	Session   string `json:"session,omitempty"`
	PrevHash  string `json:"prev_hash"`
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

// logAudit appends a single event to the audit log. Errors are swallowed
// because an audit failure must never become a policy failure.
func logAudit(action, reason, command, session string) {
	path := os.Getenv("RR_BREW_GUARD_AUDIT_LOG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		path = filepath.Join(home, ".rational-reserve", "logs", "brew-guard.jsonl")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	evt := AuditEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Action:    action,
		Reason:    reason,
		Command:   command,
		Session:   session,
		PrevHash:  prevHashForPath(path),
	}
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = f.Write(data)
}
