// exec_test.go -- tests for the hardcoded-allowlist command dispatch
// and the per-toolchain default Step generators.

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllowedBinariesLockedSet(t *testing.T) {
	expected := []string{
		"go", "npm", "npx", "cargo",
		"ruff", "mypy", "pytest",
		"cmake", "ctest", "make",
		"mix", "bundle", "mvn", "./gradlew", "act",
		"semgrep", "gitleaks", "yarn", "govulncheck",
	}
	for _, want := range expected {
		if _, ok := allowedBinaries[want]; !ok {
			t.Errorf("allowedBinaries missing %q", want)
		}
	}
	// Spot-check that obviously dangerous binaries are NOT in the set.
	for _, banned := range []string{"sh", "bash", "rm", "curl", "wget", "ssh", "sudo", "python", "node"} {
		if _, ok := allowedBinaries[banned]; ok {
			t.Errorf("allowedBinaries unexpectedly contains %q -- security regression", banned)
		}
	}
}

func TestDispatchExec_Allowed(t *testing.T) {
	ctx := context.Background()
	for binary := range allowedBinaries {
		s := Step{Toolchain: ToolchainGo, Name: "t", Cmd: binary, Args: []string{"--version"}}
		cmd, err := dispatchExec(ctx, s)
		if err != nil {
			t.Errorf("dispatchExec rejected allowlisted %q: %v", binary, err)
			continue
		}
		if cmd == nil {
			t.Errorf("dispatchExec returned nil cmd for %q", binary)
			continue
		}
		// Args should pass through; Path should be the binary name (resolved
		// later by exec at Run time).
		if len(cmd.Args) < 1 || cmd.Args[0] != binary {
			t.Errorf("dispatchExec(%q) cmd.Args[0] = %q, want %q", binary, cmd.Args[0], binary)
		}
	}
}

func TestDispatchExec_Rejected(t *testing.T) {
	ctx := context.Background()
	for _, banned := range []string{"sh", "bash", "rm", "curl", "/bin/sh", "../../bin/sh"} {
		s := Step{Cmd: banned, Args: []string{"-c", "echo pwned"}}
		cmd, err := dispatchExec(ctx, s)
		if err == nil {
			t.Errorf("dispatchExec did NOT reject %q -- security regression", banned)
		}
		if cmd != nil {
			t.Errorf("dispatchExec returned non-nil cmd for rejected %q", banned)
		}
		if !strings.Contains(err.Error(), "allowlist") {
			t.Errorf("error for %q should mention allowlist, got %v", banned, err)
		}
	}
}

func TestDefaultSteps_Go(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	steps := defaultSteps(ToolchainGo, root)
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps for Go, got %d", len(steps))
	}
	if steps[0].Cmd != "go" || steps[0].Args[0] != "vet" {
		t.Errorf("first Go step should be go vet, got %+v", steps[0])
	}
	if steps[1].Cmd != "go" || steps[1].Args[0] != "test" {
		t.Errorf("second Go step should be go test, got %+v", steps[1])
	}
	if steps[0].WorkDir != "" {
		t.Errorf("single-module repo must run steps at the root (empty WorkDir), got %q", steps[0].WorkDir)
	}
}

func TestDefaultSteps_GoMultiModule(t *testing.T) {
	root := t.TempDir()
	// No go.mod at the root: two nested module roots must each get
	// vet+test steps pinned to their own directory.
	for _, mod := range []string{"tools/alpha", "plugins/beta"} {
		dir := filepath.Join(root, mod)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/"+mod+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	steps := defaultSteps(ToolchainGo, root)
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps (vet+test x2 modules), got %d: %+v", len(steps), steps)
	}
	dirs := map[string]int{}
	for _, s := range steps {
		if s.WorkDir == "" {
			t.Errorf("multi-module step %q must set WorkDir, got empty", s.Name)
		}
		dirs[s.WorkDir]++
		if s.Cmd != "go" || (s.Args[0] != "vet" && s.Args[0] != "test") {
			t.Errorf("unexpected step %+v", s)
		}
	}
	for _, mod := range []string{"tools/alpha", "plugins/beta"} {
		if dirs[filepath.Join(root, mod)] != 2 {
			t.Errorf("module %s: expected 2 steps, got %d", mod, dirs[filepath.Join(root, mod)])
		}
	}
}

