package endpoints

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

type HTTPMethod string

const (
	GET     HTTPMethod = "GET"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	PATCH   HTTPMethod = "PATCH"
	DELETE  HTTPMethod = "DELETE"
	HEAD    HTTPMethod = "HEAD"
	OPTIONS HTTPMethod = "OPTIONS"
	ALL     HTTPMethod = "ALL"
)

type LoadedEndpoint struct {
	Handlers map[HTTPMethod]HandlerFunc
}

func Load(projectDir string) (map[string]*LoadedEndpoint, error) {
	endpoints := make(map[string]*LoadedEndpoint)

	pagesDir := filepath.Join(projectDir, "pages")
	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		return endpoints, nil
	}

	return endpoints, nil
}

func LoadPlugin(pluginPath string) (*LoadedEndpoint, error) {
	fmt.Printf("DEBUG: Loading plugin from %s\n", pluginPath)
	p, err := plugin.Open(pluginPath)
	if err != nil {
		fmt.Printf("DEBUG: plugin.Open failed: %v\n", err)
		return nil, fmt.Errorf("open plugin: %w", err)
	}

	endpoint := &LoadedEndpoint{
		Handlers: make(map[HTTPMethod]HandlerFunc),
	}

	methods := []HTTPMethod{GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, ALL}
	for _, method := range methods {
		sym, err := p.Lookup(string(method))
		if err != nil {
			fmt.Printf("DEBUG: Lookup %s failed: %v\n", method, err)
			continue
		}

		handler, ok := sym.(func(*Context) error)
		if !ok {
			fmt.Printf("DEBUG: Symbol %s wrong type: %T\n", method, sym)
			continue
		}

		endpoint.Handlers[method] = handler
		fmt.Printf("DEBUG: Loaded handler for %s\n", method)
	}

	if len(endpoint.Handlers) == 0 {
		fmt.Printf("DEBUG: No handlers found in plugin\n")
		return nil, fmt.Errorf("no HTTP method handlers found")
	}

	fmt.Printf("DEBUG: Successfully loaded %d handlers\n", len(endpoint.Handlers))
	return endpoint, nil
}
