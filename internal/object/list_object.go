package object

import (
	"bytes"
	"strings" // For String.join, used in error messages sometimes

	"github.com/deniskipeles/pylearn/internal/constants"
)

// List (Mutable Sequence, Not Hashable)
type List struct{ Elements []Object }

func (l *List) Type() ObjectType { return LIST_OBJ }
func (l *List) Inspect() string { /* ... keep inspect logic ... */
	var out bytes.Buffer; elements := []string{}; for _, e := range l.Elements { elements = append(elements, e.Inspect()) }; out.WriteString(constants.LIST_INSPECT_OPEN_BRACKET); out.WriteString(strings.Join(elements, constants.LIST_INSPECT_SEPARATOR)); out.WriteString(constants.LIST_INSPECT_CLOSE_BRACKET); return out.String()
}
// List Item Access (Get)
func (l *List) GetObjectItem(key Object) Object {
	idxObj, ok := key.(*Integer)
	if !ok { return NewError(constants.TypeError, constants.LIST_ITEM_INDEX_TYPE_ERROR, key.Type()) }
	idx := idxObj.Value
	listLen := int64(len(l.Elements))
	if idx < 0 { idx += listLen }
	if idx < 0 || idx >= listLen { return NewError(constants.IndexError, constants.LIST_ITEM_INDEX_OUT_OF_RANGE) }
	elem := l.Elements[idx]; if elem == nil { return NULL }; return elem
}
// List Item Access (Set)
func (l *List) SetObjectItem(key Object, value Object) Object {
	idxObj, ok := key.(*Integer)
	if !ok { return NewError(constants.TypeError, constants.LIST_ITEM_INDEX_TYPE_ERROR, key.Type()) }
	idx := idxObj.Value
	listLen := int64(len(l.Elements))
	if idx < 0 { idx += listLen }
	if idx < 0 || idx >= listLen { return NewError(constants.IndexError, constants.LIST_ASSIGNMENT_INDEX_OUT_OF_RANGE) }
	l.Elements[idx] = value
	return nil // Indicate success
}
// TODO: List Attribute Access (e.g., .append(), .pop())
// func (l *List) GetObjectAttribute(name string) (Object, bool) { ... }
var _ Object = (*List)(nil)
var _ ItemGetter = (*List)(nil) // List implements item getting
var _ ItemSetter = (*List)(nil) // List implements item setting
// --- Go functions for List methods ---




// GetObjectAttribute for List to expose methods
func (l *List) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makeListMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.LIST_METHOD_PREFIX + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, l) // Prepend self (the List 'l')
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case constants.LIST_APPEND_METHOD_NAME:
		return makeListMethod(constants.LIST_APPEND_METHOD_NAME, pyListAppendFn), true
	case constants.LIST_EXTEND_METHOD_NAME:
		return makeListMethod(constants.LIST_EXTEND_METHOD_NAME, pyListExtendFn), true
	case constants.LIST_INSERT_METHOD_NAME:
		return makeListMethod(constants.LIST_INSERT_METHOD_NAME, pyListInsertFn), true
	case constants.LIST_POP_METHOD_NAME:
		return makeListMethod(constants.LIST_POP_METHOD_NAME, pyListPopFn), true
	case constants.LIST_REMOVE_METHOD_NAME:
		return makeListMethod(constants.LIST_REMOVE_METHOD_NAME, pyListRemoveFn), true
	case constants.LIST_INDEX_METHOD_NAME:
		return makeListMethod(constants.LIST_INDEX_METHOD_NAME, pyListIndexFn), true
	case constants.LIST_COUNT_METHOD_NAME:
		return makeListMethod(constants.LIST_COUNT_METHOD_NAME, pyListCountFn), true
	case constants.DunderLen:
		return makeListMethod(constants.DunderLen, pyListLenFn), true
	case constants.DunderAdd:
		return makeListMethod(constants.DunderAdd, pyListAddFn), true
	case constants.DunderMul:
		return makeListMethod(constants.DunderMul, pyListMulFn), true
	case constants.DunderRMul: // For `int * list`
		return makeListMethod(constants.DunderRMul, pyListMulFn), true
	case constants.DunderContains:
		return makeListMethod(constants.DunderContains, pyListContainsFn), true
	}
	return nil, false
}
var _ AttributeGetter = (*List)(nil) // Ensure List implements AttributeGetter




