// workflow_test.go -- tests for workflow discovery and minimal-subset
// extraction of runs-on + uses.

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestExtractRunsOn(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want []string
	}{
		{
			"string form",
			"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps: []\n",
			[]string{"ubuntu-latest"},
		},
		{
			"flow array",
			"jobs:\n  build:\n    runs-on: [ubuntu-latest, gpu]\n",
			[]string{"gpu", "ubuntu-latest"},
		},
		{
			"block array",
			"jobs:\n  build:\n    runs-on:\n      - ubuntu-latest\n      - self-hosted\n",
			[]string{"self-hosted", "ubuntu-latest"},
		},
		{
			"multiple jobs unioned",
			"jobs:\n  a:\n    runs-on: ubuntu-22.04\n  b:\n    runs-on: macos-latest\n",
			[]string{"macos-latest", "ubuntu-22.04"},
		},
		{
			"quoted",
			"jobs:\n  build:\n    runs-on: \"ubuntu-latest\"\n",
			[]string{"ubuntu-latest"},
		},
		{
			"comment ignored",
			"jobs:\n  build:\n    runs-on: ubuntu-latest # primary\n",
			[]string{"ubuntu-latest"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractRunsOn(c.yaml)
			sort.Strings(got)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("extractRunsOn diff: got %v want %v", got, c.want)
			}
		})
	}
}

func TestExtractUses(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want []string
	}{
		{
			"single uses",
			"steps:\n  - uses: actions/checkout@v4\n",
			[]string{"actions/checkout@v4"},
		},
		{
			"two distinct",
			"steps:\n  - uses: actions/checkout@v4\n  - uses: actions/setup-go@v5\n",
			[]string{"actions/checkout@v4", "actions/setup-go@v5"},
		},
		{
			"local path",
			"steps:\n  - uses: ./local/composite\n",
			[]string{"./local/composite"},
		},
		{
			"docker form",
			"steps:\n  - uses: docker://alpine:3.19\n",
			[]string{"docker://alpine:3.19"},
		},
		{
			"dedupes",
			"steps:\n  - uses: a/b@v1\n  - uses: a/b@v1\n",
			[]string{"a/b@v1"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractUses(c.yaml)
			sort.Strings(got)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("extractUses diff: got %v want %v", got, c.want)
			}
		})
	}
}

func TestStripComment(t *testing.T) {
	cases := map[string]string{
		"plain":              "plain",
		"with # comment":     "with ",
		`"# in dquote" rest`: `"# in dquote" rest`,
		"'# in squote' rest": "'# in squote' rest",
	}
	for in, want := range cases {
		got := stripComment(in)
		if got != want {
			t.Errorf("stripComment(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnquote(t *testing.T) {
	if got := unquote(`"x"`); got != "x" {
		t.Errorf("dquote unquote: %q", got)
	}
	if got := unquote(`'x'`); got != "x" {
		t.Errorf("squote unquote: %q", got)
	}
	if got := unquote(`x`); got != "x" {
		t.Errorf("bare unquote: %q", got)
	}
	if got := unquote(`"`); got != `"` {
		t.Errorf("single dquote: %q", got)
	}
}

func TestDiscoverWorkflows(t *testing.T) {
	dir := t.TempDir()
	mkWf := func(rel, body string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	wfBody := "jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n"
	mkWf(".github/workflows/ci.yml", wfBody)
	mkWf(".gitea/workflows/ci.yaml", wfBody)
	mkWf(".forgejo/workflows/test.yml", wfBody)

	wfs, err := DiscoverWorkflows(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(wfs) != 3 {
		t.Fatalf("expected 3 workflows, got %d", len(wfs))
	}
	forges := map[Forge]int{}
	for _, w := range wfs {
		forges[w.InferredForge]++
		if len(w.RunsOn) != 1 || w.RunsOn[0] != "ubuntu-latest" {
			t.Errorf("unexpected RunsOn: %v", w.RunsOn)
		}
		if len(w.Uses) != 1 || w.Uses[0] != "actions/checkout@v4" {
			t.Errorf("unexpected Uses: %v", w.Uses)
		}
	}
	if forges[ForgeGitHub] != 1 || forges[ForgeGitea] != 1 || forges[ForgeForgejo] != 1 {
		t.Errorf("forge inference wrong: %+v", forges)
	}
}

func TestParseWorkflowMissingRunsOn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")
	if err := os.WriteFile(path, []byte("jobs:\n  build:\n    steps: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseWorkflow(path, ".github/workflows/", ForgeGitHub)
	if err == nil {
		t.Fatal("expected error for missing runs-on")
	}
}
