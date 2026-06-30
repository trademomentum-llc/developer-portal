// scanner.go -- rule catalog and Scan() implementation.
//
// The catalog mirrors the six commit-discipline principles (Reproducibility,
// Provenance, Security, Signal-to-noise, Portability, Reversibility) plus the
// intent-enforcement rules in validator.go. Each rule has a stable code so
// audit-log queries remain trustworthy across versions.
package main

import (
	"path/filepath"
	"strings"
)

// Rules returns the full catalog. Exposed as a function (not a global var) so
// tests can build alternative catalogs without monkey-patching.
func Rules() []Rule {
	return []Rule{
		// ---- Never-list: Security ----
		{Code: "NV-S-001", Principle: "security", Severity: "block",
			Reason: ".env file (commit .env.example instead)",
			Match:  matchEnvFile},
		{Code: "NV-S-002", Principle: "security", Severity: "block",
			Reason: "private key material",
			Match:  matchPrivateKey},
		{Code: "NV-S-003", Principle: "security", Severity: "block",
			Reason: "credential file",
			Match:  matchCredentials},

		// ---- Never-list: Reproducibility ----
		{Code: "NV-R-001", Principle: "reproducibility", Severity: "block",
			Reason: "dependency directory",
			Match:  matchDependencyDir},
		{Code: "NV-R-002", Principle: "reproducibility", Severity: "block",
			Reason: "compiled output",
			Match:  matchCompiledOutput},

		// ---- Never-list: Provenance ----
		{Code: "NV-P-001", Principle: "provenance", Severity: "block",
			Reason: "build output directory",
			Match:  matchBuildDir},

		// ---- Never-list: Signal-to-noise ----
		{Code: "NV-N-001", Principle: "signal", Severity: "block",
			Reason: "editor scratch / log file",
			Match:  matchEditorScratch},
		{Code: "NV-N-002", Principle: "signal", Severity: "block",
			Reason: "OS-generated junk file",
			Match:  matchOSJunk},
		{Code: "NV-N-003", Principle: "signal", Severity: "block",
			Reason: "file exceeds 5 MiB (use Git LFS)",
			Match:  matchHugeFile},

		// ---- Never-list: Reversibility ----
		{Code: "NV-X-001", Principle: "reversibility", Severity: "block",
			Reason: "runtime data directory at repo root",
			Match:  matchRuntimeData},

		// ---- Grey-list: warn-only ----
		{Code: "GR-P-001", Principle: "provenance", Severity: "warn",
			Reason: "generated protobuf / gRPC stub",
			Match:  matchGeneratedProto},
		{Code: "GR-N-001", Principle: "signal", Severity: "warn",
			Reason: "file between 1 MiB and 5 MiB",
			Match:  matchLargeFile},
		{Code: "GR-Y-001", Principle: "portability", Severity: "warn",
			Reason: "machine-specific IDE config",
			Match:  matchIDEConfig},
	}
}

// Scan applies every rule to every path. Returns the list of findings in the
// order (paths x rules) they were evaluated.
func Scan(paths []string, sizes map[string]int64, catalog []Rule) []Finding {
	var out []Finding
	for _, p := range paths {
		size := sizes[p]
		for _, r := range catalog {
			if r.Match(p, size) {
				out = append(out, Finding{Rule: r, Path: p, Size: size})
			}
		}
	}
	return out
}

// SplitFindings partitions findings into blocking and warning groups.
func SplitFindings(f []Finding) (blocking, warning []Finding) {
	for _, x := range f {
		if x.Rule.Severity == "block" {
			blocking = append(blocking, x)
		} else {
			warning = append(warning, x)
		}
	}
	return
}

// ---- match helpers ------------------------------------------------------

// matchEnvFile blocks `.env`, `.env.local`, `.env.production`, etc., while
// allowing `.env.example` and `.env.sample` (the documented templates).
func matchEnvFile(p string, _ int64) bool {
	base := filepath.Base(p)
	if base == ".env.example" || base == ".env.sample" {
		return false
	}
	if base == ".env" {
		return true
	}
	return strings.HasPrefix(base, ".env.")
}

