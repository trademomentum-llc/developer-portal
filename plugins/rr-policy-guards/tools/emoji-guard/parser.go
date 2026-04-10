// parser.go -- non-ASCII character detection and content extraction.
//
// This file holds pure functions only: no I/O, no stdin reading, no exit
// codes. main.go wires these functions up to the Claude Code PreToolUse
// protocol. Everything here is unit-testable in isolation.

package main

import (
	"fmt"
	"unicode/utf8"
)

// Hit describes where a non-ASCII character was found in the scanned content.
type Hit struct {
	Rune       rune   // the offending code point
	ByteOffset int    // byte offset into the scanned string
	Line       int    // 1-based line number
	Column     int    // 1-based column number within the line (in runes)
	Category   string // human-readable category name for error messages
}

// String renders a Hit for human consumption in a hook stderr message.
func (h Hit) String() string {
	return fmt.Sprintf("line %d col %d: U+%04X (%s)",
		h.Line, h.Column, h.Rune, h.Category)
}

// categorize returns a short human-readable category name for a non-ASCII
// rune. It exists purely to make rejection messages self-explanatory; it
// does not influence the decision (every non-ASCII rune is rejected).
func categorize(r rune) string {
	switch {
	case r >= 0x0080 && r <= 0x00FF:
		return "Latin-1 supplement"
	case r >= 0x2000 && r <= 0x206F:
		return "general punctuation (em dash, ellipsis, smart quotes, etc)"
	case r >= 0x2100 && r <= 0x214F:
		return "letterlike symbols"
	case r >= 0x2190 && r <= 0x21FF:
		return "arrows"
	case r >= 0x2200 && r <= 0x22FF:
		return "mathematical operators"
	case r >= 0x2300 && r <= 0x23FF:
		return "miscellaneous technical symbols"
	case r >= 0x2500 && r <= 0x257F:
		return "box-drawing characters"
	case r >= 0x2580 && r <= 0x259F:
		return "block elements"
	case r >= 0x25A0 && r <= 0x25FF:
		return "geometric shapes"
	case r >= 0x2600 && r <= 0x26FF:
		return "miscellaneous symbols"
	case r >= 0x2700 && r <= 0x27BF:
		return "dingbats"
	case r >= 0x2B00 && r <= 0x2BFF:
		return "miscellaneous symbols and arrows"
	case r >= 0x1F000 && r <= 0x1FAFF:
		return "pictographic emoji"
	case r == 0xFE0F:
		return "variation selector-16 (emoji presentation)"
	case r == 0x200D:
		return "zero-width joiner (emoji sequence)"
	default:
		return "non-ASCII character"
	}
}

// Scan walks content rune by rune and returns the first non-ASCII rune it
// finds. If nothing is blocked, Scan returns nil. Scan runs in
// O(len(content)) time with a single pass and allocates nothing beyond the
// returned Hit.
//
// The rule is absolute: any rune with value > 0x7F is rejected. The
// category field in the returned Hit is informational only; the decision
// is based purely on the ASCII cutoff.
func Scan(content string) *Hit {
	line := 1
	col := 0
	i := 0
	for i < len(content) {
		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 byte sequence. Treat as a violation because
			// a PreToolUse hook should not trust malformed input.
			return &Hit{
				Rune:       0xFFFD,
				ByteOffset: i,
				Line:       line,
				Column:     col + 1,
				Category:   "invalid UTF-8 byte sequence",
			}
		}
		col++
		if r > 0x7F {
			return &Hit{
				Rune:       r,
				ByteOffset: i,
				Line:       line,
				Column:     col,
				Category:   categorize(r),
			}
		}
		if r == '\n' {
			line++
			col = 0
		}
		i += size
	}
	return nil
}

// ExtractContent pulls the relevant text-to-scan out of a PreToolUse
// tool_input payload based on the tool name.
//
// Supported tools:
//   - Write:     tool_input.content
//   - Edit:      tool_input.new_string
//   - MultiEdit: join of every edits[*].new_string with newline separator
//
// For unsupported tool names, ExtractContent returns the empty string, which
// the caller interprets as "nothing to scan, allow".
func ExtractContent(toolName string, toolInput map[string]any) string {
	switch toolName {
	case "Write":
		if v, ok := toolInput["content"].(string); ok {
			return v
		}
	case "Edit":
		if v, ok := toolInput["new_string"].(string); ok {
			return v
		}
	case "MultiEdit":
		raw, ok := toolInput["edits"].([]any)
		if !ok {
			return ""
		}
		var out []byte
		for _, e := range raw {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}
			if s, ok := m["new_string"].(string); ok {
				if len(out) > 0 {
					out = append(out, '\n')
				}
				out = append(out, s...)
			}
		}
		return string(out)
	}
	return ""
}
