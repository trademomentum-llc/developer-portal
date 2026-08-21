// audit.go -- JSONL audit log writer.
//
// One line per invocation, append-only, mode 0600. Path resolved from
// RR_VERIFY_GUARD_AUDIT_LOG when set, otherwise
// ~/.rational-reserve/logs/verify-guard.jsonl.
//
// Each line carries prev_hash, the SHA-256 of the previous raw line
// (trailing newline included), making the log tamper-evident per
// RECORD-IMMUTABILITY-TECH-001 section 9.1. The hash is taken from the
// pre-rotation active log: whether or not rotateAuditLog then moves that
// log aside, the previous last line stays the chain predecessor (the
// rotated-away log keeps it as its own last line).

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultAuditMaxBytes = int64(8 * 1024 * 1024)
	auditBackupCount     = 3
)

// AuditLine is one record emitted to the audit log.
type AuditLine struct {
	Ts           string      `json:"ts"`
	Action       string      `json:"action"`
	Reason       string      `json:"reason"`
	Tool         string      `json:"tool"`
	Command      string      `json:"command,omitempty"`
	Session      string      `json:"session,omitempty"`
	Repo         string      `json:"repo,omitempty"`
	Forge        Forge       `json:"forge,omitempty"`
	Toolchains   []Toolchain `json:"toolchains,omitempty"`
	DurationMS   int64       `json:"duration_ms,omitempty"`
	ActUsed      bool        `json:"act_used,omitempty"`
	RunID        string      `json:"run_id,omitempty"`
	GuardVersion string      `json:"guard_version"`
	PrevHash     string      `json:"prev_hash"`
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

// auditLogPath returns the configured audit log path, creating the
// parent directory at 0700 if missing.
func auditLogPath() (string, error) {
	if p := os.Getenv("RR_VERIFY_GUARD_AUDIT_LOG"); p != "" {
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return "", err
		}
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".rational-reserve", "logs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "verify-guard.jsonl"), nil
}

// writeAudit appends one JSON line to the audit log.
func writeAudit(line AuditLine) error {
	if line.Ts == "" {
		line.Ts = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if line.GuardVersion == "" {
		line.GuardVersion = guardVersion
	}
	if line.Tool == "" {
		line.Tool = "Bash"
	}
	path, err := auditLogPath()
	if err != nil {
		return fmt.Errorf("audit path: %w", err)
	}
	// Chain-tail read, rotation, and append are one critical section: they
	// must hold the inter-process lock end to end, or concurrent guards
	// chain from the same tail (or rotate under each other) and break the
	// chain (see auditlock.go).
	return withAuditLock(path, func() error {
		// Chain to the current last line before rotation runs: with or without
		// rotation, that line remains the predecessor of the one written below.
		line.PrevHash = prevHashForPath(path)
		data, err := json.Marshal(line)
		if err != nil {
			return fmt.Errorf("audit marshal: %w", err)
		}
		data = append(data, '\n')
		if err := rotateAuditLog(path, auditMaxBytes("RR_VERIFY_GUARD_AUDIT_MAX_BYTES"), auditBackupCount, int64(len(data))); err != nil {
			return fmt.Errorf("audit rotate: %w", err)
		}
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("audit open: %w", err)
		}
		defer f.Close()
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("audit write: %w", err)
		}
		return nil
	})
}

func auditMaxBytes(envName string) int64 {
	if raw := os.Getenv(envName); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil && value > 0 {
			return value
		}
	}
	return defaultAuditMaxBytes
}

func rotateAuditLog(path string, maxBytes int64, backups int, incomingBytes int64) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Size()+incomingBytes <= maxBytes {
		return nil
	}
	oldest := path + "." + strconv.Itoa(backups)
	if err := os.Remove(oldest); err != nil && !os.IsNotExist(err) {
		return err
	}
	for index := backups - 1; index >= 1; index-- {
		source := path + "." + strconv.Itoa(index)
		destination := path + "." + strconv.Itoa(index+1)
		if err := os.Rename(source, destination); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return os.Rename(path, path+".1")
}

// auditQuiet writes an audit line and silently suppresses any error so
// the caller can still emit a useful exit decision. The audit failing
// must never itself block a commit.
func auditQuiet(line AuditLine) {
	_ = writeAudit(line)
}
