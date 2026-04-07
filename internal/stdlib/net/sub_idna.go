package net

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"golang.org/x/net/idna"
)

func pyIdnaEncode(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "encode() takes 1 argument")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "argument must be a string")
	}
	encoded, err := idna.Lookup.ToASCII(s.Value)
	if err != nil {
		return object.NewError("IDNAError", err.Error())
	}
	return &object.String{Value: encoded}
}
func createIdnaModule() *object.Module {
	env := object.NewEnvironment()
	env.Set("encode", &object.Builtin{Name: "net.idna.encode", Fn: pyIdnaEncode})
	// Add decode, etc.
	return &object.Module{
		Name: "idna",
		Path: "<builtin_net_idna>",
		Env:  env,
	}
}
