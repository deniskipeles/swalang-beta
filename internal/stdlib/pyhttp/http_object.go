// internal/stdlib/pyhttp/http_object.go
package pyhttp

import (
	// "bytes" // Not needed if helpers are moved
	gojson "encoding/json"
	"fmt"

	// "io/ioutil" // Not needed if helpers are moved
	// "net/http" // Not needed if helpers are moved
	// "net/url" // Not needed if helpers are moved
	// "strings"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

const (
	HTTP_RESPONSE_OBJ object.ObjectType = "HTTPResponse"
	HTTP_REQUEST_OBJ  object.ObjectType = "HTTPRequest"
)

// --- Pylearn HTTPResponse Object ---
type HTTPResponse struct {
	StatusCode *object.Integer
	Content    *object.Bytes
	Headers    *object.Dict
	URL        *object.String
	Encoding   *object.String
	Reason     *object.String
}

func (hr *HTTPResponse) Type() object.ObjectType { return HTTP_RESPONSE_OBJ }
func (hr *HTTPResponse) Inspect() string {
	statusCode := "?"; if hr.StatusCode != nil { statusCode = hr.StatusCode.Inspect() }
	urlStr := ""; if hr.URL != nil { urlStr = hr.URL.Value }
	return fmt.Sprintf("<HTTPResponse [%s] for %s>", statusCode, urlStr)
}
var _ object.Object = (*HTTPResponse)(nil)
var _ object.AttributeGetter = (*HTTPResponse)(nil)

func ConvertGoInterfaceToPylearnForHTTP(data interface{}, depth int) (object.Object, error) {
	if depth > 64 {
		return nil, fmt.Errorf("max recursion depth for JSON conversion")
	}
	switch v := data.(type) {
	case nil: return object.NULL, nil
	case bool: return object.NativeBoolToBooleanObject(v), nil
	case float64:
		if v == float64(int64(v)) { return &object.Integer{Value: int64(v)}, nil }
		return &object.Float{Value: v}, nil
	case string: return &object.String{Value: v}, nil
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, item := range v {
			converted, err := ConvertGoInterfaceToPylearnForHTTP(item, depth+1)
			if err != nil { return nil, err }
			elements[i] = converted
		}
		return &object.List{Elements: elements}, nil
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.DictPair, len(v))
		for key, val := range v {
			pyKey := &object.String{Value: key}
			pyVal, err := ConvertGoInterfaceToPylearnForHTTP(val, depth+1)
			if err != nil { return nil, err }
			hashKey, _ := pyKey.HashKey()
			pairs[hashKey] = object.DictPair{Key: pyKey, Value: pyVal}
		}
		return &object.Dict{Pairs: pairs}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T in JSON conversion", v)
	}
}


func (hr *HTTPResponse) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	switch name {
	case "status_code": if hr.StatusCode == nil { return object.NULL, true }; return hr.StatusCode, true
	case "text":
		if hr.Content == nil { return &object.String{Value: ""}, true }
		return &object.String{Value: string(hr.Content.Value)}, true
	case "content": if hr.Content == nil { return &object.Bytes{Value: []byte{}}, true }; return hr.Content, true
	case "headers": if hr.Headers == nil { return &object.Dict{Pairs: make(map[object.HashKey]object.DictPair)}, true }; return hr.Headers, true
	case "url": if hr.URL == nil { return object.NULL, true }; return hr.URL, true
	case "encoding": if hr.Encoding == nil { return object.NULL, true }; return hr.Encoding, true
	case "reason": if hr.Reason == nil { return object.NULL, true }; return hr.Reason, true
	case "json":
		return &object.Builtin{
			Name: "HTTPResponse.json",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 1 { return object.NewError(constants.TypeError, "json() takes no arguments (%d given)", len(args)-1) }
				selfResp, ok := args[0].(*HTTPResponse)
				if !ok { return object.NewError(constants.TypeError, "json() must be called on an HTTPResponse object") }
				if selfResp.Content == nil || len(selfResp.Content.Value) == 0 { return object.NewError(constants.JSONDecodeError, "Empty response content") }
				var goData interface{}; err := gojson.Unmarshal(selfResp.Content.Value, &goData)
				if err != nil { return object.NewError(constants.JSONDecodeError, "%s (from: %q)", err.Error(), string(selfResp.Content.Value)) }
				// Call the helper from the same package (now pyhttp)
				pylearnObj, convertErr := ConvertGoInterfaceToPylearnForHTTP(goData, 0)
				if convertErr != nil { return object.NewError(constants.JSONDecodeError, "Converting JSON to Pylearn: %v", convertErr) }
				return pylearnObj
			},
		}, true
	case "raise_for_status":
		return &object.Builtin{
			Name: "HTTPResponse.raise_for_status",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 1 { return object.NewError(constants.TypeError, "raise_for_status() takes no arguments") }
				selfResp, ok := args[0].(*HTTPResponse)
				if !ok { return object.NewError(constants.TypeError, "raise_for_status() on non-HTTPResponse") }
				if selfResp.StatusCode == nil { return object.NULL }
				statusCode := selfResp.StatusCode.Value; urlVal := ""; if selfResp.URL != nil {urlVal = selfResp.URL.Value}; reasonVal := ""; if selfResp.Reason != nil {reasonVal = selfResp.Reason.Value}
				if statusCode >= 400 && statusCode < 500 { return object.NewError(constants.HTTPClientError, "%d %s for url: %s", statusCode, reasonVal, urlVal) }
				if statusCode >= 500 && statusCode < 600 { return object.NewError(constants.HTTPServerError, "%d %s for url: %s", statusCode, reasonVal, urlVal) }
				return object.NULL
			},
		}, true
	}
	return nil, false
}

// --- Pylearn HTTPRequest Object ---
type HTTPRequest struct {
	Method  *object.String
	URL     *object.String
	Headers *object.Dict
	Body    object.Object
}

func (hr *HTTPRequest) Type() object.ObjectType { return HTTP_REQUEST_OBJ }
func (hr *HTTPRequest) Inspect() string {
	method, urlStr := "?", "/"; if hr.Method != nil { method = hr.Method.Value }; if hr.URL != nil { urlStr = hr.URL.Value }
	return fmt.Sprintf("<HTTPRequest [%s %s]>", method, urlStr)
}
var _ object.Object = (*HTTPRequest)(nil)
var _ object.AttributeGetter = (*HTTPRequest)(nil)

func (hr *HTTPRequest) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	switch name {
	case "method": if hr.Method == nil { return object.NULL, true }; return hr.Method, true
	case "url": if hr.URL == nil { return object.NULL, true }; return hr.URL, true
	case "headers": if hr.Headers == nil { return &object.Dict{Pairs: make(map[object.HashKey]object.DictPair)}, true }; return hr.Headers, true
	case "body": if hr.Body == nil { return object.NULL, true }; return hr.Body, true
	}
	return nil, false
}