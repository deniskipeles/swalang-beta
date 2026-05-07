package builtins

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
)

func pyFormatStringFunction(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsFstringFormatStringFnArgCountError, len(args))
	}
	fstrLiteral, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.BuiltinsFstringFormatStringFnArgTypeError, args[0].Type())
	}

	inputString := fstrLiteral.Value
	var result bytes.Buffer
	lastIndex := 0

	for i := 0; i < len(inputString); {
		char := inputString[i]
		if char == '{' {
			if i+1 < len(inputString) && inputString[i+1] == '{' {
				result.WriteString(inputString[lastIndex:i])
				result.WriteByte('{')
				i += 2
				lastIndex = i
				continue
			}
			result.WriteString(inputString[lastIndex:i])
			braceDepth := 1
			exprStart := i + 1
			exprEnd := -1
			for j := exprStart; j < len(inputString); j++ {
				if inputString[j] == '{' {
					braceDepth++
				} else if inputString[j] == '}' {
					braceDepth--
					if braceDepth == 0 {
						exprEnd = j
						break
					}
				}
			}
			if exprEnd == -1 {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringUnterminatedExpression, exprStart-1)
			}

			// --- START OF FORMAT SPECIFIER EXTRACT ---
			fullContent := inputString[exprStart:exprEnd]
			expressionStr := fullContent
			formatSpec := ""

			colonPos := -1
			innerBraceDepth := 0
			for k := 0; k < len(fullContent); k++ {
				if fullContent[k] == '{' {
					innerBraceDepth++
				} else if fullContent[k] == '}' {
					innerBraceDepth--
				} else if fullContent[k] == ':' && innerBraceDepth == 0 {
					colonPos = k
					break
				}
			}

			if colonPos != -1 {
				expressionStr = strings.TrimSpace(fullContent[:colonPos])
				formatSpec = fullContent[colonPos+1:]
			}
			// --- END OF FORMAT SPECIFIER EXTRACT ---

			if expressionStr == "" {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringEmptyExpressionNotAllowed)
			}

			exprLexer := lexer.New(expressionStr)
			exprParser := parser.New(exprLexer)
			programAST := exprParser.ParseProgram()
			if len(exprParser.Errors()) > 0 {
				errMsg := fmt.Sprintf(constants.BuiltinsFstringSyntaxErrorInExpression, expressionStr, strings.Join(exprParser.Errors(), constants.SemiColonWithSpace))
				return object.NewError(constants.SyntaxError, errMsg)
			}
			if len(programAST.Statements) == 0 || programAST.Statements[0] == nil {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringExpressionParsedToNoStatements, expressionStr)
			}
			var astNodeToEvaluate ast.Node
			if exprStmt, isExprStmt := programAST.Statements[0].(*ast.ExpressionStatement); isExprStmt {
				astNodeToEvaluate = exprStmt.Expression
			} else {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringContentMustBeExpression, expressionStr)
			}

			// Use the ExecutionContext to evaluate the AST node.
			evaluatedObj := ctx.EvaluateASTNode(astNodeToEvaluate, nil)

			if object.IsError(evaluatedObj) {
				return evaluatedObj
			}

			// --- APPLY FORMAT SPECIFIER ---
			var finalStr string
			if formatSpec != "" {
				formatted, err := applyFormatSpec(evaluatedObj, formatSpec)
				if err != nil {
					return object.NewError(constants.ValueError, err.Error())
				}
				finalStr = formatted
			} else {
				// No format spec, just convert to string
				strBuiltin, strBuiltinFound := Builtins[constants.BuiltinsStrFuncName]
				if !strBuiltinFound {
					return object.NewError(constants.InternalError, constants.BuiltinsFstringStrBuiltinNotFound)
				}
				strRepresentationObj := ctx.Execute(strBuiltin, evaluatedObj)
				if object.IsError(strRepresentationObj) {
					errObj := strRepresentationObj.(*object.Error)
					errObj.Message = fmt.Sprintf(constants.BuiltinsFstringErrorConvertingResult, expressionStr, errObj.Message)
					return errObj
				}
				strRepresentation, isStr := strRepresentationObj.(*object.String)
				if !isStr {
					return object.NewError(constants.InternalError, constants.BuiltinsFstringStrDidNotReturnString, strRepresentationObj.Type())
				}
				finalStr = strRepresentation.Value
			}
			result.WriteString(finalStr)
			// --- END APPLY FORMAT SPECIFIER ---

			i = exprEnd + 1
			lastIndex = i
		} else if char == '}' {
			if i+1 < len(inputString) && inputString[i+1] == '}' {
				result.WriteString(inputString[lastIndex:i])
				result.WriteByte('}')
				i += 2
				lastIndex = i
				continue
			}
			return object.NewError(constants.ValueError, constants.BuiltinsFstringSingleBraceNotAllowed)
		} else {
			i++
		}
	}
	if lastIndex < len(inputString) {
		result.WriteString(inputString[lastIndex:])
	}
	return &object.String{Value: result.String()}
}

