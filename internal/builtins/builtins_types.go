package builtins

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer" // For token info
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- type ---
// Accepts ExecutionContext (unused but required)
func pyTypeFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesTypeArgCountError, len(args))
	}
	obj := args[0]

	// --- Return the object's actual type object ---
	// We need a way to get the canonical type object associated with a value.
	// This requires either:
	// 1. Storing a *Type object pointer on each object.Object.
	// 2. Having global type objects (e.g., IntType, StrType) and checking the type.

	// Simple approach for now (matching previous behavior partly):
	// - Return Class object for Instances/Classes
	// - Return descriptive string for built-ins (mimicking <class 'int'>)

	// Handle Instances specifically (assuming object.Instance)
	if inst, ok := obj.(*object.Instance); ok && inst.Class != nil {
		return inst.Class // Return the object.Class
	}
	// Handle Classes specifically
	if classObj, ok := obj.(*object.Class); ok {
		// TODO: Need a 'TypeType' object to return here to match Python's <class 'type'>
		// For now, return the class itself or a string representation
		return classObj // Or maybe return object.NewError("type() for classes not fully implemented")
	}
	// Handle VM Types if they are distinct and passed here
	// if vmInst, ok := obj.(*vm.Instance); ok && vmInst.Class != nil { return vmInst.InstanceClass }
	// if vmClass, ok := obj.(*vm.Class); ok { return vmClass }

	// For other built-in types, return a string representation <class 'name'>
	// Ideally, these would return actual Type objects later.
	var typeName string
	switch obj.Type() { // Use obj.Type() which should be defined for all object.Object
	case object.INTEGER_OBJ:
		typeName = constants.BuiltinsIntFuncName
	case object.FLOAT_OBJ:
		typeName = constants.BuiltinsFloatFuncName
	case object.STRING_OBJ:
		typeName = constants.BuiltinsStrFuncName
	case object.BOOLEAN_OBJ:
		typeName = constants.BuiltinsBoolFuncName
	case object.LIST_OBJ:
		typeName = constants.BuiltinsListFuncName
	case object.DICT_OBJ:
		typeName = constants.BuiltinsDictFuncName
	case object.NULL_OBJ:
		typeName = constants.BuiltinsNoneType
	case object.FUNCTION_OBJ:
		typeName = constants.BuiltinsFunctionType // Interpreter function
	case object.BUILTIN_OBJ:
		typeName = constants.BuiltinsBuiltinFunctionOrMethod
	case object.RANGE_OBJ:
		typeName = constants.BuiltinsRangeFuncName
	case object.MODULE_OBJ:
		typeName = constants.BuiltinsModuleType
	case object.TUPLE_OBJ:
		typeName = constants.BuiltinsTupleFuncName
	case object.SET_OBJ:
		typeName = constants.BuiltinsSetFuncName
	case object.BYTES_OBJ:
		typeName = constants.BuiltinsBytesFuncName
	case object.FILE_OBJ:
		typeName = constants.BuiltinsTextIOWrapper // Mimic _io type name
	case object.ITERATOR_OBJ:
		typeName = constants.BuiltinsIteratorType // Generic internal type name
	case object.STOP_ITER_OBJ:
		typeName = constants.StopIteration // Exception type name
	case object.BOUND_METHOD_OBJ:
		typeName = constants.BuiltinsMethodType // Interpreter bound method
	// Add VM types if they have distinct Type() values
	// case object.VM_CLOSURE_OBJ: typeName = "function" // VM closure
	// case object.VM_CLASS_OBJ: typeName = "type" // VM class object
	// case object.VM_BOUND_METHOD_OBJ: typeName = "method" // VM bound method
	default:
		// Fallback for unknown types
		typeName = string(obj.Type()) // Use the raw type string
	}
	// Mimic <class 'int'> format - return this string for now
	return &object.String{Value: fmt.Sprintf(constants.BuiltinsClassFormat, typeName)}
}

