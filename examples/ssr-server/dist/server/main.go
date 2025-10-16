package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/endpoints"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/middleware"
	"github.com/galaxy/galaxy/pkg/parser"
	"github.com/galaxy/galaxy/pkg/router"
	"github.com/galaxy/galaxy/pkg/ssr"
	"github.com/galaxy/galaxy/pkg/template"
	"github.com/galaxy/galaxy/pkg/wasm"

	
	api "galaxy-server/pages/api"
	
	
)

var (
	rt           *router.Router
	comp         *compiler.ComponentCompiler
	baseDir      string
	pagesDir     = "pages"
	wasmManifest *wasm.WasmManifest
	routeMap     = map[string]*routeInfo{
		
		"/": {Pattern: "/", FilePath: "pages/index.gxc", IsEndpoint: false},
		
		"/api/hello": {Pattern: "/api/hello", FilePath: "pages/api/hello.go", IsEndpoint: true},
		
		"/counter": {Pattern: "/counter", FilePath: "pages/counter.gxc", IsEndpoint: false},
		
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
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir = filepath.Dir(exePath)
	comp = compiler.NewComponentCompiler(baseDir)

	manifestPath := filepath.Join(baseDir, "_assets", "wasm-manifest.json")
	wasmManifest, _ = wasm.LoadManifest(manifestPath)

	http.HandleFunc("/", handleRequest)

	addr := "localhost:4322"
	log.Printf("🚀 Server running at http://%s\n", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/_assets/") {
		assetsPath := filepath.Join(baseDir, r.URL.Path)
		http.ServeFile(w, r, assetsPath)
		return
	}

	if r.URL.Path == "/wasm_exec.js" {
		wasmExecPath := filepath.Join(baseDir, "wasm_exec.js")
		http.ServeFile(w, r, wasmExecPath)
		return
	}

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
	filePath := filepath.Join(baseDir, route.FilePath)
	content, err := os.ReadFile(filePath)
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
	resolver.SetCurrentFile(filePath)

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

	comp.CollectedStyles = nil
	processedTemplate := comp.ProcessComponentTags(parsed.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, nil)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	allStyles := append(parsed.Styles, comp.CollectedStyles...)
	if len(allStyles) > 0 {
		var styleContent string
		for _, style := range allStyles {
			styleContent += style.Content + "\n"
		}
		styleTag := "<style>" + styleContent + "</style>"
		rendered = strings.Replace(rendered, "</head>", styleTag+"\n</head>", 1)
	}

	if wasmManifest != nil {
		pageAssets, ok := wasmManifest.Assets[route.FilePath]
		if ok && len(pageAssets.WasmModules) > 0 {
			wasmExecTag := "<script src=\"/wasm_exec.js\"></script>"
			rendered = strings.Replace(rendered, "</body>", wasmExecTag+"\n</body>", 1)

			for _, mod := range pageAssets.WasmModules {
				loaderTag := fmt.Sprintf("<script src=\"%s\"></script>", mod.LoaderPath)
				rendered = strings.Replace(rendered, "</body>", loaderTag+"\n</body>", 1)
			}
		}

		if len(pageAssets.JSScripts) > 0 {
			for _, jsPath := range pageAssets.JSScripts {
				jsTag := fmt.Sprintf("<script type=\"module\" src=\"%s\"></script>", jsPath)
				rendered = strings.Replace(rendered, "</body>", jsTag+"\n</body>", 1)
			}
		}
	}

	mwCtx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	mwCtx.Response.Write([]byte(rendered))
}
