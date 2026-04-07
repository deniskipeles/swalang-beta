package parser

import (
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/lexer"
	// "github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
)

// parseLambdaLiteral parses a lambda expression: lambda <params>: <expression>
func (p *Parser) parseLambdaLiteral() ast.Expression {
	lit := &ast.LambdaLiteral{Token: p.curToken} // Current token is LAMBDA

	// Lambda parameters are optional. If a colon follows immediately, there are no params.
	if !p.peekTokenIs(lexer.COLON) {
		p.nextToken() // Consume LAMBDA, move to first parameter
		// We can reuse the parameter parsing logic, but it needs to stop at COLON.
		// We'll create a simplified version here for clarity.
		lit.Parameters = p.parseLambdaParameters()
	} else {
		lit.Parameters = []*ast.Parameter{}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	// Current token is now COLON

	p.nextToken() // Consume COLON, move to start of body expression

	// The body of a lambda is a single expression.
	// Its precedence should be low to allow complex expressions, but not assignments.
	// We parse until we hit a comma, paren, etc., which would be part of an outer expression.
	lit.Body = p.parseExpression(LOWEST)
	if lit.Body == nil {
		return nil
	}

	return lit
}

// parseLambdaParameters is a helper to parse the parameter list for a lambda.
// It's similar to parseFunctionParameters but stops at a COLON instead of a RPAREN.
func (p *Parser) parseLambdaParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Parse first parameter
	if p.curTokenIs(lexer.COLON) {
		return params // No parameters
	}

	// This logic can be simplified from the full function parameter parser,
	// as we don't need to handle *args and **kwargs in this first pass for simplicity.
	// A full implementation would reuse a more generic parameter list parser.
	for {
		if !p.curTokenIs(lexer.IDENT) {
			p.errorExpected("identifier in lambda parameter list", p.curToken.String())
			return nil
		}
		param := &ast.Parameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		// Check for default value
		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken() // move to ASSIGN
			p.nextToken() // move to start of default value
			param.DefaultValue = p.parseExpression(OR)
		}

		params = append(params, param)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume param name
			p.nextToken() // consume comma
		} else if p.peekTokenIs(lexer.COLON) {
			break // End of parameter list
		} else {
			p.peekErrorMsg("expected ',' or ':' after lambda parameter")
			return nil
		}
	}

	return params
}