// --- int ---
// Accepts ExecutionContext (needed for potential __int__ call)
func pyIntFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	base := 10 // Default base for string conversion
	var arg object.Object

	// Argument Parsing (handle optional base)
	switch len(args) {
	case 1:
		arg = args[0]
	case 2:
		arg = args[0]
		baseArg, ok := args[1].(*object.Integer)
		if !ok {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntBaseRangeError)
		}
		base = int(baseArg.Value)
		// Python's base rules: 0 means auto-detect (like ParseInt), valid range 2-36
		if base != 0 && (base < 2 || base > 36) {
			return object.NewError(constants.ValueError, constants.BuiltinsTypesIntBaseRangeError)
		}
		// If base is specified, arg must be string/bytes/bytearray
		// TODO: Add bytes/bytearray support later
		if _, isStr := arg.(*object.String); !isStr {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		}

	default:
		return object.NewError(constants.TypeError, constants.BuiltinsMathIntPowerResultTooLarge, len(args))
	}

	// --- Check for __int__ or __index__ or __trunc__ ---
	// TODO: Implement dunder method calls using ctx in this order: __int__, __index__, __trunc__
	// if inst, ok := arg.(*object.Instance); ok { ... check methods ... }

	// --- Default conversion logic ---
	switch obj := arg.(type) {
	case *object.Integer:
		if base != 10 && base != 0 {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		} // Base only applies to strings
		return obj // Return integer itself
	case *object.Float:
		if base != 10 && base != 0 {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		}
		// Python truncates floats towards zero
		return &object.Integer{Value: int64(obj.Value)}
	case *object.Boolean:
		if base != 10 && base != 0 {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		}
		if obj.Value {
			return &object.Integer{Value: 1}
		}
		return &object.Integer{Value: 0}
	case *object.String:
		// Use strconv.ParseInt with the determined base
		// Trim leading/trailing whitespace first
		strVal := strings.TrimSpace(obj.Value)
		// Handle optional prefixes 0b, 0o, 0x if base is 0 or matches
		cleanedVal, detectedBase, errBase := handleIntPrefix(strVal, base)
		if errBase != nil {
			return object.NewError(constants.ValueError, errBase.Error())
		}

		val, err := strconv.ParseInt(cleanedVal, detectedBase, 64)
		if err != nil {
			// Check for more specific errors if possible (e.g., invalid digit)
			// Use object.NewError
			return object.NewError(constants.ValueError, constants.BuiltinsTypesIntInvalidLiteral, detectedBase, obj.Inspect())
		}
		return &object.Integer{Value: val}
		// TODO: Add bytes/bytearray support
	default:
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsTypesIntArgTypeError, obj.Type())
	}
}

// Helper for int() string parsing base/prefix
func handleIntPrefix(s string, requestedBase int) (string, int, error) {
	val := s
	base := requestedBase

	if strings.HasPrefix(val, constants.PlusSign) || strings.HasPrefix(val, constants.MinusSign) {
		// Sign is handled by ParseInt, remove for prefix check
		val = val[1:]
	}

	prefix := constants.EmptyString
	detectedBase := 0 // Base inferred from prefix

	if strings.HasPrefix(val, constants.BinPrefixLower) || strings.HasPrefix(val, constants.BinPrefixUpper) {
		prefix = val[:2]
		detectedBase = 2
	} else if strings.HasPrefix(val, constants.OctPrefixLower) || strings.HasPrefix(val, constants.OctPrefixUpper) {
		prefix = val[:2]
		detectedBase = 8
	} else if strings.HasPrefix(val, constants.HexPrefixLower) || strings.HasPrefix(val, constants.HexPrefixUpper) {
		prefix = val[:2]
		detectedBase = 16
	}

	if prefix != constants.EmptyString {
		if base == 0 {
			base = detectedBase // Auto-detect base
		} else if base != detectedBase {
			// Mismatched base and prefix (e.g., int("0b10", 10))
			return constants.EmptyString, 0, fmt.Errorf(constants.BuiltinsTypesIntInvalidIntLiteral, base, s)
		}
		// If bases match or auto-detected, remove prefix for ParseInt
		s = strings.Replace(s, prefix, constants.EmptyString, 1) // Remove first occurrence
		// Handle sign potentially removed earlier
		if strings.HasPrefix(s, constants.PlusSign) || strings.HasPrefix(s, constants.MinusSign) {
			// Okay
		} else if strings.HasPrefix(s, constants.Space) { // Check after prefix removal
			return constants.EmptyString, 0, fmt.Errorf(constants.BuiltinsTypesIntInvalidIntLiteral, base, s)
		}

	} else {
		// No prefix found
		if base == 0 {
			base = 10
		} // Default to base 10 if no prefix and base 0
		// If requestedBase was non-zero and != 10, use it.
		// Base is already set correctly here.
	}

	// Return original string (potentially with sign), and determined base
	return s, base, nil
}

