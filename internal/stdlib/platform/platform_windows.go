// pylearn/internal/stdlib/ffi/platform/platform_windows.go
//go:build windows

package platform

import (
	"fmt"
	"syscall"
	"unsafe"
)

// --- THIS IS THE NEW, EXPORTED HELPER FUNCTION ---
// AddDllDirectory adds a directory to the process's DLL search path.
// This is crucial for ensuring that dependencies of a loaded library can be found.
func AddDllDirectory(path string) error {
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return fmt.Errorf("could not load kernel32.dll: %w", err)
	}
	// No defer for FreeLibrary on kernel32, it should stay loaded.

	proc, err := syscall.GetProcAddress(kernel32, "SetDllDirectoryW")
	if err != nil {
		// Fallback for older Windows versions or different APIs if needed,
		// but SetDllDirectory is standard.
		return fmt.Errorf("could not find SetDllDirectoryW in kernel32.dll: %w", err)
	}

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("could not convert path '%s' to UTF16: %w", path, err)
	}

	ret, _, callErr := syscall.Syscall(uintptr(proc), 1, uintptr(unsafe.Pointer(pathPtr)), 0, 0)
	if ret == 0 {
		// A non-zero return value indicates success.
		return fmt.Errorf("call to SetDllDirectoryW for path '%s' failed: %w", path, callErr)
	}

	return nil
}

// --- The rest of the file is the existing implementation ---

// windowsManager implements the LibManager interface for Windows.
type windowsManager struct{}

// getManager is the platform-specific implementation for GetManager().
func getManager() LibManager {
	return &windowsManager{}
}

// LoadLibrary uses the Win32 LoadLibrary function.
func (m *windowsManager) LoadLibrary(name string) (LibraryHandle, error) {
	handle, err := syscall.LoadLibrary(name)
	if err != nil {
		return 0, fmt.Errorf("LoadLibrary failed for '%s': %w", name, err)
	}
	return LibraryHandle(handle), nil
}

// FreeLibrary uses the Win32 FreeLibrary function.
func (m *windowsManager) FreeLibrary(handle LibraryHandle) error {
	err := syscall.FreeLibrary(syscall.Handle(handle))
	if err != nil {
		return fmt.Errorf("FreeLibrary failed: %w", err)
	}
	return nil
}

// GetProcAddress uses the Win32 GetProcAddress function.
func (m *windowsManager) GetProcAddress(handle LibraryHandle, procName string) (FuncPtr, error) {
	ptr, err := syscall.GetProcAddress(syscall.Handle(handle), procName)
	if err != nil {
		return 0, fmt.Errorf("GetProcAddress failed for '%s': %w", procName, err)
	}
	return FuncPtr(ptr), nil
}

// LibraryExtension returns the correct extension for Windows.
func (m *windowsManager) LibraryExtension() string {
	return ".dll"
}

