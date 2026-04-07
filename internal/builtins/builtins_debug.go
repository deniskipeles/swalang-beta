package builtins

import (
	"fmt"
	"os"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- help() ---
// Very basic placeholder. Real help() interacts with pydoc system.
// Accepts ExecutionContext (unused but required by signature)
func pyHelpFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsDebugHelpArgCountError)
	}
	if len(args) == 0 {
		// Enter interactive help mode (not feasible here)
		fmt.Println(constants.BuiltinsDebugHelpWelcomeMessage)
		fmt.Println(constants.BuiltinsDebugHelpNotImplemented)
		return object.NULL
	}

	obj := args[0]
	// TODO: Extract docstring (__doc__) or generate help text based on type/methods/etc.
	// TODO: Potentially use ctx to call getattr(obj, "__doc__") if implemented
	// Basic version: Print type and maybe Inspect() output
	fmt.Printf(constants.BuiltinsDebugHelpOnObjectFormat, obj.Type())
	fmt.Println(obj.Inspect()) // Very basic help

	// Look for __doc__ attribute? Requires attribute access support.
	// Example using context (if getattr builtin exists and works via ctx):
	getattrBuiltin, okGetAttr := Builtins[constants.BuiltinsGetattrFuncName]
	if okGetAttr {
		// Create a temporary String object for "__doc__" as the attribute name
		docNameObj := object.NewString(constants.DunderDoc)
		
		docAttr := ctx.Execute(getattrBuiltin, obj, docNameObj) // Pass docNameObj

		if object.IsError(docAttr) {
			// If getattr itself returned an error (e.g., AttributeError), report it.
			errObj := docAttr.(*object.Error)
			fmt.Fprintf(os.Stderr, constants.BuiltinsDebugHelpAttributeError, errObj.Message)
		} else if docAttr != object.NULL {
			strBuiltin, okStr := Builtins[constants.BuiltinsStrFuncName]
			if okStr {
				docStr := ctx.Execute(strBuiltin, docAttr)
				if !object.IsError(docStr) {
					// Assuming str returns String and doc is plain string content
					if s, isString := docStr.(*object.String); isString {
						fmt.Println(constants.BuiltinsDebugHelpDocumentationTitle, s.Value) 
					}
				}
			}
		}
	}

	return object.NULL
}

// --- Registration ---
func init() {
	// Signature now matches object.BuiltinFunction
	registerBuiltin(constants.BuiltinsHelpFuncName, &object.Builtin{Fn: pyHelpFn})
}