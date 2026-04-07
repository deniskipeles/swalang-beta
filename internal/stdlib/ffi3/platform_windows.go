// pylearn/internal/stdlib/ffi3/platform_windows.go
//go:build windows

package ffi3

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/stdlib/platform"
)

var libManager = platform.GetManager()
var (
	libCache       = make(map[string]*Library)
	libCacheMu     sync.RWMutex
	addedDllDirs   = make(map[string]bool)
	addedDllDirsMu sync.Mutex
)
var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetLastError = kernel32.NewProc("GetLastError")
)

// pyGetLastError is the implementation for the Windows-specific builtin.
func pyGetLastError(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError("TypeError", "get_last_error() takes no arguments")
	}
	ret, _, _ := procGetLastError.Call()
	return &object.Integer{Value: int64(ret)}
}

// registerPlatformSpecifics is called by the main init() function.
func registerPlatformSpecifics(env *object.Environment) {
	env.Set("get_last_error", &object.Builtin{Name: "_ffi.get_last_error", Fn: pyGetLastError})
}

// findProjectRoot is still useful for development with `go run`.
func findProjectRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, true
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", false
		}
		dir = parentDir
	}
}

// discoverDynamicPaths scans a base directory for library subdirectories and returns their potential library paths.
func discoverDynamicPaths(baseDir string) []string {
	var discoveredPaths []string
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return discoveredPaths
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return discoveredPaths
	}

	for _, entry := range entries {
		if entry.IsDir() {
			libraryBundleDir := filepath.Join(baseDir, entry.Name())
			potentialSubDirs := []string{"lib", "bin"}
			for _, subDir := range potentialSubDirs {
				fullPath := filepath.Join(libraryBundleDir, subDir)
				if _, err := os.Stat(fullPath); err == nil {
					discoveredPaths = append(discoveredPaths, fullPath)
				}
			}
			discoveredPaths = append(discoveredPaths, libraryBundleDir)
		}
	}
	return discoveredPaths
}

// addDllSearchPath adds a directory to the DLL search path if not already added.
func addDllSearchPath(dir string) {
	addedDllDirsMu.Lock()
	defer addedDllDirsMu.Unlock()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return // Silently fail on absolute path error
	}
	if !addedDllDirs[absDir] {
		if err := platform.AddDllDirectory(absDir); err == nil {
			addedDllDirs[absDir] = true
		}
	}
}

// findLibrary is the Windows-specific library search logic.
func findLibrary(name string) string {
	if strings.ContainsRune(name, os.PathSeparator) || strings.ContainsRune(name, '/') {
		if _, err := os.Stat(name); err == nil {
			addDllSearchPath(filepath.Dir(name))
			return name
		}
	}

	// --- Proactive Dependency Path Loading Logic ---
	potentialRoots := make(map[string]bool)

	// Identify potential root directories where a 'bin' folder might exist.
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		potentialRoots[exeDir] = true                 // Root could be the exe's own directory
		potentialRoots[filepath.Dir(exeDir)] = true // Root could be the parent of the exe's directory
	}
	if projectRoot, found := findProjectRoot(); found {
		potentialRoots[projectRoot] = true // Root could be the project root for 'go run'
	}

	// From each potential root, look for a 'bin' directory and add all sub-paths to the DLL search path.
	allPossibleDependencyPaths := []string{}
	for root := range potentialRoots {
		moduleBasePath := filepath.Join(root, "bin")
		discoveredPaths := discoverDynamicPaths(moduleBasePath)
		allPossibleDependencyPaths = append(allPossibleDependencyPaths, discoveredPaths...)
	}
	for _, path := range allPossibleDependencyPaths {
		addDllSearchPath(path)
	}
	// --- End Proactive Logic ---

	// Now that all dependency paths are registered, search for the target library.
	for _, searchDir := range allPossibleDependencyPaths {
		possibleNames := []string{
			name,
			name + libManager.LibraryExtension(),
			"lib" + name + libManager.LibraryExtension(),
		}
		for _, libName := range possibleNames {
			fullPath := filepath.Join(searchDir, libName)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}

	// If not found in bundled paths, search system paths.
	systemPaths := []string{
		filepath.Join(os.Getenv("WINDIR"), "System32"),
		filepath.Join(os.Getenv("WINDIR"), "SysWOW64"),
	}
	for _, sysPath := range systemPaths {
		fullPath := filepath.Join(sysPath, name+libManager.LibraryExtension())
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// Final fallback to let the OS loader find it.
	return name
}

func LoadLibrary(name string) (*Library, error) {
	libCacheMu.RLock()
	if cachedLib, found := libCache[name]; found {
		libCacheMu.RUnlock()
		return cachedLib, nil
	}
	libCacheMu.RUnlock()

	libPath := findLibrary(name)
	
	handle, err := libManager.LoadLibrary(libPath)
	if err != nil {
		if libPath != name {
			handle, err = libManager.LoadLibrary(name)
		}
		if err != nil {
			return nil, &FFIError{Code: ErrLibNotFound, Message: fmt.Sprintf("could not load library '%s': %v", name, err)}
		}
	}

	lib := &Library{
		Name: name, Path: libPath, handle: handle, funcs: make(map[string]*Function),
	}
	libCacheMu.Lock()
	libCache[name] = lib
	libCacheMu.Unlock()
	return lib, nil
}