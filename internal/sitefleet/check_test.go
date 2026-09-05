package sitefleet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func TestCheckValidatesGeneratedProjectFamily(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(response, request)
			return
		}
		slug := parts[0]
		base := server.URL + "/" + slug + "/"
		other := server.URL + "/beta/"
		if slug == "beta" {
			other = server.URL + "/alpha/"
		}
		suffix := strings.TrimPrefix(request.URL.Path, "/"+slug+"/")
		switch suffix {
		case "":
			fmt.Fprintf(response, `<html><head><link rel="canonical" href="%s"><link rel="alternate" hreflang="en" href="%s"><link rel="alternate" hreflang="ru" href="%sru/"><link rel="alternate" hreflang="zh-CN" href="%szh-cn/"><link rel="alternate" hreflang="x-default" href="%s"><script type="application/ld+json">{}</script></head><body data-rekurt-family><a href="https://rekurt.github.io/">Hub</a></body></html>`, base, base, base, base, base)
		case "family-build.json":
			response.Header().Set("Content-Type", "application/json")
			fmt.Fprint(response, `{"mode":"build"}`)
		case "ru/", "zh-cn/", "robots.txt":
			fmt.Fprint(response, "ok")
		case "projects/", "ru/projects/", "zh-cn/projects/":
			fmt.Fprintf(response, `<html><body data-rekurt-family><a href="%s">Sibling</a></body></html>`, other)
		case "sitemap.xml":
			fmt.Fprintf(response, `<urlset><loc>%s</loc></urlset>`, base)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	snapshot := catalog.Snapshot{Owner: "rekurt", Products: []catalog.Product{
		{Slug: "alpha", PrimaryRepo: "rekurt/alpha", Links: []catalog.Link{{Kind: "website", URL: server.URL + "/alpha/"}}},
		{Slug: "beta", PrimaryRepo: "rekurt/beta", Links: []catalog.Link{{Kind: "website", URL: server.URL + "/beta/"}}},
	}}
	snapshotPath := filepath.Join(t.TempDir(), "catalog.json")
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(snapshotPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	results, err := Check(context.Background(), Options{
		SnapshotPath: snapshotPath,
		Client:       server.Client(),
		BaseURL: func(product catalog.Product) string {
			return server.URL + "/" + product.Slug + "/"
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].Mode != "build" || len(results[0].Checked) != 9 {
		t.Fatalf("results = %#v", results)
	}
}

func TestCheckRejectsRedirectsAndInsecureAssets(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{name: "redirect", handler: func(response http.ResponseWriter, request *http.Request) {
			http.Redirect(response, request, "/elsewhere", http.StatusFound)
		}, want: "status 302"},
		{name: "mixed content", handler: func(response http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(response, `<html><head><link rel="canonical" href="BASE"><script type="application/ld+json">{}</script></head><body data-rekurt-family><a href="https://rekurt.github.io/">Hub</a><img src="http://example.com/a.png"></body></html>`)
		}, want: "insecure loaded resource"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(tt.handler)
			defer server.Close()
			snapshotPath := writeSnapshot(t, catalog.Snapshot{Owner: "rekurt", Products: []catalog.Product{{Slug: "alpha", PrimaryRepo: "rekurt/alpha"}}})
			_, err := Check(context.Background(), Options{
				SnapshotPath: snapshotPath,
				Client:       server.Client(),
				BaseURL:      func(catalog.Product) string { return server.URL + "/" },
			})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func writeSnapshot(t *testing.T, snapshot catalog.Snapshot) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catalog.json")
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
