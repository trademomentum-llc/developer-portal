// forge_clients_test.go -- request-shape verification for each forge
// client using net/http/httptest.
//
// Uses httptest.NewTLSServer (not NewServer) so the test server hands
// back https:// URLs and a TLS-aware client. The TLS-aware client is
// injected into the forge client under test via c.client = srv.Client()
// so cert verification works against the locally-generated cert.

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recordedReq captures a request for assertions.
type recordedReq struct {
	Method     string
	Path       string
	AuthHeader string
}

func makeServer(reqs *[]recordedReq, payload string, status int) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*reqs = append(*reqs, recordedReq{
			Method:     r.Method,
			Path:       r.URL.Path,
			AuthHeader: r.Header.Get("Authorization"),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != "" {
			// Decode the payload into a generic value, then re-encode
			// via the JSON encoder. This is structurally identical to
			// w.Write([]byte(payload)) but uses json.NewEncoder so the
			// data flow goes through encoding/json rather than a raw
			// byte write to the ResponseWriter.
			var v any
			if err := json.Unmarshal([]byte(payload), &v); err == nil {
				_ = json.NewEncoder(w).Encode(v)
			}
		}
	}))
}

func TestGitHubClient_AuthFallback(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"primary", map[string]string{"RR_VERIFY_GUARD_GITHUB_TOKEN": "rrgh"}, "Bearer rrgh"},
		{"fallback gh", map[string]string{"GITHUB_TOKEN": "ghtok"}, "Bearer ghtok"},
		{"fallback gh_token", map[string]string{"GH_TOKEN": "gh2"}, "Bearer gh2"},
		{"none", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("RR_VERIFY_GUARD_GITHUB_TOKEN", "")
			t.Setenv("GITHUB_TOKEN", "")
			t.Setenv("GH_TOKEN", "")
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			c2 := newGitHubClient()
			var captured string
			req, _ := http.NewRequest("GET", "https://x", nil)
			c2.auth(req)
			captured = req.Header.Get("Authorization")
			if captured != c.want {
				t.Errorf("env=%v -> auth=%q, want %q", c.env, captured, c.want)
			}
		})
	}
}

