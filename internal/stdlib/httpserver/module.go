// pylearn/internal/stdlib/pyhttpserver/module.go
package pyhttpserver

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyNewCorsWrapper is the native constructor for the CORSWrapper object.
func pyNewCorsWrapper(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, "_new_cors_wrapper takes 2 arguments (app, options_dict)")
	}
	appHandler := args[0]
	optionsDict, ok := args[1].(*object.Dict)
	if !ok {
		return object.NewError(constants.TypeError, "_new_cors_wrapper options must be a dict")
	}

	opts := &CORSOptions{}

	extractStringList := func(key string) []string {
		keyObj := object.NewString(key)
		hashKey, _ := keyObj.HashKey()
		if pair, found := optionsDict.Pairs[hashKey]; found {
			if strList, isList := pair.Value.(*object.List); isList {
				list := make([]string, len(strList.Elements))
				for i, item := range strList.Elements {
					if s, isStr := item.(*object.String); isStr {
						list[i] = s.Value
					}
				}
				return list
			} else if strVal, isStr := pair.Value.(*object.String); isStr {
				return []string{strVal.Value}
			}
		}
		return nil
	}

	opts.AllowedOrigins = extractStringList("origins")
	opts.AllowedMethods = extractStringList("methods")
	opts.AllowedHeaders = extractStringList("headers")

	credsKey, _ := object.NewString("allow_credentials").HashKey()
	if credsPair, found := optionsDict.Pairs[credsKey]; found {
		if b, isBool := credsPair.Value.(*object.Boolean); isBool {
			opts.AllowCredentials = b.Value
		}
	}

	maxAgeKey, _ := object.NewString("max_age").HashKey()
	if maxAgePair, found := optionsDict.Pairs[maxAgeKey]; found {
		if i, isInt := maxAgePair.Value.(*object.Integer); isInt {
			opts.MaxAge = int(i.Value)
		}
	}

	return &CORSWrapper{
		Options:    opts,
		PylnHandler: appHandler,
	}
}

func init() {
	env := object.NewEnvironment()

	env.Set("serve", Serve)
	env.Set("json_response", JsonResponse)
	env.Set("text_response", TextResponse)
	env.Set("_new_cors_wrapper", &object.Builtin{Name: "httpserver._new_cors_wrapper", Fn: pyNewCorsWrapper})

	webSocketErrorClass := object.CreateExceptionClass(constants.WebSocketClosedError, object.ExceptionClass)
	object.BuiltinExceptionClasses[constants.WebSocketClosedError] = webSocketErrorClass
	env.Set("WebSocketClosedError", webSocketErrorClass)

	httpServerModule := &object.Module{
		Name: "httpserver",
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("httpserver", httpServerModule)
}