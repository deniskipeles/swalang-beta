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
		Name: constants.IMPORTLIB_LOAD_MODULE_BUILTIN_NAME, // More descriptive name
		Fn: func(ctx object.ExecutionContext, args ...object.Object) object.Object {
			if loadModuleFunc == nil {
				// This indicates an initialization problem in the interpreter.
				return object.NewError(constants.InternalError, constants.IMPORTLIB_LOAD_MODULE_NOT_INIT_ERROR)
			}
			return loadModuleFunc(ctx, args...)
		},
	}
	env.Set(constants.IMPORTLIB_LOAD_MODULE_METHOD_NAME, loadBuiltin)

	module := &object.Module{
		Name: constants.IMPORTLIB_MODULE_NAME, // Pylearn module name
		Path: constants.SYS_MODULE_BUILTIN_PATH,
		Env:  env,
	}
	object.RegisterNativeModule(constants.IMPORTLIB_MODULE_NAME, module)
}