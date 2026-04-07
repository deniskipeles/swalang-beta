// pylearn/internal/interpreter/context.go
package interpreter

import (
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// InterpreterContext holds the environment and other contextual information needed to evaluate code.
type InterpreterContext struct {
	Env *object.Environment

	// Fields for zero-argument super() calls
	superSelf  object.Object
	superClass *object.Class

	// InstructionPtr is used by generators to track where to resume execution.
	InstructionPtr int

	// SentInValue holds the value passed from a generator's .send() method.
	SentInValue object.Object

	// IsResuming is a flag used by the generator evaluation logic.
	// If true, it means the generator is being resumed from a yield,
	// and the `yield` expression should evaluate to `SentInValue`.
	// If false, it means the generator is running normally, and a
	// `yield` expression should pause the generator.
	IsResuming bool

	// ActiveIterators maps a for-loop AST node to its active iterator.
	// This is crucial for preserving iterator state across yield calls.
	ActiveIterators map[*ast.ForStatement]object.Iterator
	PendingForItem    map[*ast.ForStatement]object.Object // <<< NEW: store the item that was being processed when we yielded
}

// SetSuperContext sets the context for a subsequent super() call.
func (ic *InterpreterContext) SetSuperContext(self object.Object, class *object.Class) {
	ic.superSelf = self
	ic.superClass = class
}

// GetSuperSelf retrieves the 'self' object for a zero-argument super() call.
func (ic *InterpreterContext) GetSuperSelf() object.Object {
	return ic.superSelf
}

// GetSuperClass retrieves the '__class__' for a zero-argument super() call.
func (ic *InterpreterContext) GetSuperClass() *object.Class {
	return ic.superClass
}

// // NewChildContext creates a new InterpreterContext with a new environment,
// // but carries over the super() and generator context from its parent.
// func (ic *InterpreterContext) NewChildContext(env *object.Environment) object.ExecutionContext {
// 	return &InterpreterContext{
// 		Env:             env,
// 		superSelf:       ic.superSelf,  // Carry forward
// 		superClass:      ic.superClass, // Carry forward
// 		IsResuming:      ic.IsResuming, // Carry forward generator state
// 		InstructionPtr:  0,
// 		ActiveIterators: make(map[*ast.ForStatement]object.Iterator), // <<< INITIALIZE THE MAP
// 	}
// }

// NewInterpreterContext creates a new, properly initialized InterpreterContext.
func NewInterpreterContext(env *object.Environment) *InterpreterContext {
	return &InterpreterContext{
		Env:             env,
		ActiveIterators: make(map[*ast.ForStatement]object.Iterator),
		PendingForItem:    make(map[*ast.ForStatement]object.Object), // <<< Initialize
	}
}
func (ic *InterpreterContext) NewChildContext(env *object.Environment) object.ExecutionContext {
    // Create the new context using the constructor
	newCtx := NewInterpreterContext(env)
	
    // Carry over the necessary fields from the parent
	newCtx.superSelf = ic.superSelf
	newCtx.superClass = ic.superClass
	newCtx.IsResuming = ic.IsResuming
	// InstructionPtr starts at 0 for a new context.
	// ActiveIterators is already initialized by the constructor.

	return newCtx
}
// Execute implements the core.ExecutionContext interface for the interpreter.
func (ic *InterpreterContext) Execute(callable object.Object, args ...object.Object) object.Object {
	for {
		// First, check for standard callable types.
		switch c := callable.(type) {
		case *object.Function, *object.Builtin, *object.Class, *object.BoundMethod:
			// For these types, `applyFunctionOrClass` handles everything.
			return applyFunctionOrClass(ic, c, args, nil, object.NoToken)
		}

		// If not a standard type, check if it has a __call__ method.
		if getter, ok := callable.(object.AttributeGetter); ok {
			callMethod, found := getter.GetObjectAttribute(ic, constants.ContextGoExecuteCall)
			if found && callMethod != nil {
				callable = callMethod
				continue
			}
		}

		// If we're here, the object is not a standard callable and has no __call__ method.
		return object.NewError(constants.ContextGoExecuteTypeError, constants.ContextGoExecuteReturnErrorValue1ObjectIsNotCallable, callable.Type())
	}
}

// EvaluateASTNode evaluates a given AST node within a specified environment.
// If the environment is nil, it uses the context's current environment.
func (ic *InterpreterContext) EvaluateASTNode(node ast.Node, env *object.Environment) object.Object {
	evalCtx := ic
	// If a specific environment is provided, create a child context for it.
	// This ensures the super() context is maintained even when evaluating
	// an AST node in a different, temporary environment.
	if env != nil && env != ic.Env {
		evalCtx = ic.NewChildContext(env).(*InterpreterContext)
	}
	return Eval(node, evalCtx)
}

// GetCurrentEnvironment returns the environment associated with this context.
func (ic *InterpreterContext) GetCurrentEnvironment() *object.Environment {
	if ic == nil {
		return nil
	}
	return ic.Env
}

// GetAsyncRuntime provides access to the global async runtime.
func (ic *InterpreterContext) GetAsyncRuntime() object.AsyncRuntimeAPI {
	if PylearnAsyncRuntime == nil {
		return nil
	}
	return PylearnAsyncRuntime
}