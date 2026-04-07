package builtins

import (
	"math"     // Go math package
	"math/big" // For potentially large pow results
	"strconv"  // For bin/oct/hex

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- abs() ---
// Accepts ExecutionContext
func pyAbsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsMathAbsArgCountError, len(args))
	}
	arg := args[0]

	// --- Check for __abs__ method on instances first ---
	if inst, ok := arg.(*object.Instance); ok && inst.Class != nil {
		if absMethodObj, methodOk := inst.Class.Methods[constants.DunderAbs]; methodOk {
			if absMethod, isFunc := absMethodObj.(*object.Function); isFunc { // <<< TYPE ASSERTION
				boundAbs := &object.BoundMethod{Instance: inst, Method: absMethod}
				return object.ApplyBoundMethod(ctx, boundAbs, []object.Object{}, object.NoToken)
			}
		}
	}

	// Default behavior for built-in types
	switch obj := arg.(type) {
	case *object.Integer:
		if obj.Value < 0 {
			return &object.Integer{Value: -obj.Value}
		}
		return obj
	case *object.Float:
		return &object.Float{Value: math.Abs(obj.Value)}
	case *object.Boolean: // abs(True) is 1, abs(False) is 0
		if obj.Value {
			return &object.Integer{Value: 1}
		}
		return &object.Integer{Value: 0}
	// TODO: Add Complex type later
	default:
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsMathBadOperandTypeAbs, arg.Type())
	}
}

// --- round() ---
// Accepts ExecutionContext (needed for potential __round__ call)
func pyRoundFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	ndigits := 0 // Python default for ndigits is 0 (rounds to integer)
	var numArg object.Object

	// --- Argument Parsing ---
	switch len(args) {
	case 1:
		numArg = args[0]
	case 2:
		numArg = args[0]
		ndigitsArg, ok := args[1].(*object.Integer)
		if !ok {
			// TODO: Allow None for ndigits? Python 3 does round(1.5, None) -> 2
			if args[1] == object.NULL {
				ndigits = 0 // Treat None like 0 for now
			} else {
				return object.NewError(constants.TypeError, constants.BuiltinsMathRoundNdigitsTypeError, args[1].Type())
			}
		} else {
			ndigits = int(ndigitsArg.Value) // Can be negative in Python
		}
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsMathRoundArgCountError, len(args))
	}
	// --- End Argument Parsing ---

	// --- Check for __round__ method ---
	if inst, ok := numArg.(*object.Instance); ok && inst.Class != nil {
		if roundMethodObj, methodOk := inst.Class.Methods[constants.DunderRound]; methodOk {
			if roundMethod, isFunc := roundMethodObj.(*object.Function); isFunc { // <<< TYPE ASSERTION
				boundRound := &object.BoundMethod{Instance: inst, Method: roundMethod}
				roundArgs := []object.Object{}
				if len(args) == 2 { // Pass ndigits if provided
					roundArgs = append(roundArgs, args[1]) // Pass original ndigits arg (Integer or Null)
				}
				// Use the generic object.ApplyBoundMethod helper
				return object.ApplyBoundMethod(ctx, boundRound, roundArgs, object.NoToken)
			}
		}
	}

	// --- Default round() for built-in types ---
	switch obj := numArg.(type) {
	case *object.Integer:
		// TODO: Handle ndigits for Integer (Python 3 returns int if ndigits <= 0, float otherwise?)
		// For simplicity, return Integer if ndigits is 0 or negative
		if ndigits >= 0 {
			// This doesn't really round, but matches Python 3 behavior for ndigits >= 0
			// E.g., round(123, 0) -> 123, round(123, 1) -> 123, round(123, -1) -> 120
			// Implementing negative ndigits rounding for integers is complex.
			// Let's return the integer itself for ndigits >= 0 for now.
			if ndigits > 0 {
				// Python 3 behavior for int and positive ndigits is just the int itself
				return obj
			}
			// Fall through for ndigits <= 0
		}
		// Handle ndigits < 0 (rounding to powers of 10)
		if ndigits < 0 {
			power := math.Pow(10, float64(-ndigits))
			// Round to nearest even for .5 cases
			roundedVal := math.RoundToEven(float64(obj.Value)/power) * power
			return &object.Integer{Value: int64(roundedVal)}
		}
		return obj // Default for ndigits = 0

	case *object.Float:
		// Use math.RoundToEven for Python's rounding behavior
		valueToRound := obj.Value
		var rounded float64

		if ndigits == 0 {
			rounded = math.RoundToEven(valueToRound)
			// Python's round() without ndigits (or 0) returns int
			return &object.Integer{Value: int64(rounded)}
		} else {
			// Handle positive/negative ndigits for float
			power := math.Pow(10, float64(ndigits))
			// Round to nearest even after shifting decimal point
			rounded = math.RoundToEven(valueToRound*power) / power
			// round() with ndigits returns a float
			return &object.Float{Value: rounded}
		}

	default:
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsMathRoundTypeNotDefineError, numArg.Type())
	}
}

