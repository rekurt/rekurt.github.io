package projectsite

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strconv"
)

//go:embed seo-titles.json
var seoTitleData []byte

func seoTitle(slug, locale, fallback string) string {
	var titles map[string]map[string]string
	if json.Unmarshal(seoTitleData, &titles) == nil {
		if title := titles[slug][locale]; title != "" {
			return title
		}
	}
	return fallback
}

func openGraphLocale(locale string) string {
	switch locale {
	case "ru":
		return "ru_RU"
	case "zh-cn":
		return "zh_CN"
	default:
		return "en_US"
	}
}

func structuredData(model Model, page LocalePage) ([]byte, error) {
	schema := map[string]any{
		"@context":            "https://schema.org",
		"@type":               "SoftwareSourceCode",
		"@id":                 model.BaseURL + "#software",
		"url":                 page.Canonical,
		"name":                model.Repository.Name,
		"author":              map[string]any{"@type": "Person", "@id": "https://rekurt.github.io/#author", "name": model.Owner, "url": "https://github.com/" + model.Owner},
		"mainEntityOfPage":    map[string]any{"@type": "WebPage", "@id": page.Canonical},
		"description":         page.Description,
		"codeRepository":      model.Repository.URL,
		"dateModified":        model.Repository.PushedAt.UTC().Format("2006-01-02"),
		"inLanguage":          page.Lang,
		"programmingLanguage": model.Repository.Language,
	}
	if model.Repository.License != "" {
		schema["license"] = model.Repository.License
	}
	if model.Product.Version != nil {
		schema["version"] = model.Product.Version.Value
	}
	return json.Marshal(schema)
}

var readmeHeading = regexp.MustCompile(`(?i)<(/?)h([1-6])(\s|>)`)

// README content is nested under the page's Documentation h2.
func nestedReadmeHTML(source string) string {
	return readmeHeading.ReplaceAllStringFunc(source, func(tag string) string {
		match := readmeHeading.FindStringSubmatch(tag)
		level, _ := strconv.Atoi(match[2])
		return "<" + match[1] + "h" + strconv.Itoa(min(6, level+2)) + match[3]
	})
}

func writeSocialImage(model Model, output string) error {
	data, err := siteFiles.ReadFile("assets/social/" + model.Product.Slug + ".png")
	if err != nil {
		return err
	}
	return writeSiteFile(output, "assets/social.png", data)
}
