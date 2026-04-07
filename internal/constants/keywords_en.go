//go:build en
package constants

// Common Keywords and Special Names
const (
	Self         = "self"
	Internal     = "internal"
	Lambda       = "<lambda>"
	Method       = "<method>"
	NotKeyword   = "not"
	AndKeyword   = "and"
	OrKeyword    = "or"
	IsKeyword    = "is"
	IsNotKeyword = "is not"
	InKeyword    = "in"
	NotInKeyword = "not in"
	NoneKeyword  = "None"
	TrueKeyword  = "True"
	FalseKeyword = "False"
)

// keywords with spaces
const (
	IfKeywordWithSpace          = "if "
	ElseKeywordWithSpace        = "else "
	ElifKeywordWithSpace        = "elif "
	ElseKeywordWithColonNewline = "else:\n"
	ElifKeywordWithColonNewline = "elif:\n"
	ForKeywordWithSpace         = "for "
	WhileKeywordWithSpace       = "while "
	TryKeywordWithSpace         = "try "
	ExceptKeywordWithSpace      = "except "
	FinallyKeywordWithSpace     = "finally "
	WithKeywordWithSpace        = "with "
	YieldKeywordWithSpace       = "yield "
	RaiseKeywordWithSpace       = "raise "
	AssertKeywordWithSpace      = "assert "
	ImportKeywordWithSpace      = "import "
	FromKeywordWithSpace        = "from "
	AsKeywordWithSpace          = "as "
	PassKeywordWithSpace        = "pass "
	ContinueKeywordWithSpace    = "continue "
	BreakKeywordWithSpace       = "break "
	ReturnKeywordWithSpace      = "return "
	GlobalKeywordWithSpace      = "global "
	NonlocalKeywordWithSpace    = "nonlocal "
	DelKeywordWithSpace         = "del "
)

// Dunder (Magic) Method Names
const (
	DunderInit     = "__init__"
	DunderStr      = "__str__"
	DunderRepr     = "__repr__"
	DunderCall     = "__call__"
	DunderLen      = "__len__"
	DunderBool     = "__bool__"
	DunderContains = "__contains__"
	DunderIter     = "__iter__"
	DunderNext     = "__next__"
	DunderEnter    = "__enter__"
	DunderExit     = "__exit__"
	DunderGetAttr  = "__getattr__"
	DunderSetAttr  = "__setattr__"
	DunderDelAttr  = "__delattr__"
	DunderGetItem  = "__getitem__"
	DunderSetItem  = "__setitem__"
	DunderDelItem  = "__delitem__"
	DunderAwait    = "__await__"

	// math
	DunderAbs    = "__abs__"
	DunderDivmod = "__divmod__"
	DunderRound  = "__round__"
	DunderSum    = "__sum__"
	DunderMin    = "__min__"
	DunderMax    = "__max__"

	// Binary Ops
	DunderAdd        = "__add__"
	DunderSub        = "__sub__"
	DunderMul        = "__mul__"
	DunderTrueDiv    = "__truediv__"
	DunderMod        = "__mod__"
	DunderPow        = "__pow__"
	DunderOr         = "__or__"
	DunderAnd        = "__and__"
	DunderXor        = "__xor__"
	DunderIsInstance = "__isinstance__" // Not used yet, but good to have

	// Reflected Binary Ops
	DunderRAdd     = "__radd__"
	DunderRSub     = "__rsub__"
	DunderRMul     = "__rmul__"
	DunderRTrueDiv = "__rtruediv__"
	DunderRMod     = "__rmod__"
	DunderRPow     = "__rpow__"
	DunderROr      = "__ror__"
	DunderRAnd     = "__rand__"
	DunderRXor     = "__rxor__"

	// Comparison Ops
	DunderEq = "__eq__"
	DunderNe = "__ne__"
	DunderLt = "__lt__"
	DunderLe = "__le__"
	DunderGt = "__gt__"
	DunderGe = "__ge__"

	// Unary Ops
	DunderNeg = "__neg__"
	DunderPos = "__pos__"
)

const (
	DunderName   = "__name__"
	DunderMain   = "__main__"
	DunderFile   = "__file__"
	DunderDir    = "__dir__"
	DunderFormat = "__format__"
	DunderDoc    = "__doc__"
	DunderHash   = "__hash__"
)

// Dunder Method Mapping from Operators
var InfixOperatorToDunder = map[string]string{
	"+":  DunderAdd,
	"-":  DunderSub,
	"*":  DunderMul,
	"**": DunderPow,
	"/":  DunderTrueDiv,
	"%":  DunderMod,
	"==": DunderEq,
	"!=": DunderNe,
	"<":  DunderLt,
	"<=": DunderLe,
	">":  DunderGt,
	">=": DunderGe,
	"|":  DunderOr,
	"&":  DunderAnd,
	"^":  DunderXor,
}

var InfixOperatorToRDunder = map[string]string{
	"+":  DunderRAdd,
	"-":  DunderRSub,
	"*":  DunderRMul,
	"**": DunderRPow,
	"/":  DunderRTrueDiv,
	"%":  DunderRMod,
	"|":  DunderROr,
	"&":  DunderRAnd,
	"^":  DunderRXor,
}

