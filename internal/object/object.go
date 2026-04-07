package object

import (
	"fmt"
	"strings"
	// "bytes"
	// "hash/fnv"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// ObjectType represents the type of an object in our runtime.
type ObjectType string

// Constants for our object types
const (
	INTEGER_OBJ      ObjectType = constants.OBJECT_TYPE_INTEGER
	FLOAT_OBJ        ObjectType = constants.OBJECT_TYPE_FLOAT
	BOOLEAN_OBJ      ObjectType = constants.OBJECT_TYPE_BOOLEAN
	STRING_OBJ       ObjectType = constants.OBJECT_TYPE_STRING
	LIST_OBJ         ObjectType = constants.OBJECT_TYPE_LIST
	DICT_OBJ         ObjectType = constants.OBJECT_TYPE_DICT
	NULL_OBJ         ObjectType = constants.OBJECT_TYPE_NULL
	RETURN_VALUE_OBJ ObjectType = constants.OBJECT_TYPE_RETURN_VALUE // Interpreter specific? Might remove for VM.
	ERROR_OBJ        ObjectType = constants.OBJECT_TYPE_ERROR
	FUNCTION_OBJ     ObjectType = constants.OBJECT_TYPE_FUNCTION // Interpreter specific? Use core.CompiledFunction + vm.Closure?
	ASYNC_RESULT_OBJ ObjectType = constants.OBJECT_TYPE_ASYNC_RESULT

	BUILTIN_OBJ  ObjectType = constants.OBJECT_TYPE_BUILTIN
	BREAK_OBJ    ObjectType = constants.OBJECT_TYPE_BREAK    // Interpreter specific
	CONTINUE_OBJ ObjectType = constants.OBJECT_TYPE_CONTINUE // Interpreter specific
	RANGE_OBJ    ObjectType = constants.OBJECT_TYPE_RANGE
	// CLASS_OBJ        ObjectType = constants.OBJECT_TYPE_CLASS    // Interpreter specific? Use vm.Class?
	// INSTANCE_OBJ     ObjectType = constants.OBJECT_TYPE_INSTANCE // Interpreter specific? Use vm.Instance?
	BOUND_METHOD_OBJ ObjectType = constants.OBJECT_TYPE_BOUND_METHOD // Interpreter specific? Use vm.BoundMethod?
	MODULE_OBJ       ObjectType = constants.OBJECT_TYPE_MODULE

	TUPLE_OBJ     ObjectType = constants.OBJECT_TYPE_TUPLE
	SET_OBJ       ObjectType = constants.OBJECT_TYPE_SET
	BYTES_OBJ     ObjectType = constants.OBJECT_TYPE_BYTES
	FILE_OBJ      ObjectType = constants.OBJECT_TYPE_FILE
	ITERATOR_OBJ  ObjectType = constants.OBJECT_TYPE_ITERATOR
	STOP_ITER_OBJ ObjectType = constants.OBJECT_TYPE_STOP_ITERATION_ERR // Specific Error type

	HTTP_RESPONSE_OBJ ObjectType = constants.OBJECT_TYPE_HTTP_RESPONSE
	HTTP_REQUEST_OBJ  ObjectType = constants.OBJECT_TYPE_HTTP_REQUEST
)

// --- Core Interfaces ---

// Object is the interface that all runtime value representations must implement.
type Object interface {
	Type() ObjectType
	Inspect() string // Returns a string representation of the object for debugging/display
}

// Hashable is an interface for objects that can be used as dictionary or set keys.
type Hashable interface {
	Object                     // Embed Object interface
	HashKey() (HashKey, error) // Return error if not hashable (e.g. contains list)
}

// HashKey is used as the key in the map representing our Dict object or Set elements.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Iterator is an internal Go interface used by the runtime for iteration protocols.
type Iterator interface {
	Object                // Iterators are Pylearn objects
	Next() (Object, bool) // Returns next object and bool indicating if finished (true=finished)
}

// --- Attribute & Item Access Interfaces ---

// AttributeGetter allows getting attributes using dot notation (obj.name)
type AttributeGetter interface {
	GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) // Returns attribute value, true if found (value can be Error)
}

// AttributeSetter allows setting attributes using dot notation (obj.name = value)
type AttributeSetter interface {
	SetObjectAttribute(name string, value Object) bool // Returns true if handled (even if error occurs internally), false otherwise
}

// ItemGetter allows getting items using square brackets (obj[key])
type ItemGetter interface {
	GetObjectItem(key Object) Object // Returns value or Error object
}

