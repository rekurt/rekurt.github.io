package projectsite

import (
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

type Options struct {
	Slug         string
	SnapshotPath string
	Repository   string
	Output       string
	BaseURL      string
	GeneratedAt  time.Time
}

type LocalePage struct {
	Locale          string
	Lang            string
	Path            string
	Canonical       string
	Title           string
	Description     string
	Summary         string
	ReadmeHTML      string
	ReadmeSourceURL string
}

type Model struct {
	Owner          string
	Product        catalog.Product
	Repository     catalog.Repository
	Products       []catalog.Product
	Pages          []LocalePage
	BaseURL        string
	Output         string
	RepositoryRoot string
	GeneratedAt    time.Time
}

type BuildManifest struct {
	SchemaVersion int       `json:"schemaVersion"`
	Slug          string    `json:"slug"`
	SourceSHA     string    `json:"sourceSha"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Mode          string    `json:"mode"`
}