// pyListAppendFn implements list.append(item)
func pyListAppendFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is item
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_APPEND_ARG_COUNT_ERROR, len(args)-1)
	}
	selfList, ok := args[0].(*List)
	if !ok {
		return NewError(constants.TypeError, constants.LIST_APPEND_ON_LIST_ERROR)
	}
	itemToAppend := args[1]

	selfList.Elements = append(selfList.Elements, itemToAppend)
	return NULL // append modifies in-place and returns None
}

// pyListExtendFn implements list.extend(iterable)
func pyListExtendFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is iterable
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_EXTEND_ARG_COUNT_ERROR, len(args)-1)
	}
	selfList, ok := args[0].(*List)
	if !ok {
		return NewError(constants.TypeError, constants.LIST_EXTEND_ON_LIST_ERROR)
	}
	iterableArg := args[1]

	// Get an iterator for the argument
	// Using NoToken for errors originating from GetObjectIterator itself.
	iterator, errObj := GetObjectIterator(ctx, iterableArg, NoToken)
	if errObj != nil {
		return errObj // Propagate TypeError if not iterable
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break // Iterator finished
		}
		if IsError(item) { // Check for errors yielded by iterator
			return item
		}
		selfList.Elements = append(selfList.Elements, item)
	}
	return NULL // extend modifies in-place and returns None
}

// pyListInsertFn implements list.insert(index, item)
func pyListInsertFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is index, args[2] is item
	if len(args) != 3 {
		return NewError(constants.TypeError, constants.LIST_INSERT_ARG_COUNT_ERROR, len(args)-1)
	}
	selfList, ok := args[0].(*List)
	if !ok { return NewError(constants.TypeError, constants.LIST_INSERT_ON_LIST_ERROR) }

	indexObj, okIndex := args[1].(*Integer)
	if !okIndex { return NewError(constants.TypeError, constants.LIST_INSERT_INDEX_TYPE_ERROR, args[1].Type()) }
	itemToInsert := args[2]

	idx := int(indexObj.Value)
	listLen := len(selfList.Elements)

	// Python's insert index semantics:
	// If index is out of bounds, it clamps to the ends.
	// index < 0: insert at beginning (effectively 0)
	// index >= len: insert at end (effectively len)
	if idx < 0 {
		idx = 0 // Python clamps negative idx to 0 for insert, unlike typical slicing.
		// A more Pythonic negative handling might be idx = listLen + idx; if idx < 0 { idx = 0 },
		// but standard insert clamps at 0 for negatives. Let's stick to simple clamping.
	}
	if idx > listLen {
		idx = listLen
	}

	if idx == listLen { // Insert at the end
		selfList.Elements = append(selfList.Elements, itemToInsert)
	} else { // Insert in the middle or at the beginning
		// Make space for the new element
		selfList.Elements = append(selfList.Elements, nil) // Grow slice by one (value doesn't matter)
		copy(selfList.Elements[idx+1:], selfList.Elements[idx:]) // Shift elements to the right
		selfList.Elements[idx] = itemToInsert                   // Insert the new element
	}
	return NULL
}

