package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
	"github.com/deniskipeles/pylearn/internal/lexer"
)

// Precedence levels for operators
const (
	_ int = iota
	LOWEST
	ASSIGN // =
	TUPLE_PRECEDENCE
	TERNARY // if else
	OR      // or
	AND     // and
	BITWISE
	EQUALS      // ==, !=, is, is not
	LESSGREATER // >, <, >=, <=, in, not in
	SHIFT       // >>, <<
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // -X or not X
	POWER       // **
	CALL        // myFunction(X)
	INDEX       // array[index]
	DOT         // object.attribute
)

// Map token types to their precedence levels
var precedences = map[lexer.TokenType]int{
	lexer.COMMA:       TUPLE_PRECEDENCE,
	lexer.ASSIGN:      ASSIGN,
	lexer.PLUS_EQ:     ASSIGN,
	lexer.MINUS_EQ:    ASSIGN,
	lexer.IF:          TERNARY,
	lexer.EQ:          EQUALS,
	lexer.NOT_EQ:      EQUALS,
	lexer.IS:          EQUALS,
	lexer.IS_NOT:      EQUALS,
	lexer.LT:          LESSGREATER,
	lexer.GT:          LESSGREATER,
	lexer.LT_EQ:       LESSGREATER,
	lexer.GT_EQ:       LESSGREATER,
	lexer.IN:          LESSGREATER,
	lexer.NOT_IN:      LESSGREATER,
	lexer.LSHIFT:      SHIFT,
	lexer.RSHIFT:      SHIFT,
	lexer.PLUS:        SUM,
	lexer.MINUS:       SUM,
	lexer.SLASH:       PRODUCT,
	lexer.FLOOR_DIV:   PRODUCT,
	lexer.ASTERISK:    PRODUCT,
	lexer.PERCENT:     PRODUCT,
	lexer.POW:         POWER,
	lexer.DOT:         DOT,   // <<< CORRECT
	lexer.LPAREN:      CALL,  // <<< CORRECT
	lexer.LBRACKET:    INDEX, // <<< CORRECT
	lexer.AND:         AND,
	lexer.OR:          OR,
	lexer.BITWISE_AND: BITWISE,
}

// Type definitions for parsing functions
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser holds the lexer, tokens, errors, and parsing functions.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

// New creates a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)

	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.FSTRING, p.parseFStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.NIL, p.parseNilLiteral)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedOrTupleExpression)
	p.registerPrefix(lexer.LBRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseDictOrSetLiteral)
	p.registerPrefix(lexer.BYTES, p.parseBytesLiteral)

	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.FLOOR_DIV, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.POW, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.IN, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignExpression)
	p.registerInfix(lexer.PLUS_EQ, p.parseAssignExpression)
	p.registerInfix(lexer.MINUS_EQ, p.parseAssignExpression)
	p.registerInfix(lexer.BITWISE_AND, p.parseInfixExpression)
	p.registerInfix(lexer.LSHIFT, p.parseInfixExpression)
	p.registerInfix(lexer.RSHIFT, p.parseInfixExpression)
	p.registerInfix(lexer.COMMA, p.parseTupleLiteralInfix)

	// p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerInfix(lexer.IF, p.parseTernaryExpression)

	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseDotExpression)

	p.registerPrefix(lexer.AWAIT, p.parseAwaitExpression)
	p.registerPrefix(lexer.YIELD, p.parseYieldExpression)
	p.registerPrefix(lexer.LAMBDA, p.parseLambdaLiteral)

	// --- Register new infix functions ---
	p.registerInfix(lexer.NOT_IN, p.parseInfixExpression)
	p.registerInfix(lexer.IS, p.parseInfixExpression)
	p.registerInfix(lexer.IS_NOT, p.parseInfixExpression)

	p.curToken = p.l.NextToken()
	p.peekToken = p.l.NextToken()

	return p
}

// Errors returns the list of parsing errors encountered.
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken advances the tokens.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EOF) {
		for p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
		}
		if p.curTokenIs(lexer.EOF) {
			break
		}

		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		} else {
			break
		}
		p.nextToken()
	}
	return program
}

// parseDecorators handles a sequence of one or more decorator expressions.
func (p *Parser) parseDecorator() ast.Statement {
	decorators := []ast.Expression{}

	// Loop through all consecutive decorators
	for p.curTokenIs(lexer.AT) {
		p.nextToken() // Consume '@'

		// Parse the full decorator expression with LOWEST precedence
		// This allows the full expression including function calls to be parsed
		decorator := p.parseExpression(LOWEST)
		if decorator == nil {
			return nil
		}
		decorators = append(decorators, decorator)

		// Move to the next token after the decorator expression
		p.nextToken()

		// Skip any newlines after the decorator
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
	}

	// After all decorators, check for async keyword
	isAsync := false
	if p.curTokenIs(lexer.ASYNC) {
		isAsync = true
		p.nextToken() // Consume 'async'
	}

	// Now we should be on 'def' or 'class'
	if p.curTokenIs(lexer.FUNCTION) {
		defStmt := p.parseDefStatement(isAsync)
		if defStmt == nil {
			return nil
		}

		// Attach decorators to the function
		if letStmt, ok := defStmt.(*ast.LetStatement); ok {
			if funcLit, ok := letStmt.Value.(*ast.FunctionLiteral); ok {
				funcLit.Decorators = decorators
			}
		}
		return defStmt
	}

	if p.curTokenIs(lexer.CLASS) {
		classStmt := p.parseClassStatement()
		if classStmt == nil {
			return nil
		}

		// Attach decorators to the class
		if cs, ok := classStmt.(*ast.ClassStatement); ok {
			cs.Decorators = decorators
		}
		return classStmt
	}

	// If we get here, we didn't find def or class after decorators
	p.errorExpected(constants.ParserExpectedDefOrClassAfterDecorator, p.curToken.String())
	return nil
}

// 2. Fixed parseStatement function
func (p *Parser) parseStatement() ast.Statement {
	// Skip any meaningless newlines that might appear between statements,
	// which is common with blank lines inside indented blocks.
	for p.curTokenIs(lexer.NEWLINE) {
		// p.nextToken()
		return p.parsePassStatement()
	}

	// Let the prefix function for '@' handle decorators
	if p.curToken.Type == lexer.AT {
		return p.parseDecorator()
	}

	// The rest of the function is the original logic for non-decorated statements
	if p.curTokenIs(lexer.ASYNC) {
		if !p.peekTokenIs(lexer.FUNCTION) {
			p.peekErrorMsg(fmt.Sprintf(constants.ParserExpectedDefAfterAsync, p.peekToken.Type))
			return nil
		}
		p.nextToken() // Consume ASYNC
		return p.parseDefStatement(true)
	}

	switch p.curToken.Type {
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.FUNCTION:
		return p.parseDefStatement(false)
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.PASS:
		return p.parsePassStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.FROM:
		return p.parseFromImportStatement()
	case lexer.WITH:
		return p.parseWithStatement()
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.RAISE:
		return p.parseRaiseStatement()
	case lexer.ASSERT:
		return p.parseAssertStatement()
	case lexer.GLOBAL:
		return p.parseGlobalStatement()
	case lexer.DEL:
		return p.parseDelStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseDelStatement() ast.Statement {
	stmt := &ast.DelStatement{Token: p.curToken}
	p.nextToken() // Consume 'del'

	stmt.Target = p.parseExpression(LOWEST)
	if stmt.Target == nil {
		return nil
	}

	// For now, we only need to support deleting from an index.
	// A full implementation would also allow `del var` and `del obj.attr`.
	switch stmt.Target.(type) {
	case *ast.IndexExpression:
		// This is a valid target.
	default:
		p.errors = append(p.errors, fmt.Sprintf("invalid deletion target at line %d", stmt.Token.Line))
		return nil
	}
	return stmt
}

func (p *Parser) parseGlobalStatement() *ast.GlobalStatement {
	stmt := &ast.GlobalStatement{Token: p.curToken}
	stmt.Names = []*ast.Identifier{}

	p.nextToken() // Consume 'global'

	// Parse the first identifier
	if !p.curTokenIs(lexer.IDENT) {
		p.errorExpected("identifier after 'global'", p.curToken.String())
		return nil
	}
	stmt.Names = append(stmt.Names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

	// Parse subsequent identifiers separated by commas
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // Consume IDENT
		p.nextToken() // Consume COMMA

		if !p.curTokenIs(lexer.IDENT) {
			p.errorExpected("identifier after comma in 'global' statement", p.curToken.String())
			return nil
		}
		stmt.Names = append(stmt.Names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	return stmt
}

// Add the new parsing function at the end of the file.
func (p *Parser) parseAssertStatement() *ast.AssertStatement {
	stmt := &ast.AssertStatement{Token: p.curToken} // curToken is ASSERT

	p.nextToken() // Consume 'assert'

	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		p.errors = append(p.errors, "expected condition after 'assert'")
		return nil
	}

	// Optional message
	if p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // Consume the condition's last token, curToken is now COMMA
		p.nextToken() // Consume COMMA, curToken is now start of message expression

		stmt.Message = p.parseExpression(LOWEST)
		if stmt.Message == nil {
			p.errors = append(p.errors, "expected message expression after comma in 'assert' statement")
			return nil
		}
	}

	// Consume optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTernaryExpression(valueIfTrue ast.Expression) ast.Expression {
	expr := &ast.TernaryExpression{
		Token:       p.curToken, // The 'if' token
		ValueIfTrue: valueIfTrue,
	}

	p.nextToken() // Consume 'if'

	// Parse the condition. Its precedence is low.
	expr.Condition = p.parseExpression(LOWEST)
	if expr.Condition == nil {
		return nil
	}

	// Expect an 'else' token after the condition
	if !p.expectPeek(lexer.ELSE) {
		return nil // Error already added by expectPeek
	}
	// curToken is now ELSE

	p.nextToken() // Consume 'else'

	// Parse the value_if_false expression.
	// Its precedence should be slightly lower than the TERNARY precedence
	// to handle chaining correctly (e.g., `a if b else c if d else e`).
	expr.ValueIfFalse = p.parseExpression(TERNARY - 1)
	if expr.ValueIfFalse == nil {
		return nil
	}

	return expr
}

// New parsing function for 'with' statements
func (p *Parser) parseWithStatement() *ast.WithStatement {
	stmt := &ast.WithStatement{Token: p.curToken} // curToken is WITH

	p.nextToken() // Consume WITH

	// Parse the context manager expression
	// The precedence for the context expression should be low enough to allow
	// complex expressions but not things like assignments without parentheses.
	// Let's use a precedence similar to what's used for 'if' conditions or 'return' values.
	stmt.ContextManager = p.parseExpression(LOWEST) // Or perhaps OR or ASSIGN-1
	if stmt.ContextManager == nil {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedExpressionAfterWith, p.curToken.Line, p.curToken.Column))
		return nil
	}

	// Check for optional 'as target_var'
	if p.peekTokenIs(lexer.AS) {
		p.nextToken() // Consume expression, curToken is AS
		p.nextToken() // Consume AS, curToken should be IDENT

		if !p.curTokenIs(lexer.IDENT) {
			p.errorExpected(constants.ParserIdentifierAfterAs, p.curToken.String())
			return nil
		}
		stmt.TargetVariable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		// p.curToken is now on the target IDENT
	}

	// Expect a colon
	// After parsing context_expr or target_var, p.curToken is on its last token.
	// So, we expectPeek for COLON.
	if !p.expectPeek(lexer.COLON) {
		// peekError already called by expectPeek
		return nil
	}
	// After expectPeek(COLON), p.curToken is COLON.

	// Expect an indented block for the body
	// Allow 'pass' on the same line: with cm: pass
	if p.peekTokenIs(lexer.PASS) && (p.l.PeekChar() == constants.NewlineRune || p.l.PeekChar() == 0 || p.l.PeekChar() == constants.HashRune) {
		p.nextToken() // Consume COLON, curToken is PASS
		passStmt := p.parsePassStatement()
		stmt.Body = &ast.BlockStatement{
			Token:      passStmt.Token,
			Statements: []ast.Statement{passStmt},
		}
	} else {
		if !p.expectPeek(lexer.INDENT) {
			p.errorExpectedNext(constants.ParserIndentedBlockForWithStatement, p.peekToken.String())
			return nil
		}
		// After expectPeek(INDENT), p.curToken is INDENT.
		stmt.Body = p.parseBlockStatement()
		if stmt.Body == nil {
			return nil
		}
		// parseBlockStatement leaves p.curToken on DEDENT.
	}
	// The main parseStatement loop will call nextToken() to consume the DEDENT
	// or the PASS token's implicit newline/semicolon.
	return stmt
}

