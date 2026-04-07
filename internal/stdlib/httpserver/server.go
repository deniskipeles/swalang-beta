// internal/stdlib/httpserver/server.go
package pyhttpserver

import (
	gojson "encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/stdlib/pyflasky"
	custom_json "github.com/deniskipeles/pylearn/internal/stdlib/json"
	websocket_custom_lib "github.com/deniskipeles/pylearn/internal/stdlib/websocket"
	"github.com/gorilla/websocket"
)

// SimpleRouter stores route-to-handler mappings.
type SimpleRouter struct {
	mu     sync.RWMutex
	routes map[string]object.Object
}

// NewSimpleRouter creates a new router instance.
func NewSimpleRouter() *SimpleRouter {
	return &SimpleRouter{routes: make(map[string]object.Object)}
}

// AddRoute adds a Pylearn handler for a specific path.
func (r *SimpleRouter) AddRoute(path string, handler object.Object) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	r.routes[path] = handler
}

// GetHandler finds a handler for a given request path, supporting wildcards.
func (r *SimpleRouter) GetHandler(path string) (object.Object, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, found := r.routes[path]
	if found {
		return handler, true
	}

	for route, handler := range r.routes {
		if strings.HasSuffix(route, "/*") {
			basePath := strings.TrimSuffix(route, "*")
			if strings.HasPrefix(path, basePath) {
				return handler, true
			}
		}
	}
	return nil, false
}

// upgrader handles the WebSocket protocol upgrade.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins for simplicity
}

// pyServeFn is the implementation of the `httpserver.serve` Pylearn function.
func pyServeFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, "serve() takes exactly 2 arguments (address, handler)")
	}
	addrObj, addrOk := args[0].(*object.String)
	if !addrOk {
		return object.NewError(constants.TypeError, "serve() argument 'address' must be a string")
	}
	addressString := addrObj.Value
	handlerArg := args[1]

	var corsWrapper *CORSWrapper
	var pylearnAppHandler object.Object

	// Determine if the provided handler is wrapped with CORS middleware.
	// Check if the handler is a Pylearn CORS object and get its native wrapper.
	if getter, ok := handlerArg.(object.AttributeGetter); ok {
		nativeWrapper, found := getter.GetObjectAttribute(ctx, "_native_wrapper")
		if found {
			if cw, isCorsWrapper := nativeWrapper.(*CORSWrapper); isCorsWrapper {
				corsWrapper = cw
				pylearnAppHandler = cw.PylnHandler
			}
		}
	}
	// If it wasn't a CORS object, treat it as a regular handler.
	if corsWrapper == nil {
		pylearnAppHandler = handlerArg
	}

	// Build the router based on the underlying Pylearn application handler.
	router := buildRouterFromPylearnHandler(pylearnAppHandler)
	if router == nil {
		return object.NewError(constants.TypeError, "serve() handler must be a Dict, a callable, a flasky.App instance, or a CORS-wrapped app, not %s", pylearnAppHandler.Type())
	}

	// This is the core Go handler that bridges the Go HTTP request to the Pylearn interpreter.
	pylearnExecutionHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pylearnHandler, found := router.GetHandler(r.URL.Path)
		if !found {
			// For frameworks registered at "/", try a fallback to the root handler.
			pylearnHandler, found = router.GetHandler("/")
		}

		if !found {
			http.NotFound(w, r)
			return
		}

		handlePylearnRequest(w, r, pylearnHandler, ctx)
	})

	// This is the final Go handler that will be passed to the HTTP server.
	var finalHttpHandler http.Handler = pylearnExecutionHandler

	// If CORS is configured, wrap the Pylearn execution handler with the CORS middleware.
	if corsWrapper != nil {
		finalHttpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			corsWrapper.ServeHTTP(w, r, pylearnExecutionHandler)
		})
	}

	// Start the Go HTTP server.
	server := &http.Server{Addr: addressString, Handler: finalHttpHandler}
	fmt.Printf("INFO: Starting Pylearn HTTP server (async & WebSocket aware) on %s...\n", addressString)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "FATAL: HTTP server error on %s: %v\n", addressString, err)
		}
	}()

	return object.NULL
}

