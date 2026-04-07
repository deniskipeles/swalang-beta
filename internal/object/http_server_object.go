// internal/object/http_server_object.go
package object

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants" // Import constants package
)

const (
	SERVER_RESPONSE_OBJ ObjectType = constants.SERVER_RESPONSE_OBJ_TYPE
)

// ServerResponse allows Pylearn handlers to specify status, headers, and body.
type ServerResponse struct {
	Body        Object // String, Bytes, or object that can be JSON-serialized
	StatusCode  *Integer
	Headers     *Dict   // Pylearn Dict of headers
	ContentType *String // Explicit content type
}

func (sr *ServerResponse) Type() ObjectType { return SERVER_RESPONSE_OBJ }
func (sr *ServerResponse) Inspect() string {
	status := constants.SERVER_RESPONSE_DEFAULT_STATUS
	if sr.StatusCode != nil {
		status = sr.StatusCode.Inspect()
	}
	return fmt.Sprintf(constants.SERVER_RESPONSE_INSPECT_FORMAT, status)
}

var _ Object = (*ServerResponse)(nil)
