package githubapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

var semverPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)

func (c *Client) manifestVersion(ctx context.Context, repo catalog.Repository) (*catalog.Version, error) {
	for _, filename := range []string{"Cargo.toml", "package.json"} {
		var content apiContent
		path := fmt.Sprintf("/repos/%s/contents/%s?ref=%s", repo.NameWithOwner, filename, url.QueryEscape(repo.DefaultBranch))
		if _, err := c.get(ctx, path, &content); err != nil {
			if isNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("load %s from %s: %w", filename, repo.NameWithOwner, err)
		}
		source, err := decodeContent(content)
		if err != nil {
			return nil, fmt.Errorf("decode %s from %s: %w", filename, repo.NameWithOwner, err)
		}
		version := ""
		switch filename {
		case "Cargo.toml":
			version = cargoVersion(source)
		case "package.json":
			var manifest struct {
				Version string `json:"version"`
			}
			if err := json.Unmarshal(source, &manifest); err != nil {
				return nil, fmt.Errorf("parse package.json from %s: %w", repo.NameWithOwner, err)
			}
			version = manifest.Version
		}
		if semverPattern.MatchString(version) {
			return &catalog.Version{Value: version, Source: "manifest", URL: content.HTMLURL}, nil
		}
	}
	return nil, nil
}

func cargoVersion(source []byte) string {
	scanner := bufio.NewScanner(strings.NewReader(string(source)))
	inPackage := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inPackage = line == "[package]"
			continue
		}
		if !inPackage || !strings.HasPrefix(line, "version") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != "version" {
			continue
		}
		return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
	}
	return ""
}
