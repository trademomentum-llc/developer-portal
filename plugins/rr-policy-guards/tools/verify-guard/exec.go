// exec.go -- per-toolchain command sequences and their execution.
//
// Each toolchain has a fixed default command sequence (overridable
// per-repo via .rr-verify-guard.json). Steps run sequentially within a
// toolchain; first non-zero exit terminates that toolchain and is
// recorded as the failure reason.

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	osexec "os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// defaultSteps returns the default Step sequence for a toolchain at a
// given repoRoot.
func defaultSteps(t Toolchain, repoRoot string) []Step {
	switch t {
	case ToolchainGo:
		return []Step{
			{Toolchain: t, Name: "vet", Cmd: "go", Args: []string{"vet", "./..."}, Required: true},
			{Toolchain: t, Name: "test", Cmd: "go", Args: []string{"test", "-race", "-count=1", "./..."}, Required: true},
		}
	case ToolchainNode:
		var steps []Step
		if fileExists(filepath.Join(repoRoot, "tsconfig.json")) {
			steps = append(steps, Step{Toolchain: t, Name: "tsc-noEmit", Cmd: "npx", Args: []string{"--no-install", "tsc", "--noEmit"}, Required: true})
		}
		if hasNpmScript(repoRoot, "test") {
			steps = append(steps, Step{Toolchain: t, Name: "npm-test", Cmd: "npm", Args: []string{"test", "--silent"}, Required: true})
		}
		return steps
	case ToolchainRust:
		return []Step{
			{Toolchain: t, Name: "fmt-check", Cmd: "cargo", Args: []string{"fmt", "--", "--check"}, Required: true},
			{Toolchain: t, Name: "clippy", Cmd: "cargo", Args: []string{"clippy", "--all-targets", "--", "-D", "warnings"}, Required: true},
			{Toolchain: t, Name: "test", Cmd: "cargo", Args: []string{"test", "--quiet"}, Required: true},
		}
	case ToolchainPython:
		var steps []Step
		if commandOnPath("ruff") {
			steps = append(steps, Step{Toolchain: t, Name: "ruff", Cmd: "ruff", Args: []string{"check", "."}, Required: true})
		}
		if commandOnPath("mypy") {
			steps = append(steps, Step{Toolchain: t, Name: "mypy", Cmd: "mypy", Args: []string{"."}, Required: true})
		}
		if commandOnPath("pytest") {
			steps = append(steps, Step{Toolchain: t, Name: "pytest", Cmd: "pytest", Args: []string{"-x", "--quiet"}, Required: true})
		}
		return steps
	case ToolchainCpp:
		buildDir := filepath.Join(repoRoot, "build")
		return []Step{
			{Toolchain: t, Name: "cmake-configure", Cmd: "cmake", Args: []string{"-S", ".", "-B", "build", "-DCMAKE_BUILD_TYPE=Debug"}, Required: true},
			{Toolchain: t, Name: "cmake-build", Cmd: "cmake", Args: []string{"--build", buildDir, "--parallel"}, Required: true},
			{Toolchain: t, Name: "ctest", Cmd: "ctest", Args: []string{"--test-dir", buildDir, "--output-on-failure"}, Required: true},
		}
	case ToolchainMake:
		return []Step{
			{Toolchain: t, Name: "make-test", Cmd: "make", Args: []string{"test"}, Required: true},
		}
	case ToolchainElixir:
		return []Step{
			{Toolchain: t, Name: "mix-test", Cmd: "mix", Args: []string{"test", "--warnings-as-errors"}, Required: true},
		}
	case ToolchainRuby:
		if dirExists(filepath.Join(repoRoot, "spec")) {
			return []Step{{Toolchain: t, Name: "rspec", Cmd: "bundle", Args: []string{"exec", "rspec"}, Required: true}}
		}
		return []Step{{Toolchain: t, Name: "rake-test", Cmd: "bundle", Args: []string{"exec", "rake", "test"}, Required: true}}
	case ToolchainJVMMaven:
		return []Step{{Toolchain: t, Name: "mvn-test", Cmd: "mvn", Args: []string{"-B", "-q", "test"}, Required: true}}
	case ToolchainJVMGradle:
		return []Step{{Toolchain: t, Name: "gradle-test", Cmd: "./gradlew", Args: []string{"test", "--console=plain"}, Required: true}}
	}
	return nil
}