// --- pow() ---
// Accepts ExecutionContext (needed for potential __pow__ call)
func pyPowFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var base, exp, mod object.Object
	mod = nil // Sentinel for no modulo argument

	// --- Argument Parsing ---
	switch len(args) {
	case 2:
		base = args[0]
		exp = args[1]
	case 3:
		base = args[0]
		exp = args[1]
		mod = args[2]
		if mod == object.NULL { // Allow None as modulo argument (equivalent to 2 args)
			mod = nil
		} else if _, ok := mod.(*object.Integer); !ok {
			return object.NewError(constants.TypeError, constants.BuiltinsMathPow3rdArgTypeError, mod.Type())
		}
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsMathPowArgCountError, len(args))
	}
	// --- End Argument Parsing ---

	// --- Check for __pow__ method ---
	// Python's __pow__ handles 2 or 3 arguments
	if inst, ok := base.(*object.Instance); ok && inst.Class != nil {
		if powMethodObj, methodOk := inst.Class.Methods[constants.DunderPow]; methodOk {
			if powMethod, isFunc := powMethodObj.(*object.Function); isFunc { // <<< TYPE ASSERTION
				boundPow := &object.BoundMethod{Instance: inst, Method: powMethod}
				powArgs := []object.Object{exp} // Always pass exponent
				if mod != nil {
					powArgs = append(powArgs, mod) // Pass modulo if present
				}
				result := object.ApplyBoundMethod(ctx, boundPow, powArgs, object.NoToken)
				// TODO: Handle NotImplemented later to try reflected __rpow__
				return result // Return result or error from __pow__
			}
		}
	}
	// --- Check for __rpow__ (reflected) ---
	// TODO: If __pow__ returned NotImplemented or wasn't present, try exp.__rpow__(base)

	// --- Default pow() for built-in types ---
	// Handle 3-argument pow (integer only in Python 3)
	if mod != nil {
		baseInt, okB := base.(*object.Integer)
		expInt, okE := exp.(*object.Integer)
		modInt := mod.(*object.Integer) // Already checked type
		if !(okB && okE) {
			return object.NewError(constants.TypeError, constants.BuiltinsMathPowUnsupportedOperandTypeModulo, base.Type(), exp.Type())
		}
		if expInt.Value < 0 {
			return object.NewError(constants.ValueError, constants.BuiltinsMathPowExponentNegativeError)
		}
		if modInt.Value == 0 {
			return object.NewError(constants.ValueError, constants.BuiltinsMathPow3rdArgZeroError)
		}

		// Use math/big for modular exponentiation
		bigBase := big.NewInt(baseInt.Value)
		bigExp := big.NewInt(expInt.Value)
		bigMod := big.NewInt(modInt.Value)
		result := new(big.Int)

		result.Exp(bigBase, bigExp, bigMod) // result = (base ** exp) % mod

		// Result of 3-arg pow is always int (or BigInt if needed)
		if result.IsInt64() {
			return &object.Integer{Value: result.Int64()}
		}
		return object.NewError(constants.OverflowError, constants.BuiltinsMathPowIntResultTooLarge) // TODO: BigInt support
	}

	// Handle 2-argument pow (numeric types)
	switch b := base.(type) {
	case *object.Integer:
		switch e := exp.(type) {
		case *object.Integer:
			// Use math/big for potentially large results
			bigBase := big.NewInt(b.Value)
			bigExp := big.NewInt(e.Value)
			if e.Value < 0 { // Negative exponent -> float result
				if b.Value == 0 {
					return object.NewError(constants.ZeroDivisionError, constants.BuiltinsMathZeroRaisedNegativePower)
				}
				floatResult := math.Pow(float64(b.Value), float64(e.Value))
				return &object.Float{Value: floatResult}
			}
			if b.Value == 0 && e.Value == 0 {
				return &object.Integer{Value: 1}
			} // 0**0 = 1

			result := new(big.Int)
			result.Exp(bigBase, bigExp, nil) // result = base ** exp
			if result.IsInt64() {
				return &object.Integer{Value: result.Int64()}
			}
			return object.NewError(constants.OverflowError, constants.BuiltinsMathIntPowerResultTooLarge) // TODO: BigInt support
		case *object.Float:
			floatResult := math.Pow(float64(b.Value), e.Value)
			return &object.Float{Value: floatResult}
		default:
			goto powTypeErrorDefault
		}
	case *object.Float:
		switch e := exp.(type) {
		case *object.Integer:
			floatResult := math.Pow(b.Value, float64(e.Value))
			return &object.Float{Value: floatResult}
		case *object.Float:
			floatResult := math.Pow(b.Value, e.Value)
			return &object.Float{Value: floatResult}
		default:
			goto powTypeErrorDefault
		}
	// TODO: Add complex numbers
	default:
		goto powTypeErrorDefault
	}

