// internal/stdlib/net/sub_http2.go
package net

import (
	"net/http"

	"github.com/deniskipeles/pylearn/internal/object"
	// "github.com/deniskipeles/pylearn/internal/stdlib/pyhttpserver" // <-- REMOVE THIS IMPORT
	"golang.org/x/net/http2"
)

// <<< START FIX >>>
// Define an interface that describes the behavior we need from the server object.
// The real `pyhttpserver.ServerWrapper` struct implicitly satisfies this interface.
type httpServerGetter interface {
	GetGoHTTPServer() *http.Server
}

// <<< END FIX >>>

// pyHttp2ConfigureServer implements http2.configure_server(server_obj)
func pyHttp2ConfigureServer(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "configure_server() takes exactly 1 argument (an http.server object)")
	}

	// <<< START FIX >>>
	// Instead of a direct type assertion, we assert that the passed object
	// implements our new interface.
	serverWrapper, ok := args[0].(httpServerGetter)
	if !ok {
		return object.NewError("TypeError", "argument must be a compatible http.server object, not %s", args[0].Type())
	}
	// <<< END FIX >>>

	// Configure the underlying Go http.Server for HTTP/2.
	// We pass a new empty http2.Server struct for default configuration.
	if err := http2.ConfigureServer(serverWrapper.GetGoHTTPServer(), &http2.Server{}); err != nil {
		return object.NewError("HTTP2Error", "failed to configure server for HTTP/2: %v", err)
	}

	// The configuration happens in-place on the Go server object.
	return object.NULL
}

func createHttp2Module() *object.Module {
	env := object.NewEnvironment()
	env.Set("configure_server", &object.Builtin{Name: "net.http2.configure_server", Fn: pyHttp2ConfigureServer})
	env.Set("HTTP2Error", HTTP2ErrorClass) // Expose the specific error type
	env.Set("__doc__", &object.String{Value: "HTTP/2 protocol support via Go's golang.org/x/net/http2."})
	return &object.Module{
		Name: "http2",
		Path: "<builtin_net_http2>",
		Env:  env,
	}
}