func securitySteps(repoRoot string) []Step {
	steps := []Step{
		{
			Toolchain: ToolchainSecurity,
			Name:      "semgrep",
			Cmd:       "semgrep",
			Args: []string{
				"scan",
				"--config", "p/security-audit",
				"--error",
				"--metrics=off",
				"--disable-version-check",
				".",
			},
			Required: true,
		},
		{
			Toolchain: ToolchainSecurity,
			Name:      "gitleaks",
			Cmd:       "gitleaks",
			Args: []string{
				"dir",
				"--no-banner",
				"--no-color",
				"--redact=100",
				"--exit-code", "1",
				".",
			},
			Required: true,
		},
	}
	steps = append(steps, dependencySecuritySteps(repoRoot)...)
	return steps
}

// dependencySecuritySteps discovers Node and Go module roots and returns
// SCA steps. High+Critical npm advisories and any govulncheck finding block.
func dependencySecuritySteps(repoRoot string) []Step {
	var steps []Step
	for _, dir := range findMarkerDirs(repoRoot, "yarn.lock") {
		// Prefer Yarn Berry audit when a yarn.lock is present.
		steps = append(steps, Step{
			Toolchain: ToolchainSecurity,
			Name:      "yarn-npm-audit",
			Cmd:       "yarn",
			Args: []string{
				"npm", "audit",
				"--all",
				"--recursive",
				"--severity", "high",
				"--no-deprecations",
			},
			Required: true,
			WorkDir:  dir,
		})
	}
	for _, dir := range findMarkerDirs(repoRoot, "package-lock.json") {
		// npm lockfile roots without yarn.lock.
		if fileExists(filepath.Join(dir, "yarn.lock")) {
			continue
		}
		steps = append(steps, Step{
			Toolchain: ToolchainSecurity,
			Name:      "npm-audit",
			Cmd:       "npm",
			Args: []string{
				"audit",
				"--audit-level=high",
			},
			Required: true,
			WorkDir:  dir,
		})
	}
	for _, dir := range findMarkerDirs(repoRoot, "go.mod") {
		steps = append(steps, Step{
			Toolchain: ToolchainSecurity,
			Name:      "govulncheck",
			Cmd:       "govulncheck",
			Args:      []string{"./..."},
			Required:  true,
			WorkDir:   dir,
		})
	}
	return steps
}

// findMarkerDirs returns absolute directories under repoRoot (depth-capped)
// that contain the given marker file.
func findMarkerDirs(repoRoot, marker string) []string {
	var dirs []string
	_ = filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path == repoRoot {
				return nil
			}
			if _, skip := skipDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(repoRoot, path)
			depth := strings.Count(rel, string(filepath.Separator)) + 1
			if depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != marker {
			return nil
		}
		dirs = append(dirs, filepath.Dir(path))
		return nil
	})
	sort.Strings(dirs)
	return dirs
}

func RunSecurityScans(repoRoot, runID string) ([]Result, *Result) {
	var results []Result
	for _, step := range securitySteps(repoRoot) {
		if !commandOnPath(step.Cmd) {
			result := Result{
				Toolchain: step.Toolchain,
				StepName:  step.Name + "-binary-missing",
				Cmd:       step.Cmd,
				Args:      step.Args,
				ExitCode:  127,
				Truncated: fmt.Sprintf("required security scanner %q not found on PATH", step.Cmd),
				Err:       fmt.Errorf("security scanner %q missing", step.Cmd),
			}
			results = append(results, result)
			return results, &result
		}
		result := runStep(repoRoot, runID, step)
		results = append(results, result)
		if result.ExitCode != 0 {
			return results, &result
		}
	}
	return results, nil
}

