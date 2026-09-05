package projectsite

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestDecoratePreservesApplicationAndAddsFamilyLayer(t *testing.T) {
	output := copyExistingFixture(t)
	before := elementHTMLByID(t, readFile(t, filepath.Join(output, "index.html")), "application")
	options := fixtureOptions(t, output)
	manifest, err := Decorate(options)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != "decorate" || manifest.SourceSHA != "abc123" {
		t.Fatalf("manifest = %#v", manifest)
	}
	afterHTML := readFile(t, filepath.Join(output, "index.html"))
	after := elementHTMLByID(t, afterHTML, "application")
	if before != after {
		t.Fatalf("application subtree changed:\nBEFORE %s\nAFTER  %s", before, after)
	}
	for _, required := range []string{
		`data-rekurt-family=""`, `class="rk-family-bar"`, `href="https://rekurt.github.io/"`,
		`href="projects/"`, `href="assets/bridge.css"`,
		`rel="canonical" href="https://rekurt.github.io/git-barber/"`,
		`application/ld+json`, `assets/original-social.png`,
	} {
		if !strings.Contains(afterHTML, required) {
			t.Errorf("decorated page lacks %q", required)
		}
	}
	for _, file := range []string{"assets/bridge.css", "assets/family.css", "family-build.json", "family-sitemap.xml", "projects/index.html", "ru/projects/index.html", "zh-cn/projects/index.html"} {
		if _, err := os.Stat(filepath.Join(output, file)); err != nil {
			t.Errorf("generated %s: %v", file, err)
		}
	}
}

func TestDecorateIsIdempotent(t *testing.T) {
	output := copyExistingFixture(t)
	options := fixtureOptions(t, output)
	if _, err := Decorate(options); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, filepath.Join(output, "index.html"))
	if _, err := Decorate(options); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, filepath.Join(output, "index.html"))
	if first != second {
		index := 0
		for index < len(first) && index < len(second) && first[index] == second[index] {
			index++
		}
		start := max(0, index-120)
		endFirst := min(len(first), index+240)
		endSecond := min(len(second), index+240)
		t.Fatalf("second decoration changed index.html at byte %d:\nFIRST  %q\nSECOND %q", index, first[start:endFirst], second[start:endSecond])
	}
	if strings.Count(second, `data-rekurt-family=""`) != 2 {
		t.Fatalf("family markers = %d, want stylesheet and bar", strings.Count(second, `data-rekurt-family=""`))
	}
	if strings.Count(second, `rel="canonical"`) != 1 {
		t.Fatalf("canonical links = %d", strings.Count(second, `rel="canonical"`))
	}
}

func TestDecorateRejectsUnsafeOrIncompleteOutput(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  string
	}{
		{
			name: "symlink root",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				path := filepath.Join(t.TempDir(), "site")
				if err := os.Symlink(root, path); err != nil {
					t.Fatal(err)
				}
				return path
			},
			want: "output must not be a symlink",
		},
		{
			name:  "missing index",
			setup: func(t *testing.T) string { return t.TempDir() },
			want:  "existing site index.html is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := fixtureOptions(t, tt.setup(t))
			_, err := Decorate(options)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func copyExistingFixture(t *testing.T) string {
	t.Helper()
	output := t.TempDir()
	source, err := os.ReadFile("testdata/existing/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, "index.html"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, "robots.txt"), []byte("User-agent: *\nAllow: /\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return output
}

func elementHTMLByID(t *testing.T, source, id string) string {
	t.Helper()
	document, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		for _, attribute := range node.Attr {
			if attribute.Key == "id" && attribute.Val == id {
				found = node
				return
			}
		}
		for child := node.FirstChild; child != nil && found == nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(document)
	if found == nil {
		t.Fatalf("element #%s not found", id)
	}
	var output bytes.Buffer
	if err := html.Render(&output, found); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func decorateGeneratedAt() time.Time {
	return time.Date(2026, 9, 5, 6, 30, 0, 0, time.UTC)
}
