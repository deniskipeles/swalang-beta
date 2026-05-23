//go:build windows
package platform

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// --- THIS IS THE NEW, EXPORTED HELPER FUNCTION ---
// AddDllDirectory adds a directory to the process's DLL search path.
// This is crucial for ensuring that dependencies of a loaded library can be found.
func AddDllDirectory(path string) error {
	kernel32, err := syscall.LoadLibrary(constants.PLATFORM_WINDOWS_KERNEL32_DLL)
	if err != nil {
		return fmt.Errorf(constants.PLATFORM_WINDOWS_LOAD_KERNEL32_ERROR, err)
	}
	// No defer for FreeLibrary on kernel32, it should stay loaded.

	proc, err := syscall.GetProcAddress(kernel32, constants.PLATFORM_WINDOWS_SET_DLL_DIR_PROC)
	if err != nil {
		// Fallback for older Windows versions or different APIs if needed,
		// but SetDllDirectory is standard.
		return fmt.Errorf(constants.PLATFORM_WINDOWS_FIND_PROC_ERROR, err)
	}

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf(constants.PLATFORM_WINDOWS_UTF16_CONVERSION_ERROR, path, err)
	}

	ret, _, callErr := syscall.Syscall(uintptr(proc), 1, uintptr(unsafe.Pointer(pathPtr)), 0, 0)
	if ret == 0 {
		// A non-zero return value indicates success.
		return fmt.Errorf(constants.PLATFORM_WINDOWS_SET_DLL_DIR_FAIL_ERROR, path, callErr)
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
		return 0, fmt.Errorf(constants.PLATFORM_WINDOWS_LOAD_LIBRARY_ERROR, name, err)
	}
	return LibraryHandle(handle), nil
}

// FreeLibrary uses the Win32 FreeLibrary function.
func (m *windowsManager) FreeLibrary(handle LibraryHandle) error {
	err := syscall.FreeLibrary(syscall.Handle(handle))
	if err != nil {
		return fmt.Errorf(constants.PLATFORM_WINDOWS_FREE_LIBRARY_ERROR, err)
	}
	return nil
}

// GetProcAddress uses the Win32 GetProcAddress function.
func (m *windowsManager) GetProcAddress(handle LibraryHandle, procName string) (FuncPtr, error) {
	ptr, err := syscall.GetProcAddress(syscall.Handle(handle), procName)
	if err != nil {
		return 0, fmt.Errorf(constants.PLATFORM_WINDOWS_GET_PROC_ADDRESS_ERROR, procName, err)
	}
	return FuncPtr(ptr), nil
}

// LibraryExtension returns the correct extension for Windows.
func (m *windowsManager) LibraryExtension() string {
	return constants.PLATFORM_WINDOWS_LIB_EXTENSION
}