func (p *Parser) parseTryStatement() ast.Statement {
	stmt := &ast.TryStatement{Token: p.curToken}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockTry, p.curToken.Line, p.curToken.Column+1))
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil
	}
	// After parseBlockStatement, curToken is on DEDENT

	hasHandlers := false
	stmt.Handlers = []*ast.ExceptHandler{}
	for p.peekTokenIs(lexer.EXCEPT) {
		hasHandlers = true
		p.nextToken() // Consume DEDENT, curToken is now EXCEPT
		handler := p.parseExceptHandler()
		if handler == nil {
			return nil
		}
		stmt.Handlers = append(stmt.Handlers, handler)
		// After parseExceptHandler, curToken is on its DEDENT
	}

	// --- THIS IS THE NEW LOGIC FOR `finally` ---
	hasFinally := false
	if p.peekTokenIs(lexer.FINALLY) {
		hasFinally = true
		p.nextToken() // Consume DEDENT, curToken is now FINALLY

		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			p.errorExpectedNext("indented block for 'finally' statement", p.peekToken.String())
			return nil
		}

		stmt.Finally = p.parseBlockStatement()
		if stmt.Finally == nil {
			return nil
		}
	}
	// --- END OF NEW LOGIC ---

	if !hasHandlers && !hasFinally {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedExceptOrFinally, p.curToken.Line, p.curToken.Column))
		return nil
	}
	return stmt
}

func (p *Parser) parseExceptHandler() *ast.ExceptHandler {
	handler := &ast.ExceptHandler{Token: p.curToken} // curToken is EXCEPT
	p.nextToken()                                    // Consume EXCEPT

	// If the current token is not COLON, then there must be a type expression.
	if !p.curTokenIs(lexer.COLON) {
		handler.Type = p.parseExpression(LOWEST)
		if handler.Type == nil {
			return nil
		} // Error in parsing type
	}

	// Now, after parsing the type (if any), check for an optional 'as' clause.
	// The 'AS' token would be the peek token.
	if p.peekTokenIs(lexer.AS) {
		if handler.Type == nil {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserBareExceptCannotUseAs, handler.Token.Line, handler.Token.Column))
			return nil
		}
		p.nextToken() // Consume the type expression's last token. curToken is now AS.
		p.nextToken() // Consume AS. curToken is now the variable name (IDENT).

		if !p.curTokenIs(lexer.IDENT) {
			p.errorExpected(constants.ParserIdentifierAfterAsExcept, p.curToken.String())
			return nil
		}
		handler.Var = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// After the optional type and 'as' clause, we expect a COLON.
	// We use expectPeek because the current token is on the last part of what was just parsed.
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Now expect an indented block.
	if !p.expectPeek(lexer.INDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockExcept, p.curToken.Line, p.curToken.Column+1))
		return nil
	}

	handler.Body = p.parseBlockStatement()
	if handler.Body == nil {
		return nil
	}
	return handler
}

func (p *Parser) parseRaiseStatement() *ast.RaiseStatement {
	stmt := &ast.RaiseStatement{Token: p.curToken}

	// Check if there is an expression to raise. A bare `raise` is followed
	// by a statement terminator like a newline or semicolon.
	// We check if the next token is NOT a terminator.
	switch p.peekToken.Type {
	case lexer.EOF, lexer.NEWLINE, lexer.SEMICOLON, lexer.DEDENT:
		// This is a bare `raise`, so Exception remains nil.
	default:
		// There is an expression to parse.
		p.nextToken() // Consume the 'raise' token
		stmt.Exception = p.parseExpression(LOWEST)
		if stmt.Exception == nil {
			return nil // An error occurred parsing the exception expression
		}
	}

	// Consume optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Add peekErrorMsg helper if it doesn't exist, or refine peekError
func (p *Parser) peekErrorMsg(expected string) {
	msg := fmt.Sprintf(constants.ParserExpectedGotInstead,
		p.peekToken.Line, p.peekToken.Column, expected, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// Add/Update errorExpected helper
func (p *Parser) errorExpected(expected string, got string) {
	msg := fmt.Sprintf(constants.ParserExpectedGot, p.curToken.Line, p.curToken.Column, expected, got)
	p.errors = append(p.errors, msg)
}

// Add/Update errorExpectedNext helper
func (p *Parser) errorExpectedNext(expected string, got string) {
	msg := fmt.Sprintf(constants.ParserExpectedNextGotInstead, p.peekToken.Line, p.peekToken.Column, expected, got)
	p.errors = append(p.errors, msg)
}

// parseImportStatement handles `import module.submodule [as alias]`
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken} // curToken is IMPORT

	// Parse the dotted module path
	modulePath := p.parseDottedModulePath()
	if modulePath == "" {
		p.peekErrorMsg("module name after 'import'")
		return nil
	}
	// Note: p.curToken is now on the last IDENT of the module path.
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: modulePath}

	// --- THIS IS THE FIX ---
	// After parsing the module name, check if the *next* token is 'as'.
	if p.peekTokenIs(lexer.AS) {
		p.nextToken() // Consume last part of module name. curToken is now AS.
		p.nextToken() // Consume AS. curToken is now the alias identifier.

		if !p.curTokenIs(lexer.IDENT) {
			p.errorExpected("alias identifier after 'as'", p.curToken.String())
			return nil
		}
		// Create an Identifier node for the alias and attach it to the statement.
		stmt.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}
	// --- END OF FIX ---

	// Consume optional semicolon at the end of the statement.
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFromImportStatement() *ast.FromImportStatement {
	stmt := &ast.FromImportStatement{Token: p.curToken} // curToken is FROM

	// Parse the dotted module path (e.g., "concurrent.futures")
	modulePath := p.parseDottedModulePath()
	if modulePath == "" {
		return nil // Error already added by parseDottedModulePath
	}
	stmt.ModulePath = &ast.Identifier{Token: p.curToken, Value: modulePath}

	if !p.expectPeek(lexer.IMPORT) { // IMPORT keyword
		// An error is already added by expectPeek
		return nil
	}
	// After expectPeek(lexer.IMPORT), p.curToken is IMPORT.

	// Now, p.peekToken is what comes after "import"
	if p.peekTokenIs(lexer.STAR) {
		p.nextToken() // Consume IMPORT, p.curToken is STAR
		stmt.ImportAll = true
		// No names to parse. The next p.nextToken() in ParseProgram will consume STAR.
	} else if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // Consume IMPORT, p.curToken is LPAREN
		p.nextToken() // Consume LPAREN, p.curToken is the first IDENT inside, or RPAREN if empty.
		stmt.Names = p.parseImportNameList(lexer.RPAREN)
		if stmt.Names == nil && len(p.errors) > 0 { // Check if parseImportNameList failed
			return nil
		}
		// parseImportNameList leaves curToken on RPAREN if successful
		if !p.curTokenIs(lexer.RPAREN) {
			p.errorExpected(constants.ParserExpectedClosingParenImportList, p.curToken.String())
			return nil
		}
	} else if p.peekTokenIs(lexer.IDENT) {
		p.nextToken()                                     // Consume IMPORT, p.curToken is the first IDENT to import.
		stmt.Names = p.parseImportNameList(lexer.NEWLINE) // Using NEWLINE as a conceptual terminator
		if stmt.Names == nil && len(p.errors) > 0 {       // Check if parseImportNameList failed
			return nil
		}
		// For non-parenthesized lists, parseImportNameList leaves p.curToken
		// on the last IDENT of the list. The p.nextToken() in ParseProgram
		// will advance past it.
	} else {
		p.peekErrorMsg(constants.ParserExpectedParenStarOrIdentImport)
		return nil
	}

	return stmt
}

