// cmd/repl/main.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
	"github.com/deniskipeles/pylearn/internal/interpreter"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// --- Create ONE persistent environment for the REPL session ---
	env := object.NewEnvironment()
	// Inject built-ins into the global environment
	for name, builtin := range builtins.Builtins {
		env.Set(name, builtin)
	}
	// ------------------------------------------------------------

	fmt.Println(constants.CmdReplMainWelcomeMessage)
	fmt.Println(constants.CmdReplMainExitMessage)

	for {
		fmt.Fprintf(os.Stderr, constants.CmdReplMainPrompt) // Print prompt to Stderr

		scanned := scanner.Scan()
		if !scanned {
			// EOF (Ctrl+D) detected or error
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, constants.CmdReplMainErrorReadingInput, err)
			}
			break // Exit the loop
		}

		line := scanner.Text()
		if line == "" { // Skip empty lines
			continue
		}

		// --- Process each line ---
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		// Check for parser errors *for this line*
		if len(p.Errors()) != 0 {
			printParserErrors(os.Stderr, p.Errors())
			continue // Go to next prompt iteration
		}

		// Evaluate the AST *using the persistent environment*
		mainCtx := &interpreter.InterpreterContext{Env: env}
		evaluated := interpreter.Eval(program, mainCtx)

		// Check for runtime errors *for this line*
		if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
			errObj := evaluated.(*object.Error)
			// Use Stderr for errors
			fmt.Fprintf(os.Stderr, constants.CmdReplMainRuntimeErrorPrefix, errObj.Message)
			// Line/column info less useful in simple REPL, but could be added
			continue // Go to next prompt iteration
		}

		// Print the result if it's not NULL (mimics Python REPL)
		if evaluated != nil && evaluated.Type() != object.NULL_OBJ {
			// Use Stdout for results
			fmt.Println(evaluated.Inspect())
		}
		// ------------------------
	}

	fmt.Println(constants.CmdReplMainGoodbyeMessage)
}

// Helper function to display parser errors (can be shared or copied)
func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, constants.CmdReplMainParserErrorsEncountered)
	for _, msg := range errors {
		fmt.Fprintf(out, constants.CmdReplMainParserErrorFormat, msg)
	}
}
