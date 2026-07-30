// util.go -- small helpers shared across modules.

package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// readFileSafely reads up to maxBytes from path; returns the bytes
// (possibly truncated) and any read error encountered before the cap.
func readFileSafely(path string, maxBytes int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	limited := io.LimitReader(f, maxBytes)
	return io.ReadAll(limited)
}

// firstLines returns the first n lines of s, joined by '\n'.
// If s has fewer than n lines, returns s unchanged.
func firstLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	count := 0
	for i, r := range s {
		if r == '\n' {
			count++
			if count == n {
				return s[:i]
			}
		}
	}
	return s
}

// commandOnPath returns true if the named command resolves on PATH.
func commandOnPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// trimWhitespace removes leading/trailing whitespace including \r.
func trimWhitespace(s string) string {
	return strings.TrimSpace(s)
}
