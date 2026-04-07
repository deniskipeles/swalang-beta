package pyjson

import (
	"encoding/json" // Go's json package
	// "fmt"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyJsonDumpsFn implements json.dumps(obj, indent=None)
func pyJsonDumpsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, "json.dumps() takes 1 or 2 arguments (%d given)", len(args))
	}

	objToDump := args[0]
	var indent int = 0 // Default no indentation
	useIndent := false

	if len(args) == 2 {
		if args[1] != object.NULL { // Allow None for no indent
			indentObj, ok := args[1].(*object.Integer)
			if !ok {
				return object.NewError(constants.TypeError, "json.dumps() indent argument must be an integer or None, not %s", args[1].Type())
			}
			indentVal := int(indentObj.Value)
			if indentVal < 0 { // Python's json uses 0 for compact, negative is often treated as 0 or compact
				indent = 0
			} else {
				indent = indentVal
				useIndent = true
			}
		}
	}

	// Convert Pylearn object to Go interface{} structure
	goData, err := ConvertPylearnToInterface(objToDump, 0)
	if err != nil {
		// Error from conversion (e.g., recursion depth, non-string dict key, non-serializable type)
		return object.NewError(constants.JSONEncodeError,err.Error()) // convertPylearnToInterface already formats errors
	}

	// Marshal to JSON bytes
	var jsonBytes []byte
	if useIndent {
		// Create indent string (e.g., 2 spaces)
		indentPrefix := "" // For top level
		indentStr := ""
		for i := 0; i < indent; i++ {
			indentStr += " "
		}
		jsonBytes, err = json.MarshalIndent(goData, indentPrefix, indentStr)
	} else {
		jsonBytes, err = json.Marshal(goData)
	}

	if err != nil {
		// Error during Go's json.Marshal (e.g., unsupported Go type if conversion was flawed)
		return object.NewError("JSONMarshalError: %s", err.Error())
	}

	return &object.String{Value: string(jsonBytes)}
}

// Dumps is the Pylearn Builtin object for json.dumps
var Dumps = &object.Builtin{Fn: pyJsonDumpsFn}