// --- str ---
// Accepts ExecutionContext (needed for __str__/__repr__ calls)
// Note: Removed internal callStrBuiltin helper, logic moved here.
// CORRECTED: This version properly uses the MRO-aware attribute lookup.
func pyStrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesStrArgCountError, len(args))
	}
	arg := args[0]

	// Base case: str("a string") is "a string".
	if strObj, ok := arg.(*object.String); ok {
		return strObj
	}

	// Try to find and call __str__ using the proper MRO search.
	// object.CallGetAttr correctly finds the attribute on a parent class if needed.
	strMethod, strFound := object.CallGetAttr(ctx, arg, constants.DunderStr, object.NoToken)
	if strFound && !object.IsError(strMethod) && object.IsCallable(strMethod) {
		// ctx.Execute will correctly call the bound method (Pylearn or native Go).
		result := ctx.Execute(strMethod)
		if object.IsError(result) {
			return result // Propagate error from __str__
		}
		if _, isStr := result.(*object.String); !isStr {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesStrStrNonStringError, result.Type())
		}
		return result
	}

	// Fallback: If __str__ is not found, call repr(). This is Python's behavior.
	return pyReprFn(ctx, arg)
}

// --- repr() ---
// Accepts ExecutionContext (needed for __repr__)
// CORRECTED: This version also properly uses the MRO-aware attribute lookup.
func pyReprFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesReprArgCountError, len(args))
	}
	arg := args[0]

	// Try to find and call __repr__ using the proper MRO search.
	reprMethod, reprFound := object.CallGetAttr(ctx, arg, constants.DunderRepr, object.NoToken)
	if reprFound && !object.IsError(reprMethod) && object.IsCallable(reprMethod) {
		result := ctx.Execute(reprMethod)
		if object.IsError(result) {
			return result // Propagate error from __repr__
		}
		if _, isStr := result.(*object.String); !isStr {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesStrReprNonStringError, result.Type())
		}
		return result
	}

	// Final fallback for all objects if __repr__ is missing: use the internal Inspect().
	return &object.String{Value: arg.Inspect()}
}

// --- bool() ---
// Accepts ExecutionContext (needed for IsTruthy)
func pyBoolFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesBoolArgCountError, len(args))
	}
	if len(args) == 0 {
		return object.FALSE // bool() -> False
	}
	// Use the context-aware IsTruthy helper from object package
	truthy, err := object.IsTruthy(ctx, args[0])
	if err != nil {
		// --- FIX HERE ---
		// Propagate error from IsTruthy. Need to ensure it's an object.Object.
		// Check if the returned error already implements object.Object
		if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
			return pyErr // It was already a Pylearn error object (*object.Error)
		}
		// If it was a standard Go error wrapped by IsTruthy's NewError, re-wrap it
		// Or just wrap any non-Object error
		return object.NewError(constants.RuntimeError, constants.BuiltinsTypesBoolPropagatedFromIsTruthy, err)
		// --- END FIX ---
	}
	// Use NativeBoolToBooleanObject helper from object package
	return object.NativeBoolToBooleanObject(truthy)
}

