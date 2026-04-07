package tests

import (
	"strings" // Keep for specific error checks if needed
	"testing"

	// Use centralized helpers
	"github.com/deniskipeles/pylearn/internal/testhelpers"
	// Import object only if directly manipulating/checking instance internals
	"github.com/deniskipeles/pylearn/internal/object"

	// Ensure necessary stdlib/builtins are available via interpreter/builtins packages
	// _ "github.com/deniskipeles/pylearn/internal/builtins"
    // _ "github.com/deniskipeles/pylearn/internal/stdlib/..."
)


// No need for duplicated testEvalOop or assertion helpers anymore

func TestClassDefinitionEval(t *testing.T) {
	input := `
class MyClass:
  cv = 10
  def method(self):
    pass
# Access the class object itself to test definition
MyClass
`
	evaluated := testhelpers.Eval(t, input)
	// Use OOP-specific helper (could be in testhelpers or kept here)
	testhelpers.TestClassObject(t, evaluated, "MyClass")
}

func TestInstanceCreationNoInit(t *testing.T) {
	t.Run("Successful Creation", func(t *testing.T) {
		input := `
class Simple:
  pass
s = Simple()
s # Evaluate the instance
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestInstanceObject(t, evaluated, "Simple")
	})

	t.Run("Error On Args When No Init", func(t *testing.T) {
		input := `
class Simple:
  pass
Simple(1, 2) # Call class with args when no __init__ is defined
`
		evaluated := testhelpers.Eval(t, input)
		// Check error message content
		testhelpers.TestErrorObject(t, evaluated, "TypeError", "Simple() takes no arguments", "2 given")
	})
}

func TestInstanceCreationWithInit(t *testing.T) {
	input := `
class Point:
  def __init__(self, x_val, y_val):
    self.x = x_val
    self.y = y_val
p = Point(10, 20)
p # Evaluate the instance
`
	evaluated := testhelpers.Eval(t, input)
	// Check instance type
	if !testhelpers.TestInstanceObject(t, evaluated, "Point") {
		return // Stop if not the right instance type
	}

	// Check instance attributes directly (requires type assertion)
	instance, ok := evaluated.(*object.Instance)
	if !ok {
		t.Fatal("Evaluated object is not an Instance despite passing TestInstanceObject check") // Should not happen
	}

	xAttr, xOk := instance.Env.Get("x") // Access internal env for verification
	if !xOk {
		t.Fatal("Attribute 'x' not found on instance environment")
	}
	testhelpers.TestIntegerObject(t, xAttr, 10) // Use standard helper

	yAttr, yOk := instance.Env.Get("y")
	if !yOk {
		t.Fatal("Attribute 'y' not found on instance environment")
	}
	testhelpers.TestIntegerObject(t, yAttr, 20)
}

func TestInitArityErrors(t *testing.T) {
	baseCode := `
class Coord:
  def __init__(self, a, b):
    self.a = a
    self.b = b
`
	t.Run("Too Few Args", func(t *testing.T) {
		input := baseCode + "Coord(1)"
		evaluated := testhelpers.Eval(t, input)
		// Example: Check for Python-like error message parts
		testhelpers.TestErrorObject(t, evaluated, "TypeError", "__init__()", "missing 1 required positional argument", "'b'")
	})
	t.Run("Too Many Args", func(t *testing.T) {
		input := baseCode + "Coord(1, 2, 3)"
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestErrorObject(t, evaluated, "TypeError", "__init__()", "takes 3 positional arguments", "4 were given")
	})
}


func TestAttributeAccessRead(t *testing.T) {
	setupCode := `
class Data:
  class_var = "classy"
  def __init__(self):
    self.instance_var = 123
d = Data()
`
	t.Run("Read Instance Var", func(t *testing.T) {
		input := setupCode + "d.instance_var"
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestIntegerObject(t, evaluated, 123)
	})
	t.Run("Read Class Var via Instance", func(t *testing.T) {
		input := setupCode + "d.class_var"
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestStringObject(t, evaluated, "classy")
	})
	t.Run("AttributeError", func(t *testing.T) {
		input := setupCode + "d.non_existent"
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestErrorObject(t, evaluated, "AttributeError", "'Data' object has no attribute 'non_existent'")
	})
	t.Run("Instance Var Shadows Class Var", func(t *testing.T) {
		input := `
class Shadow:
  var = "class"
  def __init__(self):
    self.var = "instance"
s = Shadow()
s.var # Access should get instance var
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestStringObject(t, evaluated, "instance")
	})
}

