package wasm

import (
	"encoding/json"
	"os"
)

type WasmManifest struct {
	Assets map[string]WasmPageAssets `json:"assets"`
}

type WasmPageAssets struct {
	WasmModules []WasmModule `json:"wasmModules"`
	JSScripts   []string     `json:"jsScripts"`
}

type WasmModule struct {
	Hash       string `json:"hash"`
	WasmPath   string `json:"wasmPath"`
	LoaderPath string `json:"loaderPath"`
}

func NewManifest() *WasmManifest {
	return &WasmManifest{
		Assets: make(map[string]WasmPageAssets),
	}
}

func (m *WasmManifest) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadManifest(path string) (*WasmManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest WasmManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}
