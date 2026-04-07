// internal/stdlib/net/errors.go
package net

import (
	"github.com/deniskipeles/pylearn/internal/object"
)

var (
	ConnectionErrorClass *object.Class
	TimeoutErrorClass    *object.Class
	ProxyErrorClass      *object.Class
	ICMPErrorClass       *object.Class
	HTTP2ErrorClass      *object.Class
)

func init() {
	ConnectionErrorClass = object.CreateExceptionClass("ConnectionError", object.OSErrorClass)
	TimeoutErrorClass = object.CreateExceptionClass("TimeoutError", object.OSErrorClass)
	ProxyErrorClass = object.CreateExceptionClass("ProxyError", object.OSErrorClass)
	ICMPErrorClass = object.CreateExceptionClass("ICMPError", object.OSErrorClass)
	HTTP2ErrorClass = object.CreateExceptionClass("HTTP2Error", object.ExceptionClass)

	// Register them so they can be created with object.NewError
	object.BuiltinExceptionClasses["ConnectionError"] = ConnectionErrorClass
	object.BuiltinExceptionClasses["TimeoutError"] = TimeoutErrorClass
	object.BuiltinExceptionClasses["ProxyError"] = ProxyErrorClass
	object.BuiltinExceptionClasses["ICMPError"] = ICMPErrorClass
	object.BuiltinExceptionClasses["HTTP2Error"] = HTTP2ErrorClass
}
