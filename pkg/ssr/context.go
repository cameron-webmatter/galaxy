package ssr

import (
	"net/http"
)

type RequestContext struct {
	Request    *http.Request
	PathParams map[string]string
	Query      map[string]string
	Headers    map[string]string
}

func NewRequestContext(r *http.Request, pathParams map[string]string) *RequestContext {
	query := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &RequestContext{
		Request:    r,
		PathParams: pathParams,
		Query:      query,
		Headers:    headers,
	}
}

func (rc *RequestContext) Method() string {
	return rc.Request.Method
}

func (rc *RequestContext) URL() string {
	return rc.Request.URL.String()
}

func (rc *RequestContext) Path() string {
	return rc.Request.URL.Path
}

func (rc *RequestContext) Param(key string) string {
	return rc.PathParams[key]
}

func (rc *RequestContext) QueryParam(key string) string {
	return rc.Query[key]
}

func (rc *RequestContext) Header(key string) string {
	return rc.Headers[key]
}
