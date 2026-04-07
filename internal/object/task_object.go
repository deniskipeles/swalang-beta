// pylearn/internal/object/task_object.go
package object

import (
	"fmt"
	"github.com/deniskipeles/pylearn/internal/constants"
)

const (
	TASK_OBJ ObjectType = "Task"
)

// Task wraps a coroutine and its result, making it a manageable unit of work.
// It is an awaitable object.
type Task struct {
	Coroutine     *Function           // The Pylearn function object being run.
	ResultWrapper *AsyncResultWrapper // The future that the event loop is working on.
}

func (t *Task) Type() ObjectType { return TASK_OBJ }
func (t *Task) Inspect() string {
	status := "pending"
	if t.ResultWrapper != nil && t.ResultWrapper.GoAsyncResult.IsReady() {
		_, err := t.ResultWrapper.GoAsyncResult.GetResult()
		if err != nil {
			status = "finished with error"
		} else {
			status = "finished"
		}
	}
	return fmt.Sprintf("<Task status=%s coroutine=%s>", status, t.Coroutine.Name)
}

// GetObjectAttribute makes the Task object behave like a Python asyncio.Task
func (t *Task) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	switch name {
	// The __await__ dunder method is what makes an object usable with `await`.
	// It must return an iterator, but in our simplified model, the interpreter's
	// `await` logic looks for an AsyncResultWrapper directly. We return it here.
	case constants.DunderAwait:
		return t.ResultWrapper, true

		// Add other Task methods for compatibility with Python's asyncio.Task
		// case "done":
		// 	// return a bound method that checks t.ResultWrapper.GoAsyncResult.IsReady()
		// case "result":
		// 	// return a bound method that blocks and gets the result (or raises)
		// case "exception":
		// 	// return a bound method that gets the exception (or None)
		// case "cancel":
		// 	// return a bound method to cancel the underlying Go coroutine
	}
	return nil, false
}

var _ Object = (*Task)(nil)
var _ AttributeGetter = (*Task)(nil)