powTypeErrorDefault:
	// Use object.NewError
	return object.NewError(constants.TypeError, constants.BuiltinsMathPowUnsupportedOperandType, base.Type(), exp.Type())
}

// Helper for Python-style floor division and modulo
// Keep this helper as is
func floorDivMod(a, b int64) (q, r int64) {
	q = a / b
	r = a % b
	if (a < 0) != (b < 0) && r != 0 { // Check if signs differ and remainder is non-zero
		q -= 1
		r += b
	}
	return
}

// --- divmod() ---
// Accepts ExecutionContext (needed for potential __divmod__)
func pyDivmodFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsMathDivmodArgCountError, len(args))
	}
	num := args[0]
	den := args[1]

	// --- Check for __divmod__ ---
	if inst, ok := num.(*object.Instance); ok && inst.Class != nil {
		if dmMethodObj, methodOk := inst.Class.Methods[constants.DunderDivmod]; methodOk {
			if dmMethod, isFunc := dmMethodObj.(*object.Function); isFunc { // <<< TYPE ASSERTION
				boundDM := &object.BoundMethod{Instance: inst, Method: dmMethod}
				result := object.ApplyBoundMethod(ctx, boundDM, []object.Object{den}, object.NoToken)
				return result
			}
		}
	}
	// --- Check for __rdivmod__ ---
	// TODO: Implement reflected check if needed

	// Default behavior for built-ins
	switch n := num.(type) {
	case *object.Integer:
		switch d := den.(type) {
		case *object.Integer:
			if d.Value == 0 {
				return object.NewError(constants.ZeroDivisionError, constants.BuiltinsMathIntegerDivOrModuloByZero)
			}
			quotient, remainder := floorDivMod(n.Value, d.Value)
			// Return Tuple object
			return &object.Tuple{Elements: []object.Object{
				&object.Integer{Value: quotient},
				&object.Integer{Value: remainder},
			}}
		// TODO: Handle float denominator? (Requires math.Modf, careful type handling)
		default:
			goto divmodTypeErrorDefault
		}
	// TODO: Handle float numerator?
	default:
		goto divmodTypeErrorDefault
	}