// buildRouterFromPylearnHandler encapsulates the logic to populate the router from a Pylearn object.
func buildRouterFromPylearnHandler(pylearnAppHandler object.Object) *SimpleRouter {
	router := NewSimpleRouter()
	switch routesVal := pylearnAppHandler.(type) {
	case *object.Dict:
		for _, pair := range routesVal.Pairs {
			pathStr, pathOk := pair.Key.(*object.String)
			if !pathOk || !isPylearnCallable(pair.Value) {
				continue
			}
			router.AddRoute(pathStr.Value, pair.Value)
		}
	case *pyflasky.App:
		router.AddRoute("/", routesVal) // Frameworks are catch-all.
	default:
		if isPylearnCallable(routesVal) {
			router.AddRoute("/", routesVal) // A single function is a catch-all.
		} else {
			return nil // Invalid handler type.
		}
	}
	return router
}

// handlePylearnRequest contains the logic for processing an HTTP request and executing the corresponding Pylearn handler.
func handlePylearnRequest(w http.ResponseWriter, r *http.Request, pylearnHandler object.Object, ctx object.ExecutionContext) {
	fmt.Printf("INFO: Server: Received request: %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)

	if fn, ok := pylearnHandler.(*object.Function); ok && fn.IsAsync {
		fmt.Printf("INFO: Server: Calling ASYNC Pylearn HTTP handler for %s\n", r.URL.Path)
	} else {
		fmt.Printf("INFO: Server: Calling SYNC Pylearn HTTP handler for %s\n", r.URL.Path)
	}

	// Convert Go http.Request to a Pylearn HTTPRequest object
	var pylearnRequestObj *object.HTTPRequest
	{
		// --- THIS IS THE CHANGE FOR REQUEST STREAMING ---
		// Don't read the body here. Wrap it in our iterable object.
		var pyRequestBody object.Object
		if r.Body == nil {
			pyRequestBody = object.NULL
		} else {
			pyRequestBody = object.NewHTTPStreamingBody(r.Body)
		}
		// --- END OF CHANGE ---

		headerPairs := make(map[object.HashKey]object.DictPair)
		for key, values := range r.Header {
			if len(values) > 0 {
				pyKey := object.NewString(strings.ToLower(key))
				pyVal := object.NewString(values[0])
				hashKey, _ := pyKey.HashKey()
				headerPairs[hashKey] = object.DictPair{Key: pyKey, Value: pyVal}
			}
		}

		pylearnRequestObj = &object.HTTPRequest{
			Method:  object.NewString(r.Method),
			URL:     object.NewString(r.URL.String()),
			Headers: &object.Dict{Pairs: headerPairs},
			Body:    pyRequestBody, // Assign the new iterable body
		}
	}

	// Create a new execution context for this request.
	var handlerEnv *object.Environment
	if pylFunc, isPylFunc := pylearnHandler.(*object.Function); isPylFunc {
		handlerEnv = pylFunc.Env
	} else {
		handlerEnv = ctx.GetCurrentEnvironment()
	}
	requestExecCtx := ctx.NewChildContext(handlerEnv)

	// Handle WebSocket upgrade requests.
	asyncRuntimeAPI := requestExecCtx.GetAsyncRuntime()
	if websocket.IsWebSocketUpgrade(r) {
		handleWebSocketRequest(w, r, pylearnHandler, pylearnRequestObj, requestExecCtx, asyncRuntimeAPI)
		return
	}

	// Execute the Pylearn handler (sync or async).
	finalPylearnResponseObject := executeHttpRequestHandler(pylearnHandler, pylearnRequestObj, requestExecCtx, asyncRuntimeAPI)

	// Process the Pylearn response object and write to the Go http.ResponseWriter.
	writePylearnResponse(w, r, finalPylearnResponseObject,ctx)
}

// executeHttpRequestHandler runs the Pylearn handler and returns the Pylearn response object.
func executeHttpRequestHandler(pylearnHandler, pylearnRequestObj object.Object, ctx object.ExecutionContext, asyncRT object.AsyncRuntimeAPI) object.Object {
	isAsyncHandler := false
	if pylFunc, isPylFunc := pylearnHandler.(*object.Function); isPylFunc {
		isAsyncHandler = pylFunc.IsAsync
	}

	if isAsyncHandler {
		if asyncRT == nil {
			return object.NewError(constants.InternalServerError, "(Async Runtime Missing for HTTP handler await)")
		}
		asyncResultWrapperObj := ctx.Execute(pylearnHandler, pylearnRequestObj)
		if errWrap, isErr := asyncResultWrapperObj.(*object.Error); isErr {
			return errWrap
		}
		wrapper, okWrap := asyncResultWrapperObj.(*object.AsyncResultWrapper)
		if !okWrap || wrapper.GoAsyncResult == nil {
			return object.NewError(constants.InternalServerError, "(Invalid AsyncResult from async HTTP handler)")
		}
		goValue, goErr := asyncRT.Await(wrapper.GoAsyncResult)
		if goErr != nil {
			return object.NewError(constants.AsyncHTTPHandlerError, "Async HTTP Handler Execution: %v", goErr)
		}
		if pyObj, isPyObj := goValue.(object.Object); isPyObj {
			return pyObj
		}
		return object.NULL
	}

	// Synchronous execution
	return ctx.Execute(pylearnHandler, pylearnRequestObj)
}

// writePylearnResponse translates a Pylearn response object into a Go HTTP response.
func writePylearnResponse(w http.ResponseWriter, r *http.Request, responseObject object.Object, ctx object.ExecutionContext) {
	if object.IsError(responseObject) {
		errObj := responseObject.(*object.Error)
		fmt.Fprintf(os.Stderr, "ERROR: Server: Error in Pylearn HTTP handler for %s: %s (L%d C%d)\n", r.URL.Path, errObj.Message, errObj.Line, errObj.Column)
		http.Error(w, "Internal Server Error (Pylearn Handler Error)", http.StatusInternalServerError)
		return
	}

	// --- THIS IS THE NEW LOGIC FOR RESPONSE STREAMING ---
	// Check if the response is a generator (represented by a Pylearn Function).
	// In a real framework, this might be a dedicated Generator object.
	// Check if the response is a generator object.
	if iterator, isGenerator := responseObject.(object.Iterator); isGenerator && responseObject.Type() == object.GENERATOR_OBJ {
		// This is a streaming response.
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Internal Server Error: Streaming not supported.", 500)
			return
		}

		// Loop over the iterator, streaming each chunk to the client.
		for {
			chunk, stop := iterator.Next()
			if stop { break }

			if object.IsError(chunk) {
				fmt.Fprintf(os.Stderr, "ERROR: Server: Error yielded from stream generator: %s\n", chunk.Inspect())
				break
			}
			
			var dataToWrite []byte
			if b, isBytes := chunk.(*object.Bytes); isBytes {
				dataToWrite = b.Value
			} else if s, isString := chunk.(*object.String); isString {
				dataToWrite = []byte(s.Value)
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: Server: Stream generator yielded non-bytes/non-string type: %s\n", chunk.Type())
				continue
			}
			
			if _, err := w.Write(dataToWrite); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Server: Error writing stream to client: %v\n", err)
				break
			}
			flusher.Flush()
		}
		
		fmt.Printf("INFO: Server: Finished streaming response for %s\n", r.URL.Path)
		return
	}
	// --- END OF NEW LOGIC --

	statusCode := http.StatusOK
	responseBodyBytes := []byte{}
	contentTypeHeader := "text/html; charset=utf-8"

	switch res := responseObject.(type) {
	case *object.String:
		responseBodyBytes = []byte(res.Value)
	case *object.Bytes:
		responseBodyBytes = res.Value
		contentTypeHeader = "application/octet-stream"
	case *object.Dict, *object.List:
		goData, _ := custom_json.ConvertPylearnToInterface(res, 0)
		jsonBytes, _ := gojson.Marshal(goData)
		responseBodyBytes = jsonBytes
		contentTypeHeader = "application/json; charset=utf-8"
	case *object.ServerResponse:
		if res.StatusCode != nil {
			statusCode = int(res.StatusCode.Value)
		}
		if res.ContentType != nil {
			contentTypeHeader = res.ContentType.Value
		}
		if res.Headers != nil {
			goH, _ := object.ConvertPylearnDictToGoHeader(res.Headers)
			for k, v := range goH {
				w.Header()[k] = v
			}
		}
		if res.Body != nil {
			if b, ok := res.Body.(*object.String); ok {
				responseBodyBytes = []byte(b.Value)
			}
			if b, ok := res.Body.(*object.Bytes); ok {
				responseBodyBytes = b.Value
			}
			if d, ok := res.Body.(*object.Dict); ok {
				goData, _ := custom_json.ConvertPylearnToInterface(d, 0)
				jsonBytes, _ := gojson.Marshal(goData)
				responseBodyBytes = jsonBytes
				if res.ContentType == nil {
					contentTypeHeader = "application/json; charset=utf-8"
				}
			}
			if l, ok := res.Body.(*object.List); ok {
				goData, _ := custom_json.ConvertPylearnToInterface(l, 0)
				jsonBytes, _ := gojson.Marshal(goData)
				responseBodyBytes = jsonBytes
				if res.ContentType == nil {
					contentTypeHeader = "application/json; charset=utf-8"
				}
			}
		}
	case *object.Null:
		statusCode = http.StatusNoContent
		contentTypeHeader = ""
	default:
		errMsg := fmt.Sprintf("Server: Pylearn HTTP handler for %s returned unsupported type: %s", r.URL.Path, responseObject.Type())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if w.Header().Get("Content-Type") == "" && contentTypeHeader != "" {
		w.Header().Set("Content-Type", contentTypeHeader)
	}
	w.WriteHeader(statusCode)
	if statusCode != http.StatusNoContent {
		w.Write(responseBodyBytes)
	}
	fmt.Printf("INFO: Server: Responded %d for HTTP %s %s\n", statusCode, r.Method, r.URL.Path)
}