func TestMethodCallSimple(t *testing.T) {
	input := `
class Calc:
  def get_const(self):
    return 42
c = Calc()
c.get_const() # Call the method
`
	evaluated := testhelpers.Eval(t, input)
	testhelpers.TestIntegerObject(t, evaluated, 42)
}

func TestDunderStr(t *testing.T) {
	t.Run("Instance with __str__", func(t *testing.T) {
		input := `
class Person:
  def __init__(self, name):
    self.name = name
  def __str__(self):
    # Use f-string or equivalent for formatting if implemented
    return "Person(" + self.name + ")"
p = Person("Alice")
str(p) # Call the str() builtin
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestStringObject(t, evaluated, "Person(Alice)")
	})

	t.Run("Instance without __str__", func(t *testing.T) {
		input := `
class Thing:
  pass
t = Thing()
s = str(t)
s # Evaluate the default string representation
`
		evaluated := testhelpers.Eval(t, input)
		strResult, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("str() did not return a String object, got %T", evaluated)
		}
		// Check for Python-like default format <ClassName object at 0x...>
		if !strings.HasPrefix(strResult.Value, "<Thing object at 0x") || !strings.HasSuffix(strResult.Value, ">") {
			t.Errorf("Default str() representation format is wrong. got=%q", strResult.Value)
		}
	})

	t.Run("__str__ must return string", func(t *testing.T) {
		input := `
class BadStr:
  def __str__(self):
    return 123 # Return non-string
b = BadStr()
str(b)
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestErrorObject(t, evaluated, "TypeError", "__str__ returned non-string")
	})
}


// func TestDunderAdd(t *testing.T) { ... } // Keep dunder tests using testhelpers

func TestDunderGetItemSetItem(t *testing.T) {
	t.Run("__getitem__", func(t *testing.T) {
		input := `
class SimpleList:
  def __init__(self):
    self._data = [10, 20, 30] # Assuming lists work
  def __getitem__(self, index):
    # Basic index check, assumes index is int
    return self._data[index]
sl = SimpleList()
sl[1] # Test getting item
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestIntegerObject(t, evaluated, 20)
	})

	t.Run("__setitem__", func(t *testing.T) {
		input := `
class SimpleDict:
  def __init__(self):
    self._data = {} # Assuming dicts work
  def __setitem__(self, key, value):
    # Basic implementation, convert key to string maybe?
    self._data[str(key)] = value # Example: Force string keys
  def __getitem__(self, key):
    return self._data[str(key)]
sd = SimpleDict()
sd["foo"] = 100
sd[5] = "bar" # Test with non-string key
# Check if setting worked by getting item
sd["foo"]
`
		evaluatedGetFoo := testhelpers.Eval(t, input)
		testhelpers.TestIntegerObject(t, evaluatedGetFoo, 100)

        // Verify the item set with int key (accessed via string representation)
        inputGet5 := input + "\nsd['5']" // Access using the string key used internally
		evaluatedGet5 := testhelpers.Eval(t, inputGet5)
		testhelpers.TestStringObject(t, evaluatedGet5, "bar")

	})

    t.Run("__getitem__ IndexError", func(t *testing.T) {
		input := `
class SimpleList:
  def __init__(self): self._data = [10]
  def __getitem__(self, index): return self._data[index] # Needs real list bounds checking
sl = SimpleList()
sl[5] # Access out of bounds
`
        // This test depends on the underlying list implementation raising IndexError
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestErrorObject(t, evaluated, "IndexError", "list index out of range")
	})

     t.Run("__setitem__ TypeError (if key unhashable)", func(t *testing.T) {
		input := `
class SimpleDict:
  def __init__(self): self._data = {}
  def __setitem__(self, key, value): self._data[key] = value # Needs proper dict key handling
sd = SimpleDict()
sd[[]] = 1 # Use list as key (unhashable)
`
        // This test depends on the underlying dict implementation raising TypeError
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestErrorObject(t, evaluated, "TypeError", "unhashable type: 'list'")
	})
}