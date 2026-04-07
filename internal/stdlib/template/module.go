// internal/stdlib/template/module.go
package template

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)


// Define and initialize the TemplateErrorClass here.
var TemplateErrorClass *object.Class

func initializeErrors() {
	if TemplateErrorClass == nil {
		TemplateErrorClass = object.CreateExceptionClass("TemplateError", object.ExceptionClass)
		object.BuiltinExceptionClasses["TemplateError"] = TemplateErrorClass
	}
}

// pyTemplateFromString implements template.from_string(template_string)
func pyTemplateFromString(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "from_string() takes exactly 1 argument (template_string)")
	}
	templateStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "template source must be a string")
	}

	tpl, err := NewTemplate("string", templateStr.Value)
	if err != nil {
		return object.NewError("TemplateError", err.Error())
	}
	return tpl
}

// pyTemplateRender implements template.render(template_string, context_dict)
func pyTemplateRender(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, "render() takes exactly 2 arguments (template_string, context)")
	}
	templateStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "template source must be a string")
	}
	contextDict, ok := args[1].(*object.Dict)
	if !ok {
		return object.NewError(constants.TypeError, "context must be a dict")
	}

	tpl, err := NewTemplate("string", templateStr.Value)
	if err != nil {
		return object.NewError("TemplateError", err.Error())
	}

	// Render the template
	return tpl.Render(ctx, contextDict)
}

func init() {
	// <<< FIX: Call the initializer function >>>
	initializeErrors()

	env := object.NewEnvironment()

	env.Set("render", &object.Builtin{Name: "template.render", Fn: pyTemplateRender})
	env.Set("from_string", &object.Builtin{Name: "template.from_string", Fn: pyTemplateFromString})
	env.Set("TemplateError", TemplateErrorClass)

	module := &object.Module{
		Name: "template",
		Path: "<builtin_template>",
		Env:  env,
	}
	object.RegisterNativeModule("template", module)
}