// internal/vm/modules.go
package vm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync" // Maybe needed if importer becomes concurrent later

	"github.com/deniskipeles/pylearn/internal/lexer"  // Needs lexer
	"github.com/deniskipeles/pylearn/internal/object" // Needs object types
	"github.com/deniskipeles/pylearn/internal/parser" // Needs parser
)

// Define the placeholder object within the vm package if not already global in object
// var importPlaceholder = &object.Null{} // Or create a unique placeholder type

// Importer manages module loading, caching, and execution for the VM.
type Importer struct {
	vm          *VM // Reference back to the VM for state (cache, paths, execution)
	searchPaths []string
	moduleCache map[string]object.Object // Cache[absPath] -> Module/Error/Placeholder
	// Use RWMutex if imports could potentially happen concurrently in the future
	cacheLock          sync.RWMutex
	currentlyImporting map[string]bool // Path -> bool (for circular detection)
	importLock         sync.Mutex      // To protect currentlyImporting map
}

// NewImporter creates a new importer instance associated with a VM.
func NewImporter(vm *VM, mainScriptPath string) *Importer {
	return &Importer{
		vm:                 vm,
		searchPaths:        initializeSearchPaths(mainScriptPath),
		moduleCache:        make(map[string]object.Object),
		currentlyImporting: make(map[string]bool),
	}
}