// parseDottedModulePath parses a dotted module path like "concurrent.futures" or "os.path"
// Returns the full dotted path as a string, or empty string on error
func (p *Parser) parseDottedModulePath() string {
	if !p.expectPeek(lexer.IDENT) { // First module name
		return ""
	}

	var pathParts []string
	pathParts = append(pathParts, p.curToken.Literal)

	// Continue parsing dots and identifiers
	for p.peekTokenIs(lexer.DOT) {
		p.nextToken() // Consume current IDENT, move to DOT
		if !p.curTokenIs(lexer.DOT) {
			p.errorExpected(constants.ParserExpectedDotInModulePath, p.curToken.String())
			return ""
		}

		if !p.expectPeek(lexer.IDENT) { // Next module part after dot
			p.errorExpected(constants.ParserExpectedIdentAfterDot, p.peekToken.String())
			return ""
		}

		pathParts = append(pathParts, p.curToken.Literal)
	}

	return strings.Join(pathParts, ".")
}

// parseImportNameList parses a list of 'name [as alias]' items.
//   - endToken: For parenthesized lists, this is lexer.RPAREN.
//     For non-parenthesized lists, this is lexer.NEWLINE (conceptual).
//   - On entry:
//   - If parenthesized: p.curToken is the first IDENT in the list, or RPAREN if the list is empty.
//   - If non-parenthesized: p.curToken is the first IDENT in the list.
//   - On successful exit:
//   - If parenthesized: p.curToken is on the `endToken` (RPAREN).
//   - If non-parenthesized: p.curToken is on the last IDENT parsed.
//
// parseImportNameList parses a list of names to import, with optional aliases
// Example: name1, name2 as alias2, name3
func (p *Parser) parseImportNameList(terminator lexer.TokenType) []*ast.ImportNamePair {
	var names []*ast.ImportNamePair

	// Handle empty list case
	if p.curTokenIs(terminator) {
		return names
	}

	// Parse first name
	if !p.curTokenIs(lexer.IDENT) {
		p.errorExpected(constants.ParserExpectedIdentInImportList, p.curToken.String())
		return nil
	}

	pair := &ast.ImportNamePair{
		OriginalName: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// Check for "as alias"
	if p.peekTokenIs(lexer.AS) {
		p.nextToken()                   // Move to AS
		if !p.expectPeek(lexer.IDENT) { // Move to alias identifier
			return nil
		}
		pair.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	names = append(names, pair)

	// Parse remaining names
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()                   // Move to COMMA
		if !p.expectPeek(lexer.IDENT) { // Move to next identifier
			return nil
		}

		pair := &ast.ImportNamePair{
			OriginalName: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		// Check for "as alias"
		if p.peekTokenIs(lexer.AS) {
			p.nextToken()                   // Move to AS
			if !p.expectPeek(lexer.IDENT) { // Move to alias identifier
				return nil
			}
			pair.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}

		names = append(names, pair)
	}

	// For parenthesized lists, we expect to end on the terminator
	// For non-parenthesized lists, we end on the last identifier
	if terminator == lexer.RPAREN {
		if !p.peekTokenIs(lexer.RPAREN) {
			p.peekErrorMsg(constants.ParserExpectedClosingParenImportList)
			return nil
		}
		p.nextToken() // Move to RPAREN
	}

	return names
}

func (p *Parser) parsePassStatement() *ast.PassStatement {
	stmt := &ast.PassStatement{Token: p.curToken}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)
	if stmt.ReturnValue == nil {
		return nil
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseClassStatement() ast.Statement {
	stmt := &ast.ClassStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedClassnameClass, p.peekToken.Line, p.peekToken.Column))
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// >>> PARSE SUPERCLASS LIST <<<
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // Consume IDENT (class name), curToken is LPAREN

		// The parseIdentifierList helper you already have is perfect for this.
		// It expects the current token to be LPAREN.
		stmt.Superclasses = p.parseIdentifierList(lexer.RPAREN)
		if stmt.Superclasses == nil && len(p.errors) > 0 { // Check if parsing superclasses failed
			return nil
		}
		// parseIdentifierList leaves curToken on RPAREN if successful
		if !p.curTokenIs(lexer.RPAREN) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedCloseParenSuperclassList, p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
	} else {
		stmt.Superclasses = []*ast.Identifier{} // No explicit superclasses
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	if !p.peekTokenIs(lexer.INDENT) {
		if p.peekTokenIs(lexer.PASS) && (p.l.PeekChar() == constants.NewlineRune || p.l.PeekChar() == 0 || p.l.PeekChar() == constants.HashRune) {
			p.nextToken() // Consume COLON, curToken is PASS
			passStmt := p.parsePassStatement()
			stmt.Body = &ast.BlockStatement{
				Token:      passStmt.Token,
				Statements: []ast.Statement{passStmt},
			}
			return stmt
		}
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockOrPassClass, p.peekToken.Line, p.peekToken.Column))
		return nil
	}

	p.nextToken() // Consume COLON, curToken is INDENT
	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil
	}

	return stmt
}

// Helper to parse a list of identifiers, e.g., for superclasses
// Ends when 'end' token is encountered. Assumes curToken is on the opening delimiter (e.g. LPAREN)
// when called, or on the first identifier if no delimiter.
// Leaves curToken on the 'end' token.
func (p *Parser) parseIdentifierList(end lexer.TokenType) []*ast.Identifier {
	list := []*ast.Identifier{}

	if p.peekTokenIs(end) { // e.g. ()
		p.nextToken() // Consume the 'end' token
		return list
	}

	p.nextToken() // Move to the first identifier

	for {
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIdentInList, p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		list = append(list, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken()          // Consume IDENT, curToken is COMMA
			p.nextToken()          // Consume COMMA, curToken is next IDENT
			if p.curTokenIs(end) { // Trailing comma: (Base1, Base2,)
				break
			}
		} else if p.peekTokenIs(end) {
			p.nextToken() // Consume IDENT, curToken is 'end' token
			break
		} else {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedCommaOrCloseParenIdList, p.peekToken.Line, p.peekToken.Column, end, p.peekToken.Type))
			return nil
		}
	}
	if !p.curTokenIs(end) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedCloseDelimiterIdList, p.curToken.Line, p.curToken.Column, end, p.curToken.Type))
		return nil
	}
	return list
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	expr := &ast.DotExpression{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIdentifierAfterDot, p.curToken.Line, p.curToken.Column))
		return nil
	}

	expr.Identifier = p.parseIdentifier().(*ast.Identifier)
	return expr
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	if !p.curTokenIs(lexer.INDENT) {
		return nil
	}
	blockToken := p.curToken
	p.nextToken()

	block := &ast.BlockStatement{Token: blockToken, Statements: []ast.Statement{}}

	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			if !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			} else {
				break
			}
			continue
		}
		p.nextToken()
	}

	if !p.curTokenIs(lexer.DEDENT) {
		return nil
	}
	return block
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.curToken}
	p.nextToken()

	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		return nil
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockIfCondition, p.curToken.Line, p.curToken.Column+1))
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()
	if stmt.Consequence == nil {
		return nil
	}

	stmt.ElifBlocks = []*ast.ElifBlock{}
	for p.peekTokenIs(lexer.ELIF) {
		p.nextToken()
		elifToken := p.curToken
		p.nextToken()
		condition := p.parseExpression(LOWEST)
		if condition == nil {
			return nil
		}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockElifCondition, p.curToken.Line, p.curToken.Column+1))
			return nil
		}

		body := p.parseBlockStatement()
		if body == nil {
			return nil
		}
		stmt.ElifBlocks = append(stmt.ElifBlocks, &ast.ElifBlock{Token: elifToken, Condition: condition, Consequence: body})
	}

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken()
		if !p.curTokenIs(lexer.ELSE) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserInternalErrorExpectedElse, p.curToken.Type))
			return nil
		}
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			p.errors[len(p.errors)-1] = fmt.Sprintf(constants.ParserExpectedIndentedBlockElse,
				p.curToken.Line, p.curToken.Column+1, string(p.peekToken.Type))
			return nil
		}
		stmt.Alternative = p.parseBlockStatement()
		if stmt.Alternative == nil {
			return nil
		}
	}
	return stmt
}

