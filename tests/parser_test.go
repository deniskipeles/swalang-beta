package tests

import (
	"testing"
    "fmt"
	
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/parser"
)

// checkParserErrors utility
func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestAssignmentStatement(t *testing.T) {
	input := `myVar = 123`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(p.Errors()) > 0 {
		t.Logf("Input:\n%s", input)
		lDebug := lexer.New(input)
		t.Logf("Tokens:")
		for {
		   tok := lDebug.NextToken()
		   t.Logf("  %s", tok.String())
		   if tok.Type == lexer.EOF { break }
		}
		checkParserErrors(t, p) // Will fail test
   }

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

    // Check if it's an assignment structure (InfixExpression with =)
    assignExpr, ok := stmt.Expression.(*ast.InfixExpression)
    if !ok {
        t.Fatalf("stmt.Expression is not ast.InfixExpression. got=%T", stmt.Expression)
    }

    if assignExpr.Operator != "=" {
        t.Fatalf("assignExpr.Operator is not '='. got=%q", assignExpr.Operator)
    }

    ident, ok := assignExpr.Left.(*ast.Identifier)
    if !ok {
        t.Fatalf("assignExpr.Left is not ast.Identifier. got=%T", assignExpr.Left)
    }
    if ident.Value != "myVar" {
        t.Fatalf("ident.Value not 'myVar'. got=%s", ident.Value)
    }

	// Test the value part (Right side of assignment)
    testLiteralExpression(t, assignExpr.Right, 123)

}


func TestDefStatement(t *testing.T) {
    // *** FIX INPUT HERE ***
    input := "def myFunction(x):\n    return x\n" // Ensure newline
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    checkParserErrors(t, p) // Check errors *after* parsing

    // Add debug logging if errors persist
	if len(p.Errors()) > 0 {
		t.Logf("Input:\n%s", input)
		lDebug := lexer.New(input)
		t.Logf("Tokens:")
		for {
		   tok := lDebug.NextToken()
		   t.Logf("  %s", tok.String())
		   if tok.Type == lexer.EOF { break }
		}
		checkParserErrors(t, p) // Will fail test
    }
    if len(p.Errors()) != 0 {
         t.Logf("Input:\n%s", input)
         lDebug := lexer.New(input)
         t.Logf("Tokens:")
         for {
            tok := lDebug.NextToken()
            t.Logf("  %s", tok.String())
            if tok.Type == lexer.EOF { break }
         }
         // checkParserErrors already prints errors
         t.FailNow()
    }

    if len(program.Statements) != 1 {
        t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
    }

    stmt, ok := program.Statements[0].(*ast.LetStatement)
    if !ok {
        t.Fatalf("statement is not ast.LetStatement, got=%T", program.Statements[0])
    }

    if stmt.Name.Value != "myFunction" {
        t.Fatalf("function name is not 'myFunction', got=%s", stmt.Name.Value)
    }

    fn, ok := stmt.Value.(*ast.FunctionLiteral)
    if !ok {
        t.Fatalf("stmt.Value is not ast.FunctionLiteral, got=%T", stmt.Value)
    }

    if len(fn.Parameters) != 1 || fn.Parameters[0].Name.Value != "x" {
        t.Fatalf("expected one parameter 'x', got=%+v", fn.Parameters)
    }

    if fn.Body == nil {
         t.Fatalf("function body is nil")
    }
    if len(fn.Body.Statements) != 1 {
        t.Fatalf("function body should have 1 statement, got=%d", len(fn.Body.Statements))
    }
    // Check body content
    retStmt, ok := fn.Body.Statements[0].(*ast.ReturnStatement)
    if !ok {
         t.Fatalf("Body statement is not ReturnStatement, got=%T", fn.Body.Statements[0])
    }
    testIdentifier(t, retStmt.ReturnValue, "x")
}

// Helper to test literals
func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
    case string:
        return testIdentifier(t, exp, v) // Or string literal if testing that
    case bool:
        return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled: %T", exp)
	return false
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
    if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
        t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
        return false
    }
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}
    if ident.TokenLiteral() != value {
        t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
        return false
    }
	return true
}


func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
    bo, ok := exp.(*ast.BooleanLiteral)
    if !ok {
        t.Errorf("exp not *ast.BooleanLiteral. got=%T", exp)
        return false
    }
    if bo.Value != value {
        t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
        return false
    }
    expectedLiteral := fmt.Sprintf("%t", value)
    // Python uses capital True/False
    if value { expectedLiteral = "True" } else { expectedLiteral = "False" }

    if bo.TokenLiteral() != expectedLiteral {
        t.Errorf("bo.TokenLiteral not %s. got=%s", expectedLiteral, bo.TokenLiteral())
        return false
    }
    return true
}


// Add more parser tests for different statements and expressions