Okay, let's start building the `pylearn` language, beginning with the lexer and the initial documentation.

**1. Project Setup (Directory Creation)**

First, ensure you have the base directory structure. You can create these directories using your terminal:

```bash
mkdir -p pylearn/cmd/interpreter
mkdir -p pylearn/cmd/compiler
mkdir -p pylearn/internal/lexer
mkdir -p pylearn/internal/parser
mkdir -p pylearn/internal/interpreter
mkdir -p pylearn/internal/compiler/vm
mkdir -p pylearn/internal/ast
mkdir -p pylearn/internal/types
mkdir -p pylearn/internal/builtins
mkdir -p pylearn/internal/utils
mkdir -p pylearn/examples
mkdir -p pylearn/tests
mkdir -p pylearn/docs
touch pylearn/go.mod # Initialize later with 'go mod init github.com/yourusername/pylearn'
touch pylearn/README.md
```

**2. Lexer Implementation**

Now, let's create the lexer files.

**`pylearn/internal/lexer/token.go`**

```go
package lexer

import "fmt"

// TokenType represents the type of a token.
type TokenType string

// Token represents a lexical token.
type Token struct {
	Type    TokenType // The type of the token (e.g., IDENT, INT, LPAREN)
	Literal string    // The literal value of the token (e.g., "myVar", "123", "(")
	Line    int       // Line number where the token starts
	Column  int       // Column number where the token starts
}

// String returns a string representation of the token, useful for debugging.
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Line: %d, Column: %d}",
		t.Type, t.Literal, t.Line, t.Column)
}

// Constants for token types
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL" // Represents a token/character we don't know about
	EOF     TokenType = "EOF"     // End Of File

	// Identifiers + Literals
	IDENT  TokenType = "IDENT"  // Variable names, function names, etc.
	INT    TokenType = "INT"    // Integers (e.g., 10, 0, -5)
	FLOAT  TokenType = "FLOAT"  // Floating-point numbers (e.g., 3.14, -0.5)
	STRING TokenType = "STRING" // String literals (e.g., "hello", 'world')
	BOOL   TokenType = "BOOL"   // Boolean literals (True, False)

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	BANG     TokenType = "!" // Used for 'not' prefix or potentially !=
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	PERCENT  TokenType = "%"

	LT     TokenType = "<"
	GT     TokenType = ">"
	EQ     TokenType = "=="
	NOT_EQ TokenType = "!="
	LT_EQ  TokenType = "<="
	GT_EQ  TokenType = ">="

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";" // Might not be used in Python-like syntax but good to have
	COLON     TokenType = ":"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{" // For dictionaries
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "[" // For lists
	RBRACKET  TokenType = "]"

	// Keywords
	FUNCTION TokenType = "FUNCTION" // def
	LET      TokenType = "LET"      // Potentially for variable declaration (or implicit like Python)
	TRUE     TokenType = "TRUE"     // True
	FALSE    TokenType = "FALSE"    // False
	IF       TokenType = "IF"       // if
	ELIF     TokenType = "ELIF"     // elif
	ELSE     TokenType = "ELSE"     // else
	RETURN   TokenType = "RETURN"   // return
	FOR      TokenType = "FOR"      // for
	WHILE    TokenType = "WHILE"    // while
	IN       TokenType = "IN"       // in
	AND      TokenType = "AND"      // and
	OR       TokenType = "OR"       // or
	NOT      TokenType = "NOT"      // not (can be operator or keyword)
	NIL      TokenType = "NIL"      // None
	PRINT    TokenType = "PRINT"    // print (treating as keyword initially)
)

// keywords maps keyword strings to their corresponding TokenType
var keywords = map[string]TokenType{
	"def":    FUNCTION,
	"True":   TRUE, // Note Python's capitalization
	"False":  FALSE,
	"if":     IF,
	"elif":   ELIF,
	"else":   ELSE,
	"return": RETURN,
	"for":    FOR,
	"while":  WHILE,
	"in":     IN,
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"None":   NIL,
	"print":  PRINT, // Initially a keyword, could become a builtin identifier later
	// "let":    LET, // Decide if explicit declaration is needed
}

// LookupIdent checks the keywords table to see whether the given identifier is a keyword.
// If it is, it returns the keyword's TokenType. Otherwise, it returns IDENT.
func LookupIdent(ident string) TokenType {
	if tokType, ok := keywords[ident]; ok {
		return tokType
	}
	// Check for boolean literals explicitly if not handled by keywords map (redundant if in map)
	// if ident == "True" || ident == "False" {
	// 	return BOOL
	// }
	return IDENT
}
```

