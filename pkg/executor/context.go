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
	Variables      map[string]interface{}
	Props          map[string]interface{}
	Slots          map[string]string
	Request        interface{}
	Locals         map[string]any
	RedirectURL    string
	RedirectStatus int
	ShouldRedirect bool
}

type GalaxyAPI struct {
	ctx *Context
}

func (g *GalaxyAPI) Redirect(url string, status int) {
	g.ctx.RedirectURL = url
	g.ctx.RedirectStatus = status
	g.ctx.ShouldRedirect = true
}

func NewContext() *Context {
	ctx := &Context{
		Variables: make(map[string]interface{}),
		Props:     make(map[string]interface{}),
		Slots:     make(map[string]string),
		Request:   nil,
		Locals:    make(map[string]any),
	}
	ctx.Variables["Galaxy"] = &GalaxyAPI{ctx: ctx}
	return ctx
}

func (c *Context) Execute(code string) error {
	fset := token.NewFileSet()

	wrappedCode := "package main\nfunc init() {\n" + code + "\n}"

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
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "init" && funcDecl.Body != nil {
				for _, stmt := range funcDecl.Body.List {
					if err := c.executeStmt(stmt); err != nil {
						return err
					}
					if c.ShouldRedirect {
						return nil
					}
				}
			}
		}
	}

	return nil
}

func (c *Context) executeStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		return c.executeIfStmt(s)
	case *ast.ExprStmt:
		_, err := c.evalExpr(s.X)
		return err
	case *ast.AssignStmt:
		return c.executeAssignStmt(s)
	case *ast.DeclStmt:
		if genDecl, ok := s.Decl.(*ast.GenDecl); ok {
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
		return nil
	default:
		return nil
	}
}

func (c *Context) executeIfStmt(stmt *ast.IfStmt) error {
	cond, err := c.evalExpr(stmt.Cond)
	if err != nil {
		return err
	}

	condBool, ok := cond.(bool)
	if !ok {
		return fmt.Errorf("if condition must be boolean")
	}

	if condBool {
		for _, s := range stmt.Body.List {
			if err := c.executeStmt(s); err != nil {
				return err
			}
			if c.ShouldRedirect {
				return nil
			}
		}
	} else if stmt.Else != nil {
		switch elseStmt := stmt.Else.(type) {
		case *ast.BlockStmt:
			for _, s := range elseStmt.List {
				if err := c.executeStmt(s); err != nil {
					return err
				}
				if c.ShouldRedirect {
					return nil
				}
			}
		case *ast.IfStmt:
			return c.executeIfStmt(elseStmt)
		}
	}

	return nil
}

