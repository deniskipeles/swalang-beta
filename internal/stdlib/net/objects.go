// internal/stdlib/net/objects.go
package net

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"golang.org/x/net/proxy"
)

const (
	ADDR_OBJ         object.ObjectType = "net.Addr"
	CONNECTION_OBJ   object.ObjectType = "net.Connection"
	LISTENER_OBJ     object.ObjectType = "net.Listener"
	PROXY_DIALER_OBJ object.ObjectType = "net.proxy.Dialer"
)

var (
	AddrClass        *object.Class
	ConnectionClass  *object.Class
	ListenerClass    *object.Class
	ProxyDialerClass *object.Class
)

// --- Addr Object ---
type Addr struct{ GoAddr net.Addr }

func (a *Addr) Type() object.ObjectType { return ADDR_OBJ }
func (a *Addr) Inspect() string         { return a.GoAddr.String() }

// --- Connection Object ---
type Connection struct{ GoConn net.Conn }

func (c *Connection) Type() object.ObjectType { return CONNECTION_OBJ }
func (c *Connection) Inspect() string {
	return fmt.Sprintf("<Connection from %s to %s>", c.GoConn.LocalAddr(), c.GoConn.RemoteAddr())
}

// --- Listener Object ---
type Listener struct{ GoListener net.Listener }

func (l *Listener) Type() object.ObjectType { return LISTENER_OBJ }
func (l *Listener) Inspect() string         { return fmt.Sprintf("<Listener on %s>", l.GoListener.Addr()) }

// --- ProxyDialer Object ---
type ProxyDialer struct{ GoDialer proxy.Dialer }

func (d *ProxyDialer) Type() object.ObjectType { return PROXY_DIALER_OBJ }
func (d *ProxyDialer) Inspect() string         { return fmt.Sprintf("<net.proxy.Dialer at %p>", d) }

// --- Method Implementations ---

func addrNetwork(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Addr)
	return &object.String{Value: self.GoAddr.Network()}
}

func addrString(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Addr)
	return &object.String{Value: self.GoAddr.String()}
}

func connRead(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Connection)
	bufsize := int64(4096) // Default buffer size
	if len(args) == 2 {
		if size, ok := args[1].(*object.Integer); ok {
			bufsize = size.Value
		} else {
			return object.NewError(constants.TypeError, "read size must be an integer")
		}
	}
	buf := make([]byte, bufsize)
	n, err := self.GoConn.Read(buf)
	if err != nil {
		if err == io.EOF {
			return &object.Bytes{Value: []byte{}} // EOF is empty bytes in Python sockets
		}
		return object.NewError("IOError", err.Error())
	}
	return &object.Bytes{Value: buf[:n]}
}

func connWrite(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Connection)
	data, ok := args[1].(*object.Bytes)
	if !ok {
		return object.NewError(constants.TypeError, "write() argument must be bytes")
	}
	n, err := self.GoConn.Write(data.Value)
	if err != nil {
		return object.NewError("IOError", err.Error())
	}
	return &object.Integer{Value: int64(n)}
}

func connClose(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Connection)
	err := self.GoConn.Close()
	if err != nil {
		return object.NewError("IOError", err.Error())
	}
	return object.NULL
}

func connSetTimeout(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Connection)
	timeoutVal, ok := args[1].(*object.Float)
	if !ok {
		return object.NewError(constants.TypeError, "timeout must be a float")
	}
	duration := time.Duration(timeoutVal.Value * float64(time.Second))
	self.GoConn.SetDeadline(time.Now().Add(duration))
	return object.NULL
}

func connGetPeerName(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Connection)
	return &Addr{GoAddr: self.GoConn.RemoteAddr()}
}

func connGetSockName(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Connection)
	return &Addr{GoAddr: self.GoConn.LocalAddr()}
}

func listenerAccept(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Listener)
	conn, err := self.GoListener.Accept()
	if err != nil {
		return object.NewError("IOError", err.Error())
	}
	pylearnConn := &Connection{GoConn: conn}
	pylearnAddr := &Addr{GoAddr: conn.RemoteAddr()}
	return &object.Tuple{Elements: []object.Object{pylearnConn, pylearnAddr}}
}

