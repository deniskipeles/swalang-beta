// pylearn/internal/stdlib/pymath/module.go
package pymath

import (
	"math"

	"github.com/deniskipeles/pylearn/internal/object"
)

func init() {
	env := object.NewEnvironment()

	// --- Constants ---
	env.Set("pi", &object.Float{Value: math.Pi})
	env.Set("e", &object.Float{Value: math.E})
	env.Set("inf", &object.Float{Value: math.Inf(1)})
	env.Set("nan", &object.Float{Value: math.NaN()})

	// --- Functions ---
	env.Set("sin", MathSin)
	env.Set("cos", MathCos)
	env.Set("tan", MathTan)
	env.Set("sqrt", MathSqrt)
	env.Set("pow", MathPow)
	env.Set("log", MathLog)
	env.Set("log10", MathLog10)
	env.Set("ceil", MathCeil)
	env.Set("floor", MathFloor)
	env.Set("fabs", MathFabs)
	// Add other functions here: acos, asin, atan, degrees, radians, etc.

	mathModule := &object.Module{
		Name: "math",
		Path: "<builtin_math>",
		Env:  env,
	}

	object.RegisterNativeModule("math", mathModule)
}