func (p *Parser) parseWhileStatement() ast.Statement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	p.nextToken()

	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		return nil
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockWhileCondition, p.curToken.Line, p.curToken.Column+1))
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	// Parse the variable(s) - could be single identifier or tuple unpacking
	variables := []*ast.Identifier{}
	variables = append(variables, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

	// Check for comma-separated variables (tuple unpacking)
	for p.peekToken.Type == lexer.COMMA {
		p.nextToken() // consume the comma
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		variables = append(variables, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	// Set the variable(s) on the statement
	if len(variables) == 1 {
		stmt.Variable = variables[0]
	} else {
		// For multiple variables, we need to store them as a tuple target
		stmt.Variables = variables
	}

	if !p.expectPeek(lexer.IN) {
		return nil
	}
	p.nextToken()

	stmt.Iterable = p.parseExpression(LOWEST)
	if stmt.Iterable == nil {
		return nil
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedIndentedBlockForStatement, p.curToken.Line, p.curToken.Column+1))
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf(constants.ParserCouldNotParseAsInteger, p.curToken.Line, p.curToken.Column, p.curToken.Literal, err)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf(constants.ParserCouldNotParseAsFloat, p.curToken.Line, p.curToken.Column, p.curToken.Literal, err)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}
func (p *Parser) parseFStringLiteral() ast.Expression {
	// The current token is the FSTRING token. Its Literal field contains the raw string content.
	fstringToken := p.curToken

	// We transform this into an AST that represents: format_str("raw_content")

	// 1. Create the identifier for the built-in function `format_str`.
	callName := &ast.Identifier{
		// We can reuse the token's location info for better error reporting.
		Token: lexer.Token{Type: lexer.IDENT, Literal: "format_str", Line: fstringToken.Line, Column: fstringToken.Column},
		Value: "format_str",
	}

	// 2. Create the string literal argument containing the f-string's content.
	stringArg := &ast.StringLiteral{
		Token: fstringToken, // The original FSTRING token
		Value: fstringToken.Literal,
	}

	// 3. Build the CallExpression node.
	callExpr := &ast.CallExpression{
		Token:     fstringToken, // The '(' is implicit, use the f-string token for location.
		Function:  callName,
		Arguments: []ast.Expression{stringArg},
	}

	return callExpr
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curTokenIs(lexer.TRUE),
	}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)
	if expression.Right == nil {
		return nil
	}
	return expression
}

// func (p *Parser) parseGroupedExpression() ast.Expression {
// 	p.nextToken()

// 	exp := p.parseExpression(LOWEST)
//     if exp == nil { return nil }

// 	if !p.expectPeek(lexer.RPAREN) {
// 		return nil
// 	}
// 	return exp
// }

// This is nearly identical to parseListComprehension, but creates a SetComprehension node.
func (p *Parser) parseSetComprehension(startToken lexer.Token, element ast.Expression) ast.Expression {
	sc := &ast.SetComprehension{
		Token:   startToken, // The '{' token
		Element: element,
	}

	if !p.curTokenIs(lexer.FOR) {
		p.errorExpected(constants.ParserListComprehensionFor, p.curToken.String())
		return nil
	}

	if !p.expectPeek(lexer.IDENT) { // Expect loop variable
		return nil
	}
	sc.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.IN) { // Expect `in` keyword
		return nil
	}

	p.nextToken() // Consume `in`, move to start of iterable expression

	sc.Iterable = p.parseExpression(LOWEST)
	if sc.Iterable == nil {
		return nil
	}

	// Check for optional `if` condition
	if p.peekTokenIs(lexer.IF) {
		p.nextToken() // Consume last token of iterable
		p.nextToken() // Consume `IF`
		sc.Condition = p.parseExpression(LOWEST)
		if sc.Condition == nil {
			return nil
		}
	}

	// Expect the closing brace `}`
	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return sc
}

// They should be in the state where COMMA has TUPLE_PRECEDENCE and
// parseTupleLiteralInfix is registered.
func (p *Parser) parseTupleLiteralInfix(left ast.Expression) ast.Expression {
	// The current token is COMMA.
	tuple := &ast.TupleLiteral{Token: p.curToken}

	p.nextToken() // Consume the COMMA token.

	// Parse the expression on the right of the comma.
	// The precedence is the same as the comma itself, which allows chaining (e.g., a, b, c).
	right := p.parseExpression(TUPLE_PRECEDENCE)
	if right == nil {
		return nil
	}

	elements := []ast.Expression{}

	// If the left side was already a tuple from a previous comma (e.g., parsing `c` in `a, b, c`),
	// we flatten its elements into our new tuple.
	if leftTuple, ok := left.(*ast.TupleLiteral); ok {
		elements = append(elements, leftTuple.Elements...)
	} else {
		// Otherwise, the left side is the first element of our new tuple.
		elements = append(elements, left)
	}

	// Also check if the right side was a tuple (e.g., `a, (b, c)`) and flatten if needed.
	if rightTuple, ok := right.(*ast.TupleLiteral); ok {
		elements = append(elements, rightTuple.Elements...)
	} else {
		// Add the right-hand side element.
		elements = append(elements, right)
	}

	tuple.Elements = elements
	return tuple
}

// ============================================================================
// CORRECTED DICTIONARY / SET PARSING LOGIC
// ============================================================================

// REPLACE your existing parseDictOrSetLiteral function with this one.
func (p *Parser) parseDictOrSetLiteral() ast.Expression {
	lbraceToken := p.curToken // Save LBRACE token

	// Handle empty dict: {}
	// In Python, {} always creates an empty dictionary. An empty set is created with set().
	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken() // consume LBRACE token (make curToken LBRACE)
		// p.nextToken() // consume RBRACE token (make curToken RBRACE)
		return &ast.DictLiteral{Token: lbraceToken, Pairs: make(map[ast.Expression]ast.Expression)}
	}

	isIndentedLiteral := p.peekTokenIs(lexer.INDENT)
	p.nextToken() // Consume LBRACE.

	if isIndentedLiteral {
		if !p.curTokenIs(lexer.INDENT) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserInternalErrorMalformedIndentLbrace, p.curToken.Line, p.curToken.Column))
			return nil
		}
		p.nextToken()
		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RBRACE) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedClosingBraceDictDedent, p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			return &ast.DictLiteral{Token: lbraceToken, Pairs: make(map[ast.Expression]ast.Expression)}
		}
	}

	if p.curTokenIs(lexer.RBRACE) {
		return &ast.DictLiteral{Token: lbraceToken, Pairs: make(map[ast.Expression]ast.Expression)}
	}

	// <<< FIX IS HERE: Parse with a higher precedence to avoid comma confusion >>>
	firstItem := p.parseExpression(OR)
	if firstItem == nil {
		return nil
	}

	// Decide what kind of literal it is based on the next token.
	if p.peekTokenIs(lexer.COLON) {
		// It's a dictionary
		p.nextToken() // to COLON
		p.nextToken() // to value start

		// <<< FIX IS HERE: Parse with a higher precedence >>>
		value := p.parseExpression(OR)
		if value == nil {
			return nil
		}

		dict := &ast.DictLiteral{Token: lbraceToken, Pairs: make(map[ast.Expression]ast.Expression)}
		dict.Pairs[firstItem] = value

		if isIndentedLiteral {
			return p.parseIndentedDictRemainder(dict)
		} else {
			return p.parseSingleLineDictRemainder(dict)
		}
	} else if p.peekTokenIs(lexer.FOR) {
		p.nextToken()
		return p.parseSetComprehension(lbraceToken, firstItem)
	} else {
		// It's a set
		set := &ast.SetLiteral{Token: lbraceToken, Elements: []ast.Expression{firstItem}}

		if isIndentedLiteral {
			return p.parseIndentedSetRemainder(set)
		} else {
			return p.parseSingleLineSetRemainder(set)
		}
	}
}

// REPLACE your existing parseIndentedDictRemainder function with this one.
func (p *Parser) parseIndentedDictRemainder(dict *ast.DictLiteral) ast.Expression {
	p.nextToken()

	for {
		for p.curTokenIs(lexer.INDENT) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RBRACE) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedClosingBraceDictDedent, p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			return dict
		}
		if p.curTokenIs(lexer.EOF) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnterminatedIndentedDictLiteral, p.curToken.Line, p.curToken.Column))
			return nil
		}

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			for p.curTokenIs(lexer.INDENT) {
				p.nextToken()
			}
			if p.curTokenIs(lexer.DEDENT) {
				continue
			}
			if p.curTokenIs(lexer.RBRACE) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedBraceIndentedDict, p.curToken.Line, p.curToken.Column))
				return nil
			}
		}

		// <<< THE FIX >>>
		key := p.parseExpression(OR)
		if key == nil {
			return nil
		}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		p.nextToken()

		// <<< THE FIX >>>
		value := p.parseExpression(OR)
		if value == nil {
			return nil
		}
		dict.Pairs[key] = value

		p.nextToken()
	}
}

