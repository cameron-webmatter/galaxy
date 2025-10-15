package endpoints

import (
	"fmt"
	"net/http"
)

func HandleEndpoint(endpoint *LoadedEndpoint, w http.ResponseWriter, r *http.Request, params map[string]string, locals map[string]any) error {
	method := HTTPMethod(r.Method)

	handler, ok := endpoint.Handlers[method]
	if !ok {
		handler, ok = endpoint.Handlers[ALL]
		if !ok {
			http.Error(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
			return nil
		}
	}

	ctx := NewContext(w, r, params, locals)
	return handler(ctx)
}
