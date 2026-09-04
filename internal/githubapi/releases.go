package githubapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

type apiRelease struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
}

type apiTag struct {
	Name string `json:"name"`
}

type apiBranch struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

type apiContent struct {
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	HTMLURL  string `json:"html_url"`
	SHA      string `json:"sha"`
}

func (c *Client) Enrich(ctx context.Context, repo catalog.Repository, withReadme bool) (catalog.Repository, error) {
	fullName := repo.NameWithOwner
	var detail apiRepository
	if _, err := c.get(ctx, "/repos/"+fullName, &detail); err != nil {
		return catalog.Repository{}, fmt.Errorf("load repository %s: %w", fullName, err)
	}
	repo = mapRepository(detail)

	version, err := c.latestVersion(ctx, repo)
	if err != nil {
		return catalog.Repository{}, err
	}
	repo.Version = version

	if repo.DefaultBranch != "" {
		var branch apiBranch
		branchPath := fmt.Sprintf("/repos/%s/branches/%s", fullName, url.PathEscape(repo.DefaultBranch))
		if _, err := c.get(ctx, branchPath, &branch); err != nil {
			if !isNotFound(err) {
				return catalog.Repository{}, fmt.Errorf("load branch %s: %w", fullName, err)
			}
		} else {
			repo.HeadSHA = branch.Commit.SHA
		}
	}

	if withReadme {
		var content apiContent
		readmePath := fmt.Sprintf("/repos/%s/readme?ref=%s", fullName, url.QueryEscape(repo.DefaultBranch))
		if _, err := c.get(ctx, readmePath, &content); err != nil {
			if !isNotFound(err) {
				return catalog.Repository{}, fmt.Errorf("load README %s: %w", fullName, err)
			}
		} else {
			source, err := decodeContent(content)
			if err != nil {
				return catalog.Repository{}, fmt.Errorf("decode README %s: %w", fullName, err)
			}
			sourceURL := content.HTMLURL
			if repo.HeadSHA != "" {
				sourceURL = fmt.Sprintf("https://github.com/%s/blob/%s/README.md", fullName, repo.HeadSHA)
			}
			repo.Readme = &catalog.Readme{SourceURL: sourceURL, SHA: repo.HeadSHA, Source: string(source)}
		}
	}

	return repo, nil
}

func (c *Client) latestVersion(ctx context.Context, repo catalog.Repository) (*catalog.Version, error) {
	fullName := repo.NameWithOwner
	var releases []apiRelease
	if _, err := c.get(ctx, "/repos/"+fullName+"/releases?per_page=20", &releases); err != nil {
		return nil, fmt.Errorf("load releases %s: %w", fullName, err)
	}
	for _, release := range releases {
		if !release.Draft && !release.Prerelease && semverPattern.MatchString(release.TagName) {
			return &catalog.Version{Value: release.TagName, Source: "release", URL: release.HTMLURL}, nil
		}
	}

	var tags []apiTag
	if _, err := c.get(ctx, "/repos/"+fullName+"/tags?per_page=20", &tags); err != nil {
		return nil, fmt.Errorf("load tags %s: %w", fullName, err)
	}
	for _, tag := range tags {
		if semverPattern.MatchString(tag.Name) {
			return &catalog.Version{
				Value:  tag.Name,
				Source: "tag",
				URL:    fmt.Sprintf("https://github.com/%s/releases/tag/%s", fullName, url.PathEscape(tag.Name)),
			}, nil
		}
	}

	return c.manifestVersion(ctx, repo)
}

func decodeContent(content apiContent) ([]byte, error) {
	if content.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported encoding %q", content.Encoding)
	}
	encoded := strings.ReplaceAll(content.Content, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
