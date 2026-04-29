package pyos

import (
	"os" // For os.Getwd

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyGetCWDFn implements os.getcwd()
// Accepts ExecutionContext (unused but required by signature)
func pyGetCWDFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 0 {
		// Use object.NewError
		return object.NewError(constants.TypeError, "os.getcwd() takes no arguments (%d given)", len(args))
	}

	cwd, err := os.Getwd()
	if err != nil {
		// Error getting current working directory
		// Use object.NewError and the existing getErrno helper
		return object.NewError(constants.OSError, "[Errno %d] %s", getErrno(err), err.Error())
	}

	return &object.String{Value: cwd}
}

// GetCWD is the Pylearn Builtin object for os.getcwd
// Signature now matches object.BuiltinFunction
var GetCWD = &object.Builtin{Fn: pyGetCWDFn}

// TODO: Implement os.chdir