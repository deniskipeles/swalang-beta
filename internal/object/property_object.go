// FILE: internal/object/property_object.go

package object

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants"
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
	return fmt.Sprintf(constants.OBJECT_PROPERTY_INSPECT_FORMAT, p)
}

// GetObjectAttribute for the Property object itself (e.g., prop.setter)
func (p *Property) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makePropMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.OBJECT_PROPERTY_METHOD_PREFIX + methodName,
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
	case constants.OBJECT_PROPERTY_SETTER_METHOD_NAME:
		return makePropMethod(constants.OBJECT_PROPERTY_SETTER_METHOD_NAME, pyPropertySetter), true
	case constants.OBJECT_PROPERTY_DELETER_METHOD_NAME:
		return makePropMethod(constants.OBJECT_PROPERTY_DELETER_METHOD_NAME, pyPropertyDeleter), true
	}
	return nil, false
}

// --- Go functions for property methods (.setter, .deleter) ---

// Pylearn: prop.setter(fset)
func pyPropertySetter(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.OBJECT_PROPERTY_SETTER_ARG_COUNT_ERROR)
	}
	selfProp, ok := args[0].(*Property)
	if !ok {
		return NewError(constants.TypeError, constants.OBJECT_PROPERTY_SETTER_ON_NON_PROPERTY_ERROR)
	}
	fset := args[1]
	if !IsCallable(fset) {
		return NewError(constants.TypeError, constants.OBJECT_PROPERTY_SETTER_ARG_TYPE_ERROR)
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
		return NewError(constants.TypeError, constants.OBJECT_PROPERTY_DELETER_ARG_COUNT_ERROR)
	}
	selfProp, ok := args[0].(*Property)
	if !ok {
		return NewError(constants.TypeError, constants.OBJECT_PROPERTY_DELETER_ON_NON_PROPERTY_ERROR)
	}
	fdel := args[1]
	if !IsCallable(fdel) {
		return NewError(constants.TypeError, constants.OBJECT_PROPERTY_DELETER_ARG_TYPE_ERROR)
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