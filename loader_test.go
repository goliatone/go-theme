package theme

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestLoadBytesJSON(t *testing.T) {
	data := []byte(`{
		"name": "default",
		"version": "1.0.0",
		"tokens": {"primary": "blue"},
		"templates": {"forms.input": "forms/input.tmpl"}
	}`)

	manifest, err := LoadBytes(data, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.Name != "default" || manifest.Version != "1.0.0" {
		t.Fatalf("manifest not decoded correctly: %+v", manifest)
	}
}

func TestLoadBytesYAMLWithoutFormat(t *testing.T) {
	data := []byte(`
name: default
version: 1.0.0
tokens:
  primary: blue
assets:
  prefix: /static
  files:
    logo: logo.png
`)

	manifest, err := LoadBytes(data, "")
	if err != nil {
		t.Fatalf("unexpected error decoding yaml: %v", err)
	}
	if manifest.Assets.Prefix != "/static" {
		t.Fatalf("expected assets prefix to be decoded, got %s", manifest.Assets.Prefix)
	}
}

func TestLoadFileInfersFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.yaml")
	err := os.WriteFile(path, []byte("name: default\nversion: 1.0.0\n"), 0o644)
	if err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	manifest, err := LoadFile(os.DirFS(dir), "theme.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", manifest.Version)
	}
}

func TestLoadDirSearchesDefaultNames(t *testing.T) {
	fsys := fstest.MapFS{
		"manifest.json": &fstest.MapFile{
			Data: []byte(`{"name":"default","version":"1.0.0","tokens":{"primary":"blue"}}`),
		},
	}

	manifest, err := LoadDir(fsys, ".")
	if err != nil {
		t.Fatalf("unexpected error searching dir: %v", err)
	}
	if manifest.Name != "default" {
		t.Fatalf("expected manifest to be loaded from default filename, got %s", manifest.Name)
	}
}

func TestLoadBytesFailsValidation(t *testing.T) {
	data := []byte(`{"name": "", "version": ""}`)
	if _, err := LoadBytes(data, "json"); err == nil {
		t.Fatalf("expected validation error")
	}
}
