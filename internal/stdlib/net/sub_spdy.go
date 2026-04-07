// internal/stdlib/net/sub_spdy.go
package net

import (
	"github.com/deniskipeles/pylearn/internal/object"
)

func createSpdyModule() *object.Module {
	env := object.NewEnvironment()
	doc := "SPDY protocol support (deprecated in favor of HTTP/2). This module is a placeholder."
	env.Set("__doc__", &object.String{Value: doc})
	return &object.Module{
		Name: "spdy",
		Path: "<builtin_net_spdy>",
		Env:  env,
	}
}