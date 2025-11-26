package cms

import (
	"testing"

	"github.com/goliatone/go-theme"
)

func TestProviderSelectsWithFallback(t *testing.T) {
	reg := theme.NewRegistry()
	if err := reg.Register(&theme.Manifest{Name: "default", Version: "1.0.0", Tokens: map[string]string{"primary": "blue"}}); err != nil {
		t.Fatalf("register default: %v", err)
	}

	p := Provider{
		Registry:       reg,
		DefaultTheme:   "default",
		DefaultVariant: "dark",
	}

	sel, err := p.Select("", "")
	if err != nil {
		t.Fatalf("select default: %v", err)
	}
	if sel.Theme != "default" {
		t.Fatalf("expected default theme, got %s", sel.Theme)
	}
	if sel.Variant != "dark" {
		t.Fatalf("expected default variant, got %s", sel.Variant)
	}

	tokens := sel.Tokens()
	if tokens["primary"] != "blue" {
		t.Fatalf("expected tokens merged from manifest, got %v", tokens)
	}
}

func TestTemplateResolutionPrefersVariantThenBaseThenFallback(t *testing.T) {
	manifest := &theme.Manifest{
		Name:      "default",
		Version:   "1.0.0",
		Templates: map[string]string{"forms.input": "templates/forms/base_input.tmpl"},
		Variants: map[string]theme.Variant{
			"dark": {Templates: map[string]string{"forms.input": "templates/forms/dark_input.tmpl"}},
		},
	}

	// variant override
	tpl := ResolveTemplate(manifest, "dark", "forms.input", "fallback.tmpl")
	if tpl != "templates/forms/dark_input.tmpl" {
		t.Fatalf("expected variant override, got %s", tpl)
	}

	// base fallback
	tpl = ResolveTemplate(manifest, "dark", "forms.select", "fallback.tmpl")
	if tpl != "fallback.tmpl" {
		t.Fatalf("expected fallback when key missing, got %s", tpl)
	}
}

func TestAssetResolutionHandlesPrefixAndVariantOverride(t *testing.T) {
	manifest := &theme.Manifest{
		Name:    "default",
		Version: "1.0.0",
		Assets: theme.Assets{
			Prefix: "/static",
			Files: map[string]string{
				"logo":   "logo.png",
				"banner": "/images/banner.png",
			},
		},
		Variants: map[string]theme.Variant{
			"dark": {
				Assets: theme.Assets{
					Prefix: "https://cdn.example.com/theme/",
					Files:  map[string]string{"logo": "logo-dark.png"},
				},
			},
		},
	}

	// variant override with CDN prefix
	url, ok := ResolveAsset(manifest, "dark", "logo")
	if !ok {
		t.Fatalf("expected variant asset to be resolved")
	}
	if url != "https://cdn.example.com/theme/logo-dark.png" {
		t.Fatalf("unexpected asset url: %s", url)
	}

	// base prefix applied when variant override missing, with leading slash normalized
	url, ok = ResolveAsset(manifest, "light", "logo")
	if !ok {
		t.Fatalf("expected base asset to be resolved")
	}
	if url != "/static/logo.png" {
		t.Fatalf("unexpected base asset url: %s", url)
	}

	// leading slash in asset path should not drop prefix
	url, ok = ResolveAsset(manifest, "light", "banner")
	if !ok || url != "/static/images/banner.png" {
		t.Fatalf("expected normalized path with prefix, got %s", url)
	}

	// fallback missing asset remains unresolved
	if _, ok := ResolveAsset(manifest, "dark", "favicon"); ok {
		t.Fatalf("expected unresolved asset for missing key")
	}
}

func TestAssetResolutionSkipsEmptyKeys(t *testing.T) {
	if _, ok := ResolveAsset(nil, "dark", "logo"); ok {
		t.Fatalf("expected nil manifest to fail resolution")
	}
}
