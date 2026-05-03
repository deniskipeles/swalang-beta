package interpreter

import (
	// "fmt"
	"strings"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
)

// All functions now correctly receive and pass the *InterpreterContext.

func evalProgram(stmts []ast.Statement, ctx *InterpreterContext) object.Object {
	var result object.Object = object.NULL
	for _, statement := range stmts {
		result = Eval(statement, ctx) // Pass ctx
		switch res := result.(type) {
		case *object.ReturnValue:
			return res.Value
		case *object.Error:
			return res
		case *object.BreakObject, *object.ContinueObject:
			return object.NewError(constants.SyntaxError, constants.InterpreterEvalStatementsBreakOutsideLoop, result.Type())
		case *object.StopIterationError:
			return object.NULL
		}
	}
	if result != nil && (result.Type() == object.BREAK_OBJ || result.Type() == object.CONTINUE_OBJ) {
		return object.NULL
	}
	return result
}

func evalBlockStatement(stmts []ast.Statement, ctx *InterpreterContext) object.Object {
	var result object.Object = object.NULL
	for _, statement := range stmts {
		result = Eval(statement, ctx) // Pass ctx
		if result != nil {
			// Add a check for YieldValue here. If a statement inside the block
			// yields, the block must immediately return that signal.
			if _, isYield := result.(*object.YieldValue); isYield {
				return result
			}
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ || rt == object.BREAK_OBJ || rt == object.CONTINUE_OBJ || rt == object.STOP_ITER_OBJ {
				return result
			}
		}
		if result == nil {
			result = object.NULL
		}
	}
	return result
}

func evalLetStatement(node *ast.LetStatement, ctx *InterpreterContext) object.Object {
	valToAssign := Eval(node.Value, ctx)

	// If the expression was a `yield` that is now PAUSING,
	// `valToAssign` is a `YieldValue` signal. Propagate this signal up
	// to the driver loop without performing an assignment yet.
	if _, ok := valToAssign.(*object.YieldValue); ok {
		return valToAssign
	}

	// If the `yield` is RESUMING, `valToAssign` is the normal value sent in by `send()`.
	// The function proceeds with the assignment as normal.
	if object.IsError(valToAssign) {
		return valToAssign
	}
	
	// ... (decorator logic remains the same)
	if fnLit, ok := node.Value.(*ast.FunctionLiteral); ok && len(fnLit.Decorators) > 0 {
		decoratedFunc := valToAssign
		for i := len(fnLit.Decorators) - 1; i >= 0; i-- {
			decoratorExpr := fnLit.Decorators[i]
			decoratorObj := Eval(decoratorExpr, ctx)
			if object.IsError(decoratorObj) {
				return decoratorObj
			}
			decoratorToken := ast.GetToken(decoratorExpr)
			decoratedFunc = applyFunctionOrClass(ctx, decoratorObj, []object.Object{decoratedFunc}, nil, decoratorToken)

			if object.IsError(decoratedFunc) {
				return decoratedFunc
			}
		}
		valToAssign = decoratedFunc
	}

	ctx.Env.Set(node.Name.Value, valToAssign)
	return object.NULL
}

func evalReturnStatement(node *ast.ReturnStatement, ctx *InterpreterContext) object.Object {
	val := Eval(node.ReturnValue, ctx) // Pass ctx
	// If the return value expression yields, propagate the yield signal.
	// Upon resumption, this statement will be re-evaluated, and `val` will
	// be the value sent in, which is then correctly returned.
	if _, isYield := val.(*object.YieldValue); isYield {
		return val
	}
	if object.IsError(val) {
		return val
	}
	return &object.ReturnValue{Value: val}
}

func evalIfStatement(node *ast.IfStatement, ctx *InterpreterContext) object.Object {
	condition := Eval(node.Condition, ctx) // Pass ctx
	if object.IsError(condition) {
		return condition
	}
	// If the condition itself yields, propagate the signal immediately.
	// The `if` statement will be re-evaluated upon resumption.
	if _, isYield := condition.(*object.YieldValue); isYield {
		return condition
	}
	truthy, err := object.IsTruthy(ctx, condition)
	if err != nil {
		if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
			return pyErr
		}
		return object.NewError(constants.RuntimeError, constants.InterpreterEvalStatementsIsTruthyPropagatedError, err)
	}
	if truthy {
		return Eval(node.Consequence, ctx) // Pass ctx
	}
	for _, elifBlock := range node.ElifBlocks {
		elifCondition := Eval(elifBlock.Condition, ctx) // Pass ctx
		if object.IsError(elifCondition) {
			return elifCondition
		}
		truthy, err = object.IsTruthy(ctx, elifCondition)
		if err != nil {
			if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.InterpreterEvalStatementsIsTruthyPropagatedError, err)
		}
		if truthy {
			return Eval(elifBlock.Consequence, ctx) // Pass ctx
		}
	}
	if node.Alternative != nil {
		return Eval(node.Alternative, ctx) // Pass ctx
	}
	return object.NULL
}

