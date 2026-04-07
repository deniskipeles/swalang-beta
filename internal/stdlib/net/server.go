// internal/stdlib/net/server.go
package net

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	custom_json "github.com/deniskipeles/pylearn/internal/stdlib/json" // <<< FIX: Import the correct JSON converter
)

const (
	APP_OBJ      object.ObjectType = "net.App"
	REQUEST_OBJ  object.ObjectType = "net.Request"
	RESPONSE_OBJ object.ObjectType = "net.Response"
)

var (
	AppClass      *object.Class
	RequestClass  *object.Class
	ResponseClass *object.Class
)

// --- App Object (The Web Framework Core) ---
type App struct {
	mu              sync.RWMutex
	GoServer        *http.Server
	routes          map[string]*object.Function
	middlewares     []*object.Function
	ctx             object.ExecutionContext
	pylearnInstance *AppInstance
}

func (a *App) Type() object.ObjectType { return APP_OBJ }
func (a *App) Inspect() string {
	addr := "not running"
	if a.GoServer != nil && a.GoServer.Addr != "" {
		addr = a.GoServer.Addr
	}
	return fmt.Sprintf("<net.App listening on %s>", addr)
}
func (a *App) GetGoHTTPServer() *http.Server { return a.GoServer }
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := newRequest(r)
	res := newResponse(w)

	handler := func(rq, rs object.Object) object.Object { // <<< FIX: Handler now returns an object
		app.mu.RLock()
		routeKey := r.Method + " " + r.URL.Path
		routeHandler, found := app.routes[routeKey]
		app.mu.RUnlock()

		if found {
			result := app.ctx.Execute(routeHandler, rq, rs)
			return result // Return the result from the handler
		} else {
			res.SetStatusCode(404)
			res.SetBody([]byte("Not Found"))
			return object.NULL
		}
	}

	app.mu.RLock()
	wrappedHandler := object.Object(&object.Builtin{
		Name: "net.App.base_handler",
		Fn: func(ctx object.ExecutionContext, args ...object.Object) object.Object {
			return handler(args[0], args[1]) // <<< FIX: Return the handler's result
		},
	})
	for i := len(app.middlewares) - 1; i >= 0; i-- {
		middleware := app.middlewares[i]
		wrappedHandler = app.ctx.Execute(middleware, wrappedHandler)
		if object.IsError(wrappedHandler) {
			res.SetStatusCode(500)
			res.SetBody([]byte("Error in middleware chain: " + wrappedHandler.Inspect()))
			res.Finalize()
			app.mu.RUnlock()
			return
		}
	}
	app.mu.RUnlock()

	// Execute the final handler and get its result
	finalResult := app.ctx.Execute(wrappedHandler, req, res)

	// --- START OF STREAMING BODY FIX ---
	switch result := finalResult.(type) {
	case *StreamingBody:
		// The handler returned a streaming body.
		// We assume headers and status have already been sent.
		defer result.Source.Close() // Ensure the file handle is closed

		// Use io.Copy for efficient, chunked streaming from the source reader
		// to the HTTP response writer.
		_, err := io.Copy(w, result.Source)
		if err != nil {
			// Cannot send a 500 error here as headers are already sent.
			// Log the error server-side.
			fmt.Printf("Error during response stream: %v\n", err)
		}

	case *object.Error:
		// The handler raised an exception.
		res.SetStatusCode(500)
		res.SetBody([]byte(result.Inspect()))
		res.Finalize()

	default:
		// The handler finished without returning a streaming body or raising an exception.
		// The response should have been written using res.body() or res.json().
		res.Finalize()
	}
	// --- END OF STREAMING BODY FIX ---
}

type AppInstance struct {
	*object.Instance
	NativeApp *App
}

