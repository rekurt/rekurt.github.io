package projectsite

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
	readmemarkdown "github.com/rekurt/rekurt.github.io/internal/markdown"
	catalogsync "github.com/rekurt/rekurt.github.io/internal/sync"
)

type localeDefinition struct {
	locale  string
	lang    string
	path    string
	summary func(catalog.LocalizedText) string
	readmes []string
}

var localeDefinitions = []localeDefinition{
	{locale: "en", lang: "en", path: "/", summary: func(text catalog.LocalizedText) string { return text.EN }, readmes: []string{"README.en.md", "README_EN.md", "README.md"}},
	{locale: "ru", lang: "ru", path: "/ru/", summary: func(text catalog.LocalizedText) string { return text.RU }, readmes: []string{"README.ru.md", "README_RU.md"}},
	{locale: "zh-cn", lang: "zh-CN", path: "/zh-cn/", summary: func(text catalog.LocalizedText) string { return text.ZHCN }, readmes: []string{"README.zh-CN.md", "README.zh-Hans.md", "README.zh.md", "README_ZH.md"}},
}

func Resolve(options Options) (Model, error) {
	baseURL, err := normalizeBaseURL(options.BaseURL)
	if err != nil {
		return Model{}, err
	}
	repositoryRoot, err := filepath.Abs(options.Repository)
	if err != nil {
		return Model{}, fmt.Errorf("resolve repository path: %w", err)
	}
	info, err := os.Stat(repositoryRoot)
	if err != nil || !info.IsDir() {
		return Model{}, fmt.Errorf("repository directory is required: %s", repositoryRoot)
	}
	output, err := filepath.Abs(options.Output)
	if err != nil {
		return Model{}, fmt.Errorf("resolve output path: %w", err)
	}
	if output == repositoryRoot {
		return Model{}, fmt.Errorf("repository and output must differ")
	}

	snapshot, err := catalogsync.ReadSnapshot(options.SnapshotPath)
	if err != nil {
		return Model{}, err
	}
	var product *catalog.Product
	for index := range snapshot.Products {
		if snapshot.Products[index].Slug == options.Slug {
			product = &snapshot.Products[index]
			break
		}
	}
	if product == nil {
		return Model{}, fmt.Errorf("unknown product %q", options.Slug)
	}
	var repository *catalog.Repository
	for index := range snapshot.Repositories {
		if strings.EqualFold(snapshot.Repositories[index].NameWithOwner, product.PrimaryRepo) {
			repository = &snapshot.Repositories[index]
			break
		}
	}
	if repository == nil {
		return Model{}, fmt.Errorf("primary repository %s is missing from snapshot", product.PrimaryRepo)
	}

	generatedAt := options.GeneratedAt.UTC()
	if generatedAt.IsZero() {
		generatedAt = snapshot.SyncedAt.UTC()
	}
	model := Model{
		Owner: snapshot.Owner, Product: *product, Repository: *repository,
		Products: snapshot.Products, BaseURL: baseURL, Output: output,
		RepositoryRoot: repositoryRoot, GeneratedAt: generatedAt,
	}
	for _, locale := range localeDefinitions {
		summary := locale.summary(product.Summary)
		page := LocalePage{
			Locale: locale.locale, Lang: locale.lang, Path: locale.path,
			Canonical:   routeURL(baseURL, locale.path),
			Title:       repository.Name + " — " + summary,
			Description: summary, Summary: summary,
		}
		readme, filename, err := readLocalizedREADME(repositoryRoot, locale.readmes, *repository)
		if err != nil {
			return Model{}, err
		}
		if filename != "" {
			page.ReadmeHTML = readme.HTML
			ref := repository.HeadSHA
			if ref == "" {
				ref = repository.DefaultBranch
			}
			page.ReadmeSourceURL = fmt.Sprintf("https://github.com/%s/blob/%s/%s", repository.NameWithOwner, ref, filename)
		}
		model.Pages = append(model.Pages, page)
	}
	return model, nil
}

func normalizeBaseURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("base URL is invalid")
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("base URL must use https")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("base URL must not contain query or fragment")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("base URL must not contain credentials")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/"
	return parsed.String(), nil
}

func routeURL(baseURL, route string) string {
	return baseURL + strings.TrimPrefix(route, "/")
}

func readLocalizedREADME(root string, candidates []string, repository catalog.Repository) (catalog.Readme, string, error) {
	for _, filename := range candidates {
		path := filepath.Join(root, filename)
		source, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return catalog.Readme{}, "", fmt.Errorf("read %s: %w", filename, err)
		}
		readme, err := readmemarkdown.RenderREADME(source, repository.NameWithOwner, repository.DefaultBranch, repository.HeadSHA)
		if err != nil {
			return catalog.Readme{}, "", fmt.Errorf("render %s: %w", filename, err)
		}
		return readme, filename, nil
	}
	return catalog.Readme{}, "", nil
}
