package pysys

import (
	"github.com/deniskipeles/pylearn/internal/constants"
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
	env.Set(constants.SYS_ARGV_KEY, argvList)
	env.Set(constants.SYS_EXIT_KEY, Exit)
	env.Set(constants.SYS_PLATFORM_KEY, Platform)

	// Create the Module object
	sysModule := &object.Module{
		Name: constants.SYS_MODULE_NAME,
		Path: constants.SYS_MODULE_BUILTIN_PATH,
		Env:  env,
	}

	// Register the module using the central registry function
	object.RegisterNativeModule(constants.SYS_MODULE_NAME, sysModule)
}