func (p *Parser) parseSingleLineDictRemainder(dict *ast.DictLiteral) ast.Expression {
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(lexer.RBRACE) { // Trailing comma
			break
		}

		key := p.parseExpression(OR)
		if key == nil {
			return nil
		}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		p.nextToken()

		value := p.parseExpression(OR)
		if value == nil {
			return nil
		}
		dict.Pairs[key] = value
	}

	if p.curTokenIs(lexer.RBRACE) {
		return dict
	}
	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}
	return dict
}

// REPLACE your existing parseIndentedSetRemainder function with this one for consistency.
func (p *Parser) parseIndentedSetRemainder(set *ast.SetLiteral) ast.Expression {
	p.nextToken()

	for {
		for p.curTokenIs(lexer.INDENT) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RBRACE) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedClosingBraceSetDedent, p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			return set
		}
		if p.curTokenIs(lexer.EOF) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnterminatedIndentedSetLiteral, p.curToken.Line, p.curToken.Column))
			return nil
		}

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			for p.curTokenIs(lexer.INDENT) {
				p.nextToken()
			}
			if p.curTokenIs(lexer.DEDENT) {
				continue
			}
			if p.curTokenIs(lexer.RBRACE) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedBraceIndentedSet, p.curToken.Line, p.curToken.Column))
				return nil
			}
		}

		// Use higher precedence for set elements as well.
		element := p.parseExpression(OR)
		if element == nil {
			return nil
		}
		set.Elements = append(set.Elements, element)

		p.nextToken()
	}
}

func (p *Parser) parseSingleLineSetRemainder(set *ast.SetLiteral) ast.Expression {
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(lexer.RBRACE) { // Trailing comma
			break
		}

		element := p.parseExpression(OR)
		if element == nil {
			return nil
		}
		set.Elements = append(set.Elements, element)
	}

	if p.curTokenIs(lexer.RBRACE) {
		return set
	}
	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}
	return set
}

// parseExpression to handle INDENT tokens better in certain contexts
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// Skip leading INDENT tokens in expression contexts
	for p.curTokenIs(lexer.INDENT) {
		p.nextToken()
	}

	if p.curTokenIs(lexer.EOF) {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedEOFExpression, p.curToken.Line, p.curToken.Column))
		return nil
	}

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		// Don't treat IF as infix if it's followed by a colon (statement context)
		if p.peekTokenIs(lexer.IF) && p.isIfStatement() {
			return leftExp
		}

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
		if leftExp == nil {
			return nil
		}
	}
	return leftExp
}

// isIfStatement checks if the next IF token is part of a statement (not ternary)
func (p *Parser) isIfStatement() bool {
	if !p.peekTokenIs(lexer.IF) {
		return false
	}

	// Save current state
	savedLexer := *p.l
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Advance past IF
	p.nextToken() // now at IF
	p.nextToken() // now past IF

	// Look for colon after condition, accounting for nested expressions
	level := 0
	for !p.curTokenIs(lexer.EOF) {
		switch p.curToken.Type {
		case lexer.LPAREN, lexer.LBRACKET, lexer.LBRACE:
			level++
		case lexer.RPAREN, lexer.RBRACKET, lexer.RBRACE:
			level--
		case lexer.COLON:
			if level == 0 {
				// Restore state and return true (this is a statement)
				*p.l = savedLexer
				p.curToken = savedCur
				p.peekToken = savedPeek
				return true
			}
		case lexer.ELSE:
			if level == 0 {
				// Found ELSE before COLON, this is ternary
				*p.l = savedLexer
				p.curToken = savedCur
				p.peekToken = savedPeek
				return false
			}
		case lexer.NEWLINE, lexer.EOF:
			// Restore state and return false
			*p.l = savedLexer
			p.curToken = savedCur
			p.peekToken = savedPeek
			return false
		}
		p.nextToken()
	}

	// Restore state
	*p.l = savedLexer
	p.curToken = savedCur
	p.peekToken = savedPeek
	return false
}

// in pylearn/internal/parser/parser.go

func (p *Parser) parseFunctionParameters() ([]*ast.Parameter, *ast.Identifier, *ast.Identifier) {
	if !p.curTokenIs(lexer.LPAREN) {
		p.errorExpected(constants.ParserOpenParenParam, p.curToken.String())
		return nil, nil, nil
	}

	params := []*ast.Parameter{}
	var varArgParam *ast.Identifier
	var kwArgParam *ast.Identifier

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params, varArgParam, kwArgParam
	}

	p.nextToken()

	hasEncounteredDefault := false
	hasEncounteredVarArg := false

	for {
		if p.curTokenIs(lexer.RPAREN) {
			break
		}

		// <<< THIS IS THE CORE OF THE FIX >>>
		if p.curTokenIs(lexer.POW) { // Case for **kwargs
			starToken := p.curToken
			if kwArgParam != nil {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserOnlyOneKwargsParamAllowed, starToken.Line, starToken.Column))
				return nil, nil, nil
			}
			if !p.expectPeek(lexer.IDENT) {
				return nil, nil, nil
			}
			kwArgParam = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		} else if p.curTokenIs(lexer.ASTERISK) { // Case for *args
			starToken := p.curToken
			if !p.peekTokenIs(lexer.IDENT) {
				p.errorExpectedNext(constants.ParserListComprehensionLoopVar, p.peekToken.String())
				return nil, nil, nil
			}
			if varArgParam != nil {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserOnlyOneArgsParamAllowed, starToken.Line, starToken.Column))
				return nil, nil, nil
			}
			if kwArgParam != nil {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserArgsCannotAppearAfterKwargs, starToken.Line, starToken.Column))
				return nil, nil, nil
			}
			p.nextToken() // Consume ASTERISK
			varArgParam = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			hasEncounteredVarArg = true

		} else if p.curTokenIs(lexer.IDENT) { // Case for regular parameters
			if hasEncounteredVarArg || kwArgParam != nil {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserRegularParamCannotAppearAfterStarred, p.curToken.Line, p.curToken.Column, p.curToken.Literal))
				return nil, nil, nil
			}

			paramNameToken := p.curToken
			paramNode := &ast.Parameter{
				Token: paramNameToken,
				Name:  &ast.Identifier{Token: paramNameToken, Value: paramNameToken.Literal},
			}

			if p.peekTokenIs(lexer.ASSIGN) {
				p.nextToken() // curToken is ASSIGN
				p.nextToken() // curToken is start of default value expr
				paramNode.DefaultValue = p.parseExpression(OR)
				if paramNode.DefaultValue == nil {
					return nil, nil, nil
				}
				hasEncounteredDefault = true
			} else {
				if hasEncounteredDefault {
					p.errors = append(p.errors, fmt.Sprintf(constants.ParserNonDefaultArgFollowsDefault, paramNode.Token.Line, paramNode.Token.Column, paramNode.Name.Value))
					return nil, nil, nil
				}
			}
			params = append(params, paramNode)
		} else {
			p.errorExpected(constants.ParserIdentAsteriskDoubleAsteriskParam, p.curToken.String())
			return nil, nil, nil
		}

		// Logic for moving to the next parameter
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
			if p.peekTokenIs(lexer.RPAREN) {
				p.nextToken()
				break
			}
			p.nextToken()
		} else if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken()
			break
		} else {
			p.peekErrorMsg(constants.ParserCommaOrCloseParenAfterParam)
			return nil, nil, nil
		}
	}

	if !p.curTokenIs(lexer.RPAREN) {
		p.errorExpected(constants.ParserCloseParenParamList, p.curToken.String())
		return nil, nil, nil
	}
	return params, varArgParam, kwArgParam
}

// Modify parseDefStatement to accept isAsync boolean
func (p *Parser) parseDefStatement(isAsync bool) ast.Statement {
	// p.curToken is FUNCTION ('def') when this is called

	funcLit := &ast.FunctionLiteral{
		Token:   p.curToken, // Token is 'def'
		IsAsync: isAsync,    // <-- SET BASED ON ARGUMENT
	}

	if !p.expectPeek(lexer.IDENT) { // Function name
		return nil
	}
	funcLit.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) { // Opening parenthesis for parameters
		return nil
	}
	// After expectPeek(LPAREN), p.curToken is LPAREN

	params, varArgParam, kwArgParam := p.parseFunctionParameters() // This call expects curToken to be LPAREN
	if len(p.errors) > 0 && params == nil && varArgParam == nil && kwArgParam == nil {
		return nil
	}

	funcLit.Parameters = params
	funcLit.VarArgParam = varArgParam
	funcLit.KwArgParam = kwArgParam

	// parseFunctionParameters should leave p.curToken on RPAREN
	if !p.curTokenIs(lexer.RPAREN) { // Ensure we are on RPAREN after parsing params
		p.errorExpected(constants.ParserCloseParenParamList, p.curToken.String())
		return nil
	}

	if !p.expectPeek(lexer.COLON) { // Colon after parameters
		return nil
	}
	// After expectPeek(COLON), p.curToken is COLON

	// Handle function body (INDENT + statements + DEDENT)
	// Check for 'pass' on the same line: def foo(): pass
	if p.peekTokenIs(lexer.PASS) && (p.l.PeekChar() == constants.NewlineRune || p.l.PeekChar() == 0 || p.l.PeekChar() == constants.HashRune) {
		p.nextToken() // Consume COLON, curToken is PASS
		passStmt := p.parsePassStatement()
		funcLit.Body = &ast.BlockStatement{
			Token:      passStmt.Token, // Or funcLit.Token
			Statements: []ast.Statement{passStmt},
		}
	} else {
		if !p.expectPeek(lexer.INDENT) {
			p.errorExpectedNext(constants.ParserExpectedIndentedBlockFunctionBody, p.peekToken.String())
			return nil
		}
		// After expectPeek(INDENT), p.curToken is INDENT
		funcLit.Body = p.parseBlockStatement()
		if funcLit.Body == nil {
			return nil
		}
		// parseBlockStatement leaves p.curToken on DEDENT
	}

	// Create a LetStatement to assign the FunctionLiteral (which is an expression)
	// to the function's name in the environment.
	assignStmt := &ast.LetStatement{
		Token: funcLit.Name.Token, // Token of the function name identifier
		Name:  funcLit.Name,
		Value: funcLit, // The FunctionLiteral AST node itself
	}
	return assignStmt
}

