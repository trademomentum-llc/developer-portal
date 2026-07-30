package main

import (
	"reflect"
	"testing"
)

func TestTokenizeShell(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "two", in: "git commit", want: []string{"git", "commit"}},
		{name: "single quoted", in: "git commit -m 'fix: issue'", want: []string{"git", "commit", "-m", "fix: issue"}},
		{name: "double quoted", in: `git commit -m "fix: issue"`, want: []string{"git", "commit", "-m", "fix: issue"}},
		{name: "escaped space", in: `cmd a\ b`, want: []string{"cmd", "a b"}},
		{name: "tabs", in: "git\tpush", want: []string{"git", "push"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := tokenizeShell(test.in); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("tokenizeShell(%q) = %v, want %v", test.in, got, test.want)
			}
		})
	}
}

func TestIsShellAssignment(t *testing.T) {
	tests := map[string]bool{
		"NAME=value": true,
		"_NAME=":     true,
		"NAME_2=x":   true,
		"2NAME=x":    false,
		"--flag=x":   false,
		"NAME":       false,
	}
	for token, want := range tests {
		if got := isShellAssignment(token); got != want {
			t.Errorf("isShellAssignment(%q) = %t, want %t", token, got, want)
		}
	}
}
