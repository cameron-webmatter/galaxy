package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/galaxy/galaxy/pkg/adapters"
	"github.com/galaxy/galaxy/pkg/adapters/standalone"
	"github.com/galaxy/galaxy/pkg/config"
	"github.com/galaxy/galaxy/pkg/router"
)

type HybridBuilder struct {
	Config     *config.Config
	SrcDir     string
	PagesDir   string
	OutDir     string
	PublicDir  string
	Router     *router.Router
	SSGBuilder *SSGBuilder
	SSRBuilder *SSRBuilder
}

func NewHybridBuilder(cfg *config.Config, srcDir, pagesDir, outDir, publicDir string) *HybridBuilder {
	return &HybridBuilder{
		Config:     cfg,
		SrcDir:     srcDir,
		PagesDir:   pagesDir,
		OutDir:     outDir,
		PublicDir:  publicDir,
		Router:     router.NewRouter(pagesDir),
		SSGBuilder: NewSSGBuilder(cfg, srcDir, pagesDir, outDir, publicDir),
		SSRBuilder: NewSSRBuilder(cfg, srcDir, pagesDir, outDir, publicDir),
	}
}

func (b *HybridBuilder) Build() error {
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

	staticRoutes := []*router.Route{}
	dynamicRoutes := []*router.Route{}

	for _, route := range b.Router.Routes {
		if b.shouldPrerender(route) {
			staticRoutes = append(staticRoutes, route)
		} else {
			dynamicRoutes = append(dynamicRoutes, route)
		}
	}

	for _, route := range staticRoutes {
		if route.Type == router.RouteStatic && !route.IsEndpoint {
			if err := b.SSGBuilder.buildStaticRoute(route); err != nil {
				return fmt.Errorf("build static route %s: %w", route.Pattern, err)
			}
		}
	}

	if len(dynamicRoutes) > 0 {
		serverDir := filepath.Join(b.OutDir, "server")
		if err := os.MkdirAll(serverDir, 0755); err != nil {
			return fmt.Errorf("create server dir: %w", err)
		}

		b.SSRBuilder.Router.Routes = dynamicRoutes
		if err := b.generateServerForDynamicRoutes(serverDir, dynamicRoutes); err != nil {
			return fmt.Errorf("generate server: %w", err)
		}
	}

	if err := b.SSGBuilder.copyPublicAssets(); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}

	return nil
}

func (b *HybridBuilder) shouldPrerender(route *router.Route) bool {
	if route.IsEndpoint {
		return false
	}

	if route.Type != router.RouteStatic {
		return false
	}

	content, err := os.ReadFile(route.FilePath)
	if err != nil {
		return true
	}

	src := string(content)
	return !strings.Contains(src, "prerender = false") && !strings.Contains(src, "prerender=false")
}

func (b *HybridBuilder) generateServerForDynamicRoutes(serverDir string, routes []*router.Route) error {
	adapter := standalone.New()

	routeInfos := make([]adapters.RouteInfo, len(routes))
	for i, route := range routes {
		routeInfos[i] = adapters.RouteInfo{
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
		Routes:    routeInfos,
	}

	return adapter.Build(cfg)
}
