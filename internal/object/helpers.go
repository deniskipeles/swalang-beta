// pylearn/internal/object/helpers.go
package object

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
)

// =============================================================================
//  NEW: Native Method Registry
// =============================================================================

// nativeMethodRegistry stores Go functions registered as methods for native Pylearn types.
// Key: ObjectType (e.g., "net.Connection")
// Value: Map of method names to the Go function implementation.
var (
	nativeMethodRegistry      = make(map[ObjectType]map[string]BuiltinFunction)
	nativeMethodRegistryMutex sync.RWMutex
)

// SetNativeMethod registers a Go function as a native method for a given Pylearn object type.
// This is called during the init() phase of a native module.
func SetNativeMethod(objType ObjectType, name string, method BuiltinFunction) {
	nativeMethodRegistryMutex.Lock()
	defer nativeMethodRegistryMutex.Unlock()

	if _, ok := nativeMethodRegistry[objType]; !ok {
		nativeMethodRegistry[objType] = make(map[string]BuiltinFunction)
	}
	nativeMethodRegistry[objType][name] = method
}

// GetNativeMethod retrieves a native method for a given object instance.
// It returns a new *Builtin object (a bound method) where the first argument
// to the underlying Go function is transparently set to the `self` object.
func GetNativeMethod(self Object, name string) (Object, bool) {
	nativeMethodRegistryMutex.RLock()
	defer nativeMethodRegistryMutex.RUnlock()

	objType := self.Type()
	methods, ok := nativeMethodRegistry[objType]
	if !ok {
		return nil, false // No methods registered for this type
	}

	methodFn, ok := methods[name]
	if !ok {
		return nil, false // This specific method is not registered
	}

	// Create a new Builtin object that acts as a bound method.
	// The closure captures the `self` instance.
	boundMethod := &Builtin{
		Name: fmt.Sprintf("%s.%s", objType, name),
		Fn: func(ctx ExecutionContext, scriptArgs ...Object) Object {
			// Prepend `self` to the arguments passed from the script.
			finalArgs := make([]Object, 1+len(scriptArgs))
			finalArgs[0] = self
			copy(finalArgs[1:], scriptArgs)

			// Call the original Go function with the complete set of arguments.
			return methodFn(ctx, finalArgs...)
		},
	}

	return boundMethod, true
}

// Used when invoking magic methods internally where no specific source token is relevant
var NoToken = lexer.Token{Type: lexer.ILLEGAL, Literal: constants.HELPER_NO_TOKEN_LITERAL, Line: 0, Column: 0}

// IsError checks if an object is an Error or StopIterationError.
func IsError(obj Object) bool {
	if obj != nil {
		// Check for regular Error OR the specific StopIterationError
		return obj.Type() == ERROR_OBJ || obj.Type() == STOP_ITER_OBJ
	}
	return false
}

// IsStopIteration checks specifically for the StopIteration signal.
func IsStopIteration(obj Object) bool {
	if obj != nil {
		return obj.Type() == STOP_ITER_OBJ
	}
	return false
}

// --- Numeric Helpers ---

// isNumeric checks if an object is an Integer or Float.
func IsNumeric(obj Object) bool {
	t := obj.Type()
	return t == INTEGER_OBJ || t == FLOAT_OBJ
}

// promoteToFloats converts two numeric objects (int/float) to two Float objects.
func PromoteToFloats(left, right Object) (*Float, *Float) {
	lF := &Float{}
	if f, ok := left.(*Float); ok {
		lF = f
	} else {
		lF.Value = float64(left.(*Integer).Value)
	}

	rF := &Float{}
	if f, ok := right.(*Float); ok {
		rF = f
	} else {
		rF.Value = float64(right.(*Integer).Value)
	}
	return lF, rF
}

// NativeBoolToBooleanObject converts a Go bool to our Boolean object instance. // <<< RENAMED
func NativeBoolToBooleanObject(input bool) *Boolean { // <<< RENAMED
	if input {
		return TRUE
	}
	return FALSE
}

// --- NEW: Error Constructor Functions (Moved from interpreter) ---

