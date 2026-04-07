// pylearn/internal/interpreter/modules.go
// pylearn/internal/interpreter/modules.go

package interpreter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
)

// --- NEW HELPER: findProjectRoot ---
// findProjectRoot searches upwards from the current directory for a `go.mod` file.
// This gives us a reliable anchor for the project's root, even when using `go run`.
func findProjectRoot() (string, bool) {
	dir, err := os.Getwd() // Start from the current working directory
	if err != nil {
		return "", false
	}
	// Loop upwards until we hit the filesystem root
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found it!
			return dir, true
		}
		// Move up one directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached the root of the filesystem
			return "", false
		}
		dir = parentDir
	}
}

// --- THE CORRECTED getPluginSearchPaths ---
// getPluginSearchPaths defines the locations where we'll look for .so files.
func getPluginSearchPaths() []string {
	paths := []string{}

	// 1. From an environment variable (highest priority)
	if p := os.Getenv(constants.PluginPathEnvironmentVariable); p != "" {
		// This allows users to specify multiple paths separated by the OS list separator (e.g., ':')
		paths = append(paths, filepath.SplitList(p)...)
	}

	// 2. A 'plugins' directory in the project root (found via go.mod)
	// This makes development and running from source much more reliable.
	if projectRoot, found := findProjectRoot(); found {
		paths = append(paths, filepath.Join(projectRoot, constants.PluginsDirectory))
	}

	// 3. A user-specific directory (e.g., ~/.pylearn/plugins)
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, constants.DOT_OurLanguageDirectory, constants.PluginsDirectory))
	}

	// 4. A system-wide directory (e.g., for installed packages)
	paths = append(paths, constants.USR_SLASH_LOCAL_SLASH_LIB_SLASH_OurLanguageDirectory_SLASH_PLUGINS)

	// 5. A directory relative to the final compiled executable
	// This is important for when the interpreter is built and distributed.
	if exePath, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exePath), constants.PluginsDirectory))
	}

	// For debugging, you can print the final search paths
	// fmt.Println("DEBUG: Plugin Search Paths:", paths)

	return paths
}

// tryLoadPluginForModule attempts to find and load a Go plugin (.so) that
// might provide the requested module.
func tryLoadPluginForModule(moduleName string) {
	pluginFileName := moduleName + ".so"
	searchPaths := getPluginSearchPaths()
	// fmt.Print(searchPaths)

	for _, dir := range searchPaths {
		pluginPath := filepath.Join(dir, pluginFileName)

		// Check if the file exists before trying to open it.
		if _, err := os.Stat(pluginPath); err == nil {
			// Plugin file exists, attempt to load it.
			// The plugin's init() function will call object.RegisterNativeModule.
			_, err := plugin.Open(pluginPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, constants.WarningFoundPluginFor_STRINGFORMATER_At_STRINGFORMATER_ButFailedToLoad_VERBFORMATER_NEXTLINE, moduleName, pluginPath, err)
			}
			// Whether it succeeded or failed, we stop searching once we find a candidate file.
			return
		}
	}
}

// --- Module Cache & State --- (No changes here)
var (
	moduleCache               = make(map[string]object.Object)
	moduleCacheMutex          sync.RWMutex
	modulesBeingImported      = make(map[string]bool)
	modulesBeingImportedMutex sync.Mutex
	currentScriptDir          string
)

// SetCurrentScriptDir sets the context for relative imports. (No changes here)
func SetCurrentScriptDir(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.InterpreterModulesWarnAbsolutePath, path, err)
		currentScriptDir = filepath.Dir(path)
		return
	}
	info, err := os.Stat(absPath)
	if err == nil && info.IsDir() {
		currentScriptDir = absPath
	} else {
		currentScriptDir = filepath.Dir(absPath)
	}
}

