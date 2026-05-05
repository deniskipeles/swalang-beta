//go:build linux || darwin
// pylearn/internal/stdlib/ffi3/platform_unix.go

package ffi3

import (
	"fmt"
	"os"
	"path/filepath"
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

func registerPlatformSpecifics(env *object.Environment) {}

func findProjectRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", false
		}
		dir = parentDir
	}
}

func discoverDynamicPaths(baseDir string) []string {
	var paths []string
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return paths
	}
	paths = append(paths, baseDir)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return paths
	}

	for _, entry := range entries {
		if entry.IsDir() {
			lvl1 := filepath.Join(baseDir, entry.Name())
			paths = append(paths, lvl1)

			subEntries, err := os.ReadDir(lvl1)
			if err == nil {
				for _, subEntry := range subEntries {
					if subEntry.IsDir() {
						lvl2 := filepath.Join(lvl1, subEntry.Name())
						paths = append(paths, lvl2)
					}
				}
			}
		}
	}
	return paths
}

func findLibrary(name string) string {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			return name
		}
	}

	var allSearchPaths[]string

	// 1. Production Layout: Resolve relative to the Swalang executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath) // e.g., root-folder/bin
		rootDir := filepath.Dir(exeDir) // e.g., root-folder
		
		allSearchPaths = append(allSearchPaths, exeDir)
		allSearchPaths = append(allSearchPaths, discoverDynamicPaths(filepath.Join(rootDir, "lib"))...)
		allSearchPaths = append(allSearchPaths, discoverDynamicPaths(filepath.Join(rootDir, "bin"))...)
	}

	// 2. Development Layout: Resolve via go.mod
	if projectRoot, found := findProjectRoot(); found {
		allSearchPaths = append(allSearchPaths, discoverDynamicPaths(filepath.Join(projectRoot, "bin"))...)
		allSearchPaths = append(allSearchPaths, discoverDynamicPaths(filepath.Join(projectRoot, "lib"))...)
	}

	for _, searchDir := range allSearchPaths {
		possibleNames :=[]string{
			name,
			"lib" + name + libManager.LibraryExtension(),
			name + libManager.LibraryExtension(),
		}
		for _, libName := range possibleNames {
			fullPath := filepath.Join(searchDir, libName)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return fullPath
			}
		}
	}

	systemPaths :=[]string{"/lib", "/usr/lib", "/usr/local/lib", "/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu"}
	possibleNames :=[]string{name, "lib" + name + libManager.LibraryExtension()}

	for _, sysPath := range systemPaths {
		for _, libName := range possibleNames {
			fullPath := filepath.Join(sysPath, libName)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return fullPath
			}
		}
	}

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
		originalErr := err
		if strings.Contains(err.Error(), "invalid ELF header") {
			handle, err = libManager.LoadLibrary(name)
			if err != nil {
				return nil, &FFIError{Code: ErrLibNotFound, Message: fmt.Sprintf("could not load library '%s': %v (ELF error: %v)", name, err, originalErr)}
			}
		} else {
			if libPath != name {
				handle, err = libManager.LoadLibrary(name)
			}
			if err != nil {
				return nil, &FFIError{Code: ErrLibNotFound, Message: fmt.Sprintf("could not load library '%s': %v \n(Original path error: %v)", name, err, originalErr)}
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