// New method:
func (p *Parser) parseAwaitExpression() ast.Expression {
	expr := &ast.AwaitExpression{Token: p.curToken}
	p.nextToken() // Consume AWAIT
	// Python's 'await' has a precedence slightly lower than '**' and higher than other unary ops.
	// For simplicity, let's use a precedence like PRODUCT or slightly higher.
	expr.Expression = p.parseExpression(PRODUCT) // Or a specific AWAIT_PRECEDENCE
	if expr.Expression == nil {
		return nil
	}
	return expr
}

func (p *Parser) parseYieldExpression() ast.Expression {
	expr := &ast.YieldExpression{Token: p.curToken}

	// `yield` can be used with or without a value.
	// If the next token is a statement terminator, it's a bare `yield`.
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) || p.peekTokenIs(lexer.RPAREN) || p.peekTokenIs(lexer.EOF) {
		expr.Value = nil
	} else {
		p.nextToken() // Consume 'yield'
		// Parse the expression to be yielded with low precedence.
		expr.Value = p.parseExpression(LOWEST)
	}

	return expr
}

// parseExpressionList - modified to handle * and ** in call context.
func (p *Parser) parseExpressionList(end lexer.TokenType, isCallContext bool) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) { // e.g. func() or []
		p.nextToken() // consume 'end'
		return list
	}

	p.nextToken() // consume opening token or move to first element

	// First element
	if p.curTokenIs(end) { // e.g. call context with only a trailing comma: func(,) -> this is an error
		if isCallContext && end == lexer.RPAREN {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedTokenInArgList, p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		// For list literals like `[,]`, it might be valid or not depending on stricter parsing.
		// For now, let's assume empty list if first token is end token for non-call contexts.
		// This case is mainly for `p.nextToken()` above landing on `end` if list was `[ ]`.
		return list
	}

	list = append(list, p.parseSingleExpressionForList(isCallContext))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume expression end
		p.nextToken() // consume COMMA

		if p.curTokenIs(end) { // Trailing comma case: e.g. [1, 2,] or func(a, b,)
			break
		}
		list = append(list, p.parseSingleExpressionForList(isCallContext))
	}

	if !p.expectPeek(end) {
		return nil
	}
	return list
}

// parseSingleExpressionForList helper for parseExpressionList
func (p *Parser) parseSingleExpressionForList(isCallContext bool) ast.Expression {
	if isCallContext {
		if p.curTokenIs(lexer.ASTERISK) {
			starToken := p.curToken
			isKwUnpack := false
			if p.peekTokenIs(lexer.ASTERISK) { // Check for '**'
				p.nextToken() // Consume first *, curToken is now second *
				isKwUnpack = true
			}
			p.nextToken() // Consume * (or second *), curToken is now start of expression

			// Check if an expression follows the * or **
			if p.curTokenIs(lexer.COMMA) || p.curTokenIs(lexer.RPAREN) {
				errMsg := constants.IterableKeyword
				if isKwUnpack {
					errMsg = constants.MappingKeyword
				}
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedExpressionAfterStar, starToken.Line, starToken.Column, errMsg, starToken.Literal))
				return nil
			}

			value := p.parseExpression(OR) // Parse the expression to be unpacked
			if value == nil {
				return nil
			}
			return &ast.StarredArgument{Token: starToken, Value: value, IsKwUnpack: isKwUnpack}
		}
	}
	return p.parseExpression(LOWEST)
}

// parseCallExpression uses the modified parseExpressionList
// Modify parseExpressionList or create parseArgumentList for call expressions.
// Let's adapt parseExpressionList and how it's called by parseCallExpression.
// parseCallExpression now calls parseArgumentList
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	call := &ast.CallExpression{Token: p.curToken, Function: function} // p.curToken is LPAREN
	call.Arguments = p.parseArgumentList()                             // Call new/modified helper
	if call.Arguments == nil && len(p.errors) > 0 {
		return nil
	}
	return call
}

// parseArgumentList is specifically for parsing arguments within function calls.
// It handles positional arguments, *iterable, **mapping, and name=value.
// Assumes curToken is LPAREN when called.
// parseArgumentList is specifically for parsing arguments within function calls.
// It handles positional arguments, *iterable, **mapping, and name=value.
// Assumes curToken is LPAREN when called.
// IT NOW IGNORES INDENT/DEDENT TOKENS WITHIN THE PARENTHESES.
func (p *Parser) parseArgumentList() []ast.Expression {
	args := []ast.Expression{}

	// Check if the parentheses are immediately closed: ()
	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // Consume RPAREN
		return args
	}

	p.nextToken() // Move past LPAREN

	// Loop to consume arguments
	for {
		// Consume any leading INDENT/DEDENT tokens on new lines within the argument list
		for p.curTokenIs(lexer.INDENT) || p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
		}

		// If we hit the closing parenthesis, we're done with arguments.
		if p.curTokenIs(lexer.RPAREN) {
			break
		}
		// Handle EOF or other unexpected tokens
		if p.curTokenIs(lexer.EOF) {
			p.errorExpected(constants.ParserExpectedArgOrCloseParen, p.curToken.String())
			return nil
		}

		// Parse one argument (positional, keyword, starred)
		arg := p.parseOneArgumentFromList() // Renamed helper for clarity
		if arg == nil {
			// parseOneArgumentFromList should have added an error
			return nil
		}
		args = append(args, arg)

		// After parsing an argument, p.curToken is on the last token of that argument.
		// We need to look at peekToken for a comma or the closing parenthesis.

		// Consume any INDENT/DEDENT tokens *after* an argument and *before* a comma or RPAREN
		for p.peekTokenIs(lexer.INDENT) || p.peekTokenIs(lexer.DEDENT) {
			p.nextToken() // Consume the current argument's last token
			// Now curToken is INDENT/DEDENT, nextToken will consume it.
			// This loop structure might need refinement if p.curToken itself is INDENT/DEDENT after an expression.
			// Let's refine: consume current arg's last token IF NOT ALREADY INDENT/DEDENT.
			// The outer loop's INDENT/DEDENT skipper should handle most cases.
			// What's more robust is to skip INDENT/DEDENT *before* parsing an arg,
			// and *before* checking for comma/RPAREN.
		}
		// Refined: after parsing an arg, p.curToken is its last token.
		// Now, look at p.peekToken.

		// Skip INDENT/DEDENT before checking for comma or RPAREN
		p.skipIndentDedentPeeks()

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // Consume the argument's last token. curToken is now COMMA.
			p.nextToken() // Consume COMMA. curToken is now start of next argument or INDENT/DEDENT/RPAREN.

			// Handle trailing comma: e.g., func(a, b, )
			p.skipIndentDedentCurrent() // Skip any indents/dedents before the RPAREN
			if p.curTokenIs(lexer.RPAREN) {
				break
			}
		} else if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken() // Consume the argument's last token. curToken is now RPAREN.
			break
		} else {
			p.peekErrorMsg(constants.ParserCommaOrCloseParenAfterArg)
			return nil
		}
	} // End for loop

	// curToken should be RPAREN when the loop breaks
	if !p.curTokenIs(lexer.RPAREN) {
		p.errorExpected(constants.ParserCloseParenArgList, p.curToken.String()+constants.AfterLoopText)
		return nil
	}
	// Caller of parseCallExpression will advance past RPAREN via its expectPeek or nextToken.
	return args
}

// New helper to parse one argument from a list (call, list, tuple).
// Handles positional, keyword (if isCallContext), or starred.

