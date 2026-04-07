// internal/stdlib/pyconcurrent/futures/future_object.go
package pycfutures // pyconcurrent.futures

import (
	"fmt"
	"sync"

	// "time"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

const (
	FUTURE_OBJ object.ObjectType = "Future"
	// Potential states (optional to expose as constants)
	// PENDING   = "PENDING"
	// RUNNING   = "RUNNING"
	// FINISHED  = "FINISHED"
	// CANCELLED = "CANCELLED"
)

type Future struct {
	mu         sync.Mutex
	cond       *sync.Cond // For blocking on result()
	done       bool
	result     object.Object // Pylearn result
	exception  object.Object // Pylearn error object if task failed
	cancelled  bool
	// Could add running_state, waiters list etc. for more advanced features
}

func NewFuture() *Future {
	f := &Future{}
	f.cond = sync.NewCond(&f.mu)
	return f
}

func (f *Future) Type() object.ObjectType { return FUTURE_OBJ }
func (f *Future) Inspect() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	state := "pending"
	if f.done {
		if f.exception != nil {
			state = fmt.Sprintf("finished with exception=%s", f.exception.Inspect())
		} else {
			state = fmt.Sprintf("finished with result=%s", f.result.Inspect())
		}
	} else if f.cancelled {
		state = "cancelled"
	}
	return fmt.Sprintf("<Future state=%s at %p>", state, f)
}

// Methods for Pylearn Future object (exposed via GetObjectAttribute)

// Pylearn: future.done()
func (f *Future) PyDone(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 { // self
		return object.NewError(constants.TypeError, "done() takes no arguments (%d given)", len(args)-1)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return object.NativeBoolToBooleanObject(f.done)
}

// Pylearn: future.cancelled()
func (f *Future) PyCancelled(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 { // self
		return object.NewError(constants.TypeError, "cancelled() takes no arguments (%d given)", len(args)-1)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return object.NativeBoolToBooleanObject(f.cancelled)
}

// Pylearn: future.result(timeout=None)
func (f *Future) PyResult(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 2 { // self, [timeout]
		return object.NewError(constants.TypeError, "result() takes at most 1 argument (timeout), got %d", len(args)-1)
	}
	// TODO: Implement timeout logic using time.After and a select with f.cond.Wait()
	// For now, blocking indefinitely.

	f.mu.Lock()
	defer f.mu.Unlock()

	for !f.done && !f.cancelled {
		f.cond.Wait() // Wait until done or cancelled
	}

	if f.cancelled {
		return object.NewError(constants.CancelledError, "Future was cancelled") // Or a specific Pylearn CancelledError
	}
	if f.exception != nil {
		return f.exception // Return the Pylearn error object
	}
	return f.result
}

// Pylearn: future.exception(timeout=None)
func (f *Future) PyException(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// Similar to result(), but returns exception or None
	if len(args) > 2 {
		return object.NewError(constants.TypeError, "exception() takes at most 1 argument (timeout), got %d", len(args)-1)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for !f.done && !f.cancelled {
		f.cond.Wait()
	}
	if f.cancelled {
		return object.NewError(constants.CancelledError, "Future was cancelled")
	}
	if f.exception != nil {
		return f.exception
	}
	return object.NULL // No exception
}

// Internal Go methods to set result/exception
func (f *Future) SetResult(res object.Object) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done || f.cancelled {
		return // Already completed or cancelled
	}
	f.result = res
	f.done = true
	f.cond.Broadcast() // Wake up all waiters
}

func (f *Future) SetException(exc object.Object) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done || f.cancelled {
		return
	}
	f.exception = exc
	f.done = true
	f.cond.Broadcast()
}

// GetObjectAttribute for Future
func (f *Future) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	makeFutureMethod := func(methodName string, goFn object.BuiltinFunction) *object.Builtin {
		return &object.Builtin{
			Name: "Future." + methodName,
			Fn: func(callCtx object.ExecutionContext, scriptProvidedArgs ...object.Object) object.Object {
				methodArgs := make([]object.Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, f) // Prepend self
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}
	switch name {
	case "done":
		return makeFutureMethod("done", f.PyDone), true
	case "cancelled":
		return makeFutureMethod("cancelled", f.PyCancelled), true
	case "result":
		return makeFutureMethod("result", f.PyResult), true
	case "exception":
		return makeFutureMethod("exception", f.PyException), true
		// TODO: add_done_callback, cancel
	}
	return nil, false
}

var _ object.Object = (*Future)(nil)
var _ object.AttributeGetter = (*Future)(nil)