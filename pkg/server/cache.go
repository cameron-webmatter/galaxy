package server

import (
	"net/http"
	"sync"
)

type PagePlugin struct {
	Handler         func(http.ResponseWriter, *http.Request, map[string]string, map[string]interface{})
	Template        string
	FrontmatterHash string
	TemplateHash    string
	PluginPath      string
}

type PageCache struct {
	mu    sync.RWMutex
	pages map[string]*PagePlugin
}

func NewPageCache() *PageCache {
	return &PageCache{
		pages: make(map[string]*PagePlugin),
	}
}

func (c *PageCache) Get(pattern string) (*PagePlugin, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	plugin, ok := c.pages[pattern]
	return plugin, ok
}

func (c *PageCache) Set(pattern string, plugin *PagePlugin) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pages[pattern] = plugin
}

func (c *PageCache) Invalidate(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pages, pattern)
}
