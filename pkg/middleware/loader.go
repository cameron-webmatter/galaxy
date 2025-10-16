package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

type LoadedMiddleware struct {
	OnRequest Middleware
	Sequence  []Middleware
}

func Load(projectDir string) (*LoadedMiddleware, error) {
	middlewarePath := filepath.Join(projectDir, "src", "middleware.go")

	if _, err := os.Stat(middlewarePath); os.IsNotExist(err) {
		return &LoadedMiddleware{}, nil
	}

	return &LoadedMiddleware{}, nil
}

func LoadPlugin(pluginPath string) (*LoadedMiddleware, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("open plugin: %w", err)
	}

	lm := &LoadedMiddleware{}

	onRequestSym, err := p.Lookup("OnRequest")
	if err == nil {
		if onRequest, ok := onRequestSym.(func(*Context, func() error) error); ok {
			lm.OnRequest = onRequest
		}
	}

	sequenceSym, err := p.Lookup("Sequence")
	if err == nil {
		if seqFunc, ok := sequenceSym.(func() []Middleware); ok {
			lm.Sequence = seqFunc()
		}
	}

	if lm.OnRequest == nil && lm.Sequence == nil {
		return nil, fmt.Errorf("no OnRequest or Sequence functions found")
	}

	return lm, nil
}