// ImportModule handles the OpImportName logic.
func (imp *Importer) ImportModule(moduleName string) (object.Object, error) {
	fmt.Printf("    Importer: Attempting to import '%s'\n", moduleName)

	// --- 1. Check Native Built-ins First (via VM) ---
	if module, found := imp.vm.nativeModules[moduleName]; found {
		fmt.Printf("    Importer: Found native module '%s'\n", moduleName)
		return module, nil // Return the native module directly
	}

	// --- TODO: Handle dotted names like "os.path" ---
	// This would involve recursive calls to ImportModule or attribute lookups.
	// For now, focus on single names.

	// --- 2. Find Module File ---
	foundPath, isPackage, err := imp.findModule(moduleName)
	if err != nil {
		return nil, object.NewError("ImportError: no module named '%s'", moduleName)
	}
	if isPackage {
		fmt.Printf("    Importer: Found package '%s' at '%s'\n", moduleName, foundPath)
		// Load the __init__.py file for packages
		foundPath = filepath.Join(foundPath, "__init__.py")
		// Check if __init__.py actually exists
		if _, statErr := os.Stat(foundPath); os.IsNotExist(statErr) {
			// Found directory, but no __init__.py - treat as namespace package? Or error?
			// Python creates an empty module object for namespace packages.
			fmt.Printf("    Importer: Namespace package found (no __init__.py), creating empty module.\n")
			// Create and cache an empty module immediately
			absPath, _ := filepath.Abs(filepath.Dir(foundPath)) // Use dir path for package key
			absPath = filepath.Clean(absPath)
			moduleEnv := object.NewEnvironment()
			moduleObj := &object.Module{Name: moduleName, Path: absPath, Env: moduleEnv} // Path is dir
			imp.cacheLock.Lock()
			imp.moduleCache[absPath] = moduleObj
			imp.cacheLock.Unlock()
			return moduleObj, nil
		} else if statErr != nil {
			// Error statting __init__.py other than not exists
			return nil, object.NewError("ImportError: error accessing package init file '%s': %v", foundPath, statErr)
		}
		// If __init__.py exists, proceed to load it below using its path.
		fmt.Printf("    Importer: Found package init file: '%s'\n", foundPath)

	} else {
		fmt.Printf("    Importer: Found module file '%s' at '%s'\n", moduleName, foundPath)
	}

	// --- 3. Check Module Cache (using absolute path) ---
	absPath, err := filepath.Abs(foundPath)
	if err != nil {
		absPath = foundPath
	} // Use original on error
	absPath = filepath.Clean(absPath)

	imp.cacheLock.RLock()
	cachedModule, found := imp.moduleCache[absPath]
	imp.cacheLock.RUnlock()

	if found {
		fmt.Printf("    Importer: Found '%s' in cache (%T)\n", absPath, cachedModule)
		// Handle import-in-progress placeholder for circular dependencies
		if placeholder, ok := cachedModule.(*object.Null); ok && placeholder == object.IMPORT_PLACEHOLDER {
			// TODO: Python's behavior here is complex. It often returns the partially
			// initialized module object. For simplicity, we error out.
			// Returning the placeholder itself might work if attribute access can handle it.
			return nil, object.NewError("ImportError: circular import detected for module '%s' (placeholder found)", moduleName)
		}
		// Handle previous import error
		if object.IsError(cachedModule) {
			return nil, cachedModule.(error) // Return the cached error
		}
		// Success, return cached module
		return cachedModule, nil
	}

	// --- 4. Handle Circular Imports (Check and Mark) ---
	imp.importLock.Lock() // Lock before checking/modifying currentlyImporting
	if imp.currentlyImporting[absPath] {
		imp.importLock.Unlock()
		fmt.Printf("    Importer: Circular import detected for '%s'\n", absPath)
		// Put placeholder in cache *now* so subsequent checks find it
		imp.cacheLock.Lock()
		imp.moduleCache[absPath] = object.IMPORT_PLACEHOLDER
		imp.cacheLock.Unlock()
		return nil, object.NewError("ImportError: circular import detected during import of '%s'", moduleName)
	}
	imp.currentlyImporting[absPath] = true
	// Put placeholder in cache before starting compilation/execution
	imp.cacheLock.Lock()
	imp.moduleCache[absPath] = object.IMPORT_PLACEHOLDER
	imp.cacheLock.Unlock()
	imp.importLock.Unlock() // Unlock after modifying currentlyImporting and cache

	// Ensure removal from currentlyImporting map when this import finishes (success or error)
	defer func() {
		imp.importLock.Lock()
		delete(imp.currentlyImporting, absPath)
		imp.importLock.Unlock()
	}()

	// --- 5. Read, Compile, and Execute Module ---
	fmt.Printf("    Importer: Loading and executing module '%s' from '%s'\n", moduleName, absPath)
	moduleObject, execErr := imp.loadAndExecModule(moduleName, absPath)

	// --- 6. Update Cache (Overwrite placeholder with result or error) ---
	imp.cacheLock.Lock()
	if execErr != nil {
		fmt.Printf("    Importer: Error executing module '%s': %v\n", moduleName, execErr)
		pylearnErr := object.NewErrorFromGoErr(execErr) // Convert Go error to Pylearn Error
		imp.moduleCache[absPath] = pylearnErr           // Cache the error
		imp.cacheLock.Unlock()
		return nil, pylearnErr // Propagate the Pylearn error
	}

	fmt.Printf("    Importer: Successfully imported and executed '%s'. Caching.\n", moduleName)
	imp.moduleCache[absPath] = moduleObject // Cache the successful module object
	imp.cacheLock.Unlock()
	return moduleObject, nil // Return the result
}

// findModule searches sys.path for the module file or package directory.
// Returns (foundPath, isPackage, error)
func (imp *Importer) findModule(moduleName string) (string, bool, error) {
	// Convert module name to potential file/dir names
	modulePathParts := strings.Split(moduleName, ".") // For future dotted handling
	baseName := modulePathParts[len(modulePathParts)-1]
	moduleFilename := baseName + ".py"
	packageDirname := baseName

	for _, dir := range imp.searchPaths {
		// 1. Check for package directory first
		potentialPackagePath := filepath.Join(dir, packageDirname)
		info, err := os.Stat(potentialPackagePath)
		if err == nil && info.IsDir() {
			// Found a directory. Check for __init__.py inside later.
			return potentialPackagePath, true, nil // Return dir path, indicate package
		}

		// 2. Check for module file
		potentialFilePath := filepath.Join(dir, moduleFilename)
		if _, err := os.Stat(potentialFilePath); err == nil {
			// Found .py file
			return potentialFilePath, false, nil // Return file path, not a package
		} else if !os.IsNotExist(err) {
			// Report other errors (permissions etc.)?
			fmt.Printf("Warning: Error checking import path '%s': %v\n", potentialFilePath, err)
		}
	}
	return "", false, fmt.Errorf("module/package not found")
}

