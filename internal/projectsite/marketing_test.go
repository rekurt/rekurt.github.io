package projectsite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			`marketing.css`, `git barber --list`,
		} {
			if !strings.Contains(page, required) {
				t.Errorf("%s lacks %s", locale.locale, required)
			}
		}
		if strings.Contains(page, "git-barber.system") {
			t.Errorf("%s still shows a generic system visual", locale.locale)
		}
		if locale.locale == "en" && !strings.Contains(page, "documentation-disclosure") {
			t.Error("English documentation should remain available in a disclosure")
		}
	}
}
