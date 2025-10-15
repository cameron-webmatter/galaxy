package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewComponentResolver(t *testing.T) {
	resolver := NewComponentResolver("/base", []string{"comp1", "comp2"})

	if resolver.BaseDir != "/base" {
		t.Errorf("Expected BaseDir /base, got %s", resolver.BaseDir)
	}

	if len(resolver.ComponentDirs) != 2 {
		t.Errorf("Expected 2 component dirs, got %d", len(resolver.ComponentDirs))
	}

	if resolver.ExplicitPaths == nil || resolver.Cache == nil {
		t.Error("Expected maps to be initialized")
	}
}

func TestNewComponentResolverDefaultDirs(t *testing.T) {
	resolver := NewComponentResolver("/base", nil)

	if len(resolver.ComponentDirs) != 1 || resolver.ComponentDirs[0] != "components" {
		t.Errorf("Expected default ['components'], got %v", resolver.ComponentDirs)
	}
}

func TestSetCurrentFile(t *testing.T) {
	resolver := NewComponentResolver("/base", nil)
	resolver.SetCurrentFile("/base/pages/index.gxc")

	if resolver.CurrentFile != "/base/pages/index.gxc" {
		t.Errorf("Expected CurrentFile to be set")
	}
}

func TestParseImports(t *testing.T) {
	resolver := NewComponentResolver("/base", nil)

	imports := []Import{
		{Path: "../components/Card.gxc", Alias: "Card", IsComponent: true},
		{Path: "@/layouts/Base.gxc", Alias: "Base", IsComponent: true},
		{Path: "./helper.js", Alias: "helper", IsComponent: false},
	}

	resolver.ParseImports(imports)

	if resolver.ExplicitPaths["Card"] != "../components/Card.gxc" {
		t.Error("Expected Card to be in ExplicitPaths")
	}

	if resolver.ExplicitPaths["Base"] != "@/layouts/Base.gxc" {
		t.Error("Expected Base to be in ExplicitPaths")
	}

	if _, ok := resolver.ExplicitPaths["helper"]; ok {
		t.Error("Non-component imports should not be in ExplicitPaths")
	}
}

func TestExtractComponentRefs(t *testing.T) {
	resolver := NewComponentResolver("/base", nil)

	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			"single component",
			"<Card title='test' />",
			[]string{"Card"},
		},
		{
			"multiple components",
			"<Layout><Header /><Footer /></Layout>",
			[]string{"Layout", "Header", "Footer"},
		},
		{
			"duplicate components",
			"<Card /><Card />",
			[]string{"Card"},
		},
		{
			"no components",
			"<div><span>hello</span></div>",
			[]string{},
		},
		{
			"mixed case",
			"<MyComponent123 /><AnotherOne />",
			[]string{"MyComponent123", "AnotherOne"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolver.ExtractComponentRefs(tt.template)

			if len(got) != len(tt.want) {
				t.Errorf("Expected %d components, got %d: %v", len(tt.want), len(got), got)
				return
			}

			for i, name := range tt.want {
				if got[i] != name {
					t.Errorf("Expected component[%d] = %s, got %s", i, name, got[i])
				}
			}
		})
	}
}

func TestResolveFromComponentDirs(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "components"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "components", "Card.gxc"), []byte(""), 0644)

	resolver := NewComponentResolver(tmpDir, []string{"components"})

	path, err := resolver.Resolve("Card")
	if err != nil {
		t.Fatalf("Expected to resolve Card, got error: %v", err)
	}

	expected := filepath.Join(tmpDir, "components", "Card.gxc")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestResolveFromCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "pages"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "pages", "Header.gxc"), []byte(""), 0644)

	resolver := NewComponentResolver(tmpDir, []string{"components"})
	resolver.SetCurrentFile(filepath.Join(tmpDir, "pages", "index.gxc"))

	path, err := resolver.Resolve("Header")
	if err != nil {
		t.Fatalf("Expected to resolve Header from current dir, got error: %v", err)
	}

	expected := filepath.Join(tmpDir, "pages", "Header.gxc")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestResolveNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	resolver := NewComponentResolver(tmpDir, []string{"components"})

	_, err := resolver.Resolve("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent component")
	}
}

func TestResolveCaching(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "components"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "components", "Card.gxc"), []byte(""), 0644)

	resolver := NewComponentResolver(tmpDir, []string{"components"})

	path1, _ := resolver.Resolve("Card")
	path2, _ := resolver.Resolve("Card")

	if path1 != path2 {
		t.Error("Resolve should return cached result")
	}

	if len(resolver.Cache) != 1 {
		t.Errorf("Expected 1 cached entry, got %d", len(resolver.Cache))
	}
}

func TestResolveImportPathRelative(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "pages"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "components"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "components", "Card.gxc"), []byte(""), 0644)

	resolver := NewComponentResolver(tmpDir, nil)
	resolver.SetCurrentFile(filepath.Join(tmpDir, "pages", "index.gxc"))

	path, err := resolver.resolveImportPath("../components/Card.gxc")
	if err != nil {
		t.Fatalf("Expected to resolve relative path, got error: %v", err)
	}

	expected := filepath.Join(tmpDir, "components", "Card.gxc")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestResolveImportPathAlias(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "layouts"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "layouts", "Base.gxc"), []byte(""), 0644)

	resolver := NewComponentResolver(tmpDir, nil)

	path, err := resolver.resolveImportPath("@/layouts/Base.gxc")
	if err != nil {
		t.Fatalf("Expected to resolve @ alias path, got error: %v", err)
	}

	expected := filepath.Join(tmpDir, "layouts", "Base.gxc")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestResolveImportPathRelativeNoContext(t *testing.T) {
	resolver := NewComponentResolver("/base", nil)

	_, err := resolver.resolveImportPath("./Card.gxc")
	if err == nil {
		t.Error("Expected error when resolving relative path without current file context")
	}
}

func TestResolveImportPathNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	resolver := NewComponentResolver(tmpDir, nil)
	resolver.SetCurrentFile(filepath.Join(tmpDir, "index.gxc"))

	_, err := resolver.resolveImportPath("./NonExistent.gxc")
	if err == nil {
		t.Error("Expected error for non-existent import path")
	}
}

func TestResolveWithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "components"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "components", "Card.gxc"), []byte(""), 0644)

	resolver := NewComponentResolver(tmpDir, nil)
	resolver.SetCurrentFile(filepath.Join(tmpDir, "pages", "index.gxc"))

	imports := []Import{
		{Path: "@/components/Card.gxc", Alias: "Card", IsComponent: true},
	}
	resolver.ParseImports(imports)

	path, err := resolver.Resolve("Card")
	if err != nil {
		t.Fatalf("Expected to resolve Card from explicit path, got error: %v", err)
	}

	expected := filepath.Join(tmpDir, "components", "Card.gxc")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}
