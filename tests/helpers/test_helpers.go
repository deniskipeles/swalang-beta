package helpers

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/interpreter"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
	"github.com/deniskipeles/pylearn/internal/stdlib/pysys"
)

// --- Central Evaluation Helper ---

type EvalOptions struct {
	Args []string
	NoBuiltins bool
	NoScriptDirContext bool
	EnvSetup func(env *object.Environment)
}

func Eval(t *testing.T, input string, opts ...EvalOptions) object.Object {
	t.Helper()

	var options EvalOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	parserErrors := p.Errors()
	if len(parserErrors) != 0 {
		errorMsg := fmt.Sprintf("Parser Errors for input:\n%s\n", input)
		lDebug := lexer.New(input)
		errorMsg += "Tokens:\n"
		for {
			tok := lDebug.NextToken()
			errorMsg += fmt.Sprintf("  %s\n", tok.String())
			if tok.Type == lexer.EOF || tok.Type == lexer.ILLEGAL {
				break
			}
		}
		errorMsg += "Errors:\n"
		for _, msg := range parserErrors {
			errorMsg += fmt.Sprintf("\t- %s\n", msg)
		}
		t.Fatalf(errorMsg)
	}

	env := object.NewEnvironment()

	if !options.NoBuiltins {
		for name, builtin := range builtins.Builtins {
			env.Set(name, builtin)
		}
	}

	var pylearnArgv *object.List
	if options.Args != nil {
		pylearnArgObjs := make([]object.Object, len(options.Args))
		for i, arg := range options.Args {
			pylearnArgObjs[i] = &object.String{Value: arg}
		}
		pylearnArgv = &object.List{Elements: pylearnArgObjs}
	} else {
		pylearnArgv = &object.List{Elements: []object.Object{&object.String{Value: "test_script.py"}}}
	}
	pysys.InitializeSysModule(pylearnArgv)

	if !options.NoScriptDirContext {
		wd, err := os.Getwd()
		if err != nil {
			t.Logf("Warning: Could not get working directory for test context: %v", err)
			interpreter.SetCurrentScriptDir(".")
		} else {
			interpreter.SetCurrentScriptDir(wd)
		}
	}

	if options.EnvSetup != nil {
		options.EnvSetup(env)
	}

	mainCtx := interpreter.NewInterpreterContext(env)
	evaluated := interpreter.Eval(program, mainCtx)

	if errObj, ok := evaluated.(*object.Error); ok {
		if strings.HasPrefix(errObj.Message, "TEST_FAIL:") {
			t.Fatalf("Runtime Error indicates test failure: %s", errObj.Message)
		}
	}

	return evaluated
}

// --- Central Assertion Helpers ---

func TestIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if result.Value != expected {
		t.Errorf("Integer has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func TestFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	t.Helper()
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if result.Value != expected {
		t.Errorf("Float has wrong value. got=%g, want=%g", result.Value, expected)
		return false
	}
	return true
}

func TestStringObject(t *testing.T, obj object.Object, expected string) bool {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if result.Value != expected {
		t.Errorf("String has wrong value. got=%q, want=%q", result.Value, expected)
		return false
	}
	return true
}

func TestBytesObject(t *testing.T, obj object.Object, expected []byte) bool {
	t.Helper()
	result, ok := obj.(*object.Bytes)
	if !ok {
		t.Errorf("object is not Bytes. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if !bytes.Equal(result.Value, expected) {
		t.Errorf("Bytes has wrong value. got=%s (hex: %x), want=b'%x'", result.Inspect(), result.Value, expected)
		return false
	}
	return true
}

func TestBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	t.Helper()
	var expectedObj object.Object = object.FALSE
	if expected {
		expectedObj = object.TRUE
	}
	if obj != expectedObj {
		t.Errorf("object is not the expected Boolean singleton. got=%s (%p), want=%s (%p)", obj.Inspect(), obj, expectedObj.Inspect(), expectedObj)
		return false
	}
	return true
}

func TestNullObject(t *testing.T, obj object.Object) bool {
	t.Helper()
	if obj != object.NULL {
		t.Errorf("object is not NULL. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	return true
}

type ErrorInterface interface {
	object.Object
	GetMessage() string
}

func TestErrorObject(t *testing.T, obj object.Object, expectedMsgParts ...string) bool {
	t.Helper()
	errObj, ok := obj.(ErrorInterface)
	if !ok {
		t.Errorf("object is not an Error type. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	errMsg := errObj.GetMessage()
	// Also check Inspect() which usually contains the Error Class Name
	inspectMsg := obj.Inspect()

	for _, part := range expectedMsgParts {
		if !strings.Contains(errMsg, part) && !strings.Contains(inspectMsg, part) {
			t.Errorf("Error %q (inspect: %q) does not contain expected part %q", errMsg, inspectMsg, part)
			return false
		}
	}
	return true
}

func TestStopIteration(t *testing.T, obj object.Object) bool {
	t.Helper()
	if obj.Type() != object.STOP_ITER_OBJ {
		t.Errorf("object is not StopIteration. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	return true
}

func TestListObject(t *testing.T, obj object.Object, expectedElements []interface{}) bool {
	t.Helper()
	list, ok := obj.(*object.List)
	if !ok {
		t.Errorf("object is not List. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if len(list.Elements) != len(expectedElements) {
		t.Errorf("List has wrong number of elements. got=%d, want=%d", len(list.Elements), len(expectedElements))
		t.Logf("Got list: %s", list.Inspect())
		return false
	}
	for i, expectedElem := range expectedElements {
		if !TestObjectLiteral(t, list.Elements[i], expectedElem) {
			t.Logf("Mismatch at index %d of List", i)
			return false
		}
	}
	return true
}

func TestTupleObject(t *testing.T, obj object.Object, expectedElements []interface{}) bool {
	t.Helper()
	tuple, ok := obj.(*object.Tuple)
	if !ok {
		t.Errorf("object is not Tuple. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if len(tuple.Elements) != len(expectedElements) {
		t.Errorf("Tuple has wrong number of elements. got=%d, want=%d", len(tuple.Elements), len(expectedElements))
		t.Logf("Got tuple: %s", tuple.Inspect())
		return false
	}
	for i, expectedElem := range expectedElements {
		if !TestObjectLiteral(t, tuple.Elements[i], expectedElem) {
			t.Logf("Mismatch at index %d of Tuple", i)
			return false
		}
	}
	return true
}

func TestSetObject(t *testing.T, obj object.Object, expectedElements []interface{}) bool {
	t.Helper()
	set, ok := obj.(*object.Set)
	if !ok {
		t.Errorf("object is not Set. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if len(set.Elements) != len(expectedElements) {
		t.Errorf("Set has wrong number of elements. got=%d, want=%d", len(set.Elements), len(expectedElements))
		t.Logf("Got set: %s", set.Inspect())
		return false
	}

	expectedMap := make(map[object.HashKey]bool)
	expectedValues := make(map[object.HashKey]interface{})

	for _, expElem := range expectedElements {
		tempObj := goLiteralToPylearnObject(t, expElem)
		if tempObj == nil {
			return false
		}
		hashableTemp, ok := tempObj.(object.Hashable)
		if !ok {
			t.Errorf("Expected element %v is not hashable (%T)", expElem, tempObj)
			return false
		}
		hKey, err := hashableTemp.HashKey()
		if err != nil {
			t.Errorf("Failed to hash expected element %v: %v", expElem, err)
			return false
		}
		expectedMap[hKey] = true
		expectedValues[hKey] = expElem
	}

	foundKeys := make(map[object.HashKey]bool)
	for hKey, actualElem := range set.Elements {
		if !expectedMap[hKey] {
			t.Errorf("Set contains unexpected element: %s", actualElem.Inspect())
			return false
		}
		foundKeys[hKey] = true
	}

	return true
}

func TestObjectLiteral(t *testing.T, obj object.Object, expected interface{}) bool {
	t.Helper()
	switch exp := expected.(type) {
	case int:
		return TestIntegerObject(t, obj, int64(exp))
	case int64:
		return TestIntegerObject(t, obj, exp)
	case float64:
		return TestFloatObject(t, obj, exp)
	case string:
		return TestStringObject(t, obj, exp)
	case bool:
		return TestBooleanObject(t, obj, exp)
	case nil:
		return TestNullObject(t, obj)
	case []byte:
		return TestBytesObject(t, obj, exp)
	case []interface{}:
		if _, ok := obj.(*object.List); ok {
			return TestListObject(t, obj, exp)
		}
		if _, ok := obj.(*object.Tuple); ok {
			return TestTupleObject(t, obj, exp)
		}
		t.Errorf("Expected a Go []interface{} but got neither List nor Tuple. Got %T: %s", obj, obj.Inspect())
		return false
	default:
		t.Errorf("Unsupported literal type for comparison: %T (%v)", expected, expected)
		return false
	}
}

func goLiteralToPylearnObject(t *testing.T, literal interface{}) object.Object {
	t.Helper()
	switch v := literal.(type) {
	case int:
		return &object.Integer{Value: int64(v)}
	case int64:
		return &object.Integer{Value: v}
	case float64:
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case bool:
		return object.NativeBoolToBooleanObject(v)
	case nil:
		return object.NULL
	case []byte:
		return &object.Bytes{Value: v}
	default:
		t.Errorf("Cannot convert Go literal type %T to Pylearn object", literal)
		return nil
	}
}

func TestClassObject(t *testing.T, obj object.Object, expectedName string) bool {
	t.Helper()
	classObj, ok := obj.(*object.Class)
	if !ok {
		t.Errorf("object is not Class. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if classObj.Name != expectedName {
		t.Errorf("Class has wrong name. got=%q, want=%q", classObj.Name, expectedName)
		return false
	}
	return true
}

func TestInstanceObject(t *testing.T, obj object.Object, expectedClassName string) bool {
	t.Helper()
	instObj, ok := obj.(*object.Instance)
	if !ok {
		t.Errorf("object is not Instance. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	if instObj.Class == nil {
		t.Errorf("Instance object has nil Class field")
		return false
	}
	if instObj.Class.Name != expectedClassName {
		t.Errorf("Instance object has wrong class name. got=%q, want=%q", instObj.Class.Name, expectedClassName)
		return false
	}
	return true
}