// NewError creates a new Pylearn Error object of a given class.
func NewError(errorClassName string, format string, args ...interface{}) *Error {
	errCls, ok := BuiltinExceptionClasses[errorClassName]
	if !ok {
		fmt.Fprintf(os.Stderr, constants.HELPER_NEW_ERROR_UNKNOWN_CLASS_WARNING, errorClassName)
		errCls = BuiltinExceptionClasses[constants.Exception]
		if errCls == nil { // Should absolutely not happen if init ran
			panic(fmt.Sprintf(constants.HELPER_NEW_ERROR_CRITICAL_EXCEPTION_NOT_FOUND, fmt.Sprintf(format, args...)))
		}
	}
	return &Error{
		Message:    fmt.Sprintf(format, args...),
		ErrorClass: errCls,
	}
}

// NewErrorWithLocation creates a new Error object including line/column.
func NewErrorWithLocation(token lexer.Token, errorClassName string, format string, args ...interface{}) *Error {
	err := NewError(errorClassName, format, args...)
	if !(token.Type == lexer.ILLEGAL && token.Literal == constants.HELPER_NO_TOKEN_LITERAL) {
		err.Line = token.Line
		err.Column = token.Column
	}
	return err
}

// NewErrorWithToken creates a runtime error object including source location.
func NewErrorWithToken(fallbackLiteral string, token lexer.Token, errorClassName string, format string, args ...interface{}) *Error {
	return NewErrorWithLocation(token, errorClassName, format, args...)
}

// BasicIsTruthy - Checks simple cases without context/dunder methods
func BasicIsTruthy(obj Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	}
	// Add other non-dunder checks if needed (numbers, maybe basic containers)
	switch obj := obj.(type) {
	case *Integer:
		return obj.Value != 0
	case *Float:
		return obj.Value != 0.0
	}
	// Default to true for unknown basic types? Or false? Python defaults to true.
	return true
}

