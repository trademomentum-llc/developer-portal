// actions.go -- resolve every `uses:` reference in a workflow against
// the inferred forge.
//
// Reference forms recognised:
//   <owner>/<repo>@<ref>
//   <owner>/<repo>/<path>@<ref>
//   ./local/path
//   docker://image[:tag]

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ActionRef is a parsed `uses:` value.
type ActionRef struct {
	Raw    string
	Owner  string
	Repo   string
	Path   string
	Ref    string
	Forge  Forge
	Local  bool
	Docker bool
}

// ResolutionResult is one resolution outcome.
type ResolutionResult struct {
	Ref        ActionRef
	RepoExists bool
	RefExists  bool
	Skipped    string // non-empty when resolution was skipped (degraded)
	Error      error
}

// ParseUses splits a `uses:` value into its components, attaching the
// path-inferred forge.
func ParseUses(raw string, inferred Forge) (ActionRef, error) {
	r := ActionRef{Raw: raw, Forge: inferred}
	if strings.HasPrefix(raw, "./") || raw == "." {
		r.Local = true
		r.Path = raw
		return r, nil
	}
	if strings.HasPrefix(raw, "docker://") {
		r.Docker = true
		return r, nil
	}
	at := strings.LastIndex(raw, "@")
	if at < 0 {
		return r, fmt.Errorf("uses ref %q has no @ref component", raw)
	}
	repoSpec := raw[:at]
	r.Ref = raw[at+1:]
	parts := strings.SplitN(repoSpec, "/", 3)
	if len(parts) < 2 {
		return r, fmt.Errorf("uses ref %q is missing owner/repo", raw)
	}
	r.Owner = parts[0]
	r.Repo = parts[1]
	if len(parts) == 3 {
		r.Path = parts[2]
	}
	return r, nil
}

// ResolveAllActions resolves every `uses:` across every workflow.
// Returns one ResolutionResult per parseable reference; returns the
// first non-resolution failure as the second value (or nil on full
// success).
func ResolveAllActions(ctx context.Context, clientByForge map[Forge]ForgeClient, workflows []Workflow) ([]ResolutionResult, *ResolutionResult, []string, error) {
	var results []ResolutionResult
	var degraded []string
	for _, w := range workflows {
		for _, raw := range w.Uses {
			ref, err := ParseUses(raw, w.InferredForge)
			if err != nil {
				results = append(results, ResolutionResult{Ref: ref, Error: err})
				return results, &results[len(results)-1], degraded, err
			}
			if ref.Local {
				p := filepath.Join(filepath.Dir(w.Path), ref.Path)
				if _, statErr := os.Stat(p); statErr != nil {
					results = append(results, ResolutionResult{Ref: ref, Error: statErr})
					return results, &results[len(results)-1], degraded, statErr
				}
				results = append(results, ResolutionResult{Ref: ref, RepoExists: true, RefExists: true})
				continue
			}
			if ref.Docker {
				degraded = append(degraded, "docker-action-skipped / "+ref.Raw)
				results = append(results, ResolutionResult{Ref: ref, Skipped: "docker"})
				continue
			}
			if os.Getenv("RR_VERIFY_GUARD_SKIP_ACTIONS") == "1" {
				degraded = append(degraded, "actions-skip / env")
				results = append(results, ResolutionResult{Ref: ref, Skipped: "env-skip"})
				continue
			}
			client, ok := clientByForge[ref.Forge]
			if !ok || client == nil || !client.HasCredentials() {
				degraded = append(degraded, "action-resolution-skipped / "+string(ref.Forge))
				results = append(results, ResolutionResult{Ref: ref, Skipped: "no-credentials"})
				continue
			}
			rr, _ := resolveOne(ctx, client, ref)
			results = append(results, rr)
			if !rr.RepoExists || !rr.RefExists {
				return results, &results[len(results)-1], degraded, nil
			}
		}
	}
	return results, nil, degraded, nil
}

// resolveOne checks a single ActionRef against its forge, with caching.
func resolveOne(ctx context.Context, client ForgeClient, ref ActionRef) (ResolutionResult, error) {
	cacheKey := fmt.Sprintf("%s-%s-%s-%s", ref.Forge, ref.Owner, ref.Repo, ref.Ref)
	var cached struct {
		RepoExists bool `json:"repo_exists"`
		RefExists  bool `json:"ref_exists"`
	}
	if shortCacheRead("actions", cacheKey, 60*time.Second, &cached) {
		return ResolutionResult{Ref: ref, RepoExists: cached.RepoExists, RefExists: cached.RefExists}, nil
	}
	repoStatus, err := client.HeadRepo(ctx, ref.Owner, ref.Repo)
	if err != nil {
		return ResolutionResult{Ref: ref, Error: err}, err
	}
	repoExists := repoStatus >= 200 && repoStatus < 300
	if !repoExists {
		_ = shortCacheWrite("actions", cacheKey, struct {
			RepoExists bool `json:"repo_exists"`
			RefExists  bool `json:"ref_exists"`
		}{false, false})
		return ResolutionResult{Ref: ref, RepoExists: false, RefExists: false}, nil
	}
	refStatus, err := client.HeadRef(ctx, ref.Owner, ref.Repo, ref.Ref)
	if err != nil {
		return ResolutionResult{Ref: ref, RepoExists: true, Error: err}, err
	}
	refExists := refStatus >= 200 && refStatus < 300
	_ = shortCacheWrite("actions", cacheKey, struct {
		RepoExists bool `json:"repo_exists"`
		RefExists  bool `json:"ref_exists"`
	}{true, refExists})
	return ResolutionResult{Ref: ref, RepoExists: true, RefExists: refExists}, nil
}
