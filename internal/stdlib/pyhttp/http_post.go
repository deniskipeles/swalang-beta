// internal/stdlib/pyhttp/http_post.go
package pyhttp

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyHttpPostFn is a wrapper around http.request for POST requests
func pyHttpPostFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// Pylearn: http.post(url, data=None, json=None, headers=None)
	// Corresponds to: http.request("POST", url, options={...})
	// Args: url, [data_or_json_options_dict] or [data_or_json, headers_dict]
	// Simplified for now: url, options_dict
	if len(args) < 1 || len(args) > 2 { // url, [options_dict]
		return object.NewError(constants.TypeError, "post() takes 1 or 2 arguments (url, options_dict=None), got %d", len(args))
	}
	
	urlArg := args[0]
	var optionsDict *object.Dict
	
	if len(args) == 2 {
		if args[1] != object.NULL {
			dict, ok := args[1].(*object.Dict)
			if !ok {
				return object.NewError(constants.TypeError, "post() second argument, if provided, must be a Dict of options or None")
			}
			optionsDict = dict
		}
	}
	if optionsDict == nil { // Create empty if not provided
	    optionsDict = &object.Dict{Pairs: make(map[object.HashKey]object.DictPair)}
	}


	// Call the generic request function
	methodStr := &object.String{Value: "POST"}
	return pyHttpRequestFn(ctx, methodStr, urlArg, optionsDict)
}

var Post = &object.Builtin{Name: "http.post", Fn: pyHttpPostFn}