// pylearn/internal/stdlib/pyflasky/app_object.go
package pyflasky

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

const APP_OBJ = "App"

// RouteInfo stores a handler function and its metadata.
type RouteInfo struct {
	Handler    object.Object
	ParamNames []string
	Methods    map[string]bool // A set of allowed methods (e.g., "GET": true)
}

// The main App object for our framework.
type App struct {
	routes map[string]*RouteInfo
}

func (app *App) Type() object.ObjectType { return APP_OBJ }
func (app *App) Inspect() string {
	return fmt.Sprintf("<Flasky App with %d routes>", len(app.routes))
}

// This is the Go constructor for the App object.
func NewApp() *App {
	return &App{
		routes: make(map[string]*RouteInfo),
	}
}

// --- Pylearn Methods for App Object ---

// This is the Go implementation for the `route` method.
// It's wrapped in a Builtin that declares its accepted keywords.
func appRoute(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// The interpreter passes `self` as the first argument.
	if len(args) < 2 {
		return object.NewError(constants.TypeError, "route() missing 1 required positional argument: 'path'")
	}
	self, ok := args[0].(*App)
	if !ok {
		return object.NewError(constants.InternalError, "route() did not receive App instance as self")
	}
	pathObj, ok := args[1].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "route path must be a string")
	}

	// Default to allowing GET if methods are not specified.
	allowedMethods := map[string]bool{"GET": true}

	// Check for the optional keyword argument dictionary, which is the last argument.
	if len(args) > 2 {
		kwargs, ok := args[len(args)-1].(*object.Dict)
		if ok { // It's only kwargs if it's a dict
			methodsKey := &object.String{Value: "methods"}
			methodsHash, _ := methodsKey.HashKey()
			if pair, found := kwargs.Pairs[methodsHash]; found {
				if methodsList, isList := pair.Value.(*object.List); isList {
					allowedMethods = make(map[string]bool) // Reset the default
					for _, item := range methodsList.Elements {
						if methodStr, isStr := item.(*object.String); isStr {
							allowedMethods[strings.ToUpper(methodStr.Value)] = true
						}
					}
				}
			}
		}
	}

	// The decorator logic
	decorator := &object.Builtin{
		Name: "route.decorator",
		Fn: func(decoratorCtx object.ExecutionContext, decoratorArgs ...object.Object) object.Object {
			if len(decoratorArgs) != 1 {
				return object.NewError(constants.TypeError, "decorator expects 1 argument (the view function)")
			}
			handlerFunc := decoratorArgs[0]
			if !isPylearnCallable(handlerFunc) {
				return object.NewError(constants.TypeError, "View function is not callable")
			}
			path := pathObj.Value
			paramNames := []string{}
			parts := strings.Split(path, "/")
			for _, part := range parts {
				if strings.HasPrefix(part, "<") && strings.HasSuffix(part, ">") {
					paramNames = append(paramNames, part[1:len(part)-1])
				}
			}
			// Use the `self` captured from the appRoute scope.
			self.routes[path] = &RouteInfo{
				Handler:    handlerFunc,
				ParamNames: paramNames,
				Methods:    allowedMethods, // <<< FIX: Correctly use the 'Methods' field
			}
			return handlerFunc
		},
	}
	return decorator
}

// The Go implementation for the __call__ method.
func appCall(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 { // self, request
		return object.NewError(constants.InternalError, "App.__call__ expects 1 argument (request)")
	}
	self, _ := args[0].(*App)
	requestObj, _ := args[1].(*object.HTTPRequest)

	methodAttr, _ := requestObj.GetObjectAttribute(ctx, "method")
	requestMethod := methodAttr.(*object.String).Value

	urlAttr, _ := requestObj.GetObjectAttribute(ctx, "url")
	requestURL := urlAttr.(*object.String).Value
	parsedURL, _ := url.Parse(requestURL)
	requestPath := parsedURL.Path

	var handlerInfo *RouteInfo
	var pathParams map[string]string
	for routePath, info := range self.routes {
		params, match := matchRoute(routePath, requestPath)
		if match {
			handlerInfo = info
			pathParams = params
			break
		}
	}

	if handlerInfo == nil {
		return &object.ServerResponse{Body: object.NewString("<h1>404 Not Found</h1>"), StatusCode: &object.Integer{Value: 404}}
	}

	if _, methodIsAllowed := handlerInfo.Methods[requestMethod]; !methodIsAllowed {
		return &object.ServerResponse{Body: object.NewString("<h1>405 Method Not Allowed</h1>"), StatusCode: &object.Integer{Value: 405}}
	}

	handlerArgs := []object.Object{requestObj}
	for _, paramName := range handlerInfo.ParamNames {
		paramValue, _ := pathParams[paramName]
		handlerArgs = append(handlerArgs, &object.String{Value: paramValue})
	}

	return ctx.Execute(handlerInfo.Handler, handlerArgs...)
}

// In GetObjectAttribute, we return a Builtin that is configured to accept keywords
// and whose function body knows how to call our implementation with `self`.
func (app *App) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == "route" {
		return &object.Builtin{
			Name: "App.route",
			Fn: func(bCtx object.ExecutionContext, bArgs ...object.Object) object.Object {
				// Prepend `self` (the `app` instance) to the arguments and call the real implementation.
				allArgs := make([]object.Object, len(bArgs)+1)
				allArgs[0] = app
				copy(allArgs[1:], bArgs)
				return appRoute(bCtx, allArgs...)
			},
			AcceptsKeywords: map[string]bool{"methods": true}, // Signal that it accepts this keyword
		}, true
	}
	if name == "__call__" {
		return &object.Builtin{
			Name: "App.__call__",
			Fn: func(bCtx object.ExecutionContext, bArgs ...object.Object) object.Object {
				allArgs := make([]object.Object, len(bArgs)+1)
				allArgs[0] = app
				copy(allArgs[1:], bArgs)
				return appCall(bCtx, allArgs...)
			},
		}, true
	}
	return nil, false
}

func matchRoute(routePath, requestPath string) (map[string]string, bool) {
	routeParts := strings.Split(strings.Trim(routePath, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(routeParts) == 1 && routeParts[0] == "" {
		routeParts = []string{}
	}
	if len(requestParts) == 1 && requestParts[0] == "" {
		requestParts = []string{}
	}
	if len(routeParts) != len(requestParts) {
		return nil, false
	}
	params := make(map[string]string)
	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, "<") && strings.HasSuffix(routePart, ">") {
			paramName := routePart[1 : len(routePart)-1]
			params[paramName] = requestParts[i]
		} else if routePart != requestParts[i] {
			return nil, false
		}
	}
	return params, true
}

// --- ADD THIS HELPER FUNCTION ---
// Replicated from httpserver/server.go to avoid circular dependencies.
func isPylearnCallable(obj object.Object) bool {
	switch obj.(type) {
	case *object.Function, *object.Builtin, *object.Class, *object.BoundMethod, *object.BoundGoMethod:
		return true
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

var _ object.Object = (*App)(nil)
var _ object.AttributeGetter = (*App)(nil)
