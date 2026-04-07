// cmd/interpreter/main.go
package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deniskipeles/pylearn/internal/asyncruntime"
	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
	"github.com/deniskipeles/pylearn/internal/interpreter"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
	"github.com/deniskipeles/pylearn/internal/stdlib/pyimportlib"
	"github.com/deniskipeles/pylearn/internal/package_manager"
	pysys_init "github.com/deniskipeles/pylearn/internal/stdlib/pysys"

	"github.com/deniskipeles/pylearn/internal/stdlib/ffi3"

	_ "github.com/deniskipeles/pylearn/internal/stdlib/httpserver"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/net"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/sqlite"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/json"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/math"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/template"
	
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyaio"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyconcurrent/futures"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyflasky"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyhttp"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyos"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pytime"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyre"
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pycrypto"
	// _ "github.com/deniskipeles/pylearn/internal/stdlib/pysys"

)

var PylearnAsyncRuntime *asyncruntime.Runtime // Global instance

// InitAsyncRuntime is a placeholder and not used by the corrected code.
// The real initialization happens in interpreter.InitAsyncRuntime.
func InitAsyncRuntime() {
	if PylearnAsyncRuntime == nil {
		PylearnAsyncRuntime = asyncruntime.NewRuntime()
	}
}

func main() {
	goArgs := os.Args

	// --- THIS IS THE NEW COMMAND DISPATCHER ---
	if len(goArgs) > 1 && goArgs[1] == "get" {
		package_manager.HandleGetCommand(goArgs[2:]) // Pass the remaining args (the URL)
		return // Exit after handling the 'get' command
	}
	interpreter.InitAsyncRuntime()

	pyimportlib.SetLoadModuleFunc(interpreter.GetPyLoadModuleFromPathFn())

	// goArgs := os.Args
	pylearnArgv := make([]object.Object, len(goArgs))
	for i, arg := range goArgs {
		pylearnArgv[i] = &object.String{Value: arg}
	}
	argvList := &object.List{Elements: pylearnArgv}
	pysys_init.InitializeSysModule(argvList)

	scriptName := constants.CmdInterpreterMainStdin
	if len(goArgs) < 2 {
		fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainUsage, filepath.Base(goArgs[0]))
		os.Exit(1)
	}
	filename := goArgs[1]
	scriptName = filename
	interpreter.SetCurrentScriptDir(filename)

	sourceBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainErrorReadingFile, filename, err)
		os.Exit(1)
	}
	source := string(sourceBytes)

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(os.Stderr, p.Errors())
		os.Exit(1)
	}

	// Create the top-level environment
	env := object.NewEnvironment()
	for name, builtin := range builtins.Builtins {
		env.Set(name, builtin)
	}
	for name, class := range object.BuiltinExceptionClasses {
		env.Set(name, class)
	}
	env.Set(constants.DunderName, &object.String{Value: constants.DunderMain})

	// Evaluate the AST *using the persistent environment*
	// OLD: mainCtx := &interpreter.InterpreterContext{Env: env}
	// NEW: Use the constructor
	mainCtx := interpreter.NewInterpreterContext(env)
	ffi3.SetGlobalExecutionContext(mainCtx)

	// Evaluate the program. This will define functions, including 'main_program'.
	evaluated := interpreter.Eval(program, mainCtx) // <<< FIX: Pass mainCtx

	// --- Async Main Program Handling ---
	mainFuncObj, mainFound := env.Get("main_program")
	if mainFound {
		if mainPylFunc, isPylFunc := mainFuncObj.(*object.Function); isPylFunc {
			if mainPylFunc.IsAsync {
				fmt.Println(constants.CmdInterpreterMainFoundAsyncMainProgram)

				if interpreter.PylearnAsyncRuntime == nil || interpreter.PylearnAsyncRuntime.EventLoop == nil {
					fmt.Fprintln(os.Stderr, constants.CmdInterpreterMainFatalAsyncRuntimeNotInitialized)
					os.Exit(1)
				}

				asyncResult := interpreter.PylearnAsyncRuntime.EventLoop.CreateCoroutine(
					func(goCtx context.Context) (interface{}, error) {
						if mainPylFunc.Body == nil {
							return nil, fmt.Errorf(constants.CmdInterpreterMainAsyncFunctionHasNoBody, mainPylFunc.Name)
						}

						var callEnv *object.Environment
						if mainPylFunc.Env != nil {
							callEnv = object.NewEnclosedEnvironment(mainPylFunc.Env)
						} else {
							fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainAsyncFunctionNilClosureEnv, mainPylFunc.Name)
							callEnv = object.NewEnclosedEnvironment(env)
						}

						// <<< FIX: Create a new context for the async function's execution
						asyncEvalCtx := &interpreter.InterpreterContext{Env: callEnv}
						evalResult := interpreter.Eval(mainPylFunc.Body, asyncEvalCtx)

						if retVal, isRet := evalResult.(*object.ReturnValue); isRet {
							return retVal.Value, nil
						}
						if errObj, isErr := evalResult.(*object.Error); isErr {
							return nil, fmt.Errorf(constants.CmdInterpreterMainPylearnErrorFormat, errObj.Message, errObj.Line, errObj.Column)
						}
						return object.NULL, nil
					},
				)

				finalValInterface, finalErr := interpreter.PylearnAsyncRuntime.Await(asyncResult)
				interpreter.PylearnAsyncRuntime.EventLoop.Stop()

				if finalErr != nil {
					fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainAsyncMainProgramError, finalErr)
					os.Exit(1)
				}

				if finalPyObj, ok := finalValInterface.(object.Object); ok {
					if finalPyObj.Type() != object.NULL_OBJ {
						fmt.Println(constants.CmdInterpreterMainAsyncMainProgramResult, finalPyObj.Inspect())
					}
				} else if finalValInterface != nil {
					fmt.Println(constants.CmdInterpreterMainAsyncMainProgramReturnedNonPylearn, finalValInterface)
				}
				os.Exit(0)

			} else {
				fmt.Println(constants.CmdInterpreterMainFoundMainProgramNotAsync)
			}
		} else {
			fmt.Printf(constants.CmdInterpreterMainFoundMainProgramNotFunction, mainFuncObj.Type())
		}
	}

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		errObj := evaluated.(*object.Error)
		fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainTracebackHeader)
		if errObj.Line > 0 {
			fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainFileAndLineFormat, scriptName, errObj.Line)
		}
		fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainErrorMessageFormat, errObj.Message)
		if interpreter.PylearnAsyncRuntime != nil && interpreter.PylearnAsyncRuntime.EventLoop != nil {
			interpreter.PylearnAsyncRuntime.EventLoop.Stop()
		}
		os.Exit(1)
	}
	if interpreter.PylearnAsyncRuntime != nil && interpreter.PylearnAsyncRuntime.EventLoop != nil {
		interpreter.PylearnAsyncRuntime.EventLoop.Stop()
	}
	os.Exit(0)
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, constants.CmdInterpreterMainParserErrorsEncountered)
	for _, msg := range errors {
		fmt.Fprintf(out, constants.CmdInterpreterMainParserErrorFormat, msg)
	}
}