// evalImportStatement handles simple `import module` or `import dotted.module.name`
func evalImportStatement(stmt *ast.ImportStatement, ctx *InterpreterContext) object.Object {
	moduleName := stmt.Name.Value
	moduleObj := loadModuleByName(moduleName, stmt.Token, ctx)
	if object.IsError(moduleObj) {
		return moduleObj
	}

	// --- THIS IS THE FIX ---
	var nameToBind string
	var finalModuleObject object.Object

	if stmt.Alias != nil {
		// `import a.b.c as d` -> bind module `c` to name `d`
		nameToBind = stmt.Alias.Value
		finalModuleObject = moduleObj
	} else {
		// `import a.b.c` -> bind module `a` to name `a`
		nameToBind = strings.Split(moduleName, ".")[0]
		// We still need to load the full path to ensure submodules are created,
		// but we only bind the top-level module object. `loadModuleByName` does this.
		// We need to fetch the top-level module object.
		finalModuleObject = loadModuleByName(nameToBind, stmt.Token, ctx)
		if object.IsError(finalModuleObject) {
			return finalModuleObject
		}
	}

	// If it was a dotted import without an alias, create the nested structure
	if strings.Contains(moduleName, ".") && stmt.Alias == nil {
		parts := strings.Split(moduleName, ".")
		createNestedModuleStructure(parts, moduleObj, ctx.Env)
	}

	// Bind the correct module object to the correct name in the current environment.
	ctx.Env.Set(nameToBind, finalModuleObject)
	// --- END OF FIX ---

	return object.NULL
}

// createNestedModuleStructure creates the nested module structure for dotted imports
// For `import os.path`, this creates: os -> Module{path: actualPathModule}
func createNestedModuleStructure(parts []string, finalModule object.Object, env *object.Environment) {
	if len(parts) == 0 {
		return
	}

	if len(parts) == 1 {
		env.Set(parts[0], finalModule)
		return
	}

	topLevelName := parts[0]
	remainingParts := parts[1:]

	// Get or create the top-level module
	var topLevelModule *object.Module
	if existing, found := env.Get(topLevelName); found {
		if mod, ok := existing.(*object.Module); ok {
			topLevelModule = mod
		} else {
			// If it exists but isn't a module, we have a conflict
			return
		}
	} else {
		// Create a new module container
		topLevelModule = &object.Module{
			Name: topLevelName,
			Path: "", // This is a synthetic module
			Env:  object.NewEnclosedEnvironment(nil),
		}
		env.Set(topLevelName, topLevelModule)
	}

	// Recursively create the nested structure
	createNestedModuleStructure(remainingParts, finalModule, topLevelModule.Env)
}

// evalFromImportStatement handles `from module.submodule import name1, name2 as alias, *`
func evalFromImportStatement(stmt *ast.FromImportStatement, ctx *InterpreterContext) object.Object {
	modulePathNode, ok := stmt.ModulePath.(*ast.Identifier)
	if !ok {
		return object.NewErrorWithLocation(stmt.Token, constants.ImportError, constants.InterpreterModulesComplexImportError)
	}
	moduleName := modulePathNode.Value

	moduleObj := loadModuleByName(moduleName, stmt.Token, ctx)
	if object.IsError(moduleObj) {
		return moduleObj
	}

	mod, ok := moduleObj.(*object.Module)
	if !ok {
		return object.NewErrorWithLocation(stmt.Token, constants.InterpreterModulesLoadedNotModuleError, moduleName, moduleObj.Type())
	}

	if stmt.ImportAll {
		for name, val := range mod.Env.Items() {
			if !strings.HasPrefix(name, constants.Underscore) {
				ctx.Env.Set(name, val)
			}
		}
	} else {
		for _, importPair := range stmt.Names {
			originalName := importPair.OriginalName.Value
			val, found := mod.Env.Get(originalName)
			if !found {
				return object.NewErrorWithLocation(importPair.OriginalName.Token, constants.ImportError, constants.InterpreterModulesImportNameError, originalName, mod.Name)
			}
			nameInCurrentScope := originalName
			if importPair.Alias != nil {
				nameInCurrentScope = importPair.Alias.Value
			}
			ctx.Env.Set(nameInCurrentScope, val)
		}
	}
	return object.NULL
}

