// pylearn/internal/stdlib/pymath/math_funcs.go
package pymath

import (
	"math"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- Helper function to extract a float from a Pylearn object ---
func getFloat(arg object.Object) (float64, bool) {
	switch v := arg.(type) {
	case *object.Integer:
		return float64(v.Value), true
	case *object.Float:
		return v.Value, true
	default:
		return 0, false
	}
}

// --- Trigonometric Functions ---

func pyMathSin(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "sin() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "sin() argument must be a number")
	}
	return &object.Float{Value: math.Sin(x)}
}

func pyMathCos(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "cos() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "cos() argument must be a number")
	}
	return &object.Float{Value: math.Cos(x)}
}

func pyMathTan(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "tan() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "tan() argument must be a number")
	}
	return &object.Float{Value: math.Tan(x)}
}

// --- Power and Logarithmic Functions ---

func pyMathSqrt(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "sqrt() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "sqrt() argument must be a number")
	}
	if x < 0 {
		return object.NewError(constants.ValueError, "math domain error")
	}
	return &object.Float{Value: math.Sqrt(x)}
}

func pyMathPow(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, "pow() takes exactly 2 arguments (base, exp)")
	}
	base, ok1 := getFloat(args[0])
	exp, ok2 := getFloat(args[1])
	if !ok1 || !ok2 {
		return object.NewError(constants.TypeError, "pow() arguments must be numbers")
	}
	return &object.Float{Value: math.Pow(base, exp)}
}

func pyMathLog(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, "log() takes 1 or 2 arguments")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "log() first argument must be a number")
	}

	if len(args) == 1 {
		return &object.Float{Value: math.Log(x)} // Natural logarithm
	}
	
	base, ok := getFloat(args[1])
	if !ok {
		return object.NewError(constants.TypeError, "log() base must be a number")
	}
	// Log base b of x = log(x) / log(b)
	return &object.Float{Value: math.Log(x) / math.Log(base)}
}

func pyMathLog10(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "log10() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "log10() argument must be a number")
	}
	return &object.Float{Value: math.Log10(x)}
}

// --- Number-theoretic and representation functions ---

func pyMathCeil(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "ceil() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "ceil() argument must be a number")
	}
	// Python's math.ceil returns an integer.
	return &object.Integer{Value: int64(math.Ceil(x))}
}

func pyMathFloor(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "floor() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "floor() argument must be a number")
	}
	// Python's math.floor returns an integer.
	return &object.Integer{Value: int64(math.Floor(x))}
}

func pyMathFabs(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "fabs() takes exactly 1 argument")
	}
	x, ok := getFloat(args[0])
	if !ok {
		return object.NewError(constants.TypeError, "fabs() argument must be a number")
	}
	return &object.Float{Value: math.Abs(x)}
}

// --- Builtin Object Definitions ---
var (
	MathSin   = &object.Builtin{Name: "math.sin", Fn: pyMathSin}
	MathCos   = &object.Builtin{Name: "math.cos", Fn: pyMathCos}
	MathTan   = &object.Builtin{Name: "math.tan", Fn: pyMathTan}
	MathSqrt  = &object.Builtin{Name: "math.sqrt", Fn: pyMathSqrt}
	MathPow   = &object.Builtin{Name: "math.pow", Fn: pyMathPow}
	MathLog   = &object.Builtin{Name: "math.log", Fn: pyMathLog}
	MathLog10 = &object.Builtin{Name: "math.log10", Fn: pyMathLog10}
	MathCeil  = &object.Builtin{Name: "math.ceil", Fn: pyMathCeil}
	MathFloor = &object.Builtin{Name: "math.floor", Fn: pyMathFloor}
	MathFabs  = &object.Builtin{Name: "math.fabs", Fn: pyMathFabs}
)