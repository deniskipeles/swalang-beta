// pylearn/internal/stdlib/pyhttpserver/server_object.go
package pyhttpserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

const (
	SERVER_OBJ object.ObjectType = constants.SERVER_RESPONSE_OBJ_TYPE
)
// ... (existing code) ...

type ServerWrapper struct {
	mu           sync.Mutex
	Server       *http.Server
	PylearnClass object.Object // The Pylearn handler class
	IsRunning    bool
	closed       chan struct{}
}

// <<< START FIX >>>
// GetGoHTTPServer makes ServerWrapper satisfy the httpServerGetter interface
// defined in the pynet package, allowing for loose coupling.
func (sw *ServerWrapper) GetGoHTTPServer() *http.Server {
	return sw.Server
}

// <<< END FIX >>>

// ... (the rest of the file remains the same) ...

func (sw *ServerWrapper) Type() object.ObjectType { return SERVER_OBJ }
func (sw *ServerWrapper) Inspect() string {
	state := "idle"
	if sw.IsRunning {
		state = "running"
	}
	addr := "?"
	if sw.Server != nil {
		addr = sw.Server.Addr
	}
	return fmt.Sprintf("<http.server at %s, %s>", addr, state)
}

// ...
