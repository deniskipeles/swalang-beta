// internal/stdlib/template/template_object.go
package template

import (
	"fmt"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

const TEMPLATE_OBJ object.ObjectType = "Template"

type Template struct {
	Name string
	Root Node
}

func (t *Template) Type() object.ObjectType { return TEMPLATE_OBJ }
func (t *Template) Inspect() string         { return fmt.Sprintf("<Template source='%s'>", t.Name) }

// NewTemplate compiles a template string into a Template object.
func NewTemplate(name, source string) (*Template, error) {
	l := NewLexer(source)
	tokens := l.Run()
	for _, tok := range tokens {
		if tok.Type == TokenError {
			return nil, fmt.Errorf("lexer error at line %d: %s", tok.Line, tok.Val)
		}
	}

	p := NewParser(tokens)
	root, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return &Template{Name: name, Root: root}, nil
}

// Render executes the template with a given context.
func (t *Template) Render(execCtx object.ExecutionContext, pylearnContext *object.Dict) object.Object {
	var out strings.Builder
	err := t.RenderToBuilder(execCtx, pylearnContext, &out)
	if err != nil {
		return object.NewError(constants.TemplateError, err.Error())
	}
	return &object.String{Value: out.String()}
}

// RenderToBuilder is a helper for rendering to an existing string builder (for extends/include)
func (t *Template) RenderToBuilder(execCtx object.ExecutionContext, pylearnContext *object.Dict, out *strings.Builder) error {
	templateContext := pylearnDictToTemplateContext(pylearnContext, execCtx)

	// Add built-in functions and filters to the context by looking them up from the global environment.
	if globalEnv := execCtx.GetCurrentEnvironment(); globalEnv != nil {
		if rangeFn, ok := globalEnv.Get("range"); ok {
			templateContext["range"] = NewValue(rangeFn, execCtx)
		}
		if upperFn, ok := globalEnv.Get("upper"); ok {
			// This assumes an 'upper' builtin exists. If not, we create one.
			templateContext["upper"] = NewValue(upperFn, execCtx)
		} else {
			// Provide a default 'upper' filter if not globally available.
			templateContext["upper"] = NewValue(&object.Builtin{
				Name: "upper",
				Fn: func(ctx object.ExecutionContext, args ...object.Object) object.Object {
					if len(args) == 1 {
						if str, ok := args[0].(*object.String); ok {
							return &object.String{Value: strings.ToUpper(str.Value)}
						}
					}
					return object.NewError("TypeError", "upper filter expects a string")
				},
			}, execCtx)
		}
	}

	renderer := NewRenderer(execCtx, templateContext, out)
	return renderer.Render(t.Root)
}

// GetObjectAttribute exposes the `render` method to Pylearn.
func (t *Template) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == "render" {
		return &object.Builtin{
			Name: "Template.render",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 1 {
					return object.NewError(constants.TypeError, "render() takes 1 argument (context_dict)")
				}
				context, ok := args[0].(*object.Dict)
				if !ok {
					return object.NewError(constants.TypeError, "render() argument must be a dict")
				}
				return t.Render(callCtx, context)
			},
		}, true
	}
	return nil, false
}

var _ object.Object = (*Template)(nil)
var _ object.AttributeGetter = (*Template)(nil)