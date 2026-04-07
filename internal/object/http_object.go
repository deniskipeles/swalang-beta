// internal/object/http_object.go

package object

import (
	gojson "encoding/json" // Alias to avoid conflict
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// const (
// 	HTTP_RESPONSE_OBJ ObjectType = "HTTPResponse"
// 	HTTP_REQUEST_OBJ  ObjectType = "HTTPRequest"
// 	// No explicit HTTPError class for now, using formatted Error
// )

// --- Pylearn HTTPResponse Object ---
type HTTPResponse struct {
	StatusCode *Integer
	Content    *Bytes // Raw response body
	Headers    *Dict
	URL        *String
	Encoding   *String // Detected/assumed encoding, e.g., "utf-8"
	Reason     *String // e.g., "OK", "Not Found"
	// UnderlyingGoResponse *http.Response // For advanced use, can be added later
}

func (hr *HTTPResponse) Type() ObjectType { return HTTP_RESPONSE_OBJ }
func (hr *HTTPResponse) Inspect() string {
	statusCode := constants.QuestionMark
	if hr.StatusCode != nil {
		statusCode = hr.StatusCode.Inspect()
	}
	urlStr := constants.EmptyString
	if hr.URL != nil {
		urlStr = hr.URL.Value
	}
	return fmt.Sprintf(constants.HTTP_RESPONSE_INSPECT_FORMAT, statusCode, urlStr)
}
var _ Object = (*HTTPResponse)(nil)
var _ AttributeGetter = (*HTTPResponse)(nil)

func (hr *HTTPResponse) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	switch name {
	case constants.HTTP_RESPONSE_STATUS_CODE_ATTR:
		if hr.StatusCode == nil { return NULL, true }
		return hr.StatusCode, true
	case constants.HTTP_RESPONSE_TEXT_ATTR:
		if hr.Content == nil { return &String{Value: constants.EmptyString}, true }
		// TODO: Use hr.Encoding if set and not UTF-8. For now, assume UTF-8.
		return &String{Value: string(hr.Content.Value)}, true
	case constants.HTTP_RESPONSE_CONTENT_ATTR:
		if hr.Content == nil { return &Bytes{Value: []byte{}}, true }
		return hr.Content, true
	case constants.HTTP_RESPONSE_HEADERS_ATTR:
		if hr.Headers == nil { return &Dict{Pairs: make(map[HashKey]DictPair)}, true }
		return hr.Headers, true
	case constants.HTTP_RESPONSE_URL_ATTR:
		if hr.URL == nil { return NULL, true }
		return hr.URL, true
	case constants.HTTP_RESPONSE_ENCODING_ATTR:
		if hr.Encoding == nil { return NULL, true } // Or default to "utf-8" string?
		return hr.Encoding, true
	case constants.HTTP_RESPONSE_REASON_ATTR:
		if hr.Reason == nil { return NULL, true }
		return hr.Reason, true
	case constants.HTTP_RESPONSE_JSON_METHOD_NAME:
		// Direct definition of the Builtin.Fn
		return &Builtin{
			Name: constants.HTTPResponseJsonMethodName,
			Fn: func(callCtx ExecutionContext, pylearnArgs ...Object) Object {
				// 'hr' is the receiver of GetObjectAttribute, so it's available here.
				// This function is the one the interpreter calls directly.
				// 'pylearnArgs' are the arguments from the Pylearn script call (e.g. () for val1.json()).
	
				if len(pylearnArgs) != 0 { // json() takes no arguments from Pylearn script
					return NewError(constants.TypeError, constants.HTTP_RESPONSE_JSON_ARG_COUNT_ERROR, len(pylearnArgs))
				}
	
				// Now perform the JSON logic using 'hr' as self.
				if hr.Content == nil || len(hr.Content.Value) == 0 {
					return NewError(constants.JSONDecodeError, constants.HTTP_JSON_DECODE_EMPTY_CONTENT)
				}
				var goData interface{}
				err := gojson.Unmarshal(hr.Content.Value, &goData)
				if err != nil {
					return NewError(constants.JSONDecodeError, constants.HTTP_JSON_DECODE_ERROR_FORMAT, err.Error(), string(hr.Content.Value))
				}
				pylearnObj, convertErr := ConvertGoInterfaceToPylearnForHTTP(goData, 0)
				if convertErr != nil {
					return NewError(constants.JSONDecodeError, constants.HTTP_JSON_CONVERSION_ERROR, convertErr)
				}
				return pylearnObj
			},
		}, true
	case constants.HTTP_RESPONSE_RAISE_FOR_STATUS_METHOD_NAME:
		return &Builtin{
			Name: constants.HTTPResponseRaiseForStatusMethodName,
			Fn: func(callCtx ExecutionContext, args ...Object) Object {
				if len(args) != 1 { // self
					return NewError(constants.TypeError, constants.HTTP_RESPONSE_RAISE_FOR_STATUS_ARG_COUNT_ERROR, len(args)-1)
				}
				selfResp, ok := args[0].(*HTTPResponse)
				if !ok { return NewError(constants.TypeError, constants.HTTP_RESPONSE_RAISE_FOR_STATUS_ON_RESPONSE_ERROR) }

				if selfResp.StatusCode == nil { return NULL }
				statusCode := selfResp.StatusCode.Value
				urlVal := constants.EmptyString
				if selfResp.URL != nil {urlVal = selfResp.URL.Value}
				reasonVal := constants.EmptyString
				if selfResp.Reason != nil {reasonVal = selfResp.Reason.Value}


				if statusCode >= 400 && statusCode < 500 {
					return NewError(constants.HTTPClientError, constants.HTTPClientErrorFormat, statusCode, reasonVal, urlVal)
				}
				if statusCode >= 500 && statusCode < 600 {
					return NewError(constants.HTTPServerError, constants.HTTPServerErrorFormat, statusCode, reasonVal, urlVal)
				}
				return NULL
			},
		}, true
	}
	return nil, false
}

