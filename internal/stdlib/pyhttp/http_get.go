// internal/stdlib/pyhttp/http_get.go
package pyhttp

import (
	// "io/ioutil" // Not directly needed here anymore
	// "net/http"  // Not directly needed here anymore
	// "net/url"   // Not directly needed here anymore

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyHttpGetFn is a wrapper around http.request for GET requests
func pyHttpGetFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// Pylearn: http.get(url, params=None, headers=None)
	// Corresponds to: http.request("GET", url, options={"params": params, "headers": headers})
	if len(args) < 1 || len(args) > 3 {
		return object.NewError(constants.TypeError, "get() takes 1 to 3 arguments (url, params=None, headers=None), got %d", len(args))
	}

	urlArg := args[0] // Should be String

	options := make(map[object.HashKey]object.DictPair)
	
	if len(args) >= 2 && args[1] != object.NULL { // params
		paramsKey := &object.String{Value: "params"}
		hashKey, _ := paramsKey.HashKey()
		options[hashKey] = object.DictPair{Key: paramsKey, Value: args[1]}
	}
	if len(args) == 3 && args[2] != object.NULL { // headers
		headersKey := &object.String{Value: "headers"}
		hashKey, _ := headersKey.HashKey()
		options[hashKey] = object.DictPair{Key: headersKey, Value: args[2]}
	}
	
	optionsDict := &object.Dict{Pairs: options}

	// Call the generic request function
	methodStr := &object.String{Value: "GET"}
	return pyHttpRequestFn(ctx, methodStr, urlArg, optionsDict)
}

// Example of a more flexible pyHttpGetFn (if not using options_dict)
func pyHttpGetFn_flexible(ctx object.ExecutionContext, args ...object.Object) object.Object {
    // Expected signatures from Pylearn:
    // get(url)
    // get(url, params_dict)
    // get(url, params_dict, headers_dict) - less common for requests.get direct params
    // get(url, headers_dict) - if params is None
    // We need to map these to http.request("GET", url, options={"params": ..., "headers": ...})

    if len(args) < 1 { // url is missing
        return object.NewError(constants.TypeError, "get() missing 1 required positional argument: 'url'")
    }
    urlArg := args[0]

    var paramsVal, headersVal object.Object = object.NULL, object.NULL

    // This parsing is simplified. Python's requests.get uses **kwargs
    // For Pylearn without kwargs, we might infer by type or require specific order.
    // Assume: get(url, params_or_headers, headers_if_params_was_first)
    if len(args) >= 2 {
        // If args[1] is a Dict, it could be params or headers.
        // If args[2] exists, then args[1] must be params and args[2] must be headers.
        if len(args) == 3 {
            paramsVal = args[1]
            headersVal = args[2]
        } else { // len(args) == 2
            // Heuristic: if it looks like headers (e.g. contains 'User-Agent'), assume headers.
            // This is fragile. A dedicated options dict is cleaner for Pylearn without kwargs.
            // For now, let's assume args[1] is params if it's a Dict.
            // A better way: get(url, params=None, *, headers=None) if Pylearn supports keyword-only args.
            paramsVal = args[1] // Or require options dict as implemented above.
        }
    }
    
    options := make(map[object.HashKey]object.DictPair)
    if paramsVal != object.NULL {
        paramsKey := &object.String{Value: "params"}
        hashKey, _ := paramsKey.HashKey(); options[hashKey] = object.DictPair{Key: paramsKey, Value: paramsVal}
    }
    if headersVal != object.NULL {
        headersKey := &object.String{Value: "headers"}
        hashKey, _ := headersKey.HashKey(); options[hashKey] = object.DictPair{Key: headersKey, Value: headersVal}
    }
    optionsDict := &object.Dict{Pairs: options}
    
    return pyHttpRequestFn(ctx, &object.String{Value: "GET"}, urlArg, optionsDict)
}

var Get = &object.Builtin{Name: "http.get", Fn: pyHttpGetFn}