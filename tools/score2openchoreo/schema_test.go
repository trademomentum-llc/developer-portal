package main

import (
	"os"
	"path/filepath"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("fixtures", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return b
}

func TestValidateScoreAcceptsMinimal(t *testing.T) {
	raw := readFixture(t, "minimal.score.yaml")
	if err := ValidateScore(raw); err != nil {
		t.Fatalf("valid fixture failed: %v", err)
	}
}

func TestValidateScoreRejectsInvalid(t *testing.T) {
	raw := readFixture(t, "invalid-schema.score.yaml")
	if err := ValidateScore(raw); err == nil {
		t.Fatal("invalid fixture unexpectedly passed")
	}
}
