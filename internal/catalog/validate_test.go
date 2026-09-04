package catalog

import (
	"strings"
	"testing"
)

func TestValidateManifestAcceptsValidManifest(t *testing.T) {
	m, err := LoadManifest("testdata/valid.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateManifest(m); err != nil {
		t.Fatalf("ValidateManifest() error = %v", err)
	}
}

func TestValidateManifestRejectsInvalidProducts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Manifest)
		want   string
	}{
		{
			name:   "duplicate slug",
			mutate: func(m *Manifest) { m.Products = append(m.Products, m.Products[0]) },
			want:   "duplicate slug",
		},
		{
			name:   "invalid slug",
			mutate: func(m *Manifest) { m.Products[0].Slug = "Mac_Coffee" },
			want:   "invalid slug",
		},
		{
			name:   "missing translation",
			mutate: func(m *Manifest) { m.Products[0].Summary.RU = "" },
			want:   "summary.ru is required",
		},
		{
			name:   "fork without upstream",
			mutate: func(m *Manifest) { m.Products[0].Upstream = "" },
			want:   "upstream is required",
		},
		{
			name:   "unsafe website",
			mutate: func(m *Manifest) { m.Products[0].Website = "javascript:alert(1)" },
			want:   "website must use https",
		},
		{
			name:   "primary outside repositories",
			mutate: func(m *Manifest) { m.Products[0].PrimaryRepo = "rekurt/other" },
			want:   "primary_repo must be listed in repositories",
		},
		{
			name: "duplicate primary repository",
			mutate: func(m *Manifest) {
				other := m.Products[0]
				other.Slug = "other"
				m.Products = append(m.Products, other)
			},
			want: "primary_repo is already used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validManifest(t)
			tt.mutate(&m)
			err := ValidateManifest(m)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestValidateManifestSortsErrors(t *testing.T) {
	m := validManifest(t)
	m.Owner = ""
	m.Products[0].Summary.RU = ""
	m.Products[0].Summary.EN = ""
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("ValidateManifest() error = nil")
	}
	want := "manifest owner is required\nmac-coffee: summary.en is required\nmac-coffee: summary.ru is required"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err, want)
	}
}

func validManifest(t *testing.T) Manifest {
	t.Helper()
	m, err := LoadManifest("testdata/valid.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return m
}
