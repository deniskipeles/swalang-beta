// pylearn/internal/object/file_object.go

package object

import (
	"bufio" // For readline, readlines
	"fmt"
	"io" // For io.Reader, io.Writer, io.Closer, io.EOF
	"os"
	"strings"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// File (Enhanced)
type File struct {
    mu       sync.Mutex // For thread-safe operations on the file
    File     *os.File   // The underlying Go file
    Name     string     // Path used to open the file
    Mode     string     // Mode string (e.g., "r", "wb")
    IsBinary bool       // True if opened in binary mode
    Closed   bool

    // For buffered reading (especially text mode)
    Reader *bufio.Reader // <<< THIS FIELD MUST BE PRESENT
    // For iteration
    iterExhausted bool   // <<< THIS FIELD MUST BE PRESENT
}

func (f *File) Type() ObjectType { return FILE_OBJ }
func (f *File) Inspect() string {
	status := constants.EmptyString
	if f.Closed {
		status = constants.FILE_OBJECT_CLOSED_STATUS
	}
	// Try to get fd for more Pythonic inspect, but os.File.Fd() might not always be useful or safe
	// fdStr := constants.FILE_OBJECT_FD_UNKNOWN_FORMAT
	// if f.File != nil { fdStr = fmt.Sprintf("fd=%d", f.File.Fd()) }
	return fmt.Sprintf(constants.FILE_OBJECT_INSPECT_FORMAT, f.Name, f.Mode, status)
}
var _ Object = (*File)(nil)
// File objects are not hashable
// File objects can be iterable (line by line)

// --- Go functions for File methods ---

func pyFileReadFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (File), args[1] is size (optional, int)
	selfFile, ok := args[0].(*File)
	if !ok { return NewError(constants.TypeError, constants.FILE_OBJECT_READ_ON_FILE_ERROR) }

	selfFile.mu.Lock()
	defer selfFile.mu.Unlock()
	if selfFile.Closed { return NewError(constants.ValueError, constants.FILE_OBJECT_READ_CLOSED_ERROR) }

	size := -1 // Default: read until EOF
	if len(args) == 2 {
		if args[1] != NULL {
			sizeInt, okSize := args[1].(*Integer)
			if !okSize { return NewError(constants.TypeError, constants.FILE_OBJECT_READ_SIZE_TYPE_ERROR) }
			size = int(sizeInt.Value)
		}
	}

	var data []byte
	var err error

	if selfFile.IsBinary {
		if size < 0 { // Read all
			data, err = io.ReadAll(selfFile.File)
		} else {
			data = make([]byte, size)
			n, readErr := selfFile.File.Read(data)
			if readErr != nil && readErr != io.EOF {
				err = readErr
			}
			data = data[:n] // Slice to actual number of bytes read
		}
		if err != nil { return NewError(constants.OSError, constants.FILE_OBJECT_READ_BINARY_ERROR_MSG, err) }
		return &Bytes{Value: data}
	} else { // Text mode
		// Use the buffered Reader
		if selfFile.Reader == nil { // Should have been initialized by open
			return NewError(constants.InternalError, constants.FILE_OBJECT_READER_NOT_INIT_ERROR)
		}
		if size < 0 { // Read all
			data, err = io.ReadAll(selfFile.Reader)
		} else {
			data = make([]byte, size)
			n, readErr := selfFile.Reader.Read(data) // Reads up to 'size' bytes
			if readErr != nil && readErr != io.EOF {
				err = readErr
			}
			data = data[:n]
		}
		if err != nil { return NewError(constants.OSError, constants.FILE_OBJECT_READ_TEXT_ERROR_MSG, err) }
		return &String{Value: string(data)}
	}
}

