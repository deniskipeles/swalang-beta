package object

import (
	"bytes"
	"sort"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// Set (Mutable Collection, Not Hashable itself)
// Elements are unique and must be Hashable.
type Set struct {
	Elements map[HashKey]Object // Store elements by their HashKey for uniqueness
}

func (s *Set) Type() ObjectType { return SET_OBJ }
func (s *Set) Inspect() string {
	if len(s.Elements) == 0 {
		return constants.SET_EMPTY_INSPECT_FORMAT // Python's representation of an empty set
	}
	var out bytes.Buffer
	items := make([]string, 0, len(s.Elements))

	// For deterministic Inspect, sort the string representations of elements
	tempElements := make([]Object, 0, len(s.Elements))
	for _, v := range s.Elements {
		tempElements = append(tempElements, v)
	}
	sort.Slice(tempElements, func(i, j int) bool {
		return tempElements[i].Inspect() < tempElements[j].Inspect()
	})

	for _, item := range tempElements {
		items = append(items, item.Inspect())
	}

	out.WriteString(constants.SET_INSPECT_OPEN_BRACE)
	out.WriteString(strings.Join(items, constants.SET_INSPECT_SEPARATOR))
	out.WriteString(constants.SET_INSPECT_CLOSE_BRACE)
	return out.String()
}

var _ Object = (*Set)(nil)

// Sets are not hashable because they are mutable.

// --- Helper to get an iterator for another object, used by set operations ---
func getIteratorForSetOp(ctx ExecutionContext, other Object, methodName string) (Iterator, Object) {
	iterator, errObj := GetObjectIterator(ctx, other, NoToken) // Use NoToken for internal ops
	if errObj != nil {
		return nil, NewError(constants.TypeError, constants.ITERATOR_SET_OP_ITERABLE_TYPE_ERROR, other.Type(), methodName)
	}
	return iterator, nil
}

// --- Go functions for Set methods (to be wrapped by AttributeGetter) ---

// set_add(self, element)
func pySetAddFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_ADD_ARG_COUNT_ERROR, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_ADD_ON_SET_ERROR, args[0].Type())
	}
	element := args[1]

	hashableElem, okHash := element.(Hashable)
	if !okHash {
		return NewError(constants.TypeError, constants.SET_UNHASHABLE_TYPE_ERROR_FORMAT, element.Type())
	}
	hKey, err := hashableElem.HashKey()
	if err != nil {
		return NewError(constants.TypeError, constants.SET_FAILED_TO_HASH_ERROR_FORMAT, err)
	}
	if selfSet.Elements == nil { // Initialize map if it's the first add
		selfSet.Elements = make(map[HashKey]Object)
	}
	selfSet.Elements[hKey] = element
	return NULL // add() modifies in-place
}

// set_remove(self, element) - raises KeyError if element not found
func pySetRemoveFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_REMOVE_ARG_COUNT_ERROR, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_REMOVE_ON_SET_ERROR)
	}
	element := args[1]

	hashableElem, okHash := element.(Hashable)
	if !okHash {
		return NewError(constants.TypeError, constants.SET_UNHASHABLE_TYPE_ERROR_FORMAT, element.Type())
	}
	hKey, err := hashableElem.HashKey()
	if err != nil {
		return NewError(constants.TypeError, constants.SET_FAILED_TO_HASH_ERROR_FORMAT, err)
	}

	if _, exists := selfSet.Elements[hKey]; !exists {
		return NewError(constants.KeyError, constants.SET_REMOVE_KEY_ERROR_FORMAT, element.Inspect())
	}
	delete(selfSet.Elements, hKey)
	return NULL
}

// set_discard(self, element) - does nothing if element not found
func pySetDiscardFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_DISCARD_ARG_COUNT_ERROR, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_DISCARD_ON_SET_ERROR)
	}
	element := args[1]

	hashableElem, okHash := element.(Hashable)
	if !okHash { // Non-hashable can't be in set
		return NULL
	}
	hKey, err := hashableElem.HashKey()
	if err != nil { // Error hashing means it can't be in set
		return NULL
	}
	delete(selfSet.Elements, hKey) // Safe to call delete even if key doesn't exist
	return NULL
}

// set_pop(self) - removes and returns an arbitrary element, raises KeyError if empty
func pySetPopFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 {
		return NewError(constants.TypeError, constants.SET_POP_ARG_COUNT_ERROR, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_POP_ON_SET_ERROR)
	}

	if len(selfSet.Elements) == 0 {
		return NewError(constants.KeyError, constants.SET_POP_EMPTY_SET_ERROR)
	}
	// Get an arbitrary element (Go map iteration order is not guaranteed)
	var poppedKey HashKey
	var poppedElem Object
	for k, v := range selfSet.Elements {
		poppedKey = k
		poppedElem = v
		break
	}
	delete(selfSet.Elements, poppedKey)
	return poppedElem
}

