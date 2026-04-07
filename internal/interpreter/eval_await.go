// pylearn/internal/interpreter/eval_await.go
package interpreter

import (
	"context"
	"fmt"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

func evalAwaitExpression(node *ast.AwaitExpression, ctx *InterpreterContext) object.Object {
	// Step 1: Evaluate the expression that is being awaited.
	// This could be a call to an async function (which returns a coroutine), a Task, etc.
	awaitable := Eval(node.Expression, ctx)
	if object.IsError(awaitable) {
		return awaitable
	}

	asyncRuntime := ctx.GetAsyncRuntime()
	if asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, constants.AwaitUsedButNoAsyncRuntimeIsAvailable)
	}

	// This is the "future" we will ultimately block on.
	var asyncResultWrapper *object.AsyncResultWrapper

	// Step 2: Resolve the awaitable object to its underlying future (AsyncResultWrapper).
	// This loop handles the __await__ protocol.
	for i := 0; i < 100; i++ { // Loop with a safety break to prevent infinite __await__ chains
		// Case 1: It's already the low-level wrapper. We're done resolving.
		if arw, ok := awaitable.(*object.AsyncResultWrapper); ok {
			asyncResultWrapper = arw
			break
		}

		// Case 2: It's our high-level Task object. Get its internal wrapper.
		if task, ok := awaitable.(*object.Task); ok {
			asyncResultWrapper = task.ResultWrapper
			break
		}

		// Case 3: It's a raw coroutine object (`async def` function). Schedule it.
		if coro, ok := awaitable.(*object.Function); ok && coro.IsAsync {
			goCoroutineFunc := func(goCtx context.Context) (interface{}, error) {
				taskEnv := object.NewEnclosedEnvironment(coro.Env)
				taskCtx := ctx.NewChildContext(taskEnv)
				evalResult := taskCtx.EvaluateASTNode(coro.Body, nil)
				if retVal, isRet := evalResult.(*object.ReturnValue); isRet {
					return retVal.Value, nil
				}
				if errObj, isErr := evalResult.(*object.Error); isErr {
					return nil, fmt.Errorf(errObj.Inspect())
				}
				return evalResult, nil
			}
			asyncResult := asyncRuntime.CreateCoroutine(goCoroutineFunc)
			asyncResultWrapper = &object.AsyncResultWrapper{GoAsyncResult: asyncResult}
			break
		}

		// Case 4: It's another type of object. Check if it implements __await__.
		if getter, ok := awaitable.(object.AttributeGetter); ok {
			awaitMethod, found := getter.GetObjectAttribute(ctx, constants.DunderAwait)
			if found && awaitMethod != nil {
				// Call the __await__ method. Its result becomes the new awaitable
				// and we continue the loop to resolve it.
				awaitable = ctx.Execute(awaitMethod)
				if object.IsError(awaitable) {
					return awaitable // Error calling __await__
				}
				continue // Continue loop with the new awaitable
			}
		}

		// If none of the above, the object is not awaitable.
		return object.NewErrorWithLocation(node.Token, constants.TypeError, constants.InterpreterEvalObjectNotAwaitable, awaitable.Type())
	}

	// Step 3: Perform the actual blocking await on the Go-level future.
	if asyncResultWrapper == nil || asyncResultWrapper.GoAsyncResult == nil {
		// This could happen if the __await__ chain ends in a non-awaitable or nil.
		return object.NewErrorWithLocation(node.Token, constants.RuntimeError, constants.InterpreterEvalAsyncResultNilGoError)
	}

	goValue, goErr := asyncRuntime.Await(asyncResultWrapper.GoAsyncResult)
	if goErr != nil {
		// The Go coroutine returned an error. Convert it to a Pylearn error.
		return object.NewErrorFromGoErr(goErr)
	}

	// Step 4: Convert the Go result back to a Pylearn object.
	if goValue == nil {
		return object.NULL
	}
	if pyObj, isPyObj := goValue.(object.Object); isPyObj {
		// If the Go coroutine returned a Pylearn error object as its *value*,
		// we need to propagate it as a proper Pylearn error.
		if object.IsError(pyObj) {
			return pyObj
		}
		return pyObj
	}

	// This should not happen if our native functions are well-behaved.
	return object.NewErrorWithLocation(node.Token, constants.InternalError, constants.InterpreterEvalAwaitGoReturnError, goValue)
}