// pyListPopFn implements list.pop(index=-1)
func pyListPopFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is index (optional)
	selfList, ok := args[0].(*List)
	if !ok { return NewError(constants.TypeError, constants.LIST_POP_ON_LIST_ERROR) }

	numScriptArgs := len(args) - 1
	if numScriptArgs > 1 {
		return NewError(constants.TypeError, constants.LIST_POP_ARG_COUNT_ERROR, numScriptArgs)
	}

	listLen := len(selfList.Elements)
	if listLen == 0 {
		return NewError(constants.IndexError, constants.LIST_POP_EMPTY_LIST_ERROR)
	}

	idx := listLen - 1 // Default index is -1 (last element)
	if numScriptArgs == 1 {
		if args[1] != NULL { // Allow pop(None) to act as pop()
			indexObj, okIndex := args[1].(*Integer)
			if !okIndex {
				return NewError(constants.TypeError, constants.LIST_POP_INDEX_TYPE_ERROR, args[1].Type())
			}
			idx = int(indexObj.Value)
			if idx < 0 { // Handle negative index
				idx = listLen + idx
			}
		}
	}

	if idx < 0 || idx >= listLen {
		return NewError(constants.IndexError, constants.LIST_POP_INDEX_OUT_OF_RANGE)
	}

	poppedItem := selfList.Elements[idx]
	// Remove element by slicing
	selfList.Elements = append(selfList.Elements[:idx], selfList.Elements[idx+1:]...)
	return poppedItem
}

// pyListRemoveFn implements list.remove(value)
func pyListRemoveFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is value
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_REMOVE_ARG_COUNT_ERROR, len(args)-1)
	}
	selfList, ok := args[0].(*List)
	if !ok { return NewError(constants.TypeError, constants.LIST_REMOVE_ON_LIST_ERROR) }
	valueToRemove := args[1]

	foundIndex := -1
	for i, item := range selfList.Elements {
		// Use CompareObjects for equality, passing the context
		// Assuming CompareObjects returns TRUE, FALSE, or Error
		comparisonResult := CompareObjects(constants.EqualsOperator, item, valueToRemove, ctx)
		if IsError(comparisonResult) {
			return comparisonResult // Propagate error from comparison
		}
		if comparisonResult == TRUE {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return NewError(constants.ValueError, constants.LIST_REMOVE_VALUE_NOT_IN_LIST)
	}

	// Remove element
	selfList.Elements = append(selfList.Elements[:foundIndex], selfList.Elements[foundIndex+1:]...)
	return NULL
}

// pyListIndexFn implements list.index(value, start=0, end=len(list))
func pyListIndexFn(ctx ExecutionContext, args ...Object) Object {
	selfList, ok := args[0].(*List)
	if !ok { return NewError(constants.TypeError, constants.LIST_INDEX_ON_LIST_ERROR) }

	numScriptArgs := len(args) - 1
	if numScriptArgs < 1 || numScriptArgs > 3 {
		return NewError(constants.TypeError, constants.LIST_INDEX_ARG_COUNT_ERROR, numScriptArgs)
	}
	valueToFind := args[1]

	runesSourceLen := len(selfList.Elements) // Use element count for list
	startIdx, endIdx := 0, runesSourceLen

	if numScriptArgs >= 2 && args[2] != NULL {
		startInt, okStart := args[2].(*Integer)
		if !okStart { return NewError(constants.TypeError, constants.LIST_INDEX_SLICE_INDICES_ERROR) }
		startIdx = int(startInt.Value)
	}
	if numScriptArgs == 3 && args[3] != NULL {
		endInt, okEnd := args[3].(*Integer)
		if !okEnd { return NewError(constants.TypeError, constants.LIST_INDEX_SLICE_INDICES_ERROR) }
		endIdx = int(endInt.Value)
	}

	// Python slice semantics for start/end for list.index
	if startIdx < 0 { startIdx = runesSourceLen + startIdx }
	if startIdx < 0 { startIdx = 0 }
	// For list.index, if start is beyond len, it's fine, loop just won't run
	// if startIdx > runesSourceLen { startIdx = runesSourceLen }

	if endIdx < 0 { endIdx = runesSourceLen + endIdx }
	if endIdx < 0 { endIdx = 0 }
	if endIdx > runesSourceLen { endIdx = runesSourceLen }

	if startIdx >= endIdx { // If effective slice is empty or invalid
		return NewError(constants.ValueError, constants.LIST_INDEX_VALUE_NOT_IN_LIST, valueToFind.Inspect())
	}
	
	for i := startIdx; i < endIdx; i++ {
		if i >= len(selfList.Elements) { break } // Safety, should be covered by endIdx logic
		item := selfList.Elements[i]
		comparisonResult := CompareObjects(constants.EqualsOperator, item, valueToFind, ctx)
		if IsError(comparisonResult) { return comparisonResult }
		if comparisonResult == TRUE {
			return &Integer{Value: int64(i)}
		}
	}
	return NewError(constants.ValueError, constants.LIST_INDEX_VALUE_NOT_IN_LIST, valueToFind.Inspect())
}

// pyListCountFn implements list.count(value)
func pyListCountFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_COUNT_ARG_COUNT_ERROR, len(args)-1)
	}
	selfList, ok := args[0].(*List)
	if !ok { return NewError(constants.TypeError, constants.LIST_COUNT_ON_LIST_ERROR) }
	valueToCount := args[1]

	count := 0
	for _, item := range selfList.Elements {
		comparisonResult := CompareObjects(constants.EqualsOperator, item, valueToCount, ctx)
		if IsError(comparisonResult) { return comparisonResult }
		if comparisonResult == TRUE {
			count++
		}
	}
	return &Integer{Value: int64(count)}
}

// pyListLenFn implements list.__len__()
func pyListLenFn(ctx ExecutionContext, args ...Object) Object {
	selfList, ok := args[0].(*List)
	if !ok { return NewError(constants.TypeError, constants.LIST_LEN_ON_LIST_ERROR) }
	if len(args) != 1 { return NewError(constants.TypeError, constants.LIST_LEN_ON_LIST_ERROR) }
	return &Integer{Value: int64(len(selfList.Elements))}
}

// pyListAddFn implements list.__add__(other_list)
func pyListAddFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is other (List)
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_ADD_ARG_COUNT_ERROR)
	}
	selfList, okSelf := args[0].(*List)
	if !okSelf { return NewError(constants.TypeError, constants.LIST_ADD_ON_LIST_ERROR) }
	otherList, okOther := args[1].(*List)
	if !okOther {
		return NewError(constants.TypeError, constants.LIST_ADD_CONCAT_ERROR, args[1].Type())
	}

	newElements := make([]Object, 0, len(selfList.Elements)+len(otherList.Elements))
	newElements = append(newElements, selfList.Elements...)
	newElements = append(newElements, otherList.Elements...)
	return &List{Elements: newElements}
}

// pyListMulFn implements list * int
func pyListMulFn(ctx ExecutionContext, args ...Object) Object {
	// For both list * int and int * list, the arguments will be (list, int)
	// because of how we register __mul__ and __rmul__.
	selfList, okList := args[0].(*List)
	countObj, okInt := args[1].(*Integer)
	if !okList || !okInt {
		// This should be caught by the interpreter's infix logic, but as a safeguard:
		return NewError(constants.TypeError, "unsupported operand type(s) for *: '%s' and '%s'", args[0].Type(), args[1].Type())
	}

	count := countObj.Value
	if count < 0 {
		count = 0 // Multiplying by a negative number results in an empty list
	}

	if len(selfList.Elements) == 0 || count == 0 {
		return &List{Elements: []Object{}}
	}
	
	newSize := int64(len(selfList.Elements)) * count
	// Basic check to prevent enormous memory allocation
	if newSize > 10000000 {
		return NewError("MemoryError", "result of list multiplication is too large")
	}

	newElements := make([]Object, 0, newSize)
	for i := int64(0); i < count; i++ {
		newElements = append(newElements, selfList.Elements...)
	}

	return &List{Elements: newElements}
}

// pyListContainsFn implements list.__contains__(item)
func pyListContainsFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (List), args[1] is item
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_CONTAINS_ARG_COUNT_ERROR)
	}
	selfList, okSelf := args[0].(*List)
	if !okSelf { return NewError(constants.TypeError, constants.LIST_CONTAINS_ON_LIST_ERROR) }
	itemToFind := args[1]

	for _, item := range selfList.Elements {
		comparisonResult := CompareObjects(constants.EqualsOperator, item, itemToFind, ctx)
		if IsError(comparisonResult) { return comparisonResult }
		if comparisonResult == TRUE {
			return TRUE
		}
	}
	return FALSE
}