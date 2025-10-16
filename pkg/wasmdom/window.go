//go:build js && wasm
// +build js,wasm

package wasmdom

import (
	"syscall/js"
)

func Alert(message string) {
	js.Global().Call("alert", message)
}

func ConsoleLog(args ...interface{}) {
	console := js.Global().Get("console")
	jsArgs := make([]interface{}, len(args))
	for i, arg := range args {
		jsArgs[i] = arg
	}
	console.Call("log", jsArgs...)
}

func ConsoleError(args ...interface{}) {
	console := js.Global().Get("console")
	jsArgs := make([]interface{}, len(args))
	for i, arg := range args {
		jsArgs[i] = arg
	}
	console.Call("error", jsArgs...)
}

func SetTimeout(callback func(), ms int) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		callback()
		return nil
	})
	js.Global().Call("setTimeout", cb, ms)
}

func SetInterval(callback func(), ms int) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		callback()
		return nil
	})
	js.Global().Call("setInterval", cb, ms)
}

func RequestAnimationFrame(callback func()) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		callback()
		return nil
	})
	js.Global().Call("requestAnimationFrame", cb)
}

func AddWindowEventListener(event string, handler func()) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handler()
		return nil
	})
	js.Global().Call("addEventListener", event, cb)
}

func GetWindowWidth() int {
	return js.Global().Get("innerWidth").Int()
}

func GetWindowHeight() int {
	return js.Global().Get("innerHeight").Int()
}