// getModuleSearchPaths returns the ordered list of paths to search for modules
func getModuleSearchPaths() []string {
	var searchPaths []string

	// 1. The standard library path (HIGHEST PRIORITY after native modules)
	stdLibPath := GetStandardLibraryPath()
	if stdLibPath != "" {
		searchPaths = append(searchPaths, stdLibPath)
	}

	// 2. The global `modules` directory (for 3rd-party installed packages)
	searchPaths = append(searchPaths, constants.ModulesDirectoryForThirdPartyPackagesInstalled)

	// 3. The `lib` directory (for project-specific modules)
	searchPaths = append(searchPaths, constants.LibDirectoryForProjectSpecificModules)

	// 4. The directory of the current script (for local/relative imports)
	if currentScriptDir != "" {
		searchPaths = append(searchPaths, currentScriptDir)
	}

	// 5. Fallback to the current working directory
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, cwd)
	}

	return searchPaths
}

// loadModuleByName loads a module (native or file-based) by its string name
// Enhanced to handle dotted module names robustly
func loadModuleByName(moduleName string, tokenForError lexer.Token, ctx *InterpreterContext) object.Object {
	// Step 1: Check if the module is already registered (either built-in or previously loaded).
	nativeMod, isNative := object.GetNativeModule(moduleName)
	if isNative {
		return nativeMod
	}

	// Step 2: If not found, try to find and load a corresponding Go plugin (.so file).
	// This is the core of the new dynamic FFI system.
	tryLoadPluginForModule(moduleName)

	// Step 3: After attempting to load a plugin, check the native module cache again.
	// If the plugin loaded successfully, it will have registered the module.
	nativeMod, isNative = object.GetNativeModule(moduleName)
	if isNative {
		return nativeMod
	}

	// Step 4: If still not found, proceed with searching for a Pylearn (.py) file.
	// This logic remains the same as before.
	searchPaths := getModuleSearchPaths()
	// // First check for native modules
	// nativeMod, isNative := object.GetNativeModule(moduleName)
	// if isNative {
	// 	return nativeMod
	// }

	// // Define search paths for Pylearn modules
	// searchPaths := getModuleSearchPaths()

	// Search for the module file in the defined paths
	foundPath, foundModuleName := findModuleInPaths(moduleName, searchPaths)
	if foundPath == "" {
		return object.NewErrorWithLocation(tokenForError, constants.ModuleNotFoundError, constants.NoModuleNamed_STRINGFORMATER, moduleName)
	}

	absPath, err := filepath.Abs(foundPath)
	if err != nil {
		return object.NewErrorWithLocation(tokenForError, constants.ImportError, constants.CouldNotDetermineAbsolutePathForModule_STRINGFORMATER, moduleName)
	}

	// Handle circular import detection
	modulesBeingImportedMutex.Lock()
	if modulesBeingImported[absPath] {
		modulesBeingImportedMutex.Unlock()
		return object.NewErrorWithLocation(tokenForError, constants.ImportError, constants.InterpreterModulesCircularImportDetected, foundModuleName, absPath)
	}
	modulesBeingImported[absPath] = true
	modulesBeingImportedMutex.Unlock()

	defer func() {
		modulesBeingImportedMutex.Lock()
		delete(modulesBeingImported, absPath)
		modulesBeingImportedMutex.Unlock()
	}()

	// Check cache
	moduleCacheMutex.RLock()
	cachedObject, found := moduleCache[absPath]
	moduleCacheMutex.RUnlock()

	if found {
		if cachedObject == object.IMPORT_PLACEHOLDER {
			return object.NewErrorWithLocation(tokenForError, constants.ImportError, constants.InterpreterModulesCircularImportBeingImported, foundModuleName, absPath)
		}
		return cachedObject
	}

	// Load and execute the module
	return loadAndExecuteModule(absPath, foundModuleName, tokenForError, ctx)
}

// findModuleInPaths searches for a module in the given search paths
// Returns the found path and the module name, or empty strings if not found
func findModuleInPaths(moduleName string, searchPaths []string) (string, string) {
	moduleParts := strings.Split(moduleName, ".")

	for _, searchDir := range searchPaths {
		// Try different combinations for finding the module
		if foundPath := tryFindModule(searchDir, moduleParts); foundPath != "" {
			return foundPath, moduleName
		}
	}

	return "", ""
}

