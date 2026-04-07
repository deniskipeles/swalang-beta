// internal/stdlib/net/http_streaming_body_object.go
package net

import (
	"io"

	"github.com/deniskipeles/pylearn/internal/object"
)

const (
	HTTP_STREAMING_BODY_OBJ object.ObjectType = "net.StreamingBody"
)

var StreamingBodyClass *object.Class

// StreamingBody is a special Pylearn object that wraps a Go io.ReadCloser.
// When returned from a handler, the server knows to stream its contents.
type StreamingBody struct {
	Source io.ReadCloser
}

func (sb *StreamingBody) Type() object.ObjectType { return HTTP_STREAMING_BODY_OBJ }
func (sb *StreamingBody) Inspect() string         { return "<net.StreamingBody>" }

func init() {
	StreamingBodyClass = object.CreateExceptionClass("StreamingBody", object.ObjectClass)
}