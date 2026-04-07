// internal/stdlib/websocket/websocket_object.go
package websocket

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/gorilla/websocket" // Go WebSocket library
)

const (
	WEBSOCKET_OBJ object.ObjectType = constants.WEBSOCKET_OBJ_TYPE
	// Note: The 'WEBSOCKET_CLOSED_ERR_OBJ' constant is no longer needed here
	// as we use the string constant from the 'constants' package.
)

// The WebSocketClosedError struct has been REMOVED.
// We now use the standard object.Error and create it with the
// appropriate class (WebSocketClosedErrorClass) via the object.NewError constructor.

type WebSocket struct {
	mu           sync.Mutex // To protect concurrent access to wsConn
	Conn         *websocket.Conn
	asyncRuntime object.AsyncRuntimeAPI // This needs to be set when WebSocket is created
	closed       bool
}

func NewWebSocket(conn *websocket.Conn, asyncRuntime object.AsyncRuntimeAPI) *WebSocket {
	if asyncRuntime == nil {
		// This is a critical issue for async operations.
	}
	return &WebSocket{Conn: conn, asyncRuntime: asyncRuntime}
}

func (ws *WebSocket) Type() object.ObjectType { return WEBSOCKET_OBJ }
func (ws *WebSocket) Inspect() string {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	addr := constants.WEBSOCKET_INSPECT_UNKNOWN_ADDR
	if ws.Conn != nil {
		addr = ws.Conn.RemoteAddr().String()
	}
	status := constants.WEBSOCKET_INSPECT_STATUS_OPEN
	if ws.closed || ws.Conn == nil {
		status = constants.WEBSOCKET_INSPECT_STATUS_CLOSED
	}
	return fmt.Sprintf(constants.WEBSOCKET_INSPECT_FORMAT, addr, status, ws)
}

// --- Pylearn Methods for WebSocket object.Object ---

// Pylearn: await ws.receive()
func (ws *WebSocket) PyReceive(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.WEBSOCKET_RECEIVE_ARG_COUNT_ERROR)
	}
	if ws.asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, constants.WEBSOCKET_ASYNC_RUNTIME_ERROR)
	}

	goAsyncResult := ws.asyncRuntime.CreateCoroutine(func(goCoroutineCtx context.Context) (interface{}, error) {
		ws.mu.Lock()
		if ws.closed || ws.Conn == nil {
			ws.mu.Unlock()
			// Return a Pylearn error object as the *value* of the coroutine.
			// This allows `await` to resolve to this error, which can be caught in Pylearn.
			return object.NewError(constants.WebSocketClosedError, constants.WEBSOCKET_CLOSED_ERROR_MESSAGE), nil
		}
		conn := ws.Conn
		ws.mu.Unlock()

		messageType, p, err := conn.ReadMessage()

		ws.mu.Lock()
		if err != nil {
			ws.closed = true
			ws.mu.Unlock()

			// Check for clean close conditions. These are not Go-level errors but
			// should be raised as catchable WebSocketClosedError in Pylearn.
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return object.NewError(constants.WebSocketClosedError, constants.WEBSOCKET_CLOSED_BY_PEER_ERROR), nil
			}
			if e, ok := err.(*websocket.CloseError); ok {
				return object.NewError(constants.WebSocketClosedError, e.Error()), nil
			}

			// This is a genuine network or protocol error. Propagate as a real Go error.
			// The `await` in the interpreter will turn this into a RuntimeError.
			return nil, fmt.Errorf(constants.WEBSOCKET_READ_ERROR_MESSAGE, err)
		}
		ws.mu.Unlock()

		switch messageType {
		case websocket.TextMessage:
			return &object.String{Value: string(p)}, nil
		case websocket.BinaryMessage:
			return &object.Bytes{Value: p}, nil
		case websocket.CloseMessage:
			ws.mu.Lock()
			ws.closed = true
			ws.mu.Unlock()
			var closeReason string = constants.WEBSOCKET_RECEIVED_CLOSE_FRAME_REASON
			if len(p) > 2 {
				// Gorilla decodes the close message into the error, but if we get a raw
				// close frame, we can extract the reason.
				closeReason = string(p[2:])
			}
			return object.NewError(constants.WebSocketClosedError, closeReason), nil
		default:
			// Unexpected message type, treat as a Go-level error.
			return nil, fmt.Errorf(constants.WEBSOCKET_UNEXPECTED_DATA_TYPE_ERROR, messageType)
		}
	})

	return &object.AsyncResultWrapper{GoAsyncResult: goAsyncResult}
}

