package formgen

import (
	"testing"

	"github.com/goliatone/go-theme"
)

func TestProviderSelectsWithDefaults(t *testing.T) {
	reg := theme.NewRegistry()
	reg.Register(&theme.Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens:  map[string]string{"primary": "blue"},
	})

	p := Provider{
		Registry:       reg,
		DefaultTheme:   "default",
		DefaultVariant: "light",
	}

	sel, err := p.Select("", "")
	if err != nil {
		t.Fatalf("unexpected error selecting default theme: %v", err)
	}
	if sel.Theme != "default" || sel.Variant != "light" {
		t.Fatalf("unexpected selection: %+v", sel)
	}
}

func TestPartialsResolveVariantOverrides(t *testing.T) {
	manifest := &theme.Manifest{
		Name:      "default",
		Version:   "1.0.0",
		Templates: map[string]string{"forms.input": "templates/forms/base_input.tmpl"},
		Variants: map[string]theme.Variant{
			"dark": {Templates: map[string]string{"forms.input": "templates/forms/dark_input.tmpl"}},
		},
	}

	s := Selection{Theme: "default", Variant: "dark", Manifest: manifest}
	result := s.Partials(map[string]string{
		"forms.input":  "fallback/input.tmpl",
		"forms.select": "fallback/select.tmpl",
	})

	if result["forms.input"] != "templates/forms/dark_input.tmpl" {
		t.Fatalf("expected variant partial, got %s", result["forms.input"])
	}
	if result["forms.select"] != "fallback/select.tmpl" {
		t.Fatalf("expected fallback partial for missing key, got %s", result["forms.select"])
	}
}

func TestAssetResolutionForRendererTheme(t *testing.T) {
	manifest := &theme.Manifest{
		Name:    "default",
		Version: "1.0.0",
		Assets: theme.Assets{
			Prefix: "/static/",
			Files: map[string]string{
				"logo":   "logo.png",
				"banner": "/images/banner.png",
			},
		},
		Variants: map[string]theme.Variant{
			"dark": {
				Assets: theme.Assets{
					Prefix: "https://cdn.example.com/theme",
					Files:  map[string]string{"logo": "logo-dark.png"},
				},
			},
		},
	}

	s := Selection{Theme: "default", Variant: "dark", Manifest: manifest}
	cfg := s.RendererTheme(map[string]string{})

	url := cfg.AssetURL("logo")
	if url != "https://cdn.example.com/theme/logo-dark.png" {
		t.Fatalf("unexpected asset url: %s", url)
	}

	url = cfg.AssetURL("missing")
	if url != "" {
		t.Fatalf("expected empty url for missing asset, got %s", url)
	}

	// base prefix normalization for assets with leading slash
	cfg = Selection{Theme: "default", Variant: "", Manifest: manifest}.RendererTheme(nil)
	url = cfg.AssetURL("banner")
	if url != "/static/images/banner.png" {
		t.Fatalf("expected normalized base prefix, got %s", url)
	}
}

func TestRendererThemeProvidesTokensAndCSSVars(t *testing.T) {
	manifest := &theme.Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens:  map[string]string{"primary": "blue"},
		Variants: map[string]theme.Variant{
			"dark": {Tokens: map[string]string{"primary": "black"}},
		},
	}
	s := Selection{Theme: "default", Variant: "dark", Manifest: manifest}
	cfg := s.RendererTheme(nil)

	if cfg.Tokens["primary"] != "black" {
		t.Fatalf("expected variant token override, got %s", cfg.Tokens["primary"])
	}
	if cfg.CSSVars["--primary"] != "black" {
		t.Fatalf("expected CSS vars to reflect variant tokens, got %s", cfg.CSSVars["--primary"])
	}
}
