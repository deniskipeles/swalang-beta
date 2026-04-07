// internal/interpreter/apply.go
package interpreter

import (
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
)

func applyFunctionOrClass(
	ctx *InterpreterContext, // Takes the context directly
	callable object.Object,
	providedPositionalArgs []object.Object,
	callsiteKeywordArgs map[string]object.Object,
	callToken lexer.Token,
) object.Object {
	// THIS IS THE FIX: Save the caller's super() context and restore it when this function returns.
	// This ensures that nested method calls (like super().__init__ inside another method)
	// don't permanently alter the super() context of the outer method's frame.
	callerSuperSelf, callerSuperClass := ctx.GetSuperSelf(), ctx.GetSuperClass()
	defer ctx.SetSuperContext(callerSuperSelf, callerSuperClass)

	// Set the super() context if we are about to call a method
	if bm, isBoundMethod := callable.(*object.BoundMethod); isBoundMethod && bm.Method.IsAMethod {
		ctx.SetSuperContext(bm.Instance, bm.Method.OriginalClass)
	}

	switch fn := callable.(type) {
	case *object.Function:
		funcName := fn.Name
		if funcName == constants.EmptyString {
			funcName = constants.FunctionLiteralFunctionPlaceholder
		}

		// Create the environment for the function call, enclosed by the function's definition environment.
		extendedEnv := object.NewEnclosedEnvironment(fn.Env)

		// --- Argument Binding Logic (No changes here) ---
		paramIsAssigned := make([]bool, len(fn.Parameters))
		paramValues := make([]object.Object, len(fn.Parameters))
		for i := 0; i < len(fn.Parameters); i++ {
			if i < len(providedPositionalArgs) {
				paramValues[i] = providedPositionalArgs[i]
				paramIsAssigned[i] = true
			} else {
				break
			}
		}
		numPositionalArgsConsumedByFormalParams := 0
		for _, assigned := range paramIsAssigned {
			if assigned {
				numPositionalArgsConsumedByFormalParams++
			}
		}
		tempKwargsForVarKw := make(map[string]object.Object)
		if callsiteKeywordArgs != nil {
			for key, value := range callsiteKeywordArgs {
				foundParam := false
				for i, formalParam := range fn.Parameters {
					if formalParam.Name == key {
						if paramIsAssigned[i] {
							return object.NewErrorWithToken(constants.EmptyString, callToken, constants.TypeError, constants.InterpreterApplyFunctionKwargMultipleValuesError, funcName, key)
						}
						paramValues[i] = value
						paramIsAssigned[i] = true
						foundParam = true
						break
					}
				}
				if !foundParam {
					if fn.KwArgParam != constants.EmptyString {
						tempKwargsForVarKw[key] = value
					} else {
						return object.NewErrorWithToken(constants.EmptyString, callToken, constants.TypeError, constants.InterpreterApplyFunctionUnexpectedKwargError, funcName, key)
					}
				}
			}
		}
		for i, formalParam := range fn.Parameters {
			if !paramIsAssigned[i] {
				if formalParam.DefaultValue != nil {
					paramValues[i] = formalParam.DefaultValue
				} else {
					return object.NewErrorWithToken(constants.EmptyString, callToken, constants.TypeError, constants.InterpreterApplyFunctionMissingPositionalArgError, funcName, formalParam.Name)
				}
			}
			extendedEnv.Set(formalParam.Name, paramValues[i])
		}
		if fn.VarArgParam != constants.EmptyString {
			varArgsCollected := []object.Object{}
			if len(providedPositionalArgs) > numPositionalArgsConsumedByFormalParams {
				varArgsCollected = providedPositionalArgs[numPositionalArgsConsumedByFormalParams:]
			}
			extendedEnv.Set(fn.VarArgParam, &object.List{Elements: varArgsCollected})
		} else {
			if len(providedPositionalArgs) > len(fn.Parameters) {
				return object.NewErrorWithToken(constants.EmptyString, callToken, constants.TypeError, constants.InterpreterApplyFunctionTooManyPositionalArgsError, funcName, len(fn.Parameters), len(providedPositionalArgs))
			}
		}
		if fn.KwArgParam != constants.EmptyString {
			kwDictPairs := make(map[object.HashKey]object.DictPair)
			for key, val := range tempKwargsForVarKw {
				keyObj := &object.String{Value: key}
				hashKey, _ := keyObj.HashKey()
				kwDictPairs[hashKey] = object.DictPair{Key: keyObj, Value: val}
			}
			extendedEnv.Set(fn.KwArgParam, &object.Dict{Pairs: kwDictPairs})
		}
		// --- End of Argument Binding Logic ---

		// First, check if the function being called is a generator.
		// We'll determine this by a quick scan of its AST body.
		// A more efficient implementation would do this once in the parser.
		isGenerator := astContainsYield(fn.Body)
		if isGenerator {
			generatorCtx := ctx.NewChildContext(extendedEnv).(*InterpreterContext)

			// The SendFn is the heart of the generator's execution.
			// It maintains its state via the captured `generatorCtx`.
			genSendFn := func(valueToSend object.Object) (object.Object, bool) {
				generatorCtx.SentInValue = valueToSend
			
				for {
					if generatorCtx.InstructionPtr >= len(fn.Body.Statements) {
						return object.STOP_ITERATION, true // End of function
					}
			
					stmt := fn.Body.Statements[generatorCtx.InstructionPtr]
					result := Eval(stmt, generatorCtx)
			
					if yv, ok := result.(*object.YieldValue); ok {
						generatorCtx.IsResuming = true
						return yv.Value, false
					}
			
					// The statement completed.
					// We only advance the instruction pointer if it was NOT a continue.
					isContinue := false
					if _, ok := result.(*object.ContinueObject); ok {
						isContinue = true
					}
			
					// --- THIS IS THE FIX ---
					// Only advance to the next statement if the loop body didn't 'continue'.
					// If it was a 'continue', we stay on the current statement (the for/while loop)
					// so it can be re-evaluated to get the next iteration.
					if !isContinue {
						generatorCtx.InstructionPtr++
					}
					// --- END OF FIX ---
			
					generatorCtx.IsResuming = false
			
					if object.IsError(result) {
						return result, true
					}
					if _, isReturn := result.(*object.ReturnValue); isReturn {
						return object.STOP_ITERATION, true
					}
				}
			}

			gen := &object.Generator{
				Name:   fn.Name,
				SendFn: genSendFn,
			}
			return gen
		}

		// --- THIS IS THE FIX ---
		if fn.IsAsync {
			// This is an async function call. We do NOT execute the body.
			// Instead, we return a "coroutine" object, which is a new Function
			// instance that has captured the arguments in its environment.
			coroutine := &object.Function{
				Name:          fn.Name,
				Parameters:    fn.Parameters, // Parameters definition can be shared
				VarArgParam:   fn.VarArgParam,
				KwArgParam:    fn.KwArgParam,
				Body:          fn.Body,     // The AST body can be shared
				Env:           extendedEnv, // CRUCIAL: Use the new env with bound arguments
				IsAsync:       true,
				OriginalClass: fn.OriginalClass, // For methods
				IsAMethod:     fn.IsAMethod,
			}
			return coroutine
		}
		// --- END OF FIX ---

		// For regular synchronous functions, execute the body as before.
		bodyEvalCtx := ctx.NewChildContext(extendedEnv).(*InterpreterContext)
		evaluated := Eval(fn.Body, bodyEvalCtx)

		if returnValue, ok := evaluated.(*object.ReturnValue); ok {
			return returnValue.Value
		}
		if object.IsError(evaluated) {
			return evaluated
		}

		// Synchronous functions implicitly return NULL.
		return object.NULL

	case *object.Class: // Instantiating a class
		instance := &object.Instance{Class: fn, Env: object.NewEnvironment()}

		// <<< THIS IS THE CORRECTED LOGIC >>>
		var initCallable object.Object
		initExists := false

		// Search the MRO for the __init__ method.
		for _, classInMRO := range fn.MRO {
			// Find the __init__ attribute, which can be any object type.
			if methodObj, found := classInMRO.Methods[constants.DunderInit]; found {
				initCallable = methodObj
				initExists = true
				break // Found the first __init__ in the MRO, stop searching.
			}
		}
		// <<< END OF CORRECTED LOGIC >>>

		if initExists {
			// Prepare arguments for __init__: prepend the new instance as `self`.
			initPositionalArgs := make([]object.Object, 1+len(providedPositionalArgs))
			initPositionalArgs[0] = instance
			copy(initPositionalArgs[1:], providedPositionalArgs)

			// Set the context for any super() calls that might happen inside __init__.
			// We need to know which class provided the __init__ method.
			if initFunc, isFunc := initCallable.(*object.Function); isFunc && initFunc.IsAMethod {
				ctx.SetSuperContext(instance, initFunc.OriginalClass)
			} else {
				// For built-ins or non-method functions, the class being instantiated is the context.
				ctx.SetSuperContext(instance, fn)
			}

			// Call the found __init__ callable (could be a Function or a Builtin).
			// The recursive call to applyFunctionOrClass will handle dispatching it correctly.
			initResult := applyFunctionOrClass(ctx, initCallable, initPositionalArgs, callsiteKeywordArgs, callToken)

			if object.IsError(initResult) {
				return initResult
			}
			if initResult != object.NULL {
				return object.NewErrorWithToken(constants.EmptyString, callToken, constants.TypeError, constants.InterpreterApplyFunctionInitReturnNoneError, initResult.Type())
			}
		} else { // No __init__ method found in the entire class hierarchy.
			if len(providedPositionalArgs) > 0 || len(callsiteKeywordArgs) > 0 {
				// The default object.__init__ takes no arguments.
				return object.NewErrorWithToken(constants.EmptyString, callToken, constants.TypeError, constants.InterpreterApplyFunctionTakesNoArgumentsError, fn.Name)
			}
		}
		return instance

	case *object.BoundMethod:
		methodPositionalArgs := make([]object.Object, 1+len(providedPositionalArgs))
		methodPositionalArgs[0] = fn.Instance
		copy(methodPositionalArgs[1:], providedPositionalArgs)
		if fn.Method.IsAMethod {
			ctx.SetSuperContext(fn.Instance, fn.Method.OriginalClass)
		}
		return applyFunctionOrClass(ctx, fn.Method, methodPositionalArgs, callsiteKeywordArgs, callToken)

	case *object.Builtin:
		// <<< THIS IS THE NEW, CORRECTED LOGIC FOR BUILTINS >>>
		finalArgs := []object.Object{}
		finalArgs = append(finalArgs, providedPositionalArgs...)

		// Check and process keyword arguments.
		if len(callsiteKeywordArgs) > 0 {
			// Check if this builtin is configured to accept any keywords at all.
			if fn.AcceptsKeywords == nil {
				firstKey := constants.EmptyString
				for k := range callsiteKeywordArgs {
					firstKey = k
					break
				}
				return object.NewError(constants.TypeError, constants.InterpreterApplyFunctionUnexpectedKwargInBuiltin, fn.Name, firstKey)
			}

			// Validate the provided keywords against the builtin's accepted list.
			for key := range callsiteKeywordArgs {
				if !fn.AcceptsKeywords[key] {
					return object.NewError(constants.TypeError, constants.InterpreterApplyFunctionUnexpectedKwargInBuiltin, fn.Name, key)
				}
			}

			// If keywords are valid, pack them into a dictionary and pass it
			// as the *last* argument to the Go function.
			kwargsDict := &object.Dict{Pairs: make(map[object.HashKey]object.DictPair)}
			for key, value := range callsiteKeywordArgs {
				keyObj := &object.String{Value: key}
				hashKey, _ := keyObj.HashKey()
				kwargsDict.Pairs[hashKey] = object.DictPair{Key: keyObj, Value: value}
			}
			finalArgs = append(finalArgs, kwargsDict)
		}

		// Call the underlying Go function with the final list of arguments.
		return fn.Fn(ctx, finalArgs...)

	default:
		return object.NewErrorWithLocation(callToken, constants.TypeError, constants.InterpreterApplyFunctionObjectNotCallable, callable.Type())
	}
}

