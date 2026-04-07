package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// Tuple (Immutable Sequence, Hashable if elements are)
type Tuple struct {
	Elements []Object
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }
func (t *Tuple) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, len(t.Elements))
	for i, e := range t.Elements {
		elements[i] = e.Inspect()
	}
	out.WriteString(constants.TUPLE_INSPECT_OPEN_PAREN)
	out.WriteString(strings.Join(elements, constants.TUPLE_INSPECT_SEPARATOR))
	// Python prints a trailing comma for single-element tuples: (1,)
	if len(t.Elements) == 1 {
		out.WriteString(constants.TUPLE_INSPECT_TRAILING_COMMA)
	}
	out.WriteString(constants.TUPLE_INSPECT_CLOSE_PAREN)
	return out.String()
}

// HashKey for Tuple (hashable if all elements are hashable)
func (t *Tuple) HashKey() (HashKey, error) {
	// Algorithm similar to Python's tuple hash:
	// https://github.com/python/cpython/blob/main/Objects/tupleobject.c#L433
	var acc uint64 = 0x345678 // Arbitrary starting prime
	mult := uint64(1000003)   // Another prime
	for _, elem := range t.Elements {
		hashableElem, ok := elem.(Hashable)
		if !ok {
			return HashKey{}, fmt.Errorf(constants.TUPLE_HASH_UNHASHABLE_ERROR, elem.Type())
		}
		elemKey, err := hashableElem.HashKey()
		if err != nil {
			// This error comes from the element's HashKey method
			return HashKey{}, fmt.Errorf(constants.TUPLE_HASH_FAILED_ERROR, err)
		}
		acc = (acc ^ elemKey.Value) * mult
		mult += uint64(82520 + len(t.Elements)*2) // Vary multiplier
	}
	acc += 97531 // Final mixing
	return HashKey{Type: t.Type(), Value: acc}, nil
}

var _ Object = (*Tuple)(nil)
var _ Hashable = (*Tuple)(nil)

// --- ItemGetter for Tuple ---
func (t *Tuple) GetObjectItem(key Object) Object { // obj[key]
	idxObj, ok := key.(*Integer)
	if !ok {
		return NewError(constants.TypeError, constants.TUPLE_ITEM_INDEX_TYPE_ERROR, key.Type())
	}
	idx := idxObj.Value
	tupleLen := int64(len(t.Elements))

	if idx < 0 { // Handle negative indexing
		idx += tupleLen
	}

	if idx < 0 || idx >= tupleLen {
		return NewError(constants.IndexError, constants.TUPLE_ITEM_INDEX_OUT_OF_RANGE)
	}
	elem := t.Elements[idx]
	if elem == nil { // Should not happen with valid Pylearn objects
		return NULL
	}
	return elem
}

var _ ItemGetter = (*Tuple)(nil)

// --- Go functions for Tuple methods (to be wrapped by AttributeGetter) ---

// tuple_count(self, value)
func pyTupleCountFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.TUPLE_COUNT_ARG_COUNT_ERROR, len(args)-1)
	}
	selfTuple, ok := args[0].(*Tuple)
	if !ok {
		return NewError(constants.TypeError, constants.TUPLE_COUNT_ON_TUPLE_ERROR, args[0].Type())
	}
	valueToCount := args[1]
	count := 0
	for _, item := range selfTuple.Elements {
		// Use CompareObjects for equality
		comparisonResult := CompareObjects(constants.EqualsOperator, item, valueToCount, ctx) // Pass context
		if IsError(comparisonResult) {
			return comparisonResult // Propagate error from comparison
		}
		if comparisonResult == TRUE {
			count++
		}
	}
	return &Integer{Value: int64(count)}
}

// tuple_index(self, value, start=0, end=len(tuple))
func pyTupleIndexFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) < 2 || len(args) > 4 {
		return NewError(constants.TypeError, constants.TUPLE_INDEX_ARG_COUNT_ERROR, len(args)-1)
	}
	selfTuple, ok := args[0].(*Tuple)
	if !ok {
		return NewError(constants.TypeError, constants.TUPLE_INDEX_ON_TUPLE_ERROR, args[0].Type())
	}
	valueToFind := args[1]

	tupleLen := len(selfTuple.Elements)
	startIdx, endIdx := 0, tupleLen

	if len(args) >= 3 { // start provided
		startInt, okStart := args[2].(*Integer)
		if !okStart {
			return NewError(constants.TypeError, constants.TUPLE_INDEX_SLICE_INDICES_ERROR, "start", args[2].Type())
		}
		startIdx = int(startInt.Value)
	}
	if len(args) == 4 { // end provided
		endInt, okEnd := args[3].(*Integer)
		if !okEnd {
			return NewError(constants.TypeError, constants.TUPLE_INDEX_SLICE_INDICES_ERROR, "end", args[3].Type())
		}
		endIdx = int(endInt.Value)
	}

	// Python slice semantics for start/end for index
	if startIdx < 0 {
		startIdx = tupleLen + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	// if startIdx > tupleLen { startIdx = tupleLen } // If start > len, Python find/index returns error or empty

	if endIdx < 0 {
		endIdx = tupleLen + endIdx
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > tupleLen {
		endIdx = tupleLen
	}

	if startIdx >= tupleLen || startIdx >= endIdx { // If start is beyond or at the end, or slice is empty/invalid
		return NewError(constants.ValueError, constants.TUPLE_INDEX_VALUE_NOT_IN_TUPLE)
	}

	for i := startIdx; i < endIdx; i++ {
		item := selfTuple.Elements[i]
		comparisonResult := CompareObjects(constants.EqualsOperator, item, valueToFind, ctx)
		if IsError(comparisonResult) {
			return comparisonResult
		}
		if comparisonResult == TRUE {
			return &Integer{Value: int64(i)}
		}
	}
	return NewError(constants.ValueError, constants.TUPLE_INDEX_VALUE_NOT_IN_TUPLE)
}

// tuple_len(self)
func pyTupleLenFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 {
		return NewError(constants.TypeError, constants.TUPLE_LEN_ON_TUPLE_ERROR)
	}
	selfTuple, ok := args[0].(*Tuple)
	if !ok {
		return NewError(constants.TypeError, constants.TUPLE_LEN_ON_TUPLE_ERROR)
	}
	return &Integer{Value: int64(len(selfTuple.Elements))}
}

