// pylearn/internal/interpreter/dynamic_loader.go

package interpreter

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
)

// pyLoadModuleFromPathFn implements pylearn_importlib.load_module_from_path(path_string)
// It will be called from Pylearn code.
func pyLoadModuleFromPathFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.InterpreterDynamicLoaderLoadModulePathArgCountError, len(args))
	}
	pathObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.InterpreterDynamicLoaderLoadModulePathArgTypeError, args[0].Type())
	}
	modulePath := pathObj.Value

	// 1. Resolve Path
	resolvedPath := modulePath
	if !filepath.IsAbs(modulePath) {
		if currentScriptDir == constants.EmptyString {
			return object.NewError(constants.ImportError, constants.InterpreterDynamicLoaderRelativePathError, modulePath)
		}
		resolvedPath = filepath.Join(currentScriptDir, modulePath)
	}

	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return object.NewError(constants.ImportError, constants.InterpreterDynamicLoaderAbsPathError, resolvedPath, err)
	}
	absPath = filepath.Clean(absPath)

	moduleName := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
	if moduleName == constants.EmptyString {
		nameParts := strings.Split(filepath.ToSlash(absPath), constants.Slash)
		if len(nameParts) > 0 {
			moduleName = nameParts[len(nameParts)-1]
			if moduleName == constants.EmptyString && len(nameParts) > 1 {
				moduleName = nameParts[len(nameParts)-2]
			}
		}
		if moduleName == constants.EmptyString {
			moduleName = constants.InterpreterDynamicLoaderDynamicModule
		}
	}

	// 2. Circular Import Check
	modulesBeingImportedMutex.Lock()
	if modulesBeingImported[absPath] {
		modulesBeingImportedMutex.Unlock()
		moduleCacheMutex.RLock()
		placeholder, exists := moduleCache[absPath]
		moduleCacheMutex.RUnlock()
		if exists && placeholder == object.IMPORT_PLACEHOLDER {
			return object.NewError(constants.ImportError, constants.InterpreterDynamicLoaderCircularImportPlaceholder, absPath)
		}
		return object.NewError(constants.ImportError, constants.InterpreterDynamicLoaderCircularImportPlaceholder, absPath)
	}
	modulesBeingImported[absPath] = true
	modulesBeingImportedMutex.Unlock()

	defer func() {
		modulesBeingImportedMutex.Lock()
		delete(modulesBeingImported, absPath)
		modulesBeingImportedMutex.Unlock()
	}()

	// 3. Cache Check
	moduleCacheMutex.RLock()
	cachedModule, found := moduleCache[absPath]
	moduleCacheMutex.RUnlock()

	if found {
		if cachedModule == object.IMPORT_PLACEHOLDER {
			return object.NewError(constants.ImportError, constants.InterpreterDynamicLoaderCircularImportBeingImported, moduleName, absPath)
		}
		if object.IsError(cachedModule) {
			return cachedModule
		}
		return cachedModule
	}

	// 4. Read Source File
	sourceBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return object.NewError(constants.ImportError, constants.InterpreterDynamicLoaderFailedToReadFile, moduleName, absPath, err)
	}
	source := string(sourceBytes)

	// 5. Create Module Object and Environment
	moduleEnv := object.NewEnclosedEnvironment(nil) // Module's own global scope

	for name, builtin := range builtins.Builtins {
		moduleEnv.Set(name, builtin)
	}
	for name, class := range object.BuiltinExceptionClasses {
		moduleEnv.Set(name, class)
	}
	moduleEnv.Set(constants.DunderName, &object.String{Value: moduleName})

	moduleObj := &object.Module{
		Name: moduleName,
		Path: absPath,
		Env:  moduleEnv,
	}

	moduleCacheMutex.Lock()
	moduleCache[absPath] = object.IMPORT_PLACEHOLDER
	moduleCacheMutex.Unlock()

	// 6. Lex, Parse
	l := lexer.New(source)
	p := parser.New(l)
	programAST := p.ParseProgram()
	if len(p.Errors()) != 0 {
		moduleCacheMutex.Lock()
		errMsg := fmt.Sprintf(constants.InterpreterDynamicLoaderSyntaxErrorInModule, moduleName, absPath)
		for _, pe := range p.Errors() {
			errMsg += constants.InterpreterDynamicLoaderSyntaxErrorFormat + pe + constants.Newline
		}
		syntaxErr := object.NewError(constants.SyntaxError, errMsg)
		moduleCache[absPath] = syntaxErr
		moduleCacheMutex.Unlock()
		return syntaxErr
	}

	// 7. Execute (Eval)
	originalCallerScriptDir := currentScriptDir
	SetCurrentScriptDir(absPath)

	// <<< FIX: Create a new execution context for the module's evaluation
	moduleCtx := ctx.NewChildContext(moduleEnv).(*InterpreterContext)
	evalResult := Eval(programAST, moduleCtx)

	SetCurrentScriptDir(originalCallerScriptDir)

	// 8. Finalize Cache and Return
	moduleCacheMutex.Lock()
	if object.IsError(evalResult) {
		errObj := evalResult.(*object.Error)
		// Prepend module context to the error message if not already there
		if !strings.HasPrefix(errObj.Message, fmt.Sprintf(constants.ErrorDuringExecutionModule, moduleName)) {
			errObj.Message = fmt.Sprintf(constants.InterpreterDynamicLoaderErrorExecutingModule, moduleName, absPath, errObj.Message)
		}
		moduleCache[absPath] = errObj
		moduleCacheMutex.Unlock()
		return errObj
	}

	moduleCache[absPath] = moduleObj
	moduleCacheMutex.Unlock()

	return moduleObj
}

// GetPyLoadModuleFromPathFn returns the function to be injected.
func GetPyLoadModuleFromPathFn() object.BuiltinFunction {
	return pyLoadModuleFromPathFn
}