// applyFormatSpec applies a Python-style format specifier to a Pylearn object.
// This is a simplified implementation supporting standard format letters.
func applyFormatSpec(obj object.Object, spec string) (string, error) {
	if spec == "" {
		if strer, ok := obj.(interface{ String() string }); ok {
			return strer.String(), nil
		}
		return obj.Inspect(), nil
	}

	lastChar := spec[len(spec)-1]
	switch lastChar {
	case 'f', 'F', 'e', 'E', 'g', 'G':
		var floatVal float64
		switch o := obj.(type) {
		case *object.Float:
			floatVal = o.Value
		case *object.Integer:
			floatVal = float64(o.Value)
		default:
			return "", fmt.Errorf(constants.BuiltinFormatSpecifier_F_RequiresFloatOrIntegerNot_STRINGFORMATER, obj.Type())
		}
		return fmt.Sprintf("%"+spec, floatVal), nil
	case 'd', 'x', 'X', 'o', 'b', 'c':
		var intVal int64
		switch o := obj.(type) {
		case *object.Integer:
			intVal = o.Value
		default:
			return obj.Inspect(), nil
		}
		return fmt.Sprintf("%"+spec, intVal), nil
	case 's':
		if strObj, ok := obj.(*object.String); ok {
			return fmt.Sprintf("%"+spec, strObj.Value), nil
		}
		return fmt.Sprintf("%"+spec, obj.Inspect()), nil
	}

	// Default/unsupported: just return the inspection of the object
	return obj.Inspect(), nil
}

func pyFStringFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsFstringFormatStringFnArgCountError, len(args))
	}
	fstrLiteral, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.BuiltinsFstringFormatStringFnArgTypeError, args[0].Type())
	}

	inputString := fstrLiteral.Value
	var result bytes.Buffer
	lastIndex := 0

	for i := 0; i < len(inputString); {
		char := inputString[i]
		if char == '{' {
			if i+1 < len(inputString) && inputString[i+1] == '{' {
				result.WriteString(inputString[lastIndex:i])
				result.WriteByte('{')
				i += 2
				lastIndex = i
				continue
			}
			result.WriteString(inputString[lastIndex:i])
			braceDepth, exprStart, exprEnd := 1, i+1, -1
			for j := exprStart; j < len(inputString); j++ {
				if inputString[j] == '{' {
					braceDepth++
				}
				if inputString[j] == '}' {
					braceDepth--
					if braceDepth == 0 {
						exprEnd = j
						break
					}
				}
			}
			if exprEnd == -1 {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringUnterminatedExpression, exprStart-1)
			}

			// --- START OF FORMAT SPECIFIER FIX ---
			fullContent := inputString[exprStart:exprEnd]
			expressionStr := fullContent
			formatSpec := ""

			// Find the format specifier separator ':'
			// We only care about the rightmost one that is not inside nested braces.
			colonPos := -1
			innerBraceDepth := 0
			for k := 0; k < len(fullContent); k++ {
				if fullContent[k] == '{' {
					innerBraceDepth++
				} else if fullContent[k] == '}' {
					innerBraceDepth--
				} else if fullContent[k] == ':' && innerBraceDepth == 0 {
					colonPos = k
					break // Found the main format spec separator
				}
			}

			if colonPos != -1 {
				expressionStr = strings.TrimSpace(fullContent[:colonPos])
				formatSpec = fullContent[colonPos+1:]
			}
			// --- END OF FORMAT SPECIFIER FIX ---

			if expressionStr == "" {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringEmptyExpressionNotAllowed)
			}

			exprLexer := lexer.New(expressionStr)
			exprParser := parser.New(exprLexer)
			programAST := exprParser.ParseProgram()
			if len(exprParser.Errors()) > 0 {
				return object.NewError(constants.SyntaxError, constants.BuiltinsFstringSyntaxErrorInExpression, expressionStr, strings.Join(exprParser.Errors(), "; "))
			}
			if len(programAST.Statements) == 0 {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringExpressionParsedToNoStatements, expressionStr)
			}

			var astNodeToEvaluate ast.Node
			if exprStmt, isExprStmt := programAST.Statements[0].(*ast.ExpressionStatement); isExprStmt {
				astNodeToEvaluate = exprStmt.Expression
			} else {
				return object.NewError(constants.ValueError, constants.BuiltinsFstringContentMustBeExpression, expressionStr)
			}

			evaluatedObj := ctx.EvaluateASTNode(astNodeToEvaluate, nil)
			if object.IsError(evaluatedObj) {
				return evaluatedObj
			}

			// --- APPLY FORMAT SPECIFIER ---
			var finalStr string
			if formatSpec != "" {
				formatted, err := applyFormatSpec(evaluatedObj, formatSpec)
				if err != nil {
					return object.NewError(constants.ValueError, err.Error())
				}
				finalStr = formatted
			} else {
				// No format spec, just convert to string
				strBuiltin, _ := Builtins[constants.BuiltinsStrFuncName]
				strRepresentationObj := ctx.Execute(strBuiltin, evaluatedObj)
				if strObj, isStr := strRepresentationObj.(*object.String); isStr {
					finalStr = strObj.Value
				} else {
					finalStr = strRepresentationObj.Inspect() // Fallback
				}
			}
			result.WriteString(finalStr)
			// --- END APPLY FORMAT SPECIFIER ---

			i = exprEnd + 1
			lastIndex = i
		} else if char == '}' {
			if i+1 < len(inputString) && inputString[i+1] == '}' {
				result.WriteString(inputString[lastIndex:i])
				result.WriteByte('}')
				i += 2
				lastIndex = i
				continue
			}
			return object.NewError(constants.ValueError, constants.BuiltinsFstringSingleBraceNotAllowed)
		} else {
			i++
		}
	}
	if lastIndex < len(inputString) {
		result.WriteString(inputString[lastIndex:])
	}
	return &object.String{Value: result.String()}
}

func init() {
	// this is a built-in function format_str("")
	registerBuiltin(constants.BuiltinsFormatStringFuncName, &object.Builtin{Fn: pyFormatStringFunction, Name: constants.BuiltinsFormatStringFuncName})
	// Add the new f-string built-in
	registerBuiltin(constants.BuiltinsFormatStrFuncName, &object.Builtin{Fn: pyFStringFn, Name: constants.BuiltinsFormatStrFuncName})
}