// loadAndExecModule reads, compiles, and executes a module's code.
// Returns the populated module object or an error.
func (imp *Importer) loadAndExecModule(moduleName, absPath string) (*object.Module, error) {
	// --- a. Read Source ---
	sourceBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module '%s': %w", absPath, err)
	}
	source := string(sourceBytes)

	// --- b. Compile Source ---
	// TODO: Implement bytecode caching (.pyc) here later
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return nil, fmt.Errorf("parse error in module '%s': %s", absPath, strings.Join(p.Errors(), "\n"))
	}

	comp := NewCompiler() // Fresh compiler isolates symbols
	errComp := comp.Compile(program)
	if errComp != nil {
		return nil, fmt.Errorf("compile error in module '%s': %w", absPath, errComp)
	}
	bytecode := comp.Bytecode()
	moduleSymbols := comp.GetFinalSymbolTable() // Assume SymbolTable() method exists
	numLocals := 0
	if moduleSymbols != nil {
		numLocals = moduleSymbols.NumDefinitions() // Use getter if needed
	} else {
		fmt.Printf("    Importer: WARNING - Module symbol table was nil after compilation for %s\n", moduleName)
	}

	// --- c. Create Module Object & Environment ---
	// Using a separate Env helps isolate module's global scope if VM supports it.
	// If OpSet/GetGlobal aren't context-aware, this Env might not be fully utilized yet.
	moduleEnv := object.NewEnvironment()
	moduleObj := &object.Module{
		Name: moduleName,
		Path: absPath,
		Env:  moduleEnv,
	}

	// --- d. Execute Module Bytecode in Current VM ---
	// Create a frame specifically for executing the module's top-level code.
	moduleFunc := &CompiledFunction{
		Instructions:  bytecode.Instructions,
		NumLocals:     numLocals, // Get actual local count from compiler
		NumParameters: 0,
		Name:          fmt.Sprintf("<module %s>", moduleName),
	}
	moduleClosure := &Closure{Fn: moduleFunc, Constants: bytecode.Constants} // Use vm.Closure
	// Calculate base pointer for module frame. Module code doesn't take arguments pushed
	// before it, so its BP can be the current SP. Globals defined will go into VM globals for now.
	basePointer := imp.vm.sp
	moduleFrame := NewFrame(moduleClosure, basePointer, bytecode.Constants, moduleEnv) // Use vm.NewFrame

	// TODO: Associate moduleEnv with moduleFrame if OpSet/GetGlobal become context-aware.

	errPush := imp.vm.pushFrame(moduleFrame)
	if errPush != nil {
		return nil, fmt.Errorf("failed to push frame for module '%s': %w", moduleName, errPush)
	}

	// Execute just this frame. Use runSingleCallbackFrame to run until this specific frame pops.
	// Store frame index before running, check after.
	moduleFrameIndex := imp.vm.frameIndex
	execModErr := imp.vm.runSingleCallbackFrame(moduleFrameIndex)

	// After runSingleCallbackFrame, the module frame should have been popped.
	// The result of the module execution (usually None) is on the stack.
	if imp.vm.frameIndex != moduleFrameIndex-1 {
		fmt.Printf("WARNING: Frame index mismatch after module execution (expected %d, got %d)\n", moduleFrameIndex-1, imp.vm.frameIndex)
		// Potentially reset frame index if needed? Risky.
	}

	// Pop the module's implicit return value (usually NULL)
	_, popErr := imp.vm.pop()
	if popErr != nil {
		fmt.Printf("WARNING: Error popping module return value: %v\n", popErr)
		// Continue, but state might be slightly off
	}

	if execModErr != nil {
		// Error occurred during module execution
		return nil, fmt.Errorf("runtime error during module '%s' execution: %w", moduleName, execModErr)
	}

	fmt.Println("    Importer: Populating module environment from VM globals...")
	if moduleSymbols != nil {
		// Get the outermost scope (where module-level defs are)
		outermostTable := moduleSymbols.OuterMost() // Use new method
		// Iterate through symbols defined in the module's outermost scope
		for name, symbol := range outermostTable.Store() { // Use new method
			if symbol.Scope == GlobalScope { // Check scope
				if symbol.Index < len(imp.vm.globals) {
					val := imp.vm.globals[symbol.Index]
					if val != nil {
						fmt.Printf("      Copying global '%s' (index %d) to module env\n", name, symbol.Index)
						moduleEnv.Set(name, val)
					} else {
						fmt.Printf("      Skipping nil global '%s' (index %d)\n", name, symbol.Index)
					}
				} else {
					fmt.Printf("      WARNING: Global index %d for '%s' out of VM globals bounds (%d)\n", symbol.Index, name, len(imp.vm.globals))
				}
			}
		}
	} else {
		fmt.Println("    Importer: WARNING - Could not get module symbol table to populate environment.")
	}

	// --- e. Populate Module Env (Still needs better solution) ---
	// If OpSet/GetGlobal are modified to use frame.Env, this step becomes unnecessary.
	// If not, we need a way to copy the relevant globals set during execution
	// into moduleObj.Env. This is hard without symbol table info from compilation preserved.
	// WORKAROUND (Inaccurate): Copy *all* current VM globals? Very bad.
	// Better WORKAROUND: Do nothing for now, rely on module functions accessing globals directly.

	return moduleObj, nil
}