func evalWhileStatement(node *ast.WhileStatement, ctx *InterpreterContext) object.Object {
	for {
		condition := Eval(node.Condition, ctx) // Pass ctx
		if object.IsError(condition) {
			return condition
		}
		truthy, err := object.IsTruthy(ctx, condition)
		if err != nil {
			if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.InterpreterEvalStatementsIsTruthyPropagatedError, err)
		}
		if !truthy {
			break
		}
		bodyResult := Eval(node.Body, ctx) // Pass ctx
		if bodyResult != nil {
			// If the body yields, we must propagate the yield signal immediately
			// to pause the entire while loop.
			if _, isYield := bodyResult.(*object.YieldValue); isYield {
				return bodyResult
			}

			rt := bodyResult.Type()
			if rt == object.ERROR_OBJ || rt == object.RETURN_VALUE_OBJ || rt == object.STOP_ITER_OBJ {
				return bodyResult
			}
			if rt == object.BREAK_OBJ {
				break
			}
			if rt == object.CONTINUE_OBJ {
				continue
			}
		}
	}
	return object.NULL
}


func evalForStatement(fs *ast.ForStatement, ctx *InterpreterContext) object.Object {
    iterableObj := Eval(fs.Iterable, ctx)
    if object.IsError(iterableObj) {
        return iterableObj
    }
    if _, isYield := iterableObj.(*object.YieldValue); isYield {
        return iterableObj
    }

    var iterator object.Iterator
    var nextItem object.Object
    var stop bool

    // Check if we're resuming
    if ctx.IsResuming {
        // Try to get the active iterator
        if iter, exists := ctx.ActiveIterators[fs]; exists {
            iterator = iter
        } else {
            // No iterator → create one
            var errObj object.Object
            iterator, errObj = object.GetObjectIterator(ctx, iterableObj, fs.Token)
            if errObj != nil {
                return errObj
            }
            ctx.ActiveIterators[fs] = iterator
            nextItem, stop = iterator.Next()
            if stop {
                delete(ctx.ActiveIterators, fs)
                return object.NULL
            }
            if object.IsError(nextItem) {
                return nextItem
            }
        }

        // On resume, reuse the pending item
        if pending, has := ctx.PendingForItem[fs]; has {
            nextItem = pending
            stop = false
        } else {
            // Should not happen, but fallback
            nextItem, stop = iterator.Next()
            if stop {
                delete(ctx.ActiveIterators, fs)
                return object.NULL
            }
            if object.IsError(nextItem) {
                return nextItem
            }
        }
    } else {
        // First time: create iterator and get first item
        var errObj object.Object
        iterator, errObj = object.GetObjectIterator(ctx, iterableObj, fs.Token)
        if errObj != nil {
            return errObj
        }
        ctx.ActiveIterators[fs] = iterator
        nextItem, stop = iterator.Next()
        if stop {
            delete(ctx.ActiveIterators, fs)
            return object.NULL
        }
        if object.IsError(nextItem) {
            return nextItem
        }
    }

    var loopResult object.Object = object.NULL
loop:
    for {
        // In Python, for loops do not create a new scope.
        // We use the current context and environment directly.
        loopCtx := ctx

        // Bind loop variable(s)
        if len(fs.Variables) > 1 {
            tuple, ok := nextItem.(*object.Tuple)
            if !ok {
                return object.NewError(constants.ValueError, constants.InterpreterEvalStatementsUnpackTooManyValues, len(fs.Variables))
            }
            if len(tuple.Elements) != len(fs.Variables) {
                return object.NewError(constants.ValueError, constants.InterpreterEvalStatementsUnpackingError, len(fs.Variables), len(tuple.Elements))
            }
            for i, ident := range fs.Variables {
                ctx.Env.Set(ident.Value, tuple.Elements[i])
            }
        } else {
            varName := ""
            if fs.Variable != nil {
                varName = fs.Variable.Value
            } else if len(fs.Variables) == 1 {
                varName = fs.Variables[0].Value
            } else {
                return object.NewError(fs.Token.String(), constants.InterpreterEvalStatementsNoLoopVariableSpecified)
            }
            ctx.Env.Set(varName, nextItem)
        }

        // Evaluate body
        bodyResult := Eval(fs.Body, loopCtx)
        if bodyResult != nil {
            if _, isYield := bodyResult.(*object.YieldValue); isYield {
                ctx.PendingForItem[fs] = nextItem
                return bodyResult
            }
            rt := bodyResult.Type()

            // FIX: If a generator nested inside the loop naturally finishes (StopIteration),
            // it is NOT an error. We simply break out of the current loop iteration.
            if rt == object.STOP_ITER_OBJ {
                // Ignore StopIteration from inner calls, it just means the inner generator is done
            } else if rt == object.ERROR_OBJ || rt == object.RETURN_VALUE_OBJ {
                loopResult = bodyResult
                break loop
            } else if rt == object.BREAK_OBJ {
                break loop
            } else if rt == object.CONTINUE_OBJ {
                // Skip to next iteration
                nextItem, stop = iterator.Next()
                if stop {
                    delete(ctx.ActiveIterators, fs)
                    break loop
                }
                if object.IsError(nextItem) {
                    loopResult = nextItem
                    break loop
                }
                continue loop
            }
        }

        // Body completed normally → move to next item
        nextItem, stop = iterator.Next()
        if stop {
            delete(ctx.ActiveIterators, fs)
            break loop
        }
        if object.IsError(nextItem) {
            loopResult = nextItem
            break loop
        }

        // Clear pending item after successful iteration
        delete(ctx.PendingForItem, fs)
        ctx.IsResuming = false
    }
    return loopResult
}


