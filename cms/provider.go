// Package cms provides helpers to wire go-theme into go-cms without adding a hard dependency.
// The helpers revolve around the existing theme.ThemeProvider interface and stay focused on
// selecting a manifest, resolving templates with variant-aware fallbacks, exposing token maps,
// and composing asset URLs with prefixes/CDN roots.
package cms

import (
	"fmt"
	"path"
	"strings"

	"github.com/goliatone/go-theme"
)

// Provider selects themes for CMS usage, applying defaults and fallback logic.
type Provider struct {
	Registry       theme.ThemeProvider
	DefaultTheme   string
	DefaultVariant string
}

// Selection captures the chosen theme/variant and provides convenience helpers for templates/helpers.
type Selection struct {
	Theme    string
	Variant  string
	Manifest *theme.Manifest
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

// Tokens returns the merged token map for the selected variant (variant tokens override base).
func (s Selection) Tokens() map[string]string {
	if s.Manifest == nil {
		return map[string]string{}
	}
	return s.Manifest.TokensForVariant(s.Variant)
}

// CSSVariables returns the CSS variable map (keys prefixed with "--" by default) for templates.
func (s Selection) CSSVariables(prefix string) map[string]string {
	if s.Manifest == nil {
		return map[string]string{}
	}
	return s.Manifest.CSSVariables(prefix, s.Variant)
}

// Template resolves a template key (e.g., "forms.input") using variant overrides, then base, then fallback.
func (s Selection) Template(key, fallback string) string {
	return ResolveTemplate(s.Manifest, s.Variant, key, fallback)
}

// Asset returns a themed asset path with prefix handling (variant assets/files override base).
// It returns the composed URL/path and a boolean indicating whether the asset key was found.
func (s Selection) Asset(key string) (string, bool) {
	return ResolveAsset(s.Manifest, s.Variant, key)
}

// ResolveTemplate finds the template path for a key, preferring variant overrides then base templates, falling back to the provided default.
func ResolveTemplate(manifest *theme.Manifest, variant, key, fallback string) string {
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

// ResolveAsset returns an asset path including prefix, honoring variant overrides first.
func ResolveAsset(manifest *theme.Manifest, variant, key string) (string, bool) {
	if manifest == nil {
		return "", false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false
	}

	prefix := strings.TrimSuffix(activePrefix(manifest.Assets.Prefix, manifest.Variants[variant].Assets.Prefix), "/")

	if pathOverride, ok := manifest.Variants[variant].Assets.Files[key]; ok && pathOverride != "" {
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
