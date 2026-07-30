// runners.go -- forge-API runner enumeration and label cross-check.

package main

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// gitHubHostedLabels is the static allowlist of GitHub-hosted runner
// labels. A runs-on value matching any of these on a github repo is
// considered satisfiable without an API call. Keep in sync with
// https://docs.github.com/en/actions/using-github-hosted-runners.
var gitHubHostedLabels = map[string]struct{}{
	"ubuntu-latest":  {},
	"ubuntu-22.04":   {},
	"ubuntu-20.04":   {},
	"macos-latest":   {},
	"macos-13":       {},
	"macos-14":       {},
	"windows-latest": {},
	"windows-2022":   {},
	"windows-2019":   {},
}

// AvailabilityResult summarises the runner-vs-label cross-check.
type AvailabilityResult struct {
	Forge       Forge
	Required    []string
	Satisfied   []string
	Unsatisfied []string
}

// CheckRunnerAvailability fetches runners from the forge for owner/repo
// (and the repo's org), unions their online labels with hosted labels
// (for GitHub), and reports any required label that is not satisfied.
func CheckRunnerAvailability(ctx context.Context, client ForgeClient, owner, repo string, workflows []Workflow) (AvailabilityResult, error) {
	required := uniqueRunsOn(workflows)
	if len(required) == 0 {
		return AvailabilityResult{Forge: client.Forge()}, nil
	}
	cacheKey := fmt.Sprintf("%s-%s-%s", client.Forge(), owner, repo)
	var runners []Runner
	if !shortCacheRead("runners", cacheKey, 60*time.Second, &runners) {
		var err error
		runners, err = fetchAllRunners(ctx, client, owner, repo)
		if err != nil {
			return AvailabilityResult{Forge: client.Forge(), Required: required}, err
		}
		_ = shortCacheWrite("runners", cacheKey, runners)
	}

	satisfied := map[string]struct{}{}
	for _, r := range runners {
		if !runnerOnline(r) {
			continue
		}
		for _, lbl := range r.Labels {
			satisfied[lbl] = struct{}{}
		}
	}
	if client.Forge() == ForgeGitHub {
		for lbl := range gitHubHostedLabels {
			satisfied[lbl] = struct{}{}
		}
	}
	res := AvailabilityResult{Forge: client.Forge(), Required: required}
	for _, lbl := range required {
		if _, ok := satisfied[lbl]; ok {
			res.Satisfied = append(res.Satisfied, lbl)
		} else {
			res.Unsatisfied = append(res.Unsatisfied, lbl)
		}
	}
	sort.Strings(res.Satisfied)
	sort.Strings(res.Unsatisfied)
	return res, nil
}

// uniqueRunsOn returns the deduplicated, sorted union of runs-on
// labels across every workflow.
func uniqueRunsOn(workflows []Workflow) []string {
	set := map[string]struct{}{}
	for _, w := range workflows {
		for _, l := range w.RunsOn {
			set[l] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for l := range set {
		out = append(out, l)
	}
	sort.Strings(out)
	return out
}

// fetchAllRunners merges repo-level and (best-effort) org-level runners.
func fetchAllRunners(ctx context.Context, client ForgeClient, owner, repo string) ([]Runner, error) {
	repoRunners, err := client.ListRepoRunners(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	orgRunners, _ := client.ListOrgRunners(ctx, owner)
	merged := append([]Runner{}, repoRunners...)
	seen := map[int64]struct{}{}
	for _, r := range repoRunners {
		seen[r.ID] = struct{}{}
	}
	for _, r := range orgRunners {
		if _, dup := seen[r.ID]; dup {
			continue
		}
		merged = append(merged, r)
		seen[r.ID] = struct{}{}
	}
	return merged, nil
}

// runnerOnline returns true for any status string we treat as live.
func runnerOnline(r Runner) bool {
	switch r.Status {
	case "online", "idle", "active":
		return true
	}
	return false
}