// IsTruthy determines the boolean value of an object based on Python's rules.
// Requires an ExecutionContext to handle potential __bool__ or __len__ calls.
func IsTruthy(ctx ExecutionContext, obj Object) (bool, error) { // Added ctx, return error
	// Handle singletons first
	switch obj {
	case TRUE:
		return true, nil
	case FALSE:
		return false, nil
	case NULL:
		return false, nil
	}

	// Numeric zero values
	switch obj := obj.(type) {
	case *Integer:
		return obj.Value != 0, nil
	case *Float:
		return obj.Value != 0.0, nil
	}

	// --- Check for __bool__ method ---
	if inst, ok := obj.(*Instance); ok && inst.Class != nil { // Check inst.Class nil
		// if boolMethod, methodOk := inst.Class.Methods["__bool__"]; methodOk {
		// 	boundBool := &BoundMethod{Instance: inst, Method: boolMethod}
		// 	// Use the generic ApplyBoundMethod helper, passing the context
		// 	result := ApplyBoundMethod(ctx, boundBool, []Object{}, lexer.Token{}) // Use placeholder token
		// 	if IsError(result) {
		if boolMethodObj, methodOk := inst.Class.Methods[constants.DunderBool]; methodOk {
			// <<< FIX: Type-assert that the found method is a callable function >>>
			if boolMethod, isFunc := boolMethodObj.(*Function); isFunc {
				boundBool := &BoundMethod{Instance: inst, Method: boolMethod}
				result := ApplyBoundMethod(ctx, boundBool, []Object{}, NoToken)
				if IsError(result) {
					// Propagate the error from __bool__
					// Need to ensure result IS an error type if IsError is true
					if errResult, ok := result.(error); ok {
						return false, errResult // Return false and the error
					}
					// Fallback if IsError was true but type assertion failed (shouldn't happen)
					return false, NewError(constants.InternalError, constants.HELPER_TRUTHY_INTERNAL_ERROR_UNKNOWN_TYPE, result)
				}
				boolResult, isBool := result.(*Boolean)
				if !isBool {
					// Python raises TypeError here. Propagate as error.
					return false, NewError(constants.TypeError, constants.HELPER_TRUTHY_BOOL_RETURN_TYPE_ERROR, result.Type())
				}
				return boolResult.Value, nil // Return the boolean value and nil error
			}
			// No __bool__, fall through to check __len__
		}
	}

	// --- Check for __len__ method ---
	var length int64 = -1  // Sentinel value means length not determined
	var lenErr error = nil // Track error from __len__ call

	// if inst, ok := obj.(*Instance); ok && inst.Class != nil { // Check inst.Class nil
	// 	if lenMethod, methodOk := inst.Class.Methods["__len__"]; methodOk {
	// 		boundLen := &BoundMethod{Instance: inst, Method: lenMethod}
	// 		result := ApplyBoundMethod(ctx, boundLen, []Object{}, lexer.Token{})
	if inst, ok := obj.(*Instance); ok && inst.Class != nil {
		if lenMethodObj, methodOk := inst.Class.Methods[constants.DunderLen]; methodOk {
			// <<< FIX: Type-assert that the found method is a callable function >>>
			if lenMethod, isFunc := lenMethodObj.(*Function); isFunc {
				boundLen := &BoundMethod{Instance: inst, Method: lenMethod}
				result := ApplyBoundMethod(ctx, boundLen, []Object{}, NoToken)
				if IsError(result) {
					// Propagate error from __len__
					if errResult, ok := result.(error); ok {
						lenErr = errResult
					} else {
						lenErr = NewError(constants.InternalError, constants.HELPER_TRUTHY_INTERNAL_ERROR_UNKNOWN_LEN_TYPE, result)
					}
				} else if intResult, isInt := result.(*Integer); isInt {
					if intResult.Value < 0 {
						// Python raises ValueError here.
						lenErr = NewError(constants.ValueError, constants.HELPER_TRUTHY_LEN_SHOULD_RETURN_NON_NEGATIVE)
					} else {
						length = intResult.Value // Store valid length
					}
				} else {
					// Python raises TypeError here.
					lenErr = NewError(constants.TypeError, constants.HELPER_TRUTHY_LEN_RETURN_TYPE_ERROR, result.Type())
				}
				// If __len__ call produced an error, return it now
				if lenErr != nil {
					return false, lenErr
				}
			}
		}
	}

	// Use length if successfully determined via __len__
	if length != -1 {
		return length != 0, nil // Truthy if length > 0, nil error
	}

	// --- Default truthiness for built-ins based on length/content ---
	// (No error possible here)
	switch obj := obj.(type) {
	case *String:
		return len(obj.Value) > 0, nil
	case *List:
		return len(obj.Elements) > 0, nil
	case *Dict:
		return len(obj.Pairs) > 0, nil
	case *Tuple:
		return len(obj.Elements) > 0, nil
	case *Set:
		return len(obj.Elements) > 0, nil
	case *Bytes:
		return len(obj.Value) > 0, nil
	default:
		// Other types (Functions, Builtins, Modules, Ranges, Instances without __bool__/__len__, etc.) are truthy
		return true, nil
	}
}