func evalDelStatement(node *ast.DelStatement, ctx *InterpreterContext) object.Object {
	switch target := node.Target.(type) {
	case *ast.IndexExpression:
		collection := Eval(target.Left, ctx)
		if object.IsError(collection) {
			return collection
		}
		index := Eval(target.Index, ctx)
		if object.IsError(index) {
			return index
		}

		deleter, ok := collection.(object.ItemDeleter)
		if !ok {
			return object.NewError(constants.TypeError, constants.STRINGFORMATER_ObjectDoesNotSupportItemDeletion, collection.Type())
		}
		
		return deleter.DeleteObjectItem(index)

	default:
		// This should have been caught by the parser.
		return object.NewError(constants.SyntaxError, constants.InvalidDeletionTarget)
	}
}

// Add the new evaluation function at the end of the file
func evalAssertStatement(node *ast.AssertStatement, ctx *InterpreterContext) object.Object {
	condition := Eval(node.Condition, ctx)
	if object.IsError(condition) {
		return condition
	}

	isTrue, err := object.IsTruthy(ctx, condition)
	if err != nil {
		if pyErr, ok := err.(object.Object); ok && object.IsError(pyErr) {
			return pyErr
		}
		return object.NewError(constants.RuntimeError, constants.PropagatedFromIsTruthyInAssert_VERBFORMATER, err)
	}

	if isTrue {
		return object.NULL // Assertion passed, do nothing
	}

	// Assertion failed
	var message string
	if node.Message != nil {
		msgObj := Eval(node.Message, ctx)
		if object.IsError(msgObj) {
			// If evaluating the message fails, Python raises that error instead.
			return msgObj
		}
		// Use str() on the message object
		strBuiltin, ok := builtins.Builtins[constants.StrBuiltinName]
		if ok {
			strResult := ctx.Execute(strBuiltin, msgObj)
			if strVal, isString := strResult.(*object.String); isString {
				message = strVal.Value
			} else {
				message = msgObj.Inspect() // Fallback
			}
		} else {
			message = msgObj.Inspect() // Fallback if str() is missing
		}
	}

	// Create and return an AssertionError
	return object.NewErrorWithLocation(node.Token, constants.AssertionError, message)
}