// set_clear(self)
func pySetClearFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 {
		return NewError(constants.TypeError, constants.SET_CLEAR_ARG_COUNT_ERROR, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_CLEAR_ON_SET_ERROR)
	}
	selfSet.Elements = make(map[HashKey]Object) // Reassign to a new empty map
	return NULL
}

// set_len(self)
func pySetLenFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 {
		return NewError(constants.TypeError, constants.SET_LEN_ON_SET_ERROR)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_LEN_ON_SET_ERROR)
	}
	return &Integer{Value: int64(len(selfSet.Elements))}
}

// set_contains(self, element)
func pySetContainsFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_CONTAINS_ARG_COUNT_ERROR)
	} // Reusing for now
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_CONTAINS_ON_SET_ERROR)
	}
	element := args[1]

	hashableElem, okHash := element.(Hashable)
	if !okHash {
		return FALSE
	}
	hKey, err := hashableElem.HashKey()
	if err != nil {
		return FALSE
	}

	_, exists := selfSet.Elements[hKey]
	return NativeBoolToBooleanObject(exists)
}

// --- Set operations ---
// helper function for binary set operations like union, intersection, difference
func pySetBinaryOpFn(opName string, ctx ExecutionContext, self *Set, other Object,
	opLogic func(s1, s2 *Set) *Set) Object {

	// resultSet := &Set{Elements: make(map[HashKey]Object)}

	// Convert 'other' to a Set if it's iterable
	otherSetElements := make(map[HashKey]Object)
	iterator, errObj := getIteratorForSetOp(ctx, other, opName)
	if errObj != nil {
		return errObj
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if IsError(item) {
			return item
		}
		hashableItem, okHash := item.(Hashable)
		if !okHash {
			return NewError(constants.TypeError, constants.ITERATOR_SET_OP_UNHASHABLE_TYPE, item.Type(), opName)
		}
		hKey, err := hashableItem.HashKey()
		if err != nil {
			return NewError(constants.TypeError, constants.ITERATOR_SET_OP_FAILED_TO_HASH, opName, err)
		}
		otherSetElements[hKey] = item
	}
	otherAsSet := &Set{Elements: otherSetElements}

	return opLogic(self, otherAsSet)
}

// set_union(self, other_iterable) / operator |
func pySetUnionFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ARG_COUNT_ERROR, constants.SET_UNION_FUNC_NAME, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ON_SET_ERROR, constants.SET_UNION_FUNC_NAME)
	}

	return pySetBinaryOpFn(constants.SET_UNION_FUNC_NAME, ctx, selfSet, args[1], func(s1, s2 *Set) *Set {
		result := &Set{Elements: make(map[HashKey]Object)}
		for k, v := range s1.Elements {
			result.Elements[k] = v
		}
		for k, v := range s2.Elements {
			result.Elements[k] = v
		}
		return result
	})
}

// set_intersection(self, other_iterable) / operator &
func pySetIntersectionFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ARG_COUNT_ERROR, constants.SET_INTERSECTION_FUNC_NAME, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ON_SET_ERROR, constants.SET_INTERSECTION_FUNC_NAME)
	}

	return pySetBinaryOpFn(constants.SET_INTERSECTION_FUNC_NAME, ctx, selfSet, args[1], func(s1, s2 *Set) *Set {
		result := &Set{Elements: make(map[HashKey]Object)}
		// Iterate over the smaller set for efficiency
		smaller, larger := s1, s2
		if len(s1.Elements) > len(s2.Elements) {
			smaller, larger = s2, s1
		}
		for k, v := range smaller.Elements {
			if _, exists := larger.Elements[k]; exists {
				result.Elements[k] = v
			}
		}
		return result
	})
}

// set_difference(self, other_iterable) / operator -
func pySetDifferenceFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ARG_COUNT_ERROR, constants.SET_DIFFERENCE_FUNC_NAME, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ON_SET_ERROR, constants.SET_DIFFERENCE_FUNC_NAME)
	}

	return pySetBinaryOpFn(constants.SET_DIFFERENCE_FUNC_NAME, ctx, selfSet, args[1], func(s1, s2 *Set) *Set {
		result := &Set{Elements: make(map[HashKey]Object)}
		for k, v := range s1.Elements {
			if _, exists := s2.Elements[k]; !exists {
				result.Elements[k] = v
			}
		}
		return result
	})
}

