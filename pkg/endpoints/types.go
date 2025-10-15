package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HandlerFunc func(*Context) error

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Params   map[string]string
	Locals   map[string]any
}

func NewContext(w http.ResponseWriter, r *http.Request, params map[string]string, locals map[string]any) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Params:   params,
		Locals:   locals,
	}
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) JSON(status int, data any) error {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(status)
	return json.NewEncoder(c.Response).Encode(data)
}

func (c *Context) Text(status int, text string) error {
	c.Response.Header().Set("Content-Type", "text/plain")
	c.Response.WriteHeader(status)
	_, err := c.Response.Write([]byte(text))
	return err
}

func (c *Context) HTML(status int, html string) error {
	c.Response.Header().Set("Content-Type", "text/html")
	c.Response.WriteHeader(status)
	_, err := c.Response.Write([]byte(html))
	return err
}

func (c *Context) Redirect(url string, status int) error {
	http.Redirect(c.Response, c.Request, url, status)
	return nil
}

func (c *Context) BindJSON(v any) error {
	if c.Request.Body == nil {
		return fmt.Errorf("request body is nil")
	}
	defer c.Request.Body.Close()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

func (c *Context) BindForm(v any) error {
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	return nil
}
