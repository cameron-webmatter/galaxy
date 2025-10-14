package executor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

type Context struct {
	Variables map[string]interface{}
	Props     map[string]interface{}
	Slots     map[string]string
	Request   interface{}
}

func NewContext() *Context {
	return &Context{
		Variables: make(map[string]interface{}),
		Props:     make(map[string]interface{}),
		Slots:     make(map[string]string),
		Request:   nil,
	}
}

func (c *Context) Execute(code string) error {
	fset := token.NewFileSet()

	wrappedCode := "package main\n" + code

	node, err := parser.ParseFile(fset, "", wrappedCode, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.VAR {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						if err := c.processVarSpec(valueSpec); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (c *Context) processVarSpec(spec *ast.ValueSpec) error {
	for i, name := range spec.Names {
		if i < len(spec.Values) {
			value, err := c.evalExpr(spec.Values[i])
			if err != nil {
				return err
			}
			c.Variables[name.Name] = value
		}
	}
	return nil
}

func (c *Context) evalExpr(expr ast.Expr) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return c.evalBasicLit(e)
	case *ast.Ident:
		if val, ok := c.Variables[e.Name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined variable: %s", e.Name)
	case *ast.BinaryExpr:
		return c.evalBinaryExpr(e)
	case *ast.UnaryExpr:
		return c.evalUnaryExpr(e)
	case *ast.CallExpr:
		return c.evalCallExpr(e)
	case *ast.CompositeLit:
		return c.evalCompositeLit(e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (c *Context) evalBasicLit(lit *ast.BasicLit) (interface{}, error) {
	switch lit.Kind {
	case token.INT:
		return strconv.ParseInt(lit.Value, 10, 64)
	case token.FLOAT:
		return strconv.ParseFloat(lit.Value, 64)
	case token.STRING:
		s := lit.Value
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
		return s, nil
	case token.CHAR:
		if len(lit.Value) >= 3 {
			return rune(lit.Value[1]), nil
		}
		return nil, fmt.Errorf("invalid char literal")
	default:
		return nil, fmt.Errorf("unsupported literal kind: %v", lit.Kind)
	}
}

func (c *Context) evalBinaryExpr(expr *ast.BinaryExpr) (interface{}, error) {
	left, err := c.evalExpr(expr.X)
	if err != nil {
		return nil, err
	}

	right, err := c.evalExpr(expr.Y)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case token.ADD:
		return c.add(left, right)
	case token.SUB:
		return c.sub(left, right)
	case token.MUL:
		return c.mul(left, right)
	case token.QUO:
		return c.div(left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operator: %v", expr.Op)
	}
}

func (c *Context) evalUnaryExpr(expr *ast.UnaryExpr) (interface{}, error) {
	x, err := c.evalExpr(expr.X)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case token.SUB:
		if v, ok := x.(int64); ok {
			return -v, nil
		}
		if v, ok := x.(float64); ok {
			return -v, nil
		}
	case token.NOT:
		if v, ok := x.(bool); ok {
			return !v, nil
		}
	}
	return nil, fmt.Errorf("unsupported unary operator: %v", expr.Op)
}

func (c *Context) evalCallExpr(expr *ast.CallExpr) (interface{}, error) {
	if ident, ok := expr.Fun.(*ast.Ident); ok {
		switch ident.Name {
		case "len":
			if len(expr.Args) != 1 {
				return nil, fmt.Errorf("len expects 1 argument")
			}
			arg, err := c.evalExpr(expr.Args[0])
			if err != nil {
				return nil, err
			}
			v := reflect.ValueOf(arg)
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array || v.Kind() == reflect.String {
				return int64(v.Len()), nil
			}
		}
	}
	return nil, fmt.Errorf("unsupported function call")
}

func (c *Context) evalCompositeLit(expr *ast.CompositeLit) (interface{}, error) {
	if _, ok := expr.Type.(*ast.ArrayType); ok {
		var result []interface{}
		for _, elt := range expr.Elts {
			val, err := c.evalExpr(elt)
			if err != nil {
				return nil, err
			}
			result = append(result, val)
		}
		return result, nil
	}
	return nil, fmt.Errorf("unsupported composite literal")
}

func (c *Context) add(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt + rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat + rFloat, nil
		}
	}
	if lStr, ok := left.(string); ok {
		if rStr, ok := right.(string); ok {
			return lStr + rStr, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for +")
}

func (c *Context) sub(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt - rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat - rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for -")
}

func (c *Context) mul(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt * rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat * rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for *")
}

func (c *Context) div(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			if rInt == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return lInt / rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			if rFloat == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return lFloat / rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for /")
}

func (c *Context) Get(name string) (interface{}, bool) {
	val, ok := c.Variables[name]
	return val, ok
}

func (c *Context) Set(name string, value interface{}) {
	c.Variables[name] = value
}

func (c *Context) SetProp(name string, value interface{}) {
	c.Props[name] = value
}

func (c *Context) SetRequest(req interface{}) {
	c.Request = req
	c.Variables["Request"] = req
}

func (c *Context) GetRequest() (interface{}, bool) {
	return c.Request, c.Request != nil
}

func (c *Context) GetProp(name string) (interface{}, bool) {
	val, ok := c.Props[name]
	return val, ok
}

func (c *Context) String() string {
	var sb strings.Builder
	sb.WriteString("Context Variables:\n")
	for k, v := range c.Variables {
		sb.WriteString(fmt.Sprintf("  %s: %v (%T)\n", k, v, v))
	}
	return sb.String()
}
