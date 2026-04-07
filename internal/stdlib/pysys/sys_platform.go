package pysys

import (
	"runtime" // Go package to get OS info

	"github.com/deniskipeles/pylearn/internal/object"
)

// Platform holds the OS identifier string (e.g., "linux", "darwin", "windows")
var Platform *object.String

func init() {
	// Map Go's GOOS to common Python sys.platform values if desired,
	// or just use GOOS directly. Let's use GOOS for simplicity.
	platformValue := runtime.GOOS
	// Example mapping:
	// switch runtime.GOOS {
	// case "darwin": platformValue = "darwin"
	// case "linux": platformValue = "linux"
	// case "windows": platformValue = "win32" // Common Python value
	// default: platformValue = runtime.GOOS // Fallback
	// }
	Platform = &object.String{Value: platformValue}
}
