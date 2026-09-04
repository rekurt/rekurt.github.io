package githubapi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func TestListOwnedPublicPaginatesFiltersAndSorts(t *testing.T) {
	page1 := mustFixture(t, "repos-page-1.json")
	page2 := mustFixture(t, "repos-page-2.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") == "" || r.Header.Get("X-GitHub-Api-Version") == "" {
			t.Fatal("GitHub version headers are required")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q", got)
		}
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/users/rekurt/repos?page=2>; rel="next"`, serverURL(r)))
			writeJSON(t, w, page1)
		case "2":
			writeJSON(t, w, page2)
		default:
			t.Fatalf("page = %q", r.URL.Query().Get("page"))
		}
	}))
	t.Cleanup(server.Close)

	got, err := New(server.URL, "test-token", server.Client()).ListOwnedPublic(t.Context(), "rekurt")
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(got))
	for _, repo := range got {
		names = append(names, repo.NameWithOwner)
		if repo.Visibility != "public" {
			t.Fatalf("private repository leaked: %#v", repo)
		}
	}
	want := []string{"rekurt/alpha", "rekurt/beta", "rekurt/zeta"}
	if !slices.Equal(names, want) {
		t.Fatalf("names = %#v, want %#v", names, want)
	}
	if !got[0].Fork || !got[0].HasPages || got[2].License != "MIT" {
		t.Fatalf("mapped metadata = %#v", got)
	}
}

func TestEnrichUsesStableReleaseAndPinsReadme(t *testing.T) {
	detail := mustFixture(t, "repo-mac-coffee.json")
	releases := mustFixture(t, "release-ymsdk.json")
	readmeSource := "# Mac Coffee\n\nSafe wake sessions.\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/rekurt/Mac-Coffee":
			writeJSON(t, w, detail)
		case "/repos/rekurt/Mac-Coffee/releases":
			writeJSON(t, w, releases)
		case "/repos/rekurt/Mac-Coffee/branches/main":
			writeJSON(t, w, []byte(`{"commit":{"sha":"commit-abc123"}}`))
		case "/repos/rekurt/Mac-Coffee/readme":
			payload := fmt.Sprintf(`{"encoding":"base64","content":%q,"sha":"blob-def456","html_url":"https://github.com/rekurt/Mac-Coffee/blob/main/README.md"}`, base64.StdEncoding.EncodeToString([]byte(readmeSource)))
			writeJSON(t, w, []byte(payload))
		case "/repos/rekurt/Mac-Coffee/tags":
			t.Fatal("tags fallback must not run after a stable release")
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	got, err := New(server.URL, "", server.Client()).Enrich(t.Context(), catalog.Repository{NameWithOwner: "rekurt/Mac-Coffee"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Parent != "Elliotwu-7/Mac-Coffee" || got.HeadSHA != "commit-abc123" {
		t.Fatalf("repository = %#v", got)
	}
	if got.Version == nil || got.Version.Value != "v0.2.0" || got.Version.Source != "release" {
		t.Fatalf("version = %#v", got.Version)
	}
	if got.Readme == nil || got.Readme.Source != readmeSource || got.Readme.SHA != "commit-abc123" {
		t.Fatalf("readme = %#v", got.Readme)
	}
}

func TestEnrichFallsBackToSemverTag(t *testing.T) {
	server := enrichmentServer(t, "dbdiff", mustFixture(t, "tags-dbdiff.json"), nil)
	t.Cleanup(server.Close)

	got, err := New(server.URL, "", server.Client()).Enrich(t.Context(), catalog.Repository{NameWithOwner: "rekurt/dbdiff"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version == nil || got.Version.Value != "v0.2.1" || got.Version.Source != "tag" {
		t.Fatalf("version = %#v", got.Version)
	}
}

func TestEnrichFallsBackToVersionedManifest(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		files      map[string]string
		want       string
	}{
		{name: "cargo", repository: "rust-tool", files: map[string]string{"Cargo.toml": "[package]\nname = \"rust-tool\"\nversion = \"1.4.2\"\n"}, want: "1.4.2"},
		{name: "npm", repository: "ts-tool", files: map[string]string{"package.json": `{"name":"ts-tool","version":"2.3.4"}`}, want: "2.3.4"},
		{name: "go has no manifest version", repository: "go-tool", files: map[string]string{"go.mod": "module github.com/rekurt/go-tool\n"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := enrichmentServer(t, tt.repository, []byte("[]"), tt.files)
			t.Cleanup(server.Close)
			got, err := New(server.URL, "", server.Client()).Enrich(t.Context(), catalog.Repository{NameWithOwner: "rekurt/" + tt.repository}, false)
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == "" {
				if got.Version != nil {
					t.Fatalf("version = %#v, want nil", got.Version)
				}
				return
			}
			if got.Version == nil || got.Version.Value != tt.want || got.Version.Source != "manifest" {
				t.Fatalf("version = %#v, want manifest %s", got.Version, tt.want)
			}
		})
	}
}

func enrichmentServer(t *testing.T, name string, tags []byte, files map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "/repos/rekurt/" + name
		switch r.URL.Path {
		case base:
			detail := fmt.Sprintf(`{"name":%q,"full_name":%q,"html_url":%q,"private":false,"visibility":"public","fork":false,"language":"Rust","topics":[],"homepage":"","has_pages":false,"default_branch":"main","license":{"spdx_id":"MIT"},"updated_at":"2026-09-01T10:00:00Z","pushed_at":"2026-09-01T09:00:00Z","archived":false,"stargazers_count":0}`, name, "rekurt/"+name, "https://github.com/rekurt/"+name)
			writeJSON(t, w, []byte(detail))
		case base + "/releases":
			writeJSON(t, w, []byte("[]"))
		case base + "/tags":
			writeJSON(t, w, tags)
		case base + "/branches/main":
			writeJSON(t, w, []byte(`{"commit":{"sha":"head-sha"}}`))
		default:
			prefix := base + "/contents/"
			if strings.HasPrefix(r.URL.Path, prefix) {
				filename := strings.TrimPrefix(r.URL.Path, prefix)
				content, ok := files[filename]
				if !ok || filename == "go.mod" {
					http.NotFound(w, r)
					return
				}
				payload := fmt.Sprintf(`{"encoding":"base64","content":%q,"html_url":%q}`, base64.StdEncoding.EncodeToString([]byte(content)), "https://github.com/rekurt/"+name+"/blob/main/"+filename)
				writeJSON(t, w, []byte(payload))
				return
			}
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
}

func mustFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "github", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeJSON(t *testing.T, w http.ResponseWriter, data []byte) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
}

func serverURL(r *http.Request) string {
	return "http://" + r.Host
}
