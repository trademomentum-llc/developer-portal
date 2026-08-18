package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitOut runs git inside repo and returns trimmed stdout, failing the test
// on any error.
func gitOut(t *testing.T, repo string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", repo}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", full, err, out)
	}
	return strings.TrimSpace(string(out))
}

// makeLinearRepo builds a scratch repo with two commits on main and returns
// the repo path and the two commit SHAs (shaA is the parent of shaB).
func makeLinearRepo(t *testing.T) (repo, shaA, shaB string) {
	t.Helper()
	repo = makeRepo(t)
	stage(t, repo, "a.txt", "a\n")
	gitOut(t, repo, "commit", "-q", "-m", "feat: a", "-m", "why a")
	shaA = gitOut(t, repo, "rev-parse", "HEAD")
	stage(t, repo, "b.txt", "b\n")
	gitOut(t, repo, "commit", "-q", "-m", "feat: b", "-m", "why b")
	shaB = gitOut(t, repo, "rev-parse", "HEAD")
	return repo, shaA, shaB
}

// chdir switches the process working directory for tests that exercise git
// commands run in cwd (isAncestor), restoring the previous directory at test
// cleanup. (t.Chdir needs go1.24; this module targets go 1.23.)
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

// runPrePushEnv drives --pre-push mode with the audit log routed into a
// temp dir and the bypass var blanked. It returns the exit code, stderr,
// and the audit log path for content assertions.
func runPrePushEnv(t *testing.T, stdin string) (int, string, string) {
	t.Helper()
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", auditPath)
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "")
	var stderr bytes.Buffer
	code := run([]string{"--pre-push", "origin", "https://example.com/repo.git"},
		strings.NewReader(stdin), &stderr)
	return code, stderr.String(), auditPath
}

// ---- parser --------------------------------------------------------------

func TestParseRefUpdates(t *testing.T) {
	shaA := strings.Repeat("a", 40)
	shaB := strings.Repeat("b", 40)

	two := "refs/heads/main " + shaA + " refs/heads/main " + shaB + "\n" +
		"refs/heads/feature " + shaB + " refs/heads/feature " + zeroSHA + "\n"
	updates, err := parseRefUpdates(two)
	if err != nil {
		t.Fatalf("two well-formed lines: %v", err)
	}
	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2", len(updates))
	}
	if updates[0] != (refUpdate{"refs/heads/main", shaA, "refs/heads/main", shaB}) {
		t.Errorf("got %+v", updates[0])
	}
	if updates[1].remoteRef != "refs/heads/feature" || updates[1].remoteSHA != zeroSHA {
		t.Errorf("got %+v", updates[1])
	}

	// A single trailing newline is tolerated (the normal git framing).
	one := "refs/heads/main " + shaA + " refs/heads/main " + shaB + "\n"
	updates, err = parseRefUpdates(one)
	if err != nil || len(updates) != 1 {
		t.Errorf("trailing newline: updates=%v err=%v", updates, err)
	}

	// Empty stdin is no updates.
	updates, err = parseRefUpdates("")
	if err != nil || len(updates) != 0 {
		t.Errorf("empty stdin: updates=%v err=%v", updates, err)
	}

	// A 3-field line is malformed.
	if _, err = parseRefUpdates("refs/heads/main " + shaA + " refs/heads/main\n"); err == nil {
		t.Error("3-field line should error")
	}
}

// ---- decision logic ------------------------------------------------------

func TestRewritesMain(t *testing.T) {
	repo, shaA, shaB := makeLinearRepo(t)
	chdir(t, repo) // isAncestor shells out to git in the current directory

	cases := []struct {
		name       string
		update     refUpdate
		wantBlock  bool
		wantReason string
	}{
		{"fast-forward allowed",
			refUpdate{"refs/heads/main", shaB, "refs/heads/main", shaA}, false, ""},
		{"non-fast-forward blocked",
			refUpdate{"refs/heads/main", shaA, "refs/heads/main", shaB}, true, "non-fast-forward update of refs/heads/main"},
		{"main deletion blocked",
			refUpdate{"refs/heads/main", zeroSHA, "refs/heads/main", shaA}, true, "deletion of refs/heads/main"},
		{"main creation allowed",
			refUpdate{"refs/heads/main", shaA, "refs/heads/main", zeroSHA}, false, ""},
		{"similar branch name allowed",
			refUpdate{"refs/heads/main-x", shaA, "refs/heads/main-x", shaB}, false, ""},
		{"tag named main allowed",
			refUpdate{"refs/tags/main", shaA, "refs/tags/main", shaB}, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			block, reason := rewritesMain(tc.update)
			if block != tc.wantBlock || reason != tc.wantReason {
				t.Errorf("got (%v, %q), want (%v, %q)", block, reason, tc.wantBlock, tc.wantReason)
			}
		})
	}
}

func TestIsAncestor(t *testing.T) {
	repo, shaA, shaB := makeLinearRepo(t)
	chdir(t, repo)

	if !isAncestor(shaA, shaB) {
		t.Errorf("linear history: %s should be an ancestor of %s", shaA, shaB)
	}
	if isAncestor(shaB, shaA) {
		t.Errorf("linear history: %s should not be an ancestor of %s", shaB, shaA)
	}

	// Diverged history: branch off shaA and commit on the branch.
	gitOut(t, repo, "checkout", "-q", "-b", "feature", shaA)
	stage(t, repo, "c.txt", "c\n")
	gitOut(t, repo, "commit", "-q", "-m", "feat: c", "-m", "why c")
	shaC := gitOut(t, repo, "rev-parse", "HEAD")
	if isAncestor(shaB, shaC) {
		t.Errorf("diverged history: %s should not be an ancestor of %s", shaB, shaC)
	}
	if !isAncestor(shaA, shaC) {
		t.Errorf("diverged history: %s should be an ancestor of %s", shaA, shaC)
	}
}

