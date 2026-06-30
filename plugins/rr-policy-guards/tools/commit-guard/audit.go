// audit.go -- append-only JSONL audit log writer.
package main

import (
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

func appendAuditStrict(r AuditRecord) error {
	if r.TS == "" {
		r.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}
	path := auditPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}
