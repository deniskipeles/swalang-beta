// +build js,wasm

package main

import (
	"strings"
	"syscall/js"
	"github.com/deniskipeles/pylearn/internal/interpreter"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/parser"
	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/stdlib/pysys"
)

func run(this js.Value, args []js.Value) any {
	source := args[0].String()
	filename := args[1].String()

	var output strings.Builder
	argvList := &object.List{Elements: []object.Object{
		&object.String{Value: filename},
	}}
	pysys.InitializeSysModule(argvList)
	interpreter.SetCurrentScriptDir(filename)

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			output.WriteString("SyntaxError: " + err + "\n")
		}
		return js.ValueOf(output.String())
	}

	env := object.NewEnvironment()
	for name, builtin := range builtins.Builtins {
		env.Set(name, builtin)
	}

	// Capture output manually
	builtins.SetOutput(&output)

	// result := interpreter.Eval(program, env)
	// <<< FIX: Create the initial execution context for the main script
	mainCtx := &interpreter.InterpreterContext{Env: env}

	// Evaluate the program. This will define functions, including 'main_program'.
	result := interpreter.Eval(program, mainCtx) // <<< FIX: Pass mainCtx

	if errObj, ok := result.(*object.Error); ok {
		output.WriteString("RuntimeError: " + errObj.Message + "\n")
		return js.ValueOf(output.String())
	}

	// Optional: Show final result
	if result != nil && result.Type() != object.NULL_OBJ {
		output.WriteString(result.Inspect() + "\n")
	}

	return js.ValueOf(output.String())
}


func main() {
	js.Global().Set("runSwalang", js.FuncOf(run))
	select {} // Keep running
}
