package projectsite

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBuildsLocalizedProjectModel(t *testing.T) {
	repository, err := filepath.Abs("testdata/repository")
	if err != nil {
		t.Fatal(err)
	}
	model, err := Resolve(Options{
		Slug:         "git-barber",
		SnapshotPath: "testdata/snapshot.json",
		Repository:   repository,
		Output:       filepath.Join(t.TempDir(), "site"),
		BaseURL:      "https://rekurt.github.io/git-barber",
	})
	if err != nil {
		t.Fatal(err)
	}
	if model.Owner != "rekurt" || model.Product.Slug != "git-barber" || model.Repository.HeadSHA != "abc123" {
		t.Fatalf("resolved identity = %#v", model)
	}
	if model.BaseURL != "https://rekurt.github.io/git-barber/" {
		t.Fatalf("BaseURL = %q", model.BaseURL)
	}
	if len(model.Pages) != 3 {
		t.Fatalf("pages = %d, want 3", len(model.Pages))
	}

	english := localePage(t, model, "en")
	if english.Path != "/" || english.Lang != "en" || english.Canonical != model.BaseURL {
		t.Fatalf("English route = %#v", english)
	}
	if !strings.Contains(english.ReadmeHTML, "English documentation") || strings.Contains(english.ReadmeHTML, "alert(1)") {
		t.Fatalf("English README was not selected and sanitized: %s", english.ReadmeHTML)
	}
	if !strings.Contains(english.ReadmeHTML, "https://github.com/rekurt/git-barber/blob/abc123/docs/usage.md") {
		t.Fatalf("English README link is not pinned: %s", english.ReadmeHTML)
	}

	russian := localePage(t, model, "ru")
	if russian.Path != "/ru/" || russian.Lang != "ru" || !strings.Contains(russian.ReadmeHTML, "Русская документация") {
		t.Fatalf("Russian route = %#v", russian)
	}

	chinese := localePage(t, model, "zh-cn")
	if chinese.Path != "/zh-cn/" || chinese.Lang != "zh-CN" || chinese.Summary != "安全清理 Git 分支。" {
		t.Fatalf("Chinese route = %#v", chinese)
	}
	if chinese.ReadmeHTML != "" {
		t.Fatalf("missing Chinese README must not fall back to English HTML: %q", chinese.ReadmeHTML)
	}
}

func TestResolveRejectsInvalidBoundaries(t *testing.T) {
	repository, err := filepath.Abs("testdata/repository")
	if err != nil {
		t.Fatal(err)
	}
	base := Options{
		Slug:         "git-barber",
		SnapshotPath: "testdata/snapshot.json",
		Repository:   repository,
		Output:       filepath.Join(t.TempDir(), "site"),
		BaseURL:      "https://rekurt.github.io/git-barber/",
	}
	tests := []struct {
		name   string
		mutate func(*Options)
		want   string
	}{
		{name: "unknown product", mutate: func(o *Options) { o.Slug = "unknown" }, want: "unknown product"},
		{name: "insecure URL", mutate: func(o *Options) { o.BaseURL = "http://rekurt.github.io/git-barber/" }, want: "base URL must use https"},
		{name: "URL query", mutate: func(o *Options) { o.BaseURL += "?preview=1" }, want: "base URL must not contain query or fragment"},
		{name: "same repository and output", mutate: func(o *Options) { o.Output = o.Repository }, want: "repository and output must differ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := base
			tt.mutate(&options)
			_, err := Resolve(options)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func localePage(t *testing.T, model Model, locale string) LocalePage {
	t.Helper()
	for _, page := range model.Pages {
		if page.Locale == locale {
			return page
		}
	}
	t.Fatalf("locale %s not found", locale)
	return LocalePage{}
}