func (ai *AppInstance) Type() object.ObjectType       { return APP_OBJ }
func (ai *AppInstance) Inspect() string               { return ai.NativeApp.Inspect() }
func (ai *AppInstance) GetGoHTTPServer() *http.Server { return ai.NativeApp.GoServer }
func (ai *AppInstance) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	method, found := AppClass.Methods[name]
	if !found {
		return nil, false
	}
	boundMethod := &object.Builtin{
		Name: fmt.Sprintf("%s.%s", ai.Class.Name, name),
		Fn: func(callCtx object.ExecutionContext, scriptArgs ...object.Object) object.Object {
			finalArgs := make([]object.Object, 1+len(scriptArgs))
			finalArgs[0] = ai
			copy(finalArgs[1:], scriptArgs)
			return method.(*object.Builtin).Fn(callCtx, finalArgs...)
		},
	}
	return boundMethod, true
}

// --- Request Object ---
type Request struct {
	GoRequest  *http.Request
	cachedBody []byte
}

func newRequest(r *http.Request) *Request  { return &Request{GoRequest: r} }
func (r *Request) Type() object.ObjectType { return REQUEST_OBJ }
func (r *Request) Inspect() string {
	return fmt.Sprintf("<net.Request %s %s>", r.GoRequest.Method, r.GoRequest.URL.Path)
}
func (r *Request) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	switch name {
	case "method":
		return &object.String{Value: r.GoRequest.Method}, true
	case "path":
		return &object.String{Value: r.GoRequest.URL.Path}, true
	}
	return object.GetNativeMethod(r, name)
}

// --- Response Object ---
type Response struct {
	GoWriter  http.ResponseWriter
	finalized bool
}

func newResponse(w http.ResponseWriter) *Response { return &Response{GoWriter: w} }
func (r *Response) Type() object.ObjectType       { return RESPONSE_OBJ }
func (r *Response) Inspect() string               { return fmt.Sprintf("<net.Response at %p>", r) }
func (r *Response) SetStatusCode(code int)        { r.GoWriter.WriteHeader(code) }
func (r *Response) SetBody(body []byte)           { r.GoWriter.Write(body) }
func (r *Response) Finalize()                     { r.finalized = true }
func (r *Response) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	return object.GetNativeMethod(r, name)
}

// --- Native Pylearn Methods ---
func appRoute(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*AppInstance)
	path, _ := args[1].(*object.String)
	method, _ := args[2].(*object.String)
	decorator := func(decoratorArgs ...object.Object) object.Object {
		handlerFunc, _ := decoratorArgs[0].(*object.Function)
		self.NativeApp.mu.Lock()
		self.NativeApp.routes[strings.ToUpper(method.Value)+" "+path.Value] = handlerFunc
		self.NativeApp.mu.Unlock()
		return handlerFunc
	}
	return object.NewBuiltin("net.App.route_decorator", decorator)
}

func appMiddleware(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*AppInstance)
	middlewareFunc, _ := args[1].(*object.Function)
	self.NativeApp.mu.Lock()
	self.NativeApp.middlewares = append(self.NativeApp.middlewares, middlewareFunc)
	self.NativeApp.mu.Unlock()
	return middlewareFunc
}

func appRun(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*AppInstance)
	addr, _ := args[1].(*object.String)
	self.NativeApp.GoServer = &http.Server{Addr: addr.Value, Handler: self.NativeApp}
	fmt.Printf("Pylearn server running on %s\n", addr.Value)
	err := self.NativeApp.GoServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return object.NewError("ServerError", "failed to start server: %v", err)
	}
	return object.NULL
}

func requestGetBody(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Request)
	if self.cachedBody != nil {
		return &object.Bytes{Value: self.cachedBody}
	}
	body, err := ioutil.ReadAll(self.GoRequest.Body)
	if err != nil {
		return object.NewError("IOError", "failed to read request body: %v", err)
	}
	self.cachedBody = body
	return &object.Bytes{Value: body}
}

