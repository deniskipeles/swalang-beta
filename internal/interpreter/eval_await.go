package interpreter

import (
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/object"
)

func evalAwaitExpression(node *ast.AwaitExpression, ctx *InterpreterContext) object.Object {
	// If the event loop has resumed us with a value, we return it as the result of the `await`.
	if ctx.IsResuming {
		ctx.IsResuming = false
		return ctx.SentInValue
	}

	// Otherwise, evaluate the expression being awaited (e.g., a coroutine or future)
	awaitable := Eval(node.Expression, ctx)
	if object.IsError(awaitable) {
		return awaitable
	}

	// Yield the awaitable object up to the Event Loop
	return &object.YieldValue{Value: awaitable}
}