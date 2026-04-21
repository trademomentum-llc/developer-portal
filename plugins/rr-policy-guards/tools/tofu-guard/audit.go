// Package main: rr-tofu-guard audit log writer.
//
// Format: JSONL (one JSON object per line), file mode 0600, parent dir 0700.
// Default path: $HOME/.rational-reserve/logs/tofu-guard.jsonl
// Override path via RR_TOFU_GUARD_AUDIT_LOG env var.
//
// Error handling: INTENTIONALLY BEST-EFFORT. Every I/O failure (bad home dir,
// bad mkdir, bad open, bad write) is silently swallowed and logAudit returns.
// Rationale: policy enforcement must never depend on a working audit
// subsystem. The guard's exit code still blocks or allows correctly; only
// the forensic record is lost. The tradeoff is deliberate and mirrors the
// pattern established by rr-brew-guard/audit.go. If you need guaranteed
// audit delivery, add a stderr fallback line on error (see tech debt
// guard-1 in TODO.md for the full discussion).
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type AuditEvent struct {
	Timestamp string `json:"ts"`
	Action    string `json:"action"`
	Reason    string `json:"reason,omitempty"`
	Command   string `json:"command,omitempty"`
	Session   string `json:"session,omitempty"`
}

func logAudit(action, reason, command, session string) {
	path := os.Getenv("RR_TOFU_GUARD_AUDIT_LOG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		path = filepath.Join(home, ".rational-reserve", "logs", "tofu-guard.jsonl")
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
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
	}
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = f.Write(data)
}
