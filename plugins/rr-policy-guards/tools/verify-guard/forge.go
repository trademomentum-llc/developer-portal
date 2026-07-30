// forge.go -- forge detection from the `origin` remote URL.
//
// Three URL forms recognised:
//   https://host/owner/repo[.git]
//   git@host:owner/repo[.git]
//   ssh://git@host[:port]/owner/repo[.git]
//
// host -> forge mapping:
//   github.com or *.github.com -> github
//   anything else: probe https://<host>/api/v1/version
//                  body containing "+gitea-" -> forgejo
//                  body matching {"version":"1.x..."} -> gitea
//                  otherwise -> unknown

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RemoteInfo captures everything the guard needs about a repo's origin.
type RemoteInfo struct {
	Forge  Forge
	Host   string
	Owner  string
	Repo   string
	Scheme string // "https" | "ssh"
	URL    string
}

// DetectForge reads the origin remote of repoRoot and classifies it.
// Returns (RemoteInfo{Forge: ForgeNone}, nil) when there is no remote.
func DetectForge(repoRoot string) (RemoteInfo, error) {
	url, err := readRemoteURL(repoRoot, "origin")
	if err != nil {
		// Try first lexical remote.
		if other, _ := firstRemoteURL(repoRoot); other != "" {
			url = other
		} else {
			return RemoteInfo{Forge: ForgeNone}, nil
		}
	}
	host, owner, repo, scheme, err := parseRemoteURL(url)
	if err != nil {
		return RemoteInfo{Forge: ForgeUnknown, URL: url}, nil
	}
	info := RemoteInfo{Host: host, Owner: owner, Repo: repo, Scheme: scheme, URL: url}
	if host == "github.com" || strings.HasSuffix(host, ".github.com") {
		info.Forge = ForgeGitHub
		return info, nil
	}
	if cached, ok := readForgeCache(host); ok {
		info.Forge = cached
		return info, nil
	}
	probed := probeForgeVersion(host, 5*time.Second)
	info.Forge = probed
	_ = writeForgeCache(host, probed)
	return info, nil
}

// readRemoteURL invokes git to read the URL of a named remote.
func readRemoteURL(repoRoot, name string) (string, error) {
	out, err := exec.Command("git", "-C", repoRoot, "remote", "get-url", name).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// firstRemoteURL returns the URL of the first remote in lexical order.
func firstRemoteURL(repoRoot string) (string, error) {
	out, err := exec.Command("git", "-C", repoRoot, "remote").CombinedOutput()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			return readRemoteURL(repoRoot, name)
		}
	}
	return "", errors.New("no remotes")
}

// parseRemoteURL extracts host, owner, repo, scheme from any of the
// three supported remote URL forms.
func parseRemoteURL(url string) (host, owner, repo, scheme string, err error) {
	u := strings.TrimSpace(url)
	switch {
	case strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://"):
		scheme = "https"
		rest := strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
		parts := strings.SplitN(rest, "/", 3)
		if len(parts) < 3 {
			return "", "", "", "", errors.New("malformed https remote")
		}
		host = parts[0]
		owner = parts[1]
		repo = strings.TrimSuffix(parts[2], ".git")
		return
	case strings.HasPrefix(u, "ssh://"):
		scheme = "ssh"
		rest := strings.TrimPrefix(u, "ssh://")
		if at := strings.Index(rest, "@"); at >= 0 {
			rest = rest[at+1:]
		}
		// rest = host[:port]/owner/repo
		parts := strings.SplitN(rest, "/", 3)
		if len(parts) < 3 {
			return "", "", "", "", errors.New("malformed ssh remote")
		}
		host = parts[0]
		if colon := strings.Index(host, ":"); colon >= 0 {
			host = host[:colon]
		}
		owner = parts[1]
		repo = strings.TrimSuffix(parts[2], ".git")
		return
	default:
		// git@host:owner/repo form
		if at := strings.Index(u, "@"); at >= 0 {
			scheme = "ssh"
			rest := u[at+1:]
			colon := strings.Index(rest, ":")
			if colon < 0 {
				return "", "", "", "", errors.New("malformed scp-style remote")
			}
			host = rest[:colon]
			path := rest[colon+1:]
			parts := strings.SplitN(path, "/", 2)
			if len(parts) < 2 {
				return "", "", "", "", errors.New("malformed scp-style path")
			}
			owner = parts[0]
			repo = strings.TrimSuffix(parts[1], ".git")
			return
		}
	}
	return "", "", "", "", errors.New("unrecognised remote URL form")
}

// probeForgeVersion calls https://<host>/api/v1/version with a hard
// timeout and classifies the response.
func probeForgeVersion(host string, timeout time.Duration) Forge {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/api/v1/version", nil)
	if err != nil {
		return ForgeUnknown
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return ForgeUnknown
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ForgeUnknown
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return ForgeUnknown
	}
	var v struct {
		Version string `json:"version"`
	}
	if jerr := json.Unmarshal(body, &v); jerr == nil {
		if strings.Contains(v.Version, "+gitea-") {
			return ForgeForgejo
		}
		if v.Version != "" {
			return ForgeGitea
		}
	}
	return ForgeUnknown
}

// readForgeCache returns the cached forge for host if the cache file is
// present and not older than 24 hours.
func readForgeCache(host string) (Forge, bool) {
	dir, err := forgeCacheDir()
	if err != nil {
		return "", false
	}
	path := filepath.Join(dir, host+".json")
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	if time.Since(info.ModTime()) > 24*time.Hour {
		return "", false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var rec struct {
		Forge Forge `json:"forge"`
	}
	if err := json.Unmarshal(b, &rec); err != nil {
		return "", false
	}
	if rec.Forge == "" {
		return "", false
	}
	return rec.Forge, true
}

// writeForgeCache persists the detection result for host.
func writeForgeCache(host string, f Forge) error {
	dir, err := forgeCacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	rec := struct {
		Forge Forge  `json:"forge"`
		Host  string `json:"host"`
		Ts    string `json:"ts"`
	}{f, host, time.Now().UTC().Format(time.RFC3339)}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, host+".json"), b, 0600)
}

func forgeCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".rational-reserve", "cache", "verify-guard", "forge"), nil
}
