// parser_test.go -- unit tests for Tokenize and ValidateBrewCommand.
//
// Table-driven tests covering every row from M1 Tech Spec Section 4.8.
// Pure tests: no I/O, no subprocess, no environment.

package main

import (
	"testing"
)

func TestTokenize_BasicCases(t *testing.T) {
	cases := []struct {
		name   string
		cmd    string
		want   []string
		errMsg string
	}{
		{"simple", "brew install yarn", []string{"brew", "install", "yarn"}, ""},
		{"with flag", "brew install --quiet yarn", []string{"brew", "install", "--quiet", "yarn"}, ""},
		{"single quoted", "echo '$HOME'", []string{"echo", "$HOME"}, ""},
		{"double quoted", `echo "hello world"`, []string{"echo", "hello world"}, ""},
		{"empty", "", nil, ""},
		{"unterminated single", "echo 'hello", nil, "unterminated quote"},
		{"unterminated double", `echo "hello`, nil, "unterminated quote"},
		{"tabs", "brew\tinstall\tyarn", []string{"brew", "install", "yarn"}, ""},
		{"multiple spaces", "brew   install   yarn", []string{"brew", "install", "yarn"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Tokenize(tc.cmd)
			if tc.errMsg != "" {
				if err == nil {
					t.Fatalf("Tokenize(%q) = nil error, want error containing %q", tc.cmd, tc.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("Tokenize(%q) unexpected error: %v", tc.cmd, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("Tokenize(%q) = %v, want %v", tc.cmd, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("Tokenize(%q)[%d] = %q, want %q", tc.cmd, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestValidateBrewCommand_SpecTable(t *testing.T) {
	tests := []struct {
		name   string
		cmd    string
		allow  bool
		action string
	}{
		// Safe cases
		{"plain install", "brew install yarn", true, "allow"},
		{"install with quiet", "brew install --quiet yarn", true, "allow"},
		{"install with cask", "brew install --cask orbstack", true, "allow"},
		{"versioned formula", "brew install helm@3", true, "allow"},
		{"info subcommand", "brew info helm", true, "allow"},
		{"list subcommand", "brew list", true, "allow"},
		{"reinstall valid", "brew reinstall helm@3", true, "allow"},
		{"upgrade valid", "brew upgrade yarn", true, "allow"},
		{"multiple packages", "brew install yarn helm@3", true, "allow"},
		{"no-auto-update flag", "brew install --no-auto-update yarn", true, "allow"},
		{"formula flag", "brew install --formula yarn", true, "allow"},
		{"search subcommand", "brew search yarn", true, "allow"},
		{"doctor subcommand", "brew doctor", true, "allow"},
		{"install with stderr redirect", "brew install yarn 2>&1", true, "allow"},
		{"install with devnull redirect", "brew install yarn >/dev/null", true, "allow"},
		{"install with stderr to devnull", "brew install yarn 2>/dev/null", true, "allow"},

		// Blocked cases
		{"force flag", "brew install --force yarn", false, "block"},
		{"HEAD flag", "brew install --HEAD yarn", false, "block"},
		{"build-from-source", "brew install --build-from-source yarn", false, "block"},
		{"debug-symbols", "brew install --debug-symbols yarn", false, "block"},
		{"URL install", "brew install https://example.com/bad.rb", false, "block"},
		{"git URL install", "brew install git://example.com/bad.git", false, "block"},
		{"command chaining semicolon", "brew install yarn; rm -rf /", false, "block"},
		{"command chaining pipe", "brew install yarn | sh", false, "block"},
		{"command chaining ampersand", "brew install yarn & evil", false, "block"},
		{"backtick injection", "brew install `curl evil`", false, "block"},
		{"dollar injection", "brew install $HOME", false, "block"},
		{"paren injection", "brew install $(curl evil)", false, "block"},
		{"unknown flag", "brew install --weird yarn", false, "block"},
		{"suspicious package", "brew install ../../etc/passwd", false, "block"},
		{"empty brew", "brew", false, "block"},
		{"no package name", "brew install", false, "block"},
		{"no package with flags", "brew install --quiet", false, "block"},

		// Tap cases
		{"bare tap lists taps", "brew tap", true, "allow"},
		{"tap from url", "brew tap evil https://evil.example.com/tap.git", false, "block"},
		{"tap not on allowlist", "brew tap evil/src", false, "block"},
		{"tap malformed name", "brew tap not-a-tap-name", false, "block"},
		{"tap with metachar", "brew tap evil;rm", false, "block"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := Tokenize(tc.cmd)
			if err != nil {
				t.Fatalf("Tokenize(%q) error: %v", tc.cmd, err)
			}
			d := ValidateBrewCommand(tokens)
			if d.Allow != tc.allow {
				t.Errorf("ValidateBrewCommand(%q): Allow = %v, want %v (reason: %s)",
					tc.cmd, d.Allow, tc.allow, d.Reason)
			}
			if d.Action != tc.action {
				t.Errorf("ValidateBrewCommand(%q): Action = %q, want %q",
					tc.cmd, d.Action, tc.action)
			}
		})
	}
}