// --- list() ---
// Accepts ExecutionContext (needed for iteration)
func pyListFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	token := lexer.Token{Line: -1, Column: 0, Type: lexer.ILLEGAL, Literal: constants.BuiltinsListFuncName} // Placeholder token
	if len(args) > 1 {
		return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsTypesListArgCountError, len(args))
	}

	elements := []object.Object{}

	if len(args) == 0 {
		return &object.List{Elements: elements} // list() -> []
	}

	iterableArg := args[0]

	// Get iterator using the context-aware helper from object package
	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, token)
	if errObj != nil {
		// Propagate TypeError if not iterable (errObj is already object.Error)
		return errObj
	}

	// Consume the iterator
	for {
		item, stop := iterator.Next()
		if stop {
			break // Iterator finished
		}
		if object.IsError(item) { // Check for errors yielded by iterator
			return item
		}
		elements = append(elements, item)
	}

	return &object.List{Elements: elements}
}

// --- dict() ---
// Accepts ExecutionContext (may be needed for future iteration/update)
// TODO: Implement dict(iterable_of_pairs), dict(**kwargs), dict(mapping)
func pyDictFn(ctx object.ExecutionContext, args ...object.Object /*, kwargs map[string]object.Object */) object.Object {
	// Python's dict() is very flexible. This is a minimal start.
	if len(args) > 0 {
		// TODO: Handle dict(mapping) or dict(iterable of pairs) using GetObjectIterator(ctx,...)
		return object.NewError(constants.TypeError, constants.BuiltinsTypesDictArgCountError)
	}
	// TODO: Handle kwargs

	// dict() called with no arguments
	return &object.Dict{Pairs: make(map[object.HashKey]object.DictPair)}
}




// --- ascii() ---
// Accepts ExecutionContext (needed for calling repr())
func pyAsciiFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesAsciiArgCountError, len(args))
	}

	// Get the repr() first using the context-aware repr function
	reprObj := pyReprFn(ctx, args...) // Pass context and original args
	if object.IsError(reprObj) {
		return reprObj // Propagate error from repr()
	}
	// reprObj is guaranteed to be String or Error (handled above)
	reprStr := reprObj.(*object.String).Value

	// Escape non-ASCII characters
	var builder strings.Builder
	for _, r := range reprStr {
		if r < 128 { // Is ASCII
			builder.WriteRune(r)
		} else if r <= 0xff {
			fmt.Fprintf(&builder, constants.HexEscapeFormat, r)
		} else if r <= 0xffff {
			fmt.Fprintf(&builder, constants.UnicodeEscapeFormat, r)
		} else { // Larger than 16 bits
			fmt.Fprintf(&builder, constants.UniversalCharacterNameEscapeFormat, r)
		}
	}
	return &object.String{Value: builder.String()}
}

// --- PLACEHOLDERS for other type conversions ---
// Update signatures to accept context

func pyTupleFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesTupleArgCountError, len(args))
	}

	elements := []object.Object{}

	if len(args) == 0 { // tuple() -> ()
		return &object.Tuple{Elements: elements}
	}

	iterableArg := args[0]
	// Use NoToken for errors from GetObjectIterator itself
	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		return errObj // Propagate TypeError if not iterable
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break // Iterator finished
		}
		if object.IsError(item) { // Check for errors yielded by iterator
			return item
		}
		elements = append(elements, item)
	}
	return &object.Tuple{Elements: elements}
}

func pySetFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesSetArgCountError, len(args))
	}

	resultSet := &object.Set{Elements: make(map[object.HashKey]object.Object)}

	if len(args) == 0 { // set() -> empty set
		return resultSet
	}

	iterableArg := args[0]
	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		return errObj // Propagate TypeError if not iterable
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(item) {
			return item
		}

		hashableItem, okHash := item.(object.Hashable)
		if !okHash {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesSetUnhashableType, item.Type())
		}
		hKey, err := hashableItem.HashKey()
		if err != nil {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesSetHashFailed, err)
		}
		resultSet.Elements[hKey] = item
	}
	return resultSet
}

