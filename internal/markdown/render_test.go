package markdown

import (
	"strings"
	"testing"
)

func TestRenderREADMERewritesAndSanitizes(t *testing.T) {
	source := []byte("[Guide](docs/guide.md) ![Shot](assets/a.png) <script>alert(1)</script> <iframe src=\"https://evil.example\"></iframe> [x](javascript:alert(1))")
	got, err := RenderREADME(source, "rekurt/demo", "main", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"<script", "<iframe", "javascript:", "alert(1)"} {
		if strings.Contains(got.HTML, forbidden) {
			t.Fatalf("HTML contains %q: %s", forbidden, got.HTML)
		}
	}
	if !strings.Contains(got.HTML, "https://github.com/rekurt/demo/blob/abc123/docs/guide.md") {
		t.Fatalf("document link was not pinned: %s", got.HTML)
	}
	if !strings.Contains(got.HTML, "https://raw.githubusercontent.com/rekurt/demo/abc123/assets/a.png") {
		t.Fatalf("image link was not pinned: %s", got.HTML)
	}
	if got.SourceURL != "https://github.com/rekurt/demo/blob/abc123/README.md" || got.SHA != "abc123" {
		t.Fatalf("readme metadata = %#v", got)
	}
}

func TestRenderREADMELinkPolicy(t *testing.T) {
	source := []byte("[Section](#install) [Site](https://example.com/docs) [Mail](mailto:dev@example.com)")
	got, err := RenderREADME(source, "rekurt/demo", "main", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.HTML, `href="#install"`) {
		t.Fatalf("fragment changed: %s", got.HTML)
	}
	if !strings.Contains(got.HTML, `href="mailto:dev@example.com"`) {
		t.Fatalf("mailto removed: %s", got.HTML)
	}
	if !strings.Contains(got.HTML, "noopener") || !strings.Contains(got.HTML, "noreferrer") {
		t.Fatalf("external rel missing: %s", got.HTML)
	}
}

func TestRenderREADMERejectsOversizedSource(t *testing.T) {
	_, err := RenderREADME(make([]byte, maxSourceBytes+1), "rekurt/demo", "main", "abc123")
	if err == nil || !strings.Contains(err.Error(), "exceeds 262144 bytes") {
		t.Fatalf("error = %v", err)
	}
}

func TestRenderREADMEUsesBranchWhenCommitIsUnavailable(t *testing.T) {
	got, err := RenderREADME([]byte("[Guide](docs/guide.md)"), "rekurt/demo", "main", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.HTML, "/blob/main/docs/guide.md") || got.SourceURL != "https://github.com/rekurt/demo/blob/main/README.md" {
		t.Fatalf("branch fallback missing: %#v", got)
	}
}

func TestRenderREADMERewritesRelativeURLsInRawHTML(t *testing.T) {
	got, err := RenderREADME([]byte(`<a href="docs/guide.md"><img src="assets/banner.jpg" width="1200"></a>`), "rekurt/demo", "main", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.HTML, `href="https://github.com/rekurt/demo/blob/abc123/docs/guide.md"`) {
		t.Fatalf("raw HTML link was not pinned: %s", got.HTML)
	}
	if !strings.Contains(got.HTML, `src="https://raw.githubusercontent.com/rekurt/demo/abc123/assets/banner.jpg"`) {
		t.Fatalf("raw HTML image was not pinned: %s", got.HTML)
	}
}

func TestRenderREADMEMakesScrollableContentKeyboardFocusable(t *testing.T) {
	got, err := RenderREADME([]byte("```sh\nlong command\n```\n\n| key | value |\n| --- | --- |\n| a | b |"), "rekurt/demo", "main", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	for _, element := range []string{`<pre tabindex="0">`, `<table tabindex="0">`} {
		if !strings.Contains(got.HTML, element) {
			t.Fatalf("missing focusable scroll region %s: %s", element, got.HTML)
		}
	}
}
