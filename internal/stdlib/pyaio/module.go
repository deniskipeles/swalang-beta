// internal/stdlib/pyaio/module.go
package pyaio

import (
	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
	"github.com/deniskipeles/pylearn/internal/object"
)

func init() {
	// Environment for the 'aio' module
	env := object.NewEnvironment()

	// Add functions to the module's environment
	env.Set(constants.StdlibAioSleepName, AioSleep)   // From aio_funcs.go
	env.Set(constants.StdlibAioGatherName, AioGather) // From aio_funcs.go

	env.Set("_create_task", AioCreateTask)          // From aio_funcs.go
	env.Set("_run_and_wait", AioRunAndWait)			// From aio_funcs.go

	// Create the Module object
	aioModule := &object.Module{
		Name: constants.StdlibAioModuleName,
		Path: constants.StdlibAioModulePath, // Using a more descriptive path
		Env:  env,
	}

	// Register the module with Pylearn's central registry
	object.RegisterNativeModule(constants.StdlibAioModuleName, aioModule)
}