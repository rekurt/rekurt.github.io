package projectsite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAcceptsGeneratedAndDecoratedArtifacts(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "generated")
	generatedOptions := fixtureOptions(t, generated)
	if _, err := Build(generatedOptions); err != nil {
		t.Fatal(err)
	}
	if err := Validate(generatedOptions); err != nil {
		t.Fatalf("generated validation: %v", err)
	}

	decorated := copyExistingFixture(t)
	decoratedOptions := fixtureOptions(t, decorated)
	if _, err := Decorate(decoratedOptions); err != nil {
		t.Fatal(err)
	}
	if err := Validate(decoratedOptions); err != nil {
		t.Fatalf("decorated validation: %v", err)
	}
}

func TestValidateRejectsBrokenCanonicalAndLinkGraph(t *testing.T) {
	output := filepath.Join(t.TempDir(), "site")
	options := fixtureOptions(t, output)
	if _, err := Build(options); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(output, "index.html")
	index := readFile(t, indexPath)
	index = strings.Replace(index, `rel="canonical" href="https://rekurt.github.io/git-barber/"`, `rel="canonical" href="https://example.com/wrong/"`, 1)
	if err := os.WriteFile(indexPath, []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	directoryPath := filepath.Join(output, "projects/index.html")
	directory := strings.ReplaceAll(readFile(t, directoryPath), "https://rekurt.github.io/ymsdk/", "")
	if err := os.WriteFile(directoryPath, []byte(directory), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Validate(options)
	if err == nil {
		t.Fatal("validation unexpectedly succeeded")
	}
	for _, required := range []string{"root canonical", "missing sibling ymsdk"} {
		if !strings.Contains(err.Error(), required) {
			t.Errorf("validation error lacks %q: %v", required, err)
		}
	}
}

func TestContainsInsecureAssetDistinguishesNavigationFromLoadedResources(t *testing.T) {
	tests := []struct {
		name string
		html string
		want bool
	}{
		{name: "external navigation", html: `<a href="http://www.tc26.ru/">TC26</a>`, want: false},
		{name: "image", html: `<img src="http://example.com/preview.png">`, want: true},
		{name: "script", html: `<script src="http://example.com/app.js"></script>`, want: true},
		{name: "stylesheet", html: `<link rel="stylesheet" href="http://example.com/app.css">`, want: true},
		{name: "data URI namespace", html: `<link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg'></svg>">`, want: false},
		{name: "mixed source set", html: `<source srcset="https://example.com/a.webp 1x, http://example.com/b.webp 2x">`, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsInsecureAsset([]byte(tt.html)); got != tt.want {
				t.Fatalf("containsInsecureAsset() = %v, want %v", got, tt.want)
			}
		})
	}
}
