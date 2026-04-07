// internal/vm/context.go
package vm

import (
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// VMContext implements object.ExecutionContext for the VM.
// It provides a way for built-ins and other native Go code to interact
// with the VM's execution state, such as calling back into Pylearn code.
// This context is designed for re-entrant, synchronous calls into the VM.
type VMContext struct {
	vm *VM
	env *object.Environment // Can be nil; used for specific scopes like f-strings
}

// SetSuperContext sets the context for a subsequent super() call on the VM.
// It interacts with the fields on the main VM struct.
func (vc *VMContext) SetSuperContext(self object.Object, class *object.Class) {
	vc.vm.superSelf = self
	vc.vm.superClass = class
}

// GetSuperSelf retrieves the 'self' object for a zero-argument super() call from the VM.
func (vc *VMContext) GetSuperSelf() object.Object {
	return vc.vm.superSelf
}

// GetSuperClass retrieves the '__class__' for a zero-argument super() call from the VM.
func (vc *VMContext) GetSuperClass() *object.Class {
	return vc.vm.superClass
}

// NewChildContext creates a new VMContext with a new environment, but shares the same VM.
// This is used for evaluating code snippets in a specific, temporary environment.
func (vc *VMContext) NewChildContext(env *object.Environment) object.ExecutionContext {
	return &VMContext{
		vm:  vc.vm,
		env: env,
	}
}

// GetCurrentEnvironment returns the environment associated with this context.
func (vc *VMContext) GetCurrentEnvironment() *object.Environment {
	if vc.env != nil {
		return vc.env
	}
	if vc.vm.frameIndex >= 0 {
		return vc.vm.currentFrame().Env
	}
	return nil
}

// GetAsyncRuntime returns the global async runtime. The VM does not currently
// have an integrated event loop, so this returns nil.
func (vc *VMContext) GetAsyncRuntime() object.AsyncRuntimeAPI {
	return nil
}

// EvaluateASTNode compiles and executes a Pylearn AST node within the VM.
// This is a synchronous, re-entrant operation used for features like f-strings.
func (vc *VMContext) EvaluateASTNode(node ast.Node, env *object.Environment) object.Object {
	var programToCompile *ast.Program
	if exprNode, ok := node.(ast.Expression); ok {
		programToCompile = &ast.Program{
			Statements: []ast.Statement{
				&ast.ExpressionStatement{Expression: exprNode},
			},
		}
	} else if programNode, ok := node.(*ast.Program); ok {
		programToCompile = programNode
	} else {
		return object.NewError(constants.InternalError, "VMContext.EvaluateASTNode expects an Expression or Program, got %T", node)
	}

	comp := NewCompiler()
	if err := comp.Compile(programToCompile); err != nil {
		return object.NewError("CompilationError", "Failed to compile dynamic code snippet: %s", err)
	}
	bytecode := comp.Bytecode()
	compiledFn := &CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     comp.GetNumDefinitions(),
		NumParameters: 0,
		Name:          "<dynamic_snippet>",
	}
	closure := &Closure{Fn: compiledFn, Constants: bytecode.Constants}

	execCtx := vc
	if env != nil {
		execCtx = vc.NewChildContext(env).(*VMContext)
	}

	return execCtx.Execute(closure)
}

// Execute performs a synchronous, re-entrant call into the VM.
// It is used by built-in functions to call Pylearn callables.
func (vc *VMContext) Execute(callable object.Object, args ...object.Object) object.Object {
	vmInstance := vc.vm

	strBuiltin, _ := builtins.Builtins["str"]
	if strBuiltin != nil && callable == strBuiltin {
		if len(args) != 1 {
			return object.NewError(constants.TypeError, "str() takes exactly one argument")
		}
		if s, ok := args[0].(*object.String); ok {
			return s
		}
		return &object.String{Value: args[0].Inspect()}
	}

	initialSP := vmInstance.sp
	initialFrameIndex := vmInstance.frameIndex

	if vmInstance.frameIndex < 0 {
		return object.NewError("VMError", "VM is terminated; cannot execute callback")
	}

	if err := vmInstance.push(callable); err != nil {
		return object.NewError("StackOverflow", "Error pushing callable: %v", err)
	}
	for _, arg := range args {
		if err := vmInstance.push(arg); err != nil {
			vmInstance.sp = initialSP
			return object.NewError("StackOverflow", "Error pushing argument: %v", err)
		}
	}

	var callErr error
	numArgs := len(args)
	isGoBuiltin := false // Flag to track if the callable is a synchronous Go builtin

	switch c := callable.(type) {
	case *Closure:
		callErr = vmInstance.callClosure(c, numArgs)
	case *object.Builtin:
		isGoBuiltin = true // It's a Go builtin, it will execute synchronously
		callErr = vmInstance.callBuiltin(c, numArgs)
	case *object.Class:
		callErr = vmInstance.callClass(c, numArgs)
	case *BoundMethod:
		callErr = vmInstance.callBoundMethod(c, numArgs)
	default:
		if getter, ok := callable.(object.AttributeGetter); ok {
			callMethod, found := getter.GetObjectAttribute(vc, "__call__")
			if found && callMethod != nil && !object.IsError(callMethod) {
				vmInstance.sp = initialSP
				// Recursive call will handle its own logic (builtin vs pylearn func)
				return vc.Execute(callMethod, args...)
			}
		}
		callErr = object.NewError(constants.TypeError, "'%s' object is not callable", c.Type())
	}

	if callErr != nil {
		vmInstance.sp = initialSP
		if pyErr, ok := callErr.(object.Object); ok && object.IsError(pyErr) {
			return pyErr
		}
		return object.NewError("VMError", "Error setting up call: %v", callErr)
	}

	// If it was a synchronous Go builtin, the work is done. The result is on the stack.
	// We don't run the VM loop; we just pop the result and return.
	if isGoBuiltin {
		poppedResult, popErr := vmInstance.pop()
		if popErr != nil {
			vmInstance.sp = initialSP // Restore SP on error
			return object.NewError("VMError", "Error popping result from synchronous builtin call: %v", popErr)
		}
		// Sanity check: stack should be back to its original state.
		if vmInstance.sp != initialSP {
			finalResult := object.NewError("VMError", "Stack pointer mismatch after sync builtin call (expected %d, got %d)", initialSP, vmInstance.sp)
			vmInstance.sp = initialSP // Force restore for safety
			return finalResult
		}
		return poppedResult
	}

	// For Pylearn functions/closures that push a new frame, we must run the VM.
	callbackTargetFrameIndex := initialFrameIndex + 1
	runErr := vmInstance.runSingleCallbackFrame(callbackTargetFrameIndex)

	var finalResult object.Object
	if runErr != nil {
		vmInstance.sp = initialSP
		if pyErr, ok := runErr.(object.Object); ok && object.IsError(pyErr) {
			finalResult = pyErr
		} else {
			finalResult = object.NewError("VMError", "Error during callback execution: %v", runErr)
		}
	} else {
		poppedResult, popErr := vmInstance.pop()
		if popErr != nil {
			vmInstance.sp = initialSP
			finalResult = object.NewError("VMError", "Error popping result from stack: %v", popErr)
		} else {
			finalResult = poppedResult
			if vmInstance.sp != initialSP {
				finalResult = object.NewError("VMError", "Stack pointer mismatch after callback (expected %d, got %d)", initialSP, vmInstance.sp)
				vmInstance.sp = initialSP
			}
		}
	}

	return finalResult
}

// Ensure VMContext implements the interface
var _ object.ExecutionContext = (*VMContext)(nil)