// ItemSetter allows setting items using square brackets (obj[key] = value)
type ItemSetter interface {
	SetObjectItem(key Object, value Object) Object // Returns nil on success, Error object on failure
}

// --- Concrete Object Types ---

// Integer
type Integer struct{ Value int64 }

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string {
	return fmt.Sprintf(constants.OBJECT_INTEGER_INSPECT_FORMAT, i.Value)
}
func (i *Integer) HashKey() (HashKey, error) {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}, nil
}

var _ Object = (*Integer)(nil)
var _ Hashable = (*Integer)(nil)

// Float (Not Hashable by default)
type Float struct{ Value float64 }

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf(constants.OBJECT_FLOAT_INSPECT_FORMAT, f.Value) }

var _ Object = (*Float)(nil)

// Boolean
type Boolean struct{ Value bool }

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	return strings.Title(fmt.Sprintf(constants.OBJECT_BOOLEAN_INSPECT_FORMAT, b.Value))
}
func (b *Boolean) HashKey() (HashKey, error) {
	var value uint64 = 0
	if b.Value {
		value = 1
	}
	return HashKey{Type: b.Type(), Value: value}, nil
}

var _ Object = (*Boolean)(nil)
var _ Hashable = (*Boolean)(nil)

// String methods were here

// Bytes

// Null
type Null struct{}

func (n *Null) Type() ObjectType          { return NULL_OBJ }
func (n *Null) Inspect() string           { return constants.OBJECT_NULL_INSPECT }
func (n *Null) HashKey() (HashKey, error) { return HashKey{Type: n.Type(), Value: 0}, nil }

var _ Object = (*Null)(nil)
var _ Hashable = (*Null)(nil)

// StopIterationError
type StopIterationError struct {
	Message    string
	ErrorClass *Class // The Pylearn class of the error
}

func (e *StopIterationError) Type() ObjectType { return STOP_ITER_OBJ }
func (e *StopIterationError) Inspect() string {
	className := constants.StopIteration
	if e.ErrorClass != nil {
		className = e.ErrorClass.Name
	}
	if e.Message != constants.EmptyString {
		return fmt.Sprintf(constants.OBJECT_STOP_ITERATION_INSPECT_FORMAT, className, e.Message)
	}
	return className
}
func (e *StopIterationError) Error() string      { return e.Message }
func (e *StopIterationError) GetMessage() string { return e.Message }

var _ Object = (*StopIterationError)(nil)
var _ error = (*StopIterationError)(nil) // Implements Go's error interface

// GenericIterator (Not Hashable)
type GenericIterator struct {
	NextFn func() (Object, bool)
	Source string
}

func (it *GenericIterator) Type() ObjectType { return ITERATOR_OBJ }
func (it *GenericIterator) Inspect() string {
	return fmt.Sprintf(constants.OBJECT_GENERIC_ITERATOR_INSPECT_FORMAT, it.Source, it)
}
func (it *GenericIterator) Next() (Object, bool) { return it.NextFn() }

// TODO: Iterator Attribute Access (__next__)
// func (it *GenericIterator) GetObjectAttribute(name string) (Object, bool) { ... }
var _ Object = (*GenericIterator)(nil)
var _ Iterator = (*GenericIterator)(nil)

// --- Potentially Interpreter-Specific Types ---
// These might be replaced by VM-specific types (vm.Closure, vm.Class, etc.)
// when running the VM. Keep definitions here for the interpreter path.

// ReturnValue (Likely Interpreter Only)
type ReturnValue struct{ Value Object }

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

var _ Object = (*ReturnValue)(nil)

// BreakObject (Interpreter Only)
type BreakObject struct{}

func (b *BreakObject) Type() ObjectType { return BREAK_OBJ }
func (b *BreakObject) Inspect() string  { return constants.OBJECT_BREAK_INSPECT }

var _ Object = (*BreakObject)(nil)

// ContinueObject (Interpreter Only)
type ContinueObject struct{}

func (c *ContinueObject) Type() ObjectType { return CONTINUE_OBJ }
func (c *ContinueObject) Inspect() string  { return constants.OBJECT_CONTINUE_INSPECT }

var _ Object = (*ContinueObject)(nil)

// Range
type Range struct{ Start, Stop, Step int64 }

