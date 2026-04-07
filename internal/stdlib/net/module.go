// internal/stdlib/net/module.go
package net

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyNewStreamingBody is the native constructor for net.StreamingBody(file_object).
func pyNewStreamingBody(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "StreamingBody() takes exactly 1 argument (a file-like object)")
	}

	// We expect the argument to be a Pylearn file object, which wraps an os.File.
	fileObj, ok := args[0].(*object.File)
	if !ok {
		return object.NewError(constants.TypeError, "StreamingBody() argument must be a file object opened in binary mode")
	}

	// The os.File from Go's standard library satisfies the io.ReadCloser interface.
	return &StreamingBody{Source: fileObj.File}
}

func init() {
	// Initialize server-side components first, so their classes are available.
	initServer()

	// Environment for the top-level 'net' module
	env := object.NewEnvironment()

	// --- Register top-level functions ---
	env.Set("dial", Dial)
	env.Set("listen", Listen)
	env.Set("parse_ip", ParseIP)

	// --- Register class types ---
	env.Set("App", &object.Builtin{Name: "net.App", Fn: pyAppConstructor})
	env.Set("Request", RequestClass)
	env.Set("Response", ResponseClass)
	env.Set("Connection", ConnectionClass)
	env.Set("Listener", ListenerClass)
	env.Set("Addr", AddrClass)
	env.Set("StreamingBody", &object.Builtin{Name: "net.StreamingBody", Fn: pyNewStreamingBody})

	// <<< START OF FIX >>>
	// Explicitly add the error classes to this module's environment so they
	// can be accessed as `_net.ConnectionError`, `_net.TimeoutError`, etc.
	env.Set("ConnectionError", ConnectionErrorClass)
	env.Set("TimeoutError", TimeoutErrorClass)
	env.Set("ProxyError", ProxyErrorClass)
	env.Set("ICMPError", ICMPErrorClass)
	env.Set("HTTP2Error", HTTP2ErrorClass)
	// <<< END OF FIX >>>

	// --- Register Submodules ---
	env.Set("websocket", createWebsocketModule())
	env.Set("http2", createHttp2Module())
	env.Set("publicsuffix", createPublicSuffixModule())
	env.Set("idna", createIdnaModule())
	env.Set("icmp", createIcmpModule())
	env.Set("spdy", createSpdyModule())
	env.Set("proxy", createProxyModule())
	env.Set("netutil", createNotImplementedModule("netutil"))
	env.Set("nettest", createNotImplementedModule("nettest"))

	// Create the final Module object
	netModule := &object.Module{
		Name: "_net",
		Path: "<builtin_net>",
		Env:  env,
	}

	object.RegisterNativeModule("_net", netModule)
}

func createNotImplementedModule(name string) *object.Module {
	env := object.NewEnvironment()
	mod := &object.Module{
		Name: name,
		Path: "<builtin_net_unimplemented>",
		Env:  env,
	}
	env.Set("__doc__", &object.String{Value: "This submodule is not yet implemented."})
	return mod
}