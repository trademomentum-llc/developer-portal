// forge_clients.go -- ForgeClient interface and three implementations
// (GitHub, Gitea, Forgejo). Forgejo and Gitea share endpoint shapes;
// Forgejo's client is a thin variant.
//
// Token material is read from env, used in-memory, and never written
// to logs, audit, or cache.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ForgeClient is the per-forge HTTP client surface used by runners.go
// and actions.go.
type ForgeClient interface {
	Forge() Forge
	BaseURL() string
	HasCredentials() bool
	ListRepoRunners(ctx context.Context, owner, repo string) ([]Runner, error)
	ListOrgRunners(ctx context.Context, org string) ([]Runner, error)
	HeadRepo(ctx context.Context, owner, repo string) (int, error)
	HeadRef(ctx context.Context, owner, repo, ref string) (int, error)
}

// NewForgeClient returns the appropriate client for the detected forge.
// Returns nil for ForgeNone or ForgeUnknown.
func NewForgeClient(f Forge) ForgeClient {
	switch f {
	case ForgeGitHub:
		return newGitHubClient()
	case ForgeGitea:
		return newGiteaClient()
	case ForgeForgejo:
		return newForgejoClient()
	}
	return nil
}

// ============================================================
// GitHub client
// ============================================================

type githubClient struct {
	base   string
	token  string
	client *http.Client
}

func newGitHubClient() *githubClient {
	tok := os.Getenv("RR_VERIFY_GUARD_GITHUB_TOKEN")
	if tok == "" {
		tok = os.Getenv("GITHUB_TOKEN")
	}
	if tok == "" {
		tok = os.Getenv("GH_TOKEN")
	}
	return &githubClient{
		base:   "https://api.github.com",
		token:  tok,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *githubClient) Forge() Forge         { return ForgeGitHub }
func (c *githubClient) BaseURL() string      { return c.base }
func (c *githubClient) HasCredentials() bool { return c.token != "" }

func (c *githubClient) auth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}

func (c *githubClient) ListRepoRunners(ctx context.Context, owner, repo string) ([]Runner, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/actions/runners?per_page=100", c.base, owner, repo)
	return c.fetchRunners(ctx, url, "github-repo-runners")
}

func (c *githubClient) ListOrgRunners(ctx context.Context, org string) ([]Runner, error) {
	url := fmt.Sprintf("%s/orgs/%s/actions/runners?per_page=100", c.base, org)
	return c.fetchRunners(ctx, url, "github-org-runners")
}

// fetchRunners handles GitHub's wrapper { total_count, runners[] }.
func (c *githubClient) fetchRunners(ctx context.Context, url, kind string) ([]Runner, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 || resp.StatusCode == 401 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: status %d", kind, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var wrapped struct {
		Runners []struct {
			ID     int64  `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
			Busy   bool   `json:"busy"`
			Labels []struct {
				Name string `json:"name"`
			} `json:"labels"`
		} `json:"runners"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}
	out := make([]Runner, 0, len(wrapped.Runners))
	for _, r := range wrapped.Runners {
		labels := make([]string, len(r.Labels))
		for i, l := range r.Labels {
			labels[i] = l.Name
		}
		out = append(out, Runner{ID: r.ID, Name: r.Name, Status: r.Status, Labels: labels})
	}
	return out, nil
}

func (c *githubClient) HeadRepo(ctx context.Context, owner, repo string) (int, error) {
	return doHead(ctx, c.client, fmt.Sprintf("%s/repos/%s/%s", c.base, owner, repo), c.auth)
}

func (c *githubClient) HeadRef(ctx context.Context, owner, repo, ref string) (int, error) {
	// Try as branch first, then tag.
	tries := []string{
		fmt.Sprintf("%s/repos/%s/%s/git/ref/heads/%s", c.base, owner, repo, ref),
		fmt.Sprintf("%s/repos/%s/%s/git/ref/tags/%s", c.base, owner, repo, ref),
	}
	for _, u := range tries {
		status, err := doHead(ctx, c.client, u, c.auth)
		if err == nil && status == 200 {
			return 200, nil
		}
	}
	return 404, nil
}

