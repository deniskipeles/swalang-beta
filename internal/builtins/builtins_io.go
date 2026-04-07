package builtins

import (
	"bufio" // Needed for input()
	"fmt"
	"io" // Needed for input() EOF check
	"os" // Needed for input()
	"strings"
	"syscall" // Needed for input()

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- print ---
// Accepts ExecutionContext
func pyPrintFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var parts []string
	for _, arg := range args {
		// --- Use str() logic for printing ---
		// Need to call the str builtin via the context
		strBuiltin, ok := Builtins[constants.BuiltinsStrFuncName] // Assumes 'str' is registered
		var strArg object.Object

		if ok {
			// Execute str() using the provided context
			strArg = ctx.Execute(strBuiltin, arg)
		} else {
			// Fallback if str builtin isn't registered for some reason
			strArg = object.NewError(constants.InternalError, constants.BuiltinsIOStrBuiltinNotFound)
		}

		// Check if the result of str() was an error
		if object.IsError(strArg) { // Use object.IsError
			// If str() failed, inspect the original argument as fallback
			// (or inspect the error object itself?)
			parts = append(parts, fmt.Sprintf(constants.BuiltinsIOErrorDuringStr, arg.Inspect(), strArg.Inspect()))
		} else if strVal, isString := strArg.(*object.String); isString {
			parts = append(parts, strVal.Value)
		} else {
			// Should not happen if str() behaves correctly
			parts = append(parts, fmt.Sprintf(constants.BuiltinsIONonStringFromStr, strArg.Inspect()))
		}
	}
	// fmt.Println(strings.Join(parts, " ")) // Print to Go's stdout(TODO:Original to be undone if necessary)
	io.WriteString(outputWriter, strings.Join(parts, constants.Space)+constants.Newline)

	return object.NULL
}

// --- input ---
// Accepts ExecutionContext
func pyInputFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	prompt := ""
	if len(args) > 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsIOInputArgCountError, len(args))
	}
	if len(args) == 1 {
		// Use str() on the prompt argument via context
		strBuiltin, ok := Builtins[constants.BuiltinsStrFuncName]
		var strPrompt object.Object
		if ok {
			strPrompt = ctx.Execute(strBuiltin, args[0])
		} else {
			strPrompt = object.NewError(constants.InternalError, constants.BuiltinsIOStrBuiltinNotFound)
		}

		if object.IsError(strPrompt) { // Use object.IsError
			// Error during str(prompt)
			return object.NewError(constants.TypeError, constants.BuiltinsIOInputArgConvertError, strPrompt.Inspect())
		}
		if strVal, isString := strPrompt.(*object.String); isString {
			prompt = strVal.Value
		} else {
			// Should not happen
			return object.NewError(constants.InternalError, constants.BuiltinsIOInputStrNotReturned)
		}
	}

	// Read from Go's stdin
	fmt.Print(prompt)
	Reader := bufio.NewReader(os.Stdin)
	line, err := Reader.ReadString(constants.NewlineRune)

	if err != nil {
		if err == io.EOF {
			// Use object.NewError
			return object.NewError(constants.EOFError, constants.BuiltinsIOEOFReadingLine)
		}
		// Use object.NewError
		return object.NewError(constants.OSError, constants.ErrorFormat, err.Error())
	}

	// Trim trailing newline characters (\r\n for Windows, \n for Unix)
	line = strings.TrimRight(line, constants.WindowsNewline)
	return &object.String{Value: line}
}

func pyOpenFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 { // Simplified: file, mode
		return object.NewError(constants.TypeError, constants.BuiltinsIOOpenArgCountError, len(args))
	}

	filePath := ""
	switch pathArg := args[0].(type) {
	case *object.String:
		filePath = pathArg.Value
	case *object.Bytes:
		filePath = string(pathArg.Value)
	default:
		return object.NewError(constants.TypeError, constants.BuiltinsIOOpenFileArgTypeError, args[0].Type())
	}

	modeStr := "r" // Python's default mode is 'r' (which implies 'rt')
	if len(args) == 2 {
		modeObj, ok := args[1].(*object.String)
		if !ok {
			return object.NewError(constants.TypeError, constants.BuiltinsIOOpenModeArgTypeError, args[1].Type())
		}
		modeStr = modeObj.Value
	}

	// --- START: New, Robust Mode Parsing Logic ---

	var baseMode rune
	var isBinary, hasPlus bool
	var goFlags int
	permissions := os.FileMode(0666) // Default for creation

	// Check for the primary mode (r, w, a, x). It must be the first character.
	if len(modeStr) > 0 {
		baseMode = rune(modeStr[0])
	} else {
		return object.NewError(constants.ValueError, constants.BuiltinsIOOpenInvalidMode, modeStr)
	}

	// Check for other characters (+, b, t) in the rest of the string.
	otherChars := ""
	if len(modeStr) > 1 {
		otherChars = modeStr[1:]
	}

	if strings.Contains(otherChars, "b") {
		isBinary = true
	}
	if strings.Contains(otherChars, "t") {
		if isBinary { // 'bt' or 'tb' is invalid
			return object.NewError(constants.ValueError, constants.BuiltinsIOOpenTextAndBinaryModeError)
		}
	}
	if strings.Contains(otherChars, "+") {
		hasPlus = true
	}

	// Set the Go os flags based on the parsed components.
	switch baseMode {
	case 'r':
		if hasPlus {
			goFlags = os.O_RDWR // 'r+' or 'rb+'
		} else {
			goFlags = os.O_RDONLY // 'r' or 'rb'
		}
	case 'w':
		if hasPlus {
			goFlags = os.O_RDWR | os.O_CREATE | os.O_TRUNC // 'w+' or 'wb+'
		} else {
			goFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC // 'w' or 'wb'
		}
	case 'a':
		if hasPlus {
			goFlags = os.O_RDWR | os.O_CREATE | os.O_APPEND // 'a+' or 'ab+'
		} else {
			goFlags = os.O_WRONLY | os.O_CREATE | os.O_APPEND // 'a' or 'ab'
		}
	case 'x':
		if hasPlus {
			goFlags = os.O_RDWR | os.O_CREATE | os.O_EXCL // 'x+' or 'xb+'
		} else {
			goFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL // 'x' or 'xb'
		}
	default:
		return object.NewError(constants.ValueError, constants.BuiltinsIOOpenInvalidMode, modeStr)
	}

	// --- END: New, Robust Mode Parsing Logic ---

	goFile, err := os.OpenFile(filePath, goFlags, permissions)
	if err != nil {
		return object.NewError(constants.OSError, constants.BuiltinsIOOpenErrorFormat, getErrno(err), err.Error(), filePath)
	}

	var bReader *bufio.Reader
	if !isBinary { // A bufio.Reader is used for text mode operations like readline.
		bReader = bufio.NewReader(goFile)
	}

	fileObj := &object.File{
		File:     goFile,
		Name:     filePath,
		Mode:     modeStr,
		IsBinary: isBinary,
		Reader:   bReader,
		Closed:   false,
	}
	return fileObj
}

// --- Registration (Continued) ---
func init() {
	// Register functions matching the new signature
	registerBuiltin(constants.BuiltinsPrintFuncName, &object.Builtin{Fn: pyPrintFn})
	registerBuiltin(constants.BuiltinsInputFuncName, &object.Builtin{Fn: pyInputFn})
	registerBuiltin(constants.BuiltinsOpenFuncName, &object.Builtin{Fn: pyOpenFn})
}

// --- Helper to get a pseudo-errno (basic) ---
// Can be copied from pyos or placed in a shared utility package later
func getErrno(err error) int {
	if e, ok := err.(*os.PathError); ok {
		if errno, ok := e.Err.(syscall.Errno); ok {
			return int(errno)
		}
	}
	if e, ok := err.(*os.LinkError); ok {
		if errno, ok := e.Err.(syscall.Errno); ok {
			return int(errno)
		}
	}
	if errno, ok := err.(syscall.Errno); ok {
		return int(errno)
	}
	if os.IsNotExist(err) {
		return int(syscall.ENOENT)
	}
	if os.IsPermission(err) {
		return int(syscall.EACCES)
	}
	if os.IsExist(err) {
		return int(syscall.EEXIST)
	}
	return 1
}
