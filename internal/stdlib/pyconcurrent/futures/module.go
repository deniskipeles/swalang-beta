// internal/stdlib/pyconcurrent/futures/module.go
package pycfutures

import (
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/constants"
)

// Constructor for Pylearn ThreadPoolExecutor class
func pyThreadPoolExecutorConstructor(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var maxWorkersArg object.Object = object.NULL
	if len(args) > 1 { // Assuming Pylearn class constructors don't get `cls` as first arg here
		return object.NewError(constants.TypeError, "ThreadPoolExecutor constructor takes at most 1 argument (max_workers), got %d", len(args))
	}
	if len(args) == 1 {
		maxWorkersArg = args[0]
	}

	// Get a base environment from the current execution context.
	// This environment will be used by the ThreadPoolExecutor to create new
	// execution contexts for the tasks it runs.
	baseEnvForTasks := ctx.GetCurrentEnvironment()
	if baseEnvForTasks == nil {
		return object.NewError(constants.InternalError, "Could not get Pylearn environment from ExecutionContext for ThreadPoolExecutor")
	}

	// NewThreadPoolExecutor is the Go constructor for the Go struct
	executor, errObj := NewThreadPoolExecutor(maxWorkersArg, baseEnvForTasks)
	if errObj != nil { // errObj is already an *object.Error
		return errObj
	}
	return executor // executor is *pycfutures.ThreadPoolExecutor which is an object.Object
}

func init() {
	futuresEnv := object.NewEnvironment()

	threadPoolExecutorClass := &object.Builtin{
		Name: "ThreadPoolExecutor",
		Fn:   pyThreadPoolExecutorConstructor,
	}
	futuresEnv.Set("ThreadPoolExecutor", threadPoolExecutorClass)

	concurrentFuturesModule := &object.Module{
		Name: "futures",
		Path: "<builtin>.concurrent.futures",
		Env:  futuresEnv,
	}

	var concurrentPackage *object.Module
	existingConcurrentMod, found := object.GetNativeModule("concurrent")
	if found {
		if existingConcurrentMod == nil {
			panic("critical: native module 'concurrent' found but is nil in registry")
		}
		concurrentPackage = existingConcurrentMod
	} else {
		concurrentPackage = &object.Module{
			Name: "concurrent",
			Path: "<builtin>.concurrent",
			Env:  object.NewEnvironment(),
		}
		object.RegisterNativeModule("concurrent", concurrentPackage)
	}
	if concurrentPackage.Env == nil {
		panic("critical: concurrentPackage.Env is nil")
	}
	concurrentPackage.Env.Set("futures", concurrentFuturesModule)
}