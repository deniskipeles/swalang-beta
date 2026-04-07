package testhelpers

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
	"github.com/deniskipeles/pylearn/internal/stdlib/pysys" // Needed for EvalWithArgs
)

// --- Central Evaluation Helper ---

// EvalOptions allows specifying configuration for the evaluation helper.
type EvalOptions struct {
	// Args to simulate for sys.argv. If nil, a default is used.
	// Set explicitly to []string{} for empty argv besides script name.
	Args []string
	// If true, do not automatically inject standard builtins.
	NoBuiltins bool
	// If true, do not automatically set a default script directory.
	NoScriptDirContext bool
	// Additional setup to run on the environment before Eval.
	EnvSetup func(env *object.Environment)
}

// Eval performs lexing, parsing, environment setup, and evaluation.
// It handles standard builtins and basic context setup.
func Eval(t *testing.T, input string, opts ...EvalOptions) object.Object {
	t.Helper()

	var options EvalOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// --- Lexing & Parsing ---
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// --- Parser Error Checking ---
	parserErrors := p.Errors()
	if len(parserErrors) != 0 {
		errorMsg := fmt.Sprintf("Parser Errors for input:\n%s\n", input)
		lDebug := lexer.New(input) // Re-lex for debugging output
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

	// --- Environment Setup ---
	env := object.NewEnvironment()

	// Inject standard builtins unless disabled
	if !options.NoBuiltins {
		for name, builtin := range builtins.Builtins {
			env.Set(name, builtin)
		}
	}

	// Simulate sys.argv if provided
	var pylearnArgv *object.List
	if options.Args != nil {
		// Use provided args
		pylearnArgObjs := make([]object.Object, len(options.Args))
		for i, arg := range options.Args {
			pylearnArgObjs[i] = &object.String{Value: arg}
		}
		pylearnArgv = &object.List{Elements: pylearnArgObjs}
	} else {
		// Default: Use a placeholder script name if Args is nil (not empty slice)
		pylearnArgv = &object.List{Elements: []object.Object{&object.String{Value: "test_script.py"}}}
	}
	// Initialize the sys module state with the determined argv
	pysys.InitializeSysModule(pylearnArgv)


	// --- Script Context Setup ---
	if !options.NoScriptDirContext {
		// Set a dummy script context (usually needed for imports)
		wd, err := os.Getwd()
		if err != nil {
			t.Logf("Warning: Could not get working directory for test context: %v", err)
			interpreter.SetCurrentScriptDir(".") // Fallback
		} else {
			// Assume tests run from a predictable location relative to project root
			interpreter.SetCurrentScriptDir(wd) // Use CWD as base for dummy path
		}
	}

	// --- Custom Env Setup ---
	if options.EnvSetup != nil {
		options.EnvSetup(env)
	}


	// --- Evaluation ---
	mainCtx := interpreter.NewInterpreterContext(env)
	evaluated := interpreter.Eval(program, mainCtx)

	// Optional: Check for runtime errors flagged as test failures
	if errObj, ok := evaluated.(*object.Error); ok {
		if strings.HasPrefix(errObj.Message, "TEST_FAIL:") {
			t.Fatalf("Runtime Error indicates test failure: %s", errObj.Message)
		}
	}

	return evaluated
}

// --- Central Assertion Helpers ---

// TestIntegerObject asserts that obj is an Integer with the expected value.
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

// TestFloatObject asserts that obj is a Float with the expected value.
func TestFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	t.Helper()
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	// Consider using tolerance for float comparison if needed:
	// const tolerance = 1e-9
	// if math.Abs(result.Value-expected) > tolerance {
	if result.Value != expected { // Simple check
		t.Errorf("Float has wrong value. got=%g, want=%g", result.Value, expected)
		return false
	}
	return true
}

// TestStringObject asserts that obj is a String with the expected value.
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

// TestBytesObject asserts that obj is a Bytes object with the expected value.
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

// TestBooleanObject asserts that obj is the expected Boolean singleton.
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

// TestNullObject asserts that obj is the NULL singleton.
func TestNullObject(t *testing.T, obj object.Object) bool {
	t.Helper()
	if obj != object.NULL {
		t.Errorf("object is not NULL. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	return true
}

// ErrorInterface is implemented by error types for testing.
type ErrorInterface interface {
	object.Object
	GetMessage() string
}

// Ensure Error and StopIteration implement ErrorInterface
var _ ErrorInterface = (*object.Error)(nil)
// Add StopIteration if it's a distinct type that needs checking
// var _ ErrorInterface = (*interpreter.StopIterationError)(nil) // Example

// TestErrorObject asserts that obj is an Error type containing the expected message parts.
func TestErrorObject(t *testing.T, obj object.Object, expectedMsgParts ...string) bool {
	t.Helper()
	errObj, ok := obj.(ErrorInterface) // Check if it's an Error or similar
	if !ok {
		t.Errorf("object is not an Error type. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	errMsg := errObj.GetMessage()
	for _, part := range expectedMsgParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message %q does not contain expected part %q (object type: %s)", errMsg, part, obj.Type())
			return false
		}
	}
	return true
}

// TestStopIteration asserts that obj is the StopIteration singleton.
func TestStopIteration(t *testing.T, obj object.Object) bool {
	t.Helper()
	// Assuming StopIteration is a specific singleton like NULL/TRUE/FALSE
	// If it's an error type, TestErrorObject might be more appropriate
	if obj != object.STOP_ITERATION { // Adjust if STOP_ITERATION is not defined this way
		t.Errorf("object is not StopIteration singleton. got=%T (%s)", obj, obj.Inspect())
		return false
	}
	return true
}

// TestListObject asserts that obj is a List with the expected elements.
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
		// Recursively test elements using TestObjectLiteral
		if !TestObjectLiteral(t, list.Elements[i], expectedElem) {
			t.Logf("Mismatch at index %d of List", i)
			return false
		}
	}
	return true
}

