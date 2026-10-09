package projectsite

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

//go:embed templates/*.html assets/*
var siteFiles embed.FS

type alternateView struct {
	Lang string
	URL  string
	Name string
}

type actionView struct {
	Label string
	URL   string
	Kind  string
}

type siblingView struct {
	Name    string
	Summary string
	Kind    string
	Domain  string
	Accent  string
	URL     string
}

type pageView struct {
	Marketing     *marketingView
	Page          LocalePage
	Copy          siteCopy
	Name          string
	Owner         string
	Kind          string
	Domain        string
	Accent        string
	Layout        string
	Version       string
	License       string
	Language      string
	Updated       string
	Install       []string
	Actions       []actionView
	Alternates    []alternateView
	Readme        template.HTML
	AssetPrefix   string
	DirectoryHref string
	ProjectHref   string
	Structured    template.JS
	GeneratedAt   string
	SiblingCount  int
}

type directoryView struct {
	Page        LocalePage
	Copy        siteCopy
	Name        string
	Accent      string
	AssetPrefix string
	ProjectHref string
	Alternates  []alternateView
	Siblings    []siblingView
	GeneratedAt string
}

func Build(options Options) (BuildManifest, error) {
	if info, err := os.Lstat(options.Output); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return BuildManifest{}, fmt.Errorf("output must not be a symlink")
	} else if err != nil && !os.IsNotExist(err) {
		return BuildManifest{}, fmt.Errorf("inspect output: %w", err)
	}
	model, err := Resolve(options)
	if err != nil {
		return BuildManifest{}, err
	}
	parent := filepath.Dir(model.Output)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return BuildManifest{}, fmt.Errorf("create output parent: %w", err)
	}
	temporary, err := os.MkdirTemp(parent, ".project-site-*")
	if err != nil {
		return BuildManifest{}, fmt.Errorf("create temporary output: %w", err)
	}
	defer os.RemoveAll(temporary)

	if err := renderSite(model, temporary); err != nil {
		return BuildManifest{}, err
	}
	manifest := BuildManifest{
		SchemaVersion: 1, Slug: model.Product.Slug, SourceSHA: model.Repository.HeadSHA,
		GeneratedAt: model.GeneratedAt, Mode: "build",
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return BuildManifest{}, err
	}
	data = append(data, '\n')
	if err := writeSiteFile(temporary, "family-build.json", data); err != nil {
		return BuildManifest{}, err
	}
	if err := replaceOutput(model.Output, temporary); err != nil {
		return BuildManifest{}, err
	}
	return manifest, nil
}

func renderSite(model Model, output string) error {
	baseTemplate, err := template.ParseFS(siteFiles, "templates/base.html")
	if err != nil {
		return fmt.Errorf("parse page template: %w", err)
	}
	directoryTemplate, err := template.ParseFS(siteFiles, "templates/directory.html")
	if err != nil {
		return fmt.Errorf("parse directory template: %w", err)
	}
	for _, page := range model.Pages {
		view, err := makePageView(model, page)
		if err != nil {
			return err
		}
		var rendered bytes.Buffer
		if err := baseTemplate.ExecuteTemplate(&rendered, "base.html", view); err != nil {
			return fmt.Errorf("render %s: %w", page.Locale, err)
		}
		if err := writeSiteFile(output, routeFile(page.Path), rendered.Bytes()); err != nil {
			return err
		}

		directory := makeDirectoryView(model, page)
		rendered.Reset()
		if err := directoryTemplate.ExecuteTemplate(&rendered, "directory.html", directory); err != nil {
			return fmt.Errorf("render %s directory: %w", page.Locale, err)
		}
		if err := writeSiteFile(output, routeFile(strings.TrimSuffix(page.Path, "/")+"/projects/"), rendered.Bytes()); err != nil {
			return err
		}
	}
	for _, name := range []string{"family.css", "family.js", "marketing.css"} {
		data, err := fs.ReadFile(siteFiles, "assets/"+name)
		if err != nil {
			return err
		}
		if err := writeSiteFile(output, "assets/"+name, data); err != nil {
			return err
		}
	}
	if err := writeSiteFile(output, ".nojekyll", nil); err != nil {
		return err
	}
	if err := writeSiteFile(output, "robots.txt", []byte("User-agent: *\nAllow: /\nSitemap: "+model.BaseURL+"sitemap.xml\n")); err != nil {
		return err
	}
	if err := writeSiteFile(output, "sitemap.xml", []byte(renderSitemap(model))); err != nil {
		return err
	}
	return writeSiteFile(output, "404.html", []byte(renderNotFound(model)))
}

func makePageView(model Model, page LocalePage) (pageView, error) {
	marketing, err := marketingFor(model, page)
	if err != nil {
		return pageView{}, err
	}
	if marketing != nil {
		page.Title = marketing.Name + " — " + marketing.Text.Headline
		page.Description = marketing.Text.Intro
	}
	data, err := structuredData(model, page)
	if err != nil {
		return pageView{}, err
	}
	version := copyFor(page.Locale).Unversioned
	if model.Product.Version != nil {
		version = model.Product.Version.Value
	}
	license := model.Repository.License
	if license == "" {
		license = copyFor(page.Locale).NotDeclared
	}
	language := model.Repository.Language
	if language == "" {
		language = "—"
	}
	return pageView{
		Marketing: marketing,
		Page:      page, Copy: copyFor(page.Locale), Name: model.Repository.Name, Owner: model.Owner,
		Kind: model.Product.Kind, Domain: model.Product.Domain, Accent: model.Product.Accent,
		Layout: layoutFor(model.Product.Kind), Version: version, License: license, Language: language,
		Updated: model.Repository.PushedAt.UTC().Format("2006-01-02"), Install: model.Product.Install,
		Actions: actionsFor(model, page.Locale), Alternates: alternatesFor(model, false),
		Readme: template.HTML(page.ReadmeHTML), AssetPrefix: assetPrefix(page.Path),
		DirectoryHref: "projects/", ProjectHref: "./", Structured: template.JS(data),
		GeneratedAt: model.GeneratedAt.UTC().Format("2006-01-02 15:04 UTC"), SiblingCount: len(model.Products) - 1,
	}, nil
}