// initializeSearchPaths determines the initial module search path list.
func initializeSearchPaths(mainScriptPath string) []string {
	paths := []string{}

	// 1. Directory of the main script
	if mainScriptPath != "" {
		absScriptPath, err := filepath.Abs(mainScriptPath)
		if err == nil {
			paths = append(paths, filepath.Dir(absScriptPath))
		} else {
			fmt.Printf("Warning: Could not get absolute path for script '%s': %v\n", mainScriptPath, err)
			// Maybe add current working directory as fallback?
			if cwd, errCwd := os.Getwd(); errCwd == nil {
				paths = append(paths, cwd)
			}
		}
	} else {
		// No script provided (e.g., REPL), start with current working directory
		if cwd, errCwd := os.Getwd(); errCwd == nil {
			paths = append(paths, cwd)
		}
	}

	// 2. Standard library path (relative to executable or fixed?)
	// TODO: Determine where your standard library modules will live
	// paths = append(paths, "/path/to/pylearn/stdlib") // Example

	// 3. PYTHONPATH environment variable
	pythonPath := os.Getenv("PYLEARNPATH") // Use a custom name
	if pythonPath != "" {
		pathEntries := filepath.SplitList(pythonPath) // Handles OS-specific separators
		paths = append(paths, pathEntries...)
	}

	// 4. Current Working Directory (if not already added)
	cwd, errCwd := os.Getwd()
	if errCwd == nil {
		alreadyAdded := false
		for _, p := range paths {
			if p == cwd {
				alreadyAdded = true
				break
			}
		}
		if !alreadyAdded {
			paths = append(paths, cwd)
		}
	}

	// Remove duplicates and clean paths
	cleanedPaths := []string{}
	seen := make(map[string]bool)
	for _, p := range paths {
		cleanP := filepath.Clean(p)
		if !seen[cleanP] {
			cleanedPaths = append(cleanedPaths, cleanP)
			seen[cleanP] = true
		}
	}

	fmt.Printf("Initialized Search Paths: %v\n", cleanedPaths) // Debug log
	return cleanedPaths
}

// Helper needed in object package or here temporarily
// Ensure object.IMPORT_PLACEHOLDER exists in object.go
// var IMPORT_PLACEHOLDER = &object.Null{} // Or define a unique type
