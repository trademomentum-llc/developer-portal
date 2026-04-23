// scanner.go -- unsafe shell pattern detection and safe-syntax correction.
//
// This file holds pure functions only: no I/O, no stdin reading, no exit
// codes. main.go wires these functions up to the Claude Code PreToolUse
// protocol. Everything here is unit-testable in isolation.
//
// Detection rules:
//
//   SIMPLE_EXPANSION  Bare $VAR references (not inside single quotes or
//                     escaped). These are non-deterministic because the
//                     expansion depends on the shell environment at runtime.
//                     Safe alternative: use the literal value or
//                     pass via run_in_background with absolute paths.
//
// Each rule has a corresponding fix function that rewrites the unsafe
// pattern into safe syntax. The fix is included in the block message so
// Claude can re-issue the command without human intervention.

package main

import (
	"strings"
	"unicode/utf8"
)

// Violation describes a single unsafe pattern found in a command.
type Violation struct {
	Rule        string // rule identifier, e.g. "SIMPLE_EXPANSION"
	Description string // human-readable explanation
	Offset      int    // byte offset into the command string
	Length      int    // byte length of the matched pattern
	Original    string // the matched text
	Replacement string // the safe replacement text
}

// ScanCommand walks a command string and returns all violations found.
// Returns nil if the command is safe.
func ScanCommand(command string) []Violation {
	var violations []Violation
	violations = append(violations, scanSimpleExpansion(command)...)
	return violations
}

// ApplyFixes takes the original command and a list of violations and
// returns a corrected command with all unsafe patterns replaced by their
// safe alternatives. Violations must be sorted by offset (ScanCommand
// returns them in order) and offsets reference the original command
// string, not a partially-modified copy.
func ApplyFixes(command string, violations []Violation) string {
	if len(violations) == 0 {
		return command
	}

	var b strings.Builder
	b.Grow(len(command))
	prev := 0

	for _, v := range violations {
		// Skip overlapping violations (should not happen with current
		// rules, but defensive).
		if v.Offset < prev {
			continue
		}

		// Write everything between the last replacement and this violation.
		b.WriteString(command[prev:v.Offset])

		// Write the replacement.
		b.WriteString(v.Replacement)

		prev = v.Offset + v.Length
	}

	// Remainder of the command after the last violation.
	if prev < len(command) {
		b.WriteString(command[prev:])
	}

	return b.String()
}

// scanSimpleExpansion detects bare $VAR patterns outside of single quotes.
// A bare $VAR is a dollar sign followed by an identifier character
// (letter, digit, underscore) that is NOT:
//   - Inside single quotes (where expansion does not occur)
//   - Preceded by a backslash (escaped)
//   - Part of $(...) command substitution (allowed, different semantics)
//   - Part of ${...} brace expansion (also flagged)
//   - $? or $! or $$ (special shell variables -- allowed, deterministic)
func scanSimpleExpansion(command string) []Violation {
	var violations []Violation
	inSingleQuote := false
	i := 0

	for i < len(command) {
		r, size := utf8.DecodeRuneInString(command[i:])

		if r == '\'' && !isEscaped(command, i) {
			inSingleQuote = !inSingleQuote
			i += size
			continue
		}

		if inSingleQuote {
			i += size
			continue
		}

		// Heredoc with a quoted delimiter (<<'DELIM', <<"DELIM", <<\DELIM,
		// and the <<- tab-stripping variants of each) has a fully literal
		// body -- no parameter, command, or arithmetic expansion. Skip the
		// body entirely so variable-looking text inside commit messages
		// and similar documentation does not produce false positives.
		// <<< (here-string) is NOT a heredoc and is handled by normal
		// scanning below.
		if r == '<' && i+1 < len(command) && command[i+1] == '<' &&
			!(i+2 < len(command) && command[i+2] == '<') {
			if end, skip := parseQuotedHeredoc(command, i); skip {
				i = end
				continue
			}
		}

		if r == '$' && !isEscaped(command, i) {
			v := classifyDollar(command, i)
			if v != nil {
				violations = append(violations, *v)
				i += v.Length
				continue
			}
		}

		i += size
	}

	return violations
}

