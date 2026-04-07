package pysys

import (
	// Remove interpreter import
	// "github.com/deniskipeles/pylearn/internal/interpreter"

	"github.com/deniskipeles/pylearn/internal/object" // Keep object import
)

// InitializeSysModule creates the 'sys' module object and registers it.
// It requires the argv list object to be passed in from the main setup.
func InitializeSysModule(argvList *object.List) {
	// Ensure argvList is not nil (defensive check)
	if argvList == nil {
		argvList = &object.List{Elements: []object.Object{}} // Default to empty if somehow nil
	}

	// Create the environment for the 'sys' module
	env := object.NewEnvironment()

	// Add variables and functions to the environment
	// Platform variable is initialized in sys_platform.go's init()
	// Exit Builtin is defined in sys_exit.go (with correct signature)
	env.Set("argv", argvList)
	env.Set("exit", Exit)
	env.Set("platform", Platform)

	// Create the Module object
	sysModule := &object.Module{
		Name: "sys",
		Path: "<builtin>",
		Env:  env,
	}

	// Register the module using the central registry function
	// Assumes object.RegisterNativeModule exists in internal/object/registry.go
	object.RegisterNativeModule("sys", sysModule)
}