// --- Pylearn HTTPRequest Object (Mainly for type consistency, used by server) ---
type HTTPRequest struct {
	Method  *String
	URL     *String // Replaces 'Path' for client-side consistency
	Headers *Dict
	Body    Object // Can be String or Bytes when server receives it
	// Client-side construction might use different fields like 'data' or 'json'
}

func (hr *HTTPRequest) Type() ObjectType { return HTTP_REQUEST_OBJ }
func (hr *HTTPRequest) Inspect() string {
	method, urlStr := constants.QuestionMark, constants.Slash
	if hr.Method != nil { method = hr.Method.Value }
	if hr.URL != nil { urlStr = hr.URL.Value }
	return fmt.Sprintf(constants.HTTP_REQUEST_INSPECT_FORMAT, method, urlStr)
}
var _ Object = (*HTTPRequest)(nil)
var _ AttributeGetter = (*HTTPRequest)(nil)

func (hr *HTTPRequest) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	switch name {
	case constants.HTTP_REQUEST_METHOD_ATTR:
		if hr.Method == nil { return NULL, true }
		return hr.Method, true
	case constants.HTTP_REQUEST_URL_ATTR:
		if hr.URL == nil { return NULL, true }
		return hr.URL, true
	case constants.HTTP_REQUEST_HEADERS_ATTR:
		if hr.Headers == nil { return &Dict{Pairs: make(map[HashKey]DictPair)}, true }
		return hr.Headers, true
	case constants.HTTP_REQUEST_BODY_ATTR:
		if hr.Body == nil { return NULL, true }
		return hr.Body, true
	}
	return nil, false
}

// --- Helper Functions (used by client functions and response.json()) ---

// convertGoInterfaceToPylearnForHTTP is a simplified converter.
// For a full implementation, integrate with your stdlib/pyjson converters.
func ConvertGoInterfaceToPylearnForHTTP(data interface{}, depth int) (Object, error) {
	if depth > 64 {
		return nil, fmt.Errorf(constants.HTTP_JSON_MAX_RECURSION_DEPTH)
	}
	switch v := data.(type) {
	case nil: return NULL, nil
	case bool: return NativeBoolToBooleanObject(v), nil
	case float64:
		if v == float64(int64(v)) { return &Integer{Value: int64(v)}, nil }
		return &Float{Value: v}, nil
	case string: return &String{Value: v}, nil
	case []interface{}:
		elements := make([]Object, len(v))
		for i, item := range v {
			converted, err := ConvertGoInterfaceToPylearnForHTTP(item, depth+1)
			if err != nil { return nil, err }
			elements[i] = converted
		}
		return &List{Elements: elements}, nil
	case map[string]interface{}:
		pairs := make(map[HashKey]DictPair, len(v))
		for key, val := range v {
			pyKey := &String{Value: key}
			pyVal, err := ConvertGoInterfaceToPylearnForHTTP(val, depth+1)
			if err != nil { return nil, err }
			hashKey, _ := pyKey.HashKey()
			pairs[hashKey] = DictPair{Key: pyKey, Value: pyVal}
		}
		return &Dict{Pairs: pairs}, nil
	default:
		return nil, fmt.Errorf(constants.HTTP_JSON_UNSUPPORTED_TYPE, v)
	}
}

// convertPylearnDictToGoHeader converts Pylearn Dict to http.Header for requests.
func ConvertPylearnDictToGoHeader(headersObj Object) (http.Header, error) {
	if headersObj == nil || headersObj == NULL {
		return make(http.Header), nil
	}
	pyDict, ok := headersObj.(*Dict)
	if !ok {
		return nil, fmt.Errorf(constants.HTTP_REQUEST_HEADERS_TYPE_ERROR, headersObj.Type())
	}
	goHeader := make(http.Header)
	for _, pair := range pyDict.Pairs {
		keyStr, keyOk := pair.Key.(*String)
		valStr, valOk := pair.Value.(*String)
		if !keyOk || !valOk {
			return nil, fmt.Errorf(constants.HTTP_REQUEST_HEADER_KEY_VALUE_ERROR, pair.Key.Type(), pair.Value.Type())
		}
		goHeader.Add(keyStr.Value, valStr.Value)
	}
	return goHeader, nil
}

