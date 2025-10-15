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

	onRequestSym, err := p.Lookup("OnRequest")
	if err != nil {
		return nil, fmt.Errorf("lookup OnRequest: %w", err)
	}

	onRequest, ok := onRequestSym.(func(*Context, func() error) error)
	if !ok {
		return nil, fmt.Errorf("OnRequest has wrong signature")
	}

	lm := &LoadedMiddleware{
		OnRequest: onRequest,
	}

	sequenceSym, err := p.Lookup("Sequence")
	if err == nil {
		if seqFunc, ok := sequenceSym.(func() []Middleware); ok {
			lm.Sequence = seqFunc()
		}
	}

	return lm, nil
}
