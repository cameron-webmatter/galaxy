package main

import (
	"net/http"
	"log"
	"regexp"
	"strings"
	
	"ssr-server/runtime"
)

func main() {
	log.Println("Starting server...")
	chain := NewMiddlewareChain()
	
	
	http.Handle("/_assets/", http.StripPrefix("/_assets/", http.FileServer(http.Dir("_assets"))))
	http.Handle("/wasm_exec.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "wasm_exec.js")
	}))
	
		http.HandleFunc("/counter", func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]string)
		chain.Execute(w, r, func(w http.ResponseWriter, r *http.Request, locals map[string]interface{}) {
			HandleCounter(w, r, params, locals)
		})
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			params := make(map[string]string)
			chain.Execute(w, r, func(w http.ResponseWriter, r *http.Request, locals map[string]interface{}) {
				HandleIndex(w, r, params, locals)
			})
			return
		}
		if regexp.MustCompile(`^\/user\/([^\/]+)$`).MatchString(r.URL.Path) {
			params := extractParams(r.URL.Path, "/user/[id]")
			chain.Execute(w, r, func(w http.ResponseWriter, r *http.Request, locals map[string]interface{}) {
				HandleUserId(w, r, params, locals)
			})
			return
		}
		http.NotFound(w, r)
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

type MiddlewareChain struct {
	middlewares []MiddlewareFunc
}

type MiddlewareFunc func(w http.ResponseWriter, r *http.Request, locals map[string]interface{}, next func())

func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]MiddlewareFunc, 0),
	}
}

func (c *MiddlewareChain) Use(mw MiddlewareFunc) {
	c.middlewares = append(c.middlewares, mw)
}

func (c *MiddlewareChain) Execute(w http.ResponseWriter, r *http.Request, handler func(http.ResponseWriter, *http.Request, map[string]interface{})) {
	locals := make(map[string]interface{})
	
	var runMiddleware func(int)
	runMiddleware = func(index int) {
		if index >= len(c.middlewares) {
			handler(w, r, locals)
			return
		}
		
		c.middlewares[index](w, r, locals, func() {
			runMiddleware(index + 1)
		})
	}
	
	runMiddleware(0)
}


func HandleIndex(w http.ResponseWriter, r *http.Request, params map[string]string, locals map[string]interface{}) {
	
	
	_ = locals
	var title = "SSR Server Example"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/index.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	for k, v := range locals {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleIndex)
	w.Write([]byte(html))
}

const templateHandleIndex = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>Galaxy SSR Server Mode</h1>
    <p>This page is rendered on-demand using Go!</p>
    
    <div galaxy:if={Request}>
        <h2>Request Info</h2>
        <p>Path: {Request.Path()}</p>
        <p>Method: {Request.Method()}</p>
    </div>

    <h2>Middleware Data</h2>
    <p>Timestamp: {Locals.timestamp}</p>
    <p>Server Name: {Locals.serverName}</p>
</body>
</html>`


func HandleCounter(w http.ResponseWriter, r *http.Request, params map[string]string, locals map[string]interface{}) {
	
	
	_ = locals
	var title = "SSR WASM Counter"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/counter.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	for k, v := range locals {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleCounter)
	w.Write([]byte(html))
}

const templateHandleCounter = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>{title}</h1>
    <p>Server-rendered page with WASM</p>
    
    <div id="counter">
        <button id="increment">+</button>
        <span id="count">0</span>
        <button id="decrement">-</button>
    </div>
</body>
</html>`


func HandleUserId(w http.ResponseWriter, r *http.Request, params map[string]string, locals map[string]interface{}) {
		id := params["id"]
	
	_ = locals
	var title = "User Profile"
var userId = id
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/user/[id].gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	for k, v := range locals {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	ctx.Set("userId", userId)
	
	html := runtime.RenderTemplate(ctx, templateHandleUserId)
	w.Write([]byte(html))
}

const templateHandleUserId = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>User Profile Page</h1>
    
    <div>
        <h2>Galaxy.Params Test</h2>
        <p>User ID via Galaxy.Params: <strong>{userId}</strong></p>
        <p>User ID via direct access: <strong>{id}</strong></p>
    </div>

    <div galaxy:if={Request}>
        <p>Path: {Request.Path()}</p>
        <p>Method: {Request.Method()}</p>
    </div>

    <a href="/">← Home</a>
</body>
</html>`

