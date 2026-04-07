// internal/stdlib/pytime/time_funcs.go
package pytime

import (
	"time" // Go's time package

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyTimeSleepFn implements time.sleep(seconds)
func pyTimeSleepFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "time.sleep() takes exactly 1 argument (seconds), got %d", len(args))
	}

	var seconds float64
	switch arg := args[0].(type) {
	case *object.Integer:
		seconds = float64(arg.Value)
	case *object.Float:
		seconds = arg.Value
	default:
		return object.NewError(constants.TypeError, "time.sleep() argument must be a number, not %s", args[0].Type())
	}

	if seconds < 0 {
		// Python's time.sleep raises InterruptedError if sleep is interrupted by a signal,
		// and ValueError for negative sleep times on some platforms.
		// For simplicity, let's make it a ValueError.
		return object.NewError(constants.ValueError, "sleep length must be non-negative")
	}

	// Convert float seconds to time.Duration
	duration := time.Duration(seconds * float64(time.Second))
	
	// Perform the sleep.
	// TODO: If integrating with your async event loop for cooperative multitasking,
	// this blocking sleep might stall the entire event loop.
	// An async-aware sleep would return an AsyncResult and use `PylearnAsyncRuntime.Sleep()`.
	// For a purely synchronous interpreter or for synchronous parts of an async one, this is fine.
	time.Sleep(duration)

	return object.NULL // time.sleep() returns None
}

// pyTimeTimeFn implements time.time()
// Returns the time in seconds since the Epoch as a floating point number.
func pyTimeTimeFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError(constants.TypeError, "time.time() takes no arguments, got %d", len(args))
	}

	// UnixNano returns nanoseconds. Convert to seconds with fractional part.
	now := time.Now().UnixNano()
	seconds := float64(now) / 1e9 // 1e9 nanoseconds in a second

	return &object.Float{Value: seconds}
}

// --- Builtin Objects for the time module ---
var (
	TimeSleep = &object.Builtin{Name: "time.sleep", Fn: pyTimeSleepFn}
	TimeTime  = &object.Builtin{Name: "time.time", Fn: pyTimeTimeFn}
)