// RunToolchains runs verification for every toolchain. Returns a slice
// of all step results, the first failure (or nil), and a list of
// degraded reasons (e.g. missing toolchain binaries).
//
// Security model: command sequences come exclusively from defaultSteps
// (a function returning hardcoded constants). The Config.Toolchains
// map can ONLY enable or disable a toolchain via tcfg.Skip; it cannot
// inject commands. ExtraCommands is intentionally not honoured in
// v0.1.0 -- arbitrary commands from on-disk config would violate the
// "no untrusted command execution" property guarded by runStep's
// allowlist.
func RunToolchains(repoRoot string, toolchains []Toolchain, runID string, cfg Config) ([]Result, *Result, []string) {
	var allResults []Result
	var degraded []string
	for _, t := range toolchains {
		// Allow per-repo skip.
		if tcfg, ok := cfg.Toolchains[t]; ok && tcfg.Skip {
			degraded = append(degraded, "config-skip / "+string(t))
			continue
		}
		// Allow env-var skip.
		if os.Getenv("RR_VERIFY_GUARD_SKIP_"+strings.ToUpper(string(t))) == "1" {
			degraded = append(degraded, "env-skip / "+string(t))
			continue
		}
		steps := defaultSteps(t, repoRoot)
		if len(steps) == 0 {
			degraded = append(degraded, "no-steps / "+string(t))
			continue
		}
		// Verify the toolchain binary is available (first step's Cmd).
		if !commandOnPath(steps[0].Cmd) {
			degraded = append(degraded, "missing-binary / "+steps[0].Cmd)
			r := Result{
				Toolchain: t,
				StepName:  "binary-missing",
				Cmd:       steps[0].Cmd,
				ExitCode:  127,
				Truncated: fmt.Sprintf("required binary %q not found on PATH", steps[0].Cmd),
				Err:       fmt.Errorf("binary %q missing", steps[0].Cmd),
			}
			allResults = append(allResults, r)
			return allResults, &r, degraded
		}
		for _, s := range steps {
			res := runStep(repoRoot, runID, s)
			allResults = append(allResults, res)
			if res.ExitCode != 0 {
				return allResults, &res, degraded
			}
		}
	}
	return allResults, nil, degraded
}

// allowedBinaries enumerates every binary that runStep is permitted to
// execute. The set is hard-coded; .rr-verify-guard.json cannot extend
// it. Any Step whose Cmd is not in this set is rejected with exit 126.
var allowedBinaries = map[string]struct{}{
	"go":        {},
	"npm":       {},
	"npx":       {},
	"cargo":     {},
	"ruff":      {},
	"mypy":      {},
	"pytest":    {},
	"cmake":     {},
	"ctest":     {},
	"make":      {},
	"mix":       {},
	"bundle":    {},
	"mvn":       {},
	"./gradlew": {},
	"act":         {},
	"semgrep":     {},
	"gitleaks":    {},
	"yarn":        {},
	"govulncheck": {},
}

// dispatchExec is the ONLY entrypoint for command execution in this
// guard. Each case below uses a string literal as the binary name so a
// static analyser can verify that no caller-controlled string can reach
// exec.CommandContext.
func dispatchExec(ctx context.Context, s Step) (*osexec.Cmd, error) {
	if _, ok := allowedBinaries[s.Cmd]; !ok {
		return nil, fmt.Errorf("binary %q is not in the verify-guard allowlist", s.Cmd)
	}
	switch s.Cmd {
	case "go":
		return osexec.CommandContext(ctx, "go", s.Args...), nil
	case "npm":
		return osexec.CommandContext(ctx, "npm", s.Args...), nil
	case "npx":
		return osexec.CommandContext(ctx, "npx", s.Args...), nil
	case "cargo":
		return osexec.CommandContext(ctx, "cargo", s.Args...), nil
	case "ruff":
		return osexec.CommandContext(ctx, "ruff", s.Args...), nil
	case "mypy":
		return osexec.CommandContext(ctx, "mypy", s.Args...), nil
	case "pytest":
		return osexec.CommandContext(ctx, "pytest", s.Args...), nil
	case "cmake":
		return osexec.CommandContext(ctx, "cmake", s.Args...), nil
	case "ctest":
		return osexec.CommandContext(ctx, "ctest", s.Args...), nil
	case "make":
		return osexec.CommandContext(ctx, "make", s.Args...), nil
	case "mix":
		return osexec.CommandContext(ctx, "mix", s.Args...), nil
	case "bundle":
		return osexec.CommandContext(ctx, "bundle", s.Args...), nil
	case "mvn":
		return osexec.CommandContext(ctx, "mvn", s.Args...), nil
	case "./gradlew":
		return osexec.CommandContext(ctx, "./gradlew", s.Args...), nil
	case "act":
		return osexec.CommandContext(ctx, "act", s.Args...), nil
	case "semgrep":
		return osexec.CommandContext(ctx, "semgrep", s.Args...), nil
	case "gitleaks":
		return osexec.CommandContext(ctx, "gitleaks", s.Args...), nil
	case "yarn":
		return osexec.CommandContext(ctx, "yarn", s.Args...), nil
	case "govulncheck":
		return osexec.CommandContext(ctx, "govulncheck", s.Args...), nil
	}
	return nil, fmt.Errorf("unreachable: %q passed allowlist but missed switch", s.Cmd)
}

