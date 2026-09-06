package projectsite

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"strings"
)

//go:embed profiles/*.json
var marketingFiles embed.FS

type marketingItem struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type marketingCopy struct {
	Eyebrow  string          `json:"eyebrow"`
	Headline string          `json:"headline"`
	Intro    string          `json:"intro"`
	Features []marketingItem `json:"features"`
	Steps    []marketingItem `json:"steps"`
	Closing  string          `json:"closing"`
	CTA      string          `json:"cta"`
	Note     string          `json:"note"`
	Demo     string          `json:"demo,omitempty"`
}

type marketingProfile struct {
	Slug    string                   `json:"slug"`
	Name    string                   `json:"name"`
	Theme   string                   `json:"theme"`
	Visual  string                   `json:"visual"`
	Image   string                   `json:"image"`
	Code    string                   `json:"code"`
	Primary string                   `json:"primary"`
	Locales map[string]marketingCopy `json:"locales"`
}

type marketingView struct {
	marketingProfile
	Text         marketingCopy
	ImageURL     string
	Benefits     string
	Workflow     string
	Example      string
	DocsLabel    string
	ReportDay    string
	SignLabel    string
	HashLabel    string
	EncryptLabel string
}

func loadMarketing(slug string) (*marketingProfile, error) {
	data, err := marketingFiles.ReadFile("profiles/" + slug + ".json")
	if err != nil {
		if strings.Contains(slug, "/") || strings.Contains(slug, "\\") {
			return nil, fmt.Errorf("invalid marketing slug")
		}
		if !fs.ValidPath("profiles/" + slug + ".json") {
			return nil, fmt.Errorf("invalid marketing path")
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		return nil, nil
	}
	var profile marketingProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("decode marketing %s: %w", slug, err)
	}
	if profile.Slug != slug || profile.Name == "" || profile.Theme == "" || profile.Code == "" {
		return nil, fmt.Errorf("incomplete marketing identity: %s", slug)
	}
	if profile.Primary != "#install" && profile.Primary != "#overview" {
		return nil, fmt.Errorf("unsupported marketing CTA: %s", slug)
	}
	for _, locale := range []string{"en", "ru", "zh-cn"} {
		copy, ok := profile.Locales[locale]
		if !ok || copy.Headline == "" || copy.Intro == "" || copy.CTA == "" || copy.Closing == "" || len(copy.Features) < 3 || len(copy.Steps) < 2 {
			return nil, fmt.Errorf("incomplete %s marketing for %s", locale, slug)
		}
		for _, item := range append(copy.Features, copy.Steps...) {
			if item.Title == "" || item.Body == "" {
				return nil, fmt.Errorf("empty %s marketing item for %s", locale, slug)
			}
		}
	}
	return &profile, nil
}

func marketingFor(model Model, page LocalePage) (*marketingView, error) {
	profile, err := loadMarketing(model.Product.Slug)
	if err != nil || profile == nil {
		return nil, err
	}
	view := &marketingView{marketingProfile: *profile, Text: profile.Locales[page.Locale]}
	if view.Text.Demo != "" {
		view.Code = view.Text.Demo
	}
	if profile.Image != "" {
		view.ImageURL = "https://raw.githubusercontent.com/" + model.Repository.NameWithOwner + "/" + model.Repository.HeadSHA + "/" + profile.Image
	}
	labels := map[string][8]string{
		"en":    {"What you can do", "Get started", "Product example", "Read the documentation", "FRI", "Sign", "Hash", "Encrypt"},
		"ru":    {"Возможности", "С чего начать", "Пример работы", "Читать документацию", "ПТ", "Подпись", "Хеш", "Шифр"},
		"zh-cn": {"功能", "开始使用", "产品示例", "阅读文档", "周五", "签名", "哈希", "加密"},
	}
	label := labels[page.Locale]
	view.Benefits, view.Workflow, view.Example, view.DocsLabel = label[0], label[1], label[2], label[3]
	view.ReportDay, view.SignLabel, view.HashLabel, view.EncryptLabel = label[4], label[5], label[6], label[7]
	return view, nil
}

func validateMarketingHTML(slug, locale string, data []byte) error {
	profile, err := loadMarketing(slug)
	if err != nil || profile == nil {
		return err
	}
	copy := profile.Locales[locale]
	page := html.UnescapeString(string(data))
	required := []string{
		`data-product-profile="` + slug + `"`,
		`class="product-hero shell"`,
		"marketing.css", "<h1>" + copy.Headline + "</h1>",
		copy.Intro, copy.CTA, copy.Closing, copy.Note, `href="` + profile.Primary + `"`,
		`id="` + strings.TrimPrefix(profile.Primary, "#") + `"`,
	}
	if profile.Visual != "image" {
		code := profile.Code
		if copy.Demo != "" {
			code = copy.Demo
		}
		required = append(required, code)
	}
	if profile.Image != "" {
		required = append(required, "/"+profile.Image+`"`)
	}
	for _, item := range append(copy.Features, copy.Steps...) {
		required = append(required, item.Title, item.Body)
	}
	for _, value := range required {
		if !strings.Contains(page, value) {
			return fmt.Errorf("%s/%s missing marketing content: %s", slug, locale, value)
		}
	}
	return nil
}