func (r *Range) Type() ObjectType { return RANGE_OBJ }
func (r *Range) Inspect() string { /* ... keep inspect logic ... */
	if r.Step == 1 {
		return fmt.Sprintf(constants.OBJECT_RANGE_INSPECT_TWO_ARGS, r.Start, r.Stop)
	} else {
		return fmt.Sprintf(constants.OBJECT_RANGE_INSPECT_THREE_ARGS, r.Start, r.Stop, r.Step)
	}
}

var _ Object = (*Range)(nil)

// BoundMethod (Interpreter representation)
type BoundMethod struct {
	Instance Object
	Method   *Function
}

func (bm *BoundMethod) Type() ObjectType { return BOUND_METHOD_OBJ }

// func (bm *BoundMethod) Inspect() string { /* ... keep inspect logic ... */
//
//		instanceInspect := "<unknown instance>"; if bm.Instance != nil { instanceInspect = bm.Instance.Inspect() }; methodName := bm.Method.Name; if methodName == "" { methodName = "<method>" }; className := "<ClassName>"; if inst, ok := bm.Instance.(*Instance); ok { className = inst.Class.Name }; return fmt.Sprintf("<bound method %s.%s of %s>", className, methodName, instanceInspect)
//	}
//
// BoundMethod: No change needed to its structure.
// Its `Instance` field is already object.Object, so it can hold an *object.Class.
// Its Inspect method may need a slight tweak if Instance is a Class.
func (bm *BoundMethod) Inspect() string {
	instanceInspect := constants.OBJECT_BOUND_METHOD_INSPECT_UNKNOWN_OWNER
	if bm.Instance != nil {
		// Check if the instance is a Class object
		if cls, okCls := bm.Instance.(*Class); okCls {
			instanceInspect = fmt.Sprintf(constants.OBJECT_BOUND_METHOD_INSPECT_CLASS_FORMAT, cls.Name)
		} else {
			instanceInspect = bm.Instance.Inspect()
		}
	}

	methodName := constants.OBJECT_BOUND_METHOD_INSPECT_METHOD_NAME_DEFAULT
	if bm.Method != nil && bm.Method.Name != constants.EmptyString {
		methodName = bm.Method.Name
	}

	// For regular instance methods, try to get class name from Instance
	className := constants.OBJECT_BOUND_METHOD_INSPECT_CLASS_PLACEHOLDER
	if inst, okInst := bm.Instance.(*Instance); okInst && inst.Class != nil {
		className = inst.Class.Name
	} else if cls, okCls := bm.Instance.(*Class); okCls {
		className = cls.Name // If bound to a class, owner is the class itself
	}

	return fmt.Sprintf(constants.OBJECT_BOUND_METHOD_INSPECT_FORMAT, className, methodName, instanceInspect)
}

var _ Object = (*BoundMethod)(nil)

// Module
type Module struct {
	Name, Path string
	Env        *Environment
}

func (m *Module) Type() ObjectType { return MODULE_OBJ }
func (m *Module) Inspect() string {
	return fmt.Sprintf(constants.OBJECT_MODULE_INSPECT_FORMAT, m.Name, m.Path)
}

// Module Attribute Access (using its Environment)
func (m *Module) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	val, ok := m.Env.Get(name)
	if !ok {
		return nil, false
	}
	return val, true
}

var _ Object = (*Module)(nil)
var _ AttributeGetter = (*Module)(nil) // Module implements attribute getting

// --- Global Constants ---
var (
	TRUE               = &Boolean{Value: true}
	FALSE              = &Boolean{Value: false}
	NULL               = &Null{}
	BREAK              = &BreakObject{}                                                                // Interpreter specific
	CONTINUE           = &ContinueObject{}                                                             // Interpreter specific
	STOP_ITERATION     = &StopIterationError{Message: constants.OBJECT_STOP_ITERATION_DEFAULT_MESSAGE} // Singleton Error
	IMPORT_PLACEHOLDER = &Null{}                                                                       // Placeholder for modules during circular import

	NOT_IMPLEMENTED = &NotImplementedType{} // Singleton for NotImplemented

)

// var IMPORT_PLACEHOLDER = &Null{}
type NotImplementedType struct{}

func (ni *NotImplementedType) Type() ObjectType { return constants.OBJECT_NOT_IMPLEMENTED_TYPE } // Or a specific const
func (ni *NotImplementedType) Inspect() string  { return constants.OBJECT_NOT_IMPLEMENTED_INSPECT }

var _ Object = (*NotImplementedType)(nil)

func NewString(value string) *String {
	return &String{Value: value}
}
