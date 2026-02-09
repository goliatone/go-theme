package theme

import "testing"

func TestSelectorSelectsWithDefaults(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens:  map[string]string{"primary": "blue"},
	})

	selector := Selector{
		Registry:       reg,
		DefaultTheme:   "default",
		DefaultVariant: "light",
	}

	sel, err := selector.Select("", "")
	if err != nil {
		t.Fatalf("unexpected error selecting default: %v", err)
	}
	if sel.Theme != "default" || sel.Variant != "light" {
		t.Fatalf("unexpected selection: %+v", sel)
	}
}

func TestSelectionTemplateResolution(t *testing.T) {
	manifest := &Manifest{
		Name:      "default",
		Version:   "1.0.0",
		Templates: map[string]string{"forms.input": "templates/forms/base_input.tmpl"},
		Variants: map[string]Variant{
			"dark": {Templates: map[string]string{"forms.input": "templates/forms/dark_input.tmpl"}},
		},
	}
	sel := Selection{Theme: "default", Variant: "dark", Manifest: manifest}

	if tpl := sel.Template("forms.input", "fallback"); tpl != "templates/forms/dark_input.tmpl" {
		t.Fatalf("expected variant template, got %s", tpl)
	}
	if tpl := sel.Template("forms.select", "fallback"); tpl != "fallback" {
		t.Fatalf("expected fallback template, got %s", tpl)
	}
}

func TestSelectionAssetResolution(t *testing.T) {
	manifest := &Manifest{
		Name:    "default",
		Version: "1.0.0",
		Assets: Assets{
			Prefix: "/static",
			Files: map[string]string{
				"logo":   "logo.png",
				"banner": "/images/banner.png",
			},
		},
		Variants: map[string]Variant{
			"dark": {
				Assets: Assets{
					Prefix: "https://cdn.example.com/theme/",
					Files:  map[string]string{"logo": "logo-dark.png"},
				},
			},
		},
	}

	sel := Selection{Theme: "default", Variant: "dark", Manifest: manifest}
	if url, ok := sel.Asset("logo"); !ok || url != "https://cdn.example.com/theme/logo-dark.png" {
		t.Fatalf("expected variant asset with CDN prefix, got %s", url)
	}
	if url, ok := sel.Asset("banner"); !ok || url != "/static/images/banner.png" {
		t.Fatalf("expected base asset with normalized prefix, got %s", url)
	}
}

func TestRendererTheme(t *testing.T) {
	manifest := &Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens:  map[string]string{"primary": "blue"},
		Variants: map[string]Variant{
			"dark": {Tokens: map[string]string{"primary": "black"}},
		},
	}
	sel := Selection{Theme: "default", Variant: "dark", Manifest: manifest}

	cfg := sel.RendererTheme(map[string]string{
		"forms.input": "fallback/input.tmpl",
	})

	if cfg.Tokens["primary"] != "black" {
		t.Fatalf("expected variant token override, got %s", cfg.Tokens["primary"])
	}
	if cfg.CSSVars["--primary"] != "black" {
		t.Fatalf("expected CSS vars from variant tokens, got %s", cfg.CSSVars["--primary"])
	}
	if cfg.Partials["forms.input"] != "fallback/input.tmpl" {
		t.Fatalf("expected fallback partial when missing, got %s", cfg.Partials["forms.input"])
	}
	if cfg.AssetURL("missing") != "" {
		t.Fatalf("expected empty asset URL for missing key")
	}
}

func TestSelectionSnapshotResolution(t *testing.T) {
	manifest := &Manifest{
		Name:    "brand",
		Version: "1.0.0",
		Tokens: map[string]string{
			"primary": "#0044ff",
			"accent":  "#ffaa00",
		},
		Assets: Assets{
			Prefix: "/static/theme",
			Files: map[string]string{
				"logo":    "logo.svg",
				"favicon": "favicon.ico",
			},
		},
		Templates: map[string]string{
			"layout.header": "templates/header.tmpl",
			"forms.input":   "templates/forms/input.tmpl",
		},
		Variants: map[string]Variant{
			"dark": {
				Tokens: map[string]string{
					"primary": "#0f172a",
				},
				Assets: Assets{
					Prefix: "https://cdn.example.com/theme",
					Files: map[string]string{
						"logo": "logo-dark.svg",
					},
				},
				Templates: map[string]string{
					"layout.header": "templates/dark/header.tmpl",
				},
			},
		},
	}

	sel := Selection{Theme: "brand", Variant: "dark", Manifest: manifest}
	snapshot := sel.Snapshot()

	if snapshot.Theme != "brand" {
		t.Fatalf("expected theme brand, got %s", snapshot.Theme)
	}
	if snapshot.Variant != "dark" {
		t.Fatalf("expected variant dark, got %s", snapshot.Variant)
	}
	if snapshot.Tokens["primary"] != "#0f172a" {
		t.Fatalf("expected variant token override, got %s", snapshot.Tokens["primary"])
	}
	if snapshot.Tokens["accent"] != "#ffaa00" {
		t.Fatalf("expected base token fallback, got %s", snapshot.Tokens["accent"])
	}
	if snapshot.Templates["layout.header"] != "templates/dark/header.tmpl" {
		t.Fatalf("expected variant template override, got %s", snapshot.Templates["layout.header"])
	}
	if snapshot.Templates["forms.input"] != "templates/forms/input.tmpl" {
		t.Fatalf("expected base template fallback, got %s", snapshot.Templates["forms.input"])
	}
	if snapshot.Assets["logo"] != "https://cdn.example.com/theme/logo-dark.svg" {
		t.Fatalf("expected variant asset override with variant prefix, got %s", snapshot.Assets["logo"])
	}
	if snapshot.Assets["favicon"] != "/static/theme/favicon.ico" {
		t.Fatalf("expected base asset fallback with base prefix, got %s", snapshot.Assets["favicon"])
	}
	if snapshot.AssetPrefix != "https://cdn.example.com/theme" {
		t.Fatalf("expected resolved prefix from variant, got %s", snapshot.AssetPrefix)
	}
}

func TestSelectionSnapshotPrefixFallback(t *testing.T) {
	sel := Selection{
		Theme:   "default",
		Variant: "light",
		Manifest: &Manifest{
			Name:    "default",
			Version: "1.0.0",
			Assets: Assets{
				Prefix: "/static/base",
				Files:  map[string]string{"logo": "logo.svg"},
			},
			Variants: map[string]Variant{
				"light": {
					Assets: Assets{
						Files: map[string]string{"badge": "badge.svg"},
					},
				},
			},
		},
	}

	snapshot := sel.Snapshot()
	if snapshot.AssetPrefix != "/static/base" {
		t.Fatalf("expected base prefix fallback, got %s", snapshot.AssetPrefix)
	}
	if snapshot.Assets["badge"] != "/static/base/badge.svg" {
		t.Fatalf("expected variant file to use base prefix when variant prefix missing, got %s", snapshot.Assets["badge"])
	}
}
