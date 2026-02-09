# go-theme

Theming utilities for Go projects. Provides a manifest schema, loaders, a registry to support themes.

## Install

```sh
go get github.com/goliatone/go-theme
```

## Core Concepts

- **Manifest**: theme metadata, tokens, assets, templates, variants. JSON or YAML.
- **Loader**: read manifests from bytes/file/dir with `fs.FS` or `embed.FS`.
- **Registry**: store and fetch themes by name/version with fallback.
- **Resolvers**: `Selector`/`Selection` expose templates, assets, tokens, CSS vars, renderer configs, and resolved snapshots for hosts.

## Quick Start

```go
import (
    "embed"
    theme "github.com/goliatone/go-theme"
)

//go:embed themes/acme/*
var themeFS embed.FS

// load manifest
m, _ := theme.LoadDir(themeFS, "themes/acme")

// register
reg := theme.NewRegistry()
_ = reg.Register(m)

// shared selector and selection
selector := theme.Selector{Registry: reg, DefaultTheme: "acme-admin", DefaultVariant: "light"}
sel, _ := selector.Select("", "")

// template usage
cssVars := sel.CSSVariables("")                         // map of "--token": value
logoURL, _ := sel.Asset("logo")                         // prefix/CDN aware
headerTpl := sel.Template("layout.header", "default/header.tmpl")

// full resolved snapshot usage (for host adapters)
snapshot := sel.Snapshot()
_ = snapshot // theme/variant + merged tokens/assets/templates + resolved asset prefix

// renderer usage
rendererCfg := sel.RendererTheme(map[string]string{
    "forms.input":  "default/forms/input.tmpl",
    "forms.select": "default/forms/select.tmpl",
})
_ = rendererCfg // pass tokens, CSS vars, partials, and AssetURL to renderers
```

## Partial Naming Conventions
- `layout.header`, `layout.footer`, `layout.nav`
- `forms.input`, `forms.select`, `forms.checkbox`, `forms.radio`, `forms.textarea`, `forms.button`, `forms.field-wrapper`
- `components.alert`, `components.card`, `components.table`
- Keep partials in `templates/<area>/<name>.tmpl` (or similar) and reference the keys above in the manifest.
- Variant overrides live under `variants.<name>.templates.<key>`.
- Selector resolution order: variant → base → your fallback path.

## Asset Handling
- `assets.prefix` is prepended to asset file paths; variant `assets.prefix` overrides the base prefix.
- Asset file paths may be relative or start with `/`; leading slashes are trimmed before joining.
- CDN prefixes (contain `://`) are concatenated with a single `/`.
- Missing asset keys return an empty string/`false` from `Selection.Asset`.

## Resolved Snapshot
- Use `Selection.Snapshot()` when integrations need one complete payload instead of per-key lookups.
- Snapshot precedence is deterministic:
  - Tokens: variant overrides base.
  - Templates: variant override key, else base key.
  - Assets: variant file override key, else base key, preserving `Selection.Asset` prefix behavior.
- `AssetPrefix` resolves to `variants.<name>.assets.prefix` when set, otherwise base `assets.prefix`.

## Examples
- Manifests: `docs/examples/basic-theme.yaml`, `docs/examples/basic-theme.json`
- Example wiring (templates + renderers): `docs/examples/example-app.md`
