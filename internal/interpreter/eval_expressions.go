package interpreter

import (
	"math"
	"math/big"
	"strings"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
)

// All functions now correctly receive and pass the *InterpreterContext.

func evalExpressions(exps []ast.Expression, ctx *InterpreterContext) []object.Object {
	result := make([]object.Object, len(exps))
	for i, e := range exps {
		evaluated := Eval(e, ctx) // Pass ctx
		if object.IsError(evaluated) {
			return []object.Object{evaluated}
		}
		result[i] = evaluated
	}
	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	return object.NewErrorWithLocation(node.Token, constants.NameError, constants.EvalExpressionsNameNotDefined, node.Value)
}

func evalPrefixExpression(node *ast.PrefixExpression, ctx *InterpreterContext) object.Object {
	right := Eval(node.Right, ctx) // Pass ctx
	if object.IsError(right) {
		return right
	}
	op, token := node.Operator, node.Token
	dunderName, foundDunder := constants.PrefixOperatorToDunder[op]
	if foundDunder {
		if inst, ok := right.(*object.Instance); ok && inst.Class != nil {
			if methodObj, methodOk := inst.Class.Methods[dunderName]; methodOk {
				// <<< THIS IS THE FIX >>>
				// Type-assert that the object we found is actually a function.
				if method, isFunc := methodObj.(*object.Function); isFunc {
					boundMethod := &object.BoundMethod{Instance: inst, Method: method}
					result := object.ApplyBoundMethod(ctx, boundMethod, []object.Object{}, token)

					if dunderName == constants.DunderBool {
						if object.IsError(result) {
							return result
						}
						truthy, err := object.IsTruthy(ctx, result)
						if err != nil {
							if pyErr, isPyErr := err.(object.Object); isPyErr && object.IsError(pyErr) {
								return pyErr
							}
							return object.NewError(constants.RuntimeError, constants.EvalExpressionsPropagatedFromIsTruthy, err)
						}
						return object.NativeBoolToBooleanObject(!truthy)
					}
					// For __neg__, __pos__, just return the result from the dunder method.
					return result
				}
				// If `methodObj` was found but wasn't a function (e.g., it was an integer),
				// we fall through to the default logic below, which is correct Python behavior.
			}
		}
	}
	switch op {
	case constants.NotKeyword, constants.BangOperator:
		truthy, err := object.IsTruthy(ctx, right) // Pass ctx
		if err != nil {
			if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.EvalExpressionsPropagatedFromIsTruthy, err)
		}
		return object.NativeBoolToBooleanObject(!truthy)
	case constants.MinusSign:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Integer{Value: -r.Value}
		case *object.Float:
			return &object.Float{Value: -r.Value}
		default:
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsBadOperandTypeUnaryMinus, right.Type())
		}
	case constants.PlusSign:
		switch right.Type() {
		case object.INTEGER_OBJ, object.FLOAT_OBJ:
			return right
		default:
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsBadOperandTypeUnaryPlus, right.Type())
		}
	case "~":
		switch r := right.(type) {
		case *object.Integer:
			return &object.Integer{Value: ^r.Value}
		default:
			return object.NewErrorWithLocation(token, constants.TypeError, "bad operand type for unary ~: '%s'", right.Type())
		}
	default:
		return object.NewErrorWithLocation(token, constants.SyntaxError, constants.EvalExpressionsUnknownPrefixOperator, op)
	}
}

