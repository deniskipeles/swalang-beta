// internal/stdlib/pyhttp/module.go
package pyhttp

import (
	"github.com/deniskipeles/pylearn/internal/object"
)

func init() {
	env := object.NewEnvironment()

	// Expose the main 'request' function and specific helpers 'get' and 'post'
	env.Set("request", Request) // From http_request.go
	env.Set("get", Get)         // From http_get.go (now wraps request)
	env.Set("post", Post)       // From http_post.go (now wraps request)

	httpModule := &object.Module{
		Name: "http",
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("http", httpModule)
}