// internal/vm/compiled_code.go
package vm

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/object" // Import the base object package
	// "github.com/deniskipeles/pylearn/internal/code" // VM objects might need code access
)

// Define VM-specific object types (can reuse names if distinct package)

// --- VM Compiled Function Object ---
type CompiledFunction struct {
	Instructions  []byte // Raw bytecode bytes
	NumLocals     int    // Includes parameters + declared locals
	NumParameters int
	Name          string // Function/method name for debugging/reflection
}

// Type returns a distinct type to differentiate from object.Function if needed,
// but for the VM's internal use, returning the base type might be okay
// as long as we use type assertions correctly. Let's use a distinct type for clarity.
const VM_COMPILED_FUNCTION_OBJ object.ObjectType = "VM_COMPILED_FUNCTION"
const VM_BOUND_METHOD_OBJ object.ObjectType = "VM_BOUND_METHOD"
const VM_CLOSURE_OBJ object.ObjectType = "VM_CLOSURE" // Distinct type
const VM_CLASS_OBJ object.ObjectType = "VM_CLASS"
const VM_COMPILED_METHODS_OBJ object.ObjectType = "VM_COMPILED_METHODS"

func (cf *CompiledFunction) Type() object.ObjectType { return VM_COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	name := cf.Name
	if name == "" {
		name = "<script>"
	} // Or <lambda>
	return fmt.Sprintf("<CompiledFunction %s #locals=%d #params=%d at %p>",
		name, cf.NumLocals, cf.NumParameters, cf)
}

// Ensure it satisfies object.Object interface (implicitly done by embedding or methods)
var _ object.Object = (*CompiledFunction)(nil)

// // --- VM Closure Object ---
// type Closure1 struct {
// 	Fn   *CompiledFunction // Holds the VM's CompiledFunction
// 	Free []object.Object   // Slice of free variables (these are base object.Object)
// }

// func (c *Closure1) Type() object.ObjectType { return VM_CLOSURE_OBJ }
// func (c *Closure1) Inspect() string {
// 	return fmt.Sprintf("<Closure Fn=%s Free=%d vars at %p>",
// 		c.Fn.Inspect(), len(c.Free), c)
// }
// var _ object.Object = (*Closure1)(nil)

// // --- VM Class Object ---
// // This might be very similar to object.Class but potentially store Closures
// // instead of Functions in its Methods map. Let's define a distinct one.
// type Class1 struct {
//     Name string
//     Methods map[string]*Closure // Stores VM Closures directly
//     // Superclass *Class // Points to VM Class type
// }

// func (c *Class1) Type() object.ObjectType { return VM_CLASS_OBJ }
// func (c *Class1) Inspect() string  { return fmt.Sprintf("<VM class '%s'>", c.Name) }
// var _ object.Object = (*Class1)(nil)
// // Note: VM Class doesn't need HashKey usually

// // --- VM Bound Method ---
// type BoundMethod1 struct {
//     Instance object.Object // Can be object.Instance or other types later
//     Method   *Closure1     // Stores the VM Closure for the method
// }

// func (bm *BoundMethod1) Type() object.ObjectType { return VM_BOUND_METHOD_OBJ }
// func (bm *BoundMethod1) Inspect() string {
//     // Similar inspect logic to object.BoundMethod but references VM types
//     instanceInspect := "<unknown instance>"
// 	if bm.Instance != nil { instanceInspect = bm.Instance.Inspect() }
//     methodInspect := "<unknown method>"
//     if bm.Method != nil { methodInspect = bm.Method.Inspect() }

//     // className := "<ClassName>" // Need a way to get class name from Instance reliably
// 	// if inst, ok := bm.Instance.(*object.Instance); ok { className = inst.Class.Name }

// 	return fmt.Sprintf("<VM bound method %s of %s>", methodInspect, instanceInspect)
// }
// var _ object.Object = (*BoundMethod1)(nil)

// --- Constants Holder (if needed by compiler/VM plan) ---
// This was used in a previous plan, might remove if not needed by the final OpClass plan
type CompiledMethodsHolder struct {
	Methods map[string]int // Method Name -> CompiledFunction Constant Index
}

func (cmh *CompiledMethodsHolder) Type() object.ObjectType { return VM_COMPILED_METHODS_OBJ }
func (cmh *CompiledMethodsHolder) Inspect() string {
	return fmt.Sprintf("<compiled_methods %d entries>", len(cmh.Methods))
}

var _ object.Object = (*CompiledMethodsHolder)(nil)
