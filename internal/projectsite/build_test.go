package projectsite

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestBuildWritesCompleteMultilingualSite(t *testing.T) {
	output := filepath.Join(t.TempDir(), "site")
	manifest, err := Build(fixtureOptions(t, output))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != "build" || manifest.Slug != "git-barber" || manifest.SourceSHA != "abc123" {
		t.Fatalf("manifest = %#v", manifest)
	}
	wantFiles := []string{
		".nojekyll", "404.html", "assets/family.css", "assets/family.js", "family-build.json",
		"index.html", "projects/index.html", "robots.txt", "ru/index.html", "ru/projects/index.html",
		"sitemap.xml", "zh-cn/index.html", "zh-cn/projects/index.html",
	}
	for _, name := range wantFiles {
		if info, err := os.Stat(filepath.Join(output, name)); err != nil || info.IsDir() {
			t.Errorf("generated file %s: info=%v err=%v", name, info, err)
		}
	}

	english := readFile(t, filepath.Join(output, "index.html"))
	document, err := html.Parse(strings.NewReader(english))
	if err != nil {
		t.Fatal(err)
	}
	if countElements(document, "h1") != 1 {
		t.Fatalf("h1 count = %d", countElements(document, "h1"))
	}
	for _, required := range []string{
		`<html lang="en">`,
		`rel="canonical" href="https://rekurt.github.io/git-barber/"`,
		`hreflang="ru" href="https://rekurt.github.io/git-barber/ru/"`,
		`hreflang="zh-CN" href="https://rekurt.github.io/git-barber/zh-cn/"`,
		`hreflang="x-default" href="https://rekurt.github.io/git-barber/"`,
		`application/ld+json`, `SoftwareSourceCode`, `brew install rekurt/tap/git-barber`,
		`v0.3.0`, `MIT OR Apache-2.0`, `https://rekurt.github.io/`,
		`href="projects/"`, `English documentation`,
	} {
		if !strings.Contains(english, required) {
			t.Errorf("English page lacks %q", required)
		}
	}
	if strings.Contains(english, "alert(1)") {
		t.Fatal("sanitized README script content leaked into output")
	}

	chinese := readFile(t, filepath.Join(output, "zh-cn/index.html"))
	for _, required := range []string{`<html lang="zh-CN">`, `安全清理 Git 分支。`, `全部项目`, `href="projects/"`} {
		if !strings.Contains(chinese, required) {
			t.Errorf("Chinese page lacks %q", required)
		}
	}
	if strings.Contains(chinese, "English documentation") {
		t.Fatal("Chinese page contains untranslated English README")
	}

	directory := readFile(t, filepath.Join(output, "projects/index.html"))
	if !strings.Contains(directory, "https://rekurt.github.io/ymsdk/") || strings.Contains(directory, `href="https://rekurt.github.io/git-barber/"`) {
		t.Fatalf("sibling directory links are incomplete or self-referential: %s", directory)
	}
	sitemap := readFile(t, filepath.Join(output, "sitemap.xml"))
	for _, required := range []string{
		"https://rekurt.github.io/git-barber/",
		"https://rekurt.github.io/git-barber/ru/",
		"https://rekurt.github.io/git-barber/zh-cn/",
		`hreflang="zh-CN"`,
	} {
		if !strings.Contains(sitemap, required) {
			t.Errorf("sitemap lacks %q", required)
		}
	}
}

func TestBuildIsDeterministicForFixedInputs(t *testing.T) {
	first := filepath.Join(t.TempDir(), "first")
	second := filepath.Join(t.TempDir(), "second")
	if _, err := Build(fixtureOptions(t, first)); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(fixtureOptions(t, second)); err != nil {
		t.Fatal(err)
	}
	if left, right := treeDigest(t, first), treeDigest(t, second); left != right {
		t.Fatalf("tree digests differ: %s != %s", left, right)
	}
}

func TestBuildRejectsOutputSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	output := filepath.Join(root, "site")
	if err := os.Symlink(outside, output); err != nil {
		t.Fatal(err)
	}
	_, err := Build(fixtureOptions(t, output))
	if err == nil || !strings.Contains(err.Error(), "output must not be a symlink") {
		t.Fatalf("error = %v", err)
	}
}

func fixtureOptions(t *testing.T, output string) Options {
	t.Helper()
	repository, err := filepath.Abs("testdata/repository")
	if err != nil {
		t.Fatal(err)
	}
	return Options{
		Slug: "git-barber", SnapshotPath: "testdata/snapshot.json", Repository: repository,
		Output: output, BaseURL: "https://rekurt.github.io/git-barber/",
		GeneratedAt: time.Date(2026, 9, 5, 6, 30, 0, 0, time.UTC),
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func countElements(node *html.Node, name string) int {
	count := 0
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.ElementNode && current.Data == name {
			count++
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return count
}

func treeDigest(t *testing.T, root string) string {
	t.Helper()
	var names []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			name, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			names = append(names, name)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
