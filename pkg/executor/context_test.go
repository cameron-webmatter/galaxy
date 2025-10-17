package executor

import (
	"fmt"
	"reflect"
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

func TestMethodInvocation(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	ts := &TestStruct{Value: "hello"}

	ctx := NewContext()
	ctx.Set("obj", ts)

	code := `var result = obj.Value`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	result, ok := ctx.Get("result")
	if !ok {
		t.Fatal("Expected result to be set")
	}

	if result != "hello" {
		t.Errorf("Expected result='hello', got %v", result)
	}
}

type Calculator struct{}

func (c *Calculator) Add(a, b int) int {
	return a + b
}

func (c *Calculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func TestMethodCall(t *testing.T) {
	calc := &Calculator{}

	ctx := NewContext()
	ctx.Set("calc", calc)

	code := `var sum = calc.Add(5, 3)`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	sum, ok := ctx.Get("sum")
	if !ok {
		t.Fatal("Expected sum to be set")
	}

	if sum != 8 {
		t.Errorf("Expected sum=8, got %v", sum)
	}
}

func TestMethodCallWithReturn(t *testing.T) {
	type User struct {
		Name string
	}

	user := &User{Name: "John"}

	ctx := NewContext()
	ctx.Set("user", user)

	code := `var name = user.Name`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	name, ok := ctx.Get("name")
	if !ok {
		t.Fatal("Expected name to be set")
	}

	if name != "John" {
		t.Errorf("Expected name='John', got %v", name)
	}
}

func TestMethodCallWithErrorReturn(t *testing.T) {
	calc := &Calculator{}

	ctx := NewContext()
	ctx.Set("calc", calc)

	code := `var result, err = calc.Divide(10, 2)`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	result, ok := ctx.Get("result")
	if !ok {
		t.Fatal("Expected result to be set")
	}

	if result != 5 {
		t.Errorf("Expected result=5, got %v", result)
	}

	errVal, ok := ctx.Get("err")
	if !ok {
		t.Fatal("Expected err to be set")
	}

	if errVal != nil {
		t.Errorf("Expected err=nil, got %v", errVal)
	}
}

func TestPointerOperation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("value", 42)

	code := `var ptr = &value`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	ptr, ok := ctx.Get("ptr")
	if !ok {
		t.Fatal("Expected ptr to be set")
	}

	// Check it's a pointer
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		t.Errorf("Expected pointer, got %v", v.Kind())
	}

	// Dereference and check value
	if v.Elem().Interface() != 42 {
		t.Errorf("Expected *ptr=42, got %v", v.Elem().Interface())
	}
}

type StringBuilder struct {
	data string
}

func (sb *StringBuilder) Append(s string) *StringBuilder {
	sb.data += s
	return sb
}

func (sb *StringBuilder) String() string {
	return sb.data
}

func TestChainedMethodCalls(t *testing.T) {
	sb := &StringBuilder{}

	ctx := NewContext()
	ctx.Set("sb", sb)

	code := `var result = sb.Append("Hello").Append(" ").Append("World").String()`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	result, ok := ctx.Get("result")
	if !ok {
		t.Fatal("Expected result to be set")
	}

	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got %v", result)
	}
}

type UserRepository struct {
	users map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: map[string]string{"1": "John"},
	}
}

func (r *UserRepository) FindByID(id string) (string, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return "", fmt.Errorf("user not found")
}

func TestConstructorFunction(t *testing.T) {
	ctx := NewContext()

	// Register constructor as package function
	ctx.RegisterPackageFunc("repo", "NewUserRepository", func(args ...interface{}) (interface{}, error) {
		return NewUserRepository(), nil
	})

	code := `import "repo"

var userRepo = repo.NewUserRepository()
var user, err = userRepo.FindByID("1")`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	user, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user to be set")
	}

	if user != "John" {
		t.Errorf("Expected user='John', got %v", user)
	}

	errVal, ok := ctx.Get("err")
	if !ok {
		t.Fatal("Expected err to be set")
	}

	if errVal != nil {
		t.Errorf("Expected err=nil, got %v", errVal)
	}
}

func TestGalaxyLocalsNilCheck(t *testing.T) {
	ctx := NewContext()
	ctx.SetLocals(map[string]any{
		"user": nil,
	})

	code := `
if Galaxy.Locals.user == nil {
	var result = "user is nil"
}
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	result, ok := ctx.Get("result")
	if !ok {
		t.Fatal("Expected result to be set (user should be nil)")
	}

	if result != "user is nil" {
		t.Errorf("Expected 'user is nil', got %v", result)
	}
}

func TestGalaxyRedirectPascalCase(t *testing.T) {
	ctx := NewContext()
	ctx.SetLocals(map[string]any{"user": nil})

	code := `
if Galaxy.Locals.user == nil {
	Galaxy.Redirect("/login", 302)
}
var shouldNotRun = "nope"
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !ctx.ShouldRedirect {
		t.Error("Expected redirect to be triggered")
	}

	if ctx.RedirectURL != "/login" {
		t.Errorf("Expected redirect to /login, got %s", ctx.RedirectURL)
	}

	_, ok := ctx.Get("shouldNotRun")
	if ok {
		t.Error("Code after redirect should not execute")
	}
}

