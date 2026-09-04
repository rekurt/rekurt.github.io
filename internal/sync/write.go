package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
)

func WriteSnapshot(path string, snapshot catalog.Snapshot) (bool, error) {
	normalized, err := normalizeSnapshot(snapshot)
	if err != nil {
		return false, err
	}

	if existing, err := ReadSnapshot(path); err == nil {
		normalizedExisting, err := normalizeSnapshot(existing)
		if err != nil {
			return false, fmt.Errorf("normalize existing snapshot: %w", err)
		}
		candidateTime := normalized.SyncedAt
		normalized.SyncedAt = time.Time{}
		normalizedExisting.SyncedAt = time.Time{}
		if reflect.DeepEqual(normalizedExisting, normalized) {
			return false, nil
		}
		normalized.SyncedAt = candidateTime
	} else if !os.IsNotExist(err) {
		return false, err
	}

	data, err := encodeSnapshot(normalized)
	if err != nil {
		return false, err
	}
	if err := writeAtomic(path, data); err != nil {
		return false, err
	}
	return true, nil
}

func SnapshotChanged(path string, snapshot catalog.Snapshot) (bool, error) {
	existing, err := ReadSnapshot(path)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	normalizedExisting, err := normalizeSnapshot(existing)
	if err != nil {
		return false, err
	}
	normalizedCandidate, err := normalizeSnapshot(snapshot)
	if err != nil {
		return false, err
	}
	normalizedExisting.SyncedAt = time.Time{}
	normalizedCandidate.SyncedAt = time.Time{}
	return !reflect.DeepEqual(normalizedExisting, normalizedCandidate), nil
}

func ReadSnapshot(path string) (catalog.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return catalog.Snapshot{}, err
	}
	var snapshot catalog.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return catalog.Snapshot{}, fmt.Errorf("decode snapshot %s: %w", path, err)
	}
	return snapshot, nil
}

func WriteBytesIfChanged(path string, data []byte) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return false, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err := writeAtomic(path, data); err != nil {
		return false, err
	}
	return true, nil
}

func normalizeSnapshot(snapshot catalog.Snapshot) (catalog.Snapshot, error) {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return catalog.Snapshot{}, fmt.Errorf("clone snapshot: %w", err)
	}
	var clone catalog.Snapshot
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return catalog.Snapshot{}, fmt.Errorf("decode cloned snapshot: %w", err)
	}
	clone.SyncedAt = clone.SyncedAt.UTC().Truncate(time.Second)
	for i := range clone.Products {
		sort.Strings(clone.Products[i].Repositories)
		sort.Strings(clone.Products[i].Install)
		sortLinks(clone.Products[i].Links)
	}
	for i := range clone.Repositories {
		sort.Strings(clone.Repositories[i].Topics)
		sortLinks(clone.Repositories[i].Links)
	}
	sort.Slice(clone.Repositories, func(i, j int) bool {
		return strings.ToLower(clone.Repositories[i].NameWithOwner) < strings.ToLower(clone.Repositories[j].NameWithOwner)
	})
	return clone, nil
}

func sortLinks(links []catalog.Link) {
	order := map[string]int{"website": 0, "documentation": 1, "source": 2, "release": 3}
	sort.Slice(links, func(i, j int) bool {
		left, right := order[links[i].Kind], order[links[j].Kind]
		if left != right {
			return left < right
		}
		return links[i].URL < links[j].URL
	})
}

func encodeSnapshot(snapshot catalog.Snapshot) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(snapshot); err != nil {
		return nil, fmt.Errorf("encode snapshot: %w", err)
	}
	return output.Bytes(), nil
}

func writeAtomic(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", directory, err)
	}
	temporary, err := os.CreateTemp(directory, ".catalog-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("set output mode: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary output: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary output: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace output %s: %w", path, err)
	}
	return nil
}
