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

func findDirectiveElement(template string, directiveName string) (tag string, attrs string, content string, start int, end int, found bool) {
	directiveAttr := directiveName + "="

	idx := strings.Index(template, directiveAttr)
	if idx == -1 {
		return "", "", "", 0, 0, false
	}

	openStart := strings.LastIndex(template[:idx], "<")
	if openStart == -1 {
		return "", "", "", 0, 0, false
	}

	openEnd := strings.Index(template[idx:], ">")
	if openEnd == -1 {
		return "", "", "", 0, 0, false
	}
	openEnd += idx

	tagNameEnd := openStart + 1
	for tagNameEnd < len(template) && template[tagNameEnd] != ' ' && template[tagNameEnd] != '>' {
		tagNameEnd++
	}
	tag = template[openStart+1 : tagNameEnd]

	attrs = strings.TrimSpace(template[tagNameEnd:openEnd])

	depth := 1
	pos := openEnd + 1
	searchTag := "<" + tag
	closeTag := "</" + tag + ">"

	for pos < len(template) && depth > 0 {
		nextOpen := strings.Index(template[pos:], searchTag)
		nextClose := strings.Index(template[pos:], closeTag)

		if nextClose == -1 {
			return "", "", "", 0, 0, false
		}

		if nextOpen != -1 && nextOpen < nextClose {
			if pos+nextOpen+len(searchTag) < len(template) {
				nextChar := template[pos+nextOpen+len(searchTag)]
				if nextChar == ' ' || nextChar == '>' {
					depth++
					pos += nextOpen + len(searchTag)
					continue
				}
			}
			pos += nextOpen + 1
			continue
		}

		depth--
		if depth == 0 {
			content = template[openEnd+1 : pos+nextClose]
			end = pos + nextClose + len(closeTag)
			return tag, attrs, content, openStart, end, true
		}
		pos += nextClose + len(closeTag)
	}

	return "", "", "", 0, 0, false
}

func replaceAllDirectives(template string, directive string, replacer func(tag, attrs, content string) string) string {
	result := template
	offset := 0

	for {
		tag, attrs, content, start, end, found := findDirectiveElement(result[offset:], directive)
		if !found {
			break
		}

		start += offset
		end += offset

		replacement := replacer(tag, attrs, content)
		result = result[:start] + replacement + result[end:]
		offset = start + len(replacement)
	}

	return result
}

func (e *Engine) renderDirectives(template string) string {
	template = e.renderIfDirective(template)
	template = e.renderForDirective(template)
	return template
}

func (e *Engine) renderIfDirective(template string) string {
	return replaceAllDirectives(template, "galaxy:if", func(tag, attrs, content string) string {
		ifStart := strings.Index(attrs, "galaxy:if={")
		if ifStart == -1 {
			return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, content, tag)
		}

		ifStart += len("galaxy:if={")
		ifEnd := strings.Index(attrs[ifStart:], "}")
		if ifEnd == -1 {
			return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, content, tag)
		}

		condition := strings.TrimSpace(attrs[ifStart : ifStart+ifEnd])

		otherAttrs := strings.TrimSpace(
			attrs[:ifStart-len("galaxy:if={")] +
				attrs[ifStart+ifEnd+1:],
		)

		if e.evaluateCondition(condition) {
			if otherAttrs != "" {
				return fmt.Sprintf("<%s %s>%s</%s>", tag, otherAttrs, content, tag)
			} else {
				return fmt.Sprintf("<%s>%s</%s>", tag, content, tag)
			}
		}

		return ""
	})
}

func (e *Engine) renderForDirective(template string) string {
	return replaceAllDirectives(template, "galaxy:for", func(tag, attrs, content string) string {
		forStart := strings.Index(attrs, "galaxy:for={")
		if forStart == -1 {
			return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, content, tag)
		}

		forStart += len("galaxy:for={")
		forEnd := strings.Index(attrs[forStart:], "}")
		if forEnd == -1 {
			return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, content, tag)
		}

		loopExpr := strings.TrimSpace(attrs[forStart : forStart+forEnd])

		otherAttrs := strings.TrimSpace(
			attrs[:forStart-len("galaxy:for={")] +
				attrs[forStart+forEnd+1:],
		)

		parts := strings.Fields(loopExpr)
		if len(parts) < 3 || parts[1] != "in" {
			return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, content, tag)
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

					if otherAttrs != "" {
						result.WriteString(fmt.Sprintf("<%s %s>%s</%s>", tag, otherAttrs, rendered, tag))
					} else {
						result.WriteString(fmt.Sprintf("<%s>%s</%s>", tag, rendered, tag))
					}

					if hadOld {
						e.ctx.Set(itemVar, oldVal)
					}
				}

				return result.String()
			}
		}

		return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, content, tag)
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
	} else {
		if m, ok := val.(map[string]interface{}); ok {
			property := parts[1]
			if propVal, ok := m[property]; ok {
				return fmt.Sprintf("%v", propVal), true
			}
		}
	}

	return "", false
}
