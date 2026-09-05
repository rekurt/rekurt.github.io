package projectsite

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	nethtml "golang.org/x/net/html"
)

func Validate(options Options) error {
	model, err := Resolve(options)
	if err != nil {
		return err
	}
	manifest, err := readBuildManifest(filepath.Join(model.Output, "family-build.json"))
	if err != nil {
		return err
	}
	var failures []error
	if manifest.Slug != model.Product.Slug {
		failures = append(failures, fmt.Errorf("build manifest slug %q does not match %q", manifest.Slug, model.Product.Slug))
	}
	if manifest.SourceSHA != model.Repository.HeadSHA {
		failures = append(failures, fmt.Errorf("build manifest source %q does not match %q", manifest.SourceSHA, model.Repository.HeadSHA))
	}
	if manifest.Mode != "build" && manifest.Mode != "decorate" {
		failures = append(failures, fmt.Errorf("unknown build mode %q", manifest.Mode))
	}

	rootPath := filepath.Join(model.Output, "index.html")
	if err := validateCanonical(rootPath, model.Pages[0].Canonical); err != nil {
		failures = append(failures, fmt.Errorf("root canonical: %w", err))
	}
	root, err := os.ReadFile(rootPath)
	if err != nil {
		failures = append(failures, fmt.Errorf("read root: %w", err))
	} else {
		if !bytesContainAll(root, []string{"data-rekurt-family", "https://rekurt.github.io/"}) {
			failures = append(failures, errors.New("root family navigation is incomplete"))
		}
		if containsInsecureAsset(root) {
			failures = append(failures, errors.New("root contains insecure asset or link"))
		}
	}

	if manifest.Mode == "build" {
		for _, page := range model.Pages[1:] {
			if err := validateCanonical(filepath.Join(model.Output, routeFile(page.Path)), page.Canonical); err != nil {
				failures = append(failures, fmt.Errorf("%s canonical: %w", page.Locale, err))
			}
		}
	}
	for _, page := range model.Pages {
		directoryPath := filepath.Join(model.Output, routeFile(strings.TrimSuffix(page.Path, "/")+"/projects/"))
		data, err := os.ReadFile(directoryPath)
		if err != nil {
			failures = append(failures, fmt.Errorf("read %s project directory: %w", page.Locale, err))
			continue
		}
		for _, sibling := range model.Products {
			if sibling.Slug == model.Product.Slug {
				continue
			}
			if !strings.Contains(string(data), productWebsite(model.Owner, sibling)) {
				failures = append(failures, fmt.Errorf("%s project directory missing sibling %s", page.Locale, sibling.Slug))
			}
		}
		if containsInsecureAsset(data) {
			failures = append(failures, fmt.Errorf("%s project directory contains insecure asset or link", page.Locale))
		}
	}
	sitemap := "sitemap.xml"
	if manifest.Mode == "decorate" {
		sitemap = "family-sitemap.xml"
	}
	if data, err := os.ReadFile(filepath.Join(model.Output, sitemap)); err != nil {
		failures = append(failures, fmt.Errorf("read %s: %w", sitemap, err))
	} else if !strings.Contains(string(data), model.BaseURL) {
		failures = append(failures, fmt.Errorf("%s does not contain canonical base URL", sitemap))
	}
	return errors.Join(failures...)
}

func readBuildManifest(path string) (BuildManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BuildManifest{}, fmt.Errorf("read build manifest: %w", err)
	}
	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return BuildManifest{}, fmt.Errorf("decode build manifest: %w", err)
	}
	return manifest, nil
}

func validateCanonical(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	document, err := nethtml.Parse(strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	var canonicals []string
	var walk func(*nethtml.Node)
	walk = func(node *nethtml.Node) {
		if isCanonical(node) {
			canonicals = append(canonicals, attributeValue(node, "href"))
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(document)
	if len(canonicals) != 1 {
		return fmt.Errorf("found %d canonical links", len(canonicals))
	}
	if canonicals[0] != expected {
		return fmt.Errorf("got %q, want %q", canonicals[0], expected)
	}
	return nil
}

func bytesContainAll(data []byte, values []string) bool {
	text := string(data)
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}

func containsInsecureAsset(data []byte) bool {
	text := strings.ToLower(string(data))
	return strings.Contains(text, `href="http://`) || strings.Contains(text, `src="http://`)
}
