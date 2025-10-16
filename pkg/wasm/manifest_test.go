package wasm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManifest(t *testing.T) {
	manifest := NewManifest()

	if manifest.Assets == nil {
		t.Error("Expected Assets map to be initialized")
	}

	if len(manifest.Assets) != 0 {
		t.Error("Expected Assets map to be empty")
	}
}

func TestManifestSave(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "wasm-manifest.json")

	manifest := NewManifest()
	manifest.Assets["/index.html"] = WasmPageAssets{
		WasmModules: []WasmModule{
			{
				Hash:       "abc12345",
				WasmPath:   "/_assets/wasm/script-abc12345.wasm",
				LoaderPath: "/_assets/script-abc12345-loader.js",
			},
		},
		JSScripts: []string{"/_assets/script-def67890.js"},
	}

	err := manifest.Save(manifestPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Expected manifest file to be created")
	}

	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("Expected manifest to have content")
	}
}

func TestLoadManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "wasm-manifest.json")

	original := NewManifest()
	original.Assets["/about.html"] = WasmPageAssets{
		WasmModules: []WasmModule{
			{
				Hash:       "xyz98765",
				WasmPath:   "/_assets/wasm/script-xyz98765.wasm",
				LoaderPath: "/_assets/script-xyz98765-loader.js",
			},
		},
		JSScripts: []string{},
	}

	err := original.Save(manifestPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if loaded.Assets == nil {
		t.Error("Expected Assets to be loaded")
	}

	aboutAssets, exists := loaded.Assets["/about.html"]
	if !exists {
		t.Error("Expected /about.html to exist in loaded manifest")
	}

	if len(aboutAssets.WasmModules) != 1 {
		t.Errorf("Expected 1 WASM module, got %d", len(aboutAssets.WasmModules))
	}

	if aboutAssets.WasmModules[0].Hash != "xyz98765" {
		t.Errorf("Expected hash xyz98765, got %s", aboutAssets.WasmModules[0].Hash)
	}
}

func TestLoadManifestNonExistent(t *testing.T) {
	_, err := LoadManifest("/nonexistent/path/manifest.json")
	if err == nil {
		t.Error("Expected error when loading non-existent manifest")
	}
}

func TestLoadManifestInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(manifestPath, []byte("invalid json {"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	_, err = LoadManifest(manifestPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}

func TestManifestMultiplePages(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewManifest()

	manifest.Assets["/index.html"] = WasmPageAssets{
		WasmModules: []WasmModule{
			{Hash: "aaa111", WasmPath: "/wasm/aaa.wasm", LoaderPath: "/loader/aaa.js"},
		},
		JSScripts: []string{"/js/index.js"},
	}

	manifest.Assets["/about.html"] = WasmPageAssets{
		WasmModules: []WasmModule{
			{Hash: "bbb222", WasmPath: "/wasm/bbb.wasm", LoaderPath: "/loader/bbb.js"},
			{Hash: "ccc333", WasmPath: "/wasm/ccc.wasm", LoaderPath: "/loader/ccc.js"},
		},
		JSScripts: []string{"/js/about.js", "/js/shared.js"},
	}

	manifest.Assets["/contact.html"] = WasmPageAssets{
		WasmModules: []WasmModule{},
		JSScripts:   []string{"/js/contact.js"},
	}

	err := manifest.Save(manifestPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Assets) != 3 {
		t.Errorf("Expected 3 pages, got %d", len(loaded.Assets))
	}

	indexAssets := loaded.Assets["/index.html"]
	if len(indexAssets.WasmModules) != 1 {
		t.Errorf("Expected 1 WASM module for index, got %d", len(indexAssets.WasmModules))
	}

	aboutAssets := loaded.Assets["/about.html"]
	if len(aboutAssets.WasmModules) != 2 {
		t.Errorf("Expected 2 WASM modules for about, got %d", len(aboutAssets.WasmModules))
	}

	if len(aboutAssets.JSScripts) != 2 {
		t.Errorf("Expected 2 JS scripts for about, got %d", len(aboutAssets.JSScripts))
	}

	contactAssets := loaded.Assets["/contact.html"]
	if len(contactAssets.WasmModules) != 0 {
		t.Errorf("Expected 0 WASM modules for contact, got %d", len(contactAssets.WasmModules))
	}
}

func TestManifestEmptyModules(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewManifest()
	manifest.Assets["/empty.html"] = WasmPageAssets{
		WasmModules: []WasmModule{},
		JSScripts:   []string{},
	}

	err := manifest.Save(manifestPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	emptyAssets, exists := loaded.Assets["/empty.html"]
	if !exists {
		t.Error("Expected /empty.html to exist")
	}

	if emptyAssets.WasmModules == nil {
		t.Error("Expected WasmModules to be initialized (empty slice)")
	}

	if emptyAssets.JSScripts == nil {
		t.Error("Expected JSScripts to be initialized (empty slice)")
	}
}

func TestWasmModuleFields(t *testing.T) {
	module := WasmModule{
		Hash:       "test123",
		WasmPath:   "/path/to/module.wasm",
		LoaderPath: "/path/to/loader.js",
	}

	if module.Hash != "test123" {
		t.Errorf("Expected Hash test123, got %s", module.Hash)
	}

	if module.WasmPath != "/path/to/module.wasm" {
		t.Errorf("Expected WasmPath /path/to/module.wasm, got %s", module.WasmPath)
	}

	if module.LoaderPath != "/path/to/loader.js" {
		t.Errorf("Expected LoaderPath /path/to/loader.js, got %s", module.LoaderPath)
	}
}

func TestWasmPageAssetsFields(t *testing.T) {
	assets := WasmPageAssets{
		WasmModules: []WasmModule{
			{Hash: "abc", WasmPath: "/wasm/abc.wasm", LoaderPath: "/loader/abc.js"},
		},
		JSScripts: []string{"/js/script.js"},
	}

	if len(assets.WasmModules) != 1 {
		t.Errorf("Expected 1 WASM module, got %d", len(assets.WasmModules))
	}

	if len(assets.JSScripts) != 1 {
		t.Errorf("Expected 1 JS script, got %d", len(assets.JSScripts))
	}
}
