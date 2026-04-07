//go:build linux || darwin || windows
// pylearn/internal/stdlib/ffi/platform/platform.go
package platform

// LibraryHandle is an opaque handle to a loaded shared library.
// It's a uintptr to hold the memory address or handle returned by the OS.
type LibraryHandle uintptr

// FuncPtr is an opaque handle to a function pointer within a library.
type FuncPtr uintptr

// LibManager is the interface for platform-specific library operations.
// This abstraction allows the main FFI code to be OS-agnostic.
type LibManager interface {
	// LoadLibrary loads a shared library from the given path or name
	// and returns a handle to it.
	LoadLibrary(name string) (LibraryHandle, error)

	// FreeLibrary unloads a shared library, freeing its resources.
	FreeLibrary(handle LibraryHandle) error

	// GetProcAddress retrieves the memory address of an exported function
	// from within a loaded library.
	GetProcAddress(handle LibraryHandle, procName string) (FuncPtr, error)

	// LibraryExtension returns the standard shared library file extension
	// for the current operating system (e.g., ".so", ".dll").
	LibraryExtension() string
}

// GetManager returns the appropriate manager for the current operating system.
// The actual implementation is provided by the OS-specific files
// (platform_linux.go, platform_windows.go, etc.) via build tags.
func GetManager() LibManager {
	return getManager()
}