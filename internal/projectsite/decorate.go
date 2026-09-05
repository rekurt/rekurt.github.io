package projectsite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Decorate(options Options) (BuildManifest, error) {
	info, err := os.Lstat(options.Output)
	if err != nil {
		return BuildManifest{}, fmt.Errorf("inspect output: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return BuildManifest{}, fmt.Errorf("output must not be a symlink")
	}
	if !info.IsDir() {
		return BuildManifest{}, fmt.Errorf("output must be a directory")
	}
	indexPath := filepath.Join(options.Output, "index.html")
	if info, err := os.Lstat(indexPath); err != nil || !info.Mode().IsRegular() {
		return BuildManifest{}, fmt.Errorf("existing site index.html is required")
	}

	model, err := Resolve(options)
	if err != nil {
		return BuildManifest{}, err
	}
	if err := decorateIndex(model, indexPath); err != nil {
		return BuildManifest{}, err
	}
	if err := renderFamilyDirectories(model, model.Output); err != nil {
		return BuildManifest{}, err
	}
	for _, name := range []string{"family.css", "bridge.css"} {
		data, err := fs.ReadFile(siteFiles, "assets/"+name)
		if err != nil {
			return BuildManifest{}, err
		}
		if err := writeSiteFile(model.Output, "assets/"+name, data); err != nil {
			return BuildManifest{}, err
		}
	}
	if err := writeSiteFile(model.Output, "family-sitemap.xml", []byte(renderSitemap(model))); err != nil {
		return BuildManifest{}, err
	}
	if err := updateRobots(model); err != nil {
		return BuildManifest{}, err
	}
	manifest := BuildManifest{
		SchemaVersion: 1, Slug: model.Product.Slug, SourceSHA: model.Repository.HeadSHA,
		GeneratedAt: model.GeneratedAt, Mode: "decorate",
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return BuildManifest{}, err
	}
	if err := writeSiteFile(model.Output, "family-build.json", append(data, '\n')); err != nil {
		return BuildManifest{}, err
	}
	return manifest, nil
}

func decorateIndex(model Model, indexPath string) error {
	source, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	document, err := nethtml.Parse(bytes.NewReader(source))
	if err != nil {
		return fmt.Errorf("parse existing index.html: %w", err)
	}
	head := findElement(document, "head")
	body := findElement(document, "body")
	if head == nil || body == nil {
		return fmt.Errorf("existing index.html requires head and body")
	}
	removeMatching(document, func(node *nethtml.Node) bool {
		return hasAttribute(node, "data-rekurt-family") || hasAttribute(node, "data-rekurt-schema") || hasAttribute(node, "data-rekurt-meta") || isCanonical(node)
	})

	page := model.Pages[0]
	appendChild(head, element("link", attr("rel", "canonical"), attr("href", page.Canonical)))
	appendChild(head, element("link", attr("rel", "stylesheet"), attr("href", "assets/bridge.css"), attr("data-rekurt-family", "")))
	ensureMeta(head, "name", "description", page.Description)
	ensureMeta(head, "name", "generator", "rekurt project family kit")
	ensureMeta(head, "property", "og:url", page.Canonical)
	ensureMeta(head, "property", "og:description", page.Description)
	data, err := structuredData(model, page)
	if err != nil {
		return err
	}
	script := element("script", attr("type", "application/ld+json"), attr("data-rekurt-schema", ""))
	appendChild(script, &nethtml.Node{Type: nethtml.TextNode, Data: string(data)})
	appendChild(head, script)

	bar := familyBar(model.Owner)
	if body.FirstChild == nil {
		appendChild(body, bar)
	} else {
		body.InsertBefore(bar, body.FirstChild)
	}
	var output bytes.Buffer
	if err := nethtml.Render(&output, document); err != nil {
		return fmt.Errorf("render decorated index.html: %w", err)
	}
	return writeAtomicFile(indexPath, output.Bytes())
}

func renderFamilyDirectories(model Model, output string) error {
	directoryTemplate, err := template.ParseFS(siteFiles, "templates/directory.html")
	if err != nil {
		return err
	}
	for _, page := range model.Pages {
		var rendered bytes.Buffer
		if err := directoryTemplate.ExecuteTemplate(&rendered, "directory.html", makeDirectoryView(model, page)); err != nil {
			return err
		}
		name := routeFile(strings.TrimSuffix(page.Path, "/") + "/projects/")
		if err := writeSiteFile(output, name, rendered.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func updateRobots(model Model) error {
	path := filepath.Join(model.Output, "robots.txt")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		data = []byte("User-agent: *\nAllow: /\n")
	} else if err != nil {
		return err
	}
	line := "Sitemap: " + model.BaseURL + "family-sitemap.xml"
	text := strings.TrimRight(string(data), "\n")
	if !strings.Contains(text, line) {
		text += "\n" + line
	}
	return writeAtomicFile(path, []byte(text+"\n"))
}

func familyBar(owner string) *nethtml.Node {
	header := element("header", attr("class", "rk-family-bar"), attr("data-rekurt-family", ""))
	nav := element("nav", attr("class", "rk-family-nav"), attr("aria-label", "Project family navigation"))
	brand := element("a", attr("class", "rk-family-brand"), attr("href", "https://rekurt.github.io/"))
	mark := element("span", attr("class", "rk-family-mark"), attr("aria-hidden", "true"))
	appendChild(mark, textNode("R"))
	label := element("span", attr("class", "rk-family-label"))
	appendChild(label, textNode("rekurt / systems"))
	appendChild(brand, mark)
	appendChild(brand, label)
	links := element("div", attr("class", "rk-family-links"))
	projects := element("a", attr("href", "projects/"))
	appendChild(projects, textNode("Projects"))
	github := element("a", attr("href", "https://github.com/"+owner))
	appendChild(github, textNode("GitHub"))
	appendChild(links, projects)
	appendChild(links, github)
	appendChild(nav, brand)
	appendChild(nav, links)
	appendChild(header, nav)
	return header
}

func ensureMeta(head *nethtml.Node, key, keyValue, content string) {
	for child := head.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == nethtml.ElementNode && child.Data == "meta" && attributeValue(child, key) == keyValue {
			return
		}
	}
	appendChild(head, element("meta", attr(key, keyValue), attr("content", content), attr("data-rekurt-meta", "")))
}

func removeMatching(root *nethtml.Node, match func(*nethtml.Node) bool) {
	for child := root.FirstChild; child != nil; {
		next := child.NextSibling
		if match(child) {
			root.RemoveChild(child)
		} else {
			removeMatching(child, match)
		}
		child = next
	}
}

func isCanonical(node *nethtml.Node) bool {
	return node.Type == nethtml.ElementNode && node.Data == "link" && strings.EqualFold(attributeValue(node, "rel"), "canonical")
}

func findElement(root *nethtml.Node, name string) *nethtml.Node {
	if root.Type == nethtml.ElementNode && root.Data == name {
		return root
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, name); found != nil {
			return found
		}
	}
	return nil
}

func hasAttribute(node *nethtml.Node, name string) bool {
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return true
		}
	}
	return false
}

func attributeValue(node *nethtml.Node, name string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func element(name string, attributes ...nethtml.Attribute) *nethtml.Node {
	return &nethtml.Node{Type: nethtml.ElementNode, Data: name, DataAtom: atom.Lookup([]byte(name)), Attr: attributes}
}

func attr(key, value string) nethtml.Attribute {
	return nethtml.Attribute{Key: key, Val: value}
}

func textNode(value string) *nethtml.Node {
	return &nethtml.Node{Type: nethtml.TextNode, Data: value}
}

func appendChild(parent, child *nethtml.Node) {
	parent.AppendChild(child)
}

func writeAtomicFile(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".project-site-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
