package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func TestWriteSnapshotIsStableAndIgnoresEmptySyncTime(t *testing.T) {
	path := filepath.Join(t.TempDir(), "generated", "catalog.json")
	firstTime := time.Date(2026, 9, 5, 1, 2, 3, 0, time.UTC)
	first := unorderedSnapshot(firstTime)
	changed, err := WriteSnapshot(path, first)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first write reported unchanged")
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(before), "\n") {
		t.Fatal("snapshot lacks trailing newline")
	}

	second := unorderedSnapshot(firstTime.Add(time.Hour))
	second.Repositories[0], second.Repositories[1] = second.Repositories[1], second.Repositories[0]
	changed, err = WriteSnapshot(path, second)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("timestamp-only sync reported changed")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("timestamp-only sync changed bytes")
	}
	if entries, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".catalog-*")); err != nil || len(entries) != 0 {
		t.Fatalf("temporary files = %#v, error = %v", entries, err)
	}
}

func TestWriteSnapshotUpdatesMaterialChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.json")
	first := unorderedSnapshot(time.Date(2026, 9, 5, 1, 0, 0, 0, time.UTC))
	if _, err := WriteSnapshot(path, first); err != nil {
		t.Fatal(err)
	}
	second := unorderedSnapshot(first.SyncedAt.Add(time.Hour))
	second.Repositories[0].Stars++
	changed, err := WriteSnapshot(path, second)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("material change reported unchanged")
	}
	got, err := ReadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.SyncedAt.Equal(second.SyncedAt) {
		t.Fatalf("syncedAt = %s, want %s", got.SyncedAt, second.SyncedAt)
	}
}

func TestSnapshotComparisonIgnoresSelfReferentialHubCommitFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.json")
	first := unorderedSnapshot(time.Date(2026, 9, 5, 1, 0, 0, 0, time.UTC))
	first.Repositories[0].Role = "portfolio-hub"
	first.Repositories[0].HeadSHA = "first"
	if _, err := WriteSnapshot(path, first); err != nil {
		t.Fatal(err)
	}

	second := unorderedSnapshot(first.SyncedAt.Add(time.Hour))
	second.Repositories[0].Role = "portfolio-hub"
	second.Repositories[0].HeadSHA = "second"
	second.Repositories[0].UpdatedAt = second.Repositories[0].UpdatedAt.Add(time.Hour)
	second.Repositories[0].PushedAt = second.Repositories[0].PushedAt.Add(time.Hour)
	changed, err := SnapshotChanged(path, second)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("portfolio hub commit metadata caused a self-referential change")
	}
}

func TestRenderAuditSummarizesRepositories(t *testing.T) {
	snapshot := unorderedSnapshot(time.Date(2026, 9, 5, 1, 2, 3, 0, time.UTC))
	report := string(RenderAudit(snapshot))
	for _, want := range []string{"2 public repositories", "1 original", "1 fork", "[rekurt/a](https://github.com/rekurt/a)", "v1.0.0", "2026-09-05T01:02:03Z"} {
		if !strings.Contains(report, want) {
			t.Fatalf("report lacks %q:\n%s", want, report)
		}
	}
	for _, line := range strings.Split(report, "\n") {
		if strings.HasSuffix(line, " ") {
			t.Fatalf("report contains trailing whitespace: %q", line)
		}
	}
}

func unorderedSnapshot(syncedAt time.Time) catalog.Snapshot {
	repositoryTime := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	return catalog.Snapshot{
		SchemaVersion: 1,
		Owner:         "rekurt",
		SyncedAt:      syncedAt,
		Products: []catalog.Product{{
			Slug: "a", PrimaryRepo: "rekurt/a", Repositories: []string{"rekurt/tap", "rekurt/a"},
			Links: []catalog.Link{{Kind: "source", URL: "https://github.com/rekurt/a"}, {Kind: "website", URL: "https://rekurt.github.io/a/"}},
		}},
		Repositories: []catalog.Repository{
			{NameWithOwner: "rekurt/z", Name: "z", URL: "https://github.com/rekurt/z", Visibility: "public", Fork: true, Topics: []string{"z", "a"}, UpdatedAt: repositoryTime, PushedAt: repositoryTime},
			{NameWithOwner: "rekurt/a", Name: "a", URL: "https://github.com/rekurt/a", Visibility: "public", Fork: false, Role: "primary", Version: &catalog.Version{Value: "v1.0.0", Source: "tag"}, UpdatedAt: repositoryTime, PushedAt: repositoryTime},
		},
	}
}