func (p *Parser) parseOneArgumentFromList() ast.Expression {
	// Check for **kwarg first - now using POW token directly
	if p.curTokenIs(lexer.POW) { // Single POW token for **
		doubleStarToken := p.curToken
		p.nextToken() // curToken is now the expression for **

		// Ensure expression after ** is not closing paren or comma prematurely
		if p.curTokenIs(lexer.RPAREN) || p.curTokenIs(lexer.COMMA) {
			p.errorExpected(constants.ParserExpectedExpressionAfterDoubleStar, p.curToken.String())
			return nil
		}

		value := p.parseExpression(OR) // Precedence for the mapping expression
		if value == nil {
			return nil
		}
		return &ast.StarredArgument{Token: doubleStarToken, Value: value, IsKwUnpack: true}
	}
	// Check for **kwarg first (only in call context, though parseArgumentList is only for calls)
	if p.curTokenIs(lexer.ASTERISK) && p.peekTokenIs(lexer.ASTERISK) {
		doubleStarToken := p.curToken
		p.nextToken() // curToken is now the second ASTERISK
		p.nextToken() // curToken is now the expression for **
		// Ensure expression after ** is not closing paren or comma prematurely
		if p.curTokenIs(lexer.RPAREN) || p.curTokenIs(lexer.COMMA) {
			p.errorExpected(constants.ParserExpectedExpressionAfterDoubleStar, p.curToken.String())
			return nil
		}
		value := p.parseExpression(OR) // Precedence for the mapping expression, OR is lower than most operators
		if value == nil {
			return nil
		}
		return &ast.StarredArgument{Token: doubleStarToken, Value: value, IsKwUnpack: true}
	}

	// Check for *arg
	if p.curTokenIs(lexer.ASTERISK) {
		starToken := p.curToken
		p.nextToken() // curToken is now the expression for *
		if p.curTokenIs(lexer.RPAREN) || p.curTokenIs(lexer.COMMA) {
			p.errorExpected(constants.ParserExpectedExpressionAfterSingleStar, p.curToken.String())
			return nil
		}
		value := p.parseExpression(OR)
		if value == nil {
			return nil
		}
		return &ast.StarredArgument{Token: starToken, Value: value, IsKwUnpack: false}
	}

	// Check for keyword argument: IDENT = expr (only in call context)
	// This logic is specific to function calls. List/tuple literals don't have keyword arguments.
	// So, parseArgumentList is fine, but parseGenericExpressionList should NOT have this.
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.ASSIGN) {
		nameIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		assignToken := p.curToken
		p.nextToken() // Consume IDENT, curToken is ASSIGN
		p.nextToken() // Consume ASSIGN, curToken is start of value expression

		if p.curTokenIs(lexer.RPAREN) || p.curTokenIs(lexer.COMMA) {
			p.errorExpected(constants.ParserValueAfterKwargAssignment, p.curToken.String())
			return nil
		}
		// Precedence for keyword argument value: should be low enough to allow most things,
		// but higher than comma if parsing multiple. Python effectively parses it as one expression.
		valueExpr := p.parseExpression(TUPLE_PRECEDENCE) // Using OR, as it's right before ASSIGN
		if valueExpr == nil {
			return nil
		}
		return &ast.KeywordArgument{Token: assignToken, Name: nameIdent, Value: valueExpr}
	}

	// Otherwise, it's a positional argument/regular expression
	return p.parseExpression(TUPLE_PRECEDENCE) // For list items, tuple items, or positional call args.
	// Python allows `(x, y for y in z)` in tuples but not calls or lists.
	// That's a generator expression, which requires different precedence.
	// For now, LOWEST allows full expressions.
	// If parsing `arg1, arg2=val`, use precedence lower than ASSIGN for arg1,
	// e.g. p.parseExpression(ASSIGN -1) or p.parseExpression(OR).
	// Let's use OR to be safe side for comma-separated expressions.
	// return p.parseExpression(OR)
}

// Helper to skip INDENT/DEDENT tokens if they are the current token
func (p *Parser) skipIndentDedentCurrent() {
	for p.curTokenIs(lexer.INDENT) || p.curTokenIs(lexer.DEDENT) {
		p.nextToken()
	}
}

// Helper to skip INDENT/DEDENT tokens if they are the peek token
func (p *Parser) skipIndentDedentPeeks() {
	for p.peekTokenIs(lexer.INDENT) || p.peekTokenIs(lexer.DEDENT) {
		p.nextToken() // This effectively consumes curToken and makes peekToken the new curToken
		// then the loop re-evaluates.
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	// <<< ADD THIS LOGIC FOR RIGHT-ASSOCIATIVITY >>>
	// For right-associative operators like `**`, we parse the right-hand side
	// with a slightly lower precedence to allow chaining.
	if expression.Operator == constants.DoubleAsteriskOperator {
		precedence--
	}
	// <<< END OF NEW LOGIC >>>
	p.nextToken()

	expression.Right = p.parseExpression(precedence)
	if expression.Right == nil {
		return nil
	}
	return expression
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	// Check if the left side is a valid assignment target (l-value).
	// This validation now happens in the interpreter, which is more robust.

	stmt := &ast.AssignStatement{
		Token:    p.curToken,
		Target:   left, // <<< REVERT: Use single Target field. `left` will be an Identifier or TupleLiteral.
		Operator: p.curToken.Literal,
	}

	// For in-place operators, the target cannot be a tuple.
	if _, isTuple := left.(*ast.TupleLiteral); isTuple && stmt.Operator != "=" {
		p.errors = append(p.errors, fmt.Sprintf("SyntaxError: '%s' operator cannot be used with multiple targets", stmt.Operator))
		return nil
	}

	precedence := p.curPrecedence()
	p.nextToken()
	stmt.Value = p.parseExpression(precedence - 1)
	if stmt.Value == nil {
		return nil
	}
	// // =========================== START: Add this Debug Line ===========================
	// fmt.Printf("[DEBUG PARSER] Created *ast.AssignStatement for Target: '%s' and Value: '%s'\n", stmt.Target.String(), stmt.Value.String())
	// // =========================== END: Add this Debug Line ===========================

	return stmt
}

// This function now correctly and always delegates to parseIndexOrSlice.
// This function now correctly handles both simple index and all slice variations.
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	lBracketToken := p.curToken // The '[' token.

	p.nextToken() // Consume '['. curToken is now the start of the index/slice.

	// --- Check for an empty `start` in a slice, e.g., `[:stop]` ---
	if p.curTokenIs(lexer.COLON) {
		// This is definitely a slice, starting with a colon.
		p.nextToken() // Consume the first ':'.

		slice := &ast.SliceExpression{Token: lBracketToken, Left: left, Start: nil}

		// Parse the `stop` part. It's optional.
		if !p.curTokenIs(lexer.RBRACKET) && !p.curTokenIs(lexer.COLON) {
			slice.Stop = p.parseExpression(LOWEST)
		}

		// Check for the optional `step` part.
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // Move to ':'
			p.nextToken() // Consume the second ':'
			if !p.curTokenIs(lexer.RBRACKET) {
				slice.Step = p.parseExpression(LOWEST)
			}
		}

		// Handle the closing bracket
		if p.curTokenIs(lexer.RBRACKET) {
			return slice
		}
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}

		return slice
	}

	// --- We have an expression at the start. It could be `index` or `start`. ---
	startOrIndexExpr := p.parseExpression(LOWEST)

	// If the next token is ']', it was a simple index.
	if p.peekTokenIs(lexer.RBRACKET) {
		p.nextToken() // Move to ']'
		return &ast.IndexExpression{Token: lBracketToken, Left: left, Index: startOrIndexExpr}
	}

	// If the next token is ':', it's a slice.
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // Move to ':'
		p.nextToken() // Consume the ':'

		slice := &ast.SliceExpression{Token: lBracketToken, Left: left, Start: startOrIndexExpr}

		// Parse the `stop` part. It's optional - check if we're already at ] or :
		if !p.curTokenIs(lexer.RBRACKET) && !p.curTokenIs(lexer.COLON) {
			slice.Stop = p.parseExpression(LOWEST)
		}

		// Check for the optional `step` part.
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // Move to ':'
			p.nextToken() // Consume the second ':'
			if !p.curTokenIs(lexer.RBRACKET) {
				slice.Step = p.parseExpression(LOWEST)
			}
		}

		// Handle the closing bracket
		if p.curTokenIs(lexer.RBRACKET) {
			return slice
		}
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}

		return slice
	}

	// If we get here, the syntax is wrong
	p.peekError(lexer.RBRACKET)
	return nil
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf(constants.ParserExpectedNextGotInstead,
		p.peekToken.Line, p.peekToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	if t == lexer.EOF {
		msg := fmt.Sprintf(constants.ParserUnexpectedEOF, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return
	}
	msg := fmt.Sprintf(constants.ParserNoPrefixParseFn, p.curToken.Line, p.curToken.Column, t)
	p.errors = append(p.errors, msg)
}

type NilLiteral struct {
	Token lexer.Token
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return constants.NoneKeyword }

func (p *Parser) parseBytesLiteral() ast.Expression {
	literalStr := p.curToken.Literal
	actualBytes, err := unescapeBytesLiteral(literalStr)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf(constants.ParserInvalidBytesLiteral, p.curToken.Line, p.curToken.Column, err))
		return nil
	}
	return &ast.BytesLiteral{Token: p.curToken, Value: actualBytes}
}

