//go:build js && wasm
// +build js,wasm

package wasmdom

import (
	"syscall/js"
)

type FetchResponse struct {
	jsValue js.Value
}

func (r FetchResponse) Status() int {
	return r.jsValue.Get("status").Int()
}

func (r FetchResponse) JSON() js.Value {
	return r.jsValue
}

func Fetch(url string, callback func(FetchResponse)) {
	fetchOpts := js.Global().Get("Object").New()
	fetchOpts.Set("method", "GET")

	promise := js.Global().Call("fetch", url, fetchOpts)
	
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		return response.Call("json")
	})).Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonData := args[0]
		callback(FetchResponse{jsValue: jsonData})
		return nil
	}))
}

func FetchWithOptions(url string, method string, headers map[string]string, body string, callback func(FetchResponse)) {
	fetchOpts := js.Global().Get("Object").New()
	fetchOpts.Set("method", method)
	
	if len(headers) > 0 {
		headersObj := js.Global().Get("Object").New()
		for k, v := range headers {
			headersObj.Set(k, v)
		}
		fetchOpts.Set("headers", headersObj)
	}
	
	if body != "" {
		fetchOpts.Set("body", body)
	}

	promise := js.Global().Call("fetch", url, fetchOpts)
	
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		status := response.Get("status").Int()
		
		if status >= 200 && status < 300 {
			return response.Call("json")
		}
		
		emptyObj := js.Global().Get("Object").New()
		emptyObj.Set("status", status)
		return emptyObj
	})).Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonData := args[0]
		callback(FetchResponse{jsValue: jsonData})
		return nil
	}))
}

func Prompt(message string) string {
	result := js.Global().Call("prompt", message)
	if result.IsNull() {
		return ""
	}
	return result.String()
}

func WindowLocation(url string) {
	js.Global().Get("window").Get("location").Set("href", url)
}
