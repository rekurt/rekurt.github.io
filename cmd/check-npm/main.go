package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
	"github.com/rekurt/rekurt.github.io/internal/npmregistry"
	catalogsync "github.com/rekurt/rekurt.github.io/internal/sync"
)

func main() {
	manifestPath := flag.String("manifest", "catalog/projects.yaml", "curated manifest")
	snapshotPath := flag.String("snapshot", "site/src/data/generated/catalog.json", "generated snapshot")
	flag.Parse()
	err := run(*manifestPath, *snapshotPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(manifestPath, snapshotPath string) error {
	manifest, err := catalog.LoadManifest(manifestPath)
	if err != nil {
		return err
	}
	snapshot, err := catalogsync.ReadSnapshot(snapshotPath)
	if err != nil {
		return err
	}
	results, err := npmregistry.Check(context.Background(), &http.Client{Timeout: 15 * time.Second}, npmregistry.RegistryURL, manifest, snapshot)
	if err != nil {
		return err
	}
	for _, result := range results {
		fmt.Printf("verified %s@%s: %s\n", result.Name, result.Version, result.URL)
		if result.Latest != result.Version {
			fmt.Printf("::warning::%s: curated npm version %s differs from registry latest %s; review the published version and installation docs\n", result.Name, result.Version, result.Latest)
		}
	}
	fmt.Printf("verified %d public npm packages\n", len(results))
	return nil
}
