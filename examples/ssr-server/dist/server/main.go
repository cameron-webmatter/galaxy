package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/endpoints"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/middleware"
	"github.com/galaxy/galaxy/pkg/parser"
	"github.com/galaxy/galaxy/pkg/router"
	"github.com/galaxy/galaxy/pkg/ssr"
	"github.com/galaxy/galaxy/pkg/template"

	
	api "galaxy-server/pages/api"
	
	
)

var (
	rt       *router.Router
	comp     *compiler.ComponentCompiler
	pagesDir = "pages"
	routeMap = map[string]*routeInfo{
		
		"/": {Pattern: "/", FilePath: "pages/index.gxc", IsEndpoint: false},
		
		"/api/hello": {Pattern: "/api/hello", FilePath: "pages/api/hello.go", IsEndpoint: true},
		
	}
	endpointHandlers = map[string]map[string]endpoints.HandlerFunc{
		
		"/api/hello": {
			
			"GET": api.GET,
			
			"POST": api.POST,
			
		},
		
	}
)

type routeInfo struct {
	Pattern    string
	FilePath   string
	IsEndpoint bool
}

func main() {
	baseDir, _ := os.Getwd()
	comp = compiler.NewComponentCompiler(baseDir)

	http.HandleFunc("/", handleRequest)

	addr := "localhost:4322"
	log.Printf("🚀 Server running at http://%s\n", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if filepath.Ext(r.URL.Path) != "" {
		http.ServeFile(w, r, filepath.Join("/Users/cameron/dev/galaxy/examples/ssr-server/dist/public", r.URL.Path))
		return
	}

	staticPath := filepath.Join("/Users/cameron/dev/galaxy/examples/ssr-server/dist", r.URL.Path)
	if r.URL.Path == "/" {
		staticPath = filepath.Join("/Users/cameron/dev/galaxy/examples/ssr-server/dist", "index.html")
	} else {
		staticPath = filepath.Join("/Users/cameron/dev/galaxy/examples/ssr-server/dist", r.URL.Path, "index.html")
	}
	
	if _, err := os.Stat(staticPath); err == nil {
		http.ServeFile(w, r, staticPath)
		return
	}

	route, ok := routeMap[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	mwCtx := middleware.NewContext(w, r)
	mwCtx.Params = make(map[string]string)

	
	if route.IsEndpoint {
		handleEndpoint(route, mwCtx)
		return
	}
	handlePage(route, mwCtx)
	
}

func handleEndpoint(route *routeInfo, mwCtx *middleware.Context) {
	ep, ok := endpointHandlers[route.Pattern]
	if !ok {
		http.Error(mwCtx.Response, "Endpoint not found", http.StatusNotFound)
		return
	}

	method := mwCtx.Request.Method
	handler, ok := ep[method]
	if !ok {
		handler, ok = ep["ALL"]
		if !ok {
			http.Error(mwCtx.Response, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	ctx := endpoints.NewContext(mwCtx.Response, mwCtx.Request, mwCtx.Params, mwCtx.Locals)
	if err := handler(ctx); err != nil {
		http.Error(mwCtx.Response, err.Error(), http.StatusInternalServerError)
	}
}

func handlePage(route *routeInfo, mwCtx *middleware.Context) {
	content, err := os.ReadFile(route.FilePath)
	if err != nil {
		http.Error(mwCtx.Response, err.Error(), http.StatusInternalServerError)
		return
	}

	parsed, err := parser.Parse(string(content))
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Parse error: %v", err), http.StatusInternalServerError)
		return
	}

	resolver := comp.Resolver
	resolver.SetCurrentFile(route.FilePath)

	imports := make([]compiler.Import, len(parsed.Imports))
	for i, imp := range parsed.Imports {
		imports[i] = compiler.Import{
			Path:        imp.Path,
			Alias:       imp.Alias,
			IsComponent: imp.IsComponent,
		}
	}
	resolver.ParseImports(imports)

	ctx := executor.NewContext()

	reqCtx := ssr.NewRequestContext(mwCtx.Request, mwCtx.Params)
	ctx.SetRequest(reqCtx)
	ctx.SetLocals(mwCtx.Locals)

	for k, v := range mwCtx.Params {
		ctx.Set(k, v)
	}

	if parsed.Frontmatter != "" {
		if err := ctx.Execute(parsed.Frontmatter); err != nil {
			http.Error(mwCtx.Response, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	processedTemplate := comp.ProcessComponentTags(parsed.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, nil)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	mwCtx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	mwCtx.Response.Write([]byte(rendered))
}