// ApplyBoundMethod prepares and executes a bound method call using the provided execution context.
// It handles arity checking and prepends the instance ('self') to the arguments before execution.
func ApplyBoundMethod(
	ctx ExecutionContext, // The context used to execute the method
	boundMethod *BoundMethod,
	args []Object,
	callToken lexer.Token, // Token for error reporting (usually '(')
) Object {
	method := boundMethod.Method     // This is the *object.Function
	instance := boundMethod.Instance // This is 'self'

	// --- Arity Check ---
	// Compare # args provided + 1 (for self) with # params defined in AST
	numExpectedParams := len(method.Parameters)
	numProvidedArgs := len(args)
	numTotalArgsSupplied := numProvidedArgs + 1 // self + provided args

	if numTotalArgsSupplied != numExpectedParams {
		// Construct helpful error message
		methodName := method.Name
		if methodName == constants.EmptyString {
			methodName = constants.FUNCTION_INSPECT_METHOD_NAME
		}

		className := constants.HELPER_APPLY_BOUND_METHOD_CLASS_PLACEHOLDER // Default
		// Safely get class name (handle potential different instance types later)
		if inst, ok := instance.(*Instance); ok {
			if inst.Class != nil { // Check if Class is not nil
				className = inst.Class.Name
			} else {
				className = constants.HELPER_APPLY_BOUND_METHOD_INSTANCE_NIL_CLASS
			}
		} else {
			// Could try instance.Type() for non-Instance types, but might be less informative
			className = string(instance.Type())
		}

		return NewErrorWithLocation(callToken, constants.TypeError, constants.HELPER_APPLY_BOUND_METHOD_ARITY_ERROR,
			className, methodName, numExpectedParams, numTotalArgsSupplied)
	}

	// --- Prepare Arguments ---
	// Prepend 'self' (the instance) to the argument list
	combinedArgs := make([]Object, 0, numTotalArgsSupplied)
	combinedArgs = append(combinedArgs, instance)
	combinedArgs = append(combinedArgs, args...)

	// --- Delegate Execution ---
	// Execute the method function itself using the context.
	// The context's implementation (Interpreter or VM) handles environment setup
	// and execution based on the combined arguments.
	return ctx.Execute(method, combinedArgs...)
}

