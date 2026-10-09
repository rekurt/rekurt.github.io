package sitefleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	nethtml "golang.org/x/net/html"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
	"github.com/rekurt/rekurt.github.io/internal/projectsite"
	catalogsync "github.com/rekurt/rekurt.github.io/internal/sync"
)

const maxResponseSize = 5 << 20

type Options struct {
	SnapshotPath string
	Client       *http.Client
	BaseURL      func(catalog.Product) string
}

type Result struct {
	Slug     string   `json:"slug"`
	URL      string   `json:"url"`
	Mode     string   `json:"mode"`
	Checked  []string `json:"checked"`
	Verified bool     `json:"verified"`
}

type buildManifest struct {
	Mode string `json:"mode"`
}

func Check(ctx context.Context, options Options) ([]Result, error) {
	snapshot, err := catalogsync.ReadSnapshot(options.SnapshotPath)
	if err != nil {
		return nil, fmt.Errorf("read catalog snapshot: %w", err)
	}
	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	copyClient := *client
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	client = &copyClient

	results := make([]Result, 0, len(snapshot.Products))
	var failures []error
	for _, product := range snapshot.Products {
		baseURL := productionURL(snapshot.Owner, product)
		if options.BaseURL != nil {
			baseURL = options.BaseURL(product)
		}
		result, err := checkProduct(ctx, client, snapshot, product, baseURL)
		results = append(results, result)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", product.Slug, err))
		}
	}
	return results, errors.Join(failures...)
}

func checkProduct(ctx context.Context, client *http.Client, snapshot catalog.Snapshot, product catalog.Product, baseURL string) (Result, error) {
	result := Result{Slug: product.Slug, URL: baseURL}
	parsedBase, err := url.Parse(baseURL)
	if err != nil || parsedBase.Scheme != "https" || parsedBase.Host == "" || parsedBase.RawQuery != "" || parsedBase.Fragment != "" || !strings.HasSuffix(parsedBase.Path, "/") {
		return result, fmt.Errorf("invalid HTTPS base URL %q", baseURL)
	}

	var failures []error
	root, err := fetch(ctx, client, baseURL)
	result.Checked = append(result.Checked, baseURL)
	if err != nil {
		failures = append(failures, err)
	} else if err := validateRoot(root, baseURL, ""); err != nil {
		failures = append(failures, err)
	}

	manifestURL := baseURL + "family-build.json"
	manifestData, err := fetch(ctx, client, manifestURL)
	result.Checked = append(result.Checked, manifestURL)
	if err != nil {
		failures = append(failures, err)
	} else {
		var manifest buildManifest
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			failures = append(failures, fmt.Errorf("decode family-build.json: %w", err))
		} else if manifest.Mode != "build" && manifest.Mode != "decorate" {
			failures = append(failures, fmt.Errorf("unknown build mode %q", manifest.Mode))
		} else {
			result.Mode = manifest.Mode
		}
	}

	if result.Mode == "build" && root != nil {
		if err := validateRoot(root, baseURL, "build"); err != nil {
			failures = append(failures, err)
		}
	}

	routes := []string{"projects/", "ru/projects/", "zh-cn/projects/"}
	if result.Mode == "build" {
		routes = append([]string{"ru/", "zh-cn/", "robots.txt"}, routes...)
	}
	for _, route := range routes {
		endpoint := baseURL + route
		data, err := fetch(ctx, client, endpoint)
		result.Checked = append(result.Checked, endpoint)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if strings.HasSuffix(route, "projects/") {
			if projectsite.HasInsecureAsset(data) {
				failures = append(failures, fmt.Errorf("%s contains an insecure loaded resource", endpoint))
			}
			for _, sibling := range snapshot.Products {
				if sibling.Slug == product.Slug {
					continue
				}
				expected := websiteURL(snapshot.Owner, sibling)
				if !strings.Contains(string(data), expected) {
					failures = append(failures, fmt.Errorf("%s is missing sibling %s (%s)", endpoint, sibling.Slug, expected))
				}
			}
		}
	}

	sitemapName := "sitemap.xml"
	if result.Mode == "decorate" {
		sitemapName = "family-sitemap.xml"
	}
	sitemapURL := baseURL + sitemapName
	sitemap, err := fetch(ctx, client, sitemapURL)
	result.Checked = append(result.Checked, sitemapURL)
	if err != nil {
		failures = append(failures, err)
	} else if !strings.Contains(string(sitemap), baseURL) {
		failures = append(failures, fmt.Errorf("%s is missing base URL", sitemapURL))
	}

	result.Verified = len(failures) == 0
	return result, errors.Join(failures...)
}

func fetch(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", endpoint, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf("GET %s: response exceeds %d bytes", endpoint, maxResponseSize)
	}
	return data, nil
}

func validateRoot(data []byte, canonical, mode string) error {
	if projectsite.HasInsecureAsset(data) {
		return errors.New("root contains an insecure loaded resource")
	}
	document, err := nethtml.Parse(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("parse root HTML: %w", err)
	}
	canonicalFound := false
	familyFound := false
	hubFound := false
	jsonLDFound := false
	alternates := map[string]bool{}
	var walk func(*nethtml.Node)
	walk = func(node *nethtml.Node) {
		if node.Type == nethtml.ElementNode {
			rel := attribute(node, "rel")
			href := attribute(node, "href")
			if node.Data == "link" && containsToken(rel, "canonical") && href == canonical {
				canonicalFound = true
			}
			if node.Data == "link" && containsToken(rel, "alternate") {
				alternates[attribute(node, "hreflang")] = true
			}
			if node.Data == "script" && strings.EqualFold(attribute(node, "type"), "application/ld+json") {
				jsonLDFound = true
			}
			if _, ok := findAttribute(node, "data-rekurt-family"); ok {
				familyFound = true
			}
			if node.Data == "a" && href == "https://rekurt.github.io/" {
				hubFound = true
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(document)
	var failures []error
	if !canonicalFound {
		failures = append(failures, fmt.Errorf("root is missing canonical %s", canonical))
	}
	if !familyFound || !hubFound {
		failures = append(failures, errors.New("root family navigation is incomplete"))
	}
	if !jsonLDFound {
		failures = append(failures, errors.New("root is missing JSON-LD"))
	}
	if mode == "build" {
		for _, language := range []string{"en", "ru", "zh-CN", "x-default"} {
			if !alternates[language] {
				failures = append(failures, fmt.Errorf("root is missing %s alternate", language))
			}
		}
	}
	return errors.Join(failures...)
}

func productionURL(owner string, product catalog.Product) string {
	name := product.PrimaryRepo
	if index := strings.IndexByte(name, '/'); index >= 0 {
		name = name[index+1:]
	}
	return fmt.Sprintf("https://%s.github.io/%s/", owner, name)
}

func websiteURL(owner string, product catalog.Product) string {
	for _, link := range product.Links {
		if link.Kind == "website" && strings.HasPrefix(link.URL, "https://") {
			return link.URL
		}
	}
	return productionURL(owner, product)
}

func attribute(node *nethtml.Node, key string) string {
	value, _ := findAttribute(node, key)
	return value
}

func findAttribute(node *nethtml.Node, key string) (string, bool) {
	for _, item := range node.Attr {
		if item.Key == key {
			return item.Val, true
		}
	}
	return "", false
}

func containsToken(value, token string) bool {
	for _, field := range strings.Fields(value) {
		if strings.EqualFold(field, token) {
			return true
		}
	}
	return false
}