func evalInfixExpression(node *ast.InfixExpression, ctx *InterpreterContext) object.Object {
	if node.Operator == constants.AssignOperator {
		return evalAssignExpression(node, ctx)
	}
	left := Eval(node.Left, ctx)
	if object.IsError(left) {
		return left
	}
	// <<< ADD THIS BLOCK >>>
	if _, isYield := left.(*object.YieldValue); isYield {
		return left
	}
	// <<< END OF FIX >>>
	op, token := node.Operator, node.Token
	if op == constants.AndKeyword {
		truthy, err := object.IsTruthy(ctx, left)
		if err != nil {
			if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.EvalExpressionsPropagatedFromIsTruthy, err)
		}
		if !truthy {
			return left
		}
		return Eval(node.Right, ctx)
	}
	if op == constants.OrKeyword {
		truthy, err := object.IsTruthy(ctx, left)
		if err != nil {
			if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.EvalExpressionsPropagatedFromIsTruthy, err)
		}
		if truthy {
			return left
		}
		return Eval(node.Right, ctx)
	}
	right := Eval(node.Right, ctx)
	if object.IsError(right) {
		return right
	}
	// Step 1: Handle `is` and `is not` which don't use dunder methods and have the highest precedence.
	switch op {
	case constants.IsKeyword:
		return object.NativeBoolToBooleanObject(left == right)
	case constants.IsNotKeyword:
		return object.NativeBoolToBooleanObject(left != right)
	case constants.InKeyword:
		return evalInOperator(left, right, token, ctx)
	case constants.NotInKeyword:
		inResult := evalInOperator(left, right, token, ctx)
		if object.IsError(inResult) {
			return inResult
		}
		return object.NativeBoolToBooleanObject(inResult != object.TRUE)
	}

	// Step 2: Try the standard dunder method on the left operand (e.g., left.__mul__(right))
	dunderName, foundDunder := constants.InfixOperatorToDunder[op]
	if foundDunder {
		if leftGetter, hasGetAttr := left.(object.AttributeGetter); hasGetAttr {
			methodObj, found := leftGetter.GetObjectAttribute(ctx, dunderName)
			if found {
				if object.IsError(methodObj) {
					return methodObj
				}
				var result object.Object
				if methodBuiltin, isBuiltin := methodObj.(*object.Builtin); isBuiltin {
					result = methodBuiltin.Fn(ctx, right)
				} else if boundMeth, isBound := methodObj.(*object.BoundMethod); isBound {
					result = object.ApplyBoundMethod(ctx, boundMeth, []object.Object{right}, token)
				} else {
					return object.NewError(constants.TypeError, constants.EvalExpressionsStrAttrNotCallable, dunderName)
				}
				if result != object.NOT_IMPLEMENTED {
					return result
				}
			}
		}
	}
	// Step 3: If the standard dunder failed or returned NotImplemented, try the reflected dunder on the right operand.
	// This is the key part of the fix that allows `3 * my_array` to work by calling `my_array.__rmul__(3)`.
	reflectedDunderName, rDunderFound := constants.InfixOperatorToRDunder[op]
	if rDunderFound {
		if getter, ok := right.(object.AttributeGetter); ok {
			method, found := getter.GetObjectAttribute(ctx, reflectedDunderName)
			if found && method != nil && !object.IsError(method) {
				// For r-dunders, the original left operand is passed as the argument
				result := ctx.Execute(method, left)
				if result != object.NOT_IMPLEMENTED {
					return result
				}
			}
		}
	}
	if op == constants.EqOperator || op == constants.NotEqOperator || op == constants.LessThanOp || op == constants.LessThanOrEqualToOp || op == constants.GreaterThanOp || op == constants.GreaterThanOrEqualToOp {
		return object.CompareObjects(op, left, right, ctx)
	}
	// Step 4: If dunder methods failed, fall back to built-in behavior for primitive types.
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right, token)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(op, left, right, token)
	case object.IsNumeric(left) && object.IsNumeric(right):
		lF, rF := object.PromoteToFloats(left, right)
		return evalFloatInfixExpression(op, lF, rF, token)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(op, left, right, token)
	case left.Type() == object.BYTES_OBJ && right.Type() == object.BYTES_OBJ:
		return evalBytesInfixExpression(op, left, right, token)
	default:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsTypeErrorUnsupportedOperandType, op, left.Type(), right.Type())
	}
}

func evalInOperator(left, right object.Object, token lexer.Token, ctx *InterpreterContext) object.Object {
	if getter, hasGetAttr := right.(object.AttributeGetter); hasGetAttr {
		containsMethodObj, foundMethod := getter.GetObjectAttribute(ctx, constants.DunderContains)
		if foundMethod {
			if object.IsError(containsMethodObj) {
				return containsMethodObj
			}
			var result object.Object
			if methodBuiltin, isBuiltin := containsMethodObj.(*object.Builtin); isBuiltin {
				result = methodBuiltin.Fn(ctx, left)
			} else if boundInstanceMethod, isBound := containsMethodObj.(*object.BoundMethod); isBound {
				result = object.ApplyBoundMethod(ctx, boundInstanceMethod, []object.Object{left}, token)
			} else {
				return object.NewError(constants.TypeError, constants.EvalExpressionsStrAttrNotCallable, containsMethodObj.Type())
			}
			if object.IsError(result) {
				return result
			}
			if boolResult, isBool := result.(*object.Boolean); isBool {
				return boolResult
			}
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsStrReturnedNonBool, result.Type())
		}
	}
	switch rightTyped := right.(type) {
	case *object.String:
		itemStr, ok := left.(*object.String)
		if !ok {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsInStringRequiresString, left.Type())
		}
		return object.NativeBoolToBooleanObject(strings.Contains(rightTyped.Value, itemStr.Value))
	case *object.List:
		for _, elem := range rightTyped.Elements {
			if eq := object.CompareObjects(constants.EqOperator, left, elem, ctx); object.IsError(eq) {
				return eq
			} else if eq == object.TRUE {
				return object.TRUE
			}
		}
		return object.FALSE
	case *object.Tuple:
		for _, elem := range rightTyped.Elements {
			if eq := object.CompareObjects(constants.EqOperator, left, elem, ctx); object.IsError(eq) {
				return eq
			} else if eq == object.TRUE {
				return object.TRUE
			}
		}
		return object.FALSE
	case *object.Dict:
		hashableKey, ok := left.(object.Hashable)
		if !ok {
			return object.FALSE
		}
		dictMapKey, err := hashableKey.HashKey()
		if err != nil {
			return object.FALSE
		}
		_, found := rightTyped.Pairs[dictMapKey]
		return object.NativeBoolToBooleanObject(found)
	default:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsArgNotIterable, right.Type())
	}
}