// IsInstance checks if an object is an instance of a class or a tuple of classes.
func IsInstance(ctx ExecutionContext, obj Object, classinfo Object) (bool, Object) {
	if tuple, ok := classinfo.(*Tuple); ok {
		for _, typeInTuple := range tuple.Elements {
			match, err := IsInstance(ctx, obj, typeInTuple)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil
	}

	targetClass, ok := classinfo.(*Class)
	if !ok {
		return false, NewError(constants.TypeError, constants.HELPER_IS_INSTANCE_ARG2_TYPE_ERROR, classinfo.Type())
	}

	var objClass *Class
	switch o := obj.(type) {
	case *Instance:
		if o.Class == nil {
			return false, NewError(constants.InternalError, constants.HELPER_IS_INSTANCE_INTERNAL_NIL_CLASS)
		}
		objClass = o.Class
	case *Error:
		if o.ErrorClass == nil {
			return false, NewError(constants.InternalError, constants.HELPER_IS_INSTANCE_INTERNAL_NIL_ERROR_CLASS)
		}
		objClass = o.ErrorClass
	case *StopIterationError:
		if o.ErrorClass == nil {
			return false, NewError(constants.InternalError, constants.HELPER_IS_INSTANCE_INTERNAL_NIL_STOP_ITERATION_CLASS)
		}
		objClass = o.ErrorClass
	default:
		// For built-in types like Integer, String, they don't have a *Class object yet.
		// They can't be instances of our defined exception classes.
		return false, nil
	}

	if objClass == nil || objClass.MRO == nil {
		// Not a class-based object or MRO not computed
		return false, NewError(constants.InternalError, constants.HELPER_IS_INSTANCE_MRO_NIL_ERROR)
	}

	// Walk the MRO of the object's class to check for a match.
	for _, clsInMRO := range objClass.MRO {
		if clsInMRO == targetClass {
			return true, nil
		}
	}

	return false, nil
}

// Placeholder for IsSubclass - needs similar logic and context
func IsSubclass(ctx ExecutionContext, class1 Object, class2 Object) (bool, Object) {
	// TODO: Implement similar checks as IsInstance, comparing Class objects and walking superclass chain.
	return false, NewError(constants.NotImplementedError, constants.HELPER_IS_SUBCLASS_NOT_IMPLEMENTED)
}

// CompareObjects implements comparison logic for various Pylearn objects.
// It prioritizes dunder methods (__eq__, __ne__, __lt__, etc.) if available.
// `op` should be one of: "==", "!=", "<", "<=", ">", ">=".
// `ctx` is the ExecutionContext required to call dunder methods.
func CompareObjects(op string, left, right Object, ctx ExecutionContext) Object {
	var dunderName string
	var reflectedDunderName string // For cases like other.__eq__(self)

	switch op {
	case constants.EqualsOperator:
		dunderName = constants.DunderEq
		reflectedDunderName = constants.DunderEq
	case constants.NotEqualsOperator:
		dunderName = constants.DunderNe
		reflectedDunderName = constants.DunderNe
	case constants.LessThanOp:
		dunderName = constants.DunderLt
		reflectedDunderName = constants.DunderGt // if a < b, then b > a
	case constants.LessThanEqualsOp:
		dunderName = constants.DunderLe
		reflectedDunderName = constants.DunderGe // if a <= b, then b >= a
	case constants.GreaterThanOp:
		dunderName = constants.DunderGt
		reflectedDunderName = constants.DunderLt // if a > b, then b < a
	case constants.GreaterThanEqualsOp:
		dunderName = constants.DunderGe
		reflectedDunderName = constants.DunderLe // if a >= b, then b <= a
	default:
		return NewError(constants.InternalError, constants.HELPER_COMPARE_OBJECTS_UNSUPPORTED_OP, op)
	}

	// --- Attempt 1: left.dunder(right) ---
	if leftGetter, hasGetAttrLeft := left.(AttributeGetter); hasGetAttrLeft {
		methodObj, found := leftGetter.GetObjectAttribute(ctx, dunderName)
		if found {
			if IsError(methodObj) { // Error retrieving the attribute itself
				return methodObj
			}

			var result Object
			// Call the retrieved method object
			if methodBuiltin, isBuiltin := methodObj.(*Builtin); isBuiltin {
				// Builtin's Fn might be a closure from GetObjectAttribute (e.g., Dict.__contains__ check).
				// Such closures usually expect (self, other_arg).
				// Here 'left' is self, 'right' is other_arg.
				result = methodBuiltin.Fn(ctx, left, right)
			} else if boundMeth, isBound := methodObj.(*BoundMethod); isBound {
				// This typically happens if 'left' is an Instance and dunderName is a method
				// defined in its Pylearn class.
				result = ApplyBoundMethod(ctx, boundMeth, []Object{right}, NoToken)
			} else {
				// The attribute was found but isn't a callable method as expected.
				// Python might raise a TypeError here, or it might mean __eq__ is a value (unlikely for std dunders).
				// For simplicity, let's assume it's an error or fall through.
				// For now, let's assume it implies NotImplemented for comparison path.
				result = NOT_IMPLEMENTED
			}

			if result != nil && result != NOT_IMPLEMENTED {
				if IsError(result) {
					return result // Propagate error from dunder method call
				}
				// Dunder methods for comparison should return Boolean or NotImplemented.
				if boolRes, isBool := result.(*Boolean); isBool {
					// If op was "!=" and __ne__ was successfully called and returned a boolean, use it.
					// Otherwise, if __eq__ was called for "!=", negate its result.
					if op == constants.NotEqualsOperator && dunderName == constants.DunderNe {
						return boolRes
					}
					if op == constants.NotEqualsOperator && dunderName == constants.DunderEq {
						return NativeBoolToBooleanObject(!boolRes.Value)
					}
					return boolRes // For ==, <, <=, >, >=
				}
				return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_DUNDER_NON_BOOL_RETURN, dunderName, result.Type())
			}
			// If result is NOT_IMPLEMENTED, proceed to Attempt 2.
		}
	}

	// --- Attempt 2: right.reflected_dunder(left) ---
	// Only if 'left' didn't handle it or returned NotImplemented.
	// And only if the operator has a distinct reflected version (e.g., < vs >).
	if reflectedDunderName != constants.EmptyString && op != constants.EqualsOperator && op != constants.NotEqualsOperator { // == and != don't typically use different reflected names
		if rightGetter, hasGetAttrRight := right.(AttributeGetter); hasGetAttrRight {
			methodObj, found := rightGetter.GetObjectAttribute(ctx, reflectedDunderName)
			if found {
				if IsError(methodObj) {
					return methodObj
				}

				var result Object
				if methodBuiltin, isBuiltin := methodObj.(*Builtin); isBuiltin {
					result = methodBuiltin.Fn(ctx, right, left) // 'right' is self, 'left' is other
				} else if boundMeth, isBound := methodObj.(*BoundMethod); isBound {
					result = ApplyBoundMethod(ctx, boundMeth, []Object{left}, NoToken)
				} else {
					result = NOT_IMPLEMENTED
				}

				if result != nil && result != NOT_IMPLEMENTED {
					if IsError(result) {
						return result
					}
					if boolRes, isBool := result.(*Boolean); isBool {
						// No need to negate here, as reflected op should give direct answer.
						return boolRes
					}
					return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_DUNDER_NON_BOOL_RETURN, reflectedDunderName, result.Type())
				}
			}
		}
	}

	// --- Attempt 3: Fallback Built-in Comparison Logic ---
	// This is reached if no dunder methods handled the comparison, or they returned NotImplemented.
	lType, rType := left.Type(), right.Type()

	// Handle == and != for any types if dunders didn't cover it
	if op == constants.EqualsOperator {
		// If types are different and not numeric (which would have been handled),
		// Python generally considers them not equal unless one is None.
		if lType != rType && left != NULL && right != NULL && !IsNumeric(left) && !IsNumeric(right) {
			return FALSE
		}
		// For same types not handled above, or involving NULL, or numeric fallback
		// (numeric would be caught by IsNumeric check below if it's simple comparison)
		// Fallback to identity for unhandled same-types.
		// For specific types like String, List, Dict, Tuple, implement content comparison here.
		// If we are here, it means __eq__ was not found or returned NotImplemented.
	}
	if op == constants.NotEqualsOperator {
		// Similar logic to ==, then negate.
		// For simplicity, call CompareObjects with "==" and negate.
		// This avoids duplicating the complex fallback logic.
		// However, be careful of infinite recursion if __ne__ itself calls this.
		// Here, we assume __ne__ was already tried if it existed.
		eqResult := CompareObjects(constants.EqualsOperator, left, right, ctx) // Recursive call for "==" part
		if IsError(eqResult) {
			return eqResult
		}
		if eqResult == TRUE {
			return FALSE
		}
		if eqResult == FALSE {
			return TRUE
		}
		return NewError(constants.InternalError, constants.HELPER_COMPARE_OBJECTS_EQ_NON_BOOL_FALLBACK)
	}

	// Numeric Comparison (Handles Int/Float Mix) - Fallback if dunders absent
	if IsNumeric(left) && IsNumeric(right) {
		lF, rF := PromoteToFloats(left, right) // Convert both to float for comparison
		lVal := lF.Value
		rVal := rF.Value
		switch op {
		// == and != should have been handled by dunders or identity by now if it reached here
		// but can be a fallback if an object doesn't implement __eq__
		case constants.EqualsOperator:
			return NativeBoolToBooleanObject(lVal == rVal)
		case constants.NotEqualsOperator:
			return NativeBoolToBooleanObject(lVal != rVal)
		case constants.LessThanOp:
			return NativeBoolToBooleanObject(lVal < rVal)
		case constants.LessThanEqualsOp:
			return NativeBoolToBooleanObject(lVal <= rVal)
		case constants.GreaterThanOp:
			return NativeBoolToBooleanObject(lVal > rVal)
		case constants.GreaterThanEqualsOp:
			return NativeBoolToBooleanObject(lVal >= rVal)
		default:
			return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_NUMERIC_FALLBACK_UNSUPPORTED_OP, op)
		}
	}

	// String Comparison - Fallback
	if lType == STRING_OBJ && rType == STRING_OBJ {
		lValS := left.(*String).Value
		rValS := right.(*String).Value
		switch op {
		case constants.EqualsOperator:
			return NativeBoolToBooleanObject(lValS == rValS)
		case constants.NotEqualsOperator:
			return NativeBoolToBooleanObject(lValS != rValS)
		case constants.LessThanOp:
			return NativeBoolToBooleanObject(lValS < rValS)
		case constants.LessThanEqualsOp:
			return NativeBoolToBooleanObject(lValS <= rValS)
		case constants.GreaterThanOp:
			return NativeBoolToBooleanObject(lValS > rValS)
		case constants.GreaterThanEqualsOp:
			return NativeBoolToBooleanObject(lValS >= rValS)
		default:
			return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_STRING_FALLBACK_UNSUPPORTED_OP, op)
		}
	}

	// Boolean Comparison - Fallback
	if lType == BOOLEAN_OBJ && rType == BOOLEAN_OBJ {
		lValB := left.(*Boolean).Value
		rValB := right.(*Boolean).Value
		switch op {
		case constants.EqualsOperator:
			return NativeBoolToBooleanObject(lValB == rValB)
		case constants.NotEqualsOperator:
			return NativeBoolToBooleanObject(lValB != rValB)
		case constants.LessThanOp, constants.LessThanEqualsOp, constants.GreaterThanOp, constants.GreaterThanEqualsOp: // Ordering not supported for booleans directly
			return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_BOOL_UNSUPPORTED_OP, op)
		default:
			return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_BOOL_FALLBACK_UNSUPPORTED_OP, op)
		}
	}

	// List Comparison (Equality only for now as fallback)
	if op == constants.EqualsOperator && lType == LIST_OBJ && rType == LIST_OBJ {
		lList := left.(*List)
		rList := right.(*List)
		if len(lList.Elements) != len(rList.Elements) {
			return FALSE
		}
		for i, lElem := range lList.Elements {
			rElem := rList.Elements[i]
			eq := CompareObjects(constants.EqualsOperator, lElem, rElem, ctx)
			if IsError(eq) {
				return eq
			}
			if eq == FALSE {
				return FALSE
			}
		}
		return TRUE
	}
	if op == constants.NotEqualsOperator && lType == LIST_OBJ && rType == LIST_OBJ {
		eqResult := CompareObjects(constants.EqualsOperator, left, right, ctx)
		if IsError(eqResult) {
			return eqResult
		}
		if eqResult == TRUE {
			return FALSE
		}
		return TRUE
	}

	// Tuple Comparison (Equality only for now as fallback) - Assuming Tuple struct exists
	if op == constants.EqualsOperator && lType == TUPLE_OBJ && rType == TUPLE_OBJ {
		lTuple := left.(*Tuple) // Assuming type cast
		rTuple := right.(*Tuple)
		if len(lTuple.Elements) != len(rTuple.Elements) {
			return FALSE
		}
		for i, lElem := range lTuple.Elements {
			rElem := rTuple.Elements[i]
			eq := CompareObjects(constants.EqualsOperator, lElem, rElem, ctx)
			if IsError(eq) {
				return eq
			}
			if eq == FALSE {
				return FALSE
			}
		}
		return TRUE
	}
	if op == constants.NotEqualsOperator && lType == TUPLE_OBJ && rType == TUPLE_OBJ {
		eqResult := CompareObjects(constants.EqualsOperator, left, right, ctx)
		if IsError(eqResult) {
			return eqResult
		}
		if eqResult == TRUE {
			return FALSE
		}
		return TRUE
	}

	// Bytes Comparison (Equality only for now as fallback)
	if op == constants.EqualsOperator && lType == BYTES_OBJ && rType == BYTES_OBJ {
		lBytes := left.(*Bytes)
		rBytes := right.(*Bytes)
		// Use bytes.Equal for efficient and correct comparison.
		return NativeBoolToBooleanObject(bytes.Equal(lBytes.Value, rBytes.Value))
	}
	if op == constants.NotEqualsOperator && lType == BYTES_OBJ && rType == BYTES_OBJ {
		eqResult := CompareObjects(constants.EqualsOperator, left, right, ctx)
		if IsError(eqResult) {
			return eqResult
		}
		if eqResult == TRUE {
			return FALSE
		}
		return TRUE
	}

	// Comparison with None (Python 3: None is only equal to None)
	if left == NULL || right == NULL {
		isEqual := (left == NULL && right == NULL)
		switch op {
		case constants.EqualsOperator:
			return NativeBoolToBooleanObject(isEqual)
		case constants.NotEqualsOperator:
			return NativeBoolToBooleanObject(!isEqual)
		case constants.LessThanOp, constants.LessThanEqualsOp, constants.GreaterThanOp, constants.GreaterThanEqualsOp: // Ordering with None raises TypeError
			lTypeName := constants.BuiltinsNoneType
			if left != NULL {
				lTypeName = string(left.Type())
			}
			rTypeName := constants.BuiltinsNoneType
			if right != NULL {
				rTypeName = string(right.Type())
			}
			return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_NONE_UNSUPPORTED_OP, op, lTypeName, rTypeName)
		default:
			return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_NONE_FALLBACK_UNSUPPORTED_OP, op)
		}
	}

	// Final Fallback for == and !=: Identity check for any other unhandled types
	if op == constants.EqualsOperator {
		return NativeBoolToBooleanObject(left == right)
	}
	if op == constants.NotEqualsOperator {
		return NativeBoolToBooleanObject(left != right)
	}

	// If we reach here for ordering operators (<, <=, >, >=) and types are incompatible
	// or not handled above, it's a TypeError.
	return NewError(constants.TypeError, constants.HELPER_COMPARE_OBJECTS_GENERAL_TYPE_ERROR, op, lType, rType)
}

