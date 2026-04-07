// pylearn/internal/stdlib/pyhttpserver/response_helpers.go
package pyhttpserver

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// Helper: httpserver.json_response(data, status_code=200, headers=None)
func pyJsonResponseFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 3 {
		return object.NewError(constants.TypeError, "json_response() takes 1 to 3 arguments (data, status_code=200, headers=None)")
	}
	data := args[0]
	switch data.(type) {
	case *object.Dict, *object.List, *object.String:
	default:
		return object.NewError(constants.TypeError, "json_response() data argument must be a Dict, List, or JSON String, not %s", data.Type())
	}

	statusCode := int64(200)
	var headers *object.Dict = nil

	if len(args) >= 2 && args[1] != object.NULL {
		statusObj, ok := args[1].(*object.Integer)
		if !ok {
			return object.NewError(constants.TypeError, "json_response() status_code argument must be an Integer or None")
		}
		statusCode = statusObj.Value
	}
	if len(args) == 3 && args[2] != object.NULL {
		headersObj, ok := args[2].(*object.Dict)
		if !ok {
			return object.NewError(constants.TypeError, "json_response() headers argument must be a Dict or None")
		}
		headers = headersObj
	}

	return &object.ServerResponse{
		Body:        data,
		StatusCode:  &object.Integer{Value: statusCode},
		Headers:     headers,
		ContentType: &object.String{Value: "application/json; charset=utf-8"},
	}
}
var JsonResponse = &object.Builtin{Name: "httpserver.json_response", Fn: pyJsonResponseFn}

// Helper: httpserver.text_response(text, status_code=200, content_type="text/plain", headers=None)
func pyTextResponseFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 4 {
		return object.NewError(constants.TypeError, "text_response() takes 1 to 4 arguments")
	}
	textObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "text_response() text argument must be a String")
	}

	statusCode := int64(200)
	contentType := "text/plain; charset=utf-8"
	var headers *object.Dict = nil

	if len(args) >= 2 && args[1] != object.NULL {
		statusObj, okS := args[1].(*object.Integer)
		if !okS {
			return object.NewError(constants.TypeError, "text_response() status_code argument must be an Integer")
		}
		statusCode = statusObj.Value
	}
	if len(args) >= 3 && args[2] != object.NULL {
		ctObj, okCt := args[2].(*object.String)
		if !okCt {
			return object.NewError(constants.TypeError, "text_response() content_type argument must be a String")
		}
		contentType = ctObj.Value
	}
	if len(args) == 4 && args[3] != object.NULL {
		hObj, okH := args[3].(*object.Dict)
		if !okH {
			return object.NewError(constants.TypeError, "text_response() headers argument must be a Dict")
		}
		headers = hObj
	}

	return &object.ServerResponse{
		Body:        textObj,
		StatusCode:  &object.Integer{Value: statusCode},
		Headers:     headers,
		ContentType: &object.String{Value: contentType},
	}
}
var TextResponse = &object.Builtin{Name: "httpserver.text_response", Fn: pyTextResponseFn}