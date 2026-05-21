package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "score2openchoreo")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}
	return bin
}

func runBinary(t *testing.T, bin string, args ...string) ([]byte, int) {
	t.Helper()
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd := exec.Command(bin, args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	}
	if code != 0 {
		t.Logf("stderr: %s", errb.String())
	}
	return out.Bytes(), code
}

func TestGoldenMinimal(t *testing.T) {
	bin := buildBinary(t)
	got, code := runBinary(t, bin,
		"--input", "fixtures/minimal.score.yaml",
		"--environment", "dev",
	)
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	want, err := os.ReadFile("fixtures/minimal.component.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatalf("golden mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGoldenWithSecret(t *testing.T) {
	bin := buildBinary(t)
	got, code := runBinary(t, bin,
		"--input", "fixtures/with-secret.score.yaml",
		"--environment", "dev",
	)
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	want, err := os.ReadFile("fixtures/with-secret.component.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatalf("golden mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestValidateOnlyExitCode(t *testing.T) {
	bin := buildBinary(t)
	_, code := runBinary(t, bin, "--input", "fixtures/minimal.score.yaml", "--validate-only")
	if code != 0 {
		t.Fatalf("valid fixture exit=%d want 0", code)
	}
	_, code = runBinary(t, bin, "--input", "fixtures/invalid-schema.score.yaml", "--validate-only")
	if code != 1 {
		t.Fatalf("invalid fixture exit=%d want 1", code)
	}
}
