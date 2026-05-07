//go:build en
// pylearn/internal/constants/all.go
package constants

// swalang/cmd/interpreter/main.go
const (
	CmdInterpreterMainStdin                       = "<stdin>"
	CmdInterpreterMainUsage                       = "Usage: %s <filename.py> [args...]\n"
	CmdInterpreterMainErrorReadingFile            = "Error reading file '%s': %v\n"
	CmdInterpreterMainParserErrorsEncountered    = "Parser errors encountered:"
	CmdInterpreterMainParserErrorFormat          = "\t%s\n"
	CmdInterpreterMainFoundAsyncMainProgram       = "Found async main_program. Executing asynchronously."
	CmdInterpreterMainFatalAsyncRuntimeNotInitialized = "FATAL: Async runtime not initialized!"
	CmdInterpreterMainAsyncFunctionHasNoBody      = "Swalang async function '%s' has no body (nil Body)"
	CmdInterpreterMainAsyncFunctionNilClosureEnv  = "Warning: Swalang async function '%s' has a nil closure environment. Using script global env.\n"
	CmdInterpreterMainPylearnErrorFormat          = "SwalangError: %s (L%d C%d)"
	CmdInterpreterMainAsyncMainProgramError       = "Async main_program error: %v\n"
	CmdInterpreterMainAsyncMainProgramResult      = "Async main_program result:"
	CmdInterpreterMainAsyncMainProgramReturnedNonPylearn = "Async main_program returned non-Swalang Go value:"
	CmdInterpreterMainFoundMainProgramNotAsync    = "Found 'main_program' but it's not async. Script will exit after synchronous evaluation."
	CmdInterpreterMainFoundMainProgramNotFunction = "Found 'main_program' but it's not a Swalang function (type: %s). Script will exit.\n"
	CmdInterpreterMainTracebackHeader             = "Traceback (most recent call last):\n"
	CmdInterpreterMainFileAndLineFormat           = "  File \"%s\", line %d\n"
	CmdInterpreterMainErrorMessageFormat          = "%s\n"
)


// swalang/cmd/repl/main.go
const (
	CmdReplMainPrompt                      = "swalang>>> "
	CmdReplMainWelcomeMessage              = "Welcome to Swalang REPL!"
	CmdReplMainExitMessage                 = "Enter code to evaluate, or press Ctrl+D to exit."
	CmdReplMainErrorReadingInput           = "Error reading input: %v\n"
	CmdReplMainParserErrorsEncountered    = "Parser errors encountered:"
	CmdReplMainParserErrorFormat          = "\t%s\n"
	CmdReplMainRuntimeErrorPrefix          = "Runtime Error: %s\n"
	CmdReplMainGoodbyeMessage              = "\nGoodbye!"
)

