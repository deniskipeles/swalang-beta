package object

import (
	"bytes"
	"fmt"
	"reflect" // Needed for hashing functions/classes by identity
	"strings"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/asyncruntime"
	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
)

// FunctionParameter holds the name and *evaluated* default value for a function parameter.
type FunctionParameter struct {
	Name         string // Parameter name
	DefaultValue Object // Evaluated default value (can be nil if no default)
}

// Function (Interpreter representation)
type Function struct {
	Name string
	// Parameters []*ast.Identifier
	Parameters    []*FunctionParameter
	VarArgParam   string // Name of *args parameter, empty if none. Stores the IDENTIFIER's value.
	KwArgParam    string // Name of **kwargs parameter, empty if none. Stores the IDENTIFIER's value.
	Body          *ast.BlockStatement
	Env           *Environment
	OriginalClass *Class // NEW: The class this function was defined in (if it's a method)
	IsAMethod     bool   // NEW: Flag to indicate if this function originated as a method
	IsAsync       bool
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		paramStr := p.Name
		if p.DefaultValue != nil {
			paramStr += constants.FUNCTION_INSPECT_PARAM_ASSIGN_OP + p.DefaultValue.Inspect()
		}
		params = append(params, paramStr)
	}
	if f.VarArgParam != constants.EmptyString {
		params = append(params, constants.FUNCTION_INSPECT_VAR_ARG_PREFIX+f.VarArgParam)
	}
	if f.KwArgParam != constants.EmptyString {
		params = append(params, constants.FUNCTION_INSPECT_KW_ARG_PREFIX+f.KwArgParam)
	}

	name := f.Name
	if name == constants.EmptyString {
		name = constants.FUNCTION_INSPECT_LAMBDA_NAME
	}
	out.WriteString(fmt.Sprintf(constants.FUNCTION_INSPECT_DEF_KEYWORD+constants.StringFormat+constants.FUNCTION_INSPECT_PAREN_OPEN, name))
	out.WriteString(strings.Join(params, constants.FUNCTION_INSPECT_PARAMS_SEPARATOR))
	out.WriteString(constants.FUNCTION_INSPECT_PAREN_CLOSE_WITH_BODY)
	return out.String()
}
func (f *Function) HashKey() (HashKey, error) {
	return HashKey{Type: f.Type(), Value: uint64(reflect.ValueOf(f).Pointer())}, nil
}

// GetObjectAttribute makes functions expose certain attributes like __name__.
func (f *Function) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	if name == constants.FUNCTION_GET_ATTR_NAME_ATTR {
		return &String{Value: f.Name}, true
	}
	// In the future, you could add __doc__, __defaults__, etc. here.
	return nil, false
}

var _ Object = (*Function)(nil)
var _ Hashable = (*Function)(nil)
var _ AttributeGetter = (*Function)(nil)

// Builtin
type BuiltinFunction func(ctx ExecutionContext, args ...Object) Object

// type Builtin struct { Name string; Fn BuiltinFunction } // Added Name for better identification/debugging
type Builtin struct {
	Name            string
	Fn              BuiltinFunction
	AcceptsKeywords map[string]bool // Optional: map of keyword names it understands
}

// NewBuiltin is a helper for creating simple built-in functions that don't need access to the ExecutionContext.
// This is particularly useful for creating decorators or other internal wrappers.
func NewBuiltin(name string, fn func(args ...Object) Object) *Builtin {
	return &Builtin{
		Name: name,
		Fn: func(ctx ExecutionContext, args ...Object) Object {
			return fn(args...)
		},
	}
}
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return fmt.Sprintf(constants.BUILTIN_INSPECT_FORMAT, b.Name) }
func (b *Builtin) HashKey() (HashKey, error) {
	return HashKey{Type: b.Type(), Value: uint64(reflect.ValueOf(b.Fn).Pointer())}, nil
}

var _ Object = (*Builtin)(nil)
var _ Hashable = (*Builtin)(nil)

type AsyncResultWrapper struct {
	GoAsyncResult *asyncruntime.AsyncResult // Assuming pyasync is the package name for your Go async runtime
}

func (arw *AsyncResultWrapper) Type() ObjectType { return ASYNC_RESULT_OBJ }
func (arw *AsyncResultWrapper) Inspect() string {
	status := constants.ASYNC_RESULT_PENDING
	if arw.GoAsyncResult.IsReady() {
		_, err := arw.GoAsyncResult.GetResult() // Non-blocking check for error
		if err != nil {
			status = fmt.Sprintf(constants.ASYNC_RESULT_FAILED_FORMAT, err)
		} else {
			status = constants.ASYNC_RESULT_COMPLETED
		}
	}
	return fmt.Sprintf(constants.ASYNC_RESULT_INSPECT_FORMAT, status)
}

var _ Object = (*AsyncResultWrapper)(nil)