func TestDashboardPattern(t *testing.T) {
	ctx := NewContext()
	ctx.SetLocals(map[string]any{"user": nil})

	code := `
if Galaxy.Locals.user == nil {
    Galaxy.Redirect("/login", 302)
}

userName := "User"
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !ctx.ShouldRedirect {
		t.Error("Expected redirect")
	}

	if ctx.RedirectURL != "/login" {
		t.Errorf("Expected /login, got %s", ctx.RedirectURL)
	}

	_, ok := ctx.Get("userName")
	if ok {
		t.Error("userName should not be set after redirect")
	}
}

func TestGalaxyLocalsFieldAccess(t *testing.T) {
	ctx := NewContext()
	
	// Test with actual user object
	ctx.SetLocals(map[string]any{
		"user": map[string]interface{}{"name": "John"},
		"db": "mock-db",
	})

	code := `
var user = Galaxy.Locals.user
var db = Galaxy.Locals.db
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	user, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user")
	}

	userMap, ok := user.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", user)
	}

	if userMap["name"] != "John" {
		t.Errorf("Expected John, got %v", userMap["name"])
	}
}

func TestGalaxyLocalsStepByStep(t *testing.T) {
	ctx := NewContext()
	ctx.SetLocals(map[string]any{
		"user": "test-user",
	})

	// Step 1: Access Galaxy
	code1 := `var g = Galaxy`
	if err := ctx.Execute(code1); err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	g, _ := ctx.Get("g")
	t.Logf("Galaxy type: %T, value: %+v", g, g)

	// Step 2: Access Galaxy.Locals
	code2 := `var locals = Galaxy.Locals`
	if err := ctx.Execute(code2); err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	locals, _ := ctx.Get("locals")
	t.Logf("Locals type: %T, value: %+v", locals, locals)

	// Step 3: Access Galaxy.Locals.user (should fail here)
	code3 := `var user = Galaxy.Locals.user`
	if err := ctx.Execute(code3); err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	user, _ := ctx.Get("user")
	t.Logf("User type: %T, value: %+v", user, user)
	
	if user != "test-user" {
		t.Errorf("Expected 'test-user', got %v", user)
	}
}

func TestShortAssignmentMethodCall(t *testing.T) {
	repo := &UserRepository{
		users: map[string]string{"1": "Alice"},
	}
	
	ctx := NewContext()
	ctx.Set("repo", repo)
	ctx.Set("id", "1")

	code := `user, err := repo.FindByID(id)`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	user, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user to be set")
	}

	if user != "Alice" {
		t.Errorf("Expected Alice, got %v", user)
	}
}

func TestParamsAndMethodCall(t *testing.T) {
	repo := &UserRepository{
		users: map[string]string{"123": "Bob"},
	}
	
	ctx := NewContext()
	ctx.SetParams(map[string]string{"id": "123"})
	ctx.Set("repo", repo)

	code := `
var idParam = id
user, err := repo.FindByID(idParam)
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	idParam, ok := ctx.Get("idParam")
	if !ok {
		t.Fatal("Expected idParam")
	}
	t.Logf("idParam: %v (%T)", idParam, idParam)

	user, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user")
	}
	t.Logf("user: %v (%T)", user, user)

	if user != "Bob" {
		t.Errorf("Expected Bob, got %v", user)
	}
}

func TestRealWorldPattern(t *testing.T) {
	repo := &UserRepository{
		users: map[string]string{"valid-id": "TestUser"},
	}
	
	ctx := NewContext()
	ctx.Set("repo", repo)
	ctx.Set("id", "valid-id")
	
	// Register uuid.Parse mock
	ctx.RegisterPackageFunc("uuid", "Parse", func(args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil // Return the ID as-is for test
		}
		return nil, fmt.Errorf("invalid")
	})

	code := `import "uuid"

projectUUID, err := uuid.Parse(id)
if err != nil {
	var failed = "parse failed"
}

user, err := repo.FindByID(projectUUID)
if err != nil {
	var failed2 = "find failed"
}
`

	execErr := ctx.Execute(code)
	if execErr != nil {
		t.Fatalf("Execute failed: %v", execErr)
	}

	projectUUID, ok := ctx.Get("projectUUID")
	if !ok {
		t.Fatal("Expected projectUUID")
	}
	t.Logf("projectUUID: %v (%T)", projectUUID, projectUUID)

	user, ok := ctx.Get("user")
	if !ok {
		t.Fatal("Expected user")
	}
	t.Logf("user: %v (%T)", user, user)

	if user != "TestUser" {
		t.Errorf("Expected TestUser, got %v", user)
	}
}

func TestMethodCallReturnsError(t *testing.T) {
	repo := &UserRepository{
		users: map[string]string{},
	}
	
	ctx := NewContext()
	ctx.Set("repo", repo)

	code := `user, err := repo.FindByID("missing")`

	execErr := ctx.Execute(code)
	if execErr != nil {
		t.Fatalf("Execute failed: %v", execErr)
	}

	user, _ := ctx.Get("user")
	t.Logf("user: %v (%T)", user, user)

	errVal, ok := ctx.Get("err")
	if !ok {
		t.Fatal("Expected err")
	}
	t.Logf("err: %v (%T)", errVal, errVal)

	if errVal == nil {
		t.Error("Expected error to be non-nil")
	}
}

func (r *UserRepository) FindNil(id string) (*UserRepository, error) {
	return nil, nil
}

func TestMethodCallReturnsNilValue(t *testing.T) {
	repo := &UserRepository{}
	
	ctx := NewContext()
	ctx.Set("repo", repo)

	code := `result, err := repo.FindNil("test")`

	execErr := ctx.Execute(code)
	if execErr != nil {
		t.Fatalf("Execute failed: %v", execErr)
	}

	result, ok := ctx.Get("result")
	if !ok {
		t.Fatal("Expected result")
	}
	t.Logf("result: %v (%T)", result, result)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}