// convertPylearnDictToURLValues converts Pylearn Dict to "key=value&key2=value2" query string.
func ConvertPylearnDictToURLValues(paramsObj Object) (string, error) {
	if paramsObj == nil || paramsObj == NULL {
		return constants.EmptyString, nil
	}
	pyDict, ok := paramsObj.(*Dict)
	if !ok {
		return constants.EmptyString, fmt.Errorf(constants.HTTP_REQUEST_PARAMS_TYPE_ERROR, paramsObj.Type())
	}

	queryValues := url.Values{}
	for _, pair := range pyDict.Pairs {
		keyStr, keyOk := pair.Key.(*String)
		if !keyOk {
			return constants.EmptyString, fmt.Errorf(constants.HTTP_REQUEST_PARAMS_KEY_ERROR, pair.Key.Type())
		}

		var valToAdd string
		switch v := pair.Value.(type) {
		case *String:
			valToAdd = v.Value
		case *Integer:
			valToAdd = fmt.Sprintf(constants.FormatInt, v.Value)
		case *Float:
			valToAdd = fmt.Sprintf(constants.FormatFloatNoTrailingZeros, v.Value)
		case *Boolean:
			if v.Value { valToAdd = constants.HTTP_TrueString } else { valToAdd = constants.HTTP_FalseString }
		case *List: // Python requests can send multiple values for the same key
			for _, listItem := range v.Elements {
				switch lv := listItem.(type) {
				case *String: queryValues.Add(keyStr.Value, lv.Value)
				case *Integer: queryValues.Add(keyStr.Value, fmt.Sprintf(constants.FormatInt, lv.Value))
				// Add other simple types if needed for list items
				default:
					// Or convert to string using Inspect()
					queryValues.Add(keyStr.Value, lv.Inspect())
				}
			}
			continue // Skip single value add for this key
		default:
			// For other types, maybe convert to string via Inspect or error
			// Python requests typically stringifies simple types.
			valToAdd = pair.Value.Inspect()
		}
		queryValues.Add(keyStr.Value, valToAdd)
	}
	return queryValues.Encode(), nil // Returns URL-encoded string "key=value&key2=value2"
}


// populateResponseObject is a helper to create the Pylearn HTTPResponse from a Go *http.Response
// It's crucial and will be used by all request functions.
func PopulateResponseObject(goResp *http.Response, originalURL string) (*HTTPResponse, Object) {
	if goResp == nil {
		return nil, NewError(constants.InternalError, constants.HTTP_INTERNAL_NIL_RESPONSE)
	}
	defer goResp.Body.Close() // Ensure body is closed after reading

	bodyBytes, err := ioutil.ReadAll(goResp.Body)
	if err != nil {
		return nil, NewError(constants.RequestError, constants.HTTP_REQUEST_BODY_READ_ERROR, originalURL, err)
	}

	pyHeaders := make(map[HashKey]DictPair)
	for k, v := range goResp.Header {
		keyObj := &String{Value: strings.ToLower(k)}
		valObj := &String{Value: strings.Join(v, constants.CommaWithSpace)}
		hashKey, _ := keyObj.HashKey()
		pyHeaders[hashKey] = DictPair{Key: keyObj, Value: valObj}
	}

	contentType := goResp.Header.Get(constants.HTTP_ContentTypeHeader)
	encoding := constants.HTTP_UTF8Encoding // Default
	if strings.Contains(strings.ToLower(contentType), constants.HTTP_CharsetEquals) {
		parts := strings.Split(contentType, constants.HTTP_CharsetEquals)
		if len(parts) > 1 {
			encoding = strings.ToLower(strings.TrimSpace(parts[1]))
			if semiColonIdx := strings.Index(encoding, constants.Semicolon); semiColonIdx != -1 {
				encoding = encoding[:semiColonIdx]
			}
		}
	}
	
	reason := strings.TrimSpace(strings.TrimPrefix(goResp.Status, fmt.Sprintf(constants.FormatInt, goResp.StatusCode)))


	responseObj := &HTTPResponse{
		StatusCode: &Integer{Value: int64(goResp.StatusCode)},
		Content:    &Bytes{Value: bodyBytes},
		Headers:    &Dict{Pairs: pyHeaders},
		URL:        &String{Value: originalURL},
		Encoding:   &String{Value: encoding},
		Reason:     &String{Value: reason},
		// UnderlyingGoResponse: goResp, // Optional: if needed for advanced access
	}
	return responseObj, nil
}