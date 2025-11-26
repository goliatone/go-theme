package theme

import (
	"fmt"
	"path"
	"strings"
)

// TemplateResolver resolves templates by key with optional fallbacks.
type TemplateResolver interface {
	Template(key, fallback string) string
	Partials(fallbacks map[string]string) map[string]string
}

// AssetResolver resolves asset URLs/paths.
type AssetResolver interface {
	Asset(key string) (string, bool)
}

// TokenProvider exposes tokens and CSS variables for a theme/variant.
type TokenProvider interface {
	Tokens() map[string]string
	CSSVariables(prefix string) map[string]string
}

// ThemeSelector selects a theme/variant and exposes resolvers.
type ThemeSelector interface {
	Select(themeName, variant string, opts ...QueryOption) (*Selection, error)
}

// Selector is the default implementation of ThemeSelector using a ThemeProvider registry.
type Selector struct {
	Registry       ThemeProvider
	DefaultTheme   string
	DefaultVariant string
}

// Selection holds the chosen theme/variant and provides resolvers for templates, assets, and tokens.
type Selection struct {
	Theme    string
	Variant  string
	Manifest *Manifest
}

// RendererConfig bundles resolved partials, tokens, CSS vars, and an asset resolver for renderers.
type RendererConfig struct {
	Theme    string
	Variant  string
	Partials map[string]string
	Tokens   map[string]string
	CSSVars  map[string]string
	AssetURL func(string) string
}

// Select resolves a theme name/variant with fallback to defaults and returns a Selection.
func (s Selector) Select(themeName, variant string, opts ...QueryOption) (*Selection, error) {
	themeName = strings.TrimSpace(themeName)
	if themeName == "" {
		themeName = s.DefaultTheme
	}

	if s.Registry == nil {
		return nil, fmt.Errorf("theme registry is nil")
	}

	manifest, err := s.Registry.Theme(themeName, opts...)
	if err != nil && s.DefaultTheme != "" && themeName != s.DefaultTheme {
		manifest, err = s.Registry.Theme(s.DefaultTheme, opts...)
	}
	if err != nil {
		return nil, fmt.Errorf("resolve theme: %w", err)
	}

	if variant == "" {
		variant = s.DefaultVariant
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

// CSSVariables returns CSS vars for the variant with the provided prefix (defaults to "--" if empty).
func (s Selection) CSSVariables(prefix string) map[string]string {
	if s.Manifest == nil {
		return map[string]string{}
	}
	return s.Manifest.CSSVariables(prefix, s.Variant)
}

// Template resolves a template key using variant overrides, then base, then fallback.
func (s Selection) Template(key, fallback string) string {
	return resolveTemplate(s.Manifest, s.Variant, key, fallback)
}

// Partials resolves a map of template keys to fallback paths.
func (s Selection) Partials(fallbacks map[string]string) map[string]string {
	out := make(map[string]string, len(fallbacks))
	for key, fallback := range fallbacks {
		out[key] = s.Template(key, fallback)
	}
	return out
}

// Asset returns a themed asset path with prefix handling (variant overrides then base). Bool indicates presence.
func (s Selection) Asset(key string) (string, bool) {
	return resolveAsset(s.Manifest, s.Variant, key)
}

// RendererTheme builds a RendererConfig given a set of fallback partials.
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

// resolveTemplate finds the template path for a key, preferring variant overrides then base templates, falling back to the provided default.
func resolveTemplate(manifest *Manifest, variant, key, fallback string) string {
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

func manifestTemplate(manifest *Manifest, variant, key string) string {
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

// resolveAsset returns an asset path including prefix, honoring variant overrides first.
func resolveAsset(manifest *Manifest, variant, key string) (string, bool) {
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
