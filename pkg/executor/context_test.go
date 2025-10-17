package executor

import (
	"fmt"
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

func TestIfStatement(t *testing.T) {
	ctx := NewContext()

	code := `
if 5 > 3 {
	var result = "greater"
}
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("result"); !ok || val != "greater" {
		t.Errorf("Expected result='greater', got %v", val)
	}
}

func TestIfElseStatement(t *testing.T) {
	ctx := NewContext()

	code := `
if 3 > 5 {
	var result = "greater"
} else {
	var result = "not greater"
}
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("result"); !ok || val != "not greater" {
		t.Errorf("Expected result='not greater', got %v", val)
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"var x = 5 == 5", true},
		{"var x = 5 != 3", true},
		{"var x = 5 > 3", true},
		{"var x = 5 < 3", false},
		{"var x = 5 >= 5", true},
		{"var x = 3 <= 5", true},
	}

	for _, tt := range tests {
		ctx := NewContext()
		err := ctx.Execute(tt.code)
		if err != nil {
			t.Fatalf("Execute failed for %s: %v", tt.code, err)
		}

		if val, ok := ctx.Get("x"); !ok || val != tt.expected {
			t.Errorf("For %s: expected %v, got %v", tt.code, tt.expected, val)
		}
	}
}

func TestNilComparison(t *testing.T) {
	ctx := NewContext()

	locals := make(map[string]any)
	locals["user"] = nil
	ctx.SetLocals(locals)

	code := `
if Locals.user == nil {
	var result = "no user"
}
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("result"); !ok || val != "no user" {
		t.Errorf("Expected result='no user', got %v", val)
	}
}

func TestGalaxyRedirect(t *testing.T) {
	ctx := NewContext()

	code := `Galaxy.redirect("/login", 302)`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !ctx.ShouldRedirect {
		t.Error("Expected ShouldRedirect to be true")
	}

	if ctx.RedirectURL != "/login" {
		t.Errorf("Expected RedirectURL='/login', got %s", ctx.RedirectURL)
	}

	if ctx.RedirectStatus != 302 {
		t.Errorf("Expected RedirectStatus=302, got %d", ctx.RedirectStatus)
	}
}

func TestConditionalRedirect(t *testing.T) {
	ctx := NewContext()

	locals := make(map[string]any)
	locals["user"] = nil
	ctx.SetLocals(locals)

	code := `
if Locals.user == nil {
	Galaxy.redirect("/login", 302)
}
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !ctx.ShouldRedirect {
		t.Error("Expected ShouldRedirect to be true")
	}

	if ctx.RedirectURL != "/login" {
		t.Errorf("Expected RedirectURL='/login', got %s", ctx.RedirectURL)
	}
}

