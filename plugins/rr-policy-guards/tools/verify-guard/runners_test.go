// runners_test.go -- uniqueRunsOn, CheckRunnerAvailability against a
// mock ForgeClient, and the GitHub-hosted allowlist.

package main

import (
	"context"
	"sort"
	"testing"
)

func TestUniqueRunsOn(t *testing.T) {
	wfs := []Workflow{
		{RunsOn: []string{"ubuntu-latest", "self-hosted"}},
		{RunsOn: []string{"ubuntu-latest", "gpu"}},
	}
	got := uniqueRunsOn(wfs)
	sort.Strings(got)
	want := []string{"gpu", "self-hosted", "ubuntu-latest"}
	if len(got) != len(want) {
		t.Fatalf("uniqueRunsOn len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("uniqueRunsOn[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRunnerOnline(t *testing.T) {
	cases := map[string]bool{
		"online":  true,
		"idle":    true,
		"active":  true,
		"offline": false,
		"":        false,
		"asleep":  false,
	}
	for status, want := range cases {
		got := runnerOnline(Runner{Status: status})
		if got != want {
			t.Errorf("runnerOnline(%q) = %v, want %v", status, got, want)
		}
	}
}

// stubClient is a ForgeClient that returns canned runner lists.
type stubClient struct {
	forge       Forge
	creds       bool
	repoRunners []Runner
	orgRunners  []Runner
}

func (s *stubClient) Forge() Forge         { return s.forge }
func (s *stubClient) BaseURL() string      { return "https://stub" }
func (s *stubClient) HasCredentials() bool { return s.creds }
func (s *stubClient) ListRepoRunners(_ context.Context, _, _ string) ([]Runner, error) {
	return s.repoRunners, nil
}
func (s *stubClient) ListOrgRunners(_ context.Context, _ string) ([]Runner, error) {
	return s.orgRunners, nil
}
func (s *stubClient) HeadRepo(_ context.Context, _, _ string) (int, error) { return 200, nil }
func (s *stubClient) HeadRef(_ context.Context, _, _, _ string) (int, error) {
	return 200, nil
}

func TestCheckRunnerAvailability_AllSatisfied_GiteaSelfHosted(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	client := &stubClient{
		forge: ForgeGitea, creds: true,
		repoRunners: []Runner{{ID: 1, Name: "host1", Status: "online", Labels: []string{"ubuntu-latest", "self-hosted"}}},
	}
	wfs := []Workflow{{RunsOn: []string{"ubuntu-latest"}}}
	res, err := CheckRunnerAvailability(context.Background(), client, "o", "r", wfs)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Unsatisfied) != 0 {
		t.Errorf("expected satisfied; got unsatisfied=%v", res.Unsatisfied)
	}
}

func TestCheckRunnerAvailability_OfflineRunnerExcluded(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	client := &stubClient{
		forge: ForgeGitea, creds: true,
		repoRunners: []Runner{{ID: 1, Status: "offline", Labels: []string{"ubuntu-latest"}}},
	}
	wfs := []Workflow{{RunsOn: []string{"ubuntu-latest"}}}
	res, _ := CheckRunnerAvailability(context.Background(), client, "o", "r", wfs)
	if len(res.Unsatisfied) != 1 || res.Unsatisfied[0] != "ubuntu-latest" {
		t.Errorf("offline runner should NOT satisfy label; got %+v", res)
	}
}

func TestCheckRunnerAvailability_GitHubHostedAllowlist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	client := &stubClient{forge: ForgeGitHub, creds: true} // no self-hosted runners
	wfs := []Workflow{{RunsOn: []string{"ubuntu-latest", "macos-latest"}}}
	res, _ := CheckRunnerAvailability(context.Background(), client, "o", "r", wfs)
	if len(res.Unsatisfied) != 0 {
		t.Errorf("hosted-runner allowlist should satisfy ubuntu-latest and macos-latest; got %+v", res)
	}
}

func TestCheckRunnerAvailability_OrgRunnerCovers(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	client := &stubClient{
		forge: ForgeGitea, creds: true,
		orgRunners: []Runner{{ID: 9, Status: "online", Labels: []string{"big-runner"}}},
	}
	wfs := []Workflow{{RunsOn: []string{"big-runner"}}}
	res, _ := CheckRunnerAvailability(context.Background(), client, "o", "r", wfs)
	if len(res.Unsatisfied) != 0 {
		t.Errorf("org-level runner should be enumerated; got %+v", res)
	}
}

func TestCheckRunnerAvailability_NoLabels(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	client := &stubClient{forge: ForgeGitea, creds: true}
	res, err := CheckRunnerAvailability(context.Background(), client, "o", "r", []Workflow{{RunsOn: nil}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Required) != 0 || len(res.Unsatisfied) != 0 {
		t.Errorf("empty required should produce empty result; got %+v", res)
	}
}
