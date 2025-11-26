package theme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

var defaultManifestNames = []string{
	"theme.json",
	"theme.yaml",
	"theme.yml",
	"manifest.json",
	"manifest.yaml",
	"manifest.yml",
}

// LoadBytes parses a manifest from raw bytes. If format is empty, it will try JSON then YAML.
func LoadBytes(data []byte, format string) (*Manifest, error) {
	format = normalizeFormat(format)

	switch format {
	case "json":
		return decodeJSON(data)
	case "yaml":
		return decodeYAML(data)
	case "":
		if manifest, err := decodeJSON(data); err == nil {
			return manifest, nil
		}
		return decodeYAML(data)
	default:
		return nil, fmt.Errorf("unsupported manifest format: %s", format)
	}
}

// LoadFile reads a manifest from a given fs.FS path, inferring format from the extension.
func LoadFile(fsys fs.FS, manifestPath string) (*Manifest, error) {
	data, err := fs.ReadFile(fsys, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	format := normalizeFormat(strings.TrimPrefix(path.Ext(manifestPath), "."))
	manifest, err := LoadBytes(data, format)
	if err != nil {
		return nil, fmt.Errorf("decode manifest %s: %w", manifestPath, err)
	}
	return manifest, nil
}

// LoadDir searches common manifest filenames within a directory in the provided fs.FS.
func LoadDir(fsys fs.FS, dir string) (*Manifest, error) {
	for _, name := range defaultManifestNames {
		candidate := path.Join(dir, name)
		info, err := fs.Stat(fsys, candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		return LoadFile(fsys, candidate)
	}
	return nil, fmt.Errorf("no manifest found in %s (looked for %s)", dir, strings.Join(defaultManifestNames, ", "))
}

func normalizeFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	default:
		return ""
	}
}

func decodeJSON(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func decodeYAML(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("yaml decode: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	return &manifest, nil
}
