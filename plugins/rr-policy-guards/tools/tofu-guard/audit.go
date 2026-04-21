// Package main: rr-tofu-guard audit log writer. JSONL, mode 0600.
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
