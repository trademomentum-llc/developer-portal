// extract.go -- parse a Bash command string into a CommitInvocation.
//
// The extractor is intentionally simple: it tokenises a shell command, handles
// single and double quotes, recognises -m / -F / --amend / -C, and STOPS at
// the first git subcommand. It does not expand variables, run subshells, or
// resolve command substitutions -- those are not safe to evaluate, and any
// commit pipeline that relies on them lies outside the guard's contract.
package main

import (
	"strings"
)

// ExtractCommit parses a Bash command and returns a CommitInvocation. The
// IsCommit field is true only when the command's leading binary is `git` and
// the first non-option token after `git` (honouring `-C <dir>`) is `commit`.
//
// Whitespace-only or empty inputs return IsCommit=false.
//
// The function never errors; ambiguous or malformed input simply yields
// IsCommit=false so the guard ignores it and lets other guards or the shell
// surface the problem.
func ExtractCommit(cmd string) CommitInvocation {
	toks := tokenize(cmd)
	if len(toks) == 0 {
		return CommitInvocation{}
	}
	// Strip any leading env-var assignments like FOO=bar.
	i := 0
	for i < len(toks) && isAssignment(toks[i]) {
		i++
	}
	if i >= len(toks) {
		return CommitInvocation{}
	}
	if toks[i] != "git" {
		return CommitInvocation{}
	}
	i++

	inv := CommitInvocation{}

	// Honour `git -C <dir>` (one or more occurrences).
	for i < len(toks) && toks[i] == "-C" {
		if i+1 >= len(toks) {
			return CommitInvocation{}
		}
		inv.RepoDir = toks[i+1]
		i += 2
	}

	// Skip non-meaningful global git options (e.g. --no-pager) until we hit a
	// subcommand. A global option starts with '-' but is not -C.
	for i < len(toks) && strings.HasPrefix(toks[i], "-") {
		i++
	}
	if i >= len(toks) {
		return CommitInvocation{}
	}
	if toks[i] != "commit" {
		return CommitInvocation{}
	}
	inv.IsCommit = true
	i++

	// Parse commit flags.
	for i < len(toks) {
		t := toks[i]
		switch {
		case t == "-m" || t == "--message":
			if i+1 < len(toks) {
				inv.MessageArgs = append(inv.MessageArgs, toks[i+1])
				i += 2
				continue
			}
			i++
		case strings.HasPrefix(t, "--message="):
			inv.MessageArgs = append(inv.MessageArgs, strings.TrimPrefix(t, "--message="))
			i++
		case t == "-F" || t == "--file":
			if i+1 < len(toks) {
				inv.FileArg = toks[i+1]
				i += 2
				continue
			}
			i++
		case strings.HasPrefix(t, "--file="):
			inv.FileArg = strings.TrimPrefix(t, "--file=")
			i++
		case t == "--amend":
			inv.Amend = true
			i++
		case t == "--no-edit":
			inv.NoEdit = true
			i++
		default:
			i++
		}
	}
	return inv
}

// isAssignment reports whether tok looks like FOO=bar (i.e. an env-var
// assignment preceding a command).
func isAssignment(tok string) bool {
	eq := strings.IndexByte(tok, '=')
	if eq <= 0 {
		return false
	}
	for i := 0; i < eq; i++ {
		c := tok[i]
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '_':
		default:
			return false
		}
	}
	return true
}

// tokenize is a small POSIX-ish shell tokeniser. Single quotes preserve their
// contents verbatim; double quotes preserve content but allow backslash
// escapes for ", $, ` and \. No expansion is performed.
func tokenize(s string) []string {
	var toks []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	prevChar := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				cur.WriteByte(c)
			}
		case inDouble:
			if c == '\\' && i+1 < len(s) {
				next := s[i+1]
				switch next {
				case '"', '\\', '$', '`':
					cur.WriteByte(next)
					i++
					continue
				}
				cur.WriteByte(c)
			} else if c == '"' {
				inDouble = false
			} else {
				cur.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == '\\' && i+1 < len(s):
			cur.WriteByte(s[i+1])
			i++
		case c == ' ' || c == '\t' || c == '\n':
			if cur.Len() > 0 {
				toks = append(toks, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
		prevChar = c
	}
	_ = prevChar
	if cur.Len() > 0 {
		toks = append(toks, cur.String())
	}
	return toks
}
