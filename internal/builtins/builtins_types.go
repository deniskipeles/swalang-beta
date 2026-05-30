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
func pyTypeFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesTypeArgCountError, len(args))
	}
	obj := args[0]

	if inst, ok := obj.(*object.Instance); ok && inst.Class != nil {
		return inst.Class 
	}
	if classObj, ok := obj.(*object.Class); ok {
		return classObj 
	}

	var typeName string
	switch obj.Type() { 
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
		typeName = constants.BuiltinsFunctionType 
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
		typeName = constants.BuiltinsTextIOWrapper 
	case object.ITERATOR_OBJ:
		typeName = constants.BuiltinsIteratorType 
	case object.STOP_ITER_OBJ:
		typeName = constants.StopIteration 
	case object.BOUND_METHOD_OBJ:
		typeName = constants.BuiltinsMethodType 
	default:
		typeName = string(obj.Type()) 
	}
	return &object.String{Value: fmt.Sprintf(constants.BuiltinsClassFormat, typeName)}
}

// --- int ---
func pyIntFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	base := 10 
	var arg object.Object

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
		if base != 0 && (base < 2 || base > 36) {
			return object.NewError(constants.ValueError, constants.BuiltinsTypesIntBaseRangeError)
		}
		if _, isStr := arg.(*object.String); !isStr {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		}

	default:
		return object.NewError(constants.TypeError, constants.BuiltinsMathIntPowerResultTooLarge, len(args))
	}

	switch obj := arg.(type) {
	case *object.Integer:
		if base != 10 && base != 0 {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		} 
		return obj 
	case *object.Float:
		if base != 10 && base != 0 {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesIntNonStringBaseError)
		}
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
		strVal := strings.TrimSpace(obj.Value)
		cleanedVal, detectedBase, errBase := handleIntPrefix(strVal, base)
		if errBase != nil {
			return object.NewError(constants.ValueError, errBase.Error())
		}

		val, err := strconv.ParseInt(cleanedVal, detectedBase, 64)
		if err != nil {
			return object.NewError(constants.ValueError, constants.BuiltinsTypesIntInvalidLiteral, detectedBase, obj.Inspect())
		}
		return &object.Integer{Value: val}
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsTypesIntArgTypeError, obj.Type())
	}
}

func handleIntPrefix(s string, requestedBase int) (string, int, error) {
	val := s
	base := requestedBase

	if strings.HasPrefix(val, constants.PlusSign) || strings.HasPrefix(val, constants.MinusSign) {
		val = val[1:]
	}

	prefix := constants.EmptyString
	detectedBase := 0 

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
			base = detectedBase 
		} else if base != detectedBase {
			return constants.EmptyString, 0, fmt.Errorf(constants.BuiltinsTypesIntInvalidIntLiteral, base, s)
		}
		s = strings.Replace(s, prefix, constants.EmptyString, 1) 
		if strings.HasPrefix(s, constants.PlusSign) || strings.HasPrefix(s, constants.MinusSign) {
		} else if strings.HasPrefix(s, constants.Space) { 
			return constants.EmptyString, 0, fmt.Errorf(constants.BuiltinsTypesIntInvalidIntLiteral, base, s)
		}

	} else {
		if base == 0 {
			base = 10
		} 
	}

	return s, base, nil
}

// --- str ---
func pyStrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesStrArgCountError, len(args))
	}
	arg := args[0]

	if strObj, ok := arg.(*object.String); ok {
		return strObj
	}

	if classObj, ok := arg.(*object.Class); ok {
		return &object.String{Value: classObj.Inspect()}
	}

	strMethod, strFound := object.CallGetAttr(ctx, arg, constants.DunderStr, object.NoToken)
	if strFound && !object.IsError(strMethod) && object.IsCallable(strMethod) {
		result := ctx.Execute(strMethod)
		if object.IsError(result) {
			return result 
		}
		if _, isStr := result.(*object.String); !isStr {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesStrStrNonStringError, result.Type())
		}
		return result
	}

	return pyReprFn(ctx, arg)
}

// --- repr() ---
func pyReprFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesReprArgCountError, len(args))
	}
	arg := args[0]

	if classObj, ok := arg.(*object.Class); ok {
		return &object.String{Value: classObj.Inspect()}
	}

	reprMethod, reprFound := object.CallGetAttr(ctx, arg, constants.DunderRepr, object.NoToken)
	if reprFound && !object.IsError(reprMethod) && object.IsCallable(reprMethod) {
		result := ctx.Execute(reprMethod)
		if object.IsError(result) {
			return result 
		}
		if _, isStr := result.(*object.String); !isStr {
			return object.NewError(constants.TypeError, constants.BuiltinsTypesStrReprNonStringError, result.Type())
		}
		return result
	}

	return &object.String{Value: arg.Inspect()}
}

