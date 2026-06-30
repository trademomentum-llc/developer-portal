// git.go -- shell out to git to enumerate staged paths and their sizes.
package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// StagedPaths returns the list of staged paths (relative to the repo root)
// using `git diff --cached --name-only -z`. The -z flag is critical so paths
// with newlines do not break parsing.
func StagedPaths(repoRoot string) ([]string, error) {
	args := []string{"diff", "--cached", "--name-only", "-z"}
	cmd := exec.Command("git", args...)
	if repoRoot != "" {
		cmd.Dir = repoRoot
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached failed: %w", err)
	}
	parts := bytes.Split(out, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		paths = append(paths, string(p))
	}
	return paths, nil
}

// StagedSize returns the size in bytes of a single staged blob. Falls back to
// 0 (i.e. "size unknown, skip size-based rules") when the blob cannot be read.
func StagedSize(repoRoot, path string) int64 {
	args := []string{}
	if repoRoot != "" {
		args = append(args, "-C", repoRoot)
	}
	args = append(args, "cat-file", "-s", ":"+path)
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return 0
	}
	n, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return n
}
