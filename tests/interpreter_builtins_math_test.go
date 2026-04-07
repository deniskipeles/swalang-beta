package tests

import (
	"testing"
	// No longer need object here directly usually
	"github.com/deniskipeles/pylearn/internal/testhelpers"
	"fmt"
	"github.com/deniskipeles/pylearn/internal/object"
)

func TestBuiltinAbs(t *testing.T) {
	successTests := []struct {
		input    string
		expected interface{} // int64 or float64
	}{
		{"abs(10)", int64(10)},
		{"abs(-5)", int64(5)},
		{"abs(0)", int64(0)},
		{"abs(10.5)", 10.5},
		{"abs(-5.2)", 5.2},
		{"abs(0.0)", 0.0},
		{"abs(True)", int64(1)},
		{"abs(False)", int64(0)},
	}

	for _, tt := range successTests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}

	errorTests := []struct {
		input       string
		expectedErr []string // Use slice for multiple parts if needed
	}{
		{"abs()", []string{"takes exactly one argument"}},
		{"abs(1, 2)", []string{"takes exactly one argument"}},
		{"abs(None)", []string{"TypeError", "bad operand type for abs(): NULL"}}, // Check type and message
		{"abs('hello')", []string{"TypeError", "bad operand type for abs(): STRING"}},
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinRound(t *testing.T) {
	// Python's round() without ndigits returns int
	// Rounds to nearest even for .5 cases
	tests := []struct {
		input    string
		expected int64
	}{
		{"round(10)", int64(10)},
		{"round(0)", int64(0)},
		{"round(-5)", int64(-5)},
		{"round(10.2)", int64(10)},
		{"round(10.8)", int64(11)},
		{"round(10.5)", int64(10)}, // Rounds to nearest even
		{"round(11.5)", int64(12)}, // Rounds to nearest even
		{"round(-10.5)", int64(-10)},
		{"round(-11.5)", int64(-12)},
		// TODO: Add tests with ndigits when implemented
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			// Use TestObjectLiteral as it handles int64
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}

	// TODO: Add error tests for round()
	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"round()", []string{"takes 1 or 2 arguments"}},
		{"round(1, 2, 3)", []string{"takes 1 or 2 arguments"}},
		{"round(None)", []string{"TypeError", "must be real number, not NULL"}}, // Example error
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinPow(t *testing.T) {
	successTests := []struct {
		input    string
		expected interface{} // int64 or float64
	}{
		{"pow(2, 3)", int64(8)},
		{"pow(2, 0)", int64(1)},
		{"pow(0, 0)", int64(1)}, // Python defines 0**0 as 1
		{"pow(10, 2)", int64(100)},
		{"pow(-2, 2)", int64(4)},
		{"pow(-2, 3)", int64(-8)},
		{"pow(2, -2)", 0.25}, // Negative exponent -> float
		{"pow(4, 0.5)", 2.0},
		{"pow(2.0, 3)", 8.0},
		{"pow(2, 3.0)", 8.0},
		{"pow(2.0, -2.0)", 0.25},
		// TODO: Test pow(x, y, z) modulo variant
		// TODO: Test potential BigInt results if implemented
	}
	for _, tt := range successTests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}

	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"pow()", []string{"takes 2 or 3 arguments"}},
		{"pow(1)", []string{"takes 2 or 3 arguments"}},
		{"pow(1, 2, 3, 4)", []string{"takes 2 or 3 arguments"}},
		{"pow(0, -2)", []string{"ValueError", "0 cannot be raised to a negative power"}}, // Example specific error
		{"pow('a', 2)", []string{"TypeError", "unsupported operand type"}},
		// TODO: Test modulo errors (e.g., non-integer modulus)
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinDivmod(t *testing.T) {
	successTests := []struct {
		input    string
		expected []interface{} // Expecting Tuple elements
	}{
		{"divmod(10, 3)", []interface{}{int64(3), int64(1)}},
		{"divmod(-10, 3)", []interface{}{int64(-4), int64(2)}},
		{"divmod(10, -3)", []interface{}{int64(-4), int64(-2)}},
		{"divmod(-10, -3)", []interface{}{int64(3), int64(-1)}},
		{"divmod(0, 3)", []interface{}{int64(0), int64(0)}},
		{"divmod(10.0, 3.0)", []interface{}{3.0, 1.0}}, // Float divmod
        {"divmod(10, 3.0)", []interface{}{3.0, 1.0}},   // Mixed types
	}

	for _, tt := range successTests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			// Use TestTupleObject directly for clarity
			if !testhelpers.TestTupleObject(t, evaluated, tt.expected) {
				t.Logf("Input was: %q", tt.input) // Context added by helper failure message
			}
		})
	}

	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"divmod(10, 0)", []string{"ZeroDivisionError", "division or modulo by zero"}},
		{"divmod()", []string{"TypeError", "takes exactly two arguments"}},
		{"divmod(1)", []string{"TypeError", "takes exactly two arguments"}},
		{"divmod(1, 2, 3)", []string{"TypeError", "takes exactly two arguments"}},
		{"divmod('a', 3)", []string{"TypeError", "unsupported operand type"}},
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinSum(t *testing.T) {
	successTests := []struct {
		input    string
		expected interface{}
	}{
		{"sum([1, 2, 3])", int64(6)},
		{"sum([])", int64(0)}, // Sum of empty is 0 (default start is 0)
		{"sum([1, 2.5, 3])", 6.5},
		{"sum([10.1, -2.1])", 8.0},
		{"sum((1, 2, 3))", int64(6)},         // Test with tuple
		{"sum([1, 2, 3], 10)", int64(16)},    // Test with start value
		{"sum([1.5, 2.5], 1)", 5.0},          // Float sum with int start
		{"sum([], 10.5)", 10.5},             // Empty with float start
        // {"sum(range(4))", int64(6)}, // Requires range support
	}
	for _, tt := range successTests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}

	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"sum()", []string{"TypeError", "takes at least 1 argument (0 given)"}}, // Adjusted error msg
		{"sum(1)", []string{"TypeError", "object is not iterable"}},
		{"sum(['a', 'b'])", []string{"TypeError", "unsupported operand type(s) for +", "int", "str"}}, // More specific
		{"sum([1, 'a'])", []string{"TypeError", "unsupported operand type(s) for +", "int", "str"}},
		{"sum([1, 2], 'a')", []string{"TypeError", "unsupported operand type(s) for +", "str", "int"}}, // Start arg error
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinMinMax(t *testing.T) {
	successTests := []struct {
		input    string
		expected interface{}
	}{
		// Single iterable arg
		{"min([3, 1, 4, 2])", int64(1)},
		{"max([3, 1, 4, 2])", int64(4)},
		{"min((3.1, 1, 4.0, 2))", int64(1)},   // Tuple arg
		{"max((3.1, 1, 4.0, 2))", 4.0},        // Tuple arg
		{"min([-5, -1, -10])", int64(-10)},
		{"max([-5, -1, -10])", int64(-1)},
		{"min('cba')", "a"},
		{"max('cba')", "c"},                   // String iterable
		{"min(['c', 'b', 'a'])", "a"},
		{"max(['c', 'b', 'a'])", "c"},

		// Multiple args
		{"min(3, 1, 4, 2)", int64(1)},
		{"max(3, 1, 4, 2)", int64(4)},
		{"min(3.1, 1, 4.0, 2)", int64(1)},
		{"max(3.1, 1, 4.0, 2)", 4.0},
		{"min('c', 'b', 'a')", "a"},
		{"max('c', 'b', 'a')", "c"},
        {"min(10)", int64(10)}, // Single argument IS the result
        {"max('z')","z"},       // Single argument IS the result
	}
	for _, tt := range successTests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}

	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"min()", []string{"TypeError", "expected at least 1 argument"}}, // Adjusted message
		{"max()", []string{"TypeError", "expected at least 1 argument"}},
		{"min([])", []string{"ValueError", "arg is an empty sequence"}},   // ValueError for empty seq
		{"max(())", []string{"ValueError", "arg is an empty sequence"}},   // Empty tuple
		{"min(1, 'a')", []string{"TypeError", "'<' not supported between instances of 'int' and 'str'"}}, // Python-like error
		{"max(1, 'a')", []string{"TypeError", "'>' not supported between instances of 'int' and 'str'"}},
		{"min([1, 'a'])", []string{"TypeError", "'<' not supported"}}, // Within list
		{"max([1, 'a'])", []string{"TypeError", "'>' not supported"}},
		{"min(None)", []string{"TypeError", "'NULL' object is not iterable"}}, // Error when single arg isn't iterable and not comparable
        {"max(123)", []string{"TypeError", "'int' object is not iterable"}}, // Check when single arg isn't iterable (but is comparable)
		{"min(1, [1])", []string{"TypeError", "'<' not supported", "int", "list"}}, // Mix iterables and non-iterables if multiple args given
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinBinOctHex(t *testing.T) {
	successTests := []struct {
		input    string
		expected string
	}{
		{"bin(10)", "0b1010"},
		{"oct(10)", "0o12"},
		{"hex(10)", "0xa"},
		{"bin(0)", "0b0"},
		{"oct(0)", "0o0"},
		{"hex(0)", "0x0"},
		{"bin(-10)", "-0b1010"},
		{"oct(-10)", "-0o12"},
		{"hex(-10)", "-0xa"},
		{"hex(255)", "0xff"},
	}
	for _, tt := range successTests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestStringObject(t, evaluated, tt.expected)
		})
	}

	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"bin()", []string{"TypeError", "takes exactly one argument"}},
		{"oct(1, 2)", []string{"TypeError", "takes exactly one argument"}},
		{"hex(None)", []string{"TypeError", "cannot be interpreted as an integer"}}, // Python-like message
		{"bin(10.5)", []string{"TypeError", "cannot be interpreted as an integer"}},
		{"hex('a')", []string{"TypeError", "cannot be interpreted as an integer"}},
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestBuiltinHash(t *testing.T) {
	// Test properties: equal immutables have equal hashes, unhashables raise error.
	equalityTests := []struct {
		input1 string
		input2 string
	}{
		{"hash(10)", "hash(10)"},
		{"hash(-5)", "hash(-5)"},
		{"hash(0)", "hash(0)"},
		{"hash('hello')", "hash('hello')"},
		{"hash('')", "hash('')"},
		{"hash(True)", "hash(1)"},    // hash(True) == hash(1)
		{"hash(False)", "hash(0)"},   // hash(False) == hash(0)
		{"hash(None)", "hash(None)"},
		{"hash(b'abc')", "hash(b'abc')"}, // Bytes
		{"hash(1.0)", "hash(1)"},         // Integer and float with same value
        {"hash(0.0)", "hash(0)"},
		{"hash((1, 2))", "hash((1, 2))"}, // Tuple of hashables
	}
	for _, tt := range equalityTests {
		t.Run(fmt.Sprintf("Equality %s vs %s", tt.input1, tt.input2), func(t *testing.T) {
			eval1 := testhelpers.Eval(t, tt.input1)
			eval2 := testhelpers.Eval(t, tt.input2)
			hash1, ok1 := eval1.(*object.Integer)
			hash2, ok2 := eval2.(*object.Integer)
			if !ok1 || !ok2 {
				t.Fatalf("hash() did not return Integer. Got %T and %T", eval1, eval2)
			}
			if hash1.Value != hash2.Value {
				t.Errorf("Hashes should be equal, but got %d (%s) != %d (%s)",
					hash1.Value, tt.input1, hash2.Value, tt.input2)
			}
		})
	}

	// Test different values likely have different hashes (collision is possible but unlikely for simple cases)
	inequalityTests := []struct {
		input1 string
		input2 string
	}{
		{"hash(1)", "hash(2)"},
		{"hash('a')", "hash('b')"},
		{"hash(0)", "hash('0')"}, // Int vs String
		{"hash(None)", "hash(0)"},
        {"hash((1, 2))", "hash((1, 3))"},
        {"hash(b'a')", "hash('a')"}, // Bytes vs String
	}
	for _, tt := range inequalityTests {
		t.Run(fmt.Sprintf("Inequality %s vs %s", tt.input1, tt.input2), func(t *testing.T) {
			eval1 := testhelpers.Eval(t, tt.input1)
			eval2 := testhelpers.Eval(t, tt.input2)
			hash1, ok1 := eval1.(*object.Integer)
			hash2, ok2 := eval2.(*object.Integer)
			if !ok1 || !ok2 {
				t.Fatalf("hash() did not return Integer. Got %T and %T", eval1, eval2)
			}
			if hash1.Value == hash2.Value {
				// This is just a warning, as collisions are allowed
				t.Logf("Warning: Hashes are equal for different inputs: %d (%s) == %d (%s). This might be a collision.",
					hash1.Value, tt.input1, hash2.Value, tt.input2)
			}
		})
	}


	errorTests := []struct {
		input       string
		expectedErr []string
	}{
		{"hash([])", []string{"TypeError", "unhashable type: 'list'"}}, // Python-like messages
		{"hash({})", []string{"TypeError", "unhashable type: 'dict'"}},
		{"hash(set())", []string{"TypeError", "unhashable type: 'set'"}}, // Requires set support
		{"hash((1, []))", []string{"TypeError", "unhashable type: 'list'"}}, // Unhashable element in tuple
		{"hash()", []string{"TypeError", "takes exactly one argument"}},
		{"hash(1, 2)", []string{"TypeError", "takes exactly one argument"}},
	}
	for _, tt := range errorTests {
		t.Run(tt.input+" (error)", func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}