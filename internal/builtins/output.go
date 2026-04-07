package builtins

import (
	"io"
	"os"
)

var outputWriter io.Writer = os.Stdout

// SetOutput sets the output destination (e.g., strings.Builder or os.Stdout)
func SetOutput(w io.Writer) {
	if w != nil {
		outputWriter = w
	}
}

// GetOutput returns the currently set output writer
func GetOutput() io.Writer {
	return outputWriter
}