// evalAssignExpression and other helpers in the file also need their signatures updated.
func evalAssignExpression(node *ast.InfixExpression, ctx *InterpreterContext) object.Object {
	val := Eval(node.Right, ctx)
	if object.IsError(val) {
		return val
	}
	// If the RHS was a yield expression that is pausing execution,
	// we must propagate the YieldValue signal upwards *without* performing
	// the assignment. The assignment will occur when the generator is resumed.
	if _, ok := val.(*object.YieldValue); ok {
		return val
	}
	token := node.Token
	switch targetNode := node.Left.(type) {
	case *ast.Identifier:
		name := targetNode.Value
		if _, ok := ctx.Env.Update(name, val); !ok {
			ctx.Env.Set(name, val)
		}
		return object.NULL
	case *ast.IndexExpression:
		collection := Eval(targetNode.Left, ctx)
		if object.IsError(collection) {
			return collection
		}
		index := Eval(targetNode.Index, ctx)
		if object.IsError(index) {
			return index
		}
		return evalIndexAssignment(collection, index, val, token, ctx)
	case *ast.DotExpression:
		obj := Eval(targetNode.Left, ctx)
		if object.IsError(obj) {
			return obj
		}
		attrName := targetNode.Identifier.Value
		tokenAttr := targetNode.Identifier.Token
		instance, ok := obj.(*object.Instance)
		if !ok {
			return object.NewErrorWithLocation(tokenAttr, constants.AttributeError, constants.EvalExpressionsTypeErrorCannotSetAttribute, obj.Type(), attrName)
		}
		if instance.Env == nil {
			instance.Env = object.NewEnvironment()
		}
		instance.Env.Set(attrName, val)
		return object.NULL
	default:
		return object.NewErrorWithLocation(token, constants.SyntaxError, constants.EvalExpressionsSyntaxErrorCannotAssignTo, node.Left.String())
	}
}

func evalCallExpression(node *ast.CallExpression, ctx *InterpreterContext) object.Object {
	callable := Eval(node.Function, ctx)
	if object.IsError(callable) {
		return callable
	}
	finalPositionalArgs := []object.Object{}
	finalCallsiteKwargs := make(map[string]object.Object)
	providedKeywordArgs := make(map[string]bool)
	for _, argNode := range node.Arguments {
		switch an := argNode.(type) {
		case *ast.KeywordArgument:
			if _, exists := providedKeywordArgs[an.Name.Value]; exists {
				return object.NewErrorWithLocation(an.Token, constants.SyntaxError, constants.EvalExpressionsSyntaxErrorKeywordArgumentRepeated, an.Name.Value)
			}
			val := Eval(an.Value, ctx)
			if object.IsError(val) {
				return val
			}
			finalCallsiteKwargs[an.Name.Value] = val
			providedKeywordArgs[an.Name.Value] = true
		case *ast.StarredArgument:
			val := Eval(an.Value, ctx)
			if object.IsError(val) {
				return val
			}
			if an.IsKwUnpack {
				dictVal, isDict := val.(*object.Dict)
				if !isDict {
					return object.NewErrorWithLocation(an.Token, constants.TypeError, constants.EvalExpressionsTypeErrorArgumentAfterStarStar, val.Type())
				}
				for _, pair := range dictVal.Pairs {
					keyStr, okStrKey := pair.Key.(*object.String)
					if !okStrKey {
						return object.NewErrorWithLocation(an.Token, constants.TypeError, constants.EvalExpressionsTypeErrorKeywordsMustBeStrings, val.Type())
					}
					if _, exists := providedKeywordArgs[keyStr.Value]; exists {
						return object.NewErrorWithLocation(an.Token, constants.TypeError, constants.EvalExpressionsTypeErrorMultipleValuesKwarg, callable.Type(), keyStr.Value)
					}
					finalCallsiteKwargs[keyStr.Value] = pair.Value
					providedKeywordArgs[keyStr.Value] = true
				}
			} else {
				iterator, errObj := object.GetObjectIterator(ctx, val, an.Token)
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
					finalPositionalArgs = append(finalPositionalArgs, item)
				}
			}
		default:
			evaluatedArg := Eval(an, ctx)
			if object.IsError(evaluatedArg) {
				return evaluatedArg
			}
			if len(providedKeywordArgs) > 0 {
				return object.NewErrorWithLocation(node.Token, constants.SyntaxError, constants.EvalExpressionsPositionalArgFollowsKeywordArg)
			}
			finalPositionalArgs = append(finalPositionalArgs, evaluatedArg)
		}
	}

	return applyFunctionOrClass(ctx, callable, finalPositionalArgs, finalCallsiteKwargs, node.Token)
}