// Helper to recursively check if an AST node contains a YieldExpression.
// This function traverses the AST to determine if a function body contains 'yield',
// which is the defining characteristic of a generator function.
func astContainsYield(node ast.Node) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	// Base case: We found a yield expression.
	case *ast.YieldExpression:
		return true

	// --- Statements that contain other statements or expressions ---

	case *ast.Program:
		for _, stmt := range n.Statements {
			if astContainsYield(stmt) {
				return true
			}
		}

	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			if astContainsYield(stmt) {
				return true
			}
		}

	case *ast.ExpressionStatement:
		return astContainsYield(n.Expression)

	case *ast.LetStatement:
		return astContainsYield(n.Value)

	case *ast.AssignStatement:
		// Iterate over all possible targets in a multiple assignment.
		// If the target is a TupleLiteral (for multi-assignment),
		// astContainsYield will recursively check its elements correctly.
		return astContainsYield(n.Target) || astContainsYield(n.Value)


	case *ast.ReturnStatement:
		return astContainsYield(n.ReturnValue)

	case *ast.IfStatement:
		if astContainsYield(n.Condition) || astContainsYield(n.Consequence) {
			return true
		}
		for _, eb := range n.ElifBlocks {
			if astContainsYield(eb.Condition) || astContainsYield(eb.Consequence) {
				return true
			}
		}
		if n.Alternative != nil {
			return astContainsYield(n.Alternative)
		}

	case *ast.WhileStatement:
		return astContainsYield(n.Condition) || astContainsYield(n.Body)

	case *ast.ForStatement:
		return astContainsYield(n.Iterable) || astContainsYield(n.Body)

	case *ast.WithStatement:
		return astContainsYield(n.ContextManager) || astContainsYield(n.Body)

	case *ast.TryStatement:
		if astContainsYield(n.Body) {
			return true
		}
		for _, handler := range n.Handlers {
			if astContainsYield(handler.Type) || astContainsYield(handler.Body) {
				return true
			}
		}
		// In a full implementation, you'd also check Finally and Else blocks here.

	case *ast.RaiseStatement:
		return astContainsYield(n.Exception)

	// --- Expressions that contain other expressions ---

	case *ast.PrefixExpression:
		return astContainsYield(n.Right)

	case *ast.InfixExpression:
		return astContainsYield(n.Left) || astContainsYield(n.Right)

	case *ast.DotExpression:
		return astContainsYield(n.Left)

	case *ast.IndexExpression:
		return astContainsYield(n.Left) || astContainsYield(n.Index)

	case *ast.SliceExpression:
		return astContainsYield(n.Left) || astContainsYield(n.Start) || astContainsYield(n.Stop) || astContainsYield(n.Step)

	case *ast.CallExpression:
		if astContainsYield(n.Function) {
			return true
		}
		for _, arg := range n.Arguments {
			if astContainsYield(arg) {
				return true
			}
		}

	case *ast.AwaitExpression:
		return astContainsYield(n.Expression)

	case *ast.TernaryExpression:
		return astContainsYield(n.ValueIfTrue) || astContainsYield(n.Condition) || astContainsYield(n.ValueIfFalse)

	case *ast.ListLiteral:
		for _, el := range n.Elements {
			if astContainsYield(el) {
				return true
			}
		}

	case *ast.TupleLiteral:
		for _, el := range n.Elements {
			if astContainsYield(el) {
				return true
			}
		}

	case *ast.DictLiteral:
		for k, v := range n.Pairs {
			if astContainsYield(k) || astContainsYield(v) {
				return true
			}
		}

	case *ast.SetLiteral:
		for _, el := range n.Elements {
			if astContainsYield(el) {
				return true
			}
		}

	case *ast.ListComprehension:
		return astContainsYield(n.Element) || astContainsYield(n.Iterable) || astContainsYield(n.Condition)

	case *ast.SetComprehension:
		return astContainsYield(n.Element) || astContainsYield(n.Iterable) || astContainsYield(n.Condition)

	case *ast.KeywordArgument:
		return astContainsYield(n.Value)

	case *ast.StarredArgument:
		return astContainsYield(n.Value)

	// --- IMPORTANT: Do not traverse into nested function or class definitions ---
	// A `yield` inside a nested function does not make the outer function a generator.
	case *ast.FunctionLiteral, *ast.ClassStatement:
		return false

	// --- Leaf nodes that cannot contain `yield` ---
	case *ast.Identifier, *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BytesLiteral, *ast.BooleanLiteral, *ast.NilLiteral, *ast.PassStatement,
		*ast.BreakStatement, *ast.ContinueStatement, *ast.ImportStatement, *ast.FromImportStatement:
		return false
	}

	// Default case for any unhandled node types.
	return false
}
