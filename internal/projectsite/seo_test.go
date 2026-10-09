package projectsite

import (
	"encoding/json"
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