divmodTypeErrorDefault:
	// Use object.NewError
	return object.NewError(constants.TypeError, constants.BuiltinsMathDivmodUnsupportedOperandType, num.Type(), den.Type())
}

// --- sum() ---
// Accepts ExecutionContext (needed for iteration and potential addition dunders)
func pySumFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var iterableArg object.Object
	var start object.Object = &object.Integer{Value: 0} // Default start is 0

	// Argument Parsing
	switch len(args) {
	case 1:
		iterableArg = args[0]
	case 2:
		iterableArg = args[0]
		start = args[1] // User-provided start value
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsMathSumArgCountError, len(args))
	}

	// Get iterator using the context-aware helper
	iterator, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		return errObj // Propagate TypeError if not iterable
	}

	currentSum := start    // Initialize with start value
	var item object.Object // <<< DECLARE 'item' OUTSIDE THE LOOP

	// Iterate and add
	for {
		// Assign the result of iterator.Next() to the outer 'item' variable
		var stop bool
		item, stop = iterator.Next() // Use existing 'item' variable
		if stop {
			break
		} // End of iteration

		if object.IsError(item) {
			return item
		} // Propagate error from iterator itself

		// Perform addition: currentSum + item
		// TODO: Add dunder method checks (__add__, __radd__) using ctx

		// --- Basic Built-in Addition (fallback) ---
		var addResult object.Object
		if ai, okSum := currentSum.(*object.Integer); okSum {
			if bi, okItem := item.(*object.Integer); okItem {
				addResult = &object.Integer{Value: ai.Value + bi.Value}
			} else if bf, okItem := item.(*object.Float); okItem {
				addResult = &object.Float{Value: float64(ai.Value) + bf.Value}
			} else {
				goto sumTypeError
			} // Jump if item type mismatch
		} else if af, okSum := currentSum.(*object.Float); okSum {
			if bi, okItem := item.(*object.Integer); okItem {
				addResult = &object.Float{Value: af.Value + float64(bi.Value)}
			} else if bf, okItem := item.(*object.Float); okItem {
				addResult = &object.Float{Value: af.Value + bf.Value}
			} else {
				goto sumTypeError
			} // Jump if item type mismatch
		} else {
			// currentSum is not Integer or Float
			// TODO: Add support for summing other types like lists if needed (via __add__)
			goto sumTypeError // Jump if currentSum type is wrong
		}
		// --- End Basic Addition ---

		if object.IsError(addResult) {
			return addResult
		} // Propagate error from addition
		currentSum = addResult // Update sum
	} // End for loop

	return currentSum // Return the final sum

	// Label for type errors during addition
sumTypeError:
	// Now 'item' is accessible here because it was declared outside the loop
	return object.NewError(constants.TypeError, constants.BuiltinsMathSumUnsupportedOperandAdd, currentSum.Type(), item.Type())
}

// --- min(), max() ---
// Helper function getNumericValue remains the same