func makeDirectoryView(model Model, page LocalePage) directoryView {
	view := directoryView{
		Page: page, Copy: copyFor(page.Locale), Name: model.Repository.Name, Accent: model.Product.Accent,
		AssetPrefix: assetPrefix(strings.TrimSuffix(page.Path, "/") + "/projects/"), ProjectHref: "../",
		Alternates: alternatesFor(model, true), GeneratedAt: model.GeneratedAt.UTC().Format("2006-01-02 15:04 UTC"),
	}
	for _, product := range model.Products {
		if product.Slug == model.Product.Slug {
			continue
		}
		view.Siblings = append(view.Siblings, siblingView{
			Name: repositoryName(product.PrimaryRepo), Summary: localizedSummary(product, page.Locale),
			Kind: product.Kind, Domain: product.Domain, Accent: product.Accent,
			URL: productWebsite(model.Owner, product),
		})
	}
	sort.SliceStable(view.Siblings, func(i, j int) bool { return view.Siblings[i].Name < view.Siblings[j].Name })
	return view
}

func actionsFor(model Model, locale string) []actionView {
	copy := copyFor(locale)
	labels := map[string]string{"source": copy.Source, "documentation": copy.Documentation, "release": copy.Release}
	var actions []actionView
	for _, link := range model.Product.Links {
		if label := labels[link.Kind]; label != "" {
			actions = append(actions, actionView{Label: label, URL: link.URL, Kind: link.Kind})
		}
	}
	return actions
}

func alternatesFor(model Model, directory bool) []alternateView {
	var values []alternateView
	for _, page := range model.Pages {
		url := page.Canonical
		if directory {
			url += "projects/"
		}
		values = append(values, alternateView{Lang: page.Lang, URL: url, Name: strings.ToUpper(page.Locale)})
	}
	defaultURL := model.Pages[0].Canonical
	if directory {
		defaultURL += "projects/"
	}
	values = append(values, alternateView{Lang: "x-default", URL: defaultURL, Name: "DEFAULT"})
	return values
}

func localizedSummary(product catalog.Product, locale string) string {
	switch locale {
	case "ru":
		return product.Summary.RU
	case "zh-cn":
		return product.Summary.ZHCN
	default:
		return product.Summary.EN
	}
}

func layoutFor(kind string) string {
	switch kind {
	case "cli", "skill":
		return "terminal"
	case "library", "sdk":
		return "code"
	default:
		return "systems"
	}
}

func repositoryName(nameWithOwner string) string {
	if index := strings.IndexByte(nameWithOwner, '/'); index >= 0 {
		return nameWithOwner[index+1:]
	}
	return nameWithOwner
}

func productWebsite(owner string, product catalog.Product) string {
	for _, link := range product.Links {
		if link.Kind == "website" {
			return link.URL
		}
	}
	return fmt.Sprintf("https://%s.github.io/%s/", owner, repositoryName(product.PrimaryRepo))
}

func assetPrefix(route string) string {
	depth := len(strings.FieldsFunc(strings.Trim(route, "/"), func(r rune) bool { return r == '/' }))
	return strings.Repeat("../", depth) + "assets/"
}

func routeFile(route string) string {
	clean := strings.Trim(route, "/")
	if clean == "" {
		return "index.html"
	}
	return filepath.ToSlash(filepath.Join(clean, "index.html"))
}

func writeSiteFile(root, name string, data []byte) error {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid output path %q", name)
	}
	path := filepath.Join(root, clean)
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output path escapes root: %q", name)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func replaceOutput(output, temporary string) error {
	if info, err := os.Lstat(output); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output must not be a symlink")
		}
		if err := os.RemoveAll(output); err != nil {
			return fmt.Errorf("replace output: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(temporary, output); err != nil {
		return fmt.Errorf("publish output: %w", err)
	}
	return nil
}

func renderSitemap(model Model) string {
	var output strings.Builder
	output.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	output.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">` + "\n")
	for _, directory := range []bool{false, true} {
		for _, page := range model.Pages {
			location := page.Canonical
			if directory {
				location += "projects/"
			}
			output.WriteString("  <url><loc>" + escapeXML(location) + "</loc>")
			for _, alternate := range alternatesFor(model, directory) {
				output.WriteString(`<xhtml:link rel="alternate" hreflang="` + escapeXML(alternate.Lang) + `" href="` + escapeXML(alternate.URL) + `"/>`)
			}
			output.WriteString("</url>\n")
		}
	}
	output.WriteString("</urlset>\n")
	return output.String()
}

func escapeXML(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;").Replace(value)
}

func renderNotFound(model Model) string {
	base, _ := url.Parse(model.BaseURL)
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="robots" content="noindex"><meta name="viewport" content="width=device-width"><title>Not found — ` + template.HTMLEscapeString(model.Repository.Name) + `</title><link rel="stylesheet" href="` + base.Path + `assets/family.css"></head><body><main class="not-found"><p class="eyebrow">404 / NOT FOUND</p><h1>This route is not part of the project site.</h1><a class="button primary" href="` + model.BaseURL + `">Return to ` + template.HTMLEscapeString(model.Repository.Name) + `</a></main></body></html>`
}
