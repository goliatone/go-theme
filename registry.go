package theme

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Registry defines methods to register and retrieve theme manifests.
type Registry interface {
	Register(manifest *Manifest) error
	Get(name string, opts ...QueryOption) (*Manifest, error)
	List() []ManifestRef
}

// ThemeProvider exposes read-only registry access for downstream consumers.
type ThemeProvider interface {
	Theme(name string, opts ...QueryOption) (*Manifest, error)
	Themes() []ManifestRef
}

// ManifestRef summarizes a stored manifest.
type ManifestRef struct {
	Name        string
	Version     string
	Description string
}

// MemoryRegistry is a minimal in-memory implementation of Registry and ThemeProvider.
type MemoryRegistry struct {
	mu     sync.RWMutex
	themes map[string]map[string]*Manifest
}

// NewRegistry constructs an empty MemoryRegistry.
func NewRegistry() *MemoryRegistry {
	return &MemoryRegistry{
		themes: make(map[string]map[string]*Manifest),
	}
}

var (
	// ErrThemeNotFound is returned when a theme name has no registered manifests.
	ErrThemeNotFound = errors.New("theme not found")
	// ErrVersionNotFound is returned when a specific version cannot be located.
	ErrVersionNotFound = errors.New("theme version not found")
)

// QueryOption modifies how registry lookups behave.
type QueryOption func(*queryOptions)

type queryOptions struct {
	version       string
	allowFallback bool
}

// WithVersion requests a specific manifest version.
func WithVersion(version string) QueryOption {
	return func(opts *queryOptions) {
		opts.version = strings.TrimSpace(version)
	}
}

// WithoutFallback disables fallback to the latest version when a specific one is missing.
func WithoutFallback() QueryOption {
	return func(opts *queryOptions) {
		opts.allowFallback = false
	}
}

// Register validates and stores a manifest. Existing entries for the same name+version are overwritten.
func (r *MemoryRegistry) Register(manifest *Manifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}
	if err := manifest.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.themes[manifest.Name]; !ok {
		r.themes[manifest.Name] = make(map[string]*Manifest)
	}

	r.themes[manifest.Name][manifest.Version] = copyManifest(manifest)
	return nil
}

// Get fetches a manifest by name, optionally constrained to a version with fallback to the latest.
func (r *MemoryRegistry) Get(name string, opts ...QueryOption) (*Manifest, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required")
	}

	settings := queryOptions{allowFallback: true}
	for _, opt := range opts {
		opt(&settings)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.themes[name]
	if !ok || len(versions) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrThemeNotFound, name)
	}

	if settings.version != "" {
		if manifest, ok := versions[settings.version]; ok {
			return copyManifest(manifest), nil
		}
		if !settings.allowFallback {
			return nil, fmt.Errorf("%w: %s@%s", ErrVersionNotFound, name, settings.version)
		}
	}

	version := latestVersion(versions)
	if version == "" {
		return nil, fmt.Errorf("%w: %s", ErrThemeNotFound, name)
	}
	return copyManifest(versions[version]), nil
}

// List returns a sorted list of all stored manifests.
func (r *MemoryRegistry) List() []ManifestRef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var refs []ManifestRef
	for name, versions := range r.themes {
		for version, manifest := range versions {
			refs = append(refs, ManifestRef{
				Name:        name,
				Version:     version,
				Description: manifest.Description,
			})
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Name == refs[j].Name {
			return compareVersions(refs[i].Version, refs[j].Version) > 0
		}
		return refs[i].Name < refs[j].Name
	})

	return refs
}

// Theme is an alias for Get to satisfy ThemeProvider.
func (r *MemoryRegistry) Theme(name string, opts ...QueryOption) (*Manifest, error) {
	return r.Get(name, opts...)
}

// Themes is an alias for List to satisfy ThemeProvider.
func (r *MemoryRegistry) Themes() []ManifestRef {
	return r.List()
}

func latestVersion(versions map[string]*Manifest) string {
	var best string
	for version := range versions {
		if best == "" || compareVersions(version, best) > 0 {
			best = version
		}
	}
	return best
}

// compareVersions performs a basic semantic version comparison (major.minor.patch).
func compareVersions(a, b string) int {
	ap := parseVersionParts(strings.TrimPrefix(a, "v"))
	bp := parseVersionParts(strings.TrimPrefix(b, "v"))

	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}

	for i := 0; i < maxLen; i++ {
		var av, bv int
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av > bv {
			return 1
		}
		if av < bv {
			return -1
		}
	}
	return 0
}

func parseVersionParts(version string) []int {
	if version == "" {
		return []int{}
	}
	parts := strings.Split(version, ".")
	out := make([]int, len(parts))
	for i, part := range parts {
		if part == "" {
			continue
		}
		var value int
		fmt.Sscanf(part, "%d", &value)
		out[i] = value
	}
	return out
}

func copyManifest(src *Manifest) *Manifest {
	if src == nil {
		return nil
	}
	cloned := Manifest{
		Name:        src.Name,
		Version:     src.Version,
		Description: src.Description,
		Tokens:      cloneStringMap(src.Tokens),
		Fonts:       cloneStringMap(src.Fonts),
		Assets: Assets{
			Prefix: src.Assets.Prefix,
			Files:  cloneStringMap(src.Assets.Files),
		},
		Templates: cloneStringMap(src.Templates),
		Variants:  make(map[string]Variant, len(src.Variants)),
	}

	for name, variant := range src.Variants {
		cloned.Variants[name] = Variant{
			Description: variant.Description,
			Tokens:      cloneStringMap(variant.Tokens),
			Templates:   cloneStringMap(variant.Templates),
			Assets: Assets{
				Prefix: variant.Assets.Prefix,
				Files:  cloneStringMap(variant.Assets.Files),
			},
		}
	}

	return &cloned
}
