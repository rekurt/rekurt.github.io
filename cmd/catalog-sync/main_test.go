package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	catalogsync "github.com/rekurt/rekurt.github.io/internal/sync"
)

func TestRunSyncWritesSnapshotAndAudit(t *testing.T) {
	server, stars := githubServer(t)
	t.Cleanup(server.Close)
	directory := t.TempDir()
	manifestPath := writeManifest(t, directory)
	snapshotPath := filepath.Join(directory, "catalog.json")
	auditPath := filepath.Join(directory, "audit.md")
	getenv := testEnvironment(server.URL)

	var output strings.Builder
	err := run(t.Context(), []string{"sync", "--manifest", manifestPath, "--snapshot", snapshotPath, "--audit", auditPath}, getenv, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := catalogsync.ReadSnapshot(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Products) != 1 || len(snapshot.Repositories) != 1 || snapshot.Repositories[0].Stars != 3 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	audit, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(audit), "rekurt/tool") || !strings.Contains(output.String(), "1 products, 1 repositories") {
		t.Fatalf("audit/output = %s\n%s", audit, output.String())
	}

	if err := run(t.Context(), []string{"check", "--manifest", manifestPath, "--snapshot", snapshotPath, "--audit", auditPath}, getenv, &output, &output); err != nil {
		t.Fatalf("unchanged check failed: %v", err)
	}
	stars.Store(4)
	err = run(t.Context(), []string{"check", "--manifest", manifestPath, "--snapshot", snapshotPath, "--audit", auditPath}, getenv, &output, &output)
	if err == nil || !strings.Contains(err.Error(), "catalog snapshot is out of date") {
		t.Fatalf("changed check error = %v", err)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var output strings.Builder
	err := run(t.Context(), []string{"unknown"}, func(string) string { return "" }, &output, &output)
	if err == nil || !strings.Contains(err.Error(), "unknown command") || !strings.Contains(output.String(), "catalog-sync <sync|check>") {
		t.Fatalf("error/output = %v / %q", err, output.String())
	}
}

func githubServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	stars := &atomic.Int32{}
	stars.Store(3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := `{"name":"tool","full_name":"rekurt/tool","html_url":"https://github.com/rekurt/tool","description":"Useful tool","private":false,"visibility":"public","fork":false,"language":"Go","topics":["tooling"],"homepage":"https://rekurt.github.io/tool/","has_pages":true,"default_branch":"main","license":{"spdx_id":"MIT"},"updated_at":"2026-09-01T10:00:00Z","pushed_at":"2026-09-01T09:00:00Z","archived":false,"stargazers_count":%d}`
		switch r.URL.Path {
		case "/users/rekurt/repos":
			fmt.Fprintf(w, "["+base+"]", stars.Load())
		case "/repos/rekurt/tool":
			fmt.Fprintf(w, base, stars.Load())
		case "/repos/rekurt/tool/releases":
			fmt.Fprint(w, `[{"tag_name":"v1.0.0","html_url":"https://github.com/rekurt/tool/releases/tag/v1.0.0","draft":false,"prerelease":false,"published_at":"2026-09-01T10:00:00Z"}]`)
		case "/repos/rekurt/tool/branches/main":
			fmt.Fprint(w, `{"commit":{"sha":"head-123"}}`)
		case "/repos/rekurt/tool/readme":
			content := base64.StdEncoding.EncodeToString([]byte("# Tool\n\nUseful tool."))
			fmt.Fprintf(w, `{"encoding":"base64","content":%q,"sha":"blob-123","html_url":"https://github.com/rekurt/tool/blob/main/README.md"}`, content)
		default:
			http.NotFound(w, r)
		}
	}))
	return server, stars
}

func writeManifest(t *testing.T, directory string) string {
	t.Helper()
	manifest := `owner: rekurt
products:
  - slug: tool
    primary_repo: rekurt/tool
    repositories: [rekurt/tool]
    kind: cli
    domain: developer-tools
    accent: cyan
    featured: true
    maintained_fork: false
    upstream: ""
    summary:
      en: Useful tool.
      ru: Полезный инструмент.
      zh-cn: 实用工具。
    install: [brew install tool]
    website: https://rekurt.github.io/tool/
    documentation: https://github.com/rekurt/tool#readme
`
	path := filepath.Join(directory, "projects.yaml")
	if err := os.WriteFile(path, []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func testEnvironment(apiURL string) func(string) string {
	return func(name string) string {
		switch name {
		case "GITHUB_API_URL":
			return apiURL
		case "GITHUB_TOKEN":
			return "test-token"
		default:
			return ""
		}
	}
}