func TestDefaultSteps_Rust(t *testing.T) {
	steps := defaultSteps(ToolchainRust, "/tmp")
	if len(steps) != 3 {
		t.Fatalf("expected 3 Rust steps, got %d", len(steps))
	}
	wantArgs := [][]string{
		{"fmt", "--", "--check"},
		{"clippy", "--all-targets", "--", "-D", "warnings"},
		{"test", "--quiet"},
	}
	for i, s := range steps {
		if s.Cmd != "cargo" {
			t.Errorf("Rust step %d Cmd = %q, want cargo", i, s.Cmd)
		}
		if len(s.Args) != len(wantArgs[i]) || s.Args[0] != wantArgs[i][0] {
			t.Errorf("Rust step %d Args = %v, want prefix %v", i, s.Args, wantArgs[i])
		}
	}
}

func TestDefaultSteps_UnknownToolchain(t *testing.T) {
	steps := defaultSteps(Toolchain("never-heard-of-it"), "/tmp")
	if steps != nil {
		t.Errorf("expected nil for unknown toolchain, got %v", steps)
	}
}

func TestSecuritySteps(t *testing.T) {
	dir := t.TempDir()
	steps := securitySteps(dir)
	if len(steps) < 2 {
		t.Fatalf("securitySteps() returned %d steps, want at least 2", len(steps))
	}
	if steps[0].Cmd != "semgrep" || steps[0].Name != "semgrep" {
		t.Fatalf("first security step = %+v, want semgrep", steps[0])
	}
	if steps[1].Cmd != "gitleaks" || steps[1].Name != "gitleaks" {
		t.Fatalf("second security step = %+v, want gitleaks", steps[1])
	}
	if !containsArgument(steps[0].Args, "--error") ||
		!containsArgument(steps[0].Args, "--metrics=off") ||
		!containsArgument(steps[0].Args, "p/security-audit") ||
		containsArgument(steps[0].Args, "auto") {
		t.Fatalf("semgrep args do not use the explicit metrics-off-compatible security ruleset: %v", steps[0].Args)
	}
	if !containsArgument(steps[1].Args, "--redact=100") {
		t.Fatalf("gitleaks args do not redact findings: %v", steps[1].Args)
	}
}

func TestDependencySecuritySteps(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte("# yarn lockfile v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	modDir := filepath.Join(dir, "svc")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module svc\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	steps := dependencySecuritySteps(dir)
	if len(steps) != 2 {
		t.Fatalf("dependencySecuritySteps returned %d steps, want 2: %+v", len(steps), steps)
	}
	if steps[0].Cmd != "yarn" || steps[0].WorkDir != dir {
		t.Fatalf("yarn step = %+v, want yarn in %s", steps[0], dir)
	}
	if steps[1].Cmd != "govulncheck" || steps[1].WorkDir != modDir {
		t.Fatalf("govulncheck step = %+v, want govulncheck in %s", steps[1], modDir)
	}
	if !containsArgument(steps[0].Args, "--severity") || !containsArgument(steps[0].Args, "high") {
		t.Fatalf("yarn audit args missing high severity gate: %v", steps[0].Args)
	}
}

func containsArgument(arguments []string, wanted string) bool {
	for _, argument := range arguments {
		if argument == wanted {
			return true
		}
	}
	return false
}

func TestNewRunIDIsUnique(t *testing.T) {
	a := NewRunID()
	b := NewRunID()
	if a == b {
		t.Errorf("NewRunID returned duplicates: %q", a)
	}
	if len(a) != 16 {
		t.Errorf("NewRunID length = %d, want 16", len(a))
	}
}
