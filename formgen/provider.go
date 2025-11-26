// Package formgen provides helpers to feed go-theme manifests into go-formgen renderers
// without creating a hard dependency on go-formgen itself. It focuses on selecting a
// theme/variant, resolving renderer partials, exposing tokens/CSS variables, and composing
// asset URLs with prefix/CDN handling.
package formgen

import (
	"fmt"
	"path"
	"strings"

	"github.com/goliatone/go-theme"
)

// Provider selects themes for formgen usage with default theme/variant fallbacks.
type Provider struct {
	Registry       theme.ThemeProvider
	DefaultTheme   string
	DefaultVariant string
}

// Selection captures the chosen theme and helpers to generate renderer config.
type Selection struct {
	Theme    string
	Variant  string
	Manifest *theme.Manifest
}

// RendererConfig bundles resolved partials, tokens, CSS vars, and an asset resolver suitable for renderers.
type RendererConfig struct {
	Theme    string
	Variant  string
	Partials map[string]string
	Tokens   map[string]string
	CSSVars  map[string]string
	AssetURL func(string) string
}

// Select resolves a theme name/variant with fallback to defaults.
func (p Provider) Select(themeName, variant string, opts ...theme.QueryOption) (*Selection, error) {
	themeName = strings.TrimSpace(themeName)
	if themeName == "" {
		themeName = p.DefaultTheme
	}

	if p.Registry == nil {
		return nil, fmt.Errorf("theme registry is nil")
	}

	manifest, err := p.Registry.Theme(themeName, opts...)
	if err != nil && p.DefaultTheme != "" && themeName != p.DefaultTheme {
		manifest, err = p.Registry.Theme(p.DefaultTheme, opts...)
	}
	if err != nil {
		return nil, fmt.Errorf("resolve theme: %w", err)
	}

	if variant == "" {
		variant = p.DefaultVariant
	}

	return &Selection{
		Theme:    themeName,
		Variant:  variant,
		Manifest: manifest,
	}, nil
}

// Tokens returns the merged token map for the selected variant.
func (s Selection) Tokens() map[string]string {
	if s.Manifest == nil {
		return map[string]string{}
	}
	return s.Manifest.TokensForVariant(s.Variant)
}

// CSSVariables returns CSS variables for the variant with the provided prefix (defaults to "--" if empty).
func (s Selection) CSSVariables(prefix string) map[string]string {
	if s.Manifest == nil {
		return map[string]string{}
	}
	return s.Manifest.CSSVariables(prefix, s.Variant)
}

// Partial resolves a template key (e.g., "forms.input") using variant overrides, base templates, then fallback.
func (s Selection) Partial(key, fallback string) string {
	return resolveTemplate(s.Manifest, s.Variant, key, fallback)
}

// Partials resolves a map of template keys to fallback paths and returns the resolved map.
func (s Selection) Partials(fallbacks map[string]string) map[string]string {
	out := make(map[string]string, len(fallbacks))
	for key, fallback := range fallbacks {
		out[key] = s.Partial(key, fallback)
	}
	return out
}

// Asset returns a themed asset path with prefix handling (variant overrides then base). Bool indicates presence.
func (s Selection) Asset(key string) (string, bool) {
	return resolveAsset(s.Manifest, s.Variant, key)
}

// RendererTheme builds a RendererConfig for renderers (vanilla/Preact) given a set of partial fallbacks.
func (s Selection) RendererTheme(fallbacks map[string]string) RendererConfig {
	partials := s.Partials(fallbacks)
	return RendererConfig{
		Theme:    s.Theme,
		Variant:  s.Variant,
		Partials: partials,
		Tokens:   s.Tokens(),
		CSSVars:  s.CSSVariables(""),
		AssetURL: func(key string) string {
			url, _ := s.Asset(key)
			return url
		},
	}
}

func resolveTemplate(manifest *theme.Manifest, variant, key, fallback string) string {
	if manifest == nil {
		return fallback
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fallback
	}

	if tpl := manifestTemplate(manifest, variant, key); tpl != "" {
		return tpl
	}
	if tpl := manifestTemplate(manifest, "", key); tpl != "" {
		return tpl
	}
	return fallback
}

func manifestTemplate(manifest *theme.Manifest, variant, key string) string {
	if manifest == nil {
		return ""
	}
	if variant == "" {
		return manifest.Templates[key]
	}
	if v, ok := manifest.Variants[variant]; ok {
		if tpl := v.Templates[key]; tpl != "" {
			return tpl
		}
	}
	return ""
}

func resolveAsset(manifest *theme.Manifest, variant, key string) (string, bool) {
	if manifest == nil {
		return "", false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false
	}

	variantAssets := manifest.Variants[variant].Assets
	prefix := strings.TrimSuffix(activePrefix(manifest.Assets.Prefix, variantAssets.Prefix), "/")

	if pathOverride, ok := variantAssets.Files[key]; ok && pathOverride != "" {
		return joinPath(prefix, pathOverride), true
	}

	if basePath, ok := manifest.Assets.Files[key]; ok && basePath != "" {
		return joinPath(prefix, basePath), true
	}

	return "", false
}

func activePrefix(base, override string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	return base
}

func joinPath(prefix, p string) string {
	p = strings.TrimPrefix(p, "/")
	if prefix == "" {
		return p
	}
	if strings.Contains(prefix, "://") {
		return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(p, "/")
	}
	return path.Join(prefix, p)
}