// ============================================================
// Gitea client
// ============================================================

type giteaClient struct {
	base   string
	token  string
	client *http.Client
	forge  Forge
}

func newGiteaClient() *giteaClient {
	return &giteaClient{
		base:   strings.TrimRight(os.Getenv("RR_VERIFY_GUARD_GITEA_URL"), "/"),
		token:  os.Getenv("RR_VERIFY_GUARD_GITEA_TOKEN"),
		client: &http.Client{Timeout: 10 * time.Second},
		forge:  ForgeGitea,
	}
}

func (c *giteaClient) Forge() Forge         { return c.forge }
func (c *giteaClient) BaseURL() string      { return c.base }
func (c *giteaClient) HasCredentials() bool { return c.base != "" && c.token != "" }

func (c *giteaClient) auth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/json")
}

func (c *giteaClient) listRunners(ctx context.Context, url string) ([]Runner, error) {
	if c.base == "" {
		return nil, errors.New("gitea base URL not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 || resp.StatusCode == 401 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gitea runners: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	// Gitea exposes either an array or an object with a runners field;
	// accept both.
	var arr []Runner
	if err := json.Unmarshal(body, &arr); err == nil && len(arr) >= 0 {
		return arr, nil
	}
	var wrapped struct {
		Runners []Runner `json:"runners"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Runners, nil
}

func (c *giteaClient) ListRepoRunners(ctx context.Context, owner, repo string) ([]Runner, error) {
	return c.listRunners(ctx, fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/runners", c.base, owner, repo))
}

func (c *giteaClient) ListOrgRunners(ctx context.Context, org string) ([]Runner, error) {
	return c.listRunners(ctx, fmt.Sprintf("%s/api/v1/orgs/%s/actions/runners", c.base, org))
}

func (c *giteaClient) HeadRepo(ctx context.Context, owner, repo string) (int, error) {
	return doHead(ctx, c.client, fmt.Sprintf("%s/api/v1/repos/%s/%s", c.base, owner, repo), c.auth)
}

func (c *giteaClient) HeadRef(ctx context.Context, owner, repo, ref string) (int, error) {
	tries := []string{
		fmt.Sprintf("%s/api/v1/repos/%s/%s/git/refs/heads/%s", c.base, owner, repo, ref),
		fmt.Sprintf("%s/api/v1/repos/%s/%s/git/refs/tags/%s", c.base, owner, repo, ref),
	}
	for _, u := range tries {
		status, err := doHead(ctx, c.client, u, c.auth)
		if err == nil && status == 200 {
			return 200, nil
		}
	}
	return 404, nil
}

// ============================================================
// Forgejo client (thin variant of giteaClient)
// ============================================================

func newForgejoClient() *giteaClient {
	c := &giteaClient{
		base:   strings.TrimRight(os.Getenv("RR_VERIFY_GUARD_FORGEJO_URL"), "/"),
		token:  os.Getenv("RR_VERIFY_GUARD_FORGEJO_TOKEN"),
		client: &http.Client{Timeout: 10 * time.Second},
		forge:  ForgeForgejo,
	}
	return c
}

// ============================================================
// shared helpers
// ============================================================

func doHead(ctx context.Context, client *http.Client, url string, authFn func(*http.Request)) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, err
	}
	if authFn != nil {
		authFn(req)
	}
	resp, err := client.Do(req)
	if err != nil {
		// HEAD may be rejected; try GET as fallback.
		gReq, gerr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if gerr != nil {
			return 0, err
		}
		if authFn != nil {
			authFn(gReq)
		}
		gResp, gerr := client.Do(gReq)
		if gerr != nil {
			return 0, err
		}
		defer gResp.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(gResp.Body, 4096))
		return gResp.StatusCode, nil
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