// ... (All other eval functions in the file must be updated similarly) ...
// The rest of the file...
func evalIntegerInfixExpression(op string, left, right object.Object, token lexer.Token) object.Object {
	lVal, rVal := left.(*object.Integer).Value, right.(*object.Integer).Value
	switch op {
	case constants.PlusOperator:
		return &object.Integer{Value: lVal + rVal}
	case constants.MinusOperator:
		return &object.Integer{Value: lVal - rVal}
	case constants.AsteriskOperator:
		return &object.Integer{Value: lVal * rVal}
	case constants.PowOperator: // <<< ADD THIS CASE
		// Python's int ** int can result in a float if the exponent is negative.
		if rVal < 0 {
			if lVal == 0 {
				return object.NewErrorWithLocation(token, constants.ZeroDivisionError, constants.EvalExpressionsZeroCannotBeRaisedNegativePower)
			}
			result := math.Pow(float64(lVal), float64(rVal))
			return &object.Float{Value: result}
		}
		// Use math/big for large integer powers to prevent overflow.
		base := big.NewInt(lVal)
		exponent := big.NewInt(rVal)
		result := new(big.Int)
		result.Exp(base, exponent, nil) // result = base^exponent

		// If the result fits in int64, return an Integer, otherwise it's an overflow
		// until you implement a BigInt object type.
		if result.IsInt64() {
			return &object.Integer{Value: result.Int64()}
		}
		return object.NewErrorWithLocation(token, constants.OverflowError, constants.EvalExpressionsIntegerPowerResultTooLarge)

	case constants.SlashOperator:
		if rVal == 0 {
			return object.NewErrorWithToken(constants.EmptyString, token, constants.ZeroDivisionError, constants.EvalExpressionsDivisionByZero)
		}
		return &object.Float{Value: float64(lVal) / float64(rVal)}
	case "//":
		if rVal == 0 {
			return object.NewErrorWithToken(constants.EmptyString, token, constants.ZeroDivisionError, constants.EvalExpressionsIntegerModuloByZero)
		}
		// Go's `/` on integers truncates towards zero. Python's `//` floors.
		// We need a helper to correctly implement floor division.
		return &object.Integer{Value: floorDiv(lVal, rVal)}
	case constants.PercentOperator:
		if rVal == 0 {
			return object.NewErrorWithToken(constants.EmptyString, token, constants.ZeroDivisionError, constants.EvalExpressionsIntegerModuloByZero)
		}
		return &object.Integer{Value: lVal % rVal}
	case "<<":
        return &object.Integer{Value: lVal << rVal}
    case ">>":
        return &object.Integer{Value: lVal >> rVal}
	case constants.LessThanOp:
		return object.NativeBoolToBooleanObject(lVal < rVal)
	case constants.LessThanOrEqualToOp:
		return object.NativeBoolToBooleanObject(lVal <= rVal)
	case constants.GreaterThanOp:
		return object.NativeBoolToBooleanObject(lVal > rVal)
	case constants.GreaterThanOrEqualToOp:
		return object.NativeBoolToBooleanObject(lVal >= rVal)
	case "&":
		return &object.Integer{Value: lVal & rVal}
	case "|":
		return &object.Integer{Value: lVal | rVal}
	case "^":
		return &object.Integer{Value: lVal ^ rVal}
	default:
		return object.NewErrorWithToken(constants.EmptyString, token, constants.TypeError, constants.EvalExpressionsTypeErrorUnsupportedOperandType, op, left.Type(), right.Type())
	}
}