func evalClassStatement(stmt *ast.ClassStatement, ctx *InterpreterContext) object.Object {
	className := stmt.Name.Value

	// --- Resolve Superclasses (no changes needed here) ---
	resolvedSuperclasses := []*object.Class{}
	if len(stmt.Superclasses) > 0 {
		for _, superIdent := range stmt.Superclasses {
			val, ok := ctx.Env.Get(superIdent.Value)
			if !ok {
				return object.NewError(constants.NameError, constants.InterpreterEvalStatementsSuperclassNotDefined, superIdent.Value, className)
			}
			sClass, ok := val.(*object.Class)
			if !ok {
				return object.NewError(constants.TypeError, constants.InterpreterEvalStatementsSuperclassNotAClass, superIdent.Value, className, val.Type())
			}
			resolvedSuperclasses = append(resolvedSuperclasses, sClass)
		}
	} else if className != constants.BuiltinsObjectType {
		resolvedSuperclasses = append(resolvedSuperclasses, object.ObjectClass)
	}

	// --- Evaluate the class body to get its contents ---
	classBodyEnv := object.NewEnclosedEnvironment(ctx.Env)
	classBodyCtx := ctx.NewChildContext(classBodyEnv).(*InterpreterContext)
	evalResult := Eval(stmt.Body, classBodyCtx)
	if object.IsError(evalResult) {
		return evalResult
	}

	// --- Create the Class Object ---
	methods := make(map[string]object.Object) // <<< Use the new map type
	classVarEnv := object.NewEnvironment()
	classObj := &object.Class{
		Name:           className,
		Superclasses:   resolvedSuperclasses,
		Methods:        methods,
		ClassVariables: classVarEnv,
	}

	// --- Populate Methods and Class Variables from the evaluated body ---
	for name, obj := range classBodyEnv.Items() {
		// The object `obj` is the *final result* after decorators have been applied.
		// It could be a *StaticMethod, *ClassMethod, or a regular *Function.
		// We store this final object directly in the methods dictionary.
		// The `GetObjectAttribute` logic will handle dispatching them correctly later.
		switch val := obj.(type) {
		case *object.Function:
			// If it's a raw function, tag it as a method.
			val.IsAMethod = true
			val.OriginalClass = classObj
			methods[name] = val
		case *object.StaticMethod, *object.ClassMethod:
			// If it's a decorator wrapper, store it as-is.
			methods[name] = val
		default:
			// Everything else is a class variable.
			classVarEnv.Set(name, val)
		}
	}

	// --- Compute MRO (no changes needed here) ---
	mro, errMRO := object.ComputeMRO(classObj, object.ObjectClass)
	if errMRO != nil {
		return object.NewError(constants.TypeError, errMRO.Error())
	}
	classObj.MRO = mro

	// --- Apply Class Decorators ---
	finalClassObj := object.Object(classObj) // Start with the raw class object
	if len(stmt.Decorators) > 0 {
		for i := len(stmt.Decorators) - 1; i >= 0; i-- {
			decoratorExpr := stmt.Decorators[i]
			decoratorObj := Eval(decoratorExpr, ctx)
			if object.IsError(decoratorObj) {
				return decoratorObj
			}
			decoratorToken := ast.GetToken(decoratorExpr)
			finalClassObj = applyFunctionOrClass(ctx, decoratorObj, []object.Object{finalClassObj}, nil, decoratorToken)
			if object.IsError(finalClassObj) {
				return finalClassObj
			}
		}
	}

	// Define the final (possibly decorated) class in the environment.
	ctx.Env.Set(className, finalClassObj)
	return object.NULL
}

