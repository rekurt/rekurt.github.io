package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	got, err := LoadManifest("testdata/valid.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got.Owner != "rekurt" {
		t.Fatalf("owner = %q, want rekurt", got.Owner)
	}
	if len(got.Products) != 1 {
		t.Fatalf("products = %d, want 1", len(got.Products))
	}
	product := got.Products[0]
	if product.Slug != "mac-coffee" {
		t.Fatalf("slug = %q, want mac-coffee", product.Slug)
	}
	if product.Upstream != "Elliotwu-7/Mac-Coffee" {
		t.Fatalf("upstream = %q", product.Upstream)
	}
}

func TestLoadManifestRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	data := []byte("owner: rekurt\nproducts: []\nproducst: []\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "producst") {
		t.Fatalf("error = %v, want unknown field producst", err)
	}
}

func TestLoadManifestRejectsMultipleDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	data := []byte("owner: rekurt\nproducts: []\n---\nowner: other\nproducts: []\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Fatalf("error = %v, want multiple document error", err)
	}
}
