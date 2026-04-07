// internal/stdlib/pyaio/aio_funcs.go
package pyaio

import (
	"context"
	"fmt"
	"time"

	"github.com/deniskipeles/pylearn/internal/asyncruntime"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/interpreter" // For PylearnAsyncRuntime
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyAioSleepFn implements aio.sleep(duration_seconds)
// It returns an AsyncResultWrapper.
func pyAioSleepFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.StdlibAioSleepArgCountError, len(args))
	}

	var seconds float64
	switch arg := args[0].(type) {
	case *object.Integer:
		seconds = float64(arg.Value)
	case *object.Float:
		seconds = arg.Value
	default:
		return object.NewError(constants.TypeError, constants.StdlibAioSleepDurationTypeError, args[0].Type())
	}

	if seconds < 0 {
		return object.NewError(constants.ValueError, constants.StdlibAioSleepNegativeDurationError)
	}

	duration := time.Duration(seconds * float64(time.Second))

	// Ensure PylearnAsyncRuntime is initialized and available
	if interpreter.PylearnAsyncRuntime == nil || interpreter.PylearnAsyncRuntime.EventLoop == nil {
		return object.NewError(constants.RuntimeError, constants.StdlibAioSleepRuntimeNotInitialized)
	}

	// Use the Sleep method from your Go asyncruntime, which returns *asyncruntime.AsyncResult
	goAsyncResult := interpreter.PylearnAsyncRuntime.Sleep(duration)

	// Wrap the Go AsyncResult in a Pylearn AsyncResultWrapper
	return &object.AsyncResultWrapper{GoAsyncResult: goAsyncResult}
}

// pyAioGatherFn implements aio.gather(*awaitables)
// It takes a variable number of Pylearn AsyncResultWrapper objects.
// It returns a new AsyncResultWrapper that will resolve to a Pylearn List of results.
func pyAioGatherFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) == 0 {
		// Return an AsyncResultWrapper that immediately resolves to an empty Pylearn List
		resolvedResult := asyncruntime.NewAsyncResult()
		resolvedResult.SetResult(&object.List{Elements: []object.Object{}}, nil) // Resolve with empty Pylearn List
		return &object.AsyncResultWrapper{GoAsyncResult: resolvedResult}
	}

	goAsyncResults := make([]*asyncruntime.AsyncResult, 0, len(args))

	for i, arg := range args {
		pylearnAwaitable, ok := arg.(*object.AsyncResultWrapper)
		if !ok {
			return object.NewError(constants.TypeError, constants.StdlibAioGatherAwaitableTypeError, arg.Type(), i)
		}
		if pylearnAwaitable.GoAsyncResult == nil {
			return object.NewError(constants.ValueError, constants.StdlibAioGatherNilGoAsyncResult, i)
		}
		goAsyncResults = append(goAsyncResults, pylearnAwaitable.GoAsyncResult)
	}

	// Ensure PylearnAsyncRuntime is initialized
	if interpreter.PylearnAsyncRuntime == nil || interpreter.PylearnAsyncRuntime.EventLoop == nil {
		return object.NewError(constants.RuntimeError, constants.StdlibAioGatherRuntimeNotInitialized)
	}

	// Call the GatherAll method from your Go asyncruntime.
	// This itself returns an *asyncruntime.AsyncResult whose value will be []interface{} (or error).
	// We need to wrap this call in *another* coroutine to handle the conversion of []interface{} to Pylearn List.
	finalResultPromise := interpreter.PylearnAsyncRuntime.EventLoop.CreateCoroutine(
		func(goCoroutineCtx context.Context) (interface{}, error) {
			// This Go function runs when the "outer" gather operation is scheduled.
			// It calls the runtime's GatherAll which internally waits for all individual results.
			gatheredGoValues, err := interpreter.PylearnAsyncRuntime.GatherAll(goAsyncResults...)
			if err != nil {
				// err from GatherAll is a Go error (likely the first error encountered).
				// It needs to be returned as such, so the outer AsyncResultWrapper reflects it.
				return nil, err // Propagate the Go error
			}

			// `gatheredGoValues` is []interface{}. Each element should be an object.Object
			// if the individual coroutines resolved with Pylearn objects.
			pylearnResultsList := make([]object.Object, len(gatheredGoValues))
			for i, goVal := range gatheredGoValues {
				if pyObj, ok := goVal.(object.Object); ok {
					pylearnResultsList[i] = pyObj
				} else if goVal == nil { // If an individual coroutine returned (nil, nil)
					pylearnResultsList[i] = object.NULL
				} else {
					// This means an underlying Go coroutine returned a non-Pylearn object and non-nil error
					// which is unexpected if our async Pylearn functions always resolve to Pylearn objects.
					return nil, fmt.Errorf(constants.StdlibAioGatherUnexpectedGoType, goVal, i)
				}
			}
			return &object.List{Elements: pylearnResultsList}, nil // Resolve with a Pylearn List
		},
	)

	return &object.AsyncResultWrapper{GoAsyncResult: finalResultPromise}
}

