// bypass.go -- RR_COMMIT_GUARD_BYPASS=1 handling.
package main

import "os"

// bypassActive reports whether the user has set the bypass env var.
// Mirrors the pattern used by bash-guard / emoji-guard / tofu-guard.
func bypassActive() bool {
	return os.Getenv("RR_COMMIT_GUARD_BYPASS") == "1"
}
