// Suggested package: internal/pyasync/runtime.go
// Or if it's more general: internal/asyncruntime/runtime.go

package asyncruntime // Or your chosen package name

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
)

// AsyncResult represents the result of an async operation
type AsyncResult struct {
	Value interface{}
	Error error
	Done  chan struct{}
	mu    sync.RWMutex
}

// NewAsyncResult creates a new AsyncResult
func NewAsyncResult() *AsyncResult {
	return &AsyncResult{
		Done: make(chan struct{}),
	}
}

// SetResult sets the result of the async operation
func (ar *AsyncResult) SetResult(value interface{}, err error) {
	ar.mu.Lock()
	// Prevent double close if SetResult is somehow called multiple times
	select {
	case <-ar.Done:
		// Already closed, do nothing or log a warning
		ar.mu.Unlock()
		return
	default:
		// Not closed yet
	}
	ar.Value = value
	ar.Error = err
	ar.mu.Unlock() // Unlock before closing channel to avoid race with GetResult RLock
	close(ar.Done)
}

// GetResult blocks until the result is available
func (ar *AsyncResult) GetResult() (interface{}, error) {
	<-ar.Done // Block until Done channel is closed
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	return ar.Value, ar.Error
}

// IsReady checks if the result is ready without blocking
func (ar *AsyncResult) IsReady() bool {
	select {
	case <-ar.Done:
		return true
	default:
		return false
	}
}

// Coroutine represents a suspended computation
type Coroutine struct {
	ID       int
	Function func(ctx context.Context) (interface{}, error) // The Go function to execute
	Result   *AsyncResult                                   // Where the result will be stored
	Context  context.Context                                // Context for this specific coroutine
	Cancel   context.CancelFunc                             // Function to cancel this coroutine's context
}

// EventLoop manages the execution of coroutines
type EventLoop struct {
	coroutines map[int]*Coroutine
	nextID     int
	mu         sync.RWMutex   // Mutex to protect coroutines map and nextID
	wg         sync.WaitGroup // To wait for all goroutines to finish
	ctx        context.Context    // Master context for the event loop itself
	cancel     context.CancelFunc // To cancel the event loop and all its coroutines
}