**`pylearn/internal/lexer/lexer.go`**

```go
package lexer

import (
	"unicode"
	"unicode/utf8"
)

// Lexer holds the state of the scanner.
type Lexer struct {
	input        string // The source code being scanned
	position     int    // Current position in input (points to current char)
	readPosition int    // Current reading position in input (after current char)
	ch           rune   // Current char under examination
	line         int    // Current line number
	column       int    // Current column number
}

// New creates a new Lexer instance.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0} // Start column at 0, will increment on first readChar
	l.readChar()                                   // Initialize l.ch, l.position, l.readPosition
	return l
}

// readChar reads the next character and advances the lexer's position.
// It supports UTF-8 characters.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // 0 is the ASCII code for "NUL" character, signifies EOF
	} else {
		// Decode the next rune
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
		l.column += 1 // Increment column count
	}
}

// peekChar looks ahead at the next character without consuming it.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// NextToken scans the input and returns the next token.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespaceAndComments()

	// Store starting position of the token
	startLine := l.line
	startColumn := l.column

	switch l.ch {
	// Operators & Delimiters
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar() // Consume the second '='
			literal := string(ch) + string(l.ch)
			tok = Token{Type: EQ, Literal: literal}
		} else {
			tok = newToken(ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar() // Consume the '='
			literal := string(ch) + string(l.ch)
			tok = Token{Type: NOT_EQ, Literal: literal}
		} else {
			// In Python '!' isn't typically a standalone operator,
			// but 'not' keyword is used. We can reserve '!' or map it to NOT later.
			// For now, let's treat standalone '!' as ILLEGAL, or potentially NOT if needed.
			// Revisit based on language design decision. For now, BANG is fine.
			tok = newToken(BANG, l.ch)
			// Alternatively: tok = Token{Type: ILLEGAL, Literal: string(l.ch)}
		}
	case '*':
		tok = newToken(ASTERISK, l.ch)
	case '/':
		tok = newToken(SLASH, l.ch)
	case '%':
		tok = newToken(PERCENT, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: LT_EQ, Literal: literal}
		} else {
			tok = newToken(LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: GT_EQ, Literal: literal}
		} else {
			tok = newToken(GT, l.ch)
		}
	case ',':
		tok = newToken(COMMA, l.ch)
	case ';': // Keep for potential future use, though not Pythonic
		tok = newToken(SEMICOLON, l.ch)
	case ':':
		tok = newToken(COLON, l.ch)
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '{':
		tok = newToken(LBRACE, l.ch)
	case '}':
		tok = newToken(RBRACE, l.ch)
	case '[':
		tok = newToken(LBRACKET, l.ch)
	case ']':
		tok = newToken(RBRACKET, l.ch)

	// String Literals
	case '"', '\'': // Support both double and single quotes
		quoteChar := l.ch
		tok.Literal = l.readString(quoteChar)
		// Check if the string was terminated correctly
		if tok.Literal == "" && l.ch != quoteChar { // Unterminated string hit EOF
			tok.Type = ILLEGAL
			tok.Literal = "Unterminated string" // Error message
		} else {
			tok.Type = STRING
			l.readChar() // Consume the closing quote
		}


	// End Of File
	case 0:
		tok.Literal = ""
		tok.Type = EOF

	// Default case (Identifiers, Keywords, Numbers, Illegal characters)
	default:
		if isLetter(l.ch) {
			// Read identifier or keyword
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal) // Check if it's a keyword
			// Update position info before returning
			tok.Line = startLine
			tok.Column = startColumn
			return tok // Early return because readIdentifier advances position
		} else if isDigit(l.ch) {
			// Read number (integer or float)
			numLiteral := l.readNumber()
			tok.Type = INT // Assume INT initially
			// Check if it contains a '.' to classify as FLOAT
			for _, r := range numLiteral {
				if r == '.' {
					tok.Type = FLOAT
					break
				}
			}
			tok.Literal = numLiteral
			// Update position info before returning
			tok.Line = startLine
			tok.Column = startColumn
			return tok // Early return because readNumber advances position
		} else {
			// Unrecognized character
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	// Set position info for tokens created within the switch case
	if tok.Line == 0 {
		tok.Line = startLine
		tok.Column = startColumn
	}

	l.readChar() // Advance to the next character for the *next* call to NextToken()
	return tok
}

// newToken is a helper function to create a new Token.
func newToken(tokenType TokenType, ch rune) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

// skipWhitespace skips over whitespace characters.
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		if l.ch == '\n' {
			l.line++
			l.column = 0 // Reset column after newline
		}
		l.readChar()
	}
}

// readIdentifier reads an identifier (letters, digits, underscore) and returns it.
func (l *Lexer) readIdentifier() string {
	startPosition := l.position
	// Identifiers can start with a letter or underscore
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// readNumber reads a number (integer or float) and returns it.
func (l *Lexer) readNumber() string {
	startPosition := l.position
	hasDecimal := false
	for isDigit(l.ch) || (l.ch == '.' && !hasDecimal && isDigit(l.peekChar())) {
		if l.ch == '.' {
			hasDecimal = true
		}
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// readString reads a string literal enclosed in quotes.
// It expects the opening quote to have already been consumed *or* to be the current l.ch.
// It reads until the closing quote or EOF. It does *not* consume the closing quote.
func (l *Lexer) readString(quote rune) string {
	startPosition := l.position + 1 // Start after the opening quote
	for {
		l.readChar()
		// Basic escape sequence handling (can be expanded)
		if l.ch == '\\' && (l.peekChar() == quote || l.peekChar() == '\\') {
			l.readChar() // Consume the escaped character
			continue
		}
		// Check for closing quote or EOF
		if l.ch == quote || l.ch == 0 {
			break
		}
		// Handle newlines within strings if needed, or treat as error
		if l.ch == '\n' {
			l.line++
			l.column = 0
			// Decide if multiline strings are allowed without triple quotes
			// For simplicity now, let's allow them implicitly
		}
	}
	// Check if we reached EOF before closing quote
	if l.ch == 0 {
		// Return the portion read so far, NextToken will mark as ILLEGAL
		return l.input[startPosition:l.position]
	}
	return l.input[startPosition:l.position]
}


// readComment reads from '#' until the end of the line or EOF.
func (l *Lexer) readComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipWhitespaceAndComments skips whitespace and full-line comments.
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		if unicode.IsSpace(l.ch) {
			if l.ch == '\n' {
				l.line++
				l.column = 0 // Reset column
			}
			l.readChar()
		} else if l.ch == '#' {
			l.readComment()
			// After reading a comment, we might encounter whitespace or another comment,
			// so we continue the loop instead of breaking immediately.
			// If readComment stops at '\n', the next loop iteration handles it.
			// If readComment stops at EOF (l.ch == 0), the loop terminates.
		} else {
			break // Not whitespace or comment start
		}
	}
}

// isLetter checks if the rune is a letter or underscore (common in identifiers).
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit checks if the rune is a digit.
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

```

