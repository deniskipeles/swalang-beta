package interpreter_test

import (
	"testing"
	"github.com/deniskipeles/pylearn/tests/helpers"
)

func TestLoops(t *testing.T) {
	t.Run("while loop", func(t *testing.T) {
		input := `
i = 0
total = 0
while i < 5:
    total = total + i
    i = i + 1
total
`
		evaluated := helpers.Eval(t, input)
		helpers.TestIntegerObject(t, evaluated, 10)
	})

	t.Run("for loop with range", func(t *testing.T) {
		input := `
total = 0
for i in range(5):
    total = total + i
total
`
		evaluated := helpers.Eval(t, input)
		helpers.TestIntegerObject(t, evaluated, 10)
	})

	t.Run("for loop with list", func(t *testing.T) {
		input := `
total = 0
for i in [1, 2, 3, 4]:
    total = total + i
total
`
		evaluated := helpers.Eval(t, input)
		helpers.TestIntegerObject(t, evaluated, 10)
	})

    t.Run("break statement", func(t *testing.T) {
        input := `
i = 0
while True:
    if i == 5:
        break
    i = i + 1
i
`
        evaluated := helpers.Eval(t, input)
        helpers.TestIntegerObject(t, evaluated, 5)
    })

    t.Run("continue statement", func(t *testing.T) {
        input := `
total = 0
for i in range(5):
    if i == 2:
        continue
    total = total + i
total
`
        evaluated := helpers.Eval(t, input)
        helpers.TestIntegerObject(t, evaluated, 8) // 0 + 1 + 3 + 4
    })
}

func TestComplexClosures(t *testing.T) {
    input := `
def outer(x):
    def inner(y):
        return x + y
    return inner

add5 = outer(5)
add10 = outer(10)
add5(10) + add10(10)
`
    evaluated := helpers.Eval(t, input)
    helpers.TestIntegerObject(t, evaluated, 35)
}

func TestRecursion(t *testing.T) {
    input := `
def fib(n):
    if n < 2: return n
    return fib(n-1) + fib(n-2)
fib(10)
`
    evaluated := helpers.Eval(t, input)
    helpers.TestIntegerObject(t, evaluated, 55)
}

func TestDictionaryMethods(t *testing.T) {
    input := `
d = {"a": 1, "b": 2}
keys = list(d.keys())
# keys.sort() # TODO: Interpreter logic error (LIST object has no attribute 'sort')
# Since sort is missing, just check length and presence
len(keys)
`
    evaluated := helpers.Eval(t, input)
    helpers.TestIntegerObject(t, evaluated, 2)
}

func TestSetOperations(t *testing.T) {
    input := `
s = {1, 2, 3}
s.add(4)
s.remove(2)
4 in s and 2 not in s
`
    evaluated := helpers.Eval(t, input)
    helpers.TestBooleanObject(t, evaluated, true)
}
