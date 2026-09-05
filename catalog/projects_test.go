package catalog_test

import (
	"slices"
	"testing"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func TestProductionManifestContract(t *testing.T) {
	manifest, err := catalog.LoadManifest("projects.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateManifest(manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Owner != "rekurt" || len(manifest.Products) != 14 {
		t.Fatalf("owner/products = %q/%d, want rekurt/14", manifest.Owner, len(manifest.Products))
	}

	wantPrimary := map[string]string{
		"chislo": "rekurt/chislo", "cortex-forge": "rekurt/cortex-forge", "dbdiff": "rekurt/dbdiff",
		"depth": "rekurt/depth", "git-barber": "rekurt/git-barber", "gitlab-downloader": "rekurt/gitlab-downloader",
		"go-propisyu": "rekurt/go-propisyu", "gost-crypto": "rekurt/gost-crypto", "mac-coffee": "rekurt/Mac-Coffee",
		"openkline": "rekurt/openkline", "prt": "rekurt/prt", "sprint-velocity": "rekurt/sprint-velocity",
		"vpn-hub": "rekurt/vpn-hub", "ymsdk": "rekurt/ymsdk",
	}
	featured := make([]string, 0, 6)
	for _, product := range manifest.Products {
		if wantPrimary[product.Slug] != product.PrimaryRepo {
			t.Fatalf("primary repo for %s = %q, want %q", product.Slug, product.PrimaryRepo, wantPrimary[product.Slug])
		}
		if product.Featured {
			featured = append(featured, product.Slug)
		}
		if product.Summary.ZHCN == "" {
			t.Fatalf("Chinese summary for %s is empty", product.Slug)
		}
		if product.Accent == "" {
			t.Fatalf("accent for %s is empty", product.Slug)
		}
		for _, repo := range product.Repositories {
			if repo == "rekurt/tsql" {
				t.Fatal("tsql must remain registry-only")
			}
		}
		delete(wantPrimary, product.Slug)
	}
	if len(wantPrimary) != 0 {
		t.Fatalf("missing products = %#v", wantPrimary)
	}
	slices.Sort(featured)
	wantFeatured := []string{"git-barber", "gost-crypto", "mac-coffee", "openkline", "vpn-hub", "ymsdk"}
	if !slices.Equal(featured, wantFeatured) {
		t.Fatalf("featured = %#v, want %#v", featured, wantFeatured)
	}

	macCoffee := findProduct(t, manifest, "mac-coffee")
	if !macCoffee.MaintainedFork || macCoffee.Upstream != "Elliotwu-7/Mac-Coffee" {
		t.Fatalf("Mac Coffee attribution = %#v", macCoffee)
	}
	if !slices.Equal(macCoffee.Repositories, []string{"rekurt/Mac-Coffee", "rekurt/homebrew-maccoffee", "rekurt/maccoffee-dist"}) {
		t.Fatalf("Mac Coffee repositories = %#v", macCoffee.Repositories)
	}
	openkline := findProduct(t, manifest, "openkline")
	if len(openkline.Repositories) != 4 || !slices.Contains(openkline.Repositories, "rekurt/openkline.tech") {
		t.Fatalf("openkline repositories = %#v", openkline.Repositories)
	}
}

func findProduct(t *testing.T, manifest catalog.Manifest, slug string) catalog.ProductConfig {
	t.Helper()
	for _, product := range manifest.Products {
		if product.Slug == slug {
			return product
		}
	}
	t.Fatalf("product %s not found", slug)
	return catalog.ProductConfig{}
}
