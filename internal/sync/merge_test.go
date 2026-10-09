package sync

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func TestBuildGroupsProductsAndClassifiesRegistry(t *testing.T) {
	manifest := catalog.Manifest{
		Owner: "rekurt",
		Products: []catalog.ProductConfig{
			{
				Slug: "tool", PrimaryRepo: "rekurt/tool",
				Repositories: []string{"rekurt/tool", "rekurt/tap"}, Kind: "cli", Domain: "developer-tools", Accent: "cyan",
				Summary: catalog.LocalizedText{EN: "Tool summary", RU: "Описание инструмента", ZHCN: "工具简介"},
			},
			{
				Slug: "forked", PrimaryRepo: "rekurt/forked", Repositories: []string{"rekurt/forked"},
				Kind: "desktop", Domain: "developer-tools", Featured: true, MaintainedFork: true,
				Upstream: "upstream/forked", Website: "https://rekurt.github.io/forked/", Accent: "amber",
				Summary: catalog.LocalizedText{EN: "Maintained fork", RU: "Поддерживаемый форк", ZHCN: "维护分支"},
			},
		},
	}
	repositories := []catalog.Repository{
		repository("rekurt/mirror", true, "upstream/mirror"),
		repository("rekurt/tap", false, ""),
		repository("rekurt/forked", true, "upstream/forked"),
		repository("rekurt/tool", false, ""),
		repository("rekurt/rekurt.github.io", false, ""),
	}
	repositories[0].Homepage = "https://upstream.example"
	repositories[2].Version = &catalog.Version{Value: "v2.0.0", Source: "release", URL: "https://github.com/rekurt/forked/releases/tag/v2.0.0"}
	repositories[3].HasPages = true
	repositories[3].Readme = &catalog.Readme{Source: "# Tool\n\n<script>bad()</script>Useful.", SHA: "abc", SourceURL: "https://github.com/rekurt/tool/blob/abc/README.md"}

	snapshot, err := Build(manifest, repositories, time.Date(2026, 9, 5, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.SchemaVersion != 1 || snapshot.Owner != "rekurt" || len(snapshot.Repositories) != 5 {
		t.Fatalf("snapshot metadata = %#v", snapshot)
	}
	if got := []string{snapshot.Products[0].Slug, snapshot.Products[1].Slug}; !slices.Equal(got, []string{"forked", "tool"}) {
		t.Fatalf("product order = %#v", got)
	}
	roles := make(map[string]string)
	for _, repo := range snapshot.Repositories {
		roles[repo.NameWithOwner] = repo.Role
		if repo.NameWithOwner == "rekurt/mirror" && linkByKind(repo.Links, "website") != "" {
			t.Fatalf("upstream homepage attributed as author website: %#v", repo.Links)
		}
	}
	wantRoles := map[string]string{
		"rekurt/forked":           "maintained-fork",
		"rekurt/mirror":           "fork",
		"rekurt/rekurt.github.io": "portfolio-hub",
		"rekurt/tap":              "support",
		"rekurt/tool":             "primary",
	}
	for name, want := range wantRoles {
		if roles[name] != want {
			t.Fatalf("role %s = %q, want %q", name, roles[name], want)
		}
	}
	tool := snapshot.Products[1]
	if tool.Accent != "cyan" || tool.Summary.ZHCN != "工具简介" {
		t.Fatalf("localized presentation metadata = %#v", tool)
	}
	if linkByKind(tool.Links, "website") != "https://rekurt.github.io/tool/" {
		t.Fatalf("derived Pages URL missing: %#v", tool.Links)
	}
	if tool.Readme == nil || strings.Contains(tool.Readme.HTML, "bad()") || !strings.Contains(tool.Readme.HTML, "Useful") {
		t.Fatalf("README was not safely rendered: %#v", tool.Readme)
	}
	if snapshot.Products[0].Upstream != "upstream/forked" || linkByKind(snapshot.Products[0].Links, "release") == "" {
		t.Fatalf("maintained fork metadata = %#v", snapshot.Products[0])
	}
}

func TestBuildRejectsUnknownAndPrivateRepositories(t *testing.T) {
	base := catalog.Manifest{Owner: "rekurt", Products: []catalog.ProductConfig{{
		Slug: "tool", PrimaryRepo: "rekurt/tool", Repositories: []string{"rekurt/tool"},
		Kind: "cli", Domain: "developer-tools", Accent: "cyan", Summary: catalog.LocalizedText{EN: "Tool", RU: "Инструмент", ZHCN: "工具"},
	}}}
	tests := []struct {
		name  string
		repos []catalog.Repository
		want  string
	}{
		{name: "unknown linked repository", repos: nil, want: "references unknown repository rekurt/tool"},
		{name: "private repository", repos: []catalog.Repository{{NameWithOwner: "rekurt/tool", Visibility: "private"}}, want: "private repository is forbidden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Build(base, tt.repos, time.Now())
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func repository(name string, fork bool, parent string) catalog.Repository {
	short := name[strings.IndexByte(name, '/')+1:]
	return catalog.Repository{
		NameWithOwner: name, Name: short, URL: "https://github.com/" + name, Visibility: "public",
		Fork: fork, Parent: parent, DefaultBranch: "main", UpdatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		PushedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}
}

func linkByKind(links []catalog.Link, kind string) string {
	for _, link := range links {
		if link.Kind == kind {
			return link.URL
		}
	}
	return ""
}
