package npmregistry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func fixture() (catalog.Manifest, catalog.Snapshot) {
	pkg := catalog.NPMPackage{Name: "@rekurt/depth", Version: "0.1.0"}
	config := catalog.ProductConfig{Slug: "depth", PrimaryRepo: "rekurt/depth", Repositories: []string{"rekurt/depth"},
		Kind: "library", Domain: "fintech", Accent: "cyan", Summary: catalog.LocalizedText{EN: "Depth", RU: "Глубина", ZHCN: "深度"},
		NPMPackage: &pkg, Install: []string{"npm install @rekurt/depth"}}
	snapshotPkg := pkg
	product := catalog.Product{Slug: config.Slug, NPMPackage: &snapshotPkg, Install: append([]string(nil), config.Install...)}
	return catalog.Manifest{Owner: "rekurt", Products: []catalog.ProductConfig{config}}, catalog.Snapshot{Products: []catalog.Product{product}}
}

func TestCheckRegistryAndInstallation(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(*catalog.Manifest, *catalog.Snapshot)
		status        int
		latest        string
		integrity     string
		wrongIdentity bool
		want          string
	}{
		{name: "scoped release and unpinned install", latest: "0.1.0"},
		{name: "new latest is reported without invalidating published version", latest: "0.2.0"},
		{name: "global CLI command", latest: "0.1.0", mutate: func(m *catalog.Manifest, s *catalog.Snapshot) {
			m.Products[0].NPMPackage.Name = "gitlab-dump-cli"
			s.Products[0].NPMPackage.Name = "gitlab-dump-cli"
			m.Products[0].Install = []string{"npm install -g gitlab-dump-cli@0.1.0 && gitlab-dump --help"}
			s.Products[0].Install = m.Products[0].Install
		}},
		{name: "stale generated version", latest: "0.1.0", want: "generated npm metadata differs", mutate: func(m *catalog.Manifest, s *catalog.Snapshot) { s.Products[0].NPMPackage.Version = "0.0.9" }},
		{name: "wrong pinned install version", latest: "0.1.0", want: "install command conflicts", mutate: func(m *catalog.Manifest, s *catalog.Snapshot) {
			m.Products[0].Install = []string{"npm install @rekurt/depth@0.0.9"}
		}},
		{name: "package missing from install command", latest: "0.1.0", want: "npm install command", mutate: func(m *catalog.Manifest, s *catalog.Snapshot) {
			m.Products[0].Install = []string{"npm install other-package"}
		}},
		{name: "generated install drift", latest: "0.1.0", want: "install command conflicts", mutate: func(m *catalog.Manifest, s *catalog.Snapshot) {
			s.Products[0].Install = []string{"npm install @rekurt/depth@0.0.9"}
		}},
		{name: "unpublished package", latest: "0.1.0", status: 404, want: "HTTP 404"},
		{name: "registry unavailable", latest: "0.1.0", status: 503, want: "HTTP 503"},
		{name: "invalid integrity", latest: "0.1.0", integrity: "sha512-invalid", want: "integrity"},
		{name: "wrong identity", latest: "0.1.0", wrongIdentity: true, want: "identity mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, snapshot := fixture()
			if tt.mutate != nil {
				tt.mutate(&manifest, &snapshot)
			}
			pkg := manifest.Products[0].NPMPackage
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodGet || r.Header.Get("Authorization") != "" {
					t.Error("registry checks must be unauthenticated GETs")
				}
				if tt.status != 0 {
					w.WriteHeader(tt.status)
					return
				}
				data := metadata{Name: pkg.Name, Version: pkg.Version}
				if r.URL.Path == "/"+pkg.Name+"/latest" {
					data.Version = tt.latest
				} else if r.URL.Path != "/"+pkg.Name+"/"+pkg.Version {
					t.Errorf("unexpected registry path %q", r.URL.Path)
				}
				if tt.wrongIdentity {
					data.Name = "wrong-package"
				}
				data.Dist.Integrity = "sha512-" + base64.StdEncoding.EncodeToString(make([]byte, 64))
				if tt.integrity != "" {
					data.Dist.Integrity = tt.integrity
				}
				data.Dist.Tarball = "https://registry.npmjs.org/" + pkg.Name + "/-/package.tgz"
				_ = json.NewEncoder(w).Encode(data)
			}))
			defer server.Close()
			results, err := Check(context.Background(), server.Client(), server.URL, manifest, snapshot)
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("error=%v want %q", err, tt.want)
				}
				if strings.Contains(tt.want, "install") || strings.Contains(tt.want, "generated") {
					if requests != 0 {
						t.Fatal("invalid local metadata reached registry")
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || results[0].Version != pkg.Version || results[0].Latest != tt.latest || results[0].URL != "https://www.npmjs.com/package/"+pkg.Name+"/v/"+pkg.Version {
				t.Fatalf("results=%#v", results)
			}
			if requests != 2 {
				t.Fatalf("requests=%d want exact-version + latest reads", requests)
			}
		})
	}
}
