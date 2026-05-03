// cmd/interpreter/main.go
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/interpreter"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
	"github.com/deniskipeles/pylearn/internal/package_manager"
	"github.com/deniskipeles/pylearn/internal/stdlib/pyimportlib"
	pysys_init "github.com/deniskipeles/pylearn/internal/stdlib/pysys"

	"github.com/deniskipeles/pylearn/internal/stdlib/ffi3"
)

func main() {
	goArgs := os.Args

	if len(goArgs) > 1 && goArgs[1] == "get" {
		package_manager.HandleGetCommand(goArgs[2:])
		return
	}

	pyimportlib.SetLoadModuleFunc(interpreter.GetPyLoadModuleFromPathFn())

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

	env := object.NewEnvironment()
	for name, builtin := range builtins.Builtins {
		env.Set(name, builtin)
	}
	for name, class := range object.BuiltinExceptionClasses {
		env.Set(name, class)
	}
	env.Set(constants.DunderName, &object.String{Value: constants.DunderMain})

	mainCtx := interpreter.NewInterpreterContext(env)
	ffi3.SetGlobalExecutionContext(mainCtx)

	evaluated := interpreter.Eval(program, mainCtx)

	// --- Asyncio Auto-Bootloader ---
	mainFuncObj, mainFound := env.Get("main_program")
	if mainFound {
		if mainPylFunc, isPylFunc := mainFuncObj.(*object.Function); isPylFunc {
			if mainPylFunc.IsAsync {
				fmt.Println("⚡ Found async main_program. Booting Swalang asyncio engine...")
				
				// Inject the asyncio launch code directly into the environment!
				bootCode := "import asyncio\nasyncio.run(main_program())\n"
				bootL := lexer.New(bootCode)
				bootP := parser.New(bootL)
				bootProg := bootP.ParseProgram()
				
				evalResult := interpreter.Eval(bootProg, mainCtx)
				if object.IsError(evalResult) {
					errObj := evalResult.(*object.Error)
					fmt.Fprintf(os.Stderr, "Asyncio Crash: %s\n", errObj.Message)
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				fmt.Println("Found 'main_program' but it's not async. Script will exit.")
			}
		}
	}

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		errObj := evaluated.(*object.Error)
		fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainTracebackHeader)
		if errObj.Line > 0 {
			fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainFileAndLineFormat, scriptName, errObj.Line)
		}
		fmt.Fprintf(os.Stderr, constants.CmdInterpreterMainErrorMessageFormat, errObj.Message)
		os.Exit(1)
	}
	os.Exit(0)
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, constants.CmdInterpreterMainParserErrorsEncountered)
	for _, msg := range errors {
		fmt.Fprintf(out, constants.CmdInterpreterMainParserErrorFormat, msg)
	}
}