// tuple_add(self, other_tuple)
func pyTupleAddFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.TUPLE_ADD_ARG_COUNT_ERROR)
	}
	selfTuple, okSelf := args[0].(*Tuple)
	if !okSelf {
		return NOT_IMPLEMENTED // To allow other type to try __radd__
	}
	otherTuple, okOther := args[1].(*Tuple)
	if !okOther {
		return NOT_IMPLEMENTED // Can only concatenate tuple (not "other.Type()") to tuple
	}

	newElements := make([]Object, 0, len(selfTuple.Elements)+len(otherTuple.Elements))
	newElements = append(newElements, selfTuple.Elements...)
	newElements = append(newElements, otherTuple.Elements...)
	return &Tuple{Elements: newElements}
}

// tuple_mul(self, count_int)
func pyTupleMulFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.TUPLE_MUL_ARG_COUNT_ERROR)
	}
	selfTuple, okSelf := args[0].(*Tuple)
	if !okSelf {
		return NOT_IMPLEMENTED
	}
	countInt, okCount := args[1].(*Integer)
	if !okCount {
		return NOT_IMPLEMENTED // Can't multiply sequence by non-int of type 'other.Type()'
	}
	count := int(countInt.Value)
	if count < 0 {
		count = 0
	} // Multiply by negative is empty tuple

	if len(selfTuple.Elements) == 0 || count == 0 {
		return &Tuple{Elements: []Object{}}
	}
	// Check for potential overflow before allocation if count is huge
	if count > 0 && len(selfTuple.Elements) > 0 && ((1<<30)/len(selfTuple.Elements)) < count {
		return NewError(constants.OverflowError, constants.TUPLE_MUL_OVERFLOW_ERROR)
	}

	newElements := make([]Object, 0, len(selfTuple.Elements)*count)
	for i := 0; i < count; i++ {
		newElements = append(newElements, selfTuple.Elements...)
	}
	return &Tuple{Elements: newElements}
}

// tuple_contains(self, item)
func pyTupleContainsFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_CONTAINS_ARG_COUNT_ERROR) // Reusing for now
	}
	selfTuple, okSelf := args[0].(*Tuple)
	if !okSelf {
		return NewError(constants.TypeError, constants.TUPLE_CONTAINS_ON_TUPLE_ERROR)
	}
	itemToFind := args[1]

	for _, item := range selfTuple.Elements {
		comparisonResult := CompareObjects(constants.EqualsOperator, item, itemToFind, ctx)
		if IsError(comparisonResult) {
			return comparisonResult
		}
		if comparisonResult == TRUE {
			return TRUE
		}
	}
	return FALSE
}

// --- AttributeGetter for Tuple ---
func (t *Tuple) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	// Helper to create a Builtin that wraps the Go function and prepends `self`
	makeTupleMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.TUPLE_METHOD_PREFIX + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				// `t` (the receiver tuple) is captured in this closure
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, t) // Prepend self
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case constants.TUPLE_COUNT_METHOD_NAME:
		return makeTupleMethod(constants.TUPLE_COUNT_METHOD_NAME, pyTupleCountFn), true
	case constants.TUPLE_INDEX_METHOD_NAME:
		return makeTupleMethod(constants.TUPLE_INDEX_METHOD_NAME, pyTupleIndexFn), true
	case constants.DunderLen:
		return makeTupleMethod(constants.DunderLen, pyTupleLenFn), true
	case constants.DunderAdd:
		return makeTupleMethod(constants.DunderAdd, pyTupleAddFn), true
	case constants.DunderMul:
		return makeTupleMethod(constants.DunderMul, pyTupleMulFn), true
	case constants.DunderRMul: // For int * tuple
		return makeTupleMethod(constants.DunderRMul, pyTupleMulFn), true // Same implementation as __mul__
	case constants.DunderContains:
		return makeTupleMethod(constants.DunderContains, pyTupleContainsFn), true
		// Comparison dunders will be handled by the generic CompareObjects logic
		// if Tuple implements Hashable correctly and its elements are comparable.
	}
	return nil, false
}

var _ AttributeGetter = (*Tuple)(nil)