**3. Initial Documentation**

Now, create the documentation file.

**`pylearn/docs/language_overview.md`**

```markdown
# PyLearn Language Overview

## 1. Introduction

PyLearn is a simplified, Python-like programming language designed primarily for educational purposes. The goal is to provide a language that is easy to learn and understand, focusing on the fundamental concepts of programming, while also serving as a practical example for learning about language design and implementation (compilers and interpreters).

This project aims to implement PyLearn using the Go programming language, demonstrating techniques for lexing, parsing, abstract syntax tree (AST) construction, interpretation, and potentially compilation to bytecode and a virtual machine.

**Target Audience:** Beginners learning programming concepts, students learning about language implementation.

## 2. Design Principles

*   **Simplicity:** The syntax should be minimal and intuitive, borrowing heavily from Python's clean style. Avoid complex features not essential for understanding core programming ideas.
*   **Pythonic Feel:** Where possible, syntax and semantics should mirror Python 3 to leverage existing familiarity and provide a stepping stone to a widely-used language.
*   **Educational Focus:** The implementation itself should be clear, well-commented, and structured to facilitate learning about language construction. Error messages (in the interpreter/compiler) should aim to be helpful.
*   **Core Features:** Focus on implementing fundamental features first: basic types, variables, arithmetic/logical operators, control flow (if/else, loops), functions, and basic input/output.

## 3. Project Structure

The project follows a standard Go project layout, separating concerns into different packages:

```
pylearn/
├── cmd/              # Command-line executables (entry points)
│   ├── interpreter/  # Source for the PyLearn interpreter executable
│   └── compiler/     # Source for the PyLearn compiler executable (future)
├── internal/         # Internal packages (core implementation)
│   ├── lexer/        # Lexical analysis (Tokenization)
│   ├── parser/       # Syntax analysis (AST Generation)
│   ├── ast/          # Abstract Syntax Tree definitions
│   ├── interpreter/  # Tree-walking interpreter logic
│   ├── compiler/     # Compiler logic (Bytecode generation - future)
│   │   └── vm/       # Virtual Machine (Bytecode execution - future)
│   ├── types/        # Language type system definitions (future)
│   ├── builtins/     # Built-in functions (e.g., print, len)
│   ├── utils/        # Utility functions shared across packages
├── examples/         # Example PyLearn programs
├── tests/            # Unit and integration tests
├── docs/             # Project documentation
├── go.mod            # Go module file
├── go.sum            # Go checksum database
└── README.md         # Project README
```

## 4. Implementation Stages

The language will be built incrementally:

1.  **Lexer:** Convert source code into a stream of tokens. (Current Stage)
2.  **Parser:** Build an Abstract Syntax Tree (AST) from the token stream.
3.  **Interpreter:** Execute the program by walking the AST.
4.  **Compiler/VM (Optional Future):** Compile the AST to bytecode and run it on a custom VM.

## 5. Lexing (Tokenization)

Lexing, or lexical analysis, is the first phase of processing the source code. The lexer scans the input text character by character and groups sequences of characters into meaningful units called **tokens**.

*   **Implementation:** The lexer is implemented in the `internal/lexer` package.
    *   `token.go`: Defines the `TokenType` enum (representing the different kinds of tokens) and the `Token` struct (which holds the type, literal value, line, and column number).
    *   `lexer.go`: Defines the `Lexer` struct, which keeps track of the input string and the current scanning position. Its primary method is `NextToken()`, which reads the input and returns the next identified token.

*   **Token Types:** The following basic token types are defined:
    *   **`ILLEGAL`**: Represents any character or sequence the lexer doesn't recognize.
    *   **`EOF`**: Marks the end of the input file.
    *   **`IDENT`**: Identifiers (variable names, function names, e.g., `my_var`, `calculate`).
    *   **Literals**: Represent constant values.
        *   `INT`: Integers (e.g., `123`, `0`, `-10`).
        *   `FLOAT`: Floating-point numbers (e.g., `3.14`, `-0.01`).
        *   `STRING`: String literals (e.g., `"hello"`, `'world'`).
    *   **Operators**: Symbols performing operations (e.g., `+`, `-`, `*`, `/`, `%`, `=`, `==`, `!=`, `<`, `>`, `<=`, `>=`).
    *   **Delimiters**: Punctuation separating parts of the syntax (e.g., `(`, `)`, `{`, `}`, `[`, `]`, `,`, `:`).
    *   **Keywords**: Reserved words with special meaning in the language (e.g., `def`, `if`, `else`, `elif`, `for`, `while`, `return`, `True`, `False`, `None`, `and`, `or`, `not`, `in`, `print`).

*   **Whitespace and Comments:**
    *   Whitespace (spaces, tabs, newlines) is generally ignored by the lexer, except where significant for separating tokens or (potentially later) for indentation. Newlines are tracked to maintain line counts.
    *   Comments start with `#` and extend to the end of the line. They are skipped by the lexer.

