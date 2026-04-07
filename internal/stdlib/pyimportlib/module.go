package pyimportlib

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

var loadModuleFunc object.BuiltinFunction

// SetLoadModuleFunc allows the interpreter package to inject the actual implementation.
// This function must be EXPORTED.
func SetLoadModuleFunc(fn object.BuiltinFunction) {
	loadModuleFunc = fn
}

func init() {
	env := object.NewEnvironment()

	loadBuiltin := &object.Builtin{
		Name: "pylearn_importlib.load_module_from_path", // More descriptive name
		Fn: func(ctx object.ExecutionContext, args ...object.Object) object.Object {
			if loadModuleFunc == nil {
				// This indicates an initialization problem in the interpreter.
				return object.NewError(constants.InternalError, "pylearn_importlib.load_module_from_path not properly initialized by the interpreter")
			}
			return loadModuleFunc(ctx, args...)
		},
	}
	env.Set("load_module_from_path", loadBuiltin)

	module := &object.Module{
		Name: "pylearn_importlib", // Pylearn module name
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("pylearn_importlib", module)
}