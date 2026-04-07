package pyjson

import (
	"encoding/json"
	// "fmt"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyJsonLoadsFn implements json.loads(json_string)
func pyJsonLoadsFn1(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "json.loads() takes exactly 1 argument (%d given)", len(args))
	}

	jsonStrObj, ok := args[0].(*object.String)
	if !ok {
		// Python's json.loads also accepts bytes-like objects.
		// TODO: Add support for Pylearn Bytes object later.
		return object.NewError(constants.TypeError, "the JSON object must be str, bytes or bytearray, not %s", args[0].Type())
	}
	jsonString := jsonStrObj.Value

	// Unmarshal into a generic Go interface{}
	var goData interface{}
	err := json.Unmarshal([]byte(jsonString), &goData)
	if err != nil {
		// Error during Go's json.Unmarshal (e.g., invalid JSON syntax)
		// Python raises JSONDecodeError. We can wrap Go's error.
		return object.NewError(constants.JSONDecodeError, "JSONDecodeError: %s", err.Error())
	}

	// Convert the Go interface{} structure back to Pylearn objects
	pylearnObj, err := ConvertInterfaceToPylearn(goData, 0)
	if err != nil {
		// Error during conversion (e.g., recursion depth, unexpected Go type)
		return object.NewError(constants.JSONDecodeError, "JSONDecodeError: %s", err.Error())
	}

	return pylearnObj
}

// pyJsonLoadsFn implements json.loads(json_string_or_bytes)
func pyJsonLoadsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "json.loads() takes exactly 1 argument (%d given)", len(args))
	}

	var jsonBytes []byte
	// --- THIS IS THE FIX ---
	// Check if the argument is a String or Bytes object.
	switch arg := args[0].(type) {
	case *object.String:
		jsonBytes = []byte(arg.Value)
	case *object.Bytes:
		jsonBytes = arg.Value
	default:
		// If it's neither, raise a TypeError.
		return object.NewError(constants.TypeError, "the JSON object must be str or bytes, not %s", args[0].Type())
	}
	// --- END OF FIX ---

	// Unmarshal into a generic Go interface{}
	var goData interface{}
	err := json.Unmarshal(jsonBytes, &goData)
	if err != nil {
		// Error during Go's json.Unmarshal (e.g., invalid JSON syntax)
		// Python raises JSONDecodeError. We can wrap Go's error.
		return object.NewError(constants.JSONDecodeError, "JSONDecodeError: %s", err.Error())
	}

	// Convert the Go interface{} structure back to Pylearn objects
	pylearnObj, err := ConvertInterfaceToPylearn(goData, 0)
	if err != nil {
		// Error during conversion (e.g., recursion depth, unexpected Go type)
		return object.NewError(constants.JSONDecodeError, "JSONDecodeError: %s", err.Error())
	}

	return pylearnObj
}

// Loads is the Pylearn Builtin object for json.loads
var Loads = &object.Builtin{Fn: pyJsonLoadsFn}