// parseQuotedHeredoc examines an occurrence of `<<` at pos. If the heredoc
// has a quoted delimiter (single-quoted, double-quoted, or backslash-
// escaped), it returns the position immediately after the closing
// delimiter line and skip=true so the caller can jump past the literal
// body. Otherwise returns (pos, false) and the caller falls through to
// normal scanning.
func parseQuotedHeredoc(s string, pos int) (end int, skip bool) {
	p := pos + 2

	dash := false
	if p < len(s) && s[p] == '-' {
		dash = true
		p++
	}

	for p < len(s) && (s[p] == ' ' || s[p] == '\t') {
		p++
	}
	if p >= len(s) {
		return pos, false
	}

	quoteChar := byte(0)
	if c := s[p]; c == '\'' || c == '"' {
		quoteChar = c
		p++
	} else if c == '\\' {
		quoteChar = '\\'
		p++
	}

	delimStart := p
	for p < len(s) && (isIdentChar(s[p]) || s[p] == '-') {
		p++
	}
	if p == delimStart {
		return pos, false
	}
	delim := s[delimStart:p]

	if quoteChar == '\'' || quoteChar == '"' {
		if p >= len(s) || s[p] != quoteChar {
			return pos, false
		}
		p++
	}

	if quoteChar == 0 {
		return pos, false
	}

	for p < len(s) && s[p] != '\n' {
		p++
	}
	if p < len(s) {
		p++
	}

	for p < len(s) {
		lineStart := p
		q := p
		for dash && q < len(s) && s[q] == '\t' {
			q++
		}
		if q+len(delim) <= len(s) && s[q:q+len(delim)] == delim {
			after := q + len(delim)
			if after == len(s) || s[after] == '\n' {
				if after < len(s) {
					after++
				}
				return after, true
			}
		}
		p = lineStart
		for p < len(s) && s[p] != '\n' {
			p++
		}
		if p < len(s) {
			p++
		}
	}

	return pos, false
}

// classifyDollar examines a $ at position pos and determines if it is
// an unsafe expansion. Returns nil if the pattern is safe.
func classifyDollar(command string, pos int) *Violation {
	if pos+1 >= len(command) {
		return nil // trailing $, harmless
	}

	next := command[pos+1]

	// $(...) command substitution -- allowed
	if next == '(' {
		return nil
	}

	// $? $! $$ $# $@ $* $- $0-$9 -- special variables, deterministic
	if isSpecialShellVar(next) {
		return nil
	}

	// ${VAR} brace expansion -- also unsafe
	if next == '{' {
		return classifyBraceExpansion(command, pos)
	}

	// $VAR simple expansion
	if isIdentStart(next) {
		return classifySimpleVar(command, pos)
	}

	return nil
}

// classifySimpleVar handles $VAR where VAR starts with a letter or underscore.
func classifySimpleVar(command string, pos int) *Violation {
	// Collect the variable name.
	end := pos + 1
	for end < len(command) && isIdentChar(command[end]) {
		end++
	}
	varName := command[pos+1 : end]
	original := command[pos:end]

	return &Violation{
		Rule:        "SIMPLE_EXPANSION",
		Description: "bare $" + varName + " expansion is non-deterministic; use the literal value or a known safe method",
		Offset:      pos,
		Length:       end - pos,
		Original:    original,
		Replacement: "<" + varName + "_VALUE>",
	}
}

// classifyBraceExpansion handles ${VAR} patterns.
func classifyBraceExpansion(command string, pos int) *Violation {
	// Find the closing brace.
	end := pos + 2
	for end < len(command) && command[end] != '}' {
		end++
	}
	if end >= len(command) {
		return nil // unclosed brace, not a valid expansion
	}
	end++ // include the closing brace

	varName := command[pos+2 : end-1]
	original := command[pos:end]

	return &Violation{
		Rule:        "SIMPLE_EXPANSION",
		Description: "bare ${" + varName + "} expansion is non-deterministic; use the literal value or a known safe method",
		Offset:      pos,
		Length:       end - pos,
		Original:    original,
		Replacement: "<" + varName + "_VALUE>",
	}
}

// isEscaped checks whether the character at position pos is preceded by
// an odd number of backslashes (i.e., it is escaped).
func isEscaped(s string, pos int) bool {
	count := 0
	for i := pos - 1; i >= 0 && s[i] == '\\'; i-- {
		count++
	}
	return count%2 == 1
}

// isIdentStart reports whether b can start a shell variable name.
func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// isIdentChar reports whether b can appear in a shell variable name.
func isIdentChar(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

// isSpecialShellVar reports whether b following $ is a special shell
// variable that is deterministic and safe to use.
func isSpecialShellVar(b byte) bool {
	switch b {
	case '?', '!', '$', '#', '@', '*', '-',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}