// matchPrivateKey blocks files whose name suggests private key material.
func matchPrivateKey(p string, _ int64) bool {
	base := filepath.Base(p)
	for _, ext := range []string{".pem", ".key", ".crt", ".cer", ".pfx", ".p12"} {
		if strings.HasSuffix(strings.ToLower(base), ext) {
			return true
		}
	}
	for _, prefix := range []string{"id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"} {
		if strings.HasPrefix(base, prefix) {
			return true
		}
	}
	return false
}

// matchCredentials catches common credential filenames: AWS credentials,
// gcloud, service-account JSON, htpasswd. This is a heuristic.
func matchCredentials(p string, _ int64) bool {
	lower := strings.ToLower(p)
	if strings.HasSuffix(lower, "/credentials") && strings.Contains(lower, ".aws/") {
		return true
	}
	base := filepath.Base(lower)
	switch base {
	case "credentials.json", "gcloud_credentials.json", ".htpasswd":
		return true
	}
	if strings.Contains(base, "service-account") && strings.HasSuffix(base, ".json") {
		return true
	}
	return false
}

// matchDependencyDir catches node_modules, venv, target/, __pycache__ anywhere
// in the path. vendor/ is intentionally NOT blocked here -- it's a Grey area
// handled separately (some Go projects vendor by policy).
func matchDependencyDir(p string, _ int64) bool {
	for _, part := range strings.Split(p, "/") {
		switch part {
		case "node_modules", "venv", ".venv", "__pycache__", "target":
			return true
		}
	}
	return false
}

// matchCompiledOutput catches compiled artefacts that should never be sourced.
func matchCompiledOutput(p string, _ int64) bool {
	base := strings.ToLower(filepath.Base(p))
	for _, ext := range []string{".pyc", ".class", ".o", ".obj", ".exe", ".dll", ".dylib", ".so"} {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}
	return false
}

// matchBuildDir catches typical build/output directories at any depth.
func matchBuildDir(p string, _ int64) bool {
	for _, part := range strings.Split(p, "/") {
		switch part {
		case "dist", "build", ".next", ".turbo", "out", "coverage":
			return true
		}
	}
	return false
}

// matchEditorScratch catches editor scratch files and log files.
func matchEditorScratch(p string, _ int64) bool {
	base := filepath.Base(p)
	for _, suf := range []string{".swp", ".swo", ".bak", ".orig", "~"} {
		if strings.HasSuffix(base, suf) {
			return true
		}
	}
	if strings.HasSuffix(base, ".log") {
		return true
	}
	if strings.HasPrefix(base, "npm-debug.log") || strings.HasPrefix(base, "yarn-debug.log") {
		return true
	}
	return false
}

// matchOSJunk catches the famous junk files OSes leave behind.
func matchOSJunk(p string, _ int64) bool {
	base := filepath.Base(p)
	switch base {
	case ".DS_Store", "Thumbs.db", "desktop.ini":
		return true
	}
	return false
}

// matchHugeFile blocks anything strictly larger than 5 MiB. LFS-tracked files
// are caught via the .gitattributes filter chain before reaching this rule,
// so a real LFS pointer (which is a tiny text file) lands as small here.
func matchHugeFile(_ string, size int64) bool {
	return size > 5*1024*1024
}

// matchLargeFile flags 1 MiB to 5 MiB inclusive as a Grey-list warning.
func matchLargeFile(_ string, size int64) bool {
	return size > 1*1024*1024 && size <= 5*1024*1024
}

// matchRuntimeData catches mutable runtime data dirs at the repo root.
func matchRuntimeData(p string, _ int64) bool {
	first := p
	if idx := strings.IndexByte(p, '/'); idx >= 0 {
		first = p[:idx]
	}
	switch first {
	case "data", "results", "uploads":
		return true
	}
	return false
}

// matchGeneratedProto flags committed protobuf / gRPC stubs as Grey because
// many teams accept them; the catch is they require regeneration discipline.
func matchGeneratedProto(p string, _ int64) bool {
	lower := strings.ToLower(p)
	for _, suf := range []string{".pb.go", "_pb2.py", "_pb2_grpc.py", ".pb.cc", ".pb.h"} {
		if strings.HasSuffix(lower, suf) {
			return true
		}
	}
	return false
}

// matchIDEConfig flags machine-specific IDE configs (workspace-level state).
// Project-level recommended-extensions configs are OUTSIDE this rule.
func matchIDEConfig(p string, _ int64) bool {
	switch p {
	case ".vscode/settings.json", ".idea/workspace.xml":
		return true
	}
	return false
}
