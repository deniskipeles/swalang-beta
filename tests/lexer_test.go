package tests

import (
	"testing"
	"github.com/deniskipeles/pylearn/internal/lexer"
)



func TestNextTokenSimple(t *testing.T) {
	input := `x = 10 + 5.0
"hello" True False None
if x > 5:
    print("yes") # comment
`
	// Corrected expectation list matching the new lexer output (21 tokens)
	expectedTokens := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string // Literal for non-indent/dedent/eof tokens
	}{
		{lexer.IDENT, "x"},         // 0
		{lexer.ASSIGN, "="},        // 1
		{lexer.INT, "10"},          // 2
		{lexer.PLUS, "+"},          // 3
		{lexer.FLOAT, "5.0"},       // 4
		{lexer.STRING, "hello"},    // 5
		{lexer.TRUE, "True"},       // 6
		{lexer.FALSE, "False"},     // 7
		{lexer.NIL, "None"},        // 8
		{lexer.IF, "if"},           // 9
		{lexer.IDENT, "x"},         // 10
		{lexer.GT, ">"},            // 11
		{lexer.INT, "5"},           // 12
		{lexer.COLON, ":"},         // 13
		{lexer.INDENT, ""},         // 14 <-- Correct indent
		{lexer.IDENT, "print"},     // 15 <-- Correctly expects IDENT for print
		{lexer.LPAREN, "("},        // 16 <-- Correct
		{lexer.STRING, "yes"},      // 17 <-- Correct
		{lexer.RPAREN, ")"},        // 18 <-- Correct
		{lexer.DEDENT, ""},         // 19 <-- Correct dedent
		{lexer.EOF, ""},            // 20 <-- Correct EOF
	}

	l := lexer.New(input)
	var actualTokens []lexer.Token

	for {
		tok := l.NextToken()
		actualTokens = append(actualTokens, tok)
		if tok.Type == lexer.EOF {
			break
		}
        if tok.Type == lexer.ILLEGAL {
             // Print actual tokens up to the illegal one for debugging
             t.Logf("Actual Tokens Before Illegal (%d):", len(actualTokens))
             for i, atok := range actualTokens { t.Logf("  [%d]: %s", i, atok.String()) }
             t.Fatalf("Lexer encountered ILLEGAL token during collection: %s", tok.String())
        }
	}

    // Compare lengths first
    if len(actualTokens) != len(expectedTokens) {
        t.Logf("Actual Tokens (%d):", len(actualTokens))
        for i, tok := range actualTokens { t.Logf("  [%d]: %s", i, tok.String()) }
        t.Logf("Expected Tokens (%d):", len(expectedTokens))
        for i, tok := range expectedTokens { t.Logf("  [%d]: %s %q", i, tok.expectedType, tok.expectedLiteral) }
        t.Fatalf("Wrong number of tokens. Got %d, expected %d", len(actualTokens), len(expectedTokens))
    }

    // Compare token by token
	for i, tt := range expectedTokens {
        tok := actualTokens[i] // Get the actual token collected

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q (Literal: %q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

        // Check literal only for relevant types
		if tok.Type != lexer.INDENT && tok.Type != lexer.DEDENT && tok.Type != lexer.EOF {
			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
					i, tt.expectedLiteral, tok.Literal)
			}
		}
	}
    // Test passes if lengths match and all tokens match type/literal
}
// Add more lexer tests for operators, keywords, errors etc.


