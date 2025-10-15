package executor

import (
	"testing"
)

func TestExecuteSimpleAssignment(t *testing.T) {
	ctx := NewContext()

	code := `
var title = "Hello"
var count = 42
var pi = 3.14
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("title"); !ok || val != "Hello" {
		t.Errorf("Expected title='Hello', got %v", val)
	}

	if val, ok := ctx.Get("count"); !ok || val != int64(42) {
		t.Errorf("Expected count=42, got %v", val)
	}

	if val, ok := ctx.Get("pi"); !ok || val != 3.14 {
		t.Errorf("Expected pi=3.14, got %v", val)
	}
}

func TestExecuteArithmetic(t *testing.T) {
	ctx := NewContext()

	code := `
var a = 10
var b = 5
var sum = a + b
var diff = a - b
var prod = a * b
var quot = a / b
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	tests := []struct {
		name     string
		expected int64
	}{
		{"sum", 15},
		{"diff", 5},
		{"prod", 50},
		{"quot", 2},
	}

	for _, tt := range tests {
		if val, ok := ctx.Get(tt.name); !ok || val != tt.expected {
			t.Errorf("Expected %s=%d, got %v", tt.name, tt.expected, val)
		}
	}
}

func TestExecuteStringConcat(t *testing.T) {
	ctx := NewContext()

	code := `
var first = "Hello"
var last = "World"
var greeting = first + " " + last
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("greeting"); !ok || val != "Hello World" {
		t.Errorf("Expected greeting='Hello World', got %v", val)
	}
}

func TestPropsAndSlots(t *testing.T) {
	ctx := NewContext()

	ctx.SetProp("title", "Page Title")
	ctx.SetProp("count", 42)

	if val, ok := ctx.GetProp("title"); !ok || val != "Page Title" {
		t.Errorf("Expected prop title='Page Title', got %v", val)
	}

	if val, ok := ctx.GetProp("count"); !ok || val != 42 {
		t.Errorf("Expected prop count=42, got %v", val)
	}
}

func TestDivisionByZero(t *testing.T) {
	ctx := NewContext()

	code := `
var x = 10
var y = 0
var result = x / y
`

	err := ctx.Execute(code)
	if err == nil {
		t.Error("Expected error for division by zero")
	}
}

func TestFloatArithmetic(t *testing.T) {
	ctx := NewContext()

	code := `
var a = 3.5
var b = 2.0
var sum = a + b
var diff = a - b
var prod = a * b
var quot = a / b
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("sum"); !ok || val != 5.5 {
		t.Errorf("Expected sum=5.5, got %v", val)
	}

	if val, ok := ctx.Get("diff"); !ok || val != 1.5 {
		t.Errorf("Expected diff=1.5, got %v", val)
	}

	if val, ok := ctx.Get("prod"); !ok || val != 7.0 {
		t.Errorf("Expected prod=7.0, got %v", val)
	}

	if val, ok := ctx.Get("quot"); !ok || val != 1.75 {
		t.Errorf("Expected quot=1.75, got %v", val)
	}
}

func TestFloatDivisionByZero(t *testing.T) {
	ctx := NewContext()

	code := `
var x = 5.5
var y = 0.0
var result = x / y
`

	err := ctx.Execute(code)
	if err == nil {
		t.Error("Expected error for float division by zero")
	}
}

func TestUnaryNegation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("x", int64(42))
	ctx.Set("y", 3.14)

	code := `
var negX = -x
var negY = -y
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("negX"); !ok || val != int64(-42) {
		t.Errorf("Expected negX=-42, got %v", val)
	}

	if val, ok := ctx.Get("negY"); !ok || val != -3.14 {
		t.Errorf("Expected negY=-3.14, got %v", val)
	}
}

func TestArrayLiteral(t *testing.T) {
	ctx := NewContext()

	code := `var items = []string{"apple", "banana", "cherry"}`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	val, ok := ctx.Get("items")
	if !ok {
		t.Fatal("Expected items to be set")
	}

	arr, ok := val.([]interface{})
	if !ok {
		t.Fatalf("Expected array, got %T", val)
	}

	if len(arr) != 3 {
		t.Errorf("Expected 3 items, got %d", len(arr))
	}

	if arr[0] != "apple" || arr[1] != "banana" || arr[2] != "cherry" {
		t.Errorf("Array values mismatch: %v", arr)
	}
}

func TestLenFunction(t *testing.T) {
	ctx := NewContext()
	ctx.Set("items", []interface{}{"a", "b", "c"})
	ctx.Set("text", "hello")

	code := `
var itemCount = len(items)
var textLen = len(text)
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("itemCount"); !ok || val != int64(3) {
		t.Errorf("Expected itemCount=3, got %v", val)
	}

	if val, ok := ctx.Get("textLen"); !ok || val != int64(5) {
		t.Errorf("Expected textLen=5, got %v", val)
	}
}

func TestSetAndGet(t *testing.T) {
	ctx := NewContext()

	ctx.Set("key1", "value1")
	ctx.Set("key2", 123)

	val, ok := ctx.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected key1=value1, got %v", val)
	}

	val, ok = ctx.Get("key2")
	if !ok || val != 123 {
		t.Errorf("Expected key2=123, got %v", val)
	}

	_, ok = ctx.Get("nonexistent")
	if ok {
		t.Error("Expected false for nonexistent key")
	}
}

func TestSetRequest(t *testing.T) {
	ctx := NewContext()

	req := struct{ Method string }{Method: "GET"}
	ctx.SetRequest(req)

	gotReq, ok := ctx.GetRequest()
	if !ok {
		t.Fatal("Expected request to be set")
	}

	reqStruct, ok := gotReq.(struct{ Method string })
	if !ok || reqStruct.Method != "GET" {
		t.Errorf("Expected request method GET, got %v", gotReq)
	}

	valFromVar, ok := ctx.Get("Request")
	if !ok {
		t.Error("Expected Request variable to be set")
	}

	if valFromVar != req {
		t.Error("Expected Request variable to match SetRequest value")
	}
}

func TestGetRequestEmpty(t *testing.T) {
	ctx := NewContext()

	_, ok := ctx.GetRequest()
	if ok {
		t.Error("Expected false when no request is set")
	}
}

func TestUndefinedVariable(t *testing.T) {
	ctx := NewContext()

	code := `var result = undefinedVar + 5`

	err := ctx.Execute(code)
	if err == nil {
		t.Error("Expected error for undefined variable")
	}
}

func TestInvalidOperandTypes(t *testing.T) {
	ctx := NewContext()
	ctx.Set("str", "hello")
	ctx.Set("num", int64(42))

	code := `var result = str * num`

	err := ctx.Execute(code)
	if err == nil {
		t.Error("Expected error for invalid operand types")
	}
}

func TestContextString(t *testing.T) {
	ctx := NewContext()
	ctx.Set("x", 10)
	ctx.Set("y", "test")

	str := ctx.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}
