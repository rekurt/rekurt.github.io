// Package npmregistry verifies curated public npm metadata without publishing or authenticating.
package npmregistry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

const RegistryURL = "https://registry.npmjs.org"

type Result struct {
	Name    string
	Version string
	Latest  string
	URL     string
}

type metadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Dist    struct {
		Integrity string `json:"integrity"`
		Tarball   string `json:"tarball"`
	} `json:"dist"`
}

// Check rejects stale generated data, conflicting installation instructions and
// missing or malformed releases. A newer latest version is reported separately.
func Check(ctx context.Context, client *http.Client, registry string, manifest catalog.Manifest, snapshot catalog.Snapshot) ([]Result, error) {
	if err := catalog.ValidateManifest(manifest); err != nil {
		return nil, err
	}
	products := make(map[string]catalog.Product, len(snapshot.Products))
	for _, product := range snapshot.Products {
		products[product.Slug] = product
	}
	var results []Result
	for _, config := range manifest.Products {
		product, exists := products[config.Slug]
		if !exists || !reflect.DeepEqual(config.NPMPackage, product.NPMPackage) {
			return nil, fmt.Errorf("%s: generated npm metadata differs from curated manifest; run catalog-sync", config.Slug)
		}
		if config.NPMPackage == nil {
			continue
		}
		pkg := *config.NPMPackage
		for _, commands := range [][]string{config.Install, product.Install} {
			if err := checkInstall(pkg, commands); err != nil {
				return nil, fmt.Errorf("%s: %w", config.Slug, err)
			}
		}
		exact, err := fetch(ctx, client, registry, pkg.Name, pkg.Version)
		if err != nil {
			return nil, err
		}
		if exact.Name != pkg.Name || exact.Version != pkg.Version {
			return nil, fmt.Errorf("%s@%s: registry identity mismatch", pkg.Name, pkg.Version)
		}
		integrity, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(exact.Dist.Integrity, "sha512-"))
		if err != nil || !strings.HasPrefix(exact.Dist.Integrity, "sha512-") || len(integrity) != 64 {
			return nil, fmt.Errorf("%s@%s: registry SHA-512 integrity is missing or malformed", pkg.Name, pkg.Version)
		}
		tarball, err := url.Parse(exact.Dist.Tarball)
		if err != nil || tarball.Scheme != "https" || tarball.Host != "registry.npmjs.org" || tarball.User != nil {
			return nil, fmt.Errorf("%s@%s: tarball must use the official HTTPS npm registry", pkg.Name, pkg.Version)
		}
		latest, err := fetch(ctx, client, registry, pkg.Name, "latest")
		if err != nil {
			return nil, err
		}
		if latest.Name != pkg.Name || latest.Version == "" {
			return nil, fmt.Errorf("%s: invalid latest registry identity", pkg.Name)
		}
		results = append(results, Result{Name: pkg.Name, Version: pkg.Version, Latest: latest.Version,
			URL: "https://www.npmjs.com/package/" + pkg.Name + "/v/" + pkg.Version})
	}
	return results, nil
}

func checkInstall(pkg catalog.NPMPackage, commands []string) error {
	found := false
	for _, command := range commands {
		args := strings.Fields(strings.SplitN(command, "&&", 2)[0])
		if len(args) < 3 || args[0] != "npm" || args[1] != "install" {
			continue
		}
		for _, arg := range args[2:] {
			if arg == pkg.Name {
				found = true
				continue
			}
			if strings.HasPrefix(arg, pkg.Name+"@") {
				if arg != pkg.Name+"@"+pkg.Version {
					return fmt.Errorf("install command conflicts with %s@%s", pkg.Name, pkg.Version)
				}
				found = true
			}
		}
	}
	if !found {
		return fmt.Errorf("npm install command for %s is missing", pkg.Name)
	}
	return nil
}

func fetch(ctx context.Context, client *http.Client, registry, name, version string) (metadata, error) {
	endpoint := strings.TrimRight(registry, "/") + "/" + url.PathEscape(name) + "/" + url.PathEscape(version)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return metadata{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return metadata{}, fmt.Errorf("%s@%s: registry request failed: %w", name, version, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return metadata{}, fmt.Errorf("%s@%s: registry returned HTTP %d", name, version, response.StatusCode)
	}
	var result metadata
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return metadata{}, fmt.Errorf("%s@%s: invalid registry JSON: %w", name, version, err)
	}
	return result, nil
}
