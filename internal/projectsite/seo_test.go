package projectsite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructuredDataUsesOnlyCatalogEvidence(t *testing.T) {
	model, err := Resolve(fixtureOptions(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	data, err := structuredData(model, localePage(t, model, "en"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"@context": "https://schema.org", "@type": "SoftwareSourceCode", "name": "git-barber",
		"codeRepository": "https://github.com/rekurt/git-barber", "programmingLanguage": "Rust",
		"license": "MIT OR Apache-2.0", "version": "v0.3.0",
	}
	for key, value := range want {
		if schema[key] != value {
			t.Errorf("schema[%q] = %#v, want %q", key, schema[key], value)
		}
	}
	encoded := string(data)
	for _, forbidden := range []string{"best", "production-ready", "secure", "fastest"} {
		if strings.Contains(strings.ToLower(encoded), forbidden) {
			t.Errorf("structured data contains unsupported claim %q: %s", forbidden, encoded)
		}
	}
}

func TestSEOTitlesCoverEveryProductAndLocale(t *testing.T) {
	var titles map[string]map[string]string
	if err := json.Unmarshal(seoTitleData, &titles); err != nil {
		t.Fatal(err)
	}
	if len(titles) != 13 {
		t.Fatalf("SEO titles: %d products", len(titles))
	}
	for slug := range titles {
		for _, locale := range []string{"en", "ru", "zh-cn"} {
			if title := seoTitle(slug, locale, "fallback"); title == "fallback" || len(title) < 15 {
				t.Errorf("missing descriptive title: %s/%s", slug, locale)
			}
		}
	}
}

func TestDecoratedSitemapDoesNotAdvertiseMissingTranslatedHomes(t *testing.T) {
	output := copyExistingFixture(t)
	model, err := Resolve(fixtureOptions(t, output))
	if err != nil {
		t.Fatal(err)
	}
	sitemap := renderDecoratedSitemap(model)
	for _, missing := range []string{model.BaseURL + "ru/", model.BaseURL + "zh-cn/"} {
		if strings.Contains(sitemap, "<loc>"+missing+"</loc>") || strings.Contains(sitemap, `href="`+missing+`"`) {
			t.Errorf("missing page in sitemap: %s", missing)
		}
	}
	if !strings.Contains(sitemap, "<loc>"+model.BaseURL+"ru/projects/</loc>") {
		t.Fatal("localized directory missing")
	}
	if err := os.MkdirAll(filepath.Join(output, "ru"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, "ru/index.html"), []byte("Russian app"), 0644); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderDecoratedSitemap(model), "<loc>"+model.BaseURL+"ru/</loc>") {
		t.Fatal("existing localized homepage excluded")
	}
}

func TestReadmeHeadingHierarchyKeepsLinksAndAttributes(t *testing.T) {
	source := `<h1 id="project">Name</h1><h2>Install</h2><h6>Note</h6><a href="#project">Jump</a>`
	want := `<h3 id="project">Name</h3><h4>Install</h4><h6>Note</h6><a href="#project">Jump</a>`
	if got := nestedReadmeHTML(source); got != want {
		t.Fatalf("nested README = %s", got)
	}
}
