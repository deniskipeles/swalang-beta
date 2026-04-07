package pyjson

import "github.com/deniskipeles/pylearn/internal/object"

func init() {
	env := object.NewEnvironment()

	env.Set("dumps", Dumps) // Dumps is imported from json_dump.go
	env.Set("loads", Loads) // Loads is imported from json_load.go

	jsonModule := &object.Module{
		Name: "json",
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("json", jsonModule)
}