// set_symmetric_difference(self, other_iterable) / operator ^
func pySetSymmetricDifferenceFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ARG_COUNT_ERROR, constants.SET_SYMMETRIC_DIFFERENCE_FUNC_NAME, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ON_SET_ERROR, constants.SET_SYMMETRIC_DIFFERENCE_FUNC_NAME)
	}

	return pySetBinaryOpFn(constants.SET_SYMMETRIC_DIFFERENCE_FUNC_NAME, ctx, selfSet, args[1], func(s1, s2 *Set) *Set {
		result := &Set{Elements: make(map[HashKey]Object)}
		for k, v := range s1.Elements {
			if _, exists := s2.Elements[k]; !exists {
				result.Elements[k] = v
			}
		}
		for k, v := range s2.Elements {
			if _, exists := s1.Elements[k]; !exists {
				result.Elements[k] = v
			}
		}
		return result
	})
}

// --- Comparison methods ---
// set_issubset(self, other_iterable) / operator <=
func pySetIsSubsetFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ARG_COUNT_ERROR, constants.SET_ISSUBSET_FUNC_NAME, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ON_SET_ERROR, constants.SET_ISSUBSET_FUNC_NAME)
	}

	// Convert 'other' to a Set for comparison
	otherSetElements := make(map[HashKey]Object)
	iterator, errObj := getIteratorForSetOp(ctx, args[1], constants.SET_ISSUBSET_FUNC_NAME)
	if errObj != nil {
		return errObj
	}
	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if IsError(item) {
			return item
		}
		h, e := item.(Hashable)
		if !e {
			return NewError(constants.TypeError, constants.SET_UNHASHABLE_TYPE_ERROR_FORMAT, constants.SET_ISSUBSET_FUNC_NAME)
		}
		hk, _ := h.HashKey()
		otherSetElements[hk] = item
	}
	otherAsSet := &Set{Elements: otherSetElements}

	if len(selfSet.Elements) > len(otherAsSet.Elements) {
		return FALSE
	} // Cannot be subset if larger
	for k := range selfSet.Elements {
		if _, exists := otherAsSet.Elements[k]; !exists {
			return FALSE
		}
	}
	return TRUE
}

// set_issuperset(self, other_iterable) / operator >=
func pySetIsSupersetFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ARG_COUNT_ERROR, constants.SET_ISSUPERSET_FUNC_NAME, len(args)-1)
	}
	selfSet, ok := args[0].(*Set)
	if !ok {
		return NewError(constants.TypeError, constants.SET_BINARY_OP_ON_SET_ERROR, constants.SET_ISSUPERSET_FUNC_NAME)
	}

	otherSetElements := make(map[HashKey]Object)
	iterator, errObj := getIteratorForSetOp(ctx, args[1], constants.SET_ISSUPERSET_FUNC_NAME)
	if errObj != nil {
		return errObj
	}
	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if IsError(item) {
			return item
		}
		h, e := item.(Hashable)
		if !e {
			return NewError(constants.TypeError, constants.SET_UNHASHABLE_TYPE_ERROR_FORMAT, constants.SET_ISSUPERSET_FUNC_NAME)
		}
		hk, _ := h.HashKey()
		otherSetElements[hk] = item
	}
	otherAsSet := &Set{Elements: otherSetElements}

	if len(selfSet.Elements) < len(otherAsSet.Elements) {
		return FALSE
	} // Cannot be superset if smaller
	for k := range otherAsSet.Elements { // Check if all elements of other are in self
		if _, exists := selfSet.Elements[k]; !exists {
			return FALSE
		}
	}
	return TRUE
}

