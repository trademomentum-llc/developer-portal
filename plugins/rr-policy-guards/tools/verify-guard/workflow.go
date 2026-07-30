// workflow.go -- workflow file discovery and minimal-subset extraction.
//
// We do NOT implement a general YAML parser. We extract two specific
// directives from CI workflow files: `runs-on:` and `uses:`. The
// extractor handles:
//   runs-on: ubuntu-latest
//   runs-on: [ubuntu-latest, gpu]
//   runs-on:
//     - ubuntu-latest
//     - gpu
//   uses: actions/checkout@v4
//   - uses: actions/checkout@v4
//
// Other YAML constructs are tolerated; any line we cannot interpret is
// passed over silently. This is by design -- the extractor is a
// purpose-built scanner, not a YAML parser.

package main

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverWorkflows finds workflow files under the three canonical
// directories and returns them parsed.
func DiscoverWorkflows(repoRoot string) ([]Workflow, error) {
	dirs := []struct {
		path  string
		forge Forge
	}{
		{".github/workflows", ForgeGitHub},
		{".gitea/workflows", ForgeGitea},
		{".forgejo/workflows", ForgeForgejo},
	}
	var out []Workflow
	for _, d := range dirs {
		full := filepath.Join(repoRoot, d.path)
		matches, _ := filepath.Glob(filepath.Join(full, "*.yml"))
		matches2, _ := filepath.Glob(filepath.Join(full, "*.yaml"))
		matches = append(matches, matches2...)
		sort.Strings(matches)
		for _, m := range matches {
			wf, err := ParseWorkflow(m, d.path, d.forge)
			if err != nil {
				return nil, err
			}
			out = append(out, *wf)
		}
	}
	return out, nil
}

// ParseWorkflow loads a workflow file and extracts runs-on and uses.
func ParseWorkflow(path, source string, inferred Forge) (*Workflow, error) {
	b, err := readFileSafely(path, 1<<20)
	if err != nil {
		return nil, err
	}
	wf := &Workflow{Path: path, Source: source, InferredForge: inferred}
	wf.RunsOn = extractRunsOn(string(b))
	wf.Uses = extractUses(string(b))
	if len(wf.RunsOn) == 0 {
		// Workflows must declare at least one runs-on; flag as parse problem.
		return wf, errors.New("no runs-on labels found in " + path)
	}
	return wf, nil
}

// extractRunsOn scans the workflow text for runs-on declarations.
// Returns deduped, sorted labels.
func extractRunsOn(text string) []string {
	set := map[string]struct{}{}
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		line := stripComment(lines[i])
		trim := strings.TrimSpace(line)
		const key = "runs-on:"
		idx := strings.Index(trim, key)
		if idx != 0 {
			continue
		}
		val := strings.TrimSpace(trim[len(key):])
		switch {
		case val == "":
			// block-array form on subsequent lines
			indent := indentOf(line)
			j := i + 1
			for j < len(lines) {
				next := lines[j]
				ntrim := strings.TrimSpace(stripComment(next))
				if ntrim == "" {
					j++
					continue
				}
				if indentOf(next) <= indent {
					break
				}
				if strings.HasPrefix(ntrim, "-") {
					label := strings.TrimSpace(strings.TrimPrefix(ntrim, "-"))
					label = unquote(label)
					if label != "" {
						set[label] = struct{}{}
					}
				}
				j++
			}
		case strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]"):
			// flow-array form
			inner := strings.TrimSuffix(strings.TrimPrefix(val, "["), "]")
			for _, item := range strings.Split(inner, ",") {
				label := unquote(strings.TrimSpace(item))
				if label != "" {
					set[label] = struct{}{}
				}
			}
		default:
			set[unquote(val)] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		if k != "" {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// extractUses scans the workflow text for uses: declarations.
func extractUses(text string) []string {
	set := map[string]struct{}{}
	for _, raw := range strings.Split(text, "\n") {
		line := stripComment(raw)
		trim := strings.TrimSpace(line)
		// step-bullet form: "- uses: x"
		stripped := strings.TrimSpace(strings.TrimPrefix(trim, "-"))
		const key = "uses:"
		if !strings.HasPrefix(stripped, key) {
			continue
		}
		val := strings.TrimSpace(stripped[len(key):])
		val = unquote(val)
		if val != "" {
			set[val] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// stripComment removes any trailing # comment outside of quotes.
func stripComment(line string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '\'' && !inDouble {
			inSingle = !inSingle
		} else if c == '"' && !inSingle {
			inDouble = !inDouble
		} else if c == '#' && !inSingle && !inDouble {
			return line[:i]
		}
	}
	return line
}

// indentOf returns the count of leading spaces (tabs treated as one).
func indentOf(line string) int {
	n := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			n++
		case '\t':
			n++
		default:
			return n
		}
	}
	return n
}

// unquote strips matching surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
