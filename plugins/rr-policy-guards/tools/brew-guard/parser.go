// parser.go -- brew command tokenizer and security validator.
//
// Pure functions only: no I/O, no stdin reading, no exit codes.
// main.go wires these functions up to the Claude Code PreToolUse protocol.
// Everything here is unit-testable in isolation.

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Decision is the outcome of validating a brew command.
type Decision struct {
	Allow  bool
	Reason string // empty when Allow=true; human-readable cause when Allow=false
	Action string // one of: "allow", "block", "bypass", "not-applicable"
}

// AllowedFlags is the exact set of brew flags permitted on install/reinstall/upgrade.
var AllowedFlags = map[string]struct{}{
	"--quiet":          {},
	"--no-auto-update": {},
	"--formula":        {},
	"--cask":           {},
}

// DangerousFlags is the enumerated rejection list for fast paths.
var DangerousFlags = map[string]struct{}{
	"--force":             {},
	"--HEAD":              {},
	"--debug-symbols":     {},
	"--build-from-source": {},
}

// AllowedTaps is the allow-list of trusted third-party taps. Default is empty,
// meaning `brew tap` is blocked by default; adding a tap requires explicit
// code review. homebrew/core is the built-in default tap -- it is never the
// subject of a `brew tap` command, so it does not need to appear here.
var AllowedTaps = map[string]struct{}{
	// intentionally empty -- override by code change with PR review
}

// urlPattern matches http/https/git URL prefixes.
var urlPattern = regexp.MustCompile(`^(https?|git)://`)

// shellMeta matches characters that could chain commands.
var shellMeta = regexp.MustCompile("[;&|<>`$()]")

// packageNamePattern is the regex safe package names must match.
var packageNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*(@[0-9.]+)?$`)

// tapNamePattern is the regex tap names must match (owner/tapname form).
var tapNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*/[a-z0-9][a-z0-9-]*$`)

// redirectPattern matches common shell redirects that are not brew arguments.
var redirectPattern = regexp.MustCompile(`^[0-9]*>[>&]?[0-9]*$`)

// isRedirect reports whether a token is a shell redirect (e.g. 2>&1, >/dev/null).
// These are not brew arguments and should be skipped during validation.
func isRedirect(token string) bool {
	return redirectPattern.MatchString(token) ||
		strings.HasPrefix(token, ">/") ||
		strings.HasPrefix(token, "2>/") ||
		strings.HasPrefix(token, "1>/") ||
		token == "/dev/null"
}

// Tokenize splits cmd into tokens respecting single and double quotes.
// Backslash escapes are supported inside double quotes only.
// Returns an error if a quote is unterminated.
func Tokenize(cmd string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escape := false
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

// ValidateBrewCommand applies the decision tree to a tokenized command.
// It is a pure function -- no I/O, no side effects.
func ValidateBrewCommand(tokens []string) Decision {
	if len(tokens) < 2 {
		return Decision{Allow: false, Reason: "empty brew command", Action: "block"}
	}

	sub := tokens[1]

	// `brew tap` is specially handled.
	if sub == "tap" {
		return validateTapCommand(tokens[2:])
	}

	// Non-install subcommands beyond tap are read-only and always safe.
	if sub != "install" && sub != "reinstall" && sub != "upgrade" {
		return Decision{Allow: true, Action: "allow"}
	}

	// Scan arguments after the subcommand.
	args := tokens[2:]
	sawPositional := false
	for _, a := range args {
		// Shell redirects are not brew arguments -- skip them.
		if isRedirect(a) {
			continue
		}

		// URL-based installs are a supply-chain risk.
		if urlPattern.MatchString(a) {
			return Decision{Allow: false, Reason: "url-based install: " + a, Action: "block"}
		}

		// Shell metacharacters suggest injection or command chaining.
		if shellMeta.MatchString(a) {
			return Decision{Allow: false, Reason: "shell metacharacter in arg: " + a, Action: "block"}
		}

		if strings.HasPrefix(a, "--") {
			// Dangerous flags: explicit block with a clear reason.
			if _, bad := DangerousFlags[a]; bad {
				return Decision{Allow: false, Reason: "disallowed flag: " + a, Action: "block"}
			}
			// Allowed flags: continue.
			if _, ok := AllowedFlags[a]; ok {
				continue
			}
			// Unknown flags: block by default.
			return Decision{Allow: false, Reason: "unknown flag: " + a, Action: "block"}
		}

		// Positional args: must be a valid package name.
		if !packageNamePattern.MatchString(a) {
			return Decision{Allow: false, Reason: "suspicious package name: " + a, Action: "block"}
		}
		sawPositional = true
	}

	if !sawPositional {
		return Decision{Allow: false, Reason: "no package name provided", Action: "block"}
	}

	return Decision{Allow: true, Action: "allow"}
}

// validateTapCommand applies the tap-specific rules.
func validateTapCommand(args []string) Decision {
	if len(args) == 0 {
		return Decision{Allow: true, Action: "allow"} // `brew tap` alone lists current taps
	}
	for _, a := range args {
		if shellMeta.MatchString(a) {
			return Decision{Allow: false, Reason: "shell metacharacter in tap arg: " + a, Action: "block"}
		}
		if urlPattern.MatchString(a) {
			return Decision{Allow: false, Reason: "url-based tap: " + a, Action: "block"}
		}
		if a == "--force-auto-update" || a == "--quiet" {
			continue
		}
		if strings.HasPrefix(a, "--") {
			return Decision{Allow: false, Reason: "disallowed tap flag: " + a, Action: "block"}
		}
		if !tapNamePattern.MatchString(a) {
			return Decision{Allow: false, Reason: "malformed tap name: " + a, Action: "block"}
		}
		if _, ok := AllowedTaps[a]; !ok {
			return Decision{Allow: false, Reason: "tap not on allow-list: " + a, Action: "block"}
		}
	}
	return Decision{Allow: true, Action: "allow"}
}