func pyFileReadLineFn(ctx ExecutionContext, args ...Object) Object {
	selfFile, ok := args[0].(*File)
	if !ok { return NewError(constants.TypeError, constants.FILE_OBJECT_READLINE_ON_FILE_ERROR) }

	selfFile.mu.Lock()
	defer selfFile.mu.Unlock()
	if selfFile.Closed { return NewError(constants.ValueError, constants.FILE_OBJECT_READ_CLOSED_ERROR) }

	// TODO: Handle optional size argument for readline
	if len(args) > 1 {
		return NewError(constants.NotImplementedError, constants.FILE_OBJECT_READLINE_SIZE_NOT_IMPL)
	}

	if selfFile.IsBinary {
		// Read until newline or EOF, keeping newline.
		// This is tricky with os.File directly for binary. bufio.Reader is better.
		// For simplicity, binary readline might be less common or behave differently.
		// Let's use a temporary bufio.Reader for binary readline for now if one isn't persistent.
		tempReader := bufio.NewReader(selfFile.File)
		lineBytes, err := tempReader.ReadBytes(constants.NewlineRune)
		if err != nil && err != io.EOF {
			return NewError(constants.OSError, constants.FILE_OBJECT_READLINE_BINARY_ERROR_MSG, err)
		}
		if len(lineBytes) == 0 && err == io.EOF { // True EOF reached before reading anything
			return &Bytes{Value: []byte{}}
		}
		return &Bytes{Value: lineBytes}
	} else { // Text mode
		if selfFile.Reader == nil {
			return NewError(constants.InternalError, constants.FILE_OBJECT_READER_NOT_INIT_ERROR)
		}
		lineStr, err := selfFile.Reader.ReadString(constants.NewlineRune)
		if err != nil && err != io.EOF {
			return NewError(constants.OSError, constants.FILE_OBJECT_READLINE_TEXT_ERROR_MSG, err)
		}
		if len(lineStr) == 0 && err == io.EOF {
			return &String{Value: constants.EmptyString}
		}
		return &String{Value: lineStr}
	}
}

func pyFileReadLinesFn(ctx ExecutionContext, args ...Object) Object {
	selfFile, ok := args[0].(*File)
	if !ok { return NewError(constants.TypeError, constants.FILE_OBJECT_READLINES_ON_FILE_ERROR) }
	
	// TODO: Handle optional hint argument
	if len(args) > 1 {
		return NewError(constants.NotImplementedError, constants.FILE_OBJECT_READLINES_HINT_NOT_IMPL)
	}
	
	selfFile.mu.Lock()
	defer selfFile.mu.Unlock()
	if selfFile.Closed { return NewError(constants.ValueError, constants.FILE_OBJECT_READ_CLOSED_ERROR) }

	linesList := &List{Elements: []Object{}}
	var Reader io.Reader
	if selfFile.IsBinary {
		Reader = selfFile.File // Or wrap in bufio.NewReader for consistency with readline
	} else {
		if selfFile.Reader == nil { return NewError(constants.InternalError, constants.FILE_OBJECT_READER_NOT_INIT_ERROR) }
		Reader = selfFile.Reader
	}

	// Use a scanner for robust line splitting
	scanner := bufio.NewScanner(Reader)
	for scanner.Scan() {
		// lineText := scanner.Text() // Text without newline
		// Python's readlines keeps the newlines if present (scanner removes them)
		// This is a deviation for simplicity. To match Python fully, we'd need ReadBytes('\n') in a loop.
		// For now, let's add newline back if it wasn't the last line without one.
		// A more robust way: check if the original data had a newline.
		// The current `readline` implementation *does* keep newlines. `readlines` should be consistent.

		// Re-do with ReadString/ReadBytes to match readline behavior
		var lineObject Object
		if selfFile.IsBinary {
			// A bit inefficient to create a new Reader each time for binary.
			// A persistent bufio.Reader for binary files could be an option.
			tempLineReader := bufio.NewReader(selfFile.File) // Assuming selfFile.File is the raw *os.File
			b, errRead := tempLineReader.ReadBytes(constants.NewlineRune)
			if errRead != nil && errRead != io.EOF { return NewError(constants.OSError, constants.FILE_OBJECT_READLINES_BINARY_READ_ERROR, errRead) }
			if len(b) == 0 && errRead == io.EOF { break } // End of file
			lineObject = &Bytes{Value: b}
		} else {
			if selfFile.Reader == nil {return NewError(constants.InternalError, constants.FILE_OBJECT_READLINES_TEXT_READER_NOT_INIT)}
			s, errRead := selfFile.Reader.ReadString(constants.NewlineRune)
			if errRead != nil && errRead != io.EOF { return NewError(constants.OSError, constants.FILE_OBJECT_READLINES_TEXT_READ_ERROR, errRead) }
			if len(s) == 0 && errRead == io.EOF { break } // End of file
			lineObject = &String{Value: s}
		}
		linesList.Elements = append(linesList.Elements, lineObject)
		if selfFile.IsBinary && len(lineObject.(*Bytes).Value) > 0 && lineObject.(*Bytes).Value[len(lineObject.(*Bytes).Value)-1] != constants.NewlineRune{
			// If binary line didn't end with newline and it wasn't EOF, means we hit EOF mid-line
			break
		}
		if !selfFile.IsBinary && len(lineObject.(*String).Value) > 0 && !strings.HasSuffix(lineObject.(*String).Value, constants.Newline) {
			// If text line didn't end with newline and it wasn't EOF, means we hit EOF mid-line
			break
		}

	}
	if err := scanner.Err(); err != nil { // Error during scan itself (rare for basic Reader)
		return NewError(constants.OSError, constants.FILE_OBJECT_READLINES_SCANNER_ERROR_MSG, err)
	}
	return linesList
}


func pyFileWriteFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (File), args[1] is data (String or Bytes)
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.FILE_OBJECT_WRITE_ARG_COUNT_ERROR, len(args)-1)
	}
	selfFile, ok := args[0].(*File)
	if !ok { return NewError(constants.TypeError, constants.FILE_OBJECT_WRITE_ON_FILE_ERROR) }

	selfFile.mu.Lock()
	defer selfFile.mu.Unlock()
	if selfFile.Closed { return NewError(constants.ValueError, constants.FILE_OBJECT_WRITE_CLOSED_ERROR) }

	var bytesToWrite []byte
	switch data := args[1].(type) {
	case *String:
		if selfFile.IsBinary {
			return NewError(constants.TypeError, constants.FILE_OBJECT_WRITE_BINARY_TYPE_ERROR)
		}
		bytesToWrite = []byte(data.Value)
	case *Bytes:
		if !selfFile.IsBinary {
			return NewError(constants.TypeError, constants.FILE_OBJECT_WRITE_TEXT_TYPE_ERROR)
		}
		bytesToWrite = data.Value
	default:
		return NewError(constants.TypeError, constants.FILE_OBJECT_WRITE_ARG_TYPE_ERROR, args[1].Type())
	}

	n, err := selfFile.File.Write(bytesToWrite)
	if err != nil {
		return NewError(constants.OSError, constants.FILE_OBJECT_WRITE_ERROR_MESSAGE, err)
	}
	return &Integer{Value: int64(n)} // Python's write returns number of characters/bytes written
}

func pyFileCloseFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (File)
	if len(args) != 1 {
		return NewError(constants.TypeError, constants.FILE_OBJECT_CLOSE_ARG_COUNT_ERROR, len(args)-1)
	}
	selfFile, ok := args[0].(*File)
	if !ok { return NewError(constants.TypeError, constants.FILE_OBJECT_CLOSE_ON_FILE_ERROR) }

	selfFile.mu.Lock()
	defer selfFile.mu.Unlock()

	if selfFile.Closed {
		return NULL // Closing an already closed file is a no-op, returns None
	}

	err := selfFile.File.Close()
	if err != nil {
		return NewError(constants.OSError, constants.FILE_OBJECT_CLOSING_ERROR, err)
	}
	selfFile.Closed = true
	selfFile.Reader = nil // Clear Reader
	return NULL
}


// pyFileEnterFn implements File.__enter__
func pyFileEnterFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 { // self
		return NewError(constants.TypeError, constants.FILE_OBJECT_ENTER_ARG_COUNT_ERROR)
	}
	selfFile, ok := args[0].(*File)
	if !ok {
		return NewError(constants.TypeError, constants.FILE_OBJECT_ENTER_ON_FILE_ERROR)
	}
	// Typically, __enter__ returns the object to be used with 'as'
	// For files, it's the file object itself.
	// It could also acquire resources here if needed.
	return selfFile
}

// pyFileExitFn implements File.__exit__(self, exc_type, exc_val, exc_tb)
func pyFileExitFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 4 { // self, type, value, traceback
		return NewError(constants.TypeError, constants.FILE_OBJECT_EXIT_ARG_COUNT_ERROR)
	}
	selfFile, ok := args[0].(*File)
	if !ok {
		return NewError(constants.TypeError, constants.FILE_OBJECT_EXIT_ON_FILE_ERROR)
	}

	// excType := args[1] // Pylearn object (e.g., ErrorClass or NULL)
	// excVal  := args[2] // Pylearn object (e.g., Error instance or NULL)
	// excTb   := args[3] // Pylearn object (e.g., Traceback object or NULL) - Traceback object not implemented yet

	// Ensure the file is closed, even if errors occurred.
	// The pyFileCloseFn already handles already-closed files gracefully.
	closeResult := pyFileCloseFn(ctx, selfFile) // Pass selfFile directly
	if IsError(closeResult) {
		// If closing itself fails, this error should probably propagate,
		// potentially masking the original exception if one occurred in the 'with' block.
		// Python's behavior here can be nuanced. For now, let's return the close error.
		return closeResult
	}

	// If __exit__ returns True, it suppresses the exception (if one occurred).
	// If it returns False (or None, which is falsy), the exception is re-raised.
	// For a simple file.close(), we don't suppress exceptions by default.
	return FALSE // Or NULL, which evaluates to False
}



