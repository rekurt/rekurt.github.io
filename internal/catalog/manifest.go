package catalog

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadManifest(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open manifest %s: %w", path, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest %s: %w", path, err)
	}

	var extra any
	err = decoder.Decode(&extra)
	if err == nil {
		return Manifest{}, fmt.Errorf("decode manifest %s: multiple YAML documents are not allowed", path)
	}
	if err != io.EOF {
		return Manifest{}, fmt.Errorf("decode trailing manifest data %s: %w", path, err)
	}

	return manifest, nil
}
