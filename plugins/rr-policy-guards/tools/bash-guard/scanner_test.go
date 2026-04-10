// scanner_test.go -- unit tests for ScanCommand and ApplyFixes.
//
// Pure tests: no I/O, no subprocess, no environment.

package main

import (
	"testing"
)

func TestScanCommand_CleanCommands_ReturnsNil(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
	}{
		{"simple ls", "ls -la /tmp"},
		{"absolute path", "/usr/bin/env python3 /tmp/script.py"},
		{"pipe", "cat /tmp/file.txt | head -5"},
		{"redirect", "echo hello > /tmp/out.txt"},
		{"single quoted dollar", "echo '$HOME is not expanded'"},
		{"escaped dollar", "echo \\$HOME"},
		{"command substitution", "echo $(date)"},
		{"exit status", "echo $?"},
		{"pid", "echo $$"},
		{"last bg pid", "echo $!"},
		{"positional", "echo $1 $2 $9"},
		{"special vars", "echo $# $@ $* $-"},
		{"no command", ""},
		{"git status", "git status"},
		{"chained safe", "git status && git diff"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			violations := ScanCommand(tc.cmd)
			if len(violations) != 0 {
				t.Errorf("ScanCommand(%q) returned %d violations, want 0", tc.cmd, len(violations))
				for _, v := range violations {
					t.Logf("  [%s] %s at offset %d", v.Rule, v.Description, v.Offset)
				}
			}
		})
	}
}

func TestScanCommand_BareVar_Detects(t *testing.T) {
	cases := []struct {
		name    string
		cmd     string
		varName string
	}{
		{"HOME", "ls $HOME/projects", "HOME"},
		{"PATH", "echo $PATH", "PATH"},
		{"USER", "whoami && echo $USER", "USER"},
		{"underscore var", "echo $MY_VAR", "MY_VAR"},
		{"mid command", "cd $HOME && ls", "HOME"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			violations := ScanCommand(tc.cmd)
			if len(violations) == 0 {
				t.Fatalf("ScanCommand(%q) returned no violations, want at least 1", tc.cmd)
			}
			if violations[0].Rule != "SIMPLE_EXPANSION" {
				t.Errorf("rule = %q, want SIMPLE_EXPANSION", violations[0].Rule)
			}
		})
	}
}

func TestScanCommand_BraceExpansion_Detects(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
	}{
		{"brace HOME", "ls ${HOME}/projects"},
		{"brace with default", "echo ${EDITOR:-vim}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			violations := ScanCommand(tc.cmd)
			if len(violations) == 0 {
				t.Fatalf("ScanCommand(%q) returned no violations, want at least 1", tc.cmd)
			}
			if violations[0].Rule != "SIMPLE_EXPANSION" {
				t.Errorf("rule = %q, want SIMPLE_EXPANSION", violations[0].Rule)
			}
		})
	}
}

func TestScanCommand_SingleQuotesProtect(t *testing.T) {
	cmd := "echo '$HOME is literal'"
	violations := ScanCommand(cmd)
	if len(violations) != 0 {
		t.Errorf("ScanCommand(%q) should allow $VAR inside single quotes", cmd)
	}
}

func TestScanCommand_EscapedDollarSafe(t *testing.T) {
	cmd := "echo \\$HOME"
	violations := ScanCommand(cmd)
	if len(violations) != 0 {
		t.Errorf("ScanCommand(%q) should allow escaped $VAR", cmd)
	}
}

func TestScanCommand_MultipleViolations(t *testing.T) {
	cmd := "cp $HOME/file $DEST/target"
	violations := ScanCommand(cmd)
	if len(violations) != 2 {
		t.Fatalf("ScanCommand(%q) returned %d violations, want 2", cmd, len(violations))
	}
	if violations[0].Original != "$HOME" {
		t.Errorf("first violation original = %q, want $HOME", violations[0].Original)
	}
	if violations[1].Original != "$DEST" {
		t.Errorf("second violation original = %q, want $DEST", violations[1].Original)
	}
}

func TestApplyFixes_SingleVar(t *testing.T) {
	cmd := "ls $HOME/projects"
	violations := ScanCommand(cmd)
	corrected := ApplyFixes(cmd, violations)
	want := "ls <HOME_VALUE>/projects"
	if corrected != want {
		t.Errorf("ApplyFixes = %q, want %q", corrected, want)
	}
}

func TestApplyFixes_MultipleVars(t *testing.T) {
	cmd := "cp $HOME/file $DEST/target"
	violations := ScanCommand(cmd)
	corrected := ApplyFixes(cmd, violations)
	want := "cp <HOME_VALUE>/file <DEST_VALUE>/target"
	if corrected != want {
		t.Errorf("ApplyFixes = %q, want %q", corrected, want)
	}
}

func TestApplyFixes_BraceExpansion(t *testing.T) {
	cmd := "ls ${HOME}/projects"
	violations := ScanCommand(cmd)
	corrected := ApplyFixes(cmd, violations)
	want := "ls <HOME_VALUE>/projects"
	if corrected != want {
		t.Errorf("ApplyFixes = %q, want %q", corrected, want)
	}
}

func TestApplyFixes_NoViolations(t *testing.T) {
	cmd := "ls /tmp"
	corrected := ApplyFixes(cmd, nil)
	if corrected != cmd {
		t.Errorf("ApplyFixes with no violations changed command: %q -> %q", cmd, corrected)
	}
}

func TestApplyFixes_PreservesContext(t *testing.T) {
	cmd := "echo hello && cd $HOME && ls -la"
	violations := ScanCommand(cmd)
	corrected := ApplyFixes(cmd, violations)
	want := "echo hello && cd <HOME_VALUE> && ls -la"
	if corrected != want {
		t.Errorf("ApplyFixes = %q, want %q", corrected, want)
	}
}