// --- GetObjectAttribute for File ---
func (f *File) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makeFileMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.FileMethodPrefix + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, f) // Prepend self (the File 'f')
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	f.mu.Lock() // Lock for reading attributes like 'closed', 'name', 'mode'
	defer f.mu.Unlock()

	switch name {
	case constants.FileReadMethodName:
		return makeFileMethod(constants.FileReadMethodName, pyFileReadFn), true
	case constants.FileReadLineMethodName:
		return makeFileMethod(constants.FileReadLineMethodName, pyFileReadLineFn), true
	case constants.FileReadLinesMethodName:
		return makeFileMethod(constants.FileReadLinesMethodName, pyFileReadLinesFn), true
	case constants.FileWriteMethodName:
		return makeFileMethod(constants.FileWriteMethodName, pyFileWriteFn), true
	case constants.FileCloseMethodName:
		return makeFileMethod(constants.FileCloseMethodName, pyFileCloseFn), true
	case constants.FileClosedAttributeName:
		return NativeBoolToBooleanObject(f.Closed), true
	case constants.FileNameAttributeName:
		return &String{Value: f.Name}, true
	case constants.FileModeAttributeName:
		return &String{Value: f.Mode}, true
	// TODO: flush, seek, tell, etc.
	// TODO: __enter__, __exit__ for context manager support
	case constants.FileEnterDunderMethodName: // <-- ADD
		return makeFileMethod(constants.FileEnterDunderMethodName, pyFileEnterFn), true
	case constants.FileExitDunderMethodName: // <-- ADD
		return makeFileMethod(constants.FileExitDunderMethodName, pyFileExitFn), true
	default:
		return nil, false
	}
	return nil, false
}
var _ AttributeGetter = (*File)(nil)


// --- Iterator Protocol for File (line by line) ---
func (f *File) Next() (Object, bool) {
	f.mu.Lock()
	if f.Closed || f.iterExhausted {
		f.mu.Unlock()
		return nil, true // Stop iteration
	}

	// If binary, we read bytes until newline. If text, use buffered Reader.
	var line Object
	var err error

	if f.IsBinary {
		// Need a way to read line-by-line for binary. For simplicity, reuse readline logic.
		// This means creating a temp bufio.Reader for each Next() call if not persistent.
		// Or, the Next() itself becomes complex.
		// For now, let's make a simple assumption that iteration over binary files is less common
		// for line-by-line and might be better handled by read().
		// Let's return error for now if iterating a binary file this way.
		// To support it, we'd need a persistent binary line Reader or more complex logic here.
		// f.mu.Unlock()
		// return NewError(constants.NotImplementedError, constants.FILE_OBJECT_ITER_BINARY_NOT_IMPL), true

		// Simpler approach: read one line for binary (less efficient if called repeatedly)
		tempReader := bufio.NewReader(f.File)
		lineBytes, readErr := tempReader.ReadBytes(constants.NewlineRune)
		err = readErr
		if len(lineBytes) > 0 {
			line = &Bytes{Value: lineBytes}
		}

	} else { // Text mode
		if f.Reader == nil {
			f.mu.Unlock()
			return NewError(constants.InternalError, constants.FILE_OBJECT_ITER_TEXT_READER_NOT_INIT), true
		}
		lineStr, readErr := f.Reader.ReadString(constants.NewlineRune)
		err = readErr
		if len(lineStr) > 0 {
			line = &String{Value: lineStr}
		}
	}
	
	f.mu.Unlock() // Unlock before returning

	if err == io.EOF {
		f.mu.Lock()
		f.iterExhausted = true
		f.mu.Unlock()
		if line != nil { // EOF reached but some data was read (line without trailing newline)
			return line, false
		}
		return nil, true // True EOF, no more data
	}
	if err != nil {
		// Propagate other errors as Pylearn errors
		return NewError(constants.OSError, constants.FILE_OBJECT_ITER_ERROR_MESSAGE, err), true
	}
	return line, false
}
var _ Iterator = (*File)(nil) // File implements the Iterator interface