func TestRedirectStopsExecution(t *testing.T) {
	ctx := NewContext()

	code := `
Galaxy.redirect("/login", 302)
var shouldNotRun = "this should not execute"
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if _, ok := ctx.Get("shouldNotRun"); ok {
		t.Error("Expected code after redirect to not execute")
	}

	if !ctx.ShouldRedirect {
		t.Error("Expected ShouldRedirect to be true")
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"and true true", "var x = (5 > 3) && (10 > 5)", true},
		{"and true false", "var x = (5 > 3) && (3 > 10)", false},
		{"or false true", "var x = (3 > 5) || (10 > 5)", true},
		{"or false false", "var x = (3 > 5) || (5 > 10)", false},
	}

	for _, tt := range tests {
		ctx := NewContext()
		err := ctx.Execute(tt.code)
		if err != nil {
			t.Fatalf("Execute failed for %s: %v", tt.name, err)
		}

		if val, ok := ctx.Get("x"); !ok || val != tt.expected {
			t.Errorf("For %s: expected %v, got %v", tt.name, tt.expected, val)
		}
	}
}

func TestSingleImportExtraction(t *testing.T) {
	ctx := NewContext()

	code := `import "database/sql"

var x = "test"`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with import failed: %v", err)
	}

	if val, ok := ctx.Get("x"); !ok || val != "test" {
		t.Errorf("Expected x='test', got %v", val)
	}
}

func TestMultipleImportExtraction(t *testing.T) {
	ctx := NewContext()

	code := `import "database/sql"
import "fmt"

var x = "test"`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with imports failed: %v", err)
	}

	if val, ok := ctx.Get("x"); !ok || val != "test" {
		t.Errorf("Expected x='test', got %v", val)
	}
}

func TestGroupedImportExtraction(t *testing.T) {
	ctx := NewContext()

	code := `import (
	"database/sql"
	"fmt"
)

var x = "test"`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with grouped imports failed: %v", err)
	}

	if val, ok := ctx.Get("x"); !ok || val != "test" {
		t.Errorf("Expected x='test', got %v", val)
	}
}

func TestAliasedImportExtraction(t *testing.T) {
	ctx := NewContext()

	code := `import db "database/sql"

var x = "test"`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with aliased import failed: %v", err)
	}

	if val, ok := ctx.Get("x"); !ok || val != "test" {
		t.Errorf("Expected x='test', got %v", val)
	}
}

func TestMixedImportsAndCode(t *testing.T) {
	ctx := NewContext()

	code := `import "fmt"

var title = "Hello"

import "database/sql"

var count = 42`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with mixed imports failed: %v", err)
	}

	if val, ok := ctx.Get("title"); !ok || val != "Hello" {
		t.Errorf("Expected title='Hello', got %v", val)
	}

	if val, ok := ctx.Get("count"); !ok || val != int64(42) {
		t.Errorf("Expected count=42, got %v", val)
	}
}

func TestTypeAssertion(t *testing.T) {
	ctx := NewContext()

	code := `var x interface{} = "hello"
var str = x.(string)`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with type assertion failed: %v", err)
	}

	if val, ok := ctx.Get("str"); !ok || val != "hello" {
		t.Errorf("Expected str='hello', got %v", val)
	}
}

func TestPackageFuncCall(t *testing.T) {
	ctx := NewContext()

	ctx.RegisterPackageFunc("models", "GetUser", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("GetUser expects 1 argument")
		}
		id := args[0].(string)
		return map[string]interface{}{"id": id, "name": "John"}, nil
	})

	code := `import "models"

var user = models.GetUser("123")`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with package func failed: %v", err)
	}

	val, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user to be set")
	}

	userMap, ok := val.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user to be map, got %T", val)
	}

	if userMap["id"] != "123" || userMap["name"] != "John" {
		t.Errorf("Expected user={id:123, name:John}, got %v", userMap)
	}
}

func TestGlobalFuncRegistry(t *testing.T) {
	RegisterGlobalFunc("testpkg", "GetData", func(args ...interface{}) (interface{}, error) {
		return "global data", nil
	})

	ctx := NewContext()

	code := `import "testpkg"

var data = testpkg.GetData()`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute with global func failed: %v", err)
	}

	val, ok := ctx.Get("data")
	if !ok || val != "global data" {
		t.Errorf("Expected data='global data', got %v", val)
	}
}

func TestMultiValueAssignment(t *testing.T) {
	ctx := NewContext()

	ctx.RegisterPackageFunc("models", "GetUser", func(args ...interface{}) (interface{}, error) {
		return map[string]interface{}{"id": "123"}, nil
	})

	code := `import "models"

var user, err = models.GetUser("123")`

	execErr := ctx.Execute(code)
	if execErr != nil {
		t.Fatalf("Execute failed: %v", execErr)
	}

	user, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user to be set")
	}

	if user == nil {
		t.Error("Expected user to have value")
	}

	errVal, ok := ctx.Get("err")
	if !ok {
		t.Fatal("Expected err to be set")
	}

	if errVal != nil {
		t.Error("Expected err to be nil")
	}
}
