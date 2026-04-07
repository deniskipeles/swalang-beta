// pylearn/internal/stdlib/ffi3/platform_unix.go
//go:build linux || darwin

package ffi3

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/stdlib/platform"
)

var libManager = platform.GetManager()
var (
	libCache   = make(map[string]*Library)
	libCacheMu sync.RWMutex
)

// On Unix-like systems, this function does nothing.
func registerPlatformSpecifics(env *object.Environment) {
	// No Windows-specific functions to register.
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

// findLibrary attempts to locate a shared library file.
func findLibrary(name string) string {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	var allSearchPaths []string

	// Helper function to get the desired directory name for the current architecture.
	getPlatformArchAlias := func() string {
		switch runtime.GOARCH {
		case "amd64":
			return "x86_64" // Translate Go's 'amd64' to the desired 'x86_64'
		default:
			return runtime.GOARCH // Use the default for all other architectures
		}
	}
	platformArch := getPlatformArchAlias()

	// Build the platform subdirectory path using our aliased architecture name.
	platformSubDir := filepath.Join("bin", fmt.Sprintf("%s-%s", runtime.GOOS, platformArch))

	// Discover dynamic paths relative to the executable (for deployed builds)
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		dynamicBase := filepath.Join(exeDir, platformSubDir)
		allSearchPaths = append(allSearchPaths, discoverDynamicPaths(dynamicBase)...)
	}

	// Discover dynamic paths relative to the project root (for 'go run')
	if projectRoot, found := findProjectRoot(); found {
		dynamicBase := filepath.Join(projectRoot, platformSubDir)
		allSearchPaths = append(allSearchPaths, discoverDynamicPaths(dynamicBase)...)
	}

	// Add static relative paths as a fallback
	staticPaths := []string{"bin", "lib"}
	if projectRoot, found := findProjectRoot(); found {
		for _, sp := range staticPaths {
			allSearchPaths = append(allSearchPaths, filepath.Join(projectRoot, sp))
		}
	}

	// Search all collected bundled paths
	for _, searchDir := range allSearchPaths {
		possibleNames := []string{
			name,
			"lib" + name + libManager.LibraryExtension(),
			name + libManager.LibraryExtension(),
		}
		for _, libName := range possibleNames {
			fullPath := filepath.Join(searchDir, libName)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}

	// If not found in bundled paths, search system paths
	systemPaths := []string{"/lib", "/usr/lib", "/usr/local/lib", "/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu"}
	possibleNames := []string{name, "lib" + name + libManager.LibraryExtension()}

	for _, sysPath := range systemPaths {
		for _, libName := range possibleNames {
			fullPath := filepath.Join(sysPath, libName)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}

	// Final fallback to let the OS loader try to find it
	return name
}

// LoadLibrary loads a shared library, handling caching and special loading cases.
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
		// If loading a full path gave an "invalid ELF header", it's likely a linker script (e.g. libc.so).
		// The correct action is to retry by passing the ORIGINAL, simple name to the system loader.
		if strings.Contains(err.Error(), "invalid ELF header") {
			handle, err = libManager.LoadLibrary(name) // Retry with just 'c' or 'm'
			if err != nil {
				return nil, &FFIError{Code: ErrLibNotFound, Message: fmt.Sprintf("could not load library '%s': %v", name, err)}
			}
		} else {
			// Original fallback for other errors (e.g. file not found)
			if libPath != name {
				handle, err = libManager.LoadLibrary(name)
			}
			if err != nil {
				return nil, &FFIError{Code: ErrLibNotFound, Message: fmt.Sprintf("could not load library '%s': %v", name, err)}
			}
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