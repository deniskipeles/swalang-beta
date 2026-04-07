package pyos

import (
	"os" // Go's os package

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyGetEnvFn implements os.getenv(key, default=None)
// Accepts ExecutionContext (unused but required by signature)
func pyGetEnvFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var key string
	var defaultVal object.Object = object.NULL // Default is None

	// Argument parsing
	if len(args) < 1 || len(args) > 2 {
		// Use object.NewError
		return object.NewError(constants.TypeError, "os.getenv() takes 1 or 2 arguments (%d given)", len(args))
	}

	// Key argument
	keyObj, ok := args[0].(*object.String)
	if !ok {
		// Use object.NewError
		return object.NewError(constants.TypeError, "os.getenv() argument 1 must be str, not %s", args[0].Type())
	}
	key = keyObj.Value

	// Optional default argument
	if len(args) == 2 {
		defaultVal = args[1] // Can be any Pylearn object
	}

	// Use Go's os.LookupEnv
	value, found := os.LookupEnv(key)
	if !found {
		return defaultVal // Return the provided default (which might be None)
	}

	return &object.String{Value: value}
}

// GetEnv is the Pylearn Builtin object for os.getenv
// Signature now matches object.BuiltinFunction
var GetEnv = &object.Builtin{Fn: pyGetEnvFn}

// TODO: Implement os.environ, os.putenv, os.unsetenv