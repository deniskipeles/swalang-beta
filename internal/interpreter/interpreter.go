// pylearn/internal/interpreter/interpreter.go
package interpreter

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// Eval is the main function that recursively evaluates AST nodes.
// It now takes the full InterpreterContext to preserve super() and other contexts.
func Eval(node ast.Node, ctx *InterpreterContext) object.Object {
	switch node := node.(type) {
	// --- Statements ---
	case *ast.Program:
		return evalProgram(node.Statements, ctx)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, ctx)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, ctx)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, ctx)
	case *ast.LetStatement:
		return evalLetStatement(node, ctx)
	case *ast.IfStatement:
		return evalIfStatement(node, ctx)
	case *ast.TryStatement:
		return evalTryStatement(node, ctx)
	case *ast.RaiseStatement:
		return evalRaiseStatement(node, ctx)
	case *ast.WhileStatement:
		return evalWhileStatement(node, ctx)
	case *ast.ForStatement:
		return evalForStatement(node, ctx)
	case *ast.BreakStatement:
		return evalBreakStatement(node, ctx)
	case *ast.ContinueStatement:
		return evalContinueStatement(node, ctx)
	case *ast.YieldExpression:
		if ctx.IsResuming {
			// Resuming from a yield. The value of this expression is what was sent.
			ctx.IsResuming = false // Reset the flag immediately after using it.
			return ctx.SentInValue
		}

		// Pausing at a yield. Evaluate the expression to be yielded
		// and wrap it in the special YieldValue signal object.
		var valToYield object.Object = object.NULL
		if node.Value != nil {
			valToYield = Eval(node.Value, ctx)
		}
		if object.IsError(valToYield) {
			return valToYield
		}
		return &object.YieldValue{Value: valToYield}
	case *ast.ClassStatement:
		return evalClassStatement(node, ctx)
	case *ast.PassStatement:
		return evalPassStatement(node, ctx)
	case *ast.AssignStatement:
		return evalAssignStatement(node, ctx)
	case *ast.ImportStatement:
		return evalImportStatement(node, ctx)
	case *ast.FromImportStatement:
		return evalFromImportStatement(node, ctx)
	case *ast.WithStatement:
		return evalWithStatement(node, ctx)
	case *ast.AssertStatement:
		return evalAssertStatement(node, ctx)
	case *ast.GlobalStatement:
		return evalGlobalStatement(node, ctx)
	case *ast.DelStatement:
		return evalDelStatement(node, ctx)
	// --- Expressions ---
	case *ast.Identifier:
		return evalIdentifier(node, ctx.Env) // Env is sufficient here
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.BooleanLiteral:
		return object.NativeBoolToBooleanObject(node.Value)
	case *ast.NilLiteral:
		return object.NULL
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.BytesLiteral:
		return &object.Bytes{Value: node.Value}
	case *ast.ListLiteral:
		return evalListLiteral(node, ctx)
	case *ast.ListComprehension:
		return evalListComprehension(node, ctx)
	case *ast.TupleLiteral:
		return evalTupleLiteral(node, ctx)
	case *ast.SetLiteral:
		return evalSetLiteral(node, ctx)
	case *ast.SetComprehension:
		return evalSetComprehension(node, ctx)
	case *ast.DictLiteral:
		return evalDictLiteral(node, ctx)
	case *ast.PrefixExpression:
		return evalPrefixExpression(node, ctx)
	case *ast.InfixExpression:
		return evalInfixExpression(node, ctx)
	case *ast.TernaryExpression:
		return evalTernaryExpression(node, ctx)
	case *ast.FunctionLiteral:
		funcName := constants.EmptyString
		if node.Name != nil {
			funcName = node.Name.Value
		}
		objParams := make([]*object.FunctionParameter, len(node.Parameters))
		for i, astParam := range node.Parameters {
			var defaultValueObject object.Object = nil
			if astParam.DefaultValue != nil {
				defaultValueObject = Eval(astParam.DefaultValue, ctx) // Pass ctx
				if object.IsError(defaultValueObject) {
					errObj := defaultValueObject.(*object.Error)
					errObj.Message = fmt.Sprintf(constants.InterpreterEvalParamDefaultError, astParam.Name.Value, errObj.Message)
					return errObj
				}
			}
			objParams[i] = &object.FunctionParameter{Name: astParam.Name.Value, DefaultValue: defaultValueObject}
		}
		var varArgName string = constants.EmptyString
		if node.VarArgParam != nil {
			varArgName = node.VarArgParam.Value
		}
		var kwArgName string = constants.EmptyString
		if node.KwArgParam != nil {
			kwArgName = node.KwArgParam.Value
		}
		return &object.Function{
			Name:        funcName,
			Parameters:  objParams,
			VarArgParam: varArgName,
			KwArgParam:  kwArgName,
			Body:        node.Body,
			Env:         ctx.Env, // Functions capture the environment they are defined in
			IsAsync:     node.IsAsync,
		}
	case *ast.AwaitExpression:
		// DELEGATE to the new, robust handler.
		return evalAwaitExpression(node, ctx)
	
	case *ast.DotExpression:
		return evalDotExpression(node, ctx)
	case *ast.CallExpression:
		return evalCallExpression(node, ctx)
	case *ast.IndexExpression:
		return evalIndexExpression(node, ctx)
	case *ast.SliceExpression:
		return evalSliceExpression(node, ctx)
	case *ast.LambdaLiteral:
		// Create FunctionParameter objects from the AST parameters
		objParams := make([]*object.FunctionParameter, len(node.Parameters))
		for i, astParam := range node.Parameters {
			var defaultValueObject object.Object = nil
			if astParam.DefaultValue != nil {
				defaultValueObject = Eval(astParam.DefaultValue, ctx)
				if object.IsError(defaultValueObject) {
					errObj := defaultValueObject.(*object.Error)
					errObj.Message = fmt.Sprintf(constants.ErrorEvaluatingDefaultForParameter_STRINGFORMATER_InLambda_STRINGFORMATER, astParam.Name.Value, errObj.Message)
					return errObj
				}
			}
			objParams[i] = &object.FunctionParameter{
				Name:         astParam.Name.Value,
				DefaultValue: defaultValueObject,
			}
		}

		// --- THIS IS THE FIX ---
		// The body of a lambda is a single expression. To execute it,
		// we wrap it in a BlockStatement and a ReturnStatement.
		returnStmt := &ast.ReturnStatement{
			// Use the ast.GetToken helper to get the actual token from the body expression.
			Token:       ast.GetToken(node.Body),
			ReturnValue: node.Body,
		}
		// --- END OF FIX ---
		
		bodyBlock := &ast.BlockStatement{
			Token:      node.Token,
			Statements: []ast.Statement{returnStmt},
		}

		return &object.Function{
			Name:       constants.Lambda,
			Parameters: objParams,
			Body:       bodyBlock,
			Env:        ctx.Env, // Lambdas capture the environment they are defined in
			IsAsync:    false,
		}
	default:
		token := object.NoToken
		if node != nil {
		}
		return object.NewErrorWithToken(node.TokenLiteral(), token, constants.SyntaxError, constants.InterpreterEvalSyntaxErrorNode, node)
	}
}


// Add the new evaluation function to the file
func evalGlobalStatement(node *ast.GlobalStatement, ctx *InterpreterContext) object.Object {
	for _, name := range node.Names {
		ctx.Env.RegisterGlobal(name.Value)
	}
	return object.NULL
}