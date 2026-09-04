package catalog

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
)

func ValidateManifest(manifest Manifest) error {
	var globalMessages []string
	var productMessages []string
	if strings.TrimSpace(manifest.Owner) == "" {
		globalMessages = append(globalMessages, "manifest owner is required")
	}

	slugs := make(map[string]string)
	primaryRepos := make(map[string]string)
	for i, product := range manifest.Products {
		name := product.Slug
		if name == "" {
			name = fmt.Sprintf("product[%d]", i)
		}
		add := func(message string) {
			productMessages = append(productMessages, name+": "+message)
		}

		if !slugPattern.MatchString(product.Slug) {
			add("invalid slug")
		}
		if first, exists := slugs[product.Slug]; exists {
			add("duplicate slug already used by " + first)
		} else {
			slugs[product.Slug] = name
		}

		if strings.TrimSpace(product.PrimaryRepo) == "" {
			add("primary_repo is required")
		} else if !repoPattern.MatchString(product.PrimaryRepo) {
			add("primary_repo must use owner/name format")
		}
		if first, exists := primaryRepos[strings.ToLower(product.PrimaryRepo)]; exists {
			add("primary_repo is already used by " + first)
		} else {
			primaryRepos[strings.ToLower(product.PrimaryRepo)] = name
		}

		listedPrimary := false
		seenRepos := make(map[string]struct{})
		for _, repo := range product.Repositories {
			key := strings.ToLower(repo)
			if _, exists := seenRepos[key]; exists {
				add("duplicate repository " + repo)
			}
			seenRepos[key] = struct{}{}
			if strings.EqualFold(repo, product.PrimaryRepo) {
				listedPrimary = true
			}
			if !repoPattern.MatchString(repo) {
				add("repository must use owner/name format: " + repo)
			}
		}
		if !listedPrimary {
			add("primary_repo must be listed in repositories")
		}
		if strings.TrimSpace(product.Kind) == "" {
			add("kind is required")
		}
		if strings.TrimSpace(product.Domain) == "" {
			add("domain is required")
		}
		if strings.TrimSpace(product.Summary.EN) == "" {
			add("summary.en is required")
		}
		if strings.TrimSpace(product.Summary.RU) == "" {
			add("summary.ru is required")
		}
		if product.MaintainedFork && strings.TrimSpace(product.Upstream) == "" {
			add("upstream is required for maintained fork")
		}
		if product.Upstream != "" && !repoPattern.MatchString(product.Upstream) {
			add("upstream must use owner/name format")
		}
		validateHTTPS := func(field, value string) {
			if value == "" {
				return
			}
			parsed, err := url.Parse(value)
			if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
				add(field + " must use https")
			}
		}
		validateHTTPS("website", product.Website)
		validateHTTPS("documentation", product.Documentation)
	}

	if len(globalMessages) == 0 && len(productMessages) == 0 {
		return nil
	}
	sort.Strings(globalMessages)
	sort.Strings(productMessages)
	messages := append(globalMessages, productMessages...)
	errList := make([]error, 0, len(messages))
	for _, message := range messages {
		errList = append(errList, errors.New(message))
	}
	return errors.Join(errList...)
}