// --- AttributeGetter for Set ---
func (s *Set) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makeSetMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.SET_METHOD_PREFIX + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, s) // Prepend self
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case constants.SET_ADD_METHOD_NAME:
		return makeSetMethod(constants.SET_ADD_METHOD_NAME, pySetAddFn), true
	case constants.SET_REMOVE_METHOD_NAME:
		return makeSetMethod(constants.SET_REMOVE_METHOD_NAME, pySetRemoveFn), true
	case constants.SET_DISCARD_METHOD_NAME:
		return makeSetMethod(constants.SET_DISCARD_METHOD_NAME, pySetDiscardFn), true
	case constants.SET_POP_METHOD_NAME:
		return makeSetMethod(constants.SET_POP_METHOD_NAME, pySetPopFn), true
	case constants.SET_CLEAR_METHOD_NAME:
		return makeSetMethod(constants.SET_CLEAR_METHOD_NAME, pySetClearFn), true
	case constants.DunderLen:
		return makeSetMethod(constants.DunderLen, pySetLenFn), true
	case constants.DunderContains:
		return makeSetMethod(constants.DunderContains, pySetContainsFn), true

	// Binary operations
	case constants.SET_UNION_METHOD_NAME, constants.DunderOr:
		return makeSetMethod(constants.SET_UNION_METHOD_NAME, pySetUnionFn), true
	case constants.SET_INTERSECTION_METHOD_NAME, constants.DunderAnd:
		return makeSetMethod(constants.SET_INTERSECTION_METHOD_NAME, pySetIntersectionFn), true
	case constants.SET_DIFFERENCE_METHOD_NAME, constants.DunderSub:
		return makeSetMethod(constants.SET_DIFFERENCE_METHOD_NAME, pySetDifferenceFn), true
	case constants.SET_SYMMETRIC_DIFFERENCE_METHOD_NAME, constants.DunderXor:
		return makeSetMethod(constants.SET_SYMMETRIC_DIFFERENCE_METHOD_NAME, pySetSymmetricDifferenceFn), true

	// Comparison operations
	case constants.SET_ISSUBSET_METHOD_NAME, constants.DunderLe:
		return makeSetMethod(constants.SET_ISSUBSET_METHOD_NAME, pySetIsSubsetFn), true
	case constants.SET_ISSUPERSET_METHOD_NAME, constants.DunderGe:
		return makeSetMethod(constants.SET_ISSUPERSET_METHOD_NAME, pySetIsSupersetFn), true
	case constants.SET_LT_METHOD_NAME: // Proper subset
		return makeSetMethod(constants.SET_LT_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			if len(a) != 2 {
				return NewError(constants.TypeError, constants.SET_LT_ARG_COUNT_ERROR)
			}
			isSubset := pySetIsSubsetFn(c, a...)
			if IsError(isSubset) || isSubset == FALSE {
				return isSubset
			}
			// To be a proper subset, len(self) must be < len(other)
			s1 := a[0].(*Set)
			iter, errIt := getIteratorForSetOp(c, a[1], constants.SET_LT_METHOD_NAME)
			if errIt != nil {
				return errIt
			}
			otherCount := 0
			for {
				_, stop := iter.Next()
				if stop {
					break
				}
				otherCount++
			}
			return NativeBoolToBooleanObject(len(s1.Elements) < otherCount)
		}), true
	case constants.SET_GT_METHOD_NAME: // Proper superset
		return makeSetMethod(constants.SET_GT_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			if len(a) != 2 {
				return NewError(constants.TypeError, constants.SET_GT_ARG_COUNT_ERROR)
			}
			isSuperset := pySetIsSupersetFn(c, a...)
			if IsError(isSuperset) || isSuperset == FALSE {
				return isSuperset
			}
			s1 := a[0].(*Set)
			iter, errIt := getIteratorForSetOp(c, a[1], constants.SET_GT_METHOD_NAME)
			if errIt != nil {
				return errIt
			}
			otherCount := 0
			for {
				_, stop := iter.Next()
				if stop {
					break
				}
				otherCount++
			}
			return NativeBoolToBooleanObject(len(s1.Elements) > otherCount)
		}), true
	case constants.SET_EQ_METHOD_NAME:
		return makeSetMethod(constants.SET_EQ_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			if len(a) != 2 {
				return NewError(constants.TypeError, constants.SET_EQ_ARG_COUNT_ERROR)
			}
			s1 := a[0].(*Set)
			s2Obj := a[1]
			s2, okS2 := s2Obj.(*Set)
			if !okS2 { // Python sets are only equal to other sets
				// Could also check if s2Obj is iterable and convert, but Python's set.__eq__ is stricter.
				return FALSE
			}
			if len(s1.Elements) != len(s2.Elements) {
				return FALSE
			}
			// Check if s1 is a subset of s2 (all elements of s1 are in s2)
			// Since lengths are equal, this also implies s2 is a subset of s1.
			for k := range s1.Elements {
				if _, exists := s2.Elements[k]; !exists {
					return FALSE
				}
			}
			return TRUE
		}), true
	case constants.SET_NE_METHOD_NAME:
		return makeSetMethod(constants.SET_NE_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			if len(a) != 2 {
				return NewError(constants.TypeError, constants.SET_NE_ARG_COUNT_ERROR)
			}
			eqResult := makeSetMethod(constants.SET_EQ_METHOD_NAME, nil).Fn(c, a...) // Call the __eq__ logic
			if IsError(eqResult) {
				return eqResult
			}
			if eqResult == TRUE {
				return FALSE
			}
			return TRUE
		}), true
	}
	return nil, false
}

var _ AttributeGetter = (*Set)(nil)
