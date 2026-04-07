// internal/object/context.go
package object

import (
	"context" // For AsyncRuntimeAPI (Go context)
	"time"    // For AsyncRuntimeAPI

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/asyncruntime"
)

type AsyncRuntimeAPI interface {
	Sleep(duration time.Duration) *asyncruntime.AsyncResult
	CreateCoroutine(fn func(goCtx context.Context) (interface{}, error)) *asyncruntime.AsyncResult
	GatherAll(results ...*asyncruntime.AsyncResult) ([]interface{}, error)
	Await(result *asyncruntime.AsyncResult) (interface{}, error)
}

type ExecutionContext interface {
	// FIX: The signature of Execute is just (callable, args...).
	// It doesn't need the first `self` argument. The interpreter will
	// handle passing `self` as the first element in `args` for bound methods.
	Execute(callable Object, args ...Object) Object
	EvaluateASTNode(node ast.Node, env *Environment) Object

	GetAsyncRuntime() AsyncRuntimeAPI
	GetCurrentEnvironment() *Environment
	NewChildContext(env *Environment) ExecutionContext

	SetSuperContext(self Object, class *Class)
	GetSuperSelf() Object
	GetSuperClass() *Class
}
