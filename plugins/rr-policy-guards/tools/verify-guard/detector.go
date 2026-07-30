// detector.go -- toolchain detection by walking the repo for marker
// files. Stdlib only.

package main

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// markerToToolchain maps a marker filename to the toolchain it implies.
var markerToToolchain = map[string]Toolchain{
	"go.mod":           ToolchainGo,
	"package.json":     ToolchainNode,
	"Cargo.toml":       ToolchainRust,
	"pyproject.toml":   ToolchainPython,
	"requirements.txt": ToolchainPython,
	"setup.py":         ToolchainPython,
	"CMakeLists.txt":   ToolchainCpp,
	"mix.exs":          ToolchainElixir,
	"Gemfile":          ToolchainRuby,
	"pom.xml":          ToolchainJVMMaven,
	"build.gradle":     ToolchainJVMGradle,
	"build.gradle.kts": ToolchainJVMGradle,
}

// skipDirs are directory names we never descend into during detection.
var skipDirs = map[string]struct{}{
	"node_modules": {},
	"vendor":       {},
	"target":       {},
	"dist":         {},
	"build":        {},
	".git":         {},
	".idea":        {},
	".vscode":      {},
}

// maxDepth caps how deep the walk goes from the repo root.
const maxDepth = 4

// DetectToolchains walks repoRoot to maxDepth looking for marker files
// and returns the deduplicated, sorted set of toolchains found.
//
// A `Makefile` containing a `test:` target is recognised as ToolchainMake
// in addition to any language-specific toolchain present.
func DetectToolchains(repoRoot string) ([]Toolchain, error) {
	found := map[Toolchain]struct{}{}
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
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
		name := d.Name()
		if t, ok := markerToToolchain[name]; ok {
			found[t] = struct{}{}
		}
		if name == "Makefile" {
			if hasMakeTestTarget(path) {
				found[ToolchainMake] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]Toolchain, 0, len(found))
	for t := range found {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i]) < string(out[j]) })
	return out, nil
}

// hasMakeTestTarget returns true if the Makefile at path declares a
// target named `test:` or `test ::`. Cheap line scan; no full Makefile
// parsing.
func hasMakeTestTarget(path string) bool {
	b, err := readFileSafely(path, 1<<20)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "test:") || strings.HasPrefix(trim, "test::") {
			return true
		}
	}
	return false
}
