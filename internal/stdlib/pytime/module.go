// internal/stdlib/pytime/module.go
package pytime

import (
	"github.com/deniskipeles/pylearn/internal/object"
)

func init() {
	// Environment for the 'time' module
	env := object.NewEnvironment()

	// Add functions to the module's environment
	env.Set("sleep", TimeSleep) // From time_funcs.go
	env.Set("time", TimeTime)   // From time_funcs.go

	// You can add constants like time.timezone, time.altzone, time.daylight if needed.
	// For simplicity, we'll omit them for now.

	// Create the Module object
	timeModule := &object.Module{
		Name: "time",
		Path: "<builtin>", // Or a more specific path if you prefer
		Env:  env,
	}

	// Register the module with Pylearn's central registry
	object.RegisterNativeModule("time", timeModule)
}