func evalFloatInfixExpression(op string, left, right object.Object, token lexer.Token) object.Object {
	lVal, rVal := left.(*object.Float).Value, right.(*object.Float).Value
	switch op {
	case constants.PlusOperator:
		return &object.Float{Value: lVal + rVal}
	case constants.MinusOperator:
		return &object.Float{Value: lVal - rVal}
	case constants.AsteriskOperator:
		return &object.Float{Value: lVal * rVal}
	case constants.PowOperator: // <<< ADD THIS CASE
		// Note: math.Pow handles edge cases like 0**0=1
		result := math.Pow(lVal, rVal)
		return &object.Float{Value: result}
	case constants.SlashOperator:
		if rVal == 0.0 {
			return object.NewErrorWithToken(constants.EmptyString, token, constants.ZeroDivisionError, constants.EvalExpressionsFloatDivisionByZero)
		}
		return &object.Float{Value: lVal / rVal}
	case constants.PercentOperator:
		return object.NewError(constants.NotImplementedError, constants.EvalExpressionsFloatModuloRequiresMath)
	case constants.LessThanOp:
		return object.NativeBoolToBooleanObject(lVal < rVal)
	case constants.LessThanOrEqualToOp:
		return object.NativeBoolToBooleanObject(lVal <= rVal)
	case constants.GreaterThanOp:
		return object.NativeBoolToBooleanObject(lVal > rVal)
	case constants.GreaterThanOrEqualToOp:
		return object.NativeBoolToBooleanObject(lVal >= rVal)
	default:
		return object.NewErrorWithToken(constants.EmptyString, token, constants.TypeError, constants.EvalExpressionsTypeErrorUnsupportedOperandType, op, left.Type(), right.Type())
	}
}

// floorDiv implements Python's floor division `//` for integers.
func floorDiv(a, b int64) int64 {
    q := a / b
    r := a % b
    // If the signs of a and b are different and there's a remainder,
    // the quotient needs to be adjusted downwards.
    if (a > 0 && b < 0 || a < 0 && b > 0) && r != 0 {
        return q - 1
    }
    return q
}

