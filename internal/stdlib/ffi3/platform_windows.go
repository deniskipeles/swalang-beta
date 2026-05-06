//go:build windows
// pylearn/internal/stdlib/ffi3/platform_windows.go

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

func pyGetLastError(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError("TypeError", "get_last_error() takes no arguments")
	}
	ret, _, _ := procGetLastError.Call()
	return &object.Integer{Value: int64(ret)}
}

func registerPlatformSpecifics(env *object.Environment) {
	env.Set("get_last_error", &object.Builtin{Name: "_ffi.get_last_error", Fn: pyGetLastError})
}

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

func addDllSearchPath(dir string) {
	addedDllDirsMu.Lock()
	defer addedDllDirsMu.Unlock()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	if !addedDllDirs[absDir] {
		if err := platform.AddDllDirectory(absDir); err == nil {
			addedDllDirs[absDir] = true
		}
	}
}

func findLibrary(name string) string {
	if strings.ContainsRune(name, os.PathSeparator) || strings.ContainsRune(name, '/') {
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			addDllSearchPath(filepath.Dir(name))
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

	for _, path := range allSearchPaths {
		addDllSearchPath(path)
	}

	for _, searchDir := range allSearchPaths {
		possibleNames :=[]string{
			name,
			name + libManager.LibraryExtension(),
			"lib" + name + libManager.LibraryExtension(),
		}
		for _, libName := range possibleNames {
			fullPath := filepath.Join(searchDir, libName)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return fullPath
			}
		}
	}

	systemPaths :=[]string{
		filepath.Join(os.Getenv("WINDIR"), "System32"),
		filepath.Join(os.Getenv("WINDIR"), "SysWOW64"),
	}
	for _, sysPath := range systemPaths {
		fullPath := filepath.Join(sysPath, name+libManager.LibraryExtension())
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			return fullPath
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
		if libPath != name {
			handle, err = libManager.LoadLibrary(name)
		}
		if err != nil {
			return nil, &FFIError{Code: ErrLibNotFound, Message: fmt.Sprintf("could not load library '%s': %v \n(Original path error: %v)", name, err, originalErr)}
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