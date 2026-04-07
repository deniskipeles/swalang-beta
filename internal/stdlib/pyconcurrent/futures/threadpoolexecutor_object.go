// internal/stdlib/pyconcurrent/futures/threadpoolexecutor_object.go
package pycfutures

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

const THREAD_POOL_EXECUTOR_OBJ object.ObjectType = "ThreadPoolExecutor"

type ThreadPoolExecutor struct {
	maxWorkers int
	workQueue  chan func()
	wg         sync.WaitGroup
	mu         sync.Mutex
	shutdown   bool
	// Store the base environment from which tasks should derive their contexts.
	// This is set during ThreadPoolExecutor creation.
	basePylearnEnvForTasks *object.Environment // Renamed for clarity
}

// NewThreadPoolExecutor Go constructor (called by Pylearn constructor)
// It now takes baseEnv directly.
func NewThreadPoolExecutor(pylearnMaxWorkers object.Object, baseEnvForTasks *object.Environment) (*ThreadPoolExecutor, object.Object) {
	if baseEnvForTasks == nil {
		// This is a critical issue if the constructor didn't get a valid base environment.
		return nil, object.NewError(constants.InternalError, "ThreadPoolExecutor requires a base environment for tasks")
	}

	maxW := runtime.NumCPU()
	if pylearnMaxWorkers != object.NULL && pylearnMaxWorkers != nil {
		mwInt, ok := pylearnMaxWorkers.(*object.Integer)
		if !ok {
			return nil, object.NewError(constants.TypeError, "ThreadPoolExecutor max_workers must be an int or None, not %s", pylearnMaxWorkers.Type())
		}
		val := int(mwInt.Value)
		if val <= 0 {
			return nil, object.NewError(constants.ValueError, "ThreadPoolExecutor max_workers must be greater than 0")
		}
		maxW = val
	}

	executor := &ThreadPoolExecutor{
		maxWorkers:             maxW,
		workQueue:              make(chan func(), maxW*2),
		basePylearnEnvForTasks: baseEnvForTasks, // Store the passed environment
	}
	executor.startWorkers()
	return executor, nil
}

func (e *ThreadPoolExecutor) Type() object.ObjectType { return THREAD_POOL_EXECUTOR_OBJ }
// ... (Inspect and startWorkers remain the same) ...
func (e *ThreadPoolExecutor) Inspect() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	state := "running"
	if e.shutdown {
		state = "shutdown"
	}
	return fmt.Sprintf("<ThreadPoolExecutor state=%s max_workers=%d at %p>", state, e.maxWorkers, e)
}

func (e *ThreadPoolExecutor) startWorkers() {
	for i := 0; i < e.maxWorkers; i++ {
		e.wg.Add(1)
		go func(workerID int) {
			defer e.wg.Done()
			for task := range e.workQueue {
				if task != nil {
					task()
				}
			}
		}(i)
	}
}


// Pylearn: executor.submit(fn, *args, **kwargs)
func (e *ThreadPoolExecutor) PySubmit(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// args[0] is self (ThreadPoolExecutor), args[1] is fn, args[2:] are pylArgs

	if len(args) < 2 {
		return object.NewError(constants.TypeError, "submit() missing 1 required argument: 'fn'")
	}
	pylFnObj := args[1] // Changed name for clarity
	pylArgs := args[2:]

	// Ensure pylFnObj is a callable Pylearn function
	pylFn, ok := pylFnObj.(*object.Function) // For now, only accept *object.Function
	if !ok {
		// TODO: Support other callable types like *object.Builtin or *object.BoundMethod
		return object.NewError(constants.TypeError, "first argument to submit() must be a Pylearn Function object, got %s", pylFnObj.Type())
	}

	e.mu.Lock()
	if e.shutdown {
		e.mu.Unlock()
		return object.NewError(constants.RuntimeError, "cannot schedule new futures after shutdown")
	}
	e.mu.Unlock()

	future := NewFuture() // This is *pycfutures.Future

	// Capture variables for the task closure
	// The environment for the *new task* should be derived from the function's
	// own closure environment (pylFn.Env) and the base environment stored in the executor.
	// A Pylearn function called in a new thread should typically run with its own
	// closure environment as the primary local scope, and that closure env's outer
	// should be the module/global scope where it was defined.
	
	// The 'basePylearnEnvForTasks' is used to create a *new ExecutionContext*.
	// The actual environment for the function call (`taskCallEnv`) needs to be
	// `NewEnclosedEnvironment(pylFn.Env)`.

	taskBaseEnv := e.basePylearnEnvForTasks // This is the env passed when TPE was created (e.g. global)
	if pylFn.Env != nil { // If the submitted function has its own closure environment
	    // This taskCallEnv will have pylFn.Env as local, and pylFn.Env.Outer (e.g. global) as its outer.
		// This is standard for function calls.
	    // taskBaseEnv = pylFn.Env // This was a misunderstanding; applyFunctionOrClass handles this.
	}


	task := func() {
		// Create a new ExecutionContext for this specific task.
		// The environment for this new context should be one suitable for executing
		// the top-level of the submitted Pylearn function.
		// If pylFn is a global function, its pylFn.Env is the global env.
		// We'll use the `basePylearnEnvForTasks` to create the initial context.
		// The actual execution of the function will then use `pylFn.Env` correctly via `applyFunctionOrClass`.
		taskExecCtx := ctx.NewChildContext(taskBaseEnv)
        if taskExecCtx == nil {
             future.SetException(object.NewError(constants.InternalError, "Failed to create child execution context for task"))
             return
        }


		// Call the Pylearn function using the new ExecutionContext.
		// This Execute call will internally use `applyFunctionOrClass` (or similar)
		// which sets up the correct enclosed environment based on `pylFn.Env`.
		resultObj := taskExecCtx.Execute(pylFn, pylArgs...)

		if errObj, isErr := resultObj.(*object.Error); isErr {
			future.SetException(errObj)
		} else {
			future.SetResult(resultObj)
		}
	}

	e.workQueue <- task
	return future
}

// ... (PyShutdown and GetObjectAttribute remain largely the same, ensuring they use ctx) ...
func (e *ThreadPoolExecutor) PyShutdown(ctx object.ExecutionContext, args ...object.Object) object.Object {
	wait := true
	if len(args) == 2 {
		if waitBool, ok := args[1].(*object.Boolean); ok {
			wait = waitBool.Value
		} else if args[1] != object.NULL {
			return object.NewError(constants.TypeError, "shutdown() wait argument must be bool or None")
		}
	} else if len(args) > 2 {
		return object.NewError(constants.TypeError, "shutdown() takes at most 1 argument (wait)")
	}

	e.mu.Lock()
	if !e.shutdown {
		e.shutdown = true
		close(e.workQueue) 
	}
	e.mu.Unlock()

	if wait {
		e.wg.Wait() 
	}
	return object.NULL
}

func (e *ThreadPoolExecutor) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	makeExecutorMethod := func(methodName string, goFn object.BuiltinFunction) *object.Builtin {
		return &object.Builtin{
			Name: "ThreadPoolExecutor." + methodName,
			Fn: func(callCtx object.ExecutionContext, scriptProvidedArgs ...object.Object) object.Object {
				methodArgs := make([]object.Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, e) 
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}
	switch name {
	case "submit":
		return makeExecutorMethod("submit", e.PySubmit), true
	case "shutdown":
		return makeExecutorMethod("shutdown", e.PyShutdown), true
	}
	return nil, false
}


var _ object.Object = (*ThreadPoolExecutor)(nil)
var _ object.AttributeGetter = (*ThreadPoolExecutor)(nil)