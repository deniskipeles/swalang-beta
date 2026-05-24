//go:build sw

package constants

// cmd/interpreter/main.go
const (
	CmdInterpreterMainStdin                   = "<stdin>"
	CmdInterpreterMainErrorReadingFile        = "Error reading file '%s': %v\n"
	CmdInterpreterMainParserErrorsEncountered = "Parser errors encountered:"
	CmdInterpreterMainParserErrorFormat       = "\t%s\n"
	CmdInterpreterMainTracebackHeader         = "Traceback (most recent call last):\n"
	CmdInterpreterMainFileAndLineFormat       = "  File \"%s\", line %d\n"
	CmdInterpreterMainErrorMessageFormat      = "%s\n"
)

// cmd/interpreter/main.go additions
const (
	CmdInterpreterMainGetCommand              = "get"
	CmdInterpreterMainMainProgramFunc         = "main_program"
	CmdInterpreterMainBootAsyncInfo           = "⚡ Found async main_program. Booting Swalang asyncio engine..."
	CmdInterpreterMainBootCode                = "import asyncio\nasyncio.run(main_program())\n"
	CmdInterpreterMainAsyncCrashErr           = "Asyncio Crash: %s\n"
	CmdInterpreterMainMainProgramNotAsyncWarn = "Found 'main_program' but it's not async. Script will exit."
	CmdInterpreterReplWelcome                 = "Welcome to Swalang REPL!"
	CmdInterpreterReplExitInfo                = "Enter code to evaluate, or press Ctrl+D to exit."
	CmdInterpreterReplPrompt                  = "swalang>>> "
	CmdInterpreterReplReadErr                 = "Error reading input: %v\n"
	CmdInterpreterReplRuntimeErr              = "Runtime Error: %s\n"
	CmdInterpreterReplGoodbye                 = "\nGoodbye!"
)
