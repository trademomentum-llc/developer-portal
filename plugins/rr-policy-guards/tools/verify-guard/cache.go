// cache.go -- on-disk caches keyed by content hashes.
//
// Four cache domains:
//   keys/    full-pipeline result, key = sha256(repo, tree, index, worktree, toolchains, schema)
//   forge/   per-host forge identification (24h TTL, see forge.go)
//   runners/ per <forge,owner,repo> runner inventory (60s TTL)
//   actions/ per <forge,owner,repo,ref> ref existence (60s TTL)

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// cacheRoot returns ~/.rational-reserve/cache/verify-guard, creating
// it at 0700 if missing.
func cacheRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(home, ".rational-reserve", "cache", "verify-guard")
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	return root, nil
}

// PipelineKey derives the sha256 of the committed tree, staged index diff,
// unstaged working-tree diff, detected toolchains, and cache schema.
func PipelineKey(repoRoot string, toolchains []Toolchain) (string, error) {
	tree, err := exec.Command("git", "-C", repoRoot, "rev-parse", "HEAD^{tree}").CombinedOutput()
	treeHash := strings.TrimSpace(string(tree))
	if err != nil {
		treeHash = "no-head"
	}
	stagedDiff, err := exec.Command("git", "-C", repoRoot, "diff", "--cached", "--no-color", "--binary").CombinedOutput()
	if err != nil {
		return "", err
	}
	stagedSum := sha256.Sum256(stagedDiff)
	stagedHash := hex.EncodeToString(stagedSum[:])

	worktreeDiff, err := exec.Command("git", "-C", repoRoot, "diff", "--no-color", "--binary").CombinedOutput()
	if err != nil {
		return "", err
	}
	worktreeSum := sha256.Sum256(worktreeDiff)
	worktreeHash := hex.EncodeToString(worktreeSum[:])

	tcs := make([]string, len(toolchains))
	for i, t := range toolchains {
		tcs[i] = string(t)
	}

	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\x00%s\x00%s",
		repoRoot, treeHash, stagedHash, worktreeHash, strings.Join(tcs, ","), cacheSchema)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// PipelineCacheMeta is what we persist for a successful pipeline run.
type PipelineCacheMeta struct {
	SchemaVersion string      `json:"schema_version"`
	Repo          string      `json:"repo"`
	Toolchains    []Toolchain `json:"toolchains"`
	Forge         Forge       `json:"forge"`
	VerifiedAt    string      `json:"verified_at"`
	DurationMS    int64       `json:"duration_ms"`
	ActUsed       bool        `json:"act_used"`
}

// LookupPipeline returns true iff a cache file exists for hexKey and is
// younger than maxAgeDays.
func LookupPipeline(hexKey string, maxAgeDays int) bool {
	root, err := cacheRoot()
	if err != nil {
		return false
	}
	path := filepath.Join(root, "keys", hexKey+".json")
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if time.Since(info.ModTime()) > time.Duration(maxAgeDays)*24*time.Hour {
		return false
	}
	return true
}

// StorePipeline persists the meta for hexKey.
func StorePipeline(hexKey string, meta PipelineCacheMeta) error {
	root, err := cacheRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "keys")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if meta.SchemaVersion == "" {
		meta.SchemaVersion = cacheSchema
	}
	if meta.VerifiedAt == "" {
		meta.VerifiedAt = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, hexKey+".json"), b, 0600)
}

// generic short-TTL cache for runners + actions.

func shortCacheRead(domain, key string, ttl time.Duration, dst any) bool {
	root, err := cacheRoot()
	if err != nil {
		return false
	}
	path := filepath.Join(root, domain, key+".json")
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if time.Since(info.ModTime()) > ttl {
		return false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return json.Unmarshal(b, dst) == nil
}

func shortCacheWrite(domain, key string, src any) error {
	root, err := cacheRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, domain)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, key+".json"), b, 0600)
}

// EvictExpired removes cache files older than maxAge in the named
// domain. Best-effort; errors are silently ignored.
func EvictExpired(domain string, maxAge time.Duration) {
	root, err := cacheRoot()
	if err != nil {
		return
	}
	dir := filepath.Join(root, domain)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
