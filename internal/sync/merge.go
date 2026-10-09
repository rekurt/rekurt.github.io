package sync

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
	readmemarkdown "github.com/rekurt/rekurt.github.io/internal/markdown"
)

func Build(manifest catalog.Manifest, repositories []catalog.Repository, syncedAt time.Time) (catalog.Snapshot, error) {
	if err := catalog.ValidateManifest(manifest); err != nil {
		return catalog.Snapshot{}, err
	}

	index := make(map[string]catalog.Repository, len(repositories))
	for _, repo := range repositories {
		if repo.Visibility != "public" {
			return catalog.Snapshot{}, fmt.Errorf("private repository is forbidden: %s", repo.NameWithOwner)
		}
		index[strings.ToLower(repo.NameWithOwner)] = repo
	}

	roles := make(map[string]string)
	products := make([]catalog.Product, 0, len(manifest.Products))
	for _, config := range manifest.Products {
		for _, name := range config.Repositories {
			if _, exists := index[strings.ToLower(name)]; !exists {
				return catalog.Snapshot{}, fmt.Errorf("product %s references unknown repository %s", config.Slug, name)
			}
		}

		primary := index[strings.ToLower(config.PrimaryRepo)]
		primaryRole := "primary"
		if config.MaintainedFork {
			primaryRole = "maintained-fork"
		}
		roles[strings.ToLower(config.PrimaryRepo)] = primaryRole
		for _, name := range config.Repositories {
			key := strings.ToLower(name)
			if key != strings.ToLower(config.PrimaryRepo) {
				roles[key] = "support"
			}
		}

		readme, err := renderReadme(primary)
		if err != nil {
			return catalog.Snapshot{}, fmt.Errorf("render README for %s: %w", config.PrimaryRepo, err)
		}
		products = append(products, catalog.Product{
			Slug:           config.Slug,
			PrimaryRepo:    config.PrimaryRepo,
			Repositories:   append([]string(nil), config.Repositories...),
			Kind:           config.Kind,
			Domain:         config.Domain,
			Accent:         config.Accent,
			Featured:       config.Featured,
			MaintainedFork: config.MaintainedFork,
			Upstream:       config.Upstream,
			Summary:        config.Summary,
			Install:        append([]string(nil), config.Install...),
			Links:          productLinks(manifest.Owner, config, primary),
			Version:        primary.Version,
			Readme:         readme,
		})
	}

	sort.SliceStable(products, func(i, j int) bool {
		return products[i].Featured && !products[j].Featured
	})

	registry := make([]catalog.Repository, 0, len(repositories))
	portfolioRepo := strings.ToLower(manifest.Owner + "/" + manifest.Owner + ".github.io")
	for _, source := range repositories {
		repo := source
		key := strings.ToLower(repo.NameWithOwner)
		switch {
		case key == portfolioRepo:
			repo.Role = "portfolio-hub"
		case roles[key] != "":
			repo.Role = roles[key]
		case repo.Fork:
			repo.Role = "fork"
		default:
			repo.Role = "unclassified"
		}
		repo.Links = repositoryLinks(manifest.Owner, repo)
		if repo.Readme != nil && repo.Readme.Source != "" {
			readme, err := renderReadme(repo)
			if err != nil {
				return catalog.Snapshot{}, fmt.Errorf("render README for %s: %w", repo.NameWithOwner, err)
			}
			repo.Readme = readme
		}
		registry = append(registry, repo)
	}
	sort.Slice(registry, func(i, j int) bool {
		return strings.ToLower(registry[i].NameWithOwner) < strings.ToLower(registry[j].NameWithOwner)
	})

	return catalog.Snapshot{
		SchemaVersion: 1,
		Owner:         manifest.Owner,
		SyncedAt:      syncedAt.UTC().Truncate(time.Second),
		Products:      products,
		Repositories:  registry,
	}, nil
}

func renderReadme(repo catalog.Repository) (*catalog.Readme, error) {
	if repo.Readme == nil {
		return nil, nil
	}
	if repo.Readme.Source == "" {
		copy := *repo.Readme
		return &copy, nil
	}
	rendered, err := readmemarkdown.RenderREADME([]byte(repo.Readme.Source), repo.NameWithOwner, repo.DefaultBranch, repo.HeadSHA)
	if err != nil {
		return nil, err
	}
	return &rendered, nil
}

func productLinks(owner string, config catalog.ProductConfig, primary catalog.Repository) []catalog.Link {
	var links []catalog.Link
	website := config.Website
	if website == "" && (!primary.Fork || config.MaintainedFork) {
		website = websiteURL(owner, primary)
	}
	links = addLink(links, "website", website)
	documentation := config.Documentation
	if documentation == "" && !primary.HasPages {
		kind, homepage := classifyHomepage(primary.Homepage)
		if kind == "documentation" {
			documentation = homepage
		}
	}
	links = addLink(links, "documentation", documentation)
	links = addLink(links, "source", primary.URL)
	if primary.Version != nil {
		links = addLink(links, "release", primary.Version.URL)
	}
	return links
}

func repositoryLinks(owner string, repo catalog.Repository) []catalog.Link {
	var links []catalog.Link
	if !repo.Fork || repo.Role == "maintained-fork" {
		if repo.HasPages {
			links = addLink(links, "website", websiteURL(owner, repo))
		} else {
			kind, homepage := classifyHomepage(repo.Homepage)
			links = addLink(links, kind, homepage)
		}
	}
	links = addLink(links, "source", repo.URL)
	if repo.Version != nil {
		links = addLink(links, "release", repo.Version.URL)
	}
	return links
}

func websiteURL(owner string, repo catalog.Repository) string {
	if repo.Homepage != "" && repo.HasPages {
		return repo.Homepage
	}
	if !repo.HasPages {
		kind, homepage := classifyHomepage(repo.Homepage)
		if kind == "website" {
			return homepage
		}
		return ""
	}
	if strings.EqualFold(repo.Name, owner+".github.io") {
		return "https://" + owner + ".github.io/"
	}
	return fmt.Sprintf("https://%s.github.io/%s/", owner, repo.Name)
}

func classifyHomepage(raw string) (string, string) {
	if raw == "" {
		return "", ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", ""
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "pkg.go.dev" || host == "crates.io" || host == "docs.rs" || host == "godoc.org" {
		return "documentation", raw
	}
	if host == "github.com" && (parsed.Fragment == "readme" || strings.Contains(parsed.Path, "/blob/")) {
		return "documentation", raw
	}
	return "website", raw
}

func addLink(links []catalog.Link, kind, rawURL string) []catalog.Link {
	if kind == "" || rawURL == "" {
		return links
	}
	for _, link := range links {
		if link.Kind == kind && link.URL == rawURL {
			return links
		}
	}
	return append(links, catalog.Link{Kind: kind, URL: rawURL})
}