// Helper function to handle the logic for both bytes() and bytearray()
func createBytesOrByteArray(ctx object.ExecutionContext, args ...object.Object) ([]byte, object.Object) {
	if len(args) == 0 {
		return []byte{}, nil // bytes() or bytearray() -> empty
	}
	if len(args) > 3 {
		return nil, object.NewError(constants.TypeError, constants.BuiltinsBytes_OR_BytearrayConstructorTakesAtMost3Arguments)
	}

	source := args[0]
	switch src := source.(type) {
	case *object.Integer:
		// bytes(int) -> create zero-filled bytes of that size
		if len(args) != 1 {
			return nil, object.NewError(constants.TypeError, constants.BuiltinsIntegerArgumentMustBeSolitary)
		}
		size := src.Value
		if size < 0 {
			return nil, object.NewError(constants.ValueError, constants.BuiltinsNegativeCount)
		}
		return make([]byte, size), nil

	case *object.String:
		// bytes(string, encoding, [errors])
		if len(args) < 2 {
			return nil, object.NewError(constants.TypeError, constants.BuiltinsStringArgumentWithoutAnEncoding)
		}
		// NOTE: For simplicity, we ignore the encoding and errors arguments and assume utf-8
		return []byte(src.Value), nil

	default:
		// Try to treat as an iterable of integers
		iterator, errObj := object.GetObjectIterator(ctx, src, object.NoToken)
		if errObj != nil {
			return nil, object.NewError(constants.TypeError, constants.BuiltinsObjectIsNotAnIterableOrCannotBeInterpretedAsBytes)
		}

		var resultBytes []byte
		for {
			item, stop := iterator.Next()
			if stop {
				break
			}
			if object.IsError(item) {
				return nil, item
			}
			intVal, ok := item.(*object.Integer)
			if !ok {
				return nil, object.NewError(constants.TypeError, constants.BuiltinsBytesLikeObjectRequiredNot_STRINGFORMATER, item.Type())
			}
			if intVal.Value < 0 || intVal.Value > 255 {
				return nil, object.NewError(constants.ValueError, constants.BuiltinsBytesMustBeInRange_0_255)
			}
			resultBytes = append(resultBytes, byte(intVal.Value))
		}
		return resultBytes, nil
	}
}

func pyBytesFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	b, err := createBytesOrByteArray(ctx, args...)
	if err != nil {
		return err // err is already an object.Object
	}
	return &object.Bytes{Value: b}
}

func pyByteArrayFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	b, err := createBytesOrByteArray(ctx, args...)
	if err != nil {
		return err // err is already an object.Object
	}
	// Note: We need to import the object package where ByteArray is defined
	return &object.ByteArray{Value: b}
}

func pyFrozenSetFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// TODO: Implement similar to set(), returning a FrozenSet object
	return object.NewError(constants.NotImplementedError, constants.BuiltinsTypesFrozensetNotImplemented)
}
func pyMemoryViewFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// TODO: Implement memoryview() (requires buffer protocol concept)
	return object.NewError(constants.NotImplementedError, constants.BuiltinsTypesMemoryViewNotImplemented)
}
func pyComplexFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// TODO: Implement complex() (requires Complex object type)
	return object.NewError(constants.NotImplementedError, constants.BuiltinsTypesComplexNotImplemented)
}
// --- float ---
// Accepts ExecutionContext (needed for potential __float__ call)
func pyFloatFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// float() with no arguments returns 0.0
	if len(args) == 0 {
		return &object.Float{Value: 0.0}
	}

	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsFloatTakesAtMost1Argument_DIGITFORMATER_Given, len(args))
	}

	arg := args[0]

	// --- Check for __float__ method ---
	// TODO: Implement dunder method call using ctx: __float__
	// if inst, ok := arg.(*object.Instance); ok { ... check methods ... }

	// --- Default conversion logic ---
	switch obj := arg.(type) {
	case *object.Float:
		// It's already a float, return it directly
		return obj
	case *object.Integer:
		// Convert integer to float
		return &object.Float{Value: float64(obj.Value)}
	case *object.Boolean:
		// Convert boolean to float (True -> 1.0, False -> 0.0)
		if obj.Value {
			return &object.Float{Value: 1.0}
		}
		return &object.Float{Value: 0.0}
	case *object.String:
		// Trim whitespace from the string before parsing
		strVal := strings.TrimSpace(obj.Value)
		lowerStrVal := strings.ToLower(strVal)

		// Check for special string values like 'inf' and 'nan'
		switch lowerStrVal {
		case constants.BuiltinsInf, constants.Builtins_PLUS_Inf,constants.BuiltinsInfinity, constants.Builtins_PLUS_Infinity:
			return &object.Float{Value: math.Inf(1)}
		case constants.Builtins_MINUS_Inf, constants.Builtins_MINUS_Infinity:
			return &object.Float{Value: math.Inf(-1)}
		case constants.BuiltinsNaN, constants.Builtins_PLUS_NaN, constants.Builtins_MINUS_NaN: // Python treats all NaN variations the same
			return &object.Float{Value: math.NaN()}
		}

		// Attempt to parse the string as a standard float
		val, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			// If parsing fails, return a ValueError, similar to Python
			return object.NewError(constants.ValueError, constants.BuiltinsCouldNotConvertStringToFloat_STRINGFORMATER, obj.Inspect())
		}
		return &object.Float{Value: val}

	default:
		// For any other type, return a TypeError
		return object.NewError(constants.TypeError, constants.BuiltinsFloatArgumentMustBeStringOrNumberNot_STRINGFORMATER, arg.Type())
	}
}


