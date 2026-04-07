package interpreter

import (
	"time"

	"github.com/deniskipeles/pylearn/internal/asyncruntime"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)


var PylearnAsyncRuntime *asyncruntime.Runtime

func InitAsyncRuntime() { // Call this from your main_interpreter.go or similar setup spot
	if PylearnAsyncRuntime == nil {
		PylearnAsyncRuntime = asyncruntime.NewRuntime()

		// Example: Register a Pylearn built-in that uses the async runtime
		// This would allow Pylearn code to do: result = await async_builtins.sleep(1)
		asyncBuiltinsEnv := object.NewEnvironment()
		asyncBuiltinsEnv.Set(constants.BuiltinsSleepFuncName, &object.Builtin{
			Name: constants.BuiltinsAsyncBuiltinsSleepFuncName,
			Fn: func(ctx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 1 {
					return object.NewError(constants.TypeError, constants.InterpreterAsyncSetupSleepArgCountError)
				}
				durationObj, ok := args[0].(*object.Integer) // Or Float
				if !ok {
					return object.NewError(constants.TypeError, constants.InterpreterAsyncSetupSleepDurationTypeError)
				}
				seconds := time.Duration(durationObj.Value) * time.Second
				goAsyncResult := PylearnAsyncRuntime.Sleep(seconds) // Uses your Go runtime's Sleep
				return &object.AsyncResultWrapper{GoAsyncResult: goAsyncResult}
			},
		})
		// Register this as a module or directly in builtins if preferred
		object.RegisterNativeModule(constants.BuiltinsAsyncBuiltinsModule, &object.Module{
			Name: constants.BuiltinsAsyncBuiltinsModule,
			Path: constants.InterpreterAsyncSetupModulePath,
			Env:  asyncBuiltinsEnv,
		})
	}
}