func requestGetJSON(ctx object.ExecutionContext, args ...object.Object) object.Object {
	bodyObj := requestGetBody(ctx, args...)
	if err, ok := bodyObj.(*object.Error); ok {
		return err
	}
	bodyBytes := bodyObj.(*object.Bytes).Value
	if len(bodyBytes) == 0 {
		return object.NewError("JSONDecodeError", "Empty body cannot be decoded as JSON")
	}
	var goData interface{}
	err := json.Unmarshal(bodyBytes, &goData)
	if err != nil {
		return object.NewError("JSONDecodeError", err.Error())
	}
	pylearnObj, _ := custom_json.ConvertInterfaceToPylearn(goData, 0)
	return pylearnObj
}

func responseSetBody(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Response)
	data, ok := args[1].(*object.Bytes)
	if !ok {
		return object.NewError(constants.TypeError, "body() argument must be bytes")
	}
	self.SetBody(data.Value)
	return object.NULL
}

func responseSetJSON(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Response)
	pylearnObj := args[1]
	goData, err := custom_json.ConvertPylearnToInterface(pylearnObj, 0)
	if err != nil {
		return object.NewError("JSONEncodeError", err.Error())
	}
	jsonBytes, err := json.Marshal(goData)
	if err != nil {
		return object.NewError("JSONEncodeError", err.Error())
	}
	self.GoWriter.Header().Set("Content-Type", "application/json")
	self.SetBody(jsonBytes)
	return object.NULL
}

func responseSetHeader(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Response)
	key, _ := args[1].(*object.String)
	value, _ := args[2].(*object.String)
	self.GoWriter.Header().Set(key.Value, value.Value)
	return object.NULL
}
func responseSetStatus(ctx object.ExecutionContext, args ...object.Object) object.Object {
	self := args[0].(*Response)
	code, ok := args[1].(*object.Integer)
	if !ok {
		return object.NewError(constants.TypeError, "status() argument must be an integer")
	}
	self.SetStatusCode(int(code.Value))
	return object.NULL // Return NULL as it modifies the response in place
}
func pyAppConstructor(ctx object.ExecutionContext, args ...object.Object) object.Object {
	nativeApp := &App{
		routes:      make(map[string]*object.Function),
		middlewares: make([]*object.Function, 0),
		ctx:         ctx,
	}
	appInstance := &AppInstance{
		Instance:  &object.Instance{Class: AppClass, Env: object.NewEnvironment()},
		NativeApp: nativeApp,
	}
	nativeApp.pylearnInstance = appInstance
	return appInstance
}

func initServer() {
	AppClass = object.CreateExceptionClass("App", object.ObjectClass)
	AppClass.Methods["route"] = &object.Builtin{Fn: appRoute, Name: "net.App.route"}
	AppClass.Methods["middleware"] = &object.Builtin{Fn: appMiddleware, Name: "net.App.middleware"}
	AppClass.Methods["run"] = &object.Builtin{Fn: appRun, Name: "net.App.run"}

	RequestClass = object.CreateExceptionClass("Request", object.ObjectClass)
	RequestClass.Methods["body"] = &object.Builtin{Fn: requestGetBody, Name: "net.Request.body"}
	RequestClass.Methods["json"] = &object.Builtin{Fn: requestGetJSON, Name: "net.Request.json"}
	object.SetNativeMethod(REQUEST_OBJ, "body", requestGetBody)
	object.SetNativeMethod(REQUEST_OBJ, "json", requestGetJSON)

	ResponseClass = object.CreateExceptionClass("Response", object.ObjectClass)
	ResponseClass.Methods["json"] = &object.Builtin{Fn: responseSetJSON, Name: "net.Response.json"}
	ResponseClass.Methods["header"] = &object.Builtin{Fn: responseSetHeader, Name: "net.Response.header"}
	ResponseClass.Methods["body"] = &object.Builtin{Fn: responseSetBody, Name: "net.Response.body"}
	object.SetNativeMethod(RESPONSE_OBJ, "json", responseSetJSON)
	object.SetNativeMethod(RESPONSE_OBJ, "header", responseSetHeader)
	object.SetNativeMethod(RESPONSE_OBJ, "body", responseSetBody)
	object.SetNativeMethod(RESPONSE_OBJ, "status", responseSetStatus)
}
