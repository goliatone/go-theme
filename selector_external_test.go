package theme_test

import (
	"testing"

	theme "github.com/goliatone/go-theme"
)

func TestResolvedSelectionAPIExternalCompile(t *testing.T) {
	reg := theme.NewRegistry()
	if err := reg.Register(&theme.Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens:  map[string]string{"primary": "blue"},
		Assets: theme.Assets{
			Prefix: "/static",
			Files:  map[string]string{"logo": "logo.svg"},
		},
	}); err != nil {
		t.Fatalf("register manifest: %v", err)
	}

	selector := theme.Selector{
		Registry:       reg,
		DefaultTheme:   "default",
		DefaultVariant: "light",
	}
	sel, err := selector.Select("", "")
	if err != nil {
		t.Fatalf("select theme: %v", err)
	}

	snapshot := sel.Snapshot()
	var _ theme.ResolvedSelection = snapshot

	if snapshot.Theme != "default" {
		t.Fatalf("expected default theme, got %s", snapshot.Theme)
	}
	if snapshot.AssetPrefix != "/static" {
		t.Fatalf("expected /static asset prefix, got %s", snapshot.AssetPrefix)
	}
	if snapshot.Assets["logo"] != "/static/logo.svg" {
		t.Fatalf("expected resolved logo asset, got %s", snapshot.Assets["logo"])
	}
}