// NewEventLoop creates a new event loop
func NewEventLoop() *EventLoop {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventLoop{
		coroutines: make(map[int]*Coroutine),
		nextID:     1,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// CreateCoroutine creates a new coroutine and schedules it to run
// fn is the Go function that represents the body of the Pylearn async task.
func (el *EventLoop) CreateCoroutine(fn func(ctx context.Context) (interface{}, error)) *AsyncResult {
	el.mu.Lock()

	id := el.nextID
	el.nextID++

	result := NewAsyncResult()
	// Create a new context for this coroutine, derived from the event loop's master context.
	// This allows individual coroutines to be cancelled or for all to be cancelled if the loop stops.
	coroCtx, coroCancel := context.WithCancel(el.ctx)

	coro := &Coroutine{
		ID:       id,
		Function: fn,
		Result:   result,
		Context:  coroCtx,
		Cancel:   coroCancel,
	}

	el.coroutines[id] = coro
	el.mu.Unlock() // Unlock before starting goroutine to avoid holding lock during Add/runCoroutine

	el.wg.Add(1)
	go el.runCoroutine(coro)

	return result
}

// runCoroutine executes a coroutine's function in a new goroutine
func (el *EventLoop) runCoroutine(coro *Coroutine) {
	defer el.wg.Done()
	defer func() {
		// Clean up the coroutine from the map once it's done
		el.mu.Lock()
		delete(el.coroutines, coro.ID)
		el.mu.Unlock()
		coro.Cancel() // Cancel its context to free resources, although it should be done if Function returned
	}()

	// Execute the function
	value, err := coro.Function(coro.Context)
	coro.Result.SetResult(value, err)
}

// Stop signals the event loop to stop and waits for all active coroutines to finish.
func (el *EventLoop) Stop() {
	el.cancel() // Cancel the master context, which propagates to all derived coroutine contexts
	el.wg.Wait() // Wait for all goroutines started by runCoroutine to complete
}

// Wait waits for all currently scheduled coroutines to complete without stopping the loop.
// This might be useful if you schedule a batch and want to wait for them before scheduling more.
func (el *EventLoop) Wait() {
	el.wg.Wait()
}

// AsyncFunction (as defined by you) represents a Go function registered with the runtime
type AsyncFunction struct {
	Name     string
	Function func(ctx context.Context, args ...interface{}) (interface{}, error)
}


// Runtime provides the async/await runtime services
type Runtime struct {
	EventLoop *EventLoop
	// This map is for Go functions registered to be callable from Pylearn as "async functions"
	// The key is the Pylearn name.
	GoAsyncFunctions map[string]*AsyncFunction
}

// NewRuntime creates a new async runtime
func NewRuntime() *Runtime {
	return &Runtime{
		EventLoop:        NewEventLoop(),
		GoAsyncFunctions: make(map[string]*AsyncFunction),
	}
}

// RegisterGoAsyncFunction registers a Go function that can be called from Pylearn
// as if it were an async function.
func (r *Runtime) RegisterGoAsyncFunction(name string, fn func(ctx context.Context, args ...interface{}) (interface{}, error)) {
	r.GoAsyncFunctions[name] = &AsyncFunction{
		Name:     name,
		Function: fn,
	}
}

// CallGoAsyncFunction is used by Pylearn when it encounters a call to a registered Go async function.
// It schedules the Go function on the event loop.
// `pylearnArgs` would be `[]object.Object` that need conversion to `[]interface{}`.
func (r *Runtime) CallGoAsyncFunction(name string, pylearnArgs ...interface{}) *AsyncResult {
	// In a real integration, pylearnArgs would be []object.Object
	// and you'd convert them to []interface{} here before passing to fn.Function.
	// This is a simplified signature.

	if goAsyncFn, exists := r.GoAsyncFunctions[name]; exists {
		return r.EventLoop.CreateCoroutine(func(ctx context.Context) (interface{}, error) {
			// When calling the registered Go function, pass its specific context and converted args.
			return goAsyncFn.Function(ctx, pylearnArgs...)
		})
	}

	result := NewAsyncResult()
	result.SetResult(nil, fmt.Errorf(constants.AsyncRuntimeGoAsyncFunctionNotFound, name))
	return result
}

// Await waits for an async result (Pylearn 'await' keyword calls this through the interpreter)
func (r *Runtime) Await(result *AsyncResult) (interface{}, error) {
	// This directly uses the Go AsyncResult's GetResult, which blocks the current goroutine.
	// This is the "Model 2 (Blocking within Go Coroutine)" approach discussed earlier.
	return result.GetResult()
}

// Sleep provides an async sleep functionality callable from Pylearn (via a wrapper)
func (r *Runtime) Sleep(duration time.Duration) *AsyncResult {
	return r.EventLoop.CreateCoroutine(func(ctx context.Context) (interface{}, error) {
		select {
		case <-time.After(duration):
			return nil, nil // Pylearn None equivalent
		case <-ctx.Done(): // Handle cancellation
			return nil, ctx.Err()
		}
	})
}

// GatherAll waits for multiple async results (Pylearn `async_gather` or similar)
func (r *Runtime) GatherAll(results ...*AsyncResult) ([]interface{}, error) {
	if len(results) == 0 {
		return []interface{}{}, nil
	}
	values := make([]interface{}, len(results))
	// Use an error channel to collect the first error non-blockingly
	errChan := make(chan error, len(results)) // Buffered channel

	var wg sync.WaitGroup
	for i, result := range results {
		wg.Add(1)
		go func(idx int, res *AsyncResult) {
			defer wg.Done()
			val, err := res.GetResult() // This blocks this specific goroutine
			if err != nil {
				// Non-blockingly send error. If channel is full, this means another error was already sent.
				select {
				case errChan <- err:
				default:
				}
				return // Don't set value if there's an error
			}
			values[idx] = val
		}(i, result)
	}

	wg.Wait()
	close(errChan) // Close channel after all goroutines are done

	// Check if any error occurred
	for err := range errChan { // Read all errors sent (though usually only first matters for GatherAll)
		if err != nil {
			// Return all values collected so far (some might be nil/zero if their task errored)
			// and the first error encountered.
			return values, err
		}
	}
	return values, nil
}


// CreateCoroutine directly on Runtime, delegating to its EventLoop
func (r *Runtime) CreateCoroutine(fn func(ctx context.Context) (interface{}, error)) *AsyncResult {
	if r.EventLoop == nil { // Safety check
		// Handle error: maybe return an already failed AsyncResult
		res := NewAsyncResult()
		res.SetResult(nil, fmt.Errorf(constants.AsyncRuntimeEventLoopNotInitialized))
		return res
	}
	return r.EventLoop.CreateCoroutine(fn)
}