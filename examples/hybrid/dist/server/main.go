package main

import (
	"net/http"
	"log"
	
	"strings"
	
	"generated-hybrid/runtime"
)

func main() {
	log.Println("Starting server...")
	
	http.Handle("/_assets/", http.StripPrefix("/_assets/", http.FileServer(http.Dir("_assets"))))
	http.Handle("/wasm_exec.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "wasm_exec.js")
	}))
	
		http.HandleFunc("/dynamic", func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]string)
		HandleDynamic(w, r, params)
	})
	http.HandleFunc("/wasm-dynamic", func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]string)
		HandleWasmDynamic(w, r, params)
	})
	
	addr := ":4322"
	log.Printf("🚀 Server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func extractParams(path, pattern string) map[string]string {
	params := make(map[string]string)
	
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	
	if len(pathParts) != len(patternParts) {
		return params
	}
	
	for i, part := range patternParts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			key := strings.Trim(part, "[]")
			params[key] = pathParts[i]
		} else if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			key := strings.Trim(part, "{}")
			params[key] = pathParts[i]
		}
	}
	
	return params
}

func HandleDynamic(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var title = "Dynamic Page"
// prerender = false
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/dynamic.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleDynamic)
	w.Write([]byte(html))
}

const templateHandleDynamic = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>Dynamic Page (SSR)</h1>
    <p>This page is rendered on-demand.</p>
    <div galaxy:if={Request}>
        <p>Request path: {Request.Path()}</p>
        <p>Request method: {Request.Method()}</p>
    </div>
    <nav>
        <a href="/">Static Home</a>
    </nav>
</body>
</html>`


func HandleWasmDynamic(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	// prerender = false
var title = "Hybrid WASM (Dynamic)"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/wasm-dynamic.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleWasmDynamic)
	w.Write([]byte(html))
}

const templateHandleWasmDynamic = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>{title}</h1>
    <p>Dynamic SSR page with WASM</p>
    
    <div id="app">
        <button id="btn">Click Count: <span id="count">0</span></button>
    </div>
</body>
</html>`

