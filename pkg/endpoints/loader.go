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
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("open plugin: %w", err)
	}

	endpoint := &LoadedEndpoint{
		Handlers: make(map[HTTPMethod]HandlerFunc),
	}

	methods := []HTTPMethod{GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, ALL}
	for _, method := range methods {
		sym, err := p.Lookup(string(method))
		if err != nil {
			continue
		}

		handler, ok := sym.(func(*Context) error)
		if !ok {
			continue
		}

		endpoint.Handlers[method] = handler
	}

	if len(endpoint.Handlers) == 0 {
		return nil, fmt.Errorf("no HTTP method handlers found")
	}

	return endpoint, nil
}
