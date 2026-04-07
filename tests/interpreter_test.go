package tests

import (
	"testing"
	"github.com/deniskipeles/pylearn/internal/testhelpers"

	// Ensure builtins are available implicitly via the Eval helper
)

// No need for testEval or duplicated basic assertion helpers


func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{} // Use interface{} to handle potential float results
	}{
		{"5", int64(5)},
		{"10", int64(10)},
		{"-5", int64(-5)},
		{"-10", int64(-10)},
		{"5 + 5 + 5 + 5 - 10", int64(10)},
		{"2 * 2 * 2 * 2 * 2", int64(32)},
		{"-50 + 100 + -50", int64(0)},
		{"5 * 2 + 10", int64(20)},
		{"5 + 2 * 10", int64(25)},
		{"20 + 2 * -10", int64(0)},
		// Division results in float in Python 3 / standard Pylearn
		{"50 / 2", 25.0},
		{"50 / 2 * 2 + 10", 60.0},        // (50 / 2 = 25.0) * 2 + 10 = 50.0 + 10 = 60.0
		{"2 * (5 + 10)", int64(30)},
		{"3 * 3 * 3 + 10", int64(37)},
		{"3 * (3 * 3) + 10", int64(37)},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50.0}, // (5 + 20 + 5.0) * 2 - 10 = 30.0 * 2 - 10 = 60.0 - 10 = 50.0
        {"10 // 3", int64(3)}, // Integer division
        {"-10 // 3", int64(-4)},
        {"10.0 // 3.0", 3.0},
        {"10 // 3.0", 3.0},
        {"5 % 2", int64(1)}, // Modulo
        {"-5 % 2", int64(1)},
        {"5 % -2", int64(-1)},
        {"5.0 % 2.0", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"True", true},
		{"False", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 <= 2", true},
        {"1 >= 2", false},
        {"1 <= 1", true},
        {"1 >= 1", true},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"True == True", true},
		{"False == False", true},
		{"True == False", false},
		{"True != False", true},
		{"False != True", true},
		{"(1 < 2) == True", true},
		{"(1 < 2) == False", false},
		{"(1 > 2) == True", false},
		{"(1 > 2) == False", true},
        // Logical operators
        {"True and True", true},
        {"True and False", false},
        {"False and True", false},
        {"False and False", false},
        {"True or True", true},
        {"True or False", true},
        {"False or True", true},
        {"False or False", false},
        {"not True", false},
		{"not False", true},
		// Truthiness
		{"not 5", false},
		{"not 0", true},
		{"not 1", false},
        {"not -1", false},
		{"not ''", true},
		{"not 'a'", false},
		{"not None", true},
        {"not []", true},
        {"not [1]", false},
        // Short-circuiting (difficult to test value directly without assignment)
        // Example: `False and exit()` shouldn't call exit. How to test?
        // Need functions/side effects to test short-circuiting properly.
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestBangOperator(t *testing.T) {
    // Assuming '!' is NOT a standard Python operator, but maybe implemented?
    // If sticking to Python syntax, use 'not'. If '!' is custom, test it.
    // Let's assume you mean 'not'. If '!' exists, rename this test.
	tests := []struct {
		input    string
		expected bool
	}{
		{"not True", false},
		{"not False", true},
		{"not 1", false},
        {"not 0", true},
        {"not None", true},
        {"not ''", true},
        {"not not True", true}, // not (not True) -> not False -> True
		{"not not False", false}, // not (not False) -> not True -> False
	}

	for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
		    evaluated := testhelpers.Eval(t, tt.input)
		    testhelpers.TestBooleanObject(t, evaluated, tt.expected)
        })
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{} // Can be Integer, Null, etc.
	}{
		{"if True: 10", int64(10)},
		{"if False: 10", nil}, // No else, returns Null
		{"if 1: 10", int64(10)}, // Truthy condition
		{"if 0: 10", nil},       // Falsy condition
		{"if 1 < 2: 10", int64(10)},
		{"if 1 > 2: 10", nil},
		{"if 1 > 2: 10 else: 20", int64(20)},
		{"if 1 < 2: 10 else: 20", int64(10)},
        {"if None: 'no' else: 'yes'", "yes"}, // None is falsy
        {"if '': 'no' else: 'yes'", "yes"}, // Empty string is falsy
        {"if 'a': 'yes' else: 'no'", "yes"}, // Non-empty string is truthy
        // Nested
        {"if True:\n  if False: 1\n  else: 2\nelse: 3", int64(2)},
        {"if False: 1 else: if True: 2 else: 3", int64(2)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}
}


func TestReturnStatements(t *testing.T) {
    // Return needs a function context
	tests := []struct {
		input    string
		expected interface{}
	}{
        {"def f(): return 10; f()", int64(10)},
        {"def f(): return 10; return 9; f()", int64(10)}, // Returns immediately
        {"def f(): 9; return 2*5; 8; f()", int64(10)},
        {"def f():\n  if 10 > 1:\n    if 10 > 1:\n      return 10\n    return 1\n  f()", int64(10)},
        {"def f(): return; f()", nil}, // Bare return gives None
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
        })
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr []string
	}{
		{"5 + True", []string{"TypeError", "unsupported operand type(s) for +", "int", "bool"}},
		{"5 + True; 9", []string{"TypeError", "unsupported operand type(s) for +"}}, // Error stops execution
		{"foobar", []string{"NameError", "name 'foobar' is not defined"}},
		{"len(1)", []string{"TypeError", "object of type 'int' has no len()"}},
		{"'a'[1]", []string{"IndexError", "string index out of range"}},
		{"{}[[]]", []string{"TypeError", "unhashable type: 'list'"}}, // Depends on dict implementation
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testhelpers.Eval(t, tt.input)
			testhelpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}


func TestLetStatements(t *testing.T) {
    // In Python syntax, this is assignment
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"a = 5; a", int64(5)},
		{"a = 5 * 5; a", int64(25)},
		{"a = 5; b = a; b", int64(5)},
		{"a = 5; b = a; c = a + b + 5; c", int64(15)},
        {"a = None; a", nil},
	}

	for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
		    evaluated := testhelpers.Eval(t, tt.input)
		    testhelpers.TestObjectLiteral(t, evaluated, tt.expected)
        })
	}
}


// Add tests for Function Application, Closures, String/List/Dict Literals and Operations etc.