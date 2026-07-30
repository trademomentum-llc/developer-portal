// actions_test.go -- ParseUses and ResolveAllActions (with a fake
// ForgeClient for the resolution path).

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseUses(t *testing.T) {
	cases := []struct {
		in     string
		want   ActionRef
		hasErr bool
	}{
		{"actions/checkout@v4", ActionRef{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Forge: ForgeGitHub}, false},
		{"a/b/sub@main", ActionRef{Raw: "a/b/sub@main", Owner: "a", Repo: "b", Path: "sub", Ref: "main", Forge: ForgeGitea}, false},
		{"./local/path", ActionRef{Raw: "./local/path", Path: "./local/path", Local: true, Forge: ForgeGitea}, false},
		{".", ActionRef{Raw: ".", Path: ".", Local: true, Forge: ForgeGitea}, false},
		{"docker://alpine:3.19", ActionRef{Raw: "docker://alpine:3.19", Docker: true, Forge: ForgeGitea}, false},
		{"missing-at-ref", ActionRef{Raw: "missing-at-ref", Forge: ForgeGitea}, true},
		{"only-owner@ref", ActionRef{Raw: "only-owner@ref", Forge: ForgeGitea}, true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			inferred := ForgeGitea
			if c.want.Forge == ForgeGitHub {
				inferred = ForgeGitHub
			}
			got, err := ParseUses(c.in, inferred)
			if c.hasErr && err == nil {
				t.Fatalf("expected error for %q", c.in)
			}
			if !c.hasErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", c.in, err)
			}
			if !c.hasErr {
				if got.Owner != c.want.Owner || got.Repo != c.want.Repo || got.Ref != c.want.Ref ||
					got.Path != c.want.Path || got.Local != c.want.Local || got.Docker != c.want.Docker {
					t.Errorf("ParseUses(%q) = %+v, want %+v", c.in, got, c.want)
				}
			}
		})
	}
}

// fakeClient implements ForgeClient with scriptable HEAD responses.
type fakeClient struct {
	forge Forge
	creds bool
	repos map[string]int // "owner/repo" -> status
	refs  map[string]int // "owner/repo/ref" -> status
}

func (f *fakeClient) Forge() Forge         { return f.forge }
func (f *fakeClient) BaseURL() string      { return "https://fake" }
func (f *fakeClient) HasCredentials() bool { return f.creds }
func (f *fakeClient) ListRepoRunners(ctx context.Context, o, r string) ([]Runner, error) {
	return nil, nil
}
func (f *fakeClient) ListOrgRunners(ctx context.Context, o string) ([]Runner, error) {
	return nil, nil
}
func (f *fakeClient) HeadRepo(ctx context.Context, o, r string) (int, error) {
	if status, ok := f.repos[o+"/"+r]; ok {
		return status, nil
	}
	return 404, nil
}
func (f *fakeClient) HeadRef(ctx context.Context, o, r, ref string) (int, error) {
	if status, ok := f.refs[o+"/"+r+"/"+ref]; ok {
		return status, nil
	}
	return 404, nil
}

func TestResolveAllActions_Positive(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RR_VERIFY_GUARD_SKIP_ACTIONS", "")
	wf := Workflow{
		Path:          "/repo/.gitea/workflows/ci.yml",
		Source:        ".gitea/workflows/",
		InferredForge: ForgeGitea,
		Uses:          []string{"foo/bar@v1"},
	}
	client := &fakeClient{
		forge: ForgeGitea, creds: true,
		repos: map[string]int{"foo/bar": 200},
		refs:  map[string]int{"foo/bar/v1": 200},
	}
	ctx := context.Background()
	results, fail, _, err := ResolveAllActions(ctx, map[Forge]ForgeClient{ForgeGitea: client}, []Workflow{wf})
	if err != nil {
		t.Fatal(err)
	}
	if fail != nil {
		t.Fatalf("expected no failure, got %+v", fail)
	}
	if len(results) != 1 || !results[0].RepoExists || !results[0].RefExists {
		t.Errorf("unexpected results: %+v", results)
	}
}