var PrefixOperatorToDunder = map[string]string{
	"-":   DunderNeg,
	"+":   DunderPos,
	"not": DunderBool,
}

// Built-in Exception Names
const (
	BaseException         = "BaseException"
	Exception             = "Exception"
	SyntaxError           = "SyntaxError"
	TypeError             = "TypeError"
	InternalError         = "InternalError"
	InternalServerError   = "InternalServerError"
	AsyncHTTPHandlerError = "AsyncHTTPHandlerError"
	ValueError            = "ValueError"
	NameError             = "NameError"
	IndexError            = "IndexError"
	KeyError              = "KeyError"
	AttributeError        = "AttributeError"
	ZeroDivisionError     = "ZeroDivisionError"
	AssertionError        = "AssertionError"
	// ErrIndexOutOfRange     = "ErrIndexOutOfRange"
	// ErrDivByZero     = "ErrDivByZero"
	// ErrNameNotDefined     = "ErrNameNotDefined"
	// ErrKeyNotFound     = "ErrKeyNotFound"
	ImportError         = "ImportError"
	ModuleImportError   = "ModuleImportError"
	ModuleNotFoundError = "ModuleNotFoundError"

	OSError              = "OSError"
	NotImplementedError  = "NotImplementedError"
	RuntimeError         = "RuntimeError"
	EOFError             = "EOFError"
	OverflowError        = "OverflowError"
	CancelledError       = "CancelledError"
	JSONDecodeError      = "JSONDecodeError"
	JSONEncodeError      = "JSONEncodeError"
	TemplateError        = "TemplateError"
	StopIteration        = "StopIteration"
	HTTPClientError      = "HTTPClientError"
	HTTPServerError      = "HTTPServerError"
	RequestError         = "RequestError"
	TimeoutError         = "TimeoutError"
	WebSocketClosedError = "WebSocketClosedError"
	WebSocketWriteError  = "WebSocketWriteError"
	WebSocketReadError   = "WebSocketReadError"
)

// Common Error Message Formats
const (
	ErrWrongNumArgs            = "%s() takes %d positional arguments but %d were given"
	ErrWrongNumArgsAtLeast     = "%s() takes at least %d arguments (%d given)"
	ErrWrongNumArgsBetween     = "%s() takes from %d to %d arguments but %d were given"
	ErrWrongNumArgsExact       = "%s() takes exactly %d arguments (%d given)"
	ErrUnsupportedOperand      = "unsupported operand type(s) for %s: '%s' and '%s'"
	ErrUnsupportedUnaryOperand = "bad operand type for unary %s: '%s'"
	ErrNotIterable             = "'%s' object is not iterable"
	ErrNotSubscriptable        = "'%s' object is not subscriptable"
	ErrNotCallable             = "'%s' object is not callable"
	ErrNoAttribute             = "'%s' object has no attribute '%s'"
	ErrUnhandledAttribute      = "'%s' object has no attribute '%s'"
	ErrUnHashableType          = "unhashable type: '%s'"
	ErrMissingRequiredArg      = "%s() missing 1 required positional argument: '%s'"
	ErrUnexpectedKeywordArg    = "%s() got an unexpected keyword argument '%s'"
	ErrIsinstanceArg2          = "isinstance() arg 2 must be a type or tuple of types, not %s"
	ErrSubclassArg1            = "issubclass() arg 1 must be a class"
	ErrSubclassArg2            = "issubclass() arg 2 must be a class or tuple of classes"
	ErrInternal                = "InternalError: %s"
)
// const (
// 	ErrWrongNumArgs            = "%s() takes %d positional arguments but %d were given"
// 	ErrWrongNumArgsAtLeast     = "%s() takes at least %d arguments (%d given)"
// 	ErrWrongNumArgsBetween     = "%s() takes from %d to %d arguments but %d were given"
// 	ErrWrongNumArgsExact       = "%s() takes exactly %d arguments (%d given)"
// 	ErrUnsupportedOperand      = "unsupported operand type(s) for %s: '%s' and '%s'"
// 	ErrUnsupportedUnaryOperand = "bad operand type for unary %s: '%s'"
// 	ErrNotIterable             = "'%s' object is not iterable"
// 	ErrNotSubscriptable        = "'%s' object is not subscriptable"
// 	ErrNotCallable             = "'%s' object is not callable"
// 	ErrNoAttribute             = "'%s' object has no attribute '%s'"
// 	ErrUnhandledAttribute      = "'%s' object has no attribute '%s'"
// 	ErrUnHashableType          = "unhashable type: '%s'"
// 	ErrMissingRequiredArg      = "%s() missing 1 required positional argument: '%s'"
// 	ErrUnexpectedKeywordArg    = "%s() got an unexpected keyword argument '%s'"
// 	ErrIsinstanceArg2          = "isinstance() arg 2 must be a type or tuple of types, not %s"
// 	ErrSubclassArg1            = "issubclass() arg 1 must be a class"
// 	ErrSubclassArg2            = "issubclass() arg 2 must be a class or tuple of classes"
// 	ErrInternal                = "InternalError: %s"
// )
