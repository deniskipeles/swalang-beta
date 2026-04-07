// pylearn/internal/stdlib/pyflasky/module.go
package pyflasky

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// Pylearn constructor for the App class.
func pyAppConstructor(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError(constants.TypeError, "App() takes no arguments")
	}
	// Call our Go constructor.
	return NewApp()
}

func init() {
	// Create the environment for the 'flasky' module.
	env := object.NewEnvironment()

	// Create a callable "type" for our App class. When called, it runs the constructor.
	appClass := &object.Builtin{
		Name: "App",
		Fn:   pyAppConstructor,
	}
	env.Set("App", appClass)

	// Create and register the module itself.
	flaskyModule := &object.Module{
		Name: "flasky",
		Path: "<builtin_flasky>",
		Env:  env,
	}
	object.RegisterNativeModule("flasky", flaskyModule)
}
