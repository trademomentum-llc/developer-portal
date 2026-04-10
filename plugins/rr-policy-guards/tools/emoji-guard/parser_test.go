// parser_test.go -- unit tests for Scan and ExtractContent.
//
// These tests are pure: no file I/O, no subprocess, no environment.
// Every table entry is a plain ASCII description of a case -- the non-ASCII
// characters under test are constructed via Go's Unicode escape syntax so
// the test source itself stays ASCII.

package main

import (
	"strings"
	"testing"
)

func TestScan_PureASCII_ReturnsNil(t *testing.T) {
	cases := []string{
		"",
		"hello world",
		"package main\n\nfunc main() {}\n",
		"# Markdown heading\n\nSome text with symbols: -> +-+ || <-\n",
		"a\nb\nc\n",
		"tab\there",
		"number 12345 and punctuation !@#$%^&*()_+-=[]{};':\",./<>?",
	}
	for _, c := range cases {
		if hit := Scan(c); hit != nil {
			t.Errorf("Scan(%q) returned hit %+v, expected nil", c, hit)
		}
	}
}

func TestScan_EmDash_Blocks(t *testing.T) {
	content := "hello \u2014 world" // em dash
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on em dash")
	}
	if hit.Rune != 0x2014 {
		t.Errorf("hit.Rune = U+%04X, want U+2014", hit.Rune)
	}
}

func TestScan_BoxDrawing_Blocks(t *testing.T) {
	content := "diagram\n\u2500\u2500\u2500\nend" // three horizontal box chars
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on box-drawing char")
	}
	if hit.Rune != 0x2500 {
		t.Errorf("hit.Rune = U+%04X, want U+2500", hit.Rune)
	}
	if hit.Line != 2 {
		t.Errorf("hit.Line = %d, want 2", hit.Line)
	}
}

func TestScan_Emoji_Blocks(t *testing.T) {
	content := "task \U0001F4AF complete" // hundred points emoji
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on emoji")
	}
	if hit.Rune != 0x1F4AF {
		t.Errorf("hit.Rune = U+%04X, want U+1F4AF", hit.Rune)
	}
}

func TestScan_CheckMark_Blocks(t *testing.T) {
	content := "\u2705 passed" // white heavy check mark
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on check mark")
	}
	if hit.Rune != 0x2705 {
		t.Errorf("hit.Rune = U+%04X, want U+2705", hit.Rune)
	}
	if hit.Line != 1 || hit.Column != 1 {
		t.Errorf("hit position = line %d col %d, want line 1 col 1", hit.Line, hit.Column)
	}
}

func TestScan_SmartQuotes_Blocks(t *testing.T) {
	content := "he said \u201Chello\u201D" // smart double quotes
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on smart quote")
	}
	if hit.Rune != 0x201C {
		t.Errorf("hit.Rune = U+%04X, want U+201C", hit.Rune)
	}
}

func TestScan_LineAndColumnTracking(t *testing.T) {
	// Put a non-ASCII rune on line 4, column 7.
	content := "line1\nline2\nline3\nabcdef\u00A7gh"
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on section sign")
	}
	if hit.Line != 4 {
		t.Errorf("hit.Line = %d, want 4", hit.Line)
	}
	if hit.Column != 7 {
		t.Errorf("hit.Column = %d, want 7", hit.Column)
	}
}

func TestScan_InvalidUTF8_Blocks(t *testing.T) {
	// Deliberately invalid UTF-8 byte.
	content := string([]byte{'a', 'b', 0xC3, 'c'})
	hit := Scan(content)
	if hit == nil {
		t.Fatal("Scan returned nil, expected hit on invalid UTF-8")
	}
	if !strings.Contains(hit.Category, "invalid UTF-8") {
		t.Errorf("hit.Category = %q, want to contain 'invalid UTF-8'", hit.Category)
	}
}

func TestExtractContent_Write(t *testing.T) {
	input := map[string]any{
		"file_path": "/tmp/x",
		"content":   "hello world",
	}
	got := ExtractContent("Write", input)
	if got != "hello world" {
		t.Errorf("ExtractContent(Write, ...) = %q, want %q", got, "hello world")
	}
}

func TestExtractContent_Edit(t *testing.T) {
	input := map[string]any{
		"file_path":  "/tmp/x",
		"old_string": "foo",
		"new_string": "bar",
	}
	got := ExtractContent("Edit", input)
	if got != "bar" {
		t.Errorf("ExtractContent(Edit, ...) = %q, want %q", got, "bar")
	}
}

func TestExtractContent_MultiEdit(t *testing.T) {
	input := map[string]any{
		"file_path": "/tmp/x",
		"edits": []any{
			map[string]any{"old_string": "a", "new_string": "alpha"},
			map[string]any{"old_string": "b", "new_string": "beta"},
			map[string]any{"old_string": "c", "new_string": "gamma"},
		},
	}
	got := ExtractContent("MultiEdit", input)
	want := "alpha\nbeta\ngamma"
	if got != want {
		t.Errorf("ExtractContent(MultiEdit, ...) = %q, want %q", got, want)
	}
}

func TestExtractContent_UnknownTool(t *testing.T) {
	got := ExtractContent("Bash", map[string]any{"command": "ls"})
	if got != "" {
		t.Errorf("ExtractContent(Bash, ...) = %q, want empty string", got)
	}
}

func TestExtractContent_MissingField(t *testing.T) {
	got := ExtractContent("Write", map[string]any{"file_path": "/tmp/x"})
	if got != "" {
		t.Errorf("ExtractContent(Write, missing content) = %q, want empty string", got)
	}
}
