package projectsite

import "encoding/json"

func structuredData(model Model, page LocalePage) ([]byte, error) {
	schema := map[string]any{
		"@context":            "https://schema.org",
		"@type":               "SoftwareSourceCode",
		"name":                model.Repository.Name,
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