// tryFindModule attempts to find a module in a given directory
func tryFindModule(searchDir string, moduleParts []string) string {
	// For a.b.c, we try:
	// 1. searchDir/a/b/c.py
	// 2. searchDir/a/b/c/__init__.py
	// 3. searchDir/a.py (if len(moduleParts) == 1)
	// 4. searchDir/a/__init__.py (if len(moduleParts) == 1)

	if len(moduleParts) == 1 {
		// Single module name
		moduleName := moduleParts[0]

		// Try moduleName.py
		potentialFile := filepath.Join(searchDir, moduleName+constants.DOT_Py)
		if _, err := os.Stat(potentialFile); err == nil {
			return potentialFile
		}

		// Try moduleName/__init__.py
		potentialInit := filepath.Join(searchDir, moduleName, constants.Init_DOT_Py)
		if _, err := os.Stat(potentialInit); err == nil {
			return potentialInit
		}
	} else {
		// Dotted module name
		currentPath := filepath.Join(searchDir, filepath.Join(moduleParts...))

		// Try .../a/b/c.py
		potentialFile := currentPath + constants.DOT_Py
		if _, err := os.Stat(potentialFile); err == nil {
			return potentialFile
		}

		// Try .../a/b/c/__init__.py
		potentialInit := filepath.Join(currentPath, constants.Init_DOT_Py)
		if _, err := os.Stat(potentialInit); err == nil {
			return potentialInit
		}
	}

	return ""
}

// loadAndExecuteModule reads, parses, and executes a module file
func loadAndExecuteModule(absPath, foundModuleName string, tokenForError lexer.Token, ctx *InterpreterContext) object.Object {
	sourceBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return object.NewErrorWithLocation(tokenForError, constants.ImportError, constants.InterpreterModulesCouldNotReadModule, foundModuleName, absPath, err)
	}
	source := string(sourceBytes)

	// Create module environment
	moduleEnv := object.NewEnclosedEnvironment(nil)
	for name, builtin := range builtins.Builtins {
		moduleEnv.Set(name, builtin)
	}
	for name, class := range object.BuiltinExceptionClasses {
		moduleEnv.Set(name, class)
	}
	moduleEnv.Set(constants.DunderName, &object.String{Value: foundModuleName})

	moduleObj := &object.Module{Name: foundModuleName, Path: absPath, Env: moduleEnv}

	// Set placeholder to detect circular imports
	moduleCacheMutex.Lock()
	moduleCache[absPath] = object.IMPORT_PLACEHOLDER
	moduleCacheMutex.Unlock()

	// Parse the module
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		errMsg := fmt.Sprintf(constants.InterpreterModulesSyntaxErrorImportedModule, foundModuleName, absPath)
		for _, parseErr := range p.Errors() {
			errMsg += fmt.Sprintf(constants.InterpreterModulesSyntaxErrorFormat, parseErr)
		}
		syntaxErr := object.NewError(constants.SyntaxError, errMsg)
		moduleCacheMutex.Lock()
		moduleCache[absPath] = syntaxErr
		moduleCacheMutex.Unlock()
		return syntaxErr
	}

	// Set up the execution context
	originalCallerScriptDir := currentScriptDir
	// When importing a package (`__init__.py`), the script dir for its relative imports
	// should be the package's directory, not the directory of the __init__.py file itself.
	SetCurrentScriptDir(filepath.Dir(absPath))

	moduleCtx := ctx.NewChildContext(moduleEnv).(*InterpreterContext)
	evaluated := Eval(program, moduleCtx)

	SetCurrentScriptDir(originalCallerScriptDir)

	moduleCacheMutex.Lock()
	defer moduleCacheMutex.Unlock()

	if object.IsError(evaluated) {
		runtimeErr := evaluated.(*object.Error)
		errorMessage := fmt.Sprintf(constants.InterpreterModulesErrorImportModuleFull, foundModuleName, absPath, runtimeErr.Message)
		if runtimeErr.Line > 0 {
			errorMessage = fmt.Sprintf(constants.InterpreterModulesErrorImportModuleLine, foundModuleName, filepath.Base(absPath), runtimeErr.Line, runtimeErr.Message)
		}
		errorToCache := object.NewError(constants.ModuleNotFoundError, errorMessage)
		errorToCache.Line = runtimeErr.Line
		errorToCache.Column = runtimeErr.Column
		moduleCache[absPath] = errorToCache
		return errorToCache
	}

	moduleCache[absPath] = moduleObj
	return moduleObj
}
