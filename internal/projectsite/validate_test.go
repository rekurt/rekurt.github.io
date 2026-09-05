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
