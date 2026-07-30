// cache_test.go -- pipeline-key determinism and short-cache roundtrip.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// initRepo creates a minimal git repo with one commit so PipelineKey
// can call git rev-parse against it.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.email=t@test", "-c", "user.name=t", "config", "user.email", "t@test"},
		{"-c", "user.name=t", "config", "user.name", "t"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	// Seed file + commit.
	seed := filepath.Join(dir, "seed")
	if err := writeFile(seed, "x\n"); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "seed"},
		{"commit", "-q", "-m", "seed"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func writeFile(p, body string) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(body)
	return err
}

func TestPipelineKey_Deterministic(t *testing.T) {
	repo := initRepo(t)
	k1, err := PipelineKey(repo, []Toolchain{ToolchainGo})
	if err != nil {
		t.Fatal(err)
	}
	k2, err := PipelineKey(repo, []Toolchain{ToolchainGo})
	if err != nil {
		t.Fatal(err)
	}
	if k1 != k2 {
		t.Fatalf("non-deterministic: %q vs %q", k1, k2)
	}
	if len(k1) != 64 {
		t.Errorf("expected 64-char hex sha256, got %d chars", len(k1))
	}
}

func TestPipelineKey_DiffersByToolchainSet(t *testing.T) {
	repo := initRepo(t)
	a, _ := PipelineKey(repo, []Toolchain{ToolchainGo})
	b, _ := PipelineKey(repo, []Toolchain{ToolchainGo, ToolchainNode})
	if a == b {
		t.Fatal("toolchain change should change key")
	}
}

func TestPipelineKey_DiffersByDiff(t *testing.T) {
	repo := initRepo(t)
	a, _ := PipelineKey(repo, []Toolchain{ToolchainGo})
	// Modify working tree without committing.
	if err := writeFile(filepath.Join(repo, "seed"), "different\n"); err != nil {
		t.Fatal(err)
	}
	b, _ := PipelineKey(repo, []Toolchain{ToolchainGo})
	if a == b {
		t.Fatal("working-tree diff should change key")
	}
}

func TestPipelineKey_DiffersByStagedDiff(t *testing.T) {
	repo := initRepo(t)
	a, err := PipelineKey(repo, []Toolchain{ToolchainGo})
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(repo, "staged.go"), "package staged\n"); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", repo, "add", "staged.go")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add staged.go: %v: %s", err, out)
	}
	b, err := PipelineKey(repo, []Toolchain{ToolchainGo})
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("staged index change should change key")
	}
}

func TestShortCacheRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	type rec struct {
		N int    `json:"n"`
		S string `json:"s"`
	}
	in := rec{N: 7, S: "hi"}
	if err := shortCacheWrite("runners", "k1", in); err != nil {
		t.Fatal(err)
	}
	var out rec
	if !shortCacheRead("runners", "k1", time.Hour, &out) {
		t.Fatal("expected cache hit")
	}
	if out != in {
		t.Errorf("roundtrip mismatch: %+v vs %+v", in, out)
	}
}

func TestShortCacheTTL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	type rec struct {
		N int `json:"n"`
	}
	if err := shortCacheWrite("runners", "k", rec{1}); err != nil {
		t.Fatal(err)
	}
	var out rec
	// TTL of 0 means anything is expired.
	if shortCacheRead("runners", "k", 0, &out) {
		t.Error("zero TTL should never be a hit")
	}
}

func TestPipelineCacheRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := StorePipeline("abc123", PipelineCacheMeta{Repo: "/x", Toolchains: []Toolchain{ToolchainGo}}); err != nil {
		t.Fatal(err)
	}
	if !LookupPipeline("abc123", 30) {
		t.Fatal("expected cache hit after store")
	}
	if LookupPipeline("does-not-exist", 30) {
		t.Fatal("missing key should not be a hit")
	}
}
