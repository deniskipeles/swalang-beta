// internal/object/iterators.go (or internal/object/enumerate_object.go)
package object

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
	// "github.com/deniskipeles/pylearn/internal/lexer" // For NoToken if used
)


const ENUMERATE_ITERATOR_OBJ ObjectType = constants.ENUMERATE_ITERATOR_OBJ_TYPE

// EnumerateIterator holds the state for an enumerate() call.
type EnumerateIterator struct {
	SourceIterator Iterator // The iterator for the underlying iterable
	CurrentIndex   int64    // The current count value, starts from 'start'
	// No need to store the original 'start' value if CurrentIndex is initialized correctly.
}

func (ei *EnumerateIterator) Type() ObjectType { return ENUMERATE_ITERATOR_OBJ }

func (ei *EnumerateIterator) Inspect() string {
	// Try to get source info if available (e.g., from GenericIterator)
	sourceInfo := constants.ENUMERATE_ITERATOR_UNKNOWN_ITERABLE
	if gi, ok := ei.SourceIterator.(*GenericIterator); ok {
		sourceInfo = gi.Source
	}
	return fmt.Sprintf(constants.ENUMERATE_ITERATOR_INSPECT_FORMAT,
		sourceInfo, ei, ei.CurrentIndex)
}

// Next yields (count, value) pairs as Tuples.
func (ei *EnumerateIterator) Next() (Object, bool) {
	// Get the next value from the source iterator
	value, stop := ei.SourceIterator.Next()

	if stop { // Source iterator is exhausted
		return nil, true // Signal stop to our caller
	}

	// If the source iterator yielded an error, propagate it
	if IsError(value) {
		return value, true // Signal stop and pass the error
	}

	// Create the (count, value) pair as a Pylearn Tuple
	countObj := &Integer{Value: ei.CurrentIndex}
	pairTuple := &Tuple{Elements: []Object{countObj, value}}

	// Increment count for the next iteration
	ei.CurrentIndex++

	return pairTuple, false // Return the pair, not stopped
}

var _ Object = (*EnumerateIterator)(nil)
var _ Iterator = (*EnumerateIterator)(nil) // EnumerateIterator is itself an Iterator