// act_test.go -- ActAvailable / DockerAvailable detect predicates.
//
// We do not actually invoke act here; the function under test only
// checks PATH (or asks docker for status). These tests confirm the
// detection functions are wired correctly without depending on act
// or docker being installed.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActAvailable_NoBinary(t *testing.T) {
	// Empty PATH -- act is definitely not on it.
	t.Setenv("PATH", "")
	if ActAvailable() {
		t.Error("expected ActAvailable=false when PATH is empty")
	}
}

func TestActAvailable_Stub(t *testing.T) {
	// Drop a fake `act` shim into a temp dir, point PATH there.
	dir := t.TempDir()
	stub := filepath.Join(dir, "act")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if !ActAvailable() {
		t.Error("expected ActAvailable=true with stub on PATH")
	}
}

func TestDockerAvailable_NoBinary(t *testing.T) {
	t.Setenv("PATH", "")
	if DockerAvailable() {
		t.Error("expected DockerAvailable=false when PATH is empty")
	}
}
