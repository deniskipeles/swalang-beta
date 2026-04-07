// internal/stdlib/pyhttp/http_request.go
package pyhttp

import (
	"bytes"
	gojson "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time" // For timeout
	"context"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// pyHttpRequestFn is the generic Pylearn http.request() function
func pyHttpRequestFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// --- Argument parsing logic remains exactly the same as before ---
	if len(args) < 2 {
		return object.NewError(constants.TypeError, "request() missing 2 required positional arguments: 'method' and 'url'")
	}
	methodObj, methodOk := args[0].(*object.String)
	if !methodOk {
		return object.NewError(constants.TypeError, "request() method argument must be a string, not %s", args[0].Type())
	}
	method := strings.ToUpper(methodObj.Value)
	urlObj, urlOk := args[1].(*object.String)
	if !urlOk {
		return object.NewError(constants.TypeError, "request() url argument must be a string, not %s", args[1].Type())
	}
	urlString := urlObj.Value
	var paramsObj, dataObj, jsonPayloadObj, headersObj, timeoutObj object.Object
	paramsObj, dataObj, jsonPayloadObj, headersObj, timeoutObj = object.NULL, object.NULL, object.NULL, object.NULL, object.NULL
	if len(args) > 2 {
		optionsDict, dictOk := args[2].(*object.Dict)
		if !dictOk {
			return object.NewError(constants.TypeError, "request() third argument, if provided, must be a Dict of options, not %s", args[2].Type())
		}
		for _, pair := range optionsDict.Pairs {
			keyStr, keyIsStr := pair.Key.(*object.String)
			if !keyIsStr { continue }
			switch keyStr.Value {
			case "params": paramsObj = pair.Value
			case "data": dataObj = pair.Value
			case "json": jsonPayloadObj = pair.Value
			case "headers": headersObj = pair.Value
			case "timeout": timeoutObj = pair.Value
			}
		}
	}
	if dataObj != object.NULL && jsonPayloadObj != object.NULL {
		return object.NewError(constants.ValueError, "cannot specify both 'data' and 'json' arguments")
	}

	// --- Prepare URL, Body, and Headers (this logic also remains the same) ---
	baseURL := urlString
	if paramsObj != object.NULL {
		paramsQuery, err := object.ConvertPylearnDictToURLValues(paramsObj)
		if err != nil {
			return object.NewError(constants.TypeError, constants.HTTP_REQUEST_PARAMS_CONVERSION_ERROR, err)
		}
		if paramsQuery != "" {
			if strings.Contains(baseURL, "?") {
				baseURL += "&" + paramsQuery
			} else {
				baseURL += "?" + paramsQuery
			}
		}
	}

	var reqBody io.Reader
	var finalContentType string
	if jsonPayloadObj != object.NULL {
		goData, err := ConvertPylearnToInterfaceForHTTP(jsonPayloadObj, 0)
		if err != nil {
			return object.NewError(constants.TypeError, constants.HTTP_REQUEST_JSON_CONVERSION_ERROR, err)
		}
		jsonBytes, err := gojson.Marshal(goData)
		if err != nil {
			return object.NewError(constants.ValueError, constants.HTTP_REQUEST_JSON_MARSHAL_ERROR, err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
		finalContentType = "application/json"
	} else if dataObj != object.NULL {
		switch d := dataObj.(type) {
		case *object.String:
			reqBody = strings.NewReader(d.Value)
			finalContentType = "application/x-www-form-urlencoded"
		case *object.Bytes:
			reqBody = bytes.NewReader(d.Value)
			finalContentType = "application/octet-stream"
		case *object.Dict:
			formDataQuery, err := object.ConvertPylearnDictToURLValues(d)
			if err != nil {
				return object.NewError(constants.TypeError, "error converting data dictionary for form: %v", err)
			}
			reqBody = strings.NewReader(formDataQuery)
			finalContentType = "application/x-www-form-urlencoded"
		default:
			return object.NewError(constants.TypeError, constants.HTTP_REQUEST_DATA_TYPE_ERROR, d.Type())
		}
	}

	// --- START OF THE FIX: WRAP THE HTTP CALL IN A COROUTINE ---
	asyncRuntime := ctx.GetAsyncRuntime()
	if asyncRuntime == nil {
		return object.NewError(constants.RuntimeError, "Async runtime not available. Cannot perform async http request.")
	}

	// Create a coroutine that will perform the blocking I/O.
	goAsyncResult := asyncRuntime.CreateCoroutine(func(goCtx context.Context) (interface{}, error) {
		// This code runs in a separate goroutine managed by the event loop.

		goReq, err := http.NewRequest(method, baseURL, reqBody)
		if err != nil {
			// Return a Go error, which the await logic will handle.
			return nil, fmt.Errorf(constants.HTTP_REQUEST_CREATE_ERROR, err)
		}

		if headersObj != object.NULL {
			userHeaders, errHeader := object.ConvertPylearnDictToGoHeader(headersObj)
			if errHeader != nil {
				return nil, fmt.Errorf("error converting headers: %v", errHeader)
			}
			for k, v := range userHeaders {
				goReq.Header[k] = v
			}
		}

		if reqBody != nil && goReq.Header.Get("Content-Type") == "" && finalContentType != "" {
			goReq.Header.Set("Content-Type", finalContentType)
		}

		client := http.DefaultClient
		if timeoutObj != object.NULL {
			var timeoutVal float64
			switch t := timeoutObj.(type) {
			case *object.Integer: timeoutVal = float64(t.Value)
			case *object.Float: timeoutVal = t.Value
			default:
				// Return a Pylearn error as the coroutine's result value
				return object.NewError(constants.TypeError, constants.HTTP_REQUEST_TIMEOUT_VALUE_ERROR, t.Type()), nil
			}
			if timeoutVal <= 0 {
				return object.NewError(constants.ValueError, constants.HTTP_REQUEST_TIMEOUT_POSITIVE_ERROR), nil
			}
			client = &http.Client{Timeout: time.Duration(timeoutVal * float64(time.Second))}
		}

		goResp, err := client.Do(goReq)
		if err != nil {
			if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
				// Return a specific Pylearn error object as the value
				return object.NewError(constants.TimeoutError, constants.HTTP_REQUEST_TIMEOUT_ERROR, baseURL), nil
			}
			// Return a general Pylearn request error as the value
			return object.NewError(constants.RequestError, constants.HTTP_REQUEST_EXECUTION_ERROR, baseURL, err), nil
		}

		// On success, populate the Pylearn HTTPResponse object.
		// This becomes the successful result of the coroutine.
		pyResp, errResp := object.PopulateResponseObject(goResp, baseURL)
		if errResp != nil {
			return errResp, nil // Return the Pylearn error object as the value
		}
		
		return pyResp, nil // Return the successful Pylearn response object
	})

	// Immediately return the awaitable wrapper.
	return &object.AsyncResultWrapper{GoAsyncResult: goAsyncResult}
	// --- END OF THE FIX ---
}

// ConvertPylearnToInterfaceForHTTP (needed for json payload)
// This is a simplified converter, similar to the one in pyjson stdlib
// For a full implementation, this should ideally be shared or use the pyjson one.
func ConvertPylearnToInterfaceForHTTP(obj object.Object, depth int) (interface{}, error) {
	if depth > 64 { return nil, fmt.Errorf("max recursion depth") }
	switch o := obj.(type) {
	case *object.Null: return nil, nil
	case *object.Boolean: return o.Value, nil
	case *object.Integer: return o.Value, nil
	case *object.Float: return o.Value, nil
	case *object.String: return o.Value, nil
	case *object.List:
		arr := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			converted, err := ConvertPylearnToInterfaceForHTTP(elem, depth+1)
			if err != nil { return nil, err }
			arr[i] = converted
		}
		return arr, nil
	case *object.Dict:
		m := make(map[string]interface{}, len(o.Pairs))
		for _, pair := range o.Pairs {
			keyStr, ok := pair.Key.(*object.String)
			if !ok { return nil, fmt.Errorf("json payload dict keys must be strings, got %s", pair.Key.Type()) }
			converted, err := ConvertPylearnToInterfaceForHTTP(pair.Value, depth+1)
			if err != nil { return nil, err }
			m[keyStr.Value] = converted
		}
		return m, nil
	default:
		return nil, fmt.Errorf("type %s not serializable to JSON for request payload", obj.Type())
	}
}

var Request = &object.Builtin{Name: "http.request", Fn: pyHttpRequestFn}