// --- bool() ---
func pyBoolFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesBoolArgCountError, len(args))
	}
	if len(args) == 0 {
		return object.FALSE 
	}
	truthy, err := object.IsTruthy(ctx, args[0])
	if err != nil {
		if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
			return pyErr 
		}
		return object.NewError(constants.RuntimeError, constants.BuiltinsTypesBoolPropagatedFromIsTruthy, err)
	}
	return object.NativeBoolToBooleanObject(truthy)
}

// --- list() ---
func pyListFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	token := lexer.Token{Line: -1, Column: 0, Type: lexer.ILLEGAL, Literal: constants.BuiltinsListFuncName} 
	if len(args) > 1 {
		return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsTypesListArgCountError, len(args))
	}

	elements := []object.Object{}

	if len(args) == 0 {
		return &object.List{Elements: elements} 
	}

	iterableArg := args[0]

	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, token)
	if errObj != nil {
		return errObj
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break 
		}
		if object.IsError(item) { 
			return item
		}
		elements = append(elements, item)
	}

	return &object.List{Elements: elements}
}

// --- dict() ---
func pyDictFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// Our new built-in dispatch injects kwargs as the LAST argument if they exist
	var kwargs *object.Dict
	var positionalArg object.Object

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if dictObj, isDict := lastArg.(*object.Dict); isDict {
			// In Swalang's current interpreter setup, kwargs are bundled into a dict
			// at the end of the arguments list if the Builtin has AcceptsKeywords map defined.
			kwargs = dictObj
			if len(args) > 1 {
				positionalArg = args[0]
			}
		} else {
			positionalArg = args[0]
		}
	}

	if len(args) > 2 || (len(args) == 2 && kwargs == nil) {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesDictArgCountError)
	}

	newDict := &object.Dict{Pairs: make(map[object.HashKey]object.DictPair)}

	// 1. Process positional argument (copy from existing dict or iterate pairs)
	if positionalArg != nil {
		if sourceDict, ok := positionalArg.(*object.Dict); ok {
			// Shallow copy the source dictionary
			for k, v := range sourceDict.Pairs {
				newDict.Pairs[k] = v
			}
		} else {
			// Treat as iterable of pairs e.g. dict([("a", 1), ("b", 2)])
			iterator, errObj := object.GetObjectIterator(ctx, positionalArg, object.NoToken)
			if errObj == nil {
				for {
					item, stop := iterator.Next()
					if stop {
						break
					}
					if object.IsError(item) {
						return item
					}
					
					// Each item must be an iterable of length 2
					pairIter, pairErr := object.GetObjectIterator(ctx, item, object.NoToken)
					if pairErr != nil {
						return object.NewError(constants.TypeError, "cannot convert dictionary update sequence element to a sequence")
					}
					
					k, stop1 := pairIter.Next()
					v, stop2 := pairIter.Next()
					_, stop3 := pairIter.Next() // Patched here
					
					if stop1 || stop2 || !stop3 {
						return object.NewError(constants.ValueError, "dictionary update sequence element has incorrect length")
					}
					
					hashableKey, ok := k.(object.Hashable)
					if !ok {
						return object.NewError(constants.TypeError, constants.EvalExpressionsUnhashableType, k.Type())
					}
					hashed, hashErr := hashableKey.HashKey()
					if hashErr != nil {
						return object.NewError(constants.TypeError, constants.EvalExpressionsFailedToHashKey, hashErr)
					}
					newDict.Pairs[hashed] = object.DictPair{Key: k, Value: v}
				}
			} else {
				return object.NewError(constants.TypeError, "'%s' object is not iterable", positionalArg.Type())
			}
		}
	}

	// 2. Process kwargs (e.g. dict(a=1, b=2))
	if kwargs != nil {
		for k, v := range kwargs.Pairs {
			newDict.Pairs[k] = v
		}
	}

	return newDict
}


// --- ascii() ---
func pyAsciiFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesAsciiArgCountError, len(args))
	}

	reprObj := pyReprFn(ctx, args...) 
	if object.IsError(reprObj) {
		return reprObj 
	}
	reprStr := reprObj.(*object.String).Value

	var builder strings.Builder
	for _, r := range reprStr {
		if r < 128 { 
			builder.WriteRune(r)
		} else if r <= 0xff {
			fmt.Fprintf(&builder, constants.HexEscapeFormat, r)
		} else if r <= 0xffff {
			fmt.Fprintf(&builder, constants.UnicodeEscapeFormat, r)
		} else { 
			fmt.Fprintf(&builder, constants.UniversalCharacterNameEscapeFormat, r)
		}
	}
	return &object.String{Value: builder.String()}
}

func pyTupleFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsTypesTupleArgCountError, len(args))
	}

	elements := []object.Object{}

	if len(args) == 0 { 
		return &object.Tuple{Elements: elements}
	}

	iterableArg := args[0]
	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		return errObj 
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break 
		}
		if object.IsError(item) { 
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

	if len(args) == 0 { 
		return resultSet
	}

	iterableArg := args[0]
	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		return errObj 
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

func createBytesOrByteArray(ctx object.ExecutionContext, args ...object.Object) ([]byte, object.Object) {
	if len(args) == 0 {
		return []byte{}, nil 
	}
	if len(args) > 3 {
		return nil, object.NewError(constants.TypeError, constants.BuiltinsBytes_OR_BytearrayConstructorTakesAtMost3Arguments)
	}

	source := args[0]
	switch src := source.(type) {
	case *object.Integer:
		if len(args) != 1 {
			return nil, object.NewError(constants.TypeError, constants.BuiltinsIntegerArgumentMustBeSolitary)
		}
		size := src.Value
		if size < 0 {
			return nil, object.NewError(constants.ValueError, constants.BuiltinsNegativeCount)
		}
		return make([]byte, size), nil

	case *object.String:
		if len(args) < 2 {
			return nil, object.NewError(constants.TypeError, constants.BuiltinsStringArgumentWithoutAnEncoding)
		}
		return []byte(src.Value), nil

	default:
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
		return err 
	}
	return &object.Bytes{Value: b}
}

func pyByteArrayFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	b, err := createBytesOrByteArray(ctx, args...)
	if err != nil {
		return err 
	}
	return &object.ByteArray{Value: b}
}

func pyFrozenSetFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsTypesFrozensetNotImplemented)
}
func pyMemoryViewFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsTypesMemoryViewNotImplemented)
}
func pyComplexFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsTypesComplexNotImplemented)
}

// --- float ---
func pyFloatFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) == 0 {
		return &object.Float{Value: 0.0}
	}

	if len(args) > 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsFloatTakesAtMost1Argument_DIGITFORMATER_Given, len(args))
	}

	arg := args[0]

	switch obj := arg.(type) {
	case *object.Float:
		return obj
	case *object.Integer:
		return &object.Float{Value: float64(obj.Value)}
	case *object.Boolean:
		if obj.Value {
			return &object.Float{Value: 1.0}
		}
		return &object.Float{Value: 0.0}
	case *object.String:
		strVal := strings.TrimSpace(obj.Value)
		lowerStrVal := strings.ToLower(strVal)

		switch lowerStrVal {
		case constants.BuiltinsInf, constants.Builtins_PLUS_Inf,constants.BuiltinsInfinity, constants.Builtins_PLUS_Infinity:
			return &object.Float{Value: math.Inf(1)}
		case constants.Builtins_MINUS_Inf, constants.Builtins_MINUS_Infinity:
			return &object.Float{Value: math.Inf(-1)}
		case constants.BuiltinsNaN, constants.Builtins_PLUS_NaN, constants.Builtins_MINUS_NaN: 
			return &object.Float{Value: math.NaN()}
		}

		val, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return object.NewError(constants.ValueError, constants.BuiltinsCouldNotConvertStringToFloat_STRINGFORMATER, obj.Inspect())
		}
		return &object.Float{Value: val}

	default:
		return object.NewError(constants.TypeError, constants.BuiltinsFloatArgumentMustBeStringOrNumberNot_STRINGFORMATER, arg.Type())
	}
}


// --- Registration ---
func init() {
	registerBuiltin(constants.BuiltinsTypeFuncName, &object.Builtin{Name: constants.BuiltinsTypeFuncName, Fn: pyTypeFn})
	registerBuiltin(constants.BuiltinsIntFuncName, &object.Builtin{Name: constants.BuiltinsIntFuncName, Fn: pyIntFn})
	registerBuiltin(constants.BuiltinsFloatFuncName, &object.Builtin{Name: constants.BuiltinsFloatFuncName, Fn: pyFloatFn})
	registerBuiltin(constants.BuiltinsStrFuncName, &object.Builtin{Name: constants.BuiltinsStrFuncName, Fn: pyStrFn})
	registerBuiltin(constants.BuiltinsBoolFuncName, &object.Builtin{Name: constants.BuiltinsBoolFuncName, Fn: pyBoolFn})
	registerBuiltin(constants.BuiltinsListFuncName, &object.Builtin{Name: constants.BuiltinsListFuncName, Fn: pyListFn})
	
	// Create dict builtin that accepts arbitrary keyword arguments
	dictBuiltin := &object.Builtin{
		Name: constants.BuiltinsDictFuncName, 
		Fn: pyDictFn,
		// Explicitly map it to accept ANY keyword arguments, otherwise they are rejected.
		// Since we don't know the keys in advance for `dict(a=1, b=2)`, we use a sentinel or skip strict validation.
		// To signal the engine to bundle keywords, we just initialize the map.
		AcceptsKeywords: make(map[string]bool),
	}
	registerBuiltin(constants.BuiltinsDictFuncName, dictBuiltin)
	
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