// CallGetAttr is a helper to get an attribute and return it.
// It's a simplified version of what the interpreter's dot operator would do.
// 'tokenForError' is used if the attribute doesn't exist.
func CallGetAttr(ctx ExecutionContext, obj Object, attrName string, tokenForError lexer.Token) (Object, bool) {
	getter, ok := obj.(AttributeGetter)
	if !ok {
		// For CallGetAttr, if it's not a getter, the attribute definitely isn't found via this mechanism.
		return nil, false
		// Or, if this should be an error:
		// return NewErrorWithLocation(tokenForError, "TypeError", constants.HELPER_CALL_GET_ATTR_TYPE_ERROR, obj.Type(), attrName), true
	}

	attrObj, found := getter.GetObjectAttribute(ctx, attrName)
	// 'found' indicates if GetObjectAttribute itself found something.
	// 'attrObj' could be an error object if GetObjectAttribute returned (Error, true).
	return attrObj, found
}

// IsCallable checks if a Pylearn object can be called.
func IsCallable(obj Object) bool {
	switch obj.(type) {
	case *Function, *Builtin, *BoundMethod, *Class: // Classes are callable (constructors)
		return true
	case *Instance:
		inst := obj.(*Instance)
		if inst.Class != nil { // Check if Class is not nil
			// Check if the instance's class has a __call__ method
			// This requires looking up __call__ without causing infinite recursion if __call__ calls IsCallable.
			// For simplicity, assume Methods map is directly accessible for this check.
			if _, hasCall := inst.Class.Methods[constants.DunderCall]; hasCall {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// NewErrorFromGoErr attempts to convert a standard Go error into a Pylearn Error.
// It preserves existing Pylearn Errors and wraps others.
func NewErrorFromGoErr(err error) *Error {
	if err == nil {
		return nil
	}
	// Check if it's already a Pylearn Error type
	if pyErr, ok := err.(*Error); ok {
		return pyErr
	}
	if pyStopIter, ok := err.(*StopIterationError); ok {
		// Wrap StopIteration in a RuntimeError if it escapes an await boundary.
		return NewError(constants.RuntimeError, constants.HELPER_NEW_ERROR_UNEXPECTED_STOP_ITERATION, pyStopIter)
	}

	// Otherwise, wrap the Go error message in a generic RuntimeError.
	return NewError(constants.RuntimeError, constants.ErrorFormat, err.Error())
}

func NewDict(kwargs map[string]Object) *Dict {
	d := &Dict{Pairs: make(map[HashKey]DictPair)}
	if kwargs == nil {
		return d
	}
	for k, v := range kwargs {
		keyObj := &String{Value: k}
		h, _ := keyObj.HashKey()
		d.Pairs[h] = DictPair{Key: keyObj, Value: v}
	}
	return d
}