func TestNextTokenMultilineString(t *testing.T) {
	input := `"""hello
world"""
x = '''line one
'line two'
"line three"'''
y = """\"\"\" escaped""" # Escaped inner quotes
z = """
""" # Empty multiline
`
	expectedTokens := []struct {
		expectedType lexer.TokenType
		expectedLiteral string
	}{
		{lexer.STRING, "hello\nworld"}, // 0
		{lexer.IDENT, "x"},             // 1
		{lexer.ASSIGN, "="},            // 2
		{lexer.STRING, "line one\n'line two'\n\"line three\""}, // 3
		{lexer.IDENT, "y"},             // 4
		{lexer.ASSIGN, "="},            // 5
		// CHANGE THIS LINE: Expect interpreted quotes, not literal backslashes
		{lexer.STRING, `""" escaped`},   // 6 <--- MODIFIED EXPECTATION
		{lexer.IDENT, "z"},             // 7
		{lexer.ASSIGN, "="},            // 8
		{lexer.STRING, "\n"},           // 9
		{lexer.EOF, ""},                // 10
	}

	l := lexer.New(input)
	for i, tt := range expectedTokens {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("Token type mismatch at index %d. Expected %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("Token literal mismatch at index %d. Expected %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
		if tok.Type == lexer.EOF {
			break
		}
	}
}

// func TestNextTokenMultilineComment(t *testing.T) {
// 	input := `
// x =1
// """ This is
// a multiline
// comment. """
// y =2
// ''' Another
// comment '''
// z =3
// `
// 	expectedTokens := []struct {
// 		expectedType lexer.TokenType
// 		expectedLiteral string
// 	}{
// 		{lexer.IDENT, "x"},
// 		{lexer.ASSIGN, "="},
// 		{lexer.INT, "1"},
// 		{lexer.STRING, " This is\na multiline\ncomment. "},
// 		{lexer.IDENT, "y"},
// 		{lexer.ASSIGN, "="},
// 		{lexer.INT, "2"},
// 		{lexer.STRING, " Another\ncomment "},
// 		{lexer.IDENT, "z"},
// 		{lexer.ASSIGN, "="},
// 		{lexer.INT, "3"},
// 		{lexer.EOF, ""},
// 	}

// 	l := lexer.New(input)
// 	for i, tt := range expectedTokens {
// 		tok := l.NextToken()
// 		if tok.Type != tt.expectedType {
// 			t.Errorf("Token type mismatch at index %d. Expected %v, got %v", i, tt.expectedType, tok.Type)
// 		}
// 		if tok.Literal != tt.expectedLiteral {
// 			t.Errorf("Token literal mismatch at index %d. Expected %q, got %q", i, tt.expectedLiteral, tok.Literal)
// 		}
// 		if tok.Type == lexer.EOF {
// 			break
// 		}
// 	}
// }


// lexer_test.go

func TestNextTokenMultilineComment(t *testing.T) {
	input := `
x = 1 /* This is a
multiline comment.
It spans several lines. */
y = 2 /* Another one */ z = 3
/* Final comment at EOF */` // Test comment right before EOF

	expectedTokens := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		// Leading newline is skipped, indentation is 0
		{lexer.IDENT, "x"},    // 0
		{lexer.ASSIGN, "="},   // 1
		{lexer.INT, "1"},      // 2
		// /* ... */ comment is skipped entirely
		{lexer.IDENT, "y"},    // 3
		{lexer.ASSIGN, "="},   // 4
		{lexer.INT, "2"},      // 5
		// /* ... */ comment is skipped entirely
		{lexer.IDENT, "z"},    // 6
		{lexer.ASSIGN, "="},   // 7
		{lexer.INT, "3"},      // 8
		// Final /* ... */ comment is skipped
		{lexer.EOF, ""},       // 9
	}

	l := lexer.New(input)
	var actualTokens []lexer.Token

	for {
		tok := l.NextToken()
		actualTokens = append(actualTokens, tok)
		if tok.Type == lexer.EOF {
			break
		}
		if tok.Type == lexer.ILLEGAL {
			t.Logf("Actual Tokens Before Illegal (%d):", len(actualTokens))
			for i, atok := range actualTokens { t.Logf("  [%d]: %s", i, atok.String()) }
			t.Fatalf("Lexer encountered ILLEGAL token during collection: %s", tok.String())
		}
	}

	// Compare lengths first
	if len(actualTokens) != len(expectedTokens) {
		t.Logf("Actual Tokens (%d):", len(actualTokens))
		for i, tok := range actualTokens { t.Logf("  [%d]: %s", i, tok.String()) }
		t.Logf("Expected Tokens (%d):", len(expectedTokens))
		for i, tok := range expectedTokens { t.Logf("  [%d]: %s %q", i, tok.expectedType, tok.expectedLiteral) }
		t.Fatalf("Wrong number of tokens. Got %d, expected %d", len(actualTokens), len(expectedTokens))
	}

	// Compare token by token
	for i, tt := range expectedTokens {
		tok := actualTokens[i] // Get the actual token collected

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q (Literal: %q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		// Check literal only for relevant types (excluding EOF as its literal is tested)
		if tok.Type != lexer.EOF && tok.Type != lexer.INDENT && tok.Type != lexer.DEDENT {
			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
					i, tt.expectedLiteral, tok.Literal)
			}
		} else if tok.Type == lexer.EOF && tok.Literal != tt.expectedLiteral {
             // Special check for EOF literal
             t.Fatalf("tests[%d] - wrong literal for EOF. expected=%q, got=%q",
                     i, tt.expectedLiteral, tok.Literal)
        }
	}
	// Test passes if lengths match and all tokens match type/literal
}


func TestNextTokenUnterminatedMultilineComment(t *testing.T) {
	input := `x = 1 /* Unterminated comment`

	expectedTokens := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
		expectedColumn int // Check column for the illegal token
	}{
		{lexer.IDENT, "x", 1},
		{lexer.ASSIGN, "=", 3},
		{lexer.INT, "1", 5},
		{lexer.ILLEGAL, "Unterminated multiline comment", 7}, // Error starts at column 7 (where '/*')
		// No EOF expected as lexing stops at ILLEGAL
	}

	l := lexer.New(input)
	for i, tt := range expectedTokens {
		tok := l.NextToken()

		if tok.Type == lexer.EOF {
			 t.Fatalf("tests[%d] - Unexpected EOF. Expected %s %q", i, tt.expectedType, tt.expectedLiteral)
		}

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q (Literal: %q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}

		// Check column specifically for the illegal token
		if tok.Type == lexer.ILLEGAL {
			if tok.Column != tt.expectedColumn {
				t.Fatalf("tests[%d] - wrong column for ILLEGAL token. expected=%d, got=%d",
					i, tt.expectedColumn, tok.Column)
			}
			// Stop after finding the expected illegal token
			return
		}
	}

	// Check if we reached here but expected an ILLEGAL token
	if len(expectedTokens) > 0 && expectedTokens[len(expectedTokens)-1].expectedType == lexer.ILLEGAL {
		t.Fatalf("Expected an ILLEGAL token but did not receive one.")
	}
}