// --- Registration ---
// Ensure function signatures match object.BuiltinFunction
func init() {
	registerBuiltin(constants.BuiltinsTypeFuncName, &object.Builtin{Name: constants.BuiltinsTypeFuncName, Fn: pyTypeFn})
	registerBuiltin(constants.BuiltinsIntFuncName, &object.Builtin{Name: constants.BuiltinsIntFuncName, Fn: pyIntFn})
	registerBuiltin(constants.BuiltinsFloatFuncName, &object.Builtin{Name: constants.BuiltinsFloatFuncName, Fn: pyFloatFn})
	registerBuiltin(constants.BuiltinsStrFuncName, &object.Builtin{Name: constants.BuiltinsStrFuncName, Fn: pyStrFn})
	registerBuiltin(constants.BuiltinsBoolFuncName, &object.Builtin{Name: constants.BuiltinsBoolFuncName, Fn: pyBoolFn})
	registerBuiltin(constants.BuiltinsListFuncName, &object.Builtin{Name: constants.BuiltinsListFuncName, Fn: pyListFn})
	registerBuiltin(constants.BuiltinsDictFuncName, &object.Builtin{Name: constants.BuiltinsDictFuncName, Fn: pyDictFn})
	registerBuiltin(constants.BuiltinsReprFuncName, &object.Builtin{Name: constants.BuiltinsReprFuncName, Fn: pyReprFn})
	registerBuiltin(constants.BuiltinsAsciiFuncName, &object.Builtin{Name: constants.BuiltinsAsciiFuncName, Fn: pyAsciiFn})
	registerBuiltin(constants.BuiltinsTupleFuncName, &object.Builtin{Name: constants.BuiltinsTupleFuncName, Fn: pyTupleFn})
	registerBuiltin(constants.BuiltinsSetFuncName, &object.Builtin{Name: constants.BuiltinsSetFuncName, Fn: pySetFn})
	registerBuiltin(constants.BuiltinsBytesFuncName, &object.Builtin{Name: constants.BuiltinsBytesFuncName, Fn: pyBytesFn})
	registerBuiltin(constants.BuiltinsByteArrayFuncName, &object.Builtin{Name: constants.BuiltinsByteArrayFuncName, Fn: pyByteArrayFn})

	// Placeholders
	registerBuiltin(constants.BuiltinsFrozensetFuncName, &object.Builtin{Name: constants.BuiltinsFrozensetFuncName, Fn: pyFrozenSetFn})
	registerBuiltin(constants.BuiltinsMemoryViewFuncName, &object.Builtin{Name: constants.BuiltinsMemoryViewFuncName, Fn: pyMemoryViewFn})
	registerBuiltin(constants.BuiltinsComplexFuncName, &object.Builtin{Name: constants.BuiltinsComplexFuncName, Fn: pyComplexFn})
}