func TestGitHubClient_ListRepoRunnersShape(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RR_VERIFY_GUARD_GITHUB_TOKEN", "tok")
	var reqs []recordedReq
	body := `{"total_count":1,"runners":[{"id":1,"name":"r1","status":"online","labels":[{"name":"ubuntu-latest"}]}]}`
	srv := makeServer(&reqs, body, 200)
	defer srv.Close()

	c := newGitHubClient()
	c.base = srv.URL
	c.client = srv.Client()

	runners, err := c.ListRepoRunners(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(runners) != 1 || runners[0].Name != "r1" || runners[0].Status != "online" {
		t.Errorf("unexpected runners: %+v", runners)
	}
	if len(runners[0].Labels) != 1 || runners[0].Labels[0] != "ubuntu-latest" {
		t.Errorf("label parse wrong: %+v", runners[0])
	}
	if len(reqs) != 1 || reqs[0].Method != "GET" || !strings.Contains(reqs[0].Path, "/repos/owner/repo/actions/runners") {
		t.Errorf("unexpected request: %+v", reqs)
	}
	if reqs[0].AuthHeader != "Bearer tok" {
		t.Errorf("expected Bearer auth, got %q", reqs[0].AuthHeader)
	}
}

func TestGitHubClient_HeadRepoNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var reqs []recordedReq
	srv := makeServer(&reqs, "", 404)
	defer srv.Close()
	c := newGitHubClient()
	c.base = srv.URL
	c.client = srv.Client()
	status, err := c.HeadRepo(context.Background(), "no", "such")
	if err != nil {
		t.Fatal(err)
	}
	if status != 404 {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestGiteaClient_AuthHeader(t *testing.T) {
	t.Setenv("RR_VERIFY_GUARD_GITEA_URL", "https://gitea.example.com")
	t.Setenv("RR_VERIFY_GUARD_GITEA_TOKEN", "gitea-tok")
	c := newGiteaClient()
	if c.BaseURL() != "https://gitea.example.com" {
		t.Errorf("base = %q", c.BaseURL())
	}
	if !c.HasCredentials() {
		t.Error("HasCredentials should be true")
	}
	req, _ := http.NewRequest("GET", "https://x", nil)
	c.auth(req)
	if got := req.Header.Get("Authorization"); got != "token gitea-tok" {
		t.Errorf("auth header = %q, want token gitea-tok", got)
	}
}

func TestGiteaClient_ListRunnersArrayBody(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RR_VERIFY_GUARD_GITEA_TOKEN", "x")
	var reqs []recordedReq
	body, _ := json.Marshal([]Runner{{ID: 5, Name: "n", Status: "online", Labels: []string{"ubuntu-latest"}}})
	srv := makeServer(&reqs, string(body), 200)
	defer srv.Close()
	c := newGiteaClient()
	c.base = srv.URL
	c.client = srv.Client()
	runners, err := c.ListRepoRunners(context.Background(), "o", "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(runners) != 1 || runners[0].ID != 5 {
		t.Errorf("unexpected runners: %+v", runners)
	}
}

func TestGiteaClient_ListRunnersWrappedBody(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RR_VERIFY_GUARD_GITEA_TOKEN", "x")
	var reqs []recordedReq
	body := `{"runners":[{"id":7,"name":"w","status":"online","labels":["x"]}]}`
	srv := makeServer(&reqs, body, 200)
	defer srv.Close()
	c := newGiteaClient()
	c.base = srv.URL
	c.client = srv.Client()
	runners, _ := c.ListRepoRunners(context.Background(), "o", "r")
	if len(runners) != 1 || runners[0].ID != 7 {
		t.Errorf("wrapped body decode failed: %+v", runners)
	}
}

func TestGiteaClient_NoBaseURL(t *testing.T) {
	t.Setenv("RR_VERIFY_GUARD_GITEA_URL", "")
	t.Setenv("RR_VERIFY_GUARD_GITEA_TOKEN", "")
	c := newGiteaClient()
	if c.HasCredentials() {
		t.Error("HasCredentials should be false without base URL")
	}
	if _, err := c.ListRepoRunners(context.Background(), "o", "r"); err == nil {
		t.Error("expected error when base URL missing")
	}
}

func TestForgejoClient_DistinctForgeID(t *testing.T) {
	t.Setenv("RR_VERIFY_GUARD_FORGEJO_URL", "https://forgejo.example.com")
	t.Setenv("RR_VERIFY_GUARD_FORGEJO_TOKEN", "fj-tok")
	c := newForgejoClient()
	if c.Forge() != ForgeForgejo {
		t.Errorf("forgejo client should report ForgeForgejo, got %v", c.Forge())
	}
}

func TestNewForgeClient_DispatchByForge(t *testing.T) {
	t.Setenv("RR_VERIFY_GUARD_GITHUB_TOKEN", "gh")
	t.Setenv("RR_VERIFY_GUARD_GITEA_URL", "https://gitea.example.com")
	t.Setenv("RR_VERIFY_GUARD_GITEA_TOKEN", "g")
	t.Setenv("RR_VERIFY_GUARD_FORGEJO_URL", "https://forgejo.example.com")
	t.Setenv("RR_VERIFY_GUARD_FORGEJO_TOKEN", "fj")

	if c := NewForgeClient(ForgeGitHub); c == nil || c.Forge() != ForgeGitHub {
		t.Errorf("github dispatch wrong: %v", c)
	}
	if c := NewForgeClient(ForgeGitea); c == nil || c.Forge() != ForgeGitea {
		t.Errorf("gitea dispatch wrong: %v", c)
	}
	if c := NewForgeClient(ForgeForgejo); c == nil || c.Forge() != ForgeForgejo {
		t.Errorf("forgejo dispatch wrong: %v", c)
	}
	if c := NewForgeClient(ForgeNone); c != nil {
		t.Errorf("None should yield nil, got %v", c)
	}
	if c := NewForgeClient(ForgeUnknown); c != nil {
		t.Errorf("Unknown should yield nil, got %v", c)
	}
}
