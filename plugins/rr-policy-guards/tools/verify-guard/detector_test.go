// detector_test.go -- DetectToolchains over synthetic repo trees.

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeMarker(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectToolchains_Single(t *testing.T) {
	cases := []struct {
		marker string
		body   string
		want   Toolchain
	}{
		{"go.mod", "module x\n", ToolchainGo},
		{"package.json", "{}\n", ToolchainNode},
		{"Cargo.toml", "[package]\n", ToolchainRust},
		{"pyproject.toml", "[project]\n", ToolchainPython},
		{"requirements.txt", "", ToolchainPython},
		{"setup.py", "", ToolchainPython},
		{"CMakeLists.txt", "", ToolchainCpp},
		{"mix.exs", "", ToolchainElixir},
		{"Gemfile", "", ToolchainRuby},
		{"pom.xml", "<project/>", ToolchainJVMMaven},
		{"build.gradle", "", ToolchainJVMGradle},
		{"build.gradle.kts", "", ToolchainJVMGradle},
	}
	for _, c := range cases {
		t.Run(c.marker, func(t *testing.T) {
			dir := t.TempDir()
			writeMarker(t, dir, c.marker, c.body)
			got, err := DetectToolchains(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[0] != c.want {
				t.Errorf("detected %v, want [%v]", got, c.want)
			}
		})
	}
}

func TestDetectToolchains_Polyglot(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "go.mod", "module x\n")
	writeMarker(t, dir, "tools/web/package.json", "{}\n")
	writeMarker(t, dir, "vendor/should-be-skipped/Cargo.toml", "[package]\n")
	writeMarker(t, dir, "node_modules/foo/Cargo.toml", "[package]\n")
	got, err := DetectToolchains(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []Toolchain{ToolchainGo, ToolchainNode}
	sort.Slice(got, func(i, j int) bool { return string(got[i]) < string(got[j]) })
	sort.Slice(want, func(i, j int) bool { return string(want[i]) < string(want[j]) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("detected %v, want %v (vendor + node_modules should be skipped)", got, want)
	}
}

func TestDetectToolchains_Make(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "Makefile", "test:\n\techo ok\n")
	got, err := DetectToolchains(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != ToolchainMake {
		t.Errorf("expected [make], got %v", got)
	}
}

func TestDetectToolchains_MakefileNoTestTarget(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "Makefile", "build:\n\techo ok\n")
	got, _ := DetectToolchains(dir)
	for _, tc := range got {
		if tc == ToolchainMake {
			t.Errorf("Makefile without test: target should NOT trigger make toolchain; got %v", got)
		}
	}
}

func TestDetectToolchains_DepthCap(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join("a", "b", "c", "d", "e", "f", "g")
	writeMarker(t, dir, filepath.Join(deep, "go.mod"), "module deep\n")
	got, _ := DetectToolchains(dir)
	for _, tc := range got {
		if tc == ToolchainGo {
			t.Errorf("walk should stop at depth 4; deep go.mod should not be detected; got %v", got)
		}
	}
}

func TestDetectToolchains_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	got, err := DetectToolchains(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("empty repo should yield no toolchains, got %v", got)
	}
}
