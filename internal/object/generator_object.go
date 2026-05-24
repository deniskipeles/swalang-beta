// pylearn/internal/object/generator_object.go
package object

import (
	"github.com/deniskipeles/pylearn/internal/constants"
)

const GENERATOR_OBJ ObjectType = constants.OBJECT_GENERATOR_TYPE

// Generator represents a suspended function execution.
// It is now decoupled from any specific execution engine (interpreter/vm).
type Generator struct {
	// Name for inspection purposes.
	Name string

	// NextFn is a function provided by the execution engine (interpreter/vm)
	// that knows how to resume this specific generator's execution.
	// It's the core of the iterator protocol for generators.
	NextFn func() (Object, bool)

	// We can add state fields if needed for inspection, but the core
	// execution state is managed by the closure in NextFn.
	// e.g., state string // "suspended", "running", "closed"
	SendFn  func(value Object) (Object, bool)
	CloseFn func() (Object, bool)
	ThrowFn func(err Object) (Object, bool)
}

func (g *Generator) Type() ObjectType { return GENERATOR_OBJ }
func (g *Generator) Inspect() string {
	return constants.OBJECT_GENERATOR_INSPECT_PREFIX + g.Name + constants.OBJECT_GENERATOR_INSPECT_SUFFIX
}

// Next implements the Iterator interface by calling the function
// provided by the execution engine that created it.
// Next implements the iterator protocol by calling SendFn with NULL
func (g *Generator) Next() (Object, bool) {
	if g.SendFn == nil {
		// Fallback for older generators or error condition
		if g.NextFn != nil {
			return g.NextFn()
		}
		return NewError(constants.InternalError, constants.OBJECT_GENERATOR_NOT_INIT_ERROR), true
	}
	return g.SendFn(NULL)
}

// YieldValue is a special internal object type used to signal that a value
// has been yielded from within an execution loop (Eval or VM loop).
type YieldValue struct {
	Value Object // The value yielded OUT to the caller
}

func (yv *YieldValue) Type() ObjectType { return constants.OBJECT_TYPE_YIELD_VALUE }
func (yv *YieldValue) Inspect() string  { return yv.Value.Inspect() }

var _ Object = (*Generator)(nil)
var _ Iterator = (*Generator)(nil)
var _ Object = (*YieldValue)(nil)

func (g *Generator) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	switch name {
	case constants.OBJECT_GENERATOR_SEND_METHOD_NAME:
		// Create a Builtin that wraps the generator's SendFn
		sendBuiltin := &Builtin{
			Name: constants.OBJECT_GENERATOR_SEND_BUILTIN_NAME,
			Fn: func(callCtx ExecutionContext, args ...Object) Object {
				if len(args) != 1 {
					return NewError(constants.TypeError, constants.OBJECT_GENERATOR_SEND_ARG_ERROR)
				}
				// The generator `g` is captured by this closure.
				result, stop := g.SendFn(args[0])
				if stop {
					if IsError(result) {
						return result // Propagate errors from the generator
					}
					// If the generator stops without an error, raise StopIteration
					return STOP_ITERATION
				}
				return result
			},
		}
		return sendBuiltin, true
	}
	return nil, false
}

var _ AttributeGetter = (*Generator)(nil)