// Pylearn: await ws.send(data)
func (ws *WebSocket) PySend(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.WEBSOCKET_SEND_ARG_COUNT_ERROR, len(args)-1)
	}
	if ws.asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, constants.WEBSOCKET_ASYNC_RUNTIME_ERROR)
	}

	dataToSend := args[1]

	// Wrap the synchronous send operation in a coroutine to make it awaitable.
	goAsyncResult := ws.asyncRuntime.CreateCoroutine(func(goCoroutineCtx context.Context) (interface{}, error) {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		if ws.closed || ws.Conn == nil {
			// Return a Pylearn error as the value, not a Go error.
			return object.NewError(constants.WebSocketClosedError, constants.WEBSOCKET_SEND_ON_CLOSED_ERROR), nil
		}

		var err error
		switch d := dataToSend.(type) {
		case *object.String:
			err = ws.Conn.WriteMessage(websocket.TextMessage, []byte(d.Value))
		case *object.Bytes:
			err = ws.Conn.WriteMessage(websocket.BinaryMessage, d.Value)
		default:
			// This error happens before the send, so return it as the value.
			return object.NewError(constants.TypeError, constants.WEBSOCKET_SEND_ARG_TYPE_ERROR, dataToSend.Type()), nil
		}

		if err != nil {
			ws.closed = true
			if _, ok := err.(*websocket.CloseError); ok {
				return object.NewError(constants.WebSocketClosedError, err.Error()), nil
			}
			return object.NewError(constants.WebSocketWriteError, constants.WEBSOCKET_WRITE_ERROR_MESSAGE, err), nil
		}

		// On success, the coroutine resolves to Pylearn's object.NULL.
		return object.NULL, nil
	})

	return &object.AsyncResultWrapper{GoAsyncResult: goAsyncResult}
}

// Pylearn: await ws.close(code=1000, reason="")
func (ws *WebSocket) PyClose(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 3 {
		return object.NewError(constants.TypeError, constants.WEBSOCKET_CLOSE_ARG_COUNT_ERROR, len(args)-1)
	}
	if ws.asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, constants.WEBSOCKET_ASYNC_RUNTIME_ERROR)
	}

	// Capture arguments for the coroutine
	var capturedArgs []object.Object
	if len(args) > 1 {
		capturedArgs = args[1:]
	}

	// Wrap the close operation in a coroutine.
	goAsyncResult := ws.asyncRuntime.CreateCoroutine(func(goCoroutineCtx context.Context) (interface{}, error) {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		if ws.closed || ws.Conn == nil {
			return object.NULL, nil // No-op, resolves to object.NULL.
		}

		closeCode := websocket.CloseNormalClosure
		reasonStr := constants.EmptyString

		if len(capturedArgs) >= 1 && capturedArgs[0] != object.NULL {
			codeObj, ok := capturedArgs[0].(*object.Integer)
			if !ok {
				return object.NewError(constants.TypeError, constants.WEBSOCKET_CLOSE_CODE_TYPE_ERROR), nil
			}
			closeCode = int(codeObj.Value)
		}
		if len(capturedArgs) == 2 && capturedArgs[1] != object.NULL {
			reasonObj, ok := capturedArgs[1].(*object.String)
			if !ok {
				return object.NewError(constants.TypeError, constants.WEBSOCKET_CLOSE_REASON_TYPE_ERROR), nil
			}
			reasonStr = reasonObj.Value
		}

		msg := websocket.FormatCloseMessage(closeCode, reasonStr)
		err := ws.Conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second*5))
		if err != nil && err != websocket.ErrCloseSent {
			fmt.Fprintf(os.Stderr, constants.WEBSOCKET_WARNING_CLOSE_WRITE_ERROR, err)
		}

		errClose := ws.Conn.Close()
		ws.closed = true
		ws.Conn = nil

		if errClose != nil {
			fmt.Fprintf(os.Stderr, constants.WEBSOCKET_WARNING_FINAL_CLOSE_ERROR, errClose)
		}

		return object.NULL, nil // Close operation resolves to object.NULL.
	})

	return &object.AsyncResultWrapper{GoAsyncResult: goAsyncResult}
}

// GetObjectAttribute for WebSocket
func (ws *WebSocket) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	makeWSMethod := func(methodName string, goFn object.BuiltinFunction) *object.Builtin {
		return &object.Builtin{
			Name: constants.WEBSOCKET_METHOD_PREFIX + methodName,
			Fn: func(callCtx object.ExecutionContext, scriptProvidedArgs ...object.Object) object.Object {
				methodArgs := make([]object.Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, ws) // Prepend self
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}
	switch name {
	case constants.WEBSOCKET_RECEIVE_METHOD_NAME:
		return makeWSMethod(constants.WEBSOCKET_RECEIVE_METHOD_NAME, ws.PyReceive), true
	case constants.WEBSOCKET_SEND_METHOD_NAME:
		return makeWSMethod(constants.WEBSOCKET_SEND_METHOD_NAME, ws.PySend), true
	case constants.WEBSOCKET_CLOSE_METHOD_NAME:
		return makeWSMethod(constants.WEBSOCKET_CLOSE_METHOD_NAME, ws.PyClose), true
	}
	return nil, false
}

var _ object.Object = (*WebSocket)(nil)
var _ object.AttributeGetter = (*WebSocket)(nil)