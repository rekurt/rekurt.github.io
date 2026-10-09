package projectsite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func TestProjectNavigationAndPublishedPackageAcrossLocales(t *testing.T) {
	options := fixtureOptions(t, filepath.Join(t.TempDir(), "site"))
	var snapshot catalog.Snapshot
	if err := json.Unmarshal([]byte(strings.ReplaceAll(readFile(t, options.SnapshotPath), "git-barber", "depth")), &snapshot); err != nil {
		t.Fatal(err)
	}
	snapshot.Products[0].NPMPackage = &catalog.NPMPackage{Name: "@rekurt/depth", Version: "0.1.0"}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	options.SnapshotPath = filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(options.SnapshotPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	options.Slug, options.BaseURL = "depth", "https://rekurt.github.io/depth/"
	if _, err := Build(options); err != nil {
		t.Fatal(err)
	}
	for _, locale := range localeDefinitions {
		page := readFile(t, filepath.Join(options.Output, routeFile(locale.path)))
		for _, wanted := range []string{
			`class="project-navigation"`, `class="project-languages"`,
			`href="#install"`, `href="#overview"`, `role="status"`,
			`https://www.npmjs.com/package/@rekurt/depth/v/0.1.0`, `GitHub · v0.3.0`,
			`data-copy-failed=`, `data-copy-label=`,
		} {
			if !strings.Contains(page, wanted) {
				t.Errorf("%s lacks %s", locale.locale, wanted)
			}
		}
	}
	if err := Validate(options); err != nil {
		t.Fatal(err)
	}
}