// ---- runPrePush mode -----------------------------------------------------

func TestRunPrePush_AllowsFastForward(t *testing.T) {
	repo, shaA, shaB := makeLinearRepo(t)
	chdir(t, repo)

	line := "refs/heads/main " + shaB + " refs/heads/main " + shaA + "\n"
	code, stderr, auditPath := runPrePushEnv(t, line)
	if code != exitAllow {
		t.Fatalf("got %d, want allow; stderr=%s", code, stderr)
	}
	raw, _ := os.ReadFile(auditPath)
	if !strings.Contains(string(raw), `"decision":"allow"`) ||
		!strings.Contains(string(raw), `"mode":"pre-push"`) {
		t.Errorf("expected allow/pre-push audit line; got %s", raw)
	}
}

func TestRunPrePush_AllowsCreationAndNonMainDeletion(t *testing.T) {
	shaA := strings.Repeat("a", 40)
	lines := "refs/heads/main " + shaA + " refs/heads/main " + zeroSHA + "\n" + // main creation
		"refs/heads/feature " + shaA + " refs/heads/feature " + zeroSHA + "\n" + // branch creation
		"refs/heads/old " + zeroSHA + " refs/heads/old " + shaA + "\n" // non-main deletion
	code, stderr, _ := runPrePushEnv(t, lines)
	if code != exitAllow {
		t.Fatalf("got %d, want allow; stderr=%s", code, stderr)
	}
}

func TestRunPrePush_AllowsNonMainNonFastForward(t *testing.T) {
	repo, shaA, shaB := makeLinearRepo(t)
	chdir(t, repo)

	// Policy scope is refs/heads/main only (FR-002): a non-fast-forward on
	// any other ref is allowed without even consulting ancestry.
	line := "refs/heads/feature " + shaA + " refs/heads/feature " + shaB + "\n"
	code, stderr, _ := runPrePushEnv(t, line)
	if code != exitAllow {
		t.Fatalf("got %d, want allow; stderr=%s", code, stderr)
	}
}

func TestRunPrePush_BlocksNonFastForward(t *testing.T) {
	repo, shaA, shaB := makeLinearRepo(t)
	chdir(t, repo)

	line := "refs/heads/main " + shaA + " refs/heads/main " + shaB + "\n"
	code, stderr, auditPath := runPrePushEnv(t, line)
	if code != exitBlock {
		t.Fatalf("got %d, want block; stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "IN-H-002") ||
		!strings.Contains(stderr, "non-fast-forward update of refs/heads/main") {
		t.Errorf("expected IN-H-002 non-fast-forward message; got %s", stderr)
	}
	if strings.Contains(stderr, "RR_COMMIT_GUARD_BYPASS") {
		t.Errorf("pre-push block must not print the bypass hint; got %s", stderr)
	}
	raw, _ := os.ReadFile(auditPath)
	if !strings.Contains(string(raw), `"rule":"IN-H-002"`) ||
		!strings.Contains(string(raw), `"mode":"pre-push"`) ||
		!strings.Contains(string(raw), `"command":"push origin https://example.com/repo.git"`) {
		t.Errorf("expected IN-H-002/pre-push audit line with push command; got %s", raw)
	}
}

func TestRunPrePush_BlocksMainDeletion(t *testing.T) {
	shaA := strings.Repeat("a", 40)
	line := "refs/heads/main " + zeroSHA + " refs/heads/main " + shaA + "\n"
	code, stderr, auditPath := runPrePushEnv(t, line)
	if code != exitBlock {
		t.Fatalf("got %d, want block; stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "deletion of refs/heads/main") {
		t.Errorf("expected deletion reason; got %s", stderr)
	}
	raw, _ := os.ReadFile(auditPath)
	if !strings.Contains(string(raw), `"rule":"IN-H-002"`) {
		t.Errorf("expected IN-H-002 in audit log; got %s", raw)
	}
}

func TestRunPrePush_BypassIgnored(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	t.Setenv("RR_COMMIT_GUARD_AUDIT_LOG", auditPath)
	// The pre-push path never consults bypassActive(). Pins the no-bypass rule.
	t.Setenv("RR_COMMIT_GUARD_BYPASS", "1")
	line := "refs/heads/main " + zeroSHA + " refs/heads/main " + strings.Repeat("a", 40) + "\n"
	var stderr bytes.Buffer
	code := run([]string{"--pre-push", "origin", "https://example.com/repo.git"},
		strings.NewReader(line), &stderr)
	if code != exitBlock {
		t.Fatalf("bypass must not waive IN-H-002; got %d, stderr=%s", code, stderr.String())
	}
}

func TestRunPrePush_MalformedStdin(t *testing.T) {
	code, stderr, _ := runPrePushEnv(t, "refs/heads/main onlythree fields\n")
	if code != exitInternal {
		t.Fatalf("got %d, want internal error; stderr=%s", code, stderr)
	}
}
