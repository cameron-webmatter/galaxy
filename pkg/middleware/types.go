package middleware

import (
	"net/http"
)

type HandlerFunc func(*Context) error

type Middleware func(*Context, func() error) error

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Params   map[string]string
	Locals   map[string]any

	index      int
	middleware []Middleware
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Params:   make(map[string]string),
		Locals:   make(map[string]any),
		index:    -1,
	}
}

func (c *Context) Set(key string, value any) {
	c.Locals[key] = value
}

func (c *Context) Get(key string) any {
	return c.Locals[key]
}

func (c *Context) Next() error {
	c.index++
	if c.index < len(c.middleware) {
		return c.middleware[c.index](c, c.Next)
	}
	return nil
}

func (c *Context) Redirect(url string, status int) {
	http.Redirect(c.Response, c.Request, url, status)
}

func (c *Context) Rewrite(path string) {
	c.Request.URL.Path = path
}
