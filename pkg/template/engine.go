package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/galaxy/galaxy/pkg/executor"
)

type Engine struct {
	ctx *executor.Context
}

func NewEngine(ctx *executor.Context) *Engine {
	return &Engine{ctx: ctx}
}

var (
	expressionRegex = regexp.MustCompile(`\{([^}]+)\}`)
	attrRegex       = regexp.MustCompile(`(\w+)=\{([^}]+)\}|(\w+)="([^"]+)"|(\w+)='([^']+)'|(\w+)`)
)

type RenderOptions struct {
	Props map[string]interface{}
	Slots map[string]string
}

func (e *Engine) Render(template string, opts *RenderOptions) (string, error) {
	if opts != nil {
		for k, v := range opts.Props {
			e.ctx.SetProp(k, v)
			e.ctx.Set(k, v)
		}
		if opts.Slots != nil {
			e.ctx.Slots = opts.Slots
		}
	}

	result := template

	result = e.renderDirectives(result)
	result = e.renderSlots(result)
	result = e.renderExpressions(result)

	return result, nil
}

func (e *Engine) renderExpressions(template string) string {
	return expressionRegex.ReplaceAllStringFunc(template, func(match string) string {
		expr := strings.Trim(match, "{}")
		expr = strings.TrimSpace(expr)

		if val, ok := e.ctx.Get(expr); ok {
			return fmt.Sprintf("%v", val)
		}

		if val, ok := e.ctx.GetProp(expr); ok {
			return fmt.Sprintf("%v", val)
		}

		if strings.Contains(expr, ".") {
			result, ok := e.evaluateExpression(expr)
			if ok {
				return result
			}
		}

		return match
	})
}

func (e *Engine) renderSlots(template string) string {
	slotRegex := regexp.MustCompile(`<slot(?:\s+name="(\w+)")?\s*/>|<slot(?:\s+name="(\w+)")?>.*?</slot>`)

	return slotRegex.ReplaceAllStringFunc(template, func(match string) string {
		matches := slotRegex.FindStringSubmatch(match)

		var slotName string
		if matches[1] != "" {
			slotName = matches[1]
		} else if matches[2] != "" {
			slotName = matches[2]
		} else {
			slotName = "default"
		}

		if content, ok := e.ctx.Slots[slotName]; ok {
			return content
		}

		innerRegex := regexp.MustCompile(`<slot[^>]*>(.*?)</slot>`)
		if innerMatch := innerRegex.FindStringSubmatch(match); len(innerMatch) > 1 {
			return innerMatch[1]
		}

		return ""
	})
}

func (e *Engine) renderDirectives(template string) string {
	template = e.renderIfDirective(template)
	template = e.renderForDirective(template)
	return template
}

func (e *Engine) renderIfDirective(template string) string {
	ifRegex := regexp.MustCompile(`(?s)<(\w+)\s+galaxy:if=\{([^}]+)\}([^>]*)>(.*?)</(\w+)>`)

	return ifRegex.ReplaceAllStringFunc(template, func(match string) string {
		matches := ifRegex.FindStringSubmatch(match)
		if len(matches) < 6 {
			return match
		}

		tagName := matches[1]
		condition := strings.TrimSpace(matches[2])
		attrs := matches[3]
		content := matches[4]
		closingTag := matches[5]

		if tagName != closingTag {
			return match
		}

		if e.evaluateCondition(condition) {
			return fmt.Sprintf("<%s%s>%s</%s>", tagName, attrs, content, tagName)
		}

		return ""
	})
}

func (e *Engine) renderForDirective(template string) string {
	forRegex := regexp.MustCompile(`(?s)<(\w+)\s+galaxy:for=\{([^}]+)\}([^>]*)>(.*?)</(\w+)>`)

	return forRegex.ReplaceAllStringFunc(template, func(match string) string {
		matches := forRegex.FindStringSubmatch(match)
		if len(matches) < 6 {
			return match
		}

		tagName := matches[1]
		loopExpr := strings.TrimSpace(matches[2])
		attrs := matches[3]
		content := matches[4]
		closingTag := matches[5]

		if tagName != closingTag {
			return match
		}

		parts := strings.Fields(loopExpr)
		if len(parts) < 3 || parts[1] != "in" {
			return match
		}

		itemVar := parts[0]
		arrayName := parts[2]

		if val, ok := e.ctx.Get(arrayName); ok {
			if arr, ok := val.([]interface{}); ok {
				var result strings.Builder

				for _, item := range arr {
					oldVal, hadOld := e.ctx.Get(itemVar)
					e.ctx.Set(itemVar, item)

					rendered := e.renderExpressions(content)
					result.WriteString(fmt.Sprintf("<%s%s>%s</%s>", tagName, attrs, rendered, tagName))

					if hadOld {
						e.ctx.Set(itemVar, oldVal)
					}
				}

				return result.String()
			}
		}

		return match
	})
}

func (e *Engine) evaluateCondition(condition string) bool {
	if val, ok := e.ctx.Get(condition); ok {
		switch v := val.(type) {
		case bool:
			return v
		case int64:
			return v != 0
		case float64:
			return v != 0
		case string:
			return v != ""
		case []interface{}:
			return len(v) > 0
		default:
			return true
		}
	}

	return false
}

func ParseAttributes(attrString string) map[string]interface{} {
	attrs := make(map[string]interface{})

	matches := attrRegex.FindAllStringSubmatch(attrString, -1)
	for _, match := range matches {
		if match[1] != "" {
			attrs[match[1]] = match[2]
		} else if match[3] != "" {
			attrs[match[3]] = match[4]
		} else if match[5] != "" {
			attrs[match[5]] = match[6]
		} else if match[7] != "" {
			attrs[match[7]] = true
		}
	}

	return attrs
}

func (e *Engine) evaluateExpression(expr string) (string, bool) {
	parts := strings.Split(expr, ".")
	if len(parts) < 2 {
		return "", false
	}

	varName := parts[0]
	val, ok := e.ctx.Get(varName)
	if !ok {
		return "", false
	}

	methodCall := parts[1]
	if strings.HasSuffix(methodCall, "()") {
		methodName := strings.TrimSuffix(methodCall, "()")

		if reqCtx, ok := val.(interface {
			Path() string
			Method() string
			URL() string
		}); ok {
			switch methodName {
			case "Path":
				return reqCtx.Path(), true
			case "Method":
				return reqCtx.Method(), true
			case "URL":
				return reqCtx.URL(), true
			}
		}
	}

	return "", false
}
