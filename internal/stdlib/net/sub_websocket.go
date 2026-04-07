package net

import (
	"fmt"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/gorilla/websocket"
	"net/http"
)

const WEBSOCKET_OBJ object.ObjectType = "net.WebSocket"

var WebSocketClass *object.Class
var WebSocketErrorClass *object.Class

// WebSocket is the Pylearn object wrapper for a websocket connection.
type WebSocket struct {
	Conn *websocket.Conn
}

func (ws *WebSocket) Type() object.ObjectType { return WEBSOCKET_OBJ }
func (ws *WebSocket) Inspect() string {
	if ws.Conn == nil {
		return "<net.WebSocket closed>"
	}
	return fmt.Sprintf("<net.WebSocket to %s>", ws.Conn.RemoteAddr())
}
func (ws *WebSocket) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	return object.GetNativeMethod(ws, name)
}

// --- WebSocket Methods ---
func wsRecv(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*WebSocket)
	messageType, p, err := self.Conn.ReadMessage()
	if err != nil {
		return object.NewError("WebSocketError", err.Error())
	}
	if messageType == websocket.TextMessage {
		return &object.String{Value: string(p)}
	}
	return &object.Bytes{Value: p}
}
func wsSend(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*WebSocket)
	data := args[1]
	var err error
	if s, ok := data.(*object.String); ok {
		err = self.Conn.WriteMessage(websocket.TextMessage, []byte(s.Value))
	} else if b, ok := data.(*object.Bytes); ok {
		err = self.Conn.WriteMessage(websocket.BinaryMessage, b.Value)
	} else {
		return object.NewError(constants.TypeError, "send() requires string or bytes")
	}
	if err != nil {
		return object.NewError("WebSocketError", err.Error())
	}
	return object.NULL
}
func wsClose(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*WebSocket)
	_ = self.Conn.Close()
	return object.NULL
}

// wsConnect implements websocket.connect(url, headers=None)
func wsConnect(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// <<< START FIX >>>
	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, "connect() takes 1 or 2 arguments (url, headers=None)")
	}
	url, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "url must be a string")
	}

	var requestHeader http.Header
	if len(args) == 2 {
		headersDict, ok := args[1].(*object.Dict)
		if !ok {
			return object.NewError(constants.TypeError, "headers must be a dict")
		}
		requestHeader = http.Header{}
		for _, pair := range headersDict.Pairs {
			key, kOk := pair.Key.(*object.String)
			val, vOk := pair.Value.(*object.String)
			if kOk && vOk {
				requestHeader.Add(key.Value, val.Value)
			}
		}
	}

	c, _, err := websocket.DefaultDialer.Dial(url.Value, requestHeader)
	// <<< END FIX >>>

	if err != nil {
		return object.NewError("ConnectionError", err.Error())
	}
	return &WebSocket{Conn: c}
}

// createWebsocketModule creates the 'net.websocket' submodule.
func createWebsocketModule() *object.Module {
	// <<< START FIX >>>
	// Initialize the WebSocketError class for this submodule
	WebSocketErrorClass = object.CreateExceptionClass("WebSocketError", object.OSErrorClass)
	object.BuiltinExceptionClasses["WebSocketError"] = WebSocketErrorClass
	// <<< END FIX >>>

	// Initialize the WebSocket class first
	WebSocketClass = object.CreateExceptionClass("WebSocket", object.ObjectClass)
	WebSocketClass.Methods["recv"] = &object.Builtin{Fn: wsRecv, Name: "net.WebSocket.recv"}
	WebSocketClass.Methods["send"] = &object.Builtin{Fn: wsSend, Name: "net.WebSocket.send"}
	WebSocketClass.Methods["close"] = &object.Builtin{Fn: wsClose, Name: "net.WebSocket.close"}
	
	// Assign native methods for the AttributeGetter
	object.SetNativeMethod(WEBSOCKET_OBJ, "recv", wsRecv)
	object.SetNativeMethod(WEBSOCKET_OBJ, "send", wsSend)
	object.SetNativeMethod(WEBSOCKET_OBJ, "close", wsClose)

	// Create the module environment
	env := object.NewEnvironment()
	env.Set("connect", &object.Builtin{Name: "net.websocket.connect", Fn: wsConnect})
	env.Set("WebSocket", WebSocketClass) // Expose the class
	env.Set("WebSocketError", WebSocketErrorClass) // <-- ADD THIS LINE

	return &object.Module{
		Name: "websocket",
		Path: "<builtin_net_websocket>",
		Env:  env,
	}
}

// Helper to create a native HTTP handler that can be used by an upgrader
func NewWebSocketHandler(pylearnHandler object.Object, ctx object.ExecutionContext) http.HandlerFunc {
	upgrader := websocket.Upgrader{} // with default options
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wsObj := &WebSocket{Conn: conn}
		go ctx.NewChildContext(nil).Execute(pylearnHandler, wsObj)
	}
}
