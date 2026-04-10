// audit.go -- append-only JSONL audit logging for the bash guard.
//
// Mirrors the emoji-guard audit pattern. Every allow, block, or bypass
// decision writes one line of JSON to the audit log.
//
// The audit log path resolves from (in order):
//
//  1. RR_BASH_GUARD_AUDIT_LOG env var (for tests and overrides)
//  2. ~/.rational-reserve/logs/bash-guard.jsonl (production default)
//
// Logging is best-effort. If the log cannot be written, the hook does not
// fail -- policy enforcement must not depend on a working audit subsystem.

package main

import (
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
	Tool      string `json:"tool,omitempty"`
	Session   string `json:"session,omitempty"`
}

// logAudit appends a single event to the audit log. Errors are swallowed
// because an audit failure must never become a policy failure.
func logAudit(action, reason, tool, session string) {
	path := os.Getenv("RR_BASH_GUARD_AUDIT_LOG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		path = filepath.Join(home, ".rational-reserve", "logs", "bash-guard.jsonl")
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
		Tool:      tool,
		Session:   session,
	}
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = f.Write(data)
}
