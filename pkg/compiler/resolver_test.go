package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComponentResolverDynamicDiscovery(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create directories
	componentsDir := filepath.Join(tmpDir, "components")
	layoutsDir := filepath.Join(tmpDir, "layouts")
	utilsDir := filepath.Join(tmpDir, "utils")

	os.MkdirAll(componentsDir, 0755)
	os.MkdirAll(layoutsDir, 0755)
	os.MkdirAll(utilsDir, 0755)

	// Create component files
	os.WriteFile(filepath.Join(componentsDir, "Button.gxc"), []byte("<button></button>"), 0644)
	os.WriteFile(filepath.Join(componentsDir, "Card.gxc"), []byte("<div></div>"), 0644)
	os.WriteFile(filepath.Join(layoutsDir, "Layout.gxc"), []byte("<html></html>"), 0644)
	os.WriteFile(filepath.Join(utilsDir, "Helper.gxc"), []byte("<span></span>"), 0644)

	// Create resolver
	resolver := NewComponentResolver(tmpDir, nil)

	// Test resolving components from different directories
	tests := []struct {
		name      string
		component string
		wantErr   bool
	}{
		{"Button in components", "Button", false},
		{"Card in components", "Card", false},
		{"Layout in layouts", "Layout", false},
		{"Helper in utils", "Helper", false},
		{"NonExistent", "NonExistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := resolver.Resolve(tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve(%q) error = %v, wantErr %v", tt.component, err, tt.wantErr)
			}
			if !tt.wantErr && path == "" {
				t.Errorf("Resolve(%q) returned empty path", tt.component)
			}
			if !tt.wantErr {
				t.Logf("Resolved %s to %s", tt.component, path)
			}
		})
	}
}

func TestComponentResolverExcludesPages(t *testing.T) {
	tmpDir := t.TempDir()

	componentsDir := filepath.Join(tmpDir, "components")
	pagesDir := filepath.Join(tmpDir, "pages")

	os.MkdirAll(componentsDir, 0755)
	os.MkdirAll(pagesDir, 0755)

	os.WriteFile(filepath.Join(componentsDir, "Button.gxc"), []byte("<button></button>"), 0644)
	os.WriteFile(filepath.Join(pagesDir, "Index.gxc"), []byte("<html></html>"), 0644)

	resolver := NewComponentResolver(tmpDir, nil)

	// Button should be found
	_, err := resolver.Resolve("Button")
	if err != nil {
		t.Errorf("Expected Button to be found, got error: %v", err)
	}

	// Index (page) should NOT be found
	_, err = resolver.Resolve("Index")
	if err == nil {
		t.Errorf("Expected Index to not be found (it's in pages dir)")
	}
}
