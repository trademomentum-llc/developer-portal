// validator.go -- commit message validation (IN-M-* rules).
package main

import (
	"regexp"
	"strings"
)

// conventionalRE enforces the Conventional Commits subject form. The type
// list mirrors widely-used presets; teams that want a custom type list can
// extend this regex.
var conventionalRE = regexp.MustCompile(
	`^(feat|fix|chore|docs|refactor|test|build|ci|perf|style|revert)(\([^)]+\))?!?: .+`)

// MsgVerdict captures the result of validating a commit message.
type MsgVerdict struct {
	Decision Decision
	RuleCode string
	Reason   string
	Subject  string
}

// VerdictFromMessage parses a raw commit message and returns the IN-M-*
// verdict. Comment lines starting with '#' are stripped before parsing,
// matching git's own behaviour. Empty input is treated as a missing subject.
//
// Auto-messages from `git merge` and `git revert` are exempt from
// Conventional Commits enforcement (IN-M-004).
func VerdictFromMessage(raw []byte) MsgVerdict {
	text := stripComments(string(raw))
	lines := strings.Split(text, "\n")

	// Drop leading blank lines.
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) == 0 {
		return MsgVerdict{Decision: DecisionBlock, RuleCode: "IN-M-001",
			Reason: "subject line is empty"}
	}

	subject := strings.TrimRight(lines[0], " \t\r")
	if subject == "" {
		return MsgVerdict{Decision: DecisionBlock, RuleCode: "IN-M-001",
			Reason: "subject line is empty"}
	}

	// Exemptions (IN-M-004) come before length and convention checks because
	// merge/revert messages from git are auto-generated and may exceed 72 chars
	// or violate Conventional Commits.
	if strings.HasPrefix(subject, "Merge ") || strings.HasPrefix(subject, "Revert ") {
		return MsgVerdict{Decision: DecisionAllow, Subject: subject}
	}

	if len(subject) > 72 {
		return MsgVerdict{Decision: DecisionBlock, RuleCode: "IN-M-001",
			Reason: "subject longer than 72 chars", Subject: subject}
	}

	if !conventionalRE.MatchString(subject) {
		return MsgVerdict{Decision: DecisionBlock, RuleCode: "IN-M-002",
			Reason: "subject must be Conventional Commits form: 'type(scope): description'",
			Subject: subject}
	}

	if !hasBody(lines) {
		return MsgVerdict{Decision: DecisionBlock, RuleCode: "IN-M-003",
			Reason: "message missing body explaining the WHY of the change",
			Subject: subject}
	}

	return MsgVerdict{Decision: DecisionAllow, Subject: subject}
}

// stripComments removes git-style comment lines (those starting with '#').
func stripComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trim, "#") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// hasBody reports whether the message has a body following the convention
// "subject\n\nbody". Specifically: at least one blank line after the subject
// followed by at least one non-empty non-comment line.
func hasBody(lines []string) bool {
	if len(lines) < 3 {
		return false
	}
	// lines[0] is the subject; lines[1] should be blank.
	if strings.TrimSpace(lines[1]) != "" {
		return false
	}
	for _, l := range lines[2:] {
		if strings.TrimSpace(l) != "" {
			return true
		}
	}
	return false
}
