//go:build linux || darwin
// pylearn/internal/stdlib/ffi/platform/platform_unix.go

package platform

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// unixManager implements the LibManager interface for Linux and macOS.
type unixManager struct{}

// getManager is the platform-specific implementation for GetManager().
func getManager() LibManager {
	return &unixManager{}
}

// LoadLibrary uses dlopen to load a shared library.
func (m *unixManager) LoadLibrary(name string) (LibraryHandle, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	// RTLD_LAZY: Resolve symbols as code that references them is executed.
	// RTLD_GLOBAL: Make symbols from this library available for other libs,
	// which can be important for dependencies.
	handle := C.dlopen(cname, C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == nil {
		// dlerror() returns a human-readable error string.
		return 0, fmt.Errorf("dlopen failed: %s", C.GoString(C.dlerror()))
	}
	return LibraryHandle(handle), nil
}

// FreeLibrary uses dlclose to unload a library.
func (m *unixManager) FreeLibrary(handle LibraryHandle) error {
	res := C.dlclose(unsafe.Pointer(handle))
	if res != 0 {
		return fmt.Errorf("dlclose failed: %s", C.GoString(C.dlerror()))
	}
	return nil
}

// GetProcAddress uses dlsym to find a function's address.
func (m *unixManager) GetProcAddress(handle LibraryHandle, procName string) (FuncPtr, error) {
	cprocName := C.CString(procName)
	defer C.free(unsafe.Pointer(cprocName))

	// Clear any old error conditions before calling dlsym.
	C.dlerror()
	ptr := C.dlsym(unsafe.Pointer(handle), cprocName)
	
	// Check for errors after the call.
	errStr := C.dlerror()
	if errStr != nil {
		return 0, fmt.Errorf("dlsym failed for '%s': %s", procName, C.GoString(errStr))
	}
	return FuncPtr(ptr), nil
}

// LibraryExtension returns the correct extension for Unix-like systems.
func (m *unixManager) LibraryExtension() string {
	// macOS uses .dylib, but .so is often used and symlinked.
	// We'll default to .so for broader compatibility. A more advanced
	// implementation could check runtime.GOOS.
	return ".so"
}