func listenerClose(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Listener)
	err := self.GoListener.Close()
	if err != nil {
		return object.NewError("IOError", err.Error())
	}
	return object.NULL
}

func listenerAddr(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Listener)
	return &Addr{GoAddr: self.GoListener.Addr()}
}

func init() {
	// Initialize AddrClass
	AddrClass = object.CreateExceptionClass("Addr", object.ObjectClass) // Use CreateExceptionClass as a generic class builder for now
	AddrClass.Methods["network"] = &object.Builtin{Fn: addrNetwork, Name: "net.Addr.network"}
	AddrClass.Methods["string"] = &object.Builtin{Fn: addrString, Name: "net.Addr.string"}

	// Initialize ConnectionClass
	ConnectionClass = object.CreateExceptionClass("Connection", object.ObjectClass)
	ConnectionClass.Methods["read"] = &object.Builtin{Fn: connRead, Name: "net.Connection.read"}
	ConnectionClass.Methods["write"] = &object.Builtin{Fn: connWrite, Name: "net.Connection.write"}
	ConnectionClass.Methods["close"] = &object.Builtin{Fn: connClose, Name: "net.Connection.close"}
	ConnectionClass.Methods["settimeout"] = &object.Builtin{Fn: connSetTimeout, Name: "net.Connection.settimeout"}
	ConnectionClass.Methods["getpeername"] = &object.Builtin{Fn: connGetPeerName, Name: "net.Connection.getpeername"}
	ConnectionClass.Methods["getsockname"] = &object.Builtin{Fn: connGetSockName, Name: "net.Connection.getsockname"}

	// Initialize ListenerClass
	ListenerClass = object.CreateExceptionClass("Listener", object.ObjectClass)
	ListenerClass.Methods["accept"] = &object.Builtin{Fn: listenerAccept, Name: "net.Listener.accept"}
	ListenerClass.Methods["close"] = &object.Builtin{Fn: listenerClose, Name: "net.Listener.close"}
	ListenerClass.Methods["addr"] = &object.Builtin{Fn: listenerAddr, Name: "net.Listener.addr"}

	// Initialize ProxyDialerClass
	ProxyDialerClass = object.CreateExceptionClass("Dialer", object.ObjectClass)

	// Assign the actual methods to the AttributeGetter implementation on the Go structs
	object.SetNativeMethod(ADDR_OBJ, "network", addrNetwork)
	object.SetNativeMethod(ADDR_OBJ, "string", addrString)

	object.SetNativeMethod(CONNECTION_OBJ, "read", connRead)
	object.SetNativeMethod(CONNECTION_OBJ, "write", connWrite)
	object.SetNativeMethod(CONNECTION_OBJ, "close", connClose)
	object.SetNativeMethod(CONNECTION_OBJ, "settimeout", connSetTimeout)
	object.SetNativeMethod(CONNECTION_OBJ, "getpeername", connGetPeerName)
	object.SetNativeMethod(CONNECTION_OBJ, "getsockname", connGetSockName)

	object.SetNativeMethod(LISTENER_OBJ, "accept", listenerAccept)
	object.SetNativeMethod(LISTENER_OBJ, "close", listenerClose)
	object.SetNativeMethod(LISTENER_OBJ, "addr", listenerAddr)
}

// func (a *Addr) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
// 	return object.GetNativeMethod(a, name)
// }
// func (c *Connection) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
// 	return object.GetNativeMethod(c, name)
// }
// func (l *Listener) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
// 	return object.GetNativeMethod(l, name)
// }

// Implement AttributeGetter for all native structs to delegate method lookup to their respective classes.
func (a *Addr) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	return AddrClass.GetObjectAttribute(ctx, name)
}
func (c *Connection) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	return ConnectionClass.GetObjectAttribute(ctx, name)
}
func (l *Listener) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	return ListenerClass.GetObjectAttribute(ctx, name)
}
func (d *ProxyDialer) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	return ProxyDialerClass.GetObjectAttribute(ctx, name)
}