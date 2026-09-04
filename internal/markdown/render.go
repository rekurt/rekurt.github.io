package markdown

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rekurt/rekurt.github.io/internal/catalog"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

const (
	maxSourceBytes = 256 << 10
	maxHTMLBytes   = 512 << 10
)

func RenderREADME(source []byte, repo, branch, sha string) (catalog.Readme, error) {
	if len(source) > maxSourceBytes {
		return catalog.Readme{}, fmt.Errorf("README exceeds %d bytes", maxSourceBytes)
	}
	if strings.Count(repo, "/") != 1 {
		return catalog.Readme{}, fmt.Errorf("repository must use owner/name format")
	}
	ref := sha
	if ref == "" {
		ref = branch
	}
	if ref == "" {
		return catalog.Readme{}, fmt.Errorf("README requires a branch or commit SHA")
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
	)
	reader := text.NewReader(source)
	document := markdown.Parser().Parse(reader)
	if err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Link:
			typed.Destination = rewriteDestination(typed.Destination, repo, ref, false)
		case *ast.Image:
			typed.Destination = rewriteDestination(typed.Destination, repo, ref, true)
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return catalog.Readme{}, fmt.Errorf("walk README: %w", err)
	}

	var rendered bytes.Buffer
	if err := markdown.Renderer().Render(&rendered, source, document); err != nil {
		return catalog.Readme{}, fmt.Errorf("render README: %w", err)
	}

	policy := bluemonday.UGCPolicy()
	policy.AllowTables()
	policy.AllowAttrs("class").OnElements("code")
	policy.AllowURLSchemes("http", "https", "mailto")
	policy.RequireParseableURLs(true)
	policy.AllowRelativeURLs(true)
	policy.RequireNoReferrerOnFullyQualifiedLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)
	policy.SkipElementsContent("script", "style", "iframe", "form", "svg")
	safeHTML := policy.SanitizeBytes(rendered.Bytes())
	if len(safeHTML) > maxHTMLBytes {
		return catalog.Readme{}, fmt.Errorf("rendered README exceeds %d bytes", maxHTMLBytes)
	}

	return catalog.Readme{
		HTML:      string(safeHTML),
		SourceURL: fmt.Sprintf("https://github.com/%s/blob/%s/README.md", repo, ref),
		SHA:       sha,
	}, nil
}

func rewriteDestination(destination []byte, repo, ref string, image bool) []byte {
	value := string(destination)
	if value == "" || strings.HasPrefix(value, "#") {
		return destination
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return destination
	}

	cleaned := strings.TrimPrefix(path.Clean("/"+parsed.Path), "/")
	segments := strings.Split(cleaned, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	escapedPath := strings.Join(segments, "/")
	base := fmt.Sprintf("https://github.com/%s/blob/%s/", repo, ref)
	if image {
		base = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/", repo, ref)
	}
	rewritten := base + escapedPath
	if parsed.RawQuery != "" {
		rewritten += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		rewritten += "#" + parsed.Fragment
	}
	return []byte(rewritten)
}
