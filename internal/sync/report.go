package sync

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func RenderAudit(snapshot catalog.Snapshot) []byte {
	originals := 0
	forks := 0
	for _, repo := range snapshot.Repositories {
		if repo.Fork {
			forks++
		} else {
			originals++
		}
	}

	var output bytes.Buffer
	fmt.Fprintln(&output, "# Public repository audit")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "Snapshot: %s\n", snapshot.SyncedAt.UTC().Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(&output, "%d public repositories · %d original · %d fork\n\n", len(snapshot.Repositories), originals, forks)
	fmt.Fprintln(&output, "| Repository | Role | Origin | Language | Version | Website | Documentation | Last push |")
	fmt.Fprintln(&output, "|---|---|---|---|---|---|---|---|")
	for _, repo := range snapshot.Repositories {
		origin := "original"
		if repo.Fork {
			origin = "fork"
			if repo.Parent != "" {
				origin += " of `" + escapeCell(repo.Parent) + "`"
			}
		}
		version := "—"
		if repo.Version != nil {
			version = escapeCell(repo.Version.Value)
		}
		website := markdownLink(linkURL(repo.Links, "website"))
		documentation := markdownLink(linkURL(repo.Links, "documentation"))
		fmt.Fprintf(&output, "| [%s](%s) | %s | %s | %s | %s | %s | %s | %s |\n",
			escapeCell(repo.NameWithOwner), repo.URL, fallback(repo.Role), origin, fallback(repo.Language), version,
			website, documentation, repo.PushedAt.UTC().Format("2006-01-02"))
	}
	return output.Bytes()
}

func linkURL(links []catalog.Link, kind string) string {
	for _, link := range links {
		if link.Kind == kind {
			return link.URL
		}
	}
	return ""
}

func markdownLink(raw string) string {
	if raw == "" {
		return "—"
	}
	return "[open](" + raw + ")"
}

func fallback(value string) string {
	if value == "" {
		return "—"
	}
	return escapeCell(value)
}

func escapeCell(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "|", "\\|"), "\n", " ")
}
