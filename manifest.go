// Package theme provides shared theming manifest definitions, loaders, and a simple in-memory registry.
package theme

import (
	"fmt"
	"strings"
)

// Manifest defines the shape of a theme file that downstream systems (go-cms, go-formgen) can consume.
type Manifest struct {
	Name        string             `json:"name" yaml:"name"`
	Version     string             `json:"version" yaml:"version"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Tokens      map[string]string  `json:"tokens,omitempty" yaml:"tokens,omitempty"`
	Fonts       map[string]string  `json:"fonts,omitempty" yaml:"fonts,omitempty"`
	Assets      Assets             `json:"assets,omitempty" yaml:"assets,omitempty"`
	Templates   map[string]string  `json:"templates,omitempty" yaml:"templates,omitempty"`
	Variants    map[string]Variant `json:"variants,omitempty" yaml:"variants,omitempty"`
}

// Assets groups static assets and optional prefix/CDN root.
type Assets struct {
	Prefix string            `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Files  map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
}

// Variant captures token/template/asset overrides for a named variant (e.g., light/dark).
type Variant struct {
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Tokens      map[string]string `json:"tokens,omitempty" yaml:"tokens,omitempty"`
	Templates   map[string]string `json:"templates,omitempty" yaml:"templates,omitempty"`
	Assets      Assets            `json:"assets,omitempty" yaml:"assets,omitempty"`
}

// ValidationError aggregates manifest validation issues.
type ValidationError struct {
	Issues []string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("manifest validation failed: %s", strings.Join(e.Issues, "; "))
}

// Validate checks required fields and basic integrity for maps/variants.
func (m *Manifest) Validate() error {
	if m == nil {
		return fmt.Errorf("manifest is nil")
	}

	var issues []string

	if strings.TrimSpace(m.Name) == "" {
		issues = append(issues, "name is required")
	}

	if strings.TrimSpace(m.Version) == "" {
		issues = append(issues, "version is required")
	}

	validateMap := func(label string, values map[string]string) {
		for k, v := range values {
			if strings.TrimSpace(k) == "" {
				issues = append(issues, fmt.Sprintf("%s has empty key", label))
			}
			if strings.TrimSpace(v) == "" {
				issues = append(issues, fmt.Sprintf("%s entry '%s' is empty", label, k))
			}
		}
	}

	validateMap("tokens", m.Tokens)
	validateMap("fonts", m.Fonts)
	validateMap("templates", m.Templates)
	validateMap("assets.files", m.Assets.Files)

	for name, variant := range m.Variants {
		if strings.TrimSpace(name) == "" {
			issues = append(issues, "variant name cannot be empty")
		}
		validateMap(fmt.Sprintf("variants.%s.tokens", name), variant.Tokens)
		validateMap(fmt.Sprintf("variants.%s.templates", name), variant.Templates)
		validateMap(fmt.Sprintf("variants.%s.assets.files", name), variant.Assets.Files)
	}

	if len(issues) > 0 {
		return ValidationError{Issues: issues}
	}

	return nil
}

// TokensForVariant merges base tokens with the requested variant (variant tokens take precedence).
func (m Manifest) TokensForVariant(variant string) map[string]string {
	if variant == "" {
		return cloneStringMap(m.Tokens)
	}

	selected, ok := m.Variants[variant]
	if !ok || len(selected.Tokens) == 0 {
		return cloneStringMap(m.Tokens)
	}

	merged := cloneStringMap(m.Tokens)
	for k, v := range selected.Tokens {
		merged[k] = v
	}
	return merged
}

// CSSVariables returns a CSS variable map (prefixed with "--" unless overridden) for a variant.
func (m Manifest) CSSVariables(prefix, variant string) map[string]string {
	if prefix == "" {
		prefix = "--"
	}
	tokenSet := m.TokensForVariant(variant)
	vars := make(map[string]string, len(tokenSet))
	for k, v := range tokenSet {
		vars[prefix+k] = v
	}
	return vars
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
