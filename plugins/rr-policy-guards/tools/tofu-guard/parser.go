// Package main: rr-tofu-guard parser. Pure functions, no I/O.
package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Decision struct {
	Allow  bool
	Reason string
	Action string // "allow" | "block" | "not-applicable"
}

var BlockedSubcommands = map[string]struct{}{
	"apply":   {},
	"destroy": {},
	"import":  {},
}

var BlockedStateSubcommands = map[string]struct{}{
	"rm":   {},
	"mv":   {},
	"push": {},
}

var shellMeta = regexp.MustCompile("[;&|<>`$()]")

// IsTofuCommandPrefix reports whether cmd, after leading whitespace, begins
// with the word "tofu" as a standalone token (followed by whitespace or
// end-of-string). It exists so main.go can short-circuit on commands that
// clearly do not invoke tofu, before running Tokenize: the tokenizer does
// not model heredocs or other shell constructs, and an unbalanced quote
// inside a heredoc must not trigger a block on a non-tofu command.
func IsTofuCommandPrefix(cmd string) bool {
	trimmed := strings.TrimLeft(cmd, " \t\n\r")
	if trimmed == "tofu" {
		return true
	}
	if !strings.HasPrefix(trimmed, "tofu") {
		return false
	}
	next := trimmed[4]
	return next == ' ' || next == '\t' || next == '\n' || next == '\r'
}

// Tokenize splits a command into tokens honoring single and double quotes.
// Backslash escapes inside double quotes. Returns error on unterminated quote.
func Tokenize(cmd string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble, escape := false, false, false
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		switch {
		case escape:
			cur.WriteByte(c)
			escape = false
		case c == '\\' && inDouble:
			escape = true
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case (c == ' ' || c == '\t') && !inSingle && !inDouble:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote")
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens, nil
}

// ValidateTofuCommand is pure: no I/O, no side effects.
func ValidateTofuCommand(tokens []string) Decision {
	if len(tokens) == 0 {
		return Decision{Allow: true, Action: "not-applicable"}
	}
	if tokens[0] != "tofu" {
		return Decision{Allow: true, Action: "not-applicable"}
	}
	if len(tokens) < 2 {
		return Decision{Allow: true, Action: "allow"}
	}

	sub := tokens[1]

	if sub == "state" && len(tokens) >= 3 {
		if _, bad := BlockedStateSubcommands[tokens[2]]; bad {
			return Decision{Allow: false, Reason: "state-mutating: tofu state " + tokens[2], Action: "block"}
		}
		return Decision{Allow: true, Action: "allow"}
	}

	if _, bad := BlockedSubcommands[sub]; bad {
		return Decision{
			Allow:  false,
			Reason: "tofu " + sub + " must run in CI, not from a Bash tool use",
			Action: "block",
		}
	}
	return Decision{Allow: true, Action: "allow"}
}