func evalWithStatement(node *ast.WithStatement, ctx *InterpreterContext) object.Object {
	contextManagerObj := Eval(node.ContextManager, ctx)
	if object.IsError(contextManagerObj) {
		return contextManagerObj
	}
	enterMethodObj, foundEnter := object.CallGetAttr(ctx, contextManagerObj, constants.DunderEnter, node.Token)
	if object.IsError(enterMethodObj) && foundEnter {
		return enterMethodObj
	}
	if !foundEnter {
		return object.NewErrorWithLocation(node.Token, constants.AttributeError, constants.InterpreterEvalStatementsWithEnterAttributeError, contextManagerObj.Type())
	}
	if !object.IsCallable(enterMethodObj) {
		return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.InterpreterEvalStatementsWithEnterNotCallable, contextManagerObj.Type())
	}
	enteredValue := ctx.Execute(enterMethodObj)
	if object.IsError(enteredValue) {
		return enteredValue
	}

	bodyEnv := ctx.Env // <<< FIX: Start with current context's env
	if node.TargetVariable != nil {
		bodyEnv = object.NewEnclosedEnvironment(ctx.Env) // <<< FIX: Enclose current context's env
		bodyEnv.Set(node.TargetVariable.Value, enteredValue)
	}

	bodyCtx := ctx.NewChildContext(bodyEnv).(*InterpreterContext) // <<< FIX: Create child context
	var bodyResult object.Object
	var caughtExceptionInBody object.Object = nil
	exitMethodObj, foundExit := object.CallGetAttr(ctx, contextManagerObj, constants.DunderExit, node.Token)
	var exitMethodCallable object.Object
	if object.IsError(exitMethodObj) && foundExit {
		return exitMethodObj
	}
	if !foundExit {
		return object.NewErrorWithLocation(node.Token, constants.AttributeError, constants.InterpreterEvalStatementsWithExitAttributeError, contextManagerObj.Type())
	}
	if !object.IsCallable(exitMethodObj) {
		return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.InterpreterEvalStatementsWithExitNotCallable, contextManagerObj.Type())
	}
	exitMethodCallable = exitMethodObj
	bodyResult = Eval(node.Body, bodyCtx) // <<< FIX: Pass child context
	if bodyResult != nil && bodyResult.Type() == object.ERROR_OBJ {
		caughtExceptionInBody = bodyResult
	} else if bodyResult != nil && bodyResult.Type() == object.STOP_ITER_OBJ {
		caughtExceptionInBody = bodyResult
	}
	var excTypeArg, excValArg, excTbArg object.Object
	if caughtExceptionInBody != nil {
		excValArg = caughtExceptionInBody
		excTypeArg = object.NewString(string(caughtExceptionInBody.Type()))
		excTbArg = object.NULL
	} else {
		excTypeArg, excValArg, excTbArg = object.NULL, object.NULL, object.NULL
	}
	exitResult := ctx.Execute(exitMethodCallable, excTypeArg, excValArg, excTbArg)
	if object.IsError(exitResult) {
		return exitResult
	}
	if caughtExceptionInBody != nil {
		exitReturnedTruthy, truthyErr := object.IsTruthy(ctx, exitResult)
		if truthyErr != nil {
			if pyErr, isPyErr := truthyErr.(object.Object); isPyErr && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewErrorWithLocation(node.Token, constants.RuntimeError, constants.InterpreterEvalStatementsTruthinessOfExitError, truthyErr)
		}
		if !exitReturnedTruthy {
			return caughtExceptionInBody
		}
	}
	if bodyResult != nil && (bodyResult.Type() == object.RETURN_VALUE_OBJ || bodyResult.Type() == object.BREAK_OBJ || bodyResult.Type() == object.CONTINUE_OBJ) {
		if caughtExceptionInBody == nil {
			return bodyResult
		}
	}
	return object.NULL
}

func evalRaiseStatement(node *ast.RaiseStatement, ctx *InterpreterContext) object.Object {
	if node.Exception == nil {
		return object.NewError(constants.RuntimeError, constants.InterpreterEvalStatementsBareRaiseNotImplemented)
	}
	raisedObj := Eval(node.Exception, ctx)
	if object.IsError(raisedObj) {
		return raisedObj
	}

	if raisedClass, isClass := raisedObj.(*object.Class); isClass {
		instance := ctx.Execute(raisedClass)
		if object.IsError(instance) {
			return instance
		}
		raisedObj = instance
	}

	isExceptionInstance, err := object.IsInstance(ctx, raisedObj, object.BaseExceptionClass)
	if err != nil {
		return err
	}
	if !isExceptionInstance {
		return object.NewError(constants.TypeError, constants.InterpreterEvalStatementsExceptionsMustDerive)
	}

	excInstance, _ := raisedObj.(*object.Instance)

	// <<< THIS IS THE FIX: Get the real message by calling __str__ >>>
	// and store the original instance in our internal error wrapper.
	message := ""
	strMethod, found := object.CallGetAttr(ctx, excInstance, constants.DunderStr, node.Token)
	if found && !object.IsError(strMethod) {
		strResult := ctx.Execute(strMethod)
		if strResultStr, isStr := strResult.(*object.String); isStr {
			message = strResultStr.Value
		}
	}
	if message == "" {
		message = excInstance.Inspect() // Fallback
	}

	internalError := &object.Error{
		Message:    message,
		ErrorClass: excInstance.Class,
		Instance:   excInstance, // <<< STORE THE ORIGINAL INSTANCE
		Line:       node.Token.Line,
		Column:     node.Token.Column,
	}
	return internalError
}

