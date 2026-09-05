package catalog

import "time"

type LocalizedText struct {
	EN   string `yaml:"en" json:"en"`
	RU   string `yaml:"ru" json:"ru"`
	ZHCN string `yaml:"zh-cn" json:"zhCN"`
}

type ProductConfig struct {
	Slug           string        `yaml:"slug"`
	PrimaryRepo    string        `yaml:"primary_repo"`
	Repositories   []string      `yaml:"repositories"`
	Kind           string        `yaml:"kind"`
	Domain         string        `yaml:"domain"`
	Accent         string        `yaml:"accent"`
	Featured       bool          `yaml:"featured"`
	MaintainedFork bool          `yaml:"maintained_fork"`
	Upstream       string        `yaml:"upstream"`
	Summary        LocalizedText `yaml:"summary"`
	Install        []string      `yaml:"install"`
	Website        string        `yaml:"website"`
	Documentation  string        `yaml:"documentation"`
}

type Manifest struct {
	Owner    string          `yaml:"owner"`
	Products []ProductConfig `yaml:"products"`
}

type Version struct {
	Value  string `json:"value"`
	Source string `json:"source"`
	URL    string `json:"url,omitempty"`
}

type Readme struct {
	HTML      string `json:"html,omitempty"`
	SourceURL string `json:"sourceUrl,omitempty"`
	SHA       string `json:"sha,omitempty"`
	Source    string `json:"-"`
}

type Link struct {
	Kind string `json:"kind"`
	URL  string `json:"url"`
}

type Repository struct {
	NameWithOwner string    `json:"nameWithOwner"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	URL           string    `json:"url"`
	Visibility    string    `json:"visibility"`
	Fork          bool      `json:"fork"`
	Parent        string    `json:"parent,omitempty"`
	Language      string    `json:"language,omitempty"`
	Topics        []string  `json:"topics,omitempty"`
	Homepage      string    `json:"homepage,omitempty"`
	HasPages      bool      `json:"hasPages"`
	DefaultBranch string    `json:"defaultBranch"`
	HeadSHA       string    `json:"headSha,omitempty"`
	License       string    `json:"license,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt"`
	PushedAt      time.Time `json:"pushedAt"`
	Archived      bool      `json:"archived"`
	Stars         int       `json:"stars"`
	Role          string    `json:"role,omitempty"`
	Version       *Version  `json:"version,omitempty"`
	Readme        *Readme   `json:"readme,omitempty"`
	Links         []Link    `json:"links,omitempty"`
}

type Product struct {
	Slug           string        `json:"slug"`
	PrimaryRepo    string        `json:"primaryRepo"`
	Repositories   []string      `json:"repositories"`
	Kind           string        `json:"kind"`
	Domain         string        `json:"domain"`
	Accent         string        `json:"accent"`
	Featured       bool          `json:"featured"`
	MaintainedFork bool          `json:"maintainedFork"`
	Upstream       string        `json:"upstream,omitempty"`
	Summary        LocalizedText `json:"summary"`
	Install        []string      `json:"install,omitempty"`
	Links          []Link        `json:"links,omitempty"`
	Version        *Version      `json:"version,omitempty"`
	Readme         *Readme       `json:"readme,omitempty"`
}

type Snapshot struct {
	SchemaVersion int          `json:"schemaVersion"`
	Owner         string       `json:"owner"`
	SyncedAt      time.Time    `json:"syncedAt"`
	Products      []Product    `json:"products"`
	Repositories  []Repository `json:"repositories"`
}