// --- min(), max() ---
// Accepts ExecutionContext (needed for iteration)
// Handles min(iterable) or min(arg1, arg2, ...)
// Handles max(iterable) or max(arg1, arg2, ...)
// Currently supports comparing numbers (int/float) OR strings, but not mixed types.
// TODO: Add support for __lt__ dunder method comparison via context.
func pyMinMaxFn(ctx object.ExecutionContext, name string, isMin bool, args ...object.Object) object.Object {
	if len(args) == 0 {
		return object.NewError(constants.TypeError, constants.BuiltinsMathMinMaxArgCountError, name)
	}

	var itemsToCompare []object.Object

	// Determine if called with a single iterable or multiple arguments
	if len(args) == 1 {
		iterableArg := args[0]
		// Use context-aware iterator helper
		iterator, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
		if errObj != nil {
			// Propagate TypeError: object is not iterable
			return errObj
		}

		items := []object.Object{}
		for {
			item, stop := iterator.Next()
			if stop {
				break
			}
			if object.IsError(item) {
				return item
			} // Propagate error from iterator
			items = append(items, item)
		}
		itemsToCompare = items
	} else {
		// Multiple arguments provided directly
		itemsToCompare = args
	}

	// Check if the resulting list of items is empty
	if len(itemsToCompare) == 0 {
		return object.NewError(constants.ValueError, constants.BuiltinsMathMinMaxEmptySequence, name)
	}

	// --- Comparison Logic ---
	var bestItem object.Object = nil
	// We need to track the best value based on type for built-in comparison
	var bestValFloat float64
	var bestValStr string
	var isComparingNumbers bool = false // Assume false initially

	for i, currentItem := range itemsToCompare {

		if i == 0 {
			// Initialize based on the first item
			bestItem = currentItem
			switch currentItem.(type) {
			case *object.Integer, *object.Float:
				isComparingNumbers = true
				val, _ := getNumericValue(currentItem) // Error check not needed after type assert
				bestValFloat = val
			case *object.String:
				isComparingNumbers = false
				bestValStr = currentItem.(*object.String).Value
			default:
				// TODO: Try using __lt__ for comparison here if types are different but comparable
				return object.NewError(constants.TypeError, constants.BuiltinsMathMinMaxNotSupportedBetweenNumber, constants.GreaterThanOp, currentItem.Type(), currentItem.Type(), name)
			}
			continue
		}

		// --- Perform Comparison ---
		// TODO: Prioritize using __lt__ (or __gt__ for max) via ctx.Execute if available on objects.
		// Example pseudo-code:
		// lt_result := ctx.Execute(lookup(bestItem, "__lt__"), currentItem) // Needs lookup helper
		// if !object.IsError(lt_result) { ... use boolean result ... } else { /* fallback or error */ }

		// Fallback to built-in comparison for numbers/strings:
		var shouldUpdate bool = false // Assume current is not better initially

		if isComparingNumbers {
			currentValFloat, ok := getNumericValue(currentItem)
			if !ok {
				opStr := constants.GreaterThanOp
				if isMin {
					opStr = constants.LessThanOp
				}
				return object.NewError(constants.TypeError, constants.BuiltinsMathMinMaxNotSupportedBetweenNumber, opStr, currentItem.Type())
			}
			if isMin {
				shouldUpdate = currentValFloat < bestValFloat
			} else {
				shouldUpdate = currentValFloat > bestValFloat
			}
		} else { // Comparing strings (assuming based on first item)
			currentValStr, ok := currentItem.(*object.String)
			if !ok {
				opStr := constants.GreaterThanOp
				if isMin {
					opStr = constants.LessThanOp
				}
				return object.NewError(constants.TypeError, constants.BuiltinsMathMinMaxNotSupportedBetweenStr, opStr, currentItem.Type())
			}
			if isMin {
				shouldUpdate = currentValStr.Value < bestValStr
			} else {
				shouldUpdate = currentValStr.Value > bestValStr
			}
		}

		if shouldUpdate {
			bestItem = currentItem
			// Update tracked best value
			if isComparingNumbers {
				bestValFloat, _ = getNumericValue(currentItem)
			} else {
				bestValStr = currentItem.(*object.String).Value
			}
		}
		// --- End Perform Comparison ---

	} // End for loop

	return bestItem // Return the best item found
}

// --- Wrapper functions for min/max ---
// Need to accept context and pass it down
func pyMinFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return pyMinMaxFn(ctx, constants.BuiltinsMinFuncName, true, args...)
}
func pyMaxFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return pyMinMaxFn(ctx, constants.BuiltinsMaxFuncName, false, args...)
}

