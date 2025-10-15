package ssr

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNewRequestContext(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test", RawQuery: "foo=bar&baz=qux"},
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer token"},
		},
	}

	pathParams := map[string]string{
		"id":   "123",
		"slug": "hello-world",
	}

	ctx := NewRequestContext(req, pathParams)

	if ctx.Request != req {
		t.Error("Expected Request to be set")
	}

	if len(ctx.PathParams) != 2 {
		t.Errorf("Expected 2 path params, got %d", len(ctx.PathParams))
	}

	if len(ctx.Query) != 2 {
		t.Errorf("Expected 2 query params, got %d", len(ctx.Query))
	}

	if len(ctx.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(ctx.Headers))
	}
}

func TestNewRequestContextEmptyQuery(t *testing.T) {
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/api/user"},
		Header: http.Header{},
	}

	ctx := NewRequestContext(req, nil)

	if len(ctx.Query) != 0 {
		t.Errorf("Expected empty query, got %d items", len(ctx.Query))
	}

	if len(ctx.Headers) != 0 {
		t.Errorf("Expected empty headers, got %d items", len(ctx.Headers))
	}
}

func TestNewRequestContextMultipleQueryValues(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{RawQuery: "tag=go&tag=rust&tag=python"},
	}

	ctx := NewRequestContext(req, nil)

	if ctx.Query["tag"] != "go" {
		t.Errorf("Expected first value 'go', got %s", ctx.Query["tag"])
	}
}

func TestRequestContextMethod(t *testing.T) {
	req := &http.Request{Method: "DELETE", URL: &url.URL{}}
	ctx := NewRequestContext(req, nil)

	if ctx.Method() != "DELETE" {
		t.Errorf("Expected DELETE, got %s", ctx.Method())
	}
}

func TestRequestContextURL(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Scheme:   "https",
			Host:     "example.com",
			Path:     "/api/test",
			RawQuery: "key=value",
		},
	}
	ctx := NewRequestContext(req, nil)

	expected := "https://example.com/api/test?key=value"
	if ctx.URL() != expected {
		t.Errorf("Expected %s, got %s", expected, ctx.URL())
	}
}

func TestRequestContextPath(t *testing.T) {
	req := &http.Request{URL: &url.URL{Path: "/blog/post/123"}}
	ctx := NewRequestContext(req, nil)

	if ctx.Path() != "/blog/post/123" {
		t.Errorf("Expected /blog/post/123, got %s", ctx.Path())
	}
}

func TestRequestContextParam(t *testing.T) {
	pathParams := map[string]string{
		"id":   "456",
		"slug": "my-post",
	}

	req := &http.Request{URL: &url.URL{}}
	ctx := NewRequestContext(req, pathParams)

	if ctx.Param("id") != "456" {
		t.Errorf("Expected 456, got %s", ctx.Param("id"))
	}

	if ctx.Param("slug") != "my-post" {
		t.Errorf("Expected my-post, got %s", ctx.Param("slug"))
	}

	if ctx.Param("nonexistent") != "" {
		t.Error("Expected empty string for missing param")
	}
}

func TestRequestContextQueryParam(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{RawQuery: "search=golang&page=2"},
	}
	ctx := NewRequestContext(req, nil)

	if ctx.QueryParam("search") != "golang" {
		t.Errorf("Expected golang, got %s", ctx.QueryParam("search"))
	}

	if ctx.QueryParam("page") != "2" {
		t.Errorf("Expected 2, got %s", ctx.QueryParam("page"))
	}

	if ctx.QueryParam("missing") != "" {
		t.Error("Expected empty string for missing query param")
	}
}

func TestRequestContextHeader(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{},
		Header: http.Header{
			"User-Agent": []string{"TestAgent/1.0"},
			"Accept":     []string{"application/json"},
		},
	}
	ctx := NewRequestContext(req, nil)

	if ctx.Header("User-Agent") != "TestAgent/1.0" {
		t.Errorf("Expected TestAgent/1.0, got %s", ctx.Header("User-Agent"))
	}

	if ctx.Header("Accept") != "application/json" {
		t.Errorf("Expected application/json, got %s", ctx.Header("Accept"))
	}

	if ctx.Header("Missing") != "" {
		t.Error("Expected empty string for missing header")
	}
}

func TestRequestContextNilPathParams(t *testing.T) {
	req := &http.Request{URL: &url.URL{}}
	ctx := NewRequestContext(req, nil)

	if ctx.PathParams != nil {
		t.Errorf("Expected PathParams to be nil, got %v", ctx.PathParams)
	}

	if ctx.Param("any") != "" {
		t.Error("Expected empty string for param when PathParams is nil")
	}
}

func TestRequestContextCompleteScenario(t *testing.T) {
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path:     "/api/posts/123",
			RawQuery: "filter=published&sort=desc",
		},
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer abc123"},
		},
	}

	pathParams := map[string]string{
		"id": "123",
	}

	ctx := NewRequestContext(req, pathParams)

	if ctx.Method() != "POST" {
		t.Error("Method mismatch")
	}

	if ctx.Path() != "/api/posts/123" {
		t.Error("Path mismatch")
	}

	if ctx.Param("id") != "123" {
		t.Error("Path param mismatch")
	}

	if ctx.QueryParam("filter") != "published" {
		t.Error("Query param mismatch")
	}

	if ctx.Header("Content-Type") != "application/json" {
		t.Error("Header mismatch")
	}
}
