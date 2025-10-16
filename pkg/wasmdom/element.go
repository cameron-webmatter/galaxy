//go:build js && wasm
// +build js,wasm

package wasmdom

import (
	"syscall/js"
)

type Element struct {
	Value js.Value
}

func GetElementById(id string) Element {
	doc := js.Global().Get("document")
	return Element{Value: doc.Call("getElementById", id)}
}

func QuerySelector(selector string) Element {
	doc := js.Global().Get("document")
	return Element{Value: doc.Call("querySelector", selector)}
}

func QuerySelectorAll(selector string) []Element {
	doc := js.Global().Get("document")
	nodeList := doc.Call("querySelectorAll", selector)
	length := nodeList.Get("length").Int()
	elements := make([]Element, length)
	for i := 0; i < length; i++ {
		elements[i] = Element{Value: nodeList.Call("item", i)}
	}
	return elements
}

func CreateElement(tag string) Element {
	doc := js.Global().Get("document")
	return Element{Value: doc.Call("createElement", tag)}
}

func (e Element) SetInnerHTML(html string) {
	e.Value.Set("innerHTML", html)
}

func (e Element) GetInnerHTML() string {
	return e.Value.Get("innerHTML").String()
}

func (e Element) SetTextContent(text string) {
	e.Value.Set("textContent", text)
}

func (e Element) GetTextContent() string {
	return e.Value.Get("textContent").String()
}

func (e Element) AddEventListener(event string, handler func()) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handler()
		return nil
	})
	e.Value.Call("addEventListener", event, cb)
}

func (e Element) AddEventListenerWithEvent(event string, handler func(js.Value)) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			handler(args[0])
		}
		return nil
	})
	e.Value.Call("addEventListener", event, cb)
}

func (e Element) SetAttribute(name, value string) {
	e.Value.Call("setAttribute", name, value)
}

func (e Element) GetAttribute(name string) string {
	return e.Value.Call("getAttribute", name).String()
}

func (e Element) RemoveAttribute(name string) {
	e.Value.Call("removeAttribute", name)
}

func (e Element) AppendChild(child Element) {
	e.Value.Call("appendChild", child.Value)
}

func (e Element) RemoveChild(child Element) {
	e.Value.Call("removeChild", child.Value)
}

func (e Element) Remove() {
	e.Value.Call("remove")
}

func (e Element) AddClass(class string) {
	e.Value.Get("classList").Call("add", class)
}

func (e Element) RemoveClass(class string) {
	e.Value.Get("classList").Call("remove", class)
}

func (e Element) ToggleClass(class string) {
	e.Value.Get("classList").Call("toggle", class)
}

func (e Element) HasClass(class string) bool {
	return e.Value.Get("classList").Call("contains", class).Bool()
}

func (e Element) SetStyle(property, value string) {
	e.Value.Get("style").Set(property, value)
}

func (e Element) GetValue() string {
	return e.Value.Get("value").String()
}

func (e Element) SetValue(value string) {
	e.Value.Set("value", value)
}

func (e Element) Focus() {
	e.Value.Call("focus")
}

func (e Element) Blur() {
	e.Value.Call("blur")
}

func (e Element) IsNull() bool {
	return e.Value.IsNull() || e.Value.IsUndefined()
}
