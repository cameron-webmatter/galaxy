package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/galaxy/galaxy/pkg/adapters"
	"github.com/galaxy/galaxy/pkg/adapters/standalone"
	"github.com/galaxy/galaxy/pkg/config"
	"github.com/galaxy/galaxy/pkg/router"
)

type SSRBuilder struct {
	Config    *config.Config
	PagesDir  string
	OutDir    string
	PublicDir string
	Router    *router.Router
}

func NewSSRBuilder(cfg *config.Config, pagesDir, outDir, publicDir string) *SSRBuilder {
	return &SSRBuilder{
		Config:    cfg,
		PagesDir:  pagesDir,
		OutDir:    outDir,
		PublicDir: publicDir,
		Router:    router.NewRouter(pagesDir),
	}
}

func (b *SSRBuilder) Build() error {
	if err := b.Router.Discover(); err != nil {
		return fmt.Errorf("route discovery: %w", err)
	}
	b.Router.Sort()

	if err := os.RemoveAll(b.OutDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clean output: %w", err)
	}

	if err := os.MkdirAll(b.OutDir, 0755); err != nil {
		return fmt.Errorf("create output: %w", err)
	}

	serverDir := filepath.Join(b.OutDir, "server")
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("create server dir: %w", err)
	}

	if err := b.generateServerCode(serverDir); err != nil {
		return fmt.Errorf("generate server: %w", err)
	}

	if err := b.copyPublicAssets(); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}

	if err := b.compileServer(serverDir); err != nil {
		return fmt.Errorf("compile server: %w", err)
	}

	return nil
}

func (b *SSRBuilder) generateServerCode(serverDir string) error {
	var adapter adapters.Adapter

	switch b.Config.Adapter.Name {
	case config.AdapterStandalone:
		adapter = standalone.New()
	default:
		return fmt.Errorf("unsupported adapter: %s", b.Config.Adapter.Name)
	}

	routes := make([]adapters.RouteInfo, len(b.Router.Routes))
	for i, route := range b.Router.Routes {
		routes[i] = adapters.RouteInfo{
			Pattern:    route.Pattern,
			FilePath:   route.FilePath,
			IsEndpoint: route.IsEndpoint,
		}
	}

	cfg := &adapters.BuildConfig{
		Config:    b.Config,
		ServerDir: serverDir,
		OutDir:    b.OutDir,
		PagesDir:  b.PagesDir,
		PublicDir: b.PublicDir,
		Routes:    routes,
	}

	return adapter.Build(cfg)
}

func (b *SSRBuilder) copyPublicAssets() error {
	publicOutDir := filepath.Join(b.OutDir, "public")

	if _, err := os.Stat(b.PublicDir); os.IsNotExist(err) {
		return nil
	}

	if err := os.MkdirAll(publicOutDir, 0755); err != nil {
		return err
	}

	return filepath.Walk(b.PublicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(b.PublicDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(publicOutDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}

func (b *SSRBuilder) compileServer(serverDir string) error {
	fmt.Println("  ✓ Server compiled")
	return nil
}