// unescapeBytesLiteral correctly parses a full bytes literal string like b"..." or b'...'.
func unescapeBytesLiteral(s string) ([]byte, error) {
	if len(s) < 2 || !(s[0] == 'b' || s[0] == 'B') {
		return nil, fmt.Errorf(constants.ParserMissingBPrefixError, s)
	}

	quoteChar := s[1]
	isTriple := false
	contentStart := 2

	// Determine if it's a triple-quoted literal and validate the quotes
	if len(s) >= 6 && s[1] == quoteChar && s[2] == quoteChar {
		isTriple = true
		contentStart = 4
		if s[len(s)-3] != quoteChar || s[len(s)-2] != quoteChar || s[len(s)-1] != quoteChar {
			return nil, fmt.Errorf(constants.ParserUnterminatedTripleQuoteError, s)
		}
	} else {
		if s[len(s)-1] != quoteChar {
			return nil, fmt.Errorf(constants.ParserUnterminatedQuoteError, s)
		}
	}

	var content string
	if isTriple {
		content = s[contentStart : len(s)-3]
	} else {
		content = s[contentStart : len(s)-1]
	}

	var buf bytes.Buffer
	i := 0
	for i < len(content) {
		char := content[i]
		if char == '\\' {
			i++
			if i >= len(content) {
				return nil, fmt.Errorf(constants.ParserTrailingBackslashError)
			}
			escapeChar := content[i]
			switch escapeChar {
			case '\\':
				buf.WriteByte('\\')
			case '\'':
				buf.WriteByte('\'')
			case '"':
				buf.WriteByte('"')
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'v':
				buf.WriteByte('\v')
			case 'x':
				if i+2 >= len(content) {
					return nil, fmt.Errorf(constants.ParserHexNeedsTwoDigitsError)
				}
				hexStr := content[i+1 : i+3]
				val, err := strconv.ParseUint(hexStr, 16, 8)
				if err != nil {
					return nil, fmt.Errorf(constants.ParserInvalidHexEscape, hexStr)
				}
				buf.WriteByte(byte(val))
				i += 2
			// Add support for octal if needed in the future
			default:
				// Unknown escapes are preserved literally in Python bytes
				buf.WriteByte('\\')
				buf.WriteByte(escapeChar)
			}
			i++
		} else {
			if char > 127 {
				return nil, fmt.Errorf(constants.ParserBytesASCIIOnlyError)
			}
			buf.WriteByte(char)
			i++
		}
	}
	return buf.Bytes(), nil
}

func (p *Parser) parseListComprehension(startToken lexer.Token, element ast.Expression) ast.Expression {
	lc := &ast.ListComprehension{
		Token:   startToken, // Use the passed-in '[' token
		Element: element,
	}

	// We are here because we just parsed `element` and saw `for`.
	// Current token should be `FOR`.
	if !p.curTokenIs(lexer.FOR) {
		p.errorExpected(constants.ParserListComprehensionFor, p.curToken.String())
		return nil
	}

	if !p.expectPeek(lexer.IDENT) { // Expect loop variable
		return nil
	}
	lc.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.IN) { // Expect `in` keyword
		return nil
	}

	p.nextToken() // Consume `in`, move to start of iterable expression

	lc.Iterable = p.parseExpression(LOWEST)
	if lc.Iterable == nil {
		return nil
	}
	// After parsing iterable, p.curToken is on its last token.

	// Check for optional `if` condition
	if p.peekTokenIs(lexer.IF) {
		p.nextToken() // Consume last token of iterable, curToken is now `IF`
		p.nextToken() // Consume `IF`, curToken is now start of condition
		lc.Condition = p.parseExpression(LOWEST)
		if lc.Condition == nil {
			return nil
		}
	}

	// Finally, expect the closing bracket `]`
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return lc
}

// pylearn/internal/parser/parser.go

// ============================================================================
// CORRECTED LIST PARSING LOGIC
// ============================================================================

func (p *Parser) parseListLiteral() ast.Expression {
	lBracketToken := p.curToken

	if p.peekTokenIs(lexer.RBRACKET) {
		p.nextToken()
		return &ast.ListLiteral{Token: lBracketToken, Elements: []ast.Expression{}}
	}

	isIndentedLiteral := p.peekTokenIs(lexer.INDENT)
	p.nextToken() // Consume LBRACKET

	if isIndentedLiteral {
		if !p.curTokenIs(lexer.INDENT) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserInternalErrorMalformedIndentList, p.curToken.Line, p.curToken.Column))
			return nil
		}
		p.nextToken()
		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RBRACKET) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedClosingBracketListDedent, p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			return &ast.ListLiteral{Token: lBracketToken, Elements: []ast.Expression{}}
		}
	}
	if p.curTokenIs(lexer.RBRACKET) {
		return &ast.ListLiteral{Token: lBracketToken, Elements: []ast.Expression{}}
	}

	// <<< FIX IS HERE: Use TUPLE_PRECEDENCE >>>
	firstElement := p.parseExpression(TUPLE_PRECEDENCE)
	if firstElement == nil {
		return nil
	}

	if p.peekTokenIs(lexer.FOR) {
		p.nextToken()
		return p.parseListComprehension(lBracketToken, firstElement)
	}

	listLiteral := &ast.ListLiteral{Token: lBracketToken, Elements: []ast.Expression{firstElement}}

	if isIndentedLiteral {
		return p.parseIndentedListRemainder(listLiteral)
	} else {
		return p.parseSingleLineListRemainder(listLiteral)
	}
}

func (p *Parser) parseIndentedListRemainder(list *ast.ListLiteral) ast.Expression {
	p.nextToken()

	for {
		for p.curTokenIs(lexer.INDENT) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RBRACKET) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedClosingBracketListDedent, p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			return list
		}
		if p.curTokenIs(lexer.EOF) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnterminatedIndentedListLiteral, p.curToken.Line, p.curToken.Column))
			return nil
		}
		if p.curTokenIs(lexer.RBRACKET) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedBracketIndentedList, p.curToken.Line, p.curToken.Column))
			return nil
		}

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			for p.curTokenIs(lexer.INDENT) {
				p.nextToken()
			}
			if p.curTokenIs(lexer.DEDENT) {
				continue
			}
			if p.curTokenIs(lexer.RBRACKET) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedBracketAfterCommaList, p.curToken.Line, p.curToken.Column))
				return nil
			}
		}

		// <<< FIX IS HERE: Use TUPLE_PRECEDENCE >>>
		element := p.parseExpression(TUPLE_PRECEDENCE)
		if element == nil {
			return nil
		}
		list.Elements = append(list.Elements, element)

		p.nextToken()
	}
}

func (p *Parser) parseSingleLineListRemainder(list *ast.ListLiteral) ast.Expression {
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(lexer.RBRACKET) { // Trailing comma
			break
		}

		element := p.parseExpression(TUPLE_PRECEDENCE)
		if element == nil {
			return nil
		}
		list.Elements = append(list.Elements, element)
	}

	if p.curTokenIs(lexer.RBRACKET) {
		return list
	}
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}
	return list
}

// ============================================================================
// CORRECTED TUPLE PARSING LOGIC
// ============================================================================

func (p *Parser) parseGroupedOrTupleExpression() ast.Expression {
	lParenToken := p.curToken

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return &ast.TupleLiteral{Token: lParenToken, Elements: []ast.Expression{}}
	}

	isIndentedLiteral := p.peekTokenIs(lexer.INDENT)
	p.nextToken()

	if isIndentedLiteral {
		if !p.curTokenIs(lexer.INDENT) {
			return nil
		}
		p.nextToken()
		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RPAREN) {
				return nil
			}
			return &ast.TupleLiteral{Token: lParenToken, Elements: []ast.Expression{}}
		}
	}
	if p.curTokenIs(lexer.RPAREN) {
		return &ast.TupleLiteral{Token: lParenToken, Elements: []ast.Expression{}}
	}

	// <<< FIX IS HERE: Use TUPLE_PRECEDENCE >>>
	firstExpr := p.parseExpression(TUPLE_PRECEDENCE)
	if firstExpr == nil {
		return nil
	}

	if !isIndentedLiteral && !p.peekTokenIs(lexer.COMMA) {
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
		return firstExpr
	}

	elements := []ast.Expression{firstExpr}
	tupleLiteral := &ast.TupleLiteral{Token: lParenToken, Elements: elements}

	if isIndentedLiteral {
		return p.parseIndentedTupleRemainder(tupleLiteral)
	} else {
		return p.parseSingleLineTupleRemainder(tupleLiteral)
	}
}

func (p *Parser) parseIndentedTupleRemainder(tuple *ast.TupleLiteral) ast.Expression {
	p.nextToken()

	for {
		for p.curTokenIs(lexer.INDENT) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.DEDENT) {
			p.nextToken()
			if !p.curTokenIs(lexer.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserExpectedCloseParenTupleDedent, p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			return tuple
		}
		if p.curTokenIs(lexer.EOF) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnterminatedIndentedTupleLiteral, p.curToken.Line, p.curToken.Column))
			return nil
		}
		if p.curTokenIs(lexer.RPAREN) {
			p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedCloseParenIndentedTuple, p.curToken.Line, p.curToken.Column))
			return nil
		}

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			for p.curTokenIs(lexer.INDENT) {
				p.nextToken()
			}
			if p.curTokenIs(lexer.DEDENT) {
				continue
			}
			if p.curTokenIs(lexer.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedCommaIndentedTuple, p.curToken.Line, p.curToken.Column))
				return nil
			}
		}

		// <<< FIX IS HERE: Use TUPLE_PRECEDENCE >>>
		element := p.parseExpression(TUPLE_PRECEDENCE)
		if element == nil {
			if p.curTokenIs(lexer.COMMA) {
				p.errors = append(p.errors, fmt.Sprintf(constants.ParserUnexpectedCommaTuple, p.curToken.Line, p.curToken.Column))
			}
			return nil
		}
		tuple.Elements = append(tuple.Elements, element)
		p.nextToken()
	}
}

func (p *Parser) parseSingleLineTupleRemainder(tuple *ast.TupleLiteral) ast.Expression {
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(lexer.RPAREN) {
			break
		}

		nextElement := p.parseExpression(TUPLE_PRECEDENCE)
		if nextElement == nil {
			return nil
		}
		tuple.Elements = append(tuple.Elements, nextElement)
	}

	if p.curTokenIs(lexer.RPAREN) {
		return tuple
	}
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return tuple
}