func evalStringInfixExpression(op string, left, right object.Object, token lexer.Token) object.Object {
	lVal, rVal := left.(*object.String).Value, right.(*object.String).Value
	switch op {
	case constants.PlusOperator:
		return &object.String{Value: lVal + rVal}
	case constants.LessThanOp:
		return object.NativeBoolToBooleanObject(lVal < rVal)
	case constants.LessThanOrEqualToOp:
		return object.NativeBoolToBooleanObject(lVal <= rVal)
	case constants.GreaterThanOp:
		return object.NativeBoolToBooleanObject(lVal > rVal)
	case constants.GreaterThanOrEqualToOp:
		return object.NativeBoolToBooleanObject(lVal >= rVal)
	default:
		return object.NewErrorWithToken(constants.EmptyString, token, constants.TypeError, constants.EvalExpressionsStringUnsupportedOperandType, op)
	}
}
func evalListLiteral(node *ast.ListLiteral, ctx *InterpreterContext) object.Object {
	elements := evalExpressions(node.Elements, ctx)
	if len(elements) == 1 && object.IsError(elements[0]) {
		return elements[0]
	}
	return &object.List{Elements: elements}
}
func evalTupleLiteral(node *ast.TupleLiteral, ctx *InterpreterContext) object.Object {
	elements := evalExpressions(node.Elements, ctx)
	if len(elements) == 1 && object.IsError(elements[0]) {
		return elements[0]
	}
	return &object.Tuple{Elements: elements}
}
func evalSetLiteral(node *ast.SetLiteral, ctx *InterpreterContext) object.Object {
	set := &object.Set{Elements: make(map[object.HashKey]object.Object)}
	for _, elNode := range node.Elements {
		el := Eval(elNode, ctx)
		if object.IsError(el) {
			return el
		}
		hashableEl, ok := el.(object.Hashable)
		if !ok {
			return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.EvalExpressionsUnhashableType, el.Type())
		}
		hKey, err := hashableEl.HashKey()
		if err != nil {
			return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.EvalExpressionsFailedToHashElementForSet, err)
		}
		set.Elements[hKey] = el
	}
	return set
}
func evalDictLiteral(node *ast.DictLiteral, ctx *InterpreterContext) object.Object {
	pairs := make(map[object.HashKey]object.DictPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, ctx)
		if object.IsError(key) {
			return key
		}
		hashableKey, ok := key.(object.Hashable)
		if !ok {
			return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.EvalExpressionsUnhashableType, key.Type())
		}
		hashed, err := hashableKey.HashKey()
		if err != nil {
			return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.EvalExpressionsFailedToHashKey, err)
		}
		value := Eval(valueNode, ctx)
		if object.IsError(value) {
			return value
		}
		pairs[hashed] = object.DictPair{Key: key, Value: value}
	}
	return &object.Dict{Pairs: pairs}
}
func evalIndexExpression(node *ast.IndexExpression, ctx *InterpreterContext) object.Object {
	left := Eval(node.Left, ctx)
	if object.IsError(left) {
		return left
	}
	index := Eval(node.Index, ctx)
	if object.IsError(index) {
		return index
	}
	token := node.Token
	if inst, ok := left.(*object.Instance); ok && inst.Class != nil {
		if getItemMethodObj, methodOk := inst.Class.Methods[constants.DunderGetItem]; methodOk {
			// <<< FIX: Type assert that the method is a function >>>
			if getItemMethod, isFunc := getItemMethodObj.(*object.Function); isFunc {
				boundGetItem := &object.BoundMethod{Instance: inst, Method: getItemMethod}
				return object.ApplyBoundMethod(ctx, boundGetItem, []object.Object{index}, token)
			}
		}
	}
	switch l := left.(type) {
	case *object.List:
		if idxObj, ok := index.(*object.Integer); ok {
			return evalListIndexExpression(l, idxObj, token)
		} else {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsListIndexMustBeInteger, index.Type())
		}
	case *object.Tuple:
		if idxObj, ok := index.(*object.Integer); ok {
			return evalTupleIndexExpression(l, idxObj, token)
		} else {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsTupleIndexMustBeInteger, index.Type())
		}
	case *object.Bytes:
		if idxObj, ok := index.(*object.Integer); ok {
			return evalBytesIndexExpression(l, idxObj, token)
		} else {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsBytesIndexMustBeInteger, index.Type())
		}
	case *object.Dict:
		return evalDictIndexExpressionRevised(l, index, token)
	case *object.String:
		if idxObj, ok := index.(*object.Integer); ok {
			return evalStringIndexExpression(l, idxObj, token)
		} else {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsStringIndexMustBeInteger, index.Type())
		}
	default:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsObjectNotSubscriptable, left.Type())
	}
}
func evalListIndexExpression(list *object.List, index *object.Integer, token lexer.Token) object.Object {
	idx, max, min := index.Value, int64(len(list.Elements)-1), -int64(len(list.Elements))
	if idx < 0 {
		if idx < min {
			return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsListIndexOutOfRange)
		}
		idx = int64(len(list.Elements)) + idx
	}
	if idx < 0 || idx > max {
		return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsListIndexOutOfRange)
	}
	return list.Elements[idx]
}
func evalTupleIndexExpression(tuple *object.Tuple, index *object.Integer, token lexer.Token) object.Object {
	idx, max, min := index.Value, int64(len(tuple.Elements)-1), -int64(len(tuple.Elements))
	if idx < 0 {
		if idx < min {
			return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsTupleIndexOutOfRange)
		}
		idx = int64(len(tuple.Elements)) + idx
	}
	if idx < 0 || idx > max {
		return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsTupleIndexOutOfRange)
	}
	return tuple.Elements[idx]
}
func evalBytesIndexExpression(bytesObj *object.Bytes, index *object.Integer, token lexer.Token) object.Object {
	idx, max, min := index.Value, int64(len(bytesObj.Value)-1), -int64(len(bytesObj.Value))
	if idx < 0 {
		if idx < min {
			return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsBytesIndexOutOfRange)
		}
		idx = int64(len(bytesObj.Value)) + idx
	}
	if idx < 0 || idx > max {
		return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsBytesIndexOutOfRange)
	}
	return &object.Integer{Value: int64(bytesObj.Value[idx])}
}
func evalStringIndexExpression(str *object.String, index *object.Integer, token lexer.Token) object.Object {
	runes := []rune(str.Value)
	idx, max, min := index.Value, int64(len(runes)-1), -int64(len(runes))
	if idx < 0 {
		if idx < min {
			return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsStringIndexOutOfRange)
		}
		idx = int64(len(runes)) + idx
	}
	if idx < 0 || idx > max {
		return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsStringIndexOutOfRange)
	}
	return &object.String{Value: string(runes[idx])}
}
func evalDictIndexExpressionRevised(dict *object.Dict, index object.Object, token lexer.Token) object.Object {
	hashableKey, ok := index.(object.Hashable)
	if !ok {
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsUnhashableType, index.Type())
	}
	dictMapKey, err := hashableKey.HashKey()
	if err != nil {
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsFailedToHashKey, err)
	}
	pair, ok := dict.Pairs[dictMapKey]
	if !ok {
		return object.NewErrorWithLocation(token, constants.KeyError, constants.EvalExpressionsKeyError, index.Inspect())
	}
	return pair.Value
}
func evalIndexAssignment(collection, index, value object.Object, token lexer.Token, ctx *InterpreterContext) object.Object {
	if inst, ok := collection.(*object.Instance); ok && inst.Class != nil {
		if setItemMethodObj, methodOk := inst.Class.Methods[constants.DunderSetItem]; methodOk {
			// <<< FIX: Type assert that the method is a function >>>
			if setItemMethod, isFunc := setItemMethodObj.(*object.Function); isFunc {
				boundSetItem := &object.BoundMethod{Instance: inst, Method: setItemMethod}
				result := object.ApplyBoundMethod(ctx, boundSetItem, []object.Object{index, value}, token)
				if object.IsError(result) {
					return result
				}
				return object.NULL
			}
		}
	}
	switch coll := collection.(type) {
	case *object.List:
		idxObj, ok := index.(*object.Integer)
		if !ok {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsListIndexMustBeInteger, index.Type())
		}
		idx, listLen := idxObj.Value, int64(len(coll.Elements))
		if idx < 0 {
			idx += listLen
		}
		if idx < 0 || idx >= listLen {
			return object.NewErrorWithLocation(token, constants.IndexError, constants.EvalExpressionsListAssignmentIndexOutOfRange)
		}
		coll.Elements[idx] = value
		return object.NULL
	case *object.Dict:
		key, ok := index.(object.Hashable)
		if !ok {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsUnhashableType, index.Type())
		}
		hashKeyVal, err := key.HashKey()
		if err != nil {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsFailedToHashKey, err)
		}
		coll.Pairs[hashKeyVal] = object.DictPair{Key: index, Value: value}
		return object.NULL
	case *object.Tuple:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsTupleDoesNotSupportItemAssignment)
	case *object.Bytes:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsBytesDoesNotSupportItemAssignment)
	case *object.String:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsStrDoesNotSupportItemAssignment)
	default:
		return object.NewErrorWithLocation(token, constants.TypeError, constants.EvalExpressionsObjectDoesNotSupportItemAssignment, collection.Type())
	}
}
func evalListComprehension(node *ast.ListComprehension, ctx *InterpreterContext) object.Object {
	iterableObj := Eval(node.Iterable, ctx)
	if object.IsError(iterableObj) {
		return iterableObj
	}
	iterator, errObj := object.GetObjectIterator(ctx, iterableObj, node.Token)
	if errObj != nil {
		return errObj
	}
	elements := []object.Object{}
	for {
		nextItem, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(nextItem) {
			return nextItem
		}
		loopEnv := object.NewEnclosedEnvironment(ctx.Env)
		loopCtx := ctx.NewChildContext(loopEnv).(*InterpreterContext)
		loopEnv.Set(node.Variable.Value, nextItem)
		if node.Condition != nil {
			conditionResult := Eval(node.Condition, loopCtx)
			if object.IsError(conditionResult) {
				return conditionResult
			}
			isTrue, truthyErr := object.IsTruthy(loopCtx, conditionResult)
			if truthyErr != nil {
				if pyErr, ok := truthyErr.(object.Object); ok {
					return pyErr
				}
				return object.NewError(constants.RuntimeError, constants.EvalExpressionsComprehensionIfConditionError, truthyErr)
			}
			if !isTrue {
				continue
			}
		}
		evaluatedElement := Eval(node.Element, loopCtx)
		if object.IsError(evaluatedElement) {
			return evaluatedElement
		}
		elements = append(elements, evaluatedElement)
	}
	return &object.List{Elements: elements}
}
func evalSetComprehension(node *ast.SetComprehension, ctx *InterpreterContext) object.Object {
	iterableObj := Eval(node.Iterable, ctx)
	if object.IsError(iterableObj) {
		return iterableObj
	}
	iterator, errObj := object.GetObjectIterator(ctx, iterableObj, node.Token)
	if errObj != nil {
		return errObj
	}
	resultSet := &object.Set{Elements: make(map[object.HashKey]object.Object)}
	for {
		nextItem, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(nextItem) {
			return nextItem
		}
		loopEnv := object.NewEnclosedEnvironment(ctx.Env)
		loopCtx := ctx.NewChildContext(loopEnv).(*InterpreterContext)
		loopEnv.Set(node.Variable.Value, nextItem)
		if node.Condition != nil {
			conditionResult := Eval(node.Condition, loopCtx)
			if object.IsError(conditionResult) {
				return conditionResult
			}
			isTrue, truthyErr := object.IsTruthy(loopCtx, conditionResult)
			if truthyErr != nil {
				if pyErr, ok := truthyErr.(object.Object); ok {
					return pyErr
				}
				return object.NewError(constants.RuntimeError, constants.EvalExpressionsComprehensionIfConditionError, truthyErr)
			}
			if !isTrue {
				continue
			}
		}
		evaluatedElement := Eval(node.Element, loopCtx)
		if object.IsError(evaluatedElement) {
			return evaluatedElement
		}
		hashable, ok := evaluatedElement.(object.Hashable)
		if !ok {
			return object.NewError(constants.TypeError, constants.ErrUnHashableType, evaluatedElement.Type())
		}
		hashKey, err := hashable.HashKey()
		if err != nil {
			return object.NewError(constants.TypeError, constants.EvalExpressionsSetHashFailed, err)
		}
		resultSet.Elements[hashKey] = evaluatedElement
	}
	return resultSet
}
func evalTernaryExpression(node *ast.TernaryExpression, ctx *InterpreterContext) object.Object {
	condition := Eval(node.Condition, ctx)
	if object.IsError(condition) {
		return condition
	}
	isTrue, err := object.IsTruthy(ctx, condition)
	if err != nil {
		if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
			return pyErr
		}
		return object.NewError(constants.RuntimeError, constants.EvalExpressionsPropagatedFromIsTruthy, err)
	}
	if isTrue {
		return Eval(node.ValueIfTrue, ctx)
	}
	return Eval(node.ValueIfFalse, ctx)
}

