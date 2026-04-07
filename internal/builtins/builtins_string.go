package builtins

import (
	"fmt"
	"unicode/utf8" // For RuneError

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- chr() ---
// Accepts ExecutionContext (unused but required by signature)
func pyChrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsStringChrArgCountError, len(args))
	}
	arg := args[0]

	codePointObj, ok := arg.(*object.Integer)
	if !ok {
		// TODO: Check __index__ via context?
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsStringChrIntInterpretationError)
	}
	codePoint := codePointObj.Value

	// Check valid Unicode code point range
	if codePoint < 0 || codePoint > 0x10FFFF {
		// Use object.NewError
		return object.NewError(constants.ValueError, constants.BuiltinsStringChrRangeError)
	}

	// Check if it's a valid rune
	r := rune(codePoint)
	// This check is generally sufficient alongside the range check
	// The extra utf8.ValidRune check might be redundant unless dealing with surrogate pairs manually
	// if r == utf8.RuneError { ... internal check ... }

	return &object.String{Value: string(r)}
}

// --- ord() ---
// Accepts ExecutionContext (unused but required by signature)
func pyOrdFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsStringOrdArgCountError, len(args))
	}
	arg := args[0]

	// Python's ord() works on str, bytes, bytearray. Let's handle str and bytes.
	switch obj := arg.(type) {
	case *object.String:
		// Use RuneCountInString for correct length check with Unicode
		if utf8.RuneCountInString(obj.Value) != 1 {
			// Use object.NewError
			return object.NewError(constants.TypeError, constants.BuiltinsStringOrdCharLengthError, utf8.RuneCountInString(obj.Value))
		}
		// Decode the single rune
		r, _ := utf8.DecodeRuneInString(obj.Value)
		return &object.Integer{Value: int64(r)}

	case *object.Bytes:
		// ord() on bytes requires length 1
		if len(obj.Value) != 1 {
			// Use object.NewError
			return object.NewError(constants.TypeError, constants.BuiltinsStringOrdByteStringLengthError, len(obj.Value))
		}
		// Return the integer value of the single byte
		return &object.Integer{Value: int64(obj.Value[0])}

	default:
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsStringOrdArgTypeError, arg.Type())
	}
}

// --- format() ---
// Accepts ExecutionContext (needed for potential __format__ call and fallback str())
func pyFormatFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsStringFormatArgCountError, len(args))
	}
	value := args[0]
	formatSpec := constants.EmptyString   // Default format specifier is empty string
	var formatSpecObj object.Object = nil // Keep track of original spec object if present

	if len(args) == 2 {
		specObj, ok := args[1].(*object.String)
		if !ok {
			// Use object.NewError
			return object.NewError(constants.TypeError, constants.BuiltinsStringFormatArg2TypeError, args[1].Type())
		}
		formatSpec = specObj.Value
		formatSpecObj = specObj // Save the string object
	}
	if formatSpecObj == nil {
		formatSpecObj = &object.String{Value: constants.EmptyString} // Ensure we have an object for __format__ call
	}

	// --- Check for __format__ method ---
	// Use object.Instance for type check
	if inst, ok := value.(*object.Instance); ok && inst.Class != nil {
		if formatMethodObj, methodOk := inst.Class.Methods[constants.DunderFormat]; methodOk {
			if formatMethod, isFunc := formatMethodObj.(*object.Function); isFunc { // <<< TYPE ASSERTION
				boundFormat := &object.BoundMethod{Instance: inst, Method: formatMethod}
				result := object.ApplyBoundMethod(ctx, boundFormat, []object.Object{formatSpecObj}, object.NoToken)
				// Result must be a string
				if object.IsError(result) {
					return result
				} // Propagate error from __format__
				if _, isStr := result.(*object.String); !isStr {
					return object.NewError(constants.TypeError, constants.BuiltinsStringFormatStrReturnError, result.Type())
				}
				return result // Return the string result from __format__
			}
			// No __format__, fall through
		}
	}

	// --- Fallback for built-ins (very basic) ---
	switch v := value.(type) {
	case *object.Integer:
		// Basic integer formatting (ignores most specifiers)
		switch formatSpec {
		case constants.EmptyString, constants.FormatSpecifierD:
			return &object.String{Value: fmt.Sprintf(constants.FormatInt, v.Value)}
		case constants.FormatSpecifierX:
			return &object.String{Value: fmt.Sprintf(constants.FormatHex, v.Value)}
		case constants.FormatSpecifierO:
			return &object.String{Value: fmt.Sprintf(constants.FormatOct, v.Value)}
		case constants.FormatSpecifierB:
			return &object.String{Value: fmt.Sprintf(constants.FormatBin, v.Value)}
		// TODO: Add more specifiers: width, alignment, sign, #, 0, grouping, etc.
		default:
			// Python raises ValueError for unsupported format codes for type
			return object.NewError(constants.ValueError, constants.BuiltinsStringFormatUnknownIntCode, formatSpec)
		}
	case *object.Float:
		// Basic float formatting
		// Python's default float format is complex (~'g'). Let's default to 'f' maybe?
		// Or require explicit specifier? Let's require explicit or empty for now.
		switch formatSpec {
		case constants.EmptyString, constants.FormatSpecifierF: // Defaulting empty to 'f' is a simplification
			// TODO: Implement precision, 'g', 'e', '%' etc.
			return &object.String{Value: fmt.Sprintf(constants.FormatFloat, v.Value)} // Basic %f for now
		case constants.FormatSpecifierE:
			return &object.String{Value: fmt.Sprintf(constants.FormatScientific, v.Value)}
		default:
			return object.NewError(constants.ValueError, constants.BuiltinsStringFormatUnknownFloatCode, formatSpec)
		}
	case *object.String:
		// Basic string formatting (ignores most specifiers)
		// TODO: Implement alignment, width, etc.
		if formatSpec == constants.EmptyString || formatSpec == constants.FormatSpecifierS { // Allow 's' explicitly
			return v // Default is just the string itself
		}
		return object.NewError(constants.ValueError, constants.BuiltinsStringFormatUnknownStrCode, formatSpec)
		// TODO: Add formatting for other types (bool, None?)
	}

	// Default if no __format__ and no specific built-in handling for the spec
	// If formatSpec was non-empty, raise error here
	if formatSpec != constants.EmptyString {
		return object.NewError(constants.NotImplementedError, constants.BuiltinsStringFormatNotImplemented, value.Type(), formatSpec)
	}

	// Fallback: If formatSpec was empty and no __format__, call str() via context
	strBuiltin, ok := Builtins[constants.BuiltinsStrFuncName]
	if !ok {
		return object.NewError(constants.InternalError, constants.BuiltinsStringStrBuiltinNotFound)
	}
	strResult := ctx.Execute(strBuiltin, value)
	// str() should return String or Error
	if _, isStr := strResult.(*object.String); !isStr {
		if object.IsError(strResult) {
			return strResult
		} // Propagate error from str()
		return object.NewError(constants.InternalError, constants.BuiltinsStringStrFallbackError, strResult.Type())
	}
	return strResult
}

// --- Registration ---
// Ensure functions match the required signature
func init() {
	registerBuiltin(constants.BuiltinsChrFuncName, &object.Builtin{Fn: pyChrFn})
	registerBuiltin(constants.BuiltinsOrdFuncName, &object.Builtin{Fn: pyOrdFn})
	registerBuiltin(constants.BuiltinsFormatFuncName, &object.Builtin{Fn: pyFormatFn})
}
