package projectsite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMarketingProfilesHaveDistinctLocalizedStories(t *testing.T) {
	files, err := marketingFiles.ReadDir("profiles")
	if err != nil {
		t.Fatal(err)
	}
	headlines := map[string]string{}
	for _, file := range files {
		slug := strings.TrimSuffix(file.Name(), ".json")
		profile, err := loadMarketing(slug)
		if err != nil {
			t.Fatal(err)
		}
		for locale, copy := range profile.Locales {
			if previous := headlines[copy.Headline]; previous != "" {
				t.Errorf("%s/%s reuses the headline from %s", slug, locale, previous)
			}
			headlines[copy.Headline] = slug + "/" + locale
		}
		if strings.HasPrefix(profile.Image, "/") || strings.Contains(profile.Image, "..") {
			t.Errorf("unsafe image path for %s", slug)
		}
	}
}

func TestDepthDemoLinkResolvesFromEveryLocale(t *testing.T) {
	output := filepath.Join(t.TempDir(), "site")
	options := fixtureOptions(t, output)
	snapshot := strings.ReplaceAll(readFile(t, options.SnapshotPath), "git-barber", "depth")
	options.SnapshotPath = filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(options.SnapshotPath, []byte(snapshot), 0600); err != nil {
		t.Fatal(err)
	}
	options.Slug = "depth"
	options.BaseURL = "https://rekurt.github.io/depth/"
	if _, err := Build(options); err != nil {
		t.Fatal(err)
	}
	if err := Validate(options); err != nil {
		t.Fatal(err)
	}
	for path, label := range map[string]string{
		"index.html": "Try the demo", "ru/index.html": "Открыть демо", "zh-cn/index.html": "体验演示",
	} {
		page := readFile(t, filepath.Join(output, path))
		if !strings.Contains(page, `href="https://rekurt.github.io/depth/demo/">`+label) {
			t.Errorf("%s lacks the localized demo link at the project root", path)
		}
	}
	other := filepath.Join(t.TempDir(), "other")
	if _, err := Build(fixtureOptions(t, other)); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(readFile(t, filepath.Join(other, "index.html")), "/demo/") {
		t.Fatal("unrelated project acquired a demo link")
	}
}

func TestDepthWorkflowBuildIsRestrictedToTrustedSource(t *testing.T) {
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct {
				Name string `yaml:"name"`
				If   string `yaml:"if"`
				Uses string `yaml:"uses"`
			}
		}
	}
	data := readFile(t, "../../.github/workflows/project-pages.yml")
	if err := yaml.Unmarshal([]byte(data), &workflow); err != nil {
		t.Fatal(err)
	}
	wanted := map[string]bool{
		"Set up Node for Depth demo":           false,
		"Build and verify Depth demo":          false,
		"Include Depth demo in Pages artifact": false,
	}
	include, validate, upload := -1, -1, -1
	for index, step := range workflow.Jobs["build"].Steps {
		if _, ok := wanted[step.Name]; ok {
			wanted[step.Name] = true
			if step.If != "inputs.slug == 'depth' && github.repository == 'rekurt/depth'" {
				t.Errorf("%s allows an untrusted caller or another slug: %q", step.Name, step.If)
			}
		}
		if step.Name == "Include Depth demo in Pages artifact" {
			include = index
		}
		if step.Name == "Validate project site" {
			validate = index
		}
		if strings.HasPrefix(step.Uses, "actions/upload-pages-artifact@") {
			upload = index
		}
	}
	for name, present := range wanted {
		if !present {
			t.Errorf("missing guarded demo step: %s", name)
		}
	}
	if !(include >= 0 && include < validate && validate < upload) {
		t.Fatal("demo must be included before site validation and the single artifact upload")
	}
}

func TestMarketingLandingHasAdoptionPathAndLocalizedBenefits(t *testing.T) {
	output := filepath.Join(t.TempDir(), "site")
	options := fixtureOptions(t, output)
	if _, err := Build(options); err != nil {
		t.Fatal(err)
	}
	profile, err := loadMarketing(options.Slug)
	if err != nil {
		t.Fatal(err)
	}
	for _, locale := range localeDefinitions {
		data, err := os.ReadFile(filepath.Join(output, routeFile(locale.path)))
		if err != nil {
			t.Fatal(err)
		}
		page := string(data)
		for _, required := range []string{
			`data-product-profile="git-barber"`, `id="install"`, `id="overview"`,
			profile.Locales[locale.locale].Headline, profile.Locales[locale.locale].Features[0].Title,
			`marketing.css`, `git barber --list`, `id="example"`,
			`class="documentation-disclosure" open`, profile.Locales[locale.locale].Benefits, profile.Locales[locale.locale].Workflow,
		} {
			if !strings.Contains(page, required) {
				t.Errorf("%s lacks %s", locale.locale, required)
			}
		}
		if strings.Contains(page, `class="family-bar"`) || strings.Contains(page, `class="family-section"`) {
			t.Errorf("%s product still leads with the shared author brand", locale.locale)
		}
		if strings.Contains(page, "git-barber.system") {
			t.Errorf("%s still shows a generic system visual", locale.locale)
		}
		if locale.locale == "en" && !strings.Contains(page, "documentation-disclosure") {
			t.Error("English documentation should remain available in a disclosure")
		}
	}
}
