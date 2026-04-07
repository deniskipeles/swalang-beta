package builtins

import (
	"fmt"
	"unicode/utf8" // For len(string)

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

var Builtins = make(map[string]*object.Builtin)

func registerBuiltin(name string, builtin *object.Builtin) {
	if _, exists := Builtins[name]; exists {
		panic(constants.BuiltinsCommonDuplicateRegistration + name)
	}
	Builtins[name] = builtin
}

func pyLenFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsCollectionsLenArgCountError, len(args))
	}
	arg := args[0]

	// --- Primary: Check for __len__ method using GetObjectAttribute ---
	if getter, hasGetAttr := arg.(object.AttributeGetter); hasGetAttr {
		lenMethodObj, foundMethod := getter.GetObjectAttribute(ctx, constants.DunderLen)
		if foundMethod {
			if object.IsError(lenMethodObj) { // Error retrieving attribute itself
				return lenMethodObj
			}

			var result object.Object
			// The lenMethodObj is expected to be callable.
			// If it's a Builtin from GetObjectAttribute (like for Dict.__len__),
			// its Fn expects 'self' (which is 'arg') to be prepended by the caller.
			// However, our GetObjectAttribute for Dict already returns a closure
			// that captures 'self' (the dict instance d).
			// So, that closure expects only the script-provided args.
			// For __len__, there are no script-provided args.
			if methodBuiltin, isBuiltin := lenMethodObj.(*object.Builtin); isBuiltin {
				// The Builtin's Fn (e.g., the closure from Dict.GetObjectAttribute for __len__)
				// expects 'self' implicitly via its closure.
				// So, call it with no explicit script arguments.
				result = methodBuiltin.Fn(ctx /*, no script args */)
			} else if boundInstanceMethod, isBound := lenMethodObj.(*object.BoundMethod); isBound {
				// This case is for object.Instance having a Pylearn-defined __len__
				result = object.ApplyBoundMethod(ctx, boundInstanceMethod, []object.Object{}, object.NoToken)
			} else {
				return object.NewError(constants.TypeError, constants.BuiltinsCollectionsLenAttrNotCallable, lenMethodObj.Type())
			}

			if object.IsError(result) {
				return result
			}
			intResult, isInt := result.(*object.Integer)
			if !isInt {
				return object.NewError(constants.TypeError, constants.BuiltinsCollectionsLenShouldReturnInteger, result.Type())
			}
			if intResult.Value < 0 {
				return object.NewError(constants.ValueError, constants.BuiltinsCollectionsLenShouldReturnNonNegative)
			}
			return intResult
		}
	}
	// --- End __len__ attribute Check ---

	// --- Fallback: Direct length for built-in types if no __len__ attribute was found/called ---
	// This section becomes less critical if all relevant types implement __len__ via GetObjectAttribute.
	// However, it's a good fallback for types that don't (or for performance in some cases).
	switch val := arg.(type) {
	case *object.String:
		return &object.Integer{Value: int64(utf8.RuneCountInString(val.Value))}
	case *object.List:
		return &object.Integer{Value: int64(len(val.Elements))}
	case *object.Tuple:
		return &object.Integer{Value: int64(len(val.Elements))}
	case *object.Dict: // Fallback if Dict.__len__ wasn't used
		return &object.Integer{Value: int64(len(val.Pairs))}
	case *object.Set:
		return &object.Integer{Value: int64(len(val.Elements))}
	case *object.Bytes:
		return &object.Integer{Value: int64(len(val.Value))}
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsCollectionsObjectHasNoLen, arg.Type())
	}
}

// ... (pyRangeFn and init()) ...
func getInt64Arg(funcName string, argIndex int, obj object.Object) (int64, error) {
	intObj, ok := obj.(*object.Integer)
	if !ok {
		return 0, fmt.Errorf(constants.BuiltinsCollectionsRangeArgTypeError, obj.Type(), argIndex+1, funcName)
	}
	return intObj.Value, nil
}

func pyRangeFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var start, stop, step int64 = 0, 0, 1
	var err error

	switch len(args) {
	case 1:
		stop, err = getInt64Arg(constants.BuiltinsRangeFuncName, 0, args[0])
		if err != nil {
			return object.NewError(constants.TypeError, err.Error())
		}
	case 2:
		start, err = getInt64Arg(constants.BuiltinsRangeFuncName, 0, args[0])
		if err != nil {
			return object.NewError(constants.TypeError, err.Error())
		}
		stop, err = getInt64Arg(constants.BuiltinsRangeFuncName, 1, args[1])
		if err != nil {
			return object.NewError(constants.TypeError, err.Error())
		}
	case 3:
		start, err = getInt64Arg(constants.BuiltinsRangeFuncName, 0, args[0])
		if err != nil {
			return object.NewError(constants.TypeError, err.Error())
		}
		stop, err = getInt64Arg(constants.BuiltinsRangeFuncName, 1, args[1])
		if err != nil {
			return object.NewError(constants.TypeError, err.Error())
		}
		step, err = getInt64Arg(constants.BuiltinsRangeFuncName, 2, args[2])
		if err != nil {
			return object.NewError(constants.TypeError, err.Error())
		}
		if step == 0 {
			return object.NewError(constants.ValueError, constants.BuiltinsCollectionsRangeStepZeroError)
		}
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsCollectionsRangeArgCountError, len(args))
	}
	return &object.Range{Start: start, Stop: stop, Step: step}
}

func init() {
	registerBuiltin(constants.BuiltinsLenFuncName, &object.Builtin{Fn: pyLenFn})
	registerBuiltin(constants.BuiltinsRangeFuncName, &object.Builtin{Fn: pyRangeFn})
}