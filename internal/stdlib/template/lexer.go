// ===========================pylearn/internal/stdlib/template/lexer.go start here===========================
// internal/stdlib/template/lexer.go
package template

import (
	"fmt"
	"strings"
)

// TokenType identifies the type of lexed tokens.
type TokenType int

// Token types
const (
	TokenError TokenType = iota
	TokenEOF
	TokenText
	TokenVarStart   // {{
	TokenVarEnd     // }}
	TokenBlockStart // {%
	TokenBlockEnd   // %}
	TokenIdent
	TokenDot
	TokenPipe   // |
	TokenComma  // ,
	TokenLparen // (
	TokenRparen // )
	TokenInteger
)

// Token represents a token returned from the scanner.
type Token struct {
	Type TokenType
	Val  string
	Line int
}

// Delimiters - easily changeable
const (
	varStart   = "{{"
	varEnd     = "}}"
	blockStart = "{%"
	blockEnd   = "%}"
)

type lexState func(*Lexer) lexState

// Lexer holds the state of the scanner.
type Lexer struct {
	input  string
	pos    int
	start  int
	line   int
	state  lexState
	tokens chan Token
}

// NewLexer creates a new scanner for the input string.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		line:   1,
		state:  lexText,
		tokens: make(chan Token, 2), // Buffer for lookahead
	}
}

// Run lexes the input by executing state functions until the state is nil.
func (l *Lexer) Run() []Token {
	go func() {
		for l.state != nil {
			l.state = l.state(l)
		}
		close(l.tokens)
	}()

	var tokens []Token
	for t := range l.tokens {
		tokens = append(tokens, t)
	}
	return tokens
}

func (l *Lexer) emit(t TokenType) {
	l.tokens <- Token{t, l.input[l.start:l.pos], l.line}
	l.start = l.pos
}

func (l *Lexer) errorf(format string, args ...interface{}) lexState {
	l.tokens <- Token{TokenError, fmt.Sprintf(format, args...), l.line}
	return nil
}

// --- State Functions ---

func lexText(l *Lexer) lexState {
	for {
		if strings.HasPrefix(l.input[l.pos:], varStart) {
			if l.pos > l.start {
				l.emit(TokenText)
			}
			return lexInsideVar
		}
		if strings.HasPrefix(l.input[l.pos:], blockStart) {
			if l.pos > l.start {
				l.emit(TokenText)
			}
			return lexInsideBlock
		}
		if l.pos >= len(l.input) {
			break
		}
		if l.input[l.pos] == '\n' {
			l.line++
		}
		l.pos++
	}
	if l.pos > l.start {
		l.emit(TokenText)
	}
	l.emit(TokenEOF)
	return nil
}

func lexInsideVar(l *Lexer) lexState {
	l.pos += len(varStart)
	l.emit(TokenVarStart)
	return lexExpression(varEnd, TokenVarEnd)
}

func lexInsideBlock(l *Lexer) lexState {
	l.pos += len(blockStart)
	l.emit(TokenBlockStart)

	// Check for raw tag
	l.skipWhitespace()
	if strings.HasPrefix(l.input[l.pos:], "raw") {
		i := l.pos + 3
		if i >= len(l.input) || (l.input[i] == ' ' || strings.HasPrefix(l.input[i:], blockEnd)) {
			return lexRaw
		}
	}

	return lexExpression(blockEnd, TokenBlockEnd)
}

func lexRaw(l *Lexer) lexState {
	// We are inside a {% raw %} tag. Find {% endraw %}
	l.pos = l.start // Rewind to start of block
	endRawTag := blockStart + " endraw " + blockEnd
	endPos := strings.Index(l.input[l.pos:], endRawTag)

	if endPos == -1 {
		return l.errorf("unterminated 'raw' block")
	}

	// Find the start of {% raw %} and its closing %}
	rawBlockStart := strings.Index(l.input[l.pos:], blockEnd)
	if rawBlockStart == -1 {
		return l.errorf("unterminated 'raw' block start tag")
	}

	// The text is between the end of the start tag and the beginning of the end tag.
	textStart := l.pos + rawBlockStart + len(blockEnd)
	l.start = textStart
	l.pos = l.pos + endPos

	l.emit(TokenText)

	// Consume the {% endraw %} tag
	l.start = l.pos
	l.pos += len(endRawTag)
	l.start = l.pos

	return lexText
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		if strings.ContainsAny(" \t\r\n", string(l.input[l.pos])) {
			if l.input[l.pos] == '\n' {
				l.line++
			}
			l.pos++
		} else {
			break
		}
	}
	l.start = l.pos
}

func lexExpression(endDelimiter string, endToken TokenType) lexState {
	return func(l *Lexer) lexState {
		for {
			l.skipWhitespace() // Always skip leading whitespace first

			if strings.HasPrefix(l.input[l.pos:], endDelimiter) {
				l.pos += len(endDelimiter)
				l.emit(endToken)
				return lexText
			}

			if l.pos >= len(l.input) {
				return l.errorf("unterminated block or variable")
			}

			// Tokenize expression components
			switch r := l.input[l.pos]; {
			case r == '.':
				l.pos++
				l.emit(TokenDot)
			case r == '|':
				l.pos++
				l.emit(TokenPipe)
			case r == '(':
				l.pos++
				l.emit(TokenLparen)
			case r == ')':
				l.pos++
				l.emit(TokenRparen)
			case r == ',':
				l.pos++
				l.emit(TokenComma)
			case isAlpha(r):
				start := l.pos
				for l.pos < len(l.input) && isAlphaNumeric(l.input[l.pos]) {
					l.pos++
				}
				l.start = start
				l.emit(TokenIdent)
			case isDigit(r):
				start := l.pos
				for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
					l.pos++
				}
				l.start = start
				l.emit(TokenInteger)
			case r == '"' || r == '\'':
				l.pos++ // consume opening quote
				start := l.pos
				for l.pos < len(l.input) && l.input[l.pos] != byte(r) {
					l.pos++
				}
				l.start = start
				l.emit(TokenText) // The parser will treat this as a string literal
				if l.pos < len(l.input) {
					l.pos++ // consume closing quote
				}
				l.start = l.pos
			default:
				// Check for operators (e.g., >, <, ==)
				// This is a simplified check. A full engine might have more.
				opFound := false
				for _, op := range []string{">", "<", "==", "!=", ">=", "<="} {
					if strings.HasPrefix(l.input[l.pos:], op) {
						l.start = l.pos
						l.pos += len(op)
						l.emit(TokenIdent) // Treat operators as identifiers for simplicity
						opFound = true
						break
					}
				}
				if opFound {
					continue
				}

				return l.errorf("unexpected character inside block: %q", l.input[l.pos])
			}
		}
	}
}

func isDigit(r byte) bool        { return r >= '0' && r <= '9' }
func isAlpha(r byte) bool        { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' }
func isAlphaNumeric(r byte) bool { return isAlpha(r) || isDigit(r) }