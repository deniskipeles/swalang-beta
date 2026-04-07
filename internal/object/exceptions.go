package object

import (
	"fmt"
	"os"

	// "github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
)

// Error (Base Error Type)
type Error struct {
	Message      string
	Line, Column int
	ErrorClass   *Class // The Pylearn class of the error (e.g., TypeError, ValueError)
	Instance     *Instance
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string {
	className := constants.ErrorKeyword
	if e.ErrorClass != nil {
		className = e.ErrorClass.Name
	}
	return fmt.Sprintf(constants.OBJECT_ERROR_INSPECT_FORMAT, className, e.Message)
}
func (e *Error) Error() string      { return e.Message }
func (e *Error) GetMessage() string { return e.Message }

var _ Object = (*Error)(nil)
var _ error = (*Error)(nil)

var (
	// These will be initialized in init()
	BaseExceptionClass         *Class
	ExceptionClass             *Class
	TypeErrorClass             *Class
	ValueErrorClass            *Class
	NameErrorClass             *Class
	IndexErrorClass            *Class
	KeyErrorClass              *Class
	AttributeErrorClass        *Class
	ZeroDivisionErrorClass     *Class
	AssertionErrorClass        *Class
	ImportErrorClass           *Class
	ModuleNotFoundErrorClass   *Class
	ModuleImportErrorClass     *Class
	OSErrorClass               *Class
	JSONDecodeErrorClass       *Class
	JSONEncodeErrorClass       *Class
	InternalErrorClass         *Class
	InternalServerErrorClass   *Class
	AsyncHTTPHandlerErrorClass *Class

	WebSocketClosedErrorClass *Class
	WebSocketWriteErrorClass  *Class
	HTTPClientErrorClass      *Class
	HTTPServerErrorClass      *Class
	RequestErrorClass         *Class
	TimeoutErrorClass         *Class
	WebSocketReadErrorClass   *Class

	TemplateErrorClass       *Class
	NotImplementedErrorClass *Class
	RuntimeErrorClass        *Class
	EOFErrorClass            *Class
	OverflowErrorClass       *Class
	CancelledErrorClass      *Class
	SyntaxErrorClass         *Class
	StopIterationClass       *Class
)

// Maps names to the class objects for easy lookup in NewError
var BuiltinExceptionClasses = make(map[string]*Class)

// CreateExceptionClass creates a new class object for a built-in exception.
// The methods are placeholder Pylearn functions, as their core logic is handled in Go.
// It now defines native Go implementations for __init__ and __str__.
func CreateExceptionClass(name string, base *Class) *Class {
	methods := make(map[string]Object)

	// A Go-based __init__ that captures all arguments and stores them in `self.args`.
	exceptionInitFn := func(ctx ExecutionContext, args ...Object) Object {
		// args[0] is self, args[1:] are the messages passed to the exception
		if len(args) < 1 {
			return NewError(constants.TypeError, constants.INIT_REQUIRES_SELF_ARGUMENT)
		}
		self, ok := args[0].(*Instance)
		if !ok {
			return NewError(constants.TypeError, constants.INIT_MUST_CALL_ON_AN_EXCEPTION_INSTANCE_NOT_OTHER, args[0].Type())
		}
		if self.Env == nil {
			self.Env = NewEnvironment()
		}
		// Store all message arguments in a tuple at self.args, mimicking Python.
		self.Env.Set("args", &Tuple{Elements: args[1:]})
		return NULL
	}
	methods[constants.DunderInit] = &Builtin{Fn: exceptionInitFn, Name: name + "." + constants.DunderInit}

	// A Go-based __str__ that retrieves the message from `self.args`.
	exceptionStrFn := func(ctx ExecutionContext, args ...Object) Object {
		// args[0] is self
		if len(args) != 1 {
			return NewError(constants.TypeError, constants.STR_TAKES_EXACTLY_ONE_ARGUMENT__SELF)
		}
		self, ok := args[0].(*Instance)
		if !ok {
			return NewError(constants.TypeError, constants.STR_MUST_BE_CALLED_ON_AN_EXCEPTION_INSTANCE_NOT_OTHER, args[0].Type())
		}
		if self.Env == nil {
			return NewString("") // No env means no message stored
		}

		argsTupleObj, found := self.Env.Get("args")
		if !found {
			return NewString("")
		}

		argsTuple, ok := argsTupleObj.(*Tuple)
		if !ok || len(argsTuple.Elements) == 0 {
			return NewString("")
		}

		// If there's only one argument, return its string representation.
		// This matches Python's behavior (e.g., str(Exception("msg")) -> "msg").
		if len(argsTuple.Elements) == 1 {
			firstArg := argsTuple.Elements[0]
			// If the argument was a string, return its raw value.
			if strVal, isStr := firstArg.(*String); isStr {
				return strVal
			}
			// For other types (like int), use their Inspect() representation.
			return NewString(firstArg.Inspect())
		}

		// If there are multiple arguments, return the string representation of the tuple.
		return NewString(argsTuple.Inspect())
	}
	methods[constants.DunderStr] = &Builtin{Fn: exceptionStrFn, Name: name + "." + constants.DunderStr}

	superclasses := []*Class{}
	if base != nil {
		superclasses = append(superclasses, base)
	} else if name != constants.BuiltinsObjectType && ObjectClass != nil {
		superclasses = append(superclasses, ObjectClass)
	}

	classObj := &Class{
		Name:           name,
		Superclasses:   superclasses,
		Methods:        methods,
		ClassVariables: NewEnvironment(),
	}

	mro, err := ComputeMRO(classObj, ObjectClass)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.EXCEPTION_MRO_COMPUTATION_FAILED, name, err)
		os.Exit(1)
	}
	classObj.MRO = mro

	return classObj
}

