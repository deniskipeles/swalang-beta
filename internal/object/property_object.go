// FILE: internal/object/property_object.go

package object

import (
	"fmt"
)

const PROPERTY_OBJ ObjectType = "property"

// Property implements the descriptor protocol for the @property decorator.
type Property struct {
	FGet Object // The getter function (e.g., the decorated `text` method)
	FSet Object // The setter function (from .setter)
	FDel Object // The deleter function (from .deleter)
	Doc  Object // The docstring
}

func (p *Property) Type() ObjectType { return PROPERTY_OBJ }
func (p *Property) Inspect() string {
	// Provide a Python-like representation
	return fmt.Sprintf("<property object at %p>", p)
}

// GetObjectAttribute for the Property object itself (e.g., prop.setter)
func (p *Property) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makePropMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: "property." + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				// Prepend `self` (the Property `p`) to the arguments
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, p)
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case "setter":
		return makePropMethod("setter", pyPropertySetter), true
	case "deleter":
		return makePropMethod("deleter", pyPropertyDeleter), true
	}
	return nil, false
}

// --- Go functions for property methods (.setter, .deleter) ---

// Pylearn: prop.setter(fset)
func pyPropertySetter(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError("TypeError", "setter() takes exactly 1 argument (the setter function)")
	}
	selfProp, ok := args[0].(*Property)
	if !ok {
		return NewError("TypeError", "setter must be called on a property object")
	}
	fset := args[1]
	if !IsCallable(fset) {
		return NewError("TypeError", "setter argument must be a callable")
	}

	// Create a *new* property object with the setter configured
	newProp := &Property{
		FGet: selfProp.FGet,
		FSet: fset,
		FDel: selfProp.FDel,
		Doc:  selfProp.Doc,
	}
	return newProp
}

// Pylearn: prop.deleter(fdel)
func pyPropertyDeleter(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError("TypeError", "deleter() takes exactly 1 argument (the deleter function)")
	}
	selfProp, ok := args[0].(*Property)
	if !ok {
		return NewError("TypeError", "deleter must be called on a property object")
	}
	fdel := args[1]
	if !IsCallable(fdel) {
		return NewError("TypeError", "deleter argument must be a callable")
	}

	// Create a *new* property object with the deleter configured
	newProp := &Property{
		FGet: selfProp.FGet,
		FSet: selfProp.FSet,
		FDel: fdel,
		Doc:  selfProp.Doc,
	}
	return newProp
}

var _ Object = (*Property)(nil)
var _ AttributeGetter = (*Property)(nil)