func evalTryStatement(ts *ast.TryStatement, ctx *InterpreterContext) object.Object {
	// This will hold the result of the try/except blocks, which might be
	// a normal value, a return, break, continue, or an unhandled error.
	var finalResult object.Object = object.NULL

	// Use `defer` to guarantee the `finally` block executes before this function returns.
	if ts.Finally != nil {
		defer func() {
			// Execute the `finally` block.
			finallyResult := Eval(ts.Finally, ctx)

			// If `finally` executes a control-flow statement (return, break, continue, raise),
			// it "wins" and overwrites the result from the try/except block.
			if finallyResult != nil {
				switch finallyResult.Type() {
				case object.RETURN_VALUE_OBJ, object.ERROR_OBJ, object.BREAK_OBJ, object.CONTINUE_OBJ, object.STOP_ITER_OBJ:
					finalResult = finallyResult
				}
			}
		}()
	}

	// Step 1: Execute the 'try' block.
	tryResult := Eval(ts.Body, ctx)

	// Step 2: Check if an exception was raised in the 'try' block.
	// We use object.IsError(), which correctly identifies both *Error and *StopIterationError.
	if object.IsError(tryResult) {
		wasHandled := false

		// Step 3: Find a matching 'except' handler.
		for _, handler := range ts.Handlers {
			matches := false
			if handler.Type == nil { // Bare except:
				// A bare `except:` catches anything inheriting from BaseException, which includes StopIteration.
				isMatch, _ := object.IsInstance(ctx, tryResult, object.BaseExceptionClass)
				matches = isMatch
			} else {
				expectedTypeObj := Eval(handler.Type, ctx)
				if object.IsError(expectedTypeObj) {
					finalResult = expectedTypeObj
					return finalResult // Return immediately if handler type is invalid
				}
				isMatch, matchErr := object.IsInstance(ctx, tryResult, expectedTypeObj)
				if matchErr != nil {
					finalResult = matchErr
					return finalResult // Return immediately on error
				}
				matches = isMatch
			}

			if matches {
				wasHandled = true
				handlerEnv := object.NewEnclosedEnvironment(ctx.Env)
				if handler.Var != nil {
					// Bind the correct Pylearn-level exception object to the variable.
					var exceptionToBind object.Object = tryResult
					// If it's our internal Error wrapper that holds a user-facing instance, bind that instance.
					if errWrapper, isWrapper := tryResult.(*object.Error); isWrapper && errWrapper.Instance != nil {
						exceptionToBind = errWrapper.Instance
					}
					handlerEnv.Set(handler.Var.Value, exceptionToBind)
				}
				handlerCtx := ctx.NewChildContext(handlerEnv).(*InterpreterContext)

				// The result of the handler block becomes the new final result.
				finalResult = Eval(handler.Body, handlerCtx)
				break // Stop checking other handlers
			}
		}

		if !wasHandled {
			// The exception was not handled, so it's our final result.
			finalResult = tryResult
		}
	} else {
		// The 'try' block completed without an exception.
		finalResult = tryResult
	}

	// Return the determined result. The deferred 'finally' block will execute
	// right before this function returns, potentially changing `finalResult`.
	return finalResult
}

func evalPassStatement(node *ast.PassStatement, ctx *InterpreterContext) object.Object {
	return object.NULL
}
func evalBreakStatement(node *ast.BreakStatement, ctx *InterpreterContext) object.Object {
	return object.BREAK
}
func evalContinueStatement(node *ast.ContinueStatement, ctx *InterpreterContext) object.Object {
	return object.CONTINUE
}

