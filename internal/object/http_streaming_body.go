// pylearn/internal/object/http_streaming_body.go
package object

import (
	"io"
)

const (
	HTTP_STREAMING_BODY_OBJ ObjectType = "StreamingBody"
	// A reasonable chunk size for reading from the request body.
	requestBodyChunkSize = 4096
)

// HTTPStreamingBody is an iterable Pylearn object that wraps an io.Reader (like http.Request.Body).
type HTTPStreamingBody struct {
	source   io.ReadCloser
	isClosed bool
}

// NewHTTPStreamingBody is the constructor for our streaming body object.
func NewHTTPStreamingBody(source io.ReadCloser) *HTTPStreamingBody {
	return &HTTPStreamingBody{source: source}
}

func (s *HTTPStreamingBody) Type() ObjectType { return HTTP_STREAMING_BODY_OBJ }
func (s *HTTPStreamingBody) Inspect() string  { return "<http.StreamingBody>" }

// Next implements the Iterator interface.
func (s *HTTPStreamingBody) Next() (Object, bool) {
	if s.isClosed {
		return nil, true // Stop iteration if already closed or finished
	}

	buffer := make([]byte, requestBodyChunkSize)
	bytesRead, err := s.source.Read(buffer)

	if bytesRead > 0 {
		// Return the chunk that was read as a Pylearn Bytes object.
		// We must return a slice of the buffer that contains actual data.
		return &Bytes{Value: buffer[:bytesRead]}, false
	}

	if err == io.EOF {
		// End of the stream. Close the source and stop iteration.
		s.source.Close()
		s.isClosed = true
		return nil, true
	}

	if err != nil {
		// An unexpected error occurred.
		s.source.Close()
		s.isClosed = true
		// Propagate the error as a Pylearn Error object.
		return NewError("IOError", "Error reading from request stream: %v", err), true
	}

	// This case should not be reached (bytesRead=0 and err=nil), but for safety:
	s.source.Close()
	s.isClosed = true
	return nil, true
}

// Ensure HTTPStreamingBody implements the necessary interfaces.
var _ Object = (*HTTPStreamingBody)(nil)
var _ Iterator = (*HTTPStreamingBody)(nil)