func init() {
	if ObjectClass == nil {
		ObjectClass = &Class{Name: constants.BuiltinsObjectType, MRO: []*Class{}}
		ObjectClass.MRO = []*Class{ObjectClass}
	}

	// Hierarchy based on Python 3: object -> BaseException -> Exception -> ...
	BaseExceptionClass = CreateExceptionClass(constants.BaseException, ObjectClass)
	ExceptionClass = CreateExceptionClass(constants.Exception, BaseExceptionClass)

	// Standard Errors inherit from Exception
	SyntaxErrorClass = CreateExceptionClass(constants.SyntaxError, ExceptionClass)
	TypeErrorClass = CreateExceptionClass(constants.TypeError, ExceptionClass)
	ValueErrorClass = CreateExceptionClass(constants.ValueError, ExceptionClass)
	NameErrorClass = CreateExceptionClass(constants.NameError, ExceptionClass)
	IndexErrorClass = CreateExceptionClass(constants.IndexError, ExceptionClass)
	KeyErrorClass = CreateExceptionClass(constants.KeyError, IndexErrorClass) // KeyError inherits from IndexError
	AttributeErrorClass = CreateExceptionClass(constants.AttributeError, ExceptionClass)
	ZeroDivisionErrorClass = CreateExceptionClass(constants.ZeroDivisionError, ExceptionClass)
	AssertionErrorClass = CreateExceptionClass(constants.AssertionError, ExceptionClass)

	ImportErrorClass = CreateExceptionClass(constants.ImportError, ExceptionClass)
	ModuleNotFoundErrorClass = CreateExceptionClass(constants.ModuleNotFoundError, ExceptionClass)
	ModuleImportErrorClass = CreateExceptionClass(constants.ModuleImportError, ExceptionClass)
	OSErrorClass = CreateExceptionClass(constants.OSError, ExceptionClass)
	NotImplementedErrorClass = CreateExceptionClass(constants.NotImplementedError, ExceptionClass)
	RuntimeErrorClass = CreateExceptionClass(constants.RuntimeError, ExceptionClass)
	EOFErrorClass = CreateExceptionClass(constants.EOFError, ExceptionClass)
	OverflowErrorClass = CreateExceptionClass(constants.OverflowError, ExceptionClass)
	CancelledErrorClass = CreateExceptionClass(constants.CancelledError, ExceptionClass)

	// websocket and http
	WebSocketClosedErrorClass = CreateExceptionClass(constants.WebSocketClosedError, ExceptionClass)
	WebSocketWriteErrorClass = CreateExceptionClass(constants.WebSocketWriteError, ExceptionClass)
	WebSocketReadErrorClass = CreateExceptionClass(constants.WebSocketReadError, ExceptionClass)
	HTTPClientErrorClass = CreateExceptionClass(constants.HTTPClientError, ExceptionClass)
	HTTPServerErrorClass = CreateExceptionClass(constants.HTTPServerError, ExceptionClass)
	RequestErrorClass = CreateExceptionClass(constants.RequestError, ExceptionClass)
	TimeoutErrorClass = CreateExceptionClass(constants.TimeoutError, ExceptionClass)

	// Custom errors for libraries
	JSONDecodeErrorClass = CreateExceptionClass(constants.JSONDecodeError, ValueErrorClass) // Subclass of ValueError
	JSONEncodeErrorClass = CreateExceptionClass(constants.JSONEncodeError, ValueErrorClass) // Subclass of ValueError
	InternalServerErrorClass = CreateExceptionClass(constants.InternalServerError, ExceptionClass)
	AsyncHTTPHandlerErrorClass = CreateExceptionClass(constants.AsyncHTTPHandlerError, ExceptionClass)
	TemplateErrorClass = CreateExceptionClass(constants.TemplateError, ExceptionClass)

	// Internal Error
	InternalErrorClass = CreateExceptionClass(constants.InternalError, ExceptionClass)
	// StopIteration is special, inherits directly from BaseException in Python 3
	StopIterationClass = CreateExceptionClass(constants.StopIteration, BaseExceptionClass)

	// Populate the lookup map
	BuiltinExceptionClasses[BaseExceptionClass.Name] = BaseExceptionClass
	BuiltinExceptionClasses[ExceptionClass.Name] = ExceptionClass
	BuiltinExceptionClasses[SyntaxErrorClass.Name] = SyntaxErrorClass
	BuiltinExceptionClasses[TypeErrorClass.Name] = TypeErrorClass
	BuiltinExceptionClasses[ValueErrorClass.Name] = ValueErrorClass
	BuiltinExceptionClasses[NameErrorClass.Name] = NameErrorClass
	BuiltinExceptionClasses[IndexErrorClass.Name] = IndexErrorClass
	BuiltinExceptionClasses[KeyErrorClass.Name] = KeyErrorClass
	BuiltinExceptionClasses[AttributeErrorClass.Name] = AttributeErrorClass
	BuiltinExceptionClasses[ZeroDivisionErrorClass.Name] = ZeroDivisionErrorClass
	BuiltinExceptionClasses[AssertionErrorClass.Name] = AssertionErrorClass
	BuiltinExceptionClasses[ImportErrorClass.Name] = ImportErrorClass
	BuiltinExceptionClasses[ModuleNotFoundErrorClass.Name] = ModuleNotFoundErrorClass
	BuiltinExceptionClasses[ModuleImportErrorClass.Name] = ModuleImportErrorClass
	BuiltinExceptionClasses[OSErrorClass.Name] = OSErrorClass
	BuiltinExceptionClasses[NotImplementedErrorClass.Name] = NotImplementedErrorClass
	BuiltinExceptionClasses[RuntimeErrorClass.Name] = RuntimeErrorClass
	BuiltinExceptionClasses[EOFErrorClass.Name] = EOFErrorClass
	BuiltinExceptionClasses[OverflowErrorClass.Name] = OverflowErrorClass
	BuiltinExceptionClasses[CancelledErrorClass.Name] = CancelledErrorClass
	BuiltinExceptionClasses[JSONDecodeErrorClass.Name] = JSONDecodeErrorClass
	BuiltinExceptionClasses[JSONEncodeErrorClass.Name] = JSONEncodeErrorClass
	BuiltinExceptionClasses[InternalServerErrorClass.Name] = InternalServerErrorClass
	BuiltinExceptionClasses[AsyncHTTPHandlerErrorClass.Name] = AsyncHTTPHandlerErrorClass
	BuiltinExceptionClasses[InternalErrorClass.Name] = InternalErrorClass
	BuiltinExceptionClasses[TemplateErrorClass.Name] = TemplateErrorClass
	BuiltinExceptionClasses[StopIterationClass.Name] = StopIterationClass

	// Update the singleton STOP_ITERATION error object with its class
	STOP_ITERATION.ErrorClass = StopIterationClass
}
