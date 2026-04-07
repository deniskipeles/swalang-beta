// ===========================pylearn/internal/stdlib/template/value.go start here===========================
// internal/stdlib/template/value.go
package template

import (
	"fmt"
	"github.com/deniskipeles/pylearn/internal/object"
)

// Value is a wrapper around a Pylearn object for use within the template engine.
// It provides methods for common template operations like truthiness, stringification, and attribute access.
type Value struct {
	obj     object.Object
	execCtx object.ExecutionContext
}

// NewValue creates a new template Value.
func NewValue(obj object.Object, ctx object.ExecutionContext) *Value {
	return &Value{obj: obj, execCtx: ctx}
}

// IsTrue delegates the truthiness check to the main Pylearn interpreter's logic.
func (v *Value) IsTrue() bool {
	if v == nil || v.obj == nil {
		return false
	}
	isTrue, err := object.IsTruthy(v.execCtx, v.obj)
	if err != nil {
		return false // Treat errors during truthiness check as false
	}
	return isTrue
}

// String delegates string conversion to the Pylearn 'str()' built-in for a Pythonic representation.
func (v *Value) String() string {
	if v == nil || v.obj == nil || v.obj.Type() == object.NULL_OBJ {
		return "" // Render None/nil as an empty string
	}
	// For raw strings, return their value directly to avoid quotes.
	if str, isStr := v.obj.(*object.String); isStr {
		return str.Value
	}
	// For all other types, use the str() built-in to get the correct representation.
	strBuiltin, found := v.execCtx.GetCurrentEnvironment().Get("str")
	if !found {
		return v.obj.Inspect() // Fallback if str is not available
	}
	strResult := v.execCtx.Execute(strBuiltin, v.obj)
	if strObj, isStr := strResult.(*object.String); isStr {
		return strObj.Value
	}
	return strResult.Inspect() // Fallback
}

// Getattr delegates attribute access to the main Pylearn interpreter's logic.
func (v *Value) Getattr(name string) *Value {
	if v == nil || v.obj == nil {
		return NewValue(object.NULL, v.execCtx)
	}

	// This is the crucial fix: If the underlying object is a Pylearn Dict,
	// we perform a key lookup on it directly.
	if dict, ok := v.obj.(*object.Dict); ok {
		val, found := dict.Get(name)
		if !found {
			return NewValue(object.NULL, v.execCtx)
		}
		return NewValue(val, v.execCtx)
	}

	// For any other object type, use the standard attribute getter.
	attr, found := object.CallGetAttr(v.execCtx, v.obj, name, object.NoToken)
	if !found || object.IsError(attr) {
		return NewValue(object.NULL, v.execCtx)
	}
	return NewValue(attr, v.execCtx)
}

// Iter returns a Go-level iterator for Pylearn iterables.
func (v *Value) Iter() (object.Iterator, error) {
	if v == nil || v.obj == nil {
		return nil, fmt.Errorf("cannot iterate over a nil value")
	}
	iterator, errObj := object.GetObjectIterator(v.execCtx, v.obj, object.NoToken)
	if errObj != nil {
		return nil, fmt.Errorf(errObj.Inspect())
	}
	return iterator, nil
}

// pylearnDictToTemplateContext converts a Pylearn dictionary to the template's internal context map.
func pylearnDictToTemplateContext(pylearnDict *object.Dict, ctx object.ExecutionContext) map[string]*Value {
	context := make(map[string]*Value)
	if pylearnDict == nil {
		return context
	}
	for _, pair := range pylearnDict.Pairs {
		if key, ok := pair.Key.(*object.String); ok {
			context[key.Value] = NewValue(pair.Value, ctx)
		}
	}
	return context
}
// ===========================pylearn/internal/stdlib/template/value.go ends here===========================