func (c *Context) executeAssignStmt(stmt *ast.AssignStmt) error {
	if stmt.Tok != token.DEFINE && stmt.Tok != token.ASSIGN {
		return fmt.Errorf("unsupported assignment operator: %v", stmt.Tok)
	}

	for i, lhs := range stmt.Lhs {
		if i >= len(stmt.Rhs) {
			break
		}

		val, err := c.evalExpr(stmt.Rhs[i])
		if err != nil {
			return err
		}

		if ident, ok := lhs.(*ast.Ident); ok {
			c.Variables[ident.Name] = val
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
		if e.Name == "nil" {
			return nil, nil
		}
		if e.Name == "true" {
			return true, nil
		}
		if e.Name == "false" {
			return false, nil
		}
		if val, ok := c.Variables[e.Name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined variable: %s", e.Name)
	case *ast.SelectorExpr:
		return c.evalSelectorExpr(e)
	case *ast.BinaryExpr:
		return c.evalBinaryExpr(e)
	case *ast.UnaryExpr:
		return c.evalUnaryExpr(e)
	case *ast.CallExpr:
		return c.evalCallExpr(e)
	case *ast.CompositeLit:
		return c.evalCompositeLit(e)
	case *ast.ParenExpr:
		return c.evalExpr(e.X)
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
	case token.EQL:
		return c.equal(left, right), nil
	case token.NEQ:
		return !c.equal(left, right), nil
	case token.LSS:
		return c.less(left, right)
	case token.LEQ:
		return c.lessEqual(left, right)
	case token.GTR:
		return c.greater(left, right)
	case token.GEQ:
		return c.greaterEqual(left, right)
	case token.LAND:
		return c.logicalAnd(left, right), nil
	case token.LOR:
		return c.logicalOr(left, right), nil
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

func (c *Context) evalSelectorExpr(expr *ast.SelectorExpr) (interface{}, error) {
	x, err := c.evalExpr(expr.X)
	if err != nil {
		return nil, err
	}

	if m, ok := x.(map[string]interface{}); ok {
		return m[expr.Sel.Name], nil
	}

	if m, ok := x.(map[string]any); ok {
		return m[expr.Sel.Name], nil
	}

	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Struct {
		field := v.FieldByName(expr.Sel.Name)
		if field.IsValid() {
			return field.Interface(), nil
		}
	}

	return nil, fmt.Errorf("cannot select field %s from type %T", expr.Sel.Name, x)
}

func (c *Context) evalCallExpr(expr *ast.CallExpr) (interface{}, error) {
	if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "Galaxy" {
			if sel.Sel.Name == "redirect" {
				return c.handleGalaxyRedirect(expr.Args)
			}
		}
	}

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

func (c *Context) handleGalaxyRedirect(args []ast.Expr) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("Galaxy.redirect expects 2 arguments (url, status)")
	}

	url, err := c.evalExpr(args[0])
	if err != nil {
		return nil, err
	}
	urlStr, ok := url.(string)
	if !ok {
		return nil, fmt.Errorf("redirect URL must be string")
	}

	status, err := c.evalExpr(args[1])
	if err != nil {
		return nil, err
	}
	statusInt, ok := status.(int64)
	if !ok {
		return nil, fmt.Errorf("redirect status must be int")
	}

	c.RedirectURL = urlStr
	c.RedirectStatus = int(statusInt)
	c.ShouldRedirect = true

	return nil, nil
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

	if _, ok := expr.Type.(*ast.MapType); ok {
		result := make(map[string]interface{})
		for _, elt := range expr.Elts {
			kvExpr, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				return nil, fmt.Errorf("expected key-value pair in map literal")
			}

			key, err := c.evalExpr(kvExpr.Key)
			if err != nil {
				return nil, err
			}

			value, err := c.evalExpr(kvExpr.Value)
			if err != nil {
				return nil, err
			}

			keyStr, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("map keys must be strings")
			}

			result[keyStr] = value
		}
		return result, nil
	}

	if expr.Type == nil && len(expr.Elts) > 0 {
		if _, ok := expr.Elts[0].(*ast.KeyValueExpr); ok {
			result := make(map[string]interface{})
			for _, elt := range expr.Elts {
				kvExpr, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					return nil, fmt.Errorf("expected key-value pair in map literal")
				}

				key, err := c.evalExpr(kvExpr.Key)
				if err != nil {
					return nil, err
				}

				value, err := c.evalExpr(kvExpr.Value)
				if err != nil {
					return nil, err
				}

				keyStr, ok := key.(string)
				if !ok {
					return nil, fmt.Errorf("map keys must be strings")
				}

				result[keyStr] = value
			}
			return result, nil
		}
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

func (c *Context) equal(left, right interface{}) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return reflect.DeepEqual(left, right)
}

func (c *Context) less(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt < rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat < rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for <")
}

func (c *Context) lessEqual(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt <= rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat <= rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for <=")
}

func (c *Context) greater(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt > rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat > rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for >")
}

func (c *Context) greaterEqual(left, right interface{}) (interface{}, error) {
	if lInt, ok := left.(int64); ok {
		if rInt, ok := right.(int64); ok {
			return lInt >= rInt, nil
		}
	}
	if lFloat, ok := left.(float64); ok {
		if rFloat, ok := right.(float64); ok {
			return lFloat >= rFloat, nil
		}
	}
	return nil, fmt.Errorf("invalid operands for >=")
}

func (c *Context) logicalAnd(left, right interface{}) bool {
	lBool, lOk := left.(bool)
	rBool, rOk := right.(bool)
	if lOk && rOk {
		return lBool && rBool
	}
	return false
}

func (c *Context) logicalOr(left, right interface{}) bool {
	lBool, lOk := left.(bool)
	rBool, rOk := right.(bool)
	if lOk && rOk {
		return lBool || rBool
	}
	return false
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

func (c *Context) SetLocals(locals map[string]any) {
	c.Locals = locals
	c.Variables["Locals"] = locals
}

func (c *Context) GetLocals() map[string]any {
	return c.Locals
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