func TestResolveAllActions_RefNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RR_VERIFY_GUARD_SKIP_ACTIONS", "")
	wf := Workflow{
		Path:          "/repo/.gitea/workflows/ci.yml",
		Source:        ".gitea/workflows/",
		InferredForge: ForgeGitea,
		Uses:          []string{"foo/bar@vGHOST"},
	}
	client := &fakeClient{
		forge: ForgeGitea, creds: true,
		repos: map[string]int{"foo/bar": 200},
		// no ref entry -> 404
	}
	_, fail, _, err := ResolveAllActions(context.Background(), map[Forge]ForgeClient{ForgeGitea: client}, []Workflow{wf})
	if err != nil {
		t.Fatalf("non-resolution should not be a hard error, got %v", err)
	}
	if fail == nil {
		t.Fatal("expected failure for missing ref")
	}
	if fail.RepoExists != true || fail.RefExists != false {
		t.Errorf("expected RepoExists=true, RefExists=false; got %+v", fail)
	}
}

func TestResolveAllActions_LocalPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	// Create a workflow file at .github/workflows/ci.yml referencing ./local
	wfDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".github", "workflows", "local"), 0755); err != nil {
		t.Fatal(err)
	}
	wfPath := filepath.Join(wfDir, "ci.yml")
	if err := os.WriteFile(wfPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	wf := Workflow{Path: wfPath, Source: ".github/workflows/", InferredForge: ForgeGitHub, Uses: []string{"./local"}}
	_, fail, _, err := ResolveAllActions(context.Background(), map[Forge]ForgeClient{}, []Workflow{wf})
	if err != nil {
		t.Fatal(err)
	}
	if fail != nil {
		t.Errorf("local path should resolve clean: %+v", fail)
	}
}

func TestResolveAllActions_LocalPathMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	wfDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatal(err)
	}
	wfPath := filepath.Join(wfDir, "ci.yml")
	_ = os.WriteFile(wfPath, []byte(""), 0644)
	wf := Workflow{Path: wfPath, Source: ".github/workflows/", InferredForge: ForgeGitHub, Uses: []string{"./does-not-exist"}}
	_, fail, _, err := ResolveAllActions(context.Background(), map[Forge]ForgeClient{}, []Workflow{wf})
	if err == nil {
		t.Fatal("expected error for missing local path")
	}
	if fail == nil {
		t.Fatal("expected non-nil failure for missing local path")
	}
}

func TestResolveAllActions_DockerSkipped(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	wf := Workflow{
		Source:        ".github/workflows/",
		InferredForge: ForgeGitHub,
		Uses:          []string{"docker://alpine:3.19"},
	}
	results, fail, degraded, err := ResolveAllActions(context.Background(), map[Forge]ForgeClient{ForgeGitHub: &fakeClient{forge: ForgeGitHub, creds: true}}, []Workflow{wf})
	if err != nil || fail != nil {
		t.Fatalf("docker should be skipped not failed: err=%v fail=%+v", err, fail)
	}
	if len(results) != 1 || results[0].Skipped != "docker" {
		t.Errorf("expected docker skip; got %+v", results)
	}
	found := false
	for _, d := range degraded {
		if d == "docker-action-skipped / docker://alpine:3.19" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected degraded entry for docker action; got %v", degraded)
	}
}

func TestResolveAllActions_NoCredentialsSkipped(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RR_VERIFY_GUARD_SKIP_ACTIONS", "")
	wf := Workflow{
		Source:        ".gitea/workflows/",
		InferredForge: ForgeGitea,
		Uses:          []string{"foo/bar@v1"},
	}
	client := &fakeClient{forge: ForgeGitea, creds: false}
	results, fail, _, err := ResolveAllActions(context.Background(), map[Forge]ForgeClient{ForgeGitea: client}, []Workflow{wf})
	if err != nil || fail != nil {
		t.Fatalf("missing creds should be degraded not failed: err=%v fail=%+v", err, fail)
	}
	if len(results) != 1 || results[0].Skipped != "no-credentials" {
		t.Errorf("expected no-credentials skip; got %+v", results)
	}
}
