package object

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer" // For token info
)

// GetObjectIterator tries to get an iterator object for a given Pylearn object.
// Follows Python's protocol: Check __iter__, then sequence types.
// Returns the Iterator object and nil error on success.
// Returns nil Iterator and an Error object on failure (e.g., TypeError).
// Requires ExecutionContext to potentially call __iter__.
func GetObjectIterator(ctx ExecutionContext, obj Object, tokenOnError lexer.Token) (Iterator, Object) { // Returns Iterator, Pylearn Error Object

	// 1. Check for __iter__ method on object.Instance
	// if inst, ok := obj.(*Instance); ok && inst.Class != nil { // Check non-nil Class
	// 	if iterMethod, methodOk := inst.Class.Methods[constants.DunderIter]; methodOk {
	// 		boundIter := &BoundMethod{Instance: inst, Method: iterMethod}
	if inst, ok := obj.(*Instance); ok && inst.Class != nil {
		if iterMethodObj, methodOk := inst.Class.Methods[constants.DunderIter]; methodOk {
			// <<< FIX: Check if the retrieved method is actually a function >>>
			if iterMethod, isFunc := iterMethodObj.(*Function); isFunc {
				boundIter := &BoundMethod{Instance: inst, Method: iterMethod}
				// Use the generic ApplyBoundMethod helper, passing the context
				iterObj := ApplyBoundMethod(ctx, boundIter, []Object{}, tokenOnError)
				if IsError(iterObj) {
					return nil, iterObj // Propagate error from __iter__ call
				}

				// Check if the returned object implements our internal Go Iterator interface
				if iterator, isIter := iterObj.(Iterator); isIter {
					return iterator, nil // Successfully got iterator from __iter__
				}

				// If it doesn't implement the Go interface, Python would check __next__.
				// For simplicity, require __iter__ to return an object satisfying object.Iterator.
				return nil, NewErrorWithLocation(tokenOnError, constants.TypeError, constants.ITERATOR_ITER_METHOD_RETURN_ERROR, iterObj.Type())
			}
		}
		// Fall through if instance has no __iter__
	}

	// 2. Check built-in sequence types and return a GenericIterator wrapper
	switch o := obj.(type) {
	case *List: // Reuse vm's iterator creation logic or define here
		return newListIterator(o), nil
	case *String:
		return newStringIterator(o), nil
	case *Tuple:
		return newTupleIterator(o), nil
	case *Dict: // Iterates keys
		return newDictIterator(o), nil
	case *Range:
		return newRangeIterator(o), nil
	case *Bytes: // Iterates ints
		return newBytesIterator(o), nil
	case *Set:
		return newSetIterator(o), nil
		// Add other iterable types like File if needed
	} // End switch o := obj.(type)

	// 3. Check if the object is already an iterator itself (implements the Go interface)
	if iterator, isIter := obj.(Iterator); isIter {
		return iterator, nil
	}

	// 4. Not iterable
	return nil, NewErrorWithLocation(tokenOnError, constants.TypeError, constants.ITERATOR_GET_OBJECT_ITER_ERROR, obj.Type())
}

// --- Iterator Creation Helpers (Copied/Adapted from VM or defined here) ---
// These create the GenericIterator wrappers for built-in types.

func newListIterator(list *List) *GenericIterator {
	idx := 0
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_LIST, NextFn: func() (Object, bool) {
		if idx >= len(list.Elements) {
			return nil, true
		}
		elem := list.Elements[idx]
		idx++
		return elem, false
	}}
}
func newTupleIterator(tuple *Tuple) *GenericIterator {
	idx := 0
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_TUPLE, NextFn: func() (Object, bool) {
		if idx >= len(tuple.Elements) {
			return nil, true
		}
		elem := tuple.Elements[idx]
		idx++
		return elem, false
	}}
}
func newStringIterator(str *String) *GenericIterator {
	runes := []rune(str.Value)
	idx := 0
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_STRING, NextFn: func() (Object, bool) {
		if idx >= len(runes) {
			return nil, true
		}
		r := runes[idx]
		idx++
		return &String{Value: string(r)}, false
	}}
}
func newRangeIterator(r *Range) *GenericIterator {
	current, step, stop := r.Start, r.Step, r.Stop
	finished := (step == 0) || (step > 0 && current >= stop) || (step < 0 && current <= stop)
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_RANGE, NextFn: func() (Object, bool) {
		if finished {
			return nil, true
		}
		val := current
		current += step
		if (step > 0 && current >= stop) || (step < 0 && current <= stop) {
			finished = true
		}
		return &Integer{Value: val}, false
	}}
}
func newDictIterator(dict *Dict) *GenericIterator {
	keys := make([]HashKey, 0, len(dict.Pairs))
	idx := 0
	for k := range dict.Pairs {
		keys = append(keys, k)
	}
	// TODO: Sort keys maybe for deterministic testing?
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_DICT_KEYS, NextFn: func() (Object, bool) {
		if idx >= len(keys) {
			return nil, true
		}
		keyHash := keys[idx]
		idx++
		// Important: Return the original key object, not the hash key struct
		pair, ok := dict.Pairs[keyHash]
		if !ok {
			return NewError(constants.InternalError, constants.ITERATOR_DICT_KEY_DISAPPEARED_ERROR), true
		} // Should not happen
		return pair.Key, false
	}}
}
func newBytesIterator(b *Bytes) *GenericIterator {
	idx := 0
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_BYTES, NextFn: func() (Object, bool) {
		if idx >= len(b.Value) {
			return nil, true
		}
		byteVal := b.Value[idx]
		idx++
		return &Integer{Value: int64(byteVal)}, false
	}}
}
func newSetIterator(s *Set) *GenericIterator {
	elems := make([]Object, 0, len(s.Elements))
	idx := 0
	for _, elem := range s.Elements {
		elems = append(elems, elem)
	}
	// TODO: Sort elements maybe for deterministic testing?
	return &GenericIterator{Source: constants.ITERATOR_SOURCE_SET, NextFn: func() (Object, bool) {
		if idx >= len(elems) {
			return nil, true
		}
		elem := elems[idx]
		idx++
		return elem, false
	}}
}

func UnpackIterator(it Iterator) ([]Object, error) {
	var items []Object
	for {
		item, stop := it.Next()
		if stop {
			break
		}
		if IsError(item) {
			return nil, fmt.Errorf("error during unpacking: %s", item.Inspect())
		}
		items = append(items, item)
	}
	return items, nil
}
