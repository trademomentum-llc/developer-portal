// audit.go -- JSONL audit log writer.
//
// One line per invocation, append-only, mode 0600. Path resolved from
// RR_VERIFY_GUARD_AUDIT_LOG when set, otherwise
// ~/.rational-reserve/logs/verify-guard.jsonl.

package main

import (
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