// evalDotExpression handles `object.attribute` access.
// evalDotExpression handles `object.attribute` access.
func evalDotExpression(node *ast.DotExpression, ctx *InterpreterContext) object.Object {
	left := Eval(node.Left, ctx)
	if object.IsError(left) {
		return left
	}

	attrName := node.Identifier.Value
	token := node.Identifier.Token

	// --- Primary Lookup: AttributeGetter interface ---
	if getter, ok := left.(object.AttributeGetter); ok {
		attrVal, found := getter.GetObjectAttribute(ctx, attrName) // Pass context
		if found {
			if object.IsError(attrVal) { // Propagate error if GetObjectAttribute returned one
				return attrVal
			}
			return attrVal // Return the found attribute (could be a method/builtin)
		}
	}

	// --- Fallback to __getattr__ for Pylearn instances ---
	// In Python, __getattr__ is only called if the attribute is not found through normal mechanisms.
	if inst, isInst := left.(*object.Instance); isInst && inst.Class != nil {
		if getattrDunderObj, hasDunder := inst.Class.Methods[constants.DunderGetAttr]; hasDunder {
			if getattrMethod, isFunc := getattrDunderObj.(*object.Function); isFunc {
				boundGetattr := &object.BoundMethod{Instance: inst, Method: getattrMethod}
				attrNameObj := &object.String{Value: attrName}
				return applyFunctionOrClass(ctx, boundGetattr, []object.Object{attrNameObj}, nil, token)
			}
		}
	}

	// If all lookups fail, raise an AttributeError.
	typeName := string(left.Type())
	if inst, ok := left.(*object.Instance); ok && inst.Class != nil {
		typeName = inst.Class.Name
	}
	return object.NewErrorWithLocation(token, constants.AttributeError, constants.ErrNoAttribute, typeName, attrName)
}



func evalBytesInfixExpression(op string, left, right object.Object, token lexer.Token) object.Object {
	lVal := left.(*object.Bytes).Value
	rVal := right.(*object.Bytes).Value

	switch op {
	case constants.PlusOperator:
		// Concatenate the byte slices
		newBytes := make([]byte, len(lVal)+len(rVal))
		copy(newBytes, lVal)
		copy(newBytes[len(lVal):], rVal)
		return &object.Bytes{Value: newBytes}
	// Other operators like '==' for bytes would be handled by CompareObjects if a dunder method is added.
	// For now, we only need to support addition.
	default:
		return object.NewErrorWithToken(constants.EmptyString, token, constants.TypeError, constants.EvalExpressionsTypeErrorUnsupportedOperandType, op, left.Type(), right.Type())
	}
}