package tests

import (
	"runtime" // Keep for platform check
	"testing"

	"github.com/deniskipeles/pylearn/internal/object" // Keep for direct object checks if needed
	"github.com/deniskipeles/pylearn/internal/testhelpers"

	// Ensure stdlib modules are registered
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pysys"
)

// No need for testEvalStdSyslib or duplicated assertion helpers

func TestSysBuiltinModule(t *testing.T) {

	t.Run("sys.argv", func(t *testing.T) {
		simArgs := []string{"program.py", "--mode=fast", "input.txt"}
		opts := testhelpers.EvalOptions{Args: simArgs}
		input := `import sys; sys.argv` // Evaluate the list itself

		evaluated := testhelpers.Eval(t, input, opts)

		// Use the centralized list tester
		expectedArgv := make([]interface{}, len(simArgs))
		for i, v := range simArgs {
			expectedArgv[i] = v
		}
		testhelpers.TestListObject(t, evaluated, expectedArgv)
	})

	t.Run("sys.argv default", func(t *testing.T) {
		// Test default args when EvalOptions.Args is nil (uses "test_script.py")
		inputLen := `import sys; len(sys.argv)`
		evalLen := testhelpers.Eval(t, inputLen) // No options provided
		testhelpers.TestIntegerObject(t, evalLen, 1)

		inputName := `import sys; sys.argv[0]`
		evalName := testhelpers.Eval(t, inputName) // No options provided
		testhelpers.TestStringObject(t, evalName, "test_script.py") // Default from helper

		// Test empty args (explicitly empty slice, only script name remains)
		optsEmpty := testhelpers.EvalOptions{Args: []string{"script_only.py"}} // Provide only script name
		evalEmpty := testhelpers.Eval(t, inputLen, optsEmpty)
		testhelpers.TestIntegerObject(t, evalEmpty, 1)
		evalEmptyName := testhelpers.Eval(t, inputName, optsEmpty)
		testhelpers.TestStringObject(t, evalEmptyName, "script_only.py")

	})

	t.Run("sys.platform", func(t *testing.T) {
		input := `import sys; sys.platform`
		evaluated := testhelpers.Eval(t, input)
		expectedPlatform := runtime.GOOS // Check against Go runtime value
		testhelpers.TestStringObject(t, evaluated, expectedPlatform)
	})

	// sys.exit tests remain tricky. Testing error cases is reliable.
	// Testing success cases relies on the interpreter *not* exiting the test process.
	t.Run("sys.exit error cases", func(t *testing.T) {
		errTests := []struct{ input string; errParts []string }{
			{`import sys; sys.exit(1, 2)`, []string{"TypeError", "takes at most 1 argument", "2 given"}},
			{`import sys; sys.exit("die")`, []string{"TypeError", "argument must be an integer", "got str"}},
			{`import sys; sys.exit(None)`, []string{"TypeError", "argument must be an integer", "got NULL"}},
		}
        for _, et := range errTests {
            t.Run(et.input+" (error)", func(t *testing.T){
                evalErr := testhelpers.Eval(t, et.input)
				testhelpers.TestErrorObject(t, evalErr, et.errParts...)
            })
        }
	})

	t.Run("sys.exit valid calls (check for no error)", func(t *testing.T) {
		// Check that valid calls do not return a Pylearn error object.
		// The Eval helper itself doesn't terminate the Go test process.
		validCalls := []string{
			`import sys; sys.exit()`,
			`import sys; sys.exit(0)`,
			`import sys; sys.exit(1)`,
			`import sys; sys.exit(127)`,
		}
		for _, input := range validCalls {
			t.Run(input, func(t *testing.T) {
				evaluated := testhelpers.Eval(t, input)
				// Check if the result is an Error type
				if _, isErr := evaluated.(*object.Error); isErr {
					t.Errorf("sys.exit() unexpectedly returned an error: %s", evaluated.Inspect())
				}
                // We might expect NULL or a specific ExitSignal object here
                // depending on how sys.exit is implemented internally.
                // For now, just checking it's not a standard Error.
			})
		}
	})
}