// --- bin(), oct(), hex() ---
// Accept ExecutionContext (unused but required)
func formatBasedInt(ctx object.ExecutionContext, prefix string, base int, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsMathFormatBasedIntArgCountError, prefix[1:], len(args))
	}
	arg := args[0]
	// TODO: Check __index__ method via context

	intObj, ok := arg.(*object.Integer)
	if !ok {
		// TODO: Check if arg has __index__ method using ctx and call it?
		return object.NewError(constants.TypeError, constants.BuiltinsMathObjectCannotBeInterpretedInteger, arg.Type())
	}

	val := intObj.Value
	if val == 0 {
		return &object.String{Value: prefix + constants.ZeroString}
	}

	sign := constants.EmptyString
	absValue := val
	if val < 0 {
		sign = constants.MinusSign
		absValue = -val
	}
	formatted := strconv.FormatInt(absValue, base)
	return &object.String{Value: sign + prefix + formatted}
}

func pyBinFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return formatBasedInt(ctx, constants.BuiltinsBinPrefix, 2, args...)
}
func pyOctFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return formatBasedInt(ctx, constants.BuiltinsOctPrefix, 8, args...)
}
func pyHexFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return formatBasedInt(ctx, constants.BuiltinsHexPrefix, 16, args...)
}

// --- hash() ---
// Accepts ExecutionContext (potentially needed for __hash__)
func pyHashFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsMathHashArgCountError, len(args))
	}
	arg := args[0]

	// --- Check for __hash__ method ---
	// TODO: Implement __hash__ lookup and call via context if arg is Instance

	// Fallback: Check if object implements the Go Hashable interface
	hashable, ok := arg.(object.Hashable)
	if !ok {
		// If no __hash__ and not Go Hashable
		return object.NewError(constants.TypeError, constants.BuiltinsMathUnhashableType, arg.Type())
	}

	// Use the Go Hashable interface method
	hashKeyVal, err := hashable.HashKey()
	if err != nil {
		// Error occurred during HashKey calculation (e.g., tuple contained list)
		// err should be a Go error containing type info
		return object.NewError(constants.TypeError, constants.BuiltinsMathFailedToHashObject, err)
	}

	// Return the hash value as a Pylearn Integer
	// Python's hash() can return negative; HashKey uses uint64. Cast carefully.
	return &object.Integer{Value: int64(hashKeyVal.Value)}
}

// --- min(), max() ---
// Simplified: Only numeric types for now, requires iterable handling
// --- Helper function to get float64 value from Integer or Float ---
// Used for numeric comparisons in min/max
func getNumericValue(obj object.Object) (float64, bool) {
	if i, ok := obj.(*object.Integer); ok {
		return float64(i.Value), true
	}
	if f, ok := obj.(*object.Float); ok {
		return f.Value, true
	}
	return 0, false // Not a numeric type
}

// --- Registration ---
// Ensure functions match the required signature
func init() {
	registerBuiltin(constants.BuiltinsAbsFuncName, &object.Builtin{Fn: pyAbsFn})
	registerBuiltin(constants.BuiltinsRoundFuncName, &object.Builtin{Fn: pyRoundFn})
	registerBuiltin(constants.BuiltinsPowFuncName, &object.Builtin{Fn: pyPowFn})
	registerBuiltin(constants.BuiltinsDivmodFuncName, &object.Builtin{Fn: pyDivmodFn})
	registerBuiltin(constants.BuiltinsSumFuncName, &object.Builtin{Fn: pySumFn})
	registerBuiltin(constants.BuiltinsMinFuncName, &object.Builtin{Fn: pyMinFn})
	registerBuiltin(constants.BuiltinsMaxFuncName, &object.Builtin{Fn: pyMaxFn})
	registerBuiltin(constants.BuiltinsBinFuncName, &object.Builtin{Fn: pyBinFn})
	registerBuiltin(constants.BuiltinsOctFuncName, &object.Builtin{Fn: pyOctFn})
	registerBuiltin(constants.BuiltinsHexFuncName, &object.Builtin{Fn: pyHexFn})
	registerBuiltin(constants.BuiltinsHashFuncName, &object.Builtin{Fn: pyHashFn})
}
