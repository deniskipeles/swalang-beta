package builtins

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer" // For token info
	"github.com/deniskipeles/pylearn/internal/object"
)

// NOTE: These functions are highly dependent on a well-defined iterator protocol
// ( __iter__ method returning an iterator object, and __next__ method on the iterator).
// The implementations below are placeholders or simplified versions assuming basic types.

// --- iter() ---
// Accepts ExecutionContext
func pyIterFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	token := lexer.Token{Line: -1, Column: 0, Type: lexer.ILLEGAL, Literal: constants.BuiltinsIterTokenLiteral} // Placeholder token
	var obj object.Object

	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsIterArgCountError, len(args))
	}
	obj = args[0]

	if len(args) == 2 { // iter(callable, sentinel)
		sentinel := args[1]

		// Check if 'obj' is callable using the 'callable' builtin via context
		callableBuiltin, ok := Builtins[constants.BuiltinsCallableFuncName]
		if !ok {
			return object.NewError(constants.InternalError, constants.BuiltinsIterCallableBuiltinNotFound)
		}
		// Execute callable() using the provided context
		callableResult := ctx.Execute(callableBuiltin, obj)
		if object.IsError(callableResult) {
			// Propagate error from callable() check itself
			return callableResult
		}

		// callable() returns TRUE or FALSE object
		if callableResult != object.TRUE {
			// Use object.NewErrorWithLocation (using placeholder token)
			return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsIterVMustBeCallable)
		}

		// Create and return a callable iterator object
		return &object.GenericIterator{
			Source: constants.BuiltinsIterCallableSource,
			NextFn: func() (object.Object, bool) { // This closure still captures 'obj' and 'sentinel'
				// --- Execute the captured callable object using the context ---
				// This allows calling functions or instances with __call__ correctly
				result := ctx.Execute(obj /* callable */) // Call with NO arguments
				// --- End Execution via Context ---

				if object.IsError(result) { // Use object.IsError
					// How to handle errors from the callable? Stop iteration?
					// Python usually propagates the error.
					// For now, treat error as the value to check against sentinel,
					// or maybe return it directly to stop iteration?
					// Let's log and continue checking against sentinel for now.
					// A dedicated error return might be better long term.
					fmt.Printf(constants.BuiltinsIterWarningErrorIgnored, result.Inspect())
				}

				// Basic comparison using Inspect (improve with proper equality check later)
				// TODO: Replace with a proper object comparison helper (e.g., ctx.Equals(result, sentinel))
				isEqual := (result.Type() == sentinel.Type() && result.Inspect() == sentinel.Inspect())

				if isEqual {
					return nil, true // Stop iteration if result == sentinel
				}
				return result, false // Return the result
			},
		}
	}

	// iter(object) variant
	// Use the GetObjectIterator helper (which should ideally be moved out of interpreter)
	// For now, let's assume a similar helper exists or move it to 'object' package.
	// Replace interpreter.GetObjectIterator with object.GetObjectIterator
	iterator, errObj := object.GetObjectIterator(ctx, obj, token) // Pass context and token
	if errObj != nil {
		// errObj should already be an object.Error
		return errObj
	}
	return iterator
}

// --- next() ---
// Accepts ExecutionContext (unused but required by signature)
func pyNextFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	token := lexer.Token{Line: -1, Column: 0, Type: lexer.ILLEGAL, Literal: constants.BuiltinsNextTokenLiteral} // Placeholder token
	hasDefault := false
	var defaultVal object.Object

	if len(args) < 1 || len(args) > 2 {
		// Use object.NewErrorWithLocation
		return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsNextArgCountError, len(args))
	}

	iteratorArg := args[0]
	if len(args) == 2 {
		hasDefault = true
		defaultVal = args[1]
	}

	// Check if the argument is actually an iterator using the Go interface
	// This remains an internal Go check.
	iterator, ok := iteratorArg.(object.Iterator)
	if !ok {
		// Try checking for __next__ method if not implementing the interface? Complex.
		// For now, require our internal Iterator interface.
		// Use object.NewErrorWithLocation
		return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsNextObjectNotIterator, iteratorArg.Type())
	}

	// Call the iterator's internal Go Next method
	nextItem, stop := iterator.Next()

	if stop {
		// Iterator finished
		if hasDefault {
			return defaultVal // Return the default value
		}
		// No default, raise StopIteration
		// Return the *singleton* error instance defined in object.go
		return object.STOP_ITERATION
	}

	// Iterator returned an item
	// If nextItem itself is an error (e.g., from iter(callable)), propagate it
	if object.IsError(nextItem) {
		return nextItem
	}

	return nextItem
}

// --- Other Iterator Builtins (Placeholders) ---
// Update signatures to accept context

