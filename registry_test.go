package theme

import "testing"

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()

	err := reg.Register(&Manifest{
		Name:    "default",
		Version: "1.0.0",
		Tokens:  map[string]string{"primary": "blue"},
	})
	if err != nil {
		t.Fatalf("unexpected error registering manifest: %v", err)
	}

	err = reg.Register(&Manifest{
		Name:    "default",
		Version: "1.1.0",
		Tokens:  map[string]string{"primary": "indigo"},
	})
	if err != nil {
		t.Fatalf("unexpected error registering second manifest: %v", err)
	}

	latest, err := reg.Get("default")
	if err != nil {
		t.Fatalf("unexpected error fetching latest: %v", err)
	}
	if latest.Version != "1.1.0" {
		t.Fatalf("expected latest version to be 1.1.0, got %s", latest.Version)
	}

	exact, err := reg.Get("default", WithVersion("1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error fetching exact version: %v", err)
	}
	if exact.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", exact.Version)
	}

	fallback, err := reg.Get("default", WithVersion("2.0.0"))
	if err != nil {
		t.Fatalf("expected fallback to latest version, got error: %v", err)
	}
	if fallback.Version != "1.1.0" {
		t.Fatalf("expected fallback to latest version 1.1.0, got %s", fallback.Version)
	}

	if _, err := reg.Get("default", WithVersion("2.0.0"), WithoutFallback()); err == nil {
		t.Fatalf("expected error when fallback disabled for missing version")
	}
}

func TestListSorting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&Manifest{Name: "b-theme", Version: "1.0.0", Tokens: map[string]string{"x": "y"}})
	reg.Register(&Manifest{Name: "a-theme", Version: "1.0.0", Tokens: map[string]string{"x": "y"}})
	reg.Register(&Manifest{Name: "a-theme", Version: "1.1.0", Tokens: map[string]string{"x": "z"}})

	refs := reg.List()
	if len(refs) != 3 {
		t.Fatalf("expected 3 manifest refs, got %d", len(refs))
	}
	if refs[0].Name != "a-theme" || refs[0].Version != "1.1.0" {
		t.Fatalf("expected newest a-theme first, got %+v", refs[0])
	}
}

func TestRegisterValidatesManifest(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(&Manifest{Name: "", Version: ""}); err == nil {
		t.Fatalf("expected validation error on register")
	}
}