// applyInfixOperation is a helper function to perform an in-place operation (like +=).
// It's needed by evalAssignStatement. It reuses the infix evaluation logic.
func applyInfixOperation(op string, left, right object.Object, token lexer.Token, ctx *InterpreterContext) object.Object {
	// This is a simplified version of evalInfixExpression that works on objects directly.
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
	// Add other type combinations (list + list, etc.) as needed.
	default:
		// Attempt to use dunder methods as a fallback
		dunderName, foundDunder := constants.InfixOperatorToDunder[op]
		if foundDunder {
			if leftGetter, hasGetAttr := left.(object.AttributeGetter); hasGetAttr {
				methodObj, found := leftGetter.GetObjectAttribute(ctx, dunderName)
				if found {
					result := ctx.Execute(methodObj, right)
					if result != object.NOT_IMPLEMENTED {
						return result
					}
				}
			}
		}
		return object.NewErrorWithLocation(token, constants.TypeError, constants.UnsupportedOperandTypesFor_STRINGFORMATER_STRINGFORMATER_STRINGFORMATER, op, left.Type(), right.Type())
	}
}

// evalAssignStatement is the new function that correctly handles all assignments.
func evalAssignStatement(node *ast.AssignStatement, ctx *InterpreterContext) object.Object {
	// Step 1: Evaluate the value on the right-hand side.
	value := Eval(node.Value, ctx)
	if object.IsError(value) {
		return value
	}

	// Step 2: Determine the targets for assignment.
	var targets []ast.Expression
	if tuple, ok := node.Target.(*ast.TupleLiteral); ok {
		targets = tuple.Elements
	} else {
		targets = []ast.Expression{node.Target}
	}

	// Step 3: Get the values to be assigned.
	var valuesToAssign []object.Object
	if len(targets) > 1 {
		// Unpack the iterable on the right-hand side.
		iterator, errObj := object.GetObjectIterator(ctx, value, node.Token)
		if errObj != nil {
			return errObj
		}
		
		unpacked, err := object.UnpackIterator(iterator)
		if err != nil {
			return object.NewErrorFromGoErr(err)
		}
		valuesToAssign = unpacked

		// Check arity
		if len(valuesToAssign) < len(targets) {
			return object.NewErrorWithLocation(node.Token, constants.ValueError, constants.NotEnoughValuesToUnpackExpected_NUMBERFORMATER_Got_NUMBERFORMATER, len(targets), len(valuesToAssign))
		}
		if len(valuesToAssign) > len(targets) {
			return object.NewErrorWithLocation(node.Token, constants.ValueError, constants.TooManyValuesToUnpackExpected_NUMBERFORMATER, len(targets))
		}
	} else {
		// Single assignment. The value might need to be calculated for in-place operators.
		valToAssign := value
		if node.Operator != "=" {
			currentTarget := Eval(targets[0], ctx)
			if object.IsError(currentTarget) {
				return currentTarget
			}
			op := strings.TrimSuffix(node.Operator, "=")
			result := applyInfixOperation(op, currentTarget, value, node.Token, ctx)
			if object.IsError(result) {
				return result
			}
			valToAssign = result
		}
		valuesToAssign = []object.Object{valToAssign}
	}

	// Step 4: Perform the assignments.
	for i, target := range targets {
		val := valuesToAssign[i]
		switch t := target.(type) {
		case *ast.Identifier:
			ctx.Env.Set(t.Value, val)
		case *ast.IndexExpression:
			collection := Eval(t.Left, ctx)
			if object.IsError(collection) { return collection }
			index := Eval(t.Index, ctx)
			if object.IsError(index) { return index }
			if setter, ok := collection.(object.ItemSetter); ok {
				 if res := setter.SetObjectItem(index, val); res != nil { return res }
			} else {
				return object.NewErrorWithLocation(t.Token, constants.TypeError, constants.STRINGFORMATER_ObectDoesNotSupportItemAssignment, collection.Type())
			}
		case *ast.DotExpression:
			obj := Eval(t.Left, ctx)
			if object.IsError(obj) { return obj }
			if inst, ok := obj.(*object.Instance); ok {
				if inst.Env == nil { inst.Env = object.NewEnvironment() }
				inst.Env.Set(t.Identifier.Value, val)
			} else {
				return object.NewErrorWithLocation(t.Token, constants.AttributeError, constants.STRINGFORMATER_ObectHasNoAttribute_STRINGFORMATER_OrCannotBeAssignedTo, obj.Type(), t.Identifier.Value)
			}
		default:
			return object.NewErrorWithLocation(node.Token, constants.SyntaxError, constants.CannotAssignTo_STRINGFORMATER, t.String())
		}
	}

	return object.NULL // Assignments are statements and return NULL
}