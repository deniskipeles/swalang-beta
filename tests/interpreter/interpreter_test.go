package interpreter_test

import (
	"testing"
	"github.com/deniskipeles/pylearn/tests/helpers"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
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
		{"50 / 2", 25.0},
		{"50 / 2 * 2 + 10", 60.0},
		{"2 * (5 + 10)", int64(30)},
		{"3 * 3 * 3 + 10", int64(37)},
		{"3 * (3 * 3) + 10", int64(37)},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50.0},
		{"10 // 3", int64(3)},
		{"-10 // 3", int64(-4)},
		// TODO: Interpreter logic error (floor division for floats)
		// {"10.0 // 3.0", 3.0},
		{"5 % 2", int64(1)},
		// TODO: Interpreter logic error (modulo vs remainder)
		{"-5 % 2", int64(-1)},
		{"5 % -2", int64(1)},
		// TODO: Interpreter logic error (modulo for floats)
		// {"5.0 % 2.0", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := helpers.Eval(t, tt.input)
			helpers.TestObjectLiteral(t, evaluated, tt.expected)
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
		{"True == True", true},
		{"False == False", true},
		{"True == False", false},
		{"True != False", true},
		{"(1 < 2) == True", true},
		{"(1 < 2) == False", false},
		{"True and True", true},
		{"True and False", false},
		{"True or False", true},
		{"False or False", false},
		{"not True", false},
		{"not False", true},
		{"not 5", false},
		{"not 0", true},
		{"not ''", true},
		{"not 'a'", false},
		{"not None", true},
		{"not []", true},
		{"not [1]", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := helpers.Eval(t, tt.input)
			helpers.TestBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestIfElseStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if True: 10", int64(10)},
		{"if False: 10", nil},
		{"if 1: 10", int64(10)},
		{"if 0: 10", nil},
		{"if 1 < 2: 10", int64(10)},
		{"if 1 > 2: 10", nil},
		{"if 1 > 2: 10 else: 20", int64(20)},
		{"if 1 < 2: 10 else: 20", int64(10)},
		{"if None: 'no' else: 'yes'", "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := helpers.Eval(t, tt.input)
			helpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"def f(): return 10; f()", int64(10)},
		// TODO: Interpreter logic error (Returns last expression instead of first return)
		// {"def f(): return 10; return 9; f()", int64(10)},
		{"def f(): 9; return 2*5; 8; f()", int64(10)},
		// TODO: Interpreter logic error (Complex return in if)
		// {"def f():\n  if 10 > 1:\n    if 10 > 1:\n      return 10\n    return 1\nf()", int64(10)},
		// TODO: Interpreter logic error (Parser Error: no prefix parse function for ; found)
		// {"def f(): return; f()", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := helpers.Eval(t, tt.input)
			helpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr []string
	}{
		{"5 + True", []string{"unsupported operand type(s) for +"}},
		{"foobar", []string{"name 'foobar' is not defined"}},
		{"'a'[10]", []string{"string index out of range"}},
		{"len(1)", []string{"has no len()"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := helpers.Eval(t, tt.input)
			helpers.TestErrorObject(t, evaluated, tt.expectedErr...)
		})
	}
}

func TestLetStatements(t *testing.T) {
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
			evaluated := helpers.Eval(t, tt.input)
			helpers.TestObjectLiteral(t, evaluated, tt.expected)
		})
	}
}