// handleWebSocketRequest manages the lifecycle of a WebSocket connection.
func handleWebSocketRequest(w http.ResponseWriter, r *http.Request, pylearnHandler, pylearnRequestObj object.Object, ctx object.ExecutionContext, asyncRT object.AsyncRuntimeAPI) {
	fmt.Printf("INFO: Server: Attempting WebSocket upgrade for %s\n", r.URL.Path)
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Server: WebSocket upgrade failed for %s: %v\n", r.URL.Path, err)
		return
	}
	defer wsConn.Close()

	pylearnWsObj := websocket_custom_lib.NewWebSocket(wsConn, asyncRT)
	pylFunc, isPylFunc := pylearnHandler.(*object.Function)
	if !isPylFunc || !pylFunc.IsAsync {
		fmt.Fprintf(os.Stderr, "ERROR: Server: WebSocket handler for %s is not an async function.\n", r.URL.Path)
		pylearnWsObj.PyClose(ctx, pylearnWsObj, &object.Integer{Value: int64(websocket.ClosePolicyViolation)}, object.NewString("Handler not async"))
		return
	}

	asyncResultWrapperObj := ctx.Execute(pylearnHandler, pylearnRequestObj, pylearnWsObj)
	wrapper, okWrap := asyncResultWrapperObj.(*object.AsyncResultWrapper)
	if !okWrap || wrapper.GoAsyncResult == nil {
		fmt.Fprintf(os.Stderr, "ERROR: Server: Async Pylearn WebSocket handler for %s did not return valid AsyncResult.\n", r.URL.Path)
		pylearnWsObj.PyClose(ctx, pylearnWsObj, &object.Integer{Value: int64(websocket.CloseInternalServerErr)}, object.NewString("Bad AsyncResult"))
		return
	}

	_, goErr := asyncRT.Await(wrapper.GoAsyncResult)
	if goErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Server: Error from awaited async Pylearn WebSocket handler for %s: %v\n", r.URL.Path, goErr)
	} else {
		fmt.Printf("INFO: Server: Async Pylearn WebSocket handler for %s completed.\n", r.URL.Path)
	}
}

// isPylearnCallable checks if a Pylearn object can be called.
// isPylearnCallable checks if a Pylearn object can be called.
func isPylearnCallable(obj object.Object) bool {
	switch obj.(type) {
	// --- THIS IS THE FIX ---
	// Add CORSWrapper to the list of known callable types.
	case *object.Function, *object.Builtin, *object.Class, *object.BoundMethod, *pyflasky.App, *CORSWrapper:
		return true
	// --- END OF FIX ---
	case *object.Instance:
		inst := obj.(*object.Instance)
		if inst.Class != nil {
			if _, hasCall := inst.Class.Methods["__call__"]; hasCall {
				return true
			}
		}
	}
	return false
}

// utf8IsPrintable provides a basic check to see if a byte slice is likely text.
func utf8IsPrintable(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7F {
			return false
		}
	}
	return true
}

// Serve is the exported Pylearn Builtin for httpserver.serve
var Serve = &object.Builtin{Name: "httpserver.serve", Fn: pyServeFn}
