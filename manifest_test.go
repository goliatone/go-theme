package theme

import "testing"

func TestManifestValidate(t *testing.T) {
	t.Run("requires name and version", func(t *testing.T) {
		m := Manifest{}
		err := m.Validate()
		if err == nil {
			t.Fatalf("expected validation error")
		}
		if _, ok := err.(ValidationError); !ok {
			t.Fatalf("expected ValidationError, got %T", err)
		}
	})

	t.Run("rejects empty token keys", func(t *testing.T) {
		m := Manifest{
			Name:    "default",
			Version: "1.0.0",
			Tokens: map[string]string{
				"": "primary",
			},
		}
		if err := m.Validate(); err == nil {
			t.Fatalf("expected validation error for empty token key")
		}
	})

	t.Run("rejects empty variant names", func(t *testing.T) {
		m := Manifest{
			Name:    "default",
			Version: "1.0.0",
			Variants: map[string]Variant{
				"": {Tokens: map[string]string{"color": "blue"}},
			},
		}
		if err := m.Validate(); err == nil {
			t.Fatalf("expected validation error for empty variant name")
		}
	})
}

func TestTokensForVariantMerges(t *testing.T) {
	m := Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens: map[string]string{
			"primary": "blue",
		},
		Variants: map[string]Variant{
			"dark": {
				Tokens: map[string]string{
					"primary": "black",
					"bg":      "#000",
				},
			},
		},
	}

	variant := m.TokensForVariant("dark")
	if variant["primary"] != "black" {
		t.Fatalf("expected variant token override, got %s", variant["primary"])
	}
	if _, ok := m.Tokens["bg"]; ok {
		t.Fatalf("base tokens mutated by variant merge")
	}
	if variant["bg"] != "#000" {
		t.Fatalf("expected variant token bg to be set")
	}
}

func TestCSSVariables(t *testing.T) {
	m := Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens: map[string]string{
			"primary": "blue",
		},
	}

	vars := m.CSSVariables("", "")
	if vars["--primary"] != "blue" {
		t.Fatalf("expected CSS variable to be prefixed, got %v", vars)
	}
}