// pyEnumerateFn implements enumerate(iterable, start=0)
func pyEnumerateFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsEnumerateArgCountError, len(args))
	}

	iterableArg := args[0]
	startValue := int64(0) // Default start is 0

	if len(args) == 2 {
		if args[1] != object.NULL { // Allow None for start to mean default 0
			startObj, ok := args[1].(*object.Integer)
			if !ok {
				return object.NewError(constants.TypeError, constants.BuiltinsEnumerateStartArgTypeError, args[1].Type())
			}
			startValue = startObj.Value
		}
	}

	// Get an iterator for the input iterable.
	// Use a placeholder token for errors originating from GetObjectIterator itself,
	// as there isn't a specific Pylearn source token for this internal step.
	sourceIter, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		// errObj is already an *object.Error (TypeError: '...' object is not iterable)
		return errObj
	}

	// Create and return the EnumerateIterator object
	return &object.EnumerateIterator{
		SourceIterator: sourceIter,
		CurrentIndex:   startValue, // Initialize with the start value
	}
}

func pyZipFn(ctx object.ExecutionContext, args ...object.Object /* *iterables */) object.Object {
	// TODO: Implement zip(). Needs GetObjectIterator on all inputs, returns zip iterator object.
	return object.NewError(constants.NotImplementedError, constants.BuiltinsZipNotImplemented)
}

func pyMapFn(ctx object.ExecutionContext, args ...object.Object /* function, *iterables */) object.Object {
	// TODO: Implement map(). Needs callable check, GetObjectIterator on inputs, returns map iterator object. Uses ctx.Execute for function.
	return object.NewError(constants.NotImplementedError, constants.BuiltinsMapNotImplemented)
}

func pyFilterFn(ctx object.ExecutionContext, args ...object.Object /* function or None, iterable */) object.Object {
	// TODO: Implement filter(). Needs callable check, GetObjectIterator on input, returns filter iterator object. Uses ctx.Execute for function, object.IsTruthy for check.
	return object.NewError(constants.NotImplementedError, constants.BuiltinsFilterNotImplemented)
}

func pyReversedFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// TODO: Implement reversed(). Needs __reversed__ or (__len__ and __getitem__) on input, returns a reverse iterator. May need ctx.Execute.
	return object.NewError(constants.NotImplementedError, constants.BuiltinsReversedNotImplemented)
}

func pyAllFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsAllArgCountError, len(args))
	}
	// Use the new object.GetObjectIterator
	iterator, errObj := object.GetObjectIterator(ctx, args[0], object.NoToken)
	if errObj != nil {
		return errObj
	} // errObj is already an object.Object error

	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(item) {
			return item
		} // Propagate errors from iterator

		// Use IsTruthy helper - requires context if that helper needs it
		isTrue, truthErr := object.IsTruthy(ctx, item)
		if truthErr != nil {
			// --- FIX HERE ---
			// Convert the Go error 'truthErr' into a Pylearn error object
			// Check if it's already one of our error types
			if pyErr, ok := truthErr.(object.Object); ok && object.IsError(pyErr) {
				return pyErr // Return the existing Pylearn error
			}
			// Otherwise, wrap it
			return object.NewError(constants.RuntimeError, constants.BuiltinsAllIsTruthyPropagatedError, truthErr)
			// --- END FIX ---
		}
		if !isTrue {
			return object.FALSE // Short-circuit on first false item
		}
	}
	return object.TRUE // All items were true (or iterable was empty)
}

func pyAnyFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsAnyArgCountError, len(args))
	}
	// Use the new object.GetObjectIterator
	iterator, errObj := object.GetObjectIterator(ctx, args[0], object.NoToken)
	if errObj != nil {
		return errObj
	} // errObj is already an object.Object error

	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(item) {
			return item
		} // Propagate errors from iterator

		// Use IsTruthy helper - requires context if that helper needs it
		isTrue, truthErr := object.IsTruthy(ctx, item)
		if truthErr != nil {
			// --- FIX HERE ---
			// Convert the Go error 'truthErr' into a Pylearn error object
			if pyErr, ok := truthErr.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.BuiltinsAnyIsTruthyPropagatedError, truthErr)
			// --- END FIX ---
		}
		if isTrue {
			return object.TRUE // Short-circuit on first true item
		}
	}
	return object.FALSE // No items were true (or iterable was empty)
}

// --- Registration ---
// Ensure function signatures match object.BuiltinFunction
func init() {
	registerBuiltin(constants.BuiltinsIterFuncName, &object.Builtin{Fn: pyIterFn})
	registerBuiltin(constants.BuiltinsNextFuncName, &object.Builtin{Fn: pyNextFn})
	registerBuiltin(constants.BuiltinsEnumerateFuncName, &object.Builtin{Fn: pyEnumerateFn})
	registerBuiltin(constants.BuiltinsZipFuncName, &object.Builtin{Fn: pyZipFn})
	registerBuiltin(constants.BuiltinsMapFuncName, &object.Builtin{Fn: pyMapFn})
	registerBuiltin(constants.BuiltinsFilterFuncName, &object.Builtin{Fn: pyFilterFn})
	registerBuiltin(constants.BuiltinsReversedFuncName, &object.Builtin{Fn: pyReversedFn})
	registerBuiltin(constants.BuiltinsAllFuncName, &object.Builtin{Fn: pyAllFn})
	registerBuiltin(constants.BuiltinsAnyFuncName, &object.Builtin{Fn: pyAnyFn})
}