// _create_task(coro) -> returns a Pylearn Task object
// This is the bridge between a Pylearn coroutine and the Go event loop.
func pyAioCreateTask(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "_create_task takes exactly one argument (a coroutine function)")
	}
	coro, ok := args[0].(*object.Function)
	if !ok || !coro.IsAsync {
		return object.NewError(constants.TypeError, "_create_task argument must be an async function")
	}

	asyncRuntime := ctx.GetAsyncRuntime()
	if asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, "Cannot create task: async runtime not available.")
	}

	// This is the Go function that the event loop will execute in a goroutine.
	// It captures the execution context and the Pylearn coroutine.
	goCoroutineFunc := func(goCtx context.Context) (interface{}, error) {
		// Create a new Pylearn execution context for this task.
		// It inherits the closure environment from the coroutine function.
		taskEnv := object.NewEnclosedEnvironment(coro.Env)
		taskCtx := ctx.NewChildContext(taskEnv)
		
		// Execute the Pylearn coroutine's body.
		evalResult := taskCtx.EvaluateASTNode(coro.Body, nil)

		// Unwrap the result for the Go async runtime.
		if retVal, isRet := evalResult.(*object.ReturnValue); isRet {
			return retVal.Value, nil
		}
		if errObj, isErr := evalResult.(*object.Error); isErr {
			// Convert Pylearn error to Go error to be handled by the runtime.
			return nil, fmt.Errorf(errObj.Inspect())
		}
		return object.NULL, nil
	}

	// Schedule the Go function on the event loop.
	asyncResult := asyncRuntime.CreateCoroutine(goCoroutineFunc)

	// Create and return the high-level Pylearn Task object.
	task := &object.Task{
		Coroutine:     coro,
		ResultWrapper: &object.AsyncResultWrapper{GoAsyncResult: asyncResult},
	}
	return task
}

// _run_and_wait(task) -> blocks until the task is complete and returns its result.
func pyAioRunAndWait(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "_run_and_wait takes exactly one argument (a Task)")
	}
	task, ok := args[0].(*object.Task)
	if !ok {
		return object.NewError(constants.TypeError, "_run_and_wait argument must be a Task object")
	}

	asyncRuntime := ctx.GetAsyncRuntime()
	if asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, "Cannot run task: async runtime not available.")
	}

	// This is the blocking call.
	goValue, goErr := asyncRuntime.Await(task.ResultWrapper.GoAsyncResult)
	
	if goErr != nil {
		// Create a Pylearn error from the Go error.
		return object.NewError(constants.RuntimeError, goErr.Error())
	}

	if pyObj, isPyObj := goValue.(object.Object); isPyObj {
		return pyObj
	}
	
	return object.NULL
}


// --- Builtin Objects for the aio module ---
var (
	AioSleep  = &object.Builtin{Name: constants.StdlibAioDotSleepFuncName, Fn: pyAioSleepFn}
	AioGather = &object.Builtin{Name: constants.StdlibAioDotGatherFuncName, Fn: pyAioGatherFn}

	AioCreateTask = &object.Builtin{Name: "aio._create_task", Fn: pyAioCreateTask}
	AioRunAndWait = &object.Builtin{Name: "aio._run_and_wait", Fn: pyAioRunAndWait}
)