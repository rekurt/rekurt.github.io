package githubapi

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

type apiRepository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Visibility  string `json:"visibility"`
	Fork        bool   `json:"fork"`
	Parent      *struct {
		FullName string `json:"full_name"`
	} `json:"parent"`
	Language      string   `json:"language"`
	Topics        []string `json:"topics"`
	Homepage      string   `json:"homepage"`
	HasPages      bool     `json:"has_pages"`
	DefaultBranch string   `json:"default_branch"`
	License       *struct {
		SPDXID string `json:"spdx_id"`
	} `json:"license"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	Archived        bool      `json:"archived"`
	StargazersCount int       `json:"stargazers_count"`
}

func (c *Client) ListOwnedPublic(ctx context.Context, owner string) ([]catalog.Repository, error) {
	path := fmt.Sprintf("/users/%s/repos?type=owner&sort=full_name&direction=asc&per_page=100&page=1", url.PathEscape(owner))
	var repositories []catalog.Repository
	for path != "" {
		var page []apiRepository
		headers, err := c.get(ctx, path, &page)
		if err != nil {
			return nil, err
		}
		for _, repo := range page {
			if repo.Private || repo.Visibility != "public" {
				continue
			}
			repositories = append(repositories, mapRepository(repo))
		}
		path = nextLink(headers.Get("Link"))
	}
	sort.Slice(repositories, func(i, j int) bool {
		return strings.ToLower(repositories[i].NameWithOwner) < strings.ToLower(repositories[j].NameWithOwner)
	})
	return repositories, nil
}

func mapRepository(repo apiRepository) catalog.Repository {
	parent := ""
	if repo.Parent != nil {
		parent = repo.Parent.FullName
	}
	license := ""
	if repo.License != nil && repo.License.SPDXID != "NOASSERTION" {
		license = repo.License.SPDXID
	}
	return catalog.Repository{
		NameWithOwner: repo.FullName,
		Name:          repo.Name,
		Description:   repo.Description,
		URL:           repo.HTMLURL,
		Visibility:    repo.Visibility,
		Fork:          repo.Fork,
		Parent:        parent,
		Language:      repo.Language,
		Topics:        repo.Topics,
		Homepage:      repo.Homepage,
		HasPages:      repo.HasPages,
		DefaultBranch: repo.DefaultBranch,
		License:       license,
		UpdatedAt:     repo.UpdatedAt,
		PushedAt:      repo.PushedAt,
		Archived:      repo.Archived,
		Stars:         repo.StargazersCount,
	}
}

func nextLink(header string) string {
	for _, part := range strings.Split(header, ",") {
		sections := strings.Split(strings.TrimSpace(part), ";")
		if len(sections) < 2 || !strings.Contains(strings.Join(sections[1:], ";"), `rel="next"`) {
			continue
		}
		return strings.Trim(strings.TrimSpace(sections[0]), "<>")
	}
	return ""
}