// TestTupleObject asserts that obj is a Tuple with the expected elements.
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
		// Recursively test elements using TestObjectLiteral
		if !TestObjectLiteral(t, tuple.Elements[i], expectedElem) {
			t.Logf("Mismatch at index %d of Tuple", i)
			return false
		}
	}
	return true
}

// TestSetObject asserts that obj is a Set containing the expected elements (order-independent).
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

	// Create a map of expected elements' hash keys for quick lookup
	expectedMap := make(map[object.HashKey]bool)
	expectedValues := make(map[object.HashKey]interface{}) // Store original values for better error messages

	for _, expElem := range expectedElements {
		// Convert expected Go literal to Pylearn object to hash it
		tempObj := goLiteralToPylearnObject(t, expElem)
		if tempObj == nil {
			return false
		} // Error during conversion
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
		if expectedMap[hKey] {
			t.Errorf("Duplicate element provided in expected set elements: %v (hash: %v)", expElem, hKey)
			return false
		}
		expectedMap[hKey] = true
		expectedValues[hKey] = expElem
	}

	// Check if all elements in the actual set exist in the expected map
	foundKeys := make(map[object.HashKey]bool)
	for hKey, actualElem := range set.Elements {
		if !expectedMap[hKey] {
			t.Errorf("Set contains unexpected element: %s", actualElem.Inspect())
			t.Logf("Got set: %s", set.Inspect())
			t.Logf("Expected elements: %v", expectedElements)
			return false
		}
		foundKeys[hKey] = true
	}

	// Check if all expected elements were found in the actual set
	if len(foundKeys) != len(expectedMap) {
		t.Errorf("Set is missing expected elements.")
		for hKey := range expectedMap {
			if !foundKeys[hKey] {
				t.Errorf("  Missing element: %v (HashKey: %v)", expectedValues[hKey], hKey)
			}
		}
		t.Logf("Got set: %s", set.Inspect())
		return false
	}

	return true
}

// TestObjectLiteral compares a Pylearn object against an expected Go literal value.
func TestObjectLiteral(t *testing.T, obj object.Object, expected interface{}) bool {
	t.Helper()
	switch exp := expected.(type) {
	case int: // Allow int convenience
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
		// Usually indicates a List or Tuple is expected.
		// Delegate to the specific helpers for better context.
		// Prefer using TestListObject or TestTupleObject directly in tests.
		if _, ok := obj.(*object.List); ok {
			return TestListObject(t, obj, exp)
		}
		if _, ok := obj.(*object.Tuple); ok {
			return TestTupleObject(t, obj, exp)
		}
		// Could add Set here too if needed, but TestSetObject is usually better.
		t.Errorf("Expected a Go []interface{} but got neither List nor Tuple. Got %T: %s", obj, obj.Inspect())
		return false
	case *object.Integer: // Allow comparing directly with object types if needed
	    return TestIntegerObject(t, obj, exp.Value)
	case *object.Float:
	    return TestFloatObject(t, obj, exp.Value)
	case *object.String:
	    return TestStringObject(t, obj, exp.Value)
	case *object.Boolean:
	    return TestBooleanObject(t, obj, exp.Value)
    case *object.Null:
        return TestNullObject(t, obj)
    case *object.Bytes:
        return TestBytesObject(t, obj, exp.Value)
	// Add other direct object comparisons if necessary

	default:
		t.Errorf("Unsupported literal type for comparison: %T (%v)", expected, expected)
		return false
	}
}

// goLiteralToPylearnObject converts simple Go literals to Pylearn Objects. Used internally by TestSetObject.
func goLiteralToPylearnObject(t *testing.T, literal interface{}) object.Object {
	t.Helper()
	switch v := literal.(type) {
	case int:
		return &object.Integer{Value: int64(v)}
	case int64:
		return &object.Integer{Value: v}
	case float64:
		// Floats are generally not good set elements due to precision,
		// but create the object if needed for testing.
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case bool:
		return object.NativeBoolToBooleanObject(v)
	case nil:
		return object.NULL
	case []byte:
		return &object.Bytes{Value: v}
	// Add Tuple conversion if needed and tuples are hashable in your implementation
	// case []interface{}: // Example for Tuple
	//  elems := make([]object.Object, len(v))
	//  for i, item := range v {
	//      elemObj := goLiteralToPylearnObject(t, item) // Recursive call
	//      if elemObj == nil { return nil }
	//      elems[i] = elemObj
	//  }
	//  return &object.Tuple{Elements: elems}
	default:
		t.Errorf("Cannot convert Go literal type %T to Pylearn object for hashing/comparison", literal)
		return nil
	}
}

// --- OOP Specific Assertions (Keep here or move to oop_test if not widely used) ---

// TestClassObject asserts obj is a Class with the expected name.
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

// TestInstanceObject asserts obj is an Instance of the expected Class name.
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


// --- Stdlib Specific Assertions ---
