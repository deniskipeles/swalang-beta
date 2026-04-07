package object

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
)

const (
	STATIC_METHOD_OBJ ObjectType = constants.STATIC_METHOD_OBJ_TYPE
	CLASS_METHOD_OBJ  ObjectType = constants.CLASS_METHOD_OBJ_TYPE
)

// --- StaticMethod ---
type StaticMethod struct {
	Function *Function // The Pylearn function it wraps
}

func (sm *StaticMethod) Type() ObjectType { return STATIC_METHOD_OBJ }
func (sm *StaticMethod) Inspect() string {
	funcName := constants.STATIC_METHOD_ANON_NAME
	if sm.Function != nil && sm.Function.Name != constants.EmptyString {
		funcName = sm.Function.Name
	}
	return fmt.Sprintf(constants.STATIC_METHOD_INSPECT_FORMAT, funcName, sm)
}

// GetObjectAttribute for StaticMethod implements the descriptor protocol.
func (sm *StaticMethod) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	// Expose the name of the wrapped function.
	if name == constants.DunderName { // Assuming "__name__" is the common attribute
		return &String{Value: sm.Function.Name}, true
	}
	// When a staticmethod wrapper is "called", it should execute the underlying function.
	if name == constants.DunderCall {
		return sm.Function, true
	}
	return nil, false
}

var _ Object = (*StaticMethod)(nil)
var _ AttributeGetter = (*StaticMethod)(nil)

// --- ClassMethod ---
type ClassMethod struct {
	Function *Function // The Pylearn function it wraps
}

func (cm *ClassMethod) Type() ObjectType { return CLASS_METHOD_OBJ }
func (cm *ClassMethod) Inspect() string {
	funcName := constants.CLASS_METHOD_ANON_NAME
	if cm.Function != nil && cm.Function.Name != constants.EmptyString {
		funcName = cm.Function.Name
	}
	return fmt.Sprintf(constants.CLASS_METHOD_INSPECT_FORMAT, funcName, cm)
}

// GetObjectAttribute for ClassMethod implements the descriptor protocol.
func (cm *ClassMethod) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	// Expose the name of the wrapped function.
	if name == constants.DunderName { // Assuming "__name__" is the common attribute
		return &String{Value: cm.Function.Name}, true
	}
	// The __call__ for a classmethod is more complex as it depends on what it's bound to.
	// This is handled by the GetObjectAttribute on the owner class/instance, so we don't
	// need a __call__ here.
	return nil, false
}

var _ Object = (*ClassMethod)(nil)
var _ AttributeGetter = (*ClassMethod)(nil)
