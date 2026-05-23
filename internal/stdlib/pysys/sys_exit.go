package pysys

import (
	"os"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyExitFn implements the sys.exit([code]) function
// Accepts ExecutionContext (unused but required by signature)
func pyExitFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	exitCode := 0 // Default exit code is 0

	if len(args) > 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.SYS_EXIT_ARG_COUNT_ERROR, len(args))
	}

	if len(args) == 1 {
		// Try to interpret the argument as an integer status code
		codeArg := args[0]
		if codeInt, ok := codeArg.(*object.Integer); ok {
			exitCode = int(codeInt.Value) // Convert PyLearn int to Go int
		} else if codeArg != object.NULL { // Allow sys.exit(None) -> sys.exit(0)
			// Python prints the argument to stderr and exits with 1 if it's not None or int
			// For simplicity, let's just raise a TypeError
			// Use object.NewError
			return object.NewError(constants.TypeError, constants.SYS_EXIT_ARG_TYPE_ERROR, codeArg.Type())
		}
	}

	// --- Perform the actual exit ---
	// NOTE: This terminates the entire interpreter process immediately.
	os.Exit(exitCode)

	// This part is technically unreachable but needed for Go's type checker
	return object.NULL
}

// Exit is the Pylearn Builtin object for sys.exit
// Signature now matches object.BuiltinFunction
var Exit = &object.Builtin{Fn: pyExitFn}