// runStep executes one Step via dispatchExec, captures stdout+stderr
// to a log file, and returns a Result with the first 50 lines of
// output for deny payloads. Any Step that names a non-allowlisted
// binary is rejected with exit 126 and never reaches the OS.
func runStep(repoRoot, runID string, s Step) Result {
	timeoutSec := 600
	if v := os.Getenv("RR_VERIFY_GUARD_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSec = n
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	cmd, dispatchErr := dispatchExec(ctx, s)
	if dispatchErr != nil {
		return Result{
			Toolchain: s.Toolchain,
			StepName:  s.Name,
			Cmd:       s.Cmd,
			Args:      s.Args,
			ExitCode:  126,
			Truncated: dispatchErr.Error(),
			Err:       dispatchErr,
		}
	}
	if s.WorkDir != "" {
		cmd.Dir = s.WorkDir
	} else {
		cmd.Dir = repoRoot
	}
	var buf bytes.Buffer
	logPath := stepLogPath(runID, s)
	if logFile, err := openLogFile(logPath); err == nil {
		defer logFile.Close()
		cmd.Stdout = io.MultiWriter(&buf, logFile)
		cmd.Stderr = io.MultiWriter(&buf, logFile)
	} else {
		cmd.Stdout = &buf
		cmd.Stderr = &buf
	}
	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start).Milliseconds()
	exitCode := 0
	if err != nil {
		var ee *osexec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			exitCode = 124
		} else {
			exitCode = 1
		}
	}
	captured := buf.String()
	const cap50KiB = 64 * 1024
	if len(captured) > cap50KiB {
		captured = captured[:cap50KiB] + "\n... (truncated; see " + logPath + ")"
	}
	return Result{
		Toolchain:  s.Toolchain,
		StepName:   s.Name,
		Cmd:        s.Cmd,
		Args:       s.Args,
		ExitCode:   exitCode,
		DurationMS: dur,
		LogPath:    logPath,
		Truncated:  firstLines(captured, 50),
		Err:        err,
	}
}

// stepLogPath returns the path inside the run-log directory.
func stepLogPath(runID string, s Step) string {
	root, err := cacheRoot()
	if err != nil {
		return ""
	}
	return filepath.Join(root, "runs", runID, fmt.Sprintf("%s-%s.log", s.Toolchain, s.Name))
}

// openLogFile creates the log file with parent dirs.
func openLogFile(path string) (*os.File, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
}

// NewRunID returns a fresh hex run identifier.
func NewRunID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// hasNpmScript returns true if package.json defines a script with name.
func hasNpmScript(repoRoot, name string) bool {
	b, err := os.ReadFile(filepath.Join(repoRoot, "package.json"))
	if err != nil {
		return false
	}
	// Cheap substring check; adequate for v0.1 -- a stricter parse can
	// land in v0.2 if false-positives bite.
	needle := "\"" + name + "\""
	idx := bytes.Index(b, []byte("\"scripts\""))
	if idx < 0 {
		return false
	}
	return bytes.Contains(b[idx:], []byte(needle))
}
