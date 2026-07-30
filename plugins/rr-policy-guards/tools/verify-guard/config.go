// config.go -- per-repo config loaded from .rr-verify-guard.json.
//
// JSON-with-comments is not parsed; the file is plain JSON. The Design
// Spec discusses YAML as a future enhancement; v0.1.0 ships JSON to
// keep us stdlib-only.

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config is the on-disk shape of .rr-verify-guard.json.
//
// Per-repo config can ONLY enable or disable existing checks. It cannot
// inject arbitrary commands. v0.1.0 deliberately does not expose a
// command-override surface to keep the static-allowlist invariant in
// exec.go intact (a malicious repo cannot push commands into the
// guard via .rr-verify-guard.json).
type Config struct {
	SchemaVersion int                           `json:"schema_version"`
	Toolchains    map[Toolchain]ToolchainConfig `json:"toolchains"`
	Act           ActConfig                     `json:"act"`
}

// ToolchainConfig describes per-toolchain on/off control. The
// `Commands` and `ExtraCommands` fields that older drafts of the
// design described are intentionally absent here -- see Config doc.
type ToolchainConfig struct {
	Skip bool `json:"skip"`
}

// ActConfig configures whether `act` runs at all for this repo.
type ActConfig struct {
	Enabled *bool `json:"enabled"`
}

// LoadConfig reads .rr-verify-guard.json from repoRoot.
// Returns (zero, false, nil) when the file does not exist.
func LoadConfig(repoRoot string) (Config, bool, error) {
	path := filepath.Join(repoRoot, ".rr-verify-guard.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, false, err
	}
	if len(b) == 0 {
		return Config{}, true, errors.New("empty config file")
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, true, err
	}
	return c, true, nil
}