*   **Process:** The `Lexer` reads the input character by character using `readChar()`. It uses `peekChar()` to look ahead when necessary (e.g., for two-character operators like `==`). Based on the current character, it consumes characters belonging to a single token (e.g., reading all digits for a number, reading all letters/digits/\_ for an identifier) and constructs a `Token` struct, which is then returned by `NextToken()`.
```

**Explanation and Design Choices:**

1.  **`token.go`:**
    *   `TokenType` is a `string` for easy debugging, though `int` constants can be slightly more performant. String representation is often clearer for educational purposes.
    *   The `Token` struct includes `Line` and `Column` which are crucial for useful error reporting later.
    *   Keywords are stored in a `map` (`keywords`) for efficient lookup (`LookupIdent`). Python keywords like `True`, `False`, `None` are included with their Python capitalization. `print` is included as a keyword initially for simplicity; it might be refactored later to be a standard identifier resolved to a built-in function.
    *   Operators like `==`, `!=`, `<=`, `>=` are defined as distinct token types.

2.  **`lexer.go`:**
    *   The `Lexer` struct holds the `input` string and positions (`position`, `readPosition`). `ch` stores the *current* character being examined (as a `rune` to support Unicode). `line` and `column` track the position for error messages.
    *   `readChar()` handles advancing the position and supports UTF-8 using `utf8.DecodeRuneInString`. It sets `ch` to `0` (rune value) to signify EOF.
    *   `peekChar()` allows looking ahead without consuming the character, essential for multi-character tokens.
    *   `NextToken()` is the main loop:
        *   It first calls `skipWhitespaceAndComments()` to ignore irrelevant characters.
        *   A `switch` statement efficiently handles single-character tokens.
        *   For potential two-character tokens (`==`, `!=`, etc.), it uses `peekChar()`.
        *   Identifiers and numbers are handled by helper functions (`readIdentifier`, `readNumber`) that consume multiple characters.
        *   `LookupIdent` is used after reading an identifier to distinguish keywords.
        *   String literals handle both `"` and `'` and include basic escape sequence handling (`\\`, `\"`, `\'`). Unterminated strings are marked `ILLEGAL`.
        *   Number parsing (`readNumber`) includes basic float detection (presence of a single '.'). More robust float parsing might be needed later.
        *   Unknown characters result in an `ILLEGAL` token.
    *   Helper functions (`isLetter`, `isDigit`, `skipWhitespace`, `readIdentifier`, `readNumber`, `readString`, `readComment`) break down the logic into manageable parts. `isLetter` includes `_` as valid in identifiers.
    *   Whitespace skipping also handles newline detection to increment the `line` counter and reset the `column`.
    *   Comment skipping reads until a newline or EOF.

3.  **`docs/language_overview.md`:**
    *   Clearly states the educational purpose and Python-like nature.
    *   Outlines the simple design goals.
    *   Explains the chosen directory structure and the role of each part.
    *   Details the lexing phase, explaining tokens, the lexer's role, whitespace/comment handling, and lists the defined token types. This connects the documentation directly to the code just written.

This provides a solid foundation for the lexer and the project documentation. The next step would be to implement the parser (`internal/parser/parser.go` and `internal/ast/ast.go`). Remember to run `go mod init <your-module-path>` (e.g., `go mod init github.com/yourusername/pylearn`) in the `pylearn` directory and potentially `go mod tidy` as you add dependencies later. You should also start writing tests for the lexer (`tests/lexer_test.go`).