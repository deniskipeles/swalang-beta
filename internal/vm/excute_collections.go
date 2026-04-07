package vm

import (
	"strings"

	"github.com/deniskipeles/pylearn/internal/object"
)

// --- executeInOperation (NEW HELPER FUNCTION) ---
func (vm *VM) executeInOperation() error {
	container, err := vm.pop()
	if err != nil { return err }
	item, err := vm.pop()
	if err != nil { return err }

	ctx := &VMContext{vm: vm}

	// --- Check for __contains__ dunder method first ---
	if getter, ok := container.(object.AttributeGetter); ok {
		containsMethod, found := getter.GetObjectAttribute(ctx, "__contains__")
		if found && containsMethod != nil {
			if object.IsError(containsMethod) {
				return containsMethod.(error)
			}
			// Execute the __contains__ method
			result := ctx.Execute(containsMethod, item)
			if object.IsError(result) {
				return result.(error)
			}
			if boolResult, isBool := result.(*object.Boolean); isBool {
				return vm.push(boolResult)
			}
			return object.NewError("TypeError", "__contains__ returned non-bool (type %s)", result.Type())
		}
	}

	// --- Fallback for built-in types ---
	switch c := container.(type) {
	case *object.String:
		itemStr, ok := item.(*object.String)
		if !ok {
			return object.NewError("TypeError", "'in <string>' requires string as left operand, not %s", item.Type())
		}
		if strings.Contains(c.Value, itemStr.Value) {
			return vm.push(object.TRUE)
		}
		return vm.push(object.FALSE)
	
	case *object.List:
		for _, elem := range c.Elements {
			eqResult := object.CompareObjects("==", item, elem, ctx)
			if object.IsError(eqResult) { return eqResult.(error) }
			if eqResult == object.TRUE { return vm.push(object.TRUE) }
		}
		return vm.push(object.FALSE)

	case *object.Tuple:
		for _, elem := range c.Elements {
			eqResult := object.CompareObjects("==", item, elem, ctx)
			if object.IsError(eqResult) { return eqResult.(error) }
			if eqResult == object.TRUE { return vm.push(object.TRUE) }
		}
		return vm.push(object.FALSE)
	
	case *object.Dict:
		hashableKey, ok := item.(object.Hashable)
		if !ok {
			return vm.push(object.FALSE) // Unhashable types can't be keys
		}
		dictMapKey, err := hashableKey.HashKey()
		if err != nil {
			return vm.push(object.FALSE)
		}
		_, found := c.Pairs[dictMapKey]
		if found {
			return vm.push(object.TRUE)
		}
		return vm.push(object.FALSE)

	default:
		return object.NewError("TypeError", "'in' operator not supported for type '%s'", container.Type())
	}
}