package lexer

import (
	"bytes" // Needed for building string literals potentially containing escapes
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
)

// Lexer struct - No changes needed here
type Lexer struct {
	input        string // The source code being scanned
	position     int    // Current position in input (points to current char)
	readPosition int    // Current reading position in input (after current char)
	ch           rune   // Current char under examination
	line         int    // Current line number
	column       int    // Current column number - Tracks position *on the current line*

	// Indentation state
	indentStack   []int   // Stack of indentation levels (number of spaces)
	atLineStart   bool    // True if we are at the start of a logical line
	pendingTokens []Token // Queue for pending DEDENT/INDENT tokens
}

// New creates a new Lexer.
func New(input string) *Lexer {
	l := Lexer{
		input:       input,
		line:        1,
		column:      0, // Start column at 0, represents beginning of line
		indentStack: []int{0},
		atLineStart: true,
	}
	l.readChar() // Initialize l.ch
	return &l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
		l.position = l.readPosition
		return
	}

	r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
	l.position = l.readPosition // position is the start of the current rune
	l.ch = r
	l.readPosition += size // readPosition is the start of the *next* rune

	l.column += 1
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) PeekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) peekCharN(n int) rune {
	if n <= 0 {
		return l.ch // Peeking 0 chars is the current char
	}
	currentReadPos := l.readPosition
	for i := 1; i < n; i++ { // Advance read position simulation N-1 times
		if currentReadPos >= len(l.input) {
			return 0 // EOF before reaching Nth char
		}
		_, size := utf8.DecodeRuneInString(l.input[currentReadPos:])
		currentReadPos += size
	}
	// Now read the Nth character
	if currentReadPos >= len(l.input) {
		return 0 // EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[currentReadPos:])
	return r
}

// readSingleLineComment reads until newline or EOF
func (l *Lexer) readSingleLineComment() {
	for l.ch != constants.NewlineRune && l.ch != 0 {
		l.readChar()
	}
}

// readMultilineComment reads /* ... */ style comments
func (l *Lexer) readMultilineComment() Token {
	startLine := l.line
	startCol := l.column - 1 // The '/' was already read and column incremented

	l.readChar() // l.ch is now the character after '/*'

	for {
		if l.ch == 0 { // Unterminated comment
			return Token{Type: ILLEGAL, Literal: constants.LexerUnterminatedMultilineComment, Line: startLine, Column: startCol}
		}
		if l.ch == constants.AsteriskRune && l.peekChar() == constants.SlashRune {
			l.readChar() // Consume '*'
			l.readChar() // Consume '/'
			return Token{Type: COMMENT, Literal: constants.SlashAsteriskComment, Line: startLine, Column: startCol}
		}
		if l.ch == constants.NewlineRune {
			l.readChar()
			l.line++
			l.column = 0 // Reset column after newline
		} else {
			l.readChar()
		}
	}
}

// measureIndent calculates the indentation level, skipping carriage returns.
func (l *Lexer) measureIndent() int {
	indent := 0
	startColForError := l.column

	for {
		if l.ch == ' ' {
			indent++
			l.readChar()
		} else if l.ch == '\r' {
			l.readChar()
		} else {
			break
		}
	}

	if l.ch == '\t' {
		return -1 * (startColForError + indent + 1) // Negative indicates error, value is column
	}
	return indent
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isBinaryDigit(ch rune) bool {
	return ch == '0' || ch == '1'
}

// readPrefixedNumber reads hexadecimal (0x), octal (0o), and binary (0b) numbers.
func (l *Lexer) readPrefixedNumber() string {
	startPosition := l.position
	
	// Consume '0'
	l.readChar()
	prefix := l.ch
	
	// Consume 'x', 'o', or 'b'
	l.readChar()
	
	if prefix == 'x' || prefix == 'X' {
		for isHexDigit(l.ch) {
			l.readChar()
		}
	} else if prefix == 'o' || prefix == 'O' {
		for isOctalDigit(l.ch) {
			l.readChar()
		}
	} else if prefix == 'b' || prefix == 'B' {
		for isBinaryDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[startPosition:l.position]
}

func (l *Lexer) NextToken() Token {
	// 1. Process pending DEDENT tokens
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	// 2. Loop to skip whitespace, comments, handle newlines, and process indentation
	for {
		if l.ch == '\r' {
			l.readChar()
			continue
		}
		if l.atLineStart {
			indentStartColumn := l.column      
			currentIndent := l.measureIndent() 

			if currentIndent < 0 { 
				return Token{Type: ILLEGAL, Literal: constants.LexerIndentationTabError, Line: l.line, Column: -currentIndent}
			}

			isSignificantLine := true
			isCommentLine := false
			if l.ch == constants.HashRune {
				isSignificantLine = false
				isCommentLine = true
			} else if l.ch == constants.SlashRune && l.peekChar() == constants.AsteriskRune { 
				isSignificantLine = false
				isCommentLine = true
			} else if l.ch == constants.NewlineRune { 
				isSignificantLine = false
			} else if l.ch == 0 { 
				isSignificantLine = false
				return l.handleEOFIndentation() 
			}

			if isSignificantLine {
				l.atLineStart = false 

				lastIndent := l.indentStack[len(l.indentStack)-1]

				if currentIndent > lastIndent {
					l.indentStack = append(l.indentStack, currentIndent)
					indentToken := Token{Type: INDENT, Literal: constants.IndentLiteral, Line: l.line, Column: 1}
					return indentToken 
				} else if currentIndent < lastIndent {
					for len(l.indentStack) > 1 && currentIndent < l.indentStack[len(l.indentStack)-1] {
						l.indentStack = l.indentStack[:len(l.indentStack)-1]
						dedentToken := Token{Type: DEDENT, Literal: constants.DedentLiteral, Line: l.line, Column: 1}
						l.pendingTokens = append(l.pendingTokens, dedentToken)
					}

					if len(l.indentStack) == 0 || currentIndent != l.indentStack[len(l.indentStack)-1] {
						return Token{Type: ILLEGAL, Literal: constants.LexerIndentationDedentError, Line: l.line, Column: indentStartColumn + 1}
					}

					if len(l.pendingTokens) > 0 {
						tok := l.pendingTokens[0]
						l.pendingTokens = l.pendingTokens[1:]
						return tok
					}
				} 
			} else { 
				if isCommentLine {
					if l.ch == constants.HashRune {
						l.readSingleLineComment()
					} else if l.ch == constants.SlashRune && l.peekChar() == constants.AsteriskRune {
						l.readChar()                             
						commentToken := l.readMultilineComment() 
						if commentToken.Type == ILLEGAL {
							return commentToken 
						}
						continue 
					}
				}
			}
		} 

		// 3. Skip non-leading horizontal whitespace
		for !l.atLineStart && (l.ch == constants.SpaceRune || l.ch == constants.TabRune || l.ch == constants.CarriageReturnRune) {
			l.readChar()
		}

		// 4. Handle Newlines
		if l.ch == constants.NewlineRune {
			l.readChar() 
			l.line++
			l.column = 0         
			l.atLineStart = true 
			continue             
		}

		// 5. Skip comments
		if l.ch == constants.HashRune {
			l.readSingleLineComment() 
			continue                  
		}
		if l.ch == constants.SlashRune && l.peekChar() == constants.AsteriskRune {
			startColComment := l.column 
			l.readChar()                
			commentToken := l.readMultilineComment()
			if commentToken.Type == ILLEGAL {
				commentToken.Column = startColComment
				return commentToken
			}
			continue
		}

		// 6. Handle EOF
		if l.ch == 0 {
			return l.handleEOFIndentation()
		}

		// 7. Found significant token character
		break 
	} 

	// --- Regular Token Lexing ---
	startLine := l.line
	startColumn := l.column

	var tok Token
	makeToken := func(typ TokenType, lit string) Token {
		return Token{Type: typ, Literal: lit, Line: startLine, Column: startColumn}
	}

	isBytes := false
	isFString := false
	isRaw := false

	// Check for string prefixes (f, b, r)
	for {
		switch l.ch {
		case 'f', 'F':
			if l.peekChar() == '\'' || l.peekChar() == '"' {
				isFString = true
				l.readChar()
				startColumn = l.column
				goto endPrefixLoop
			}
		case 'b', 'B':
			if l.peekChar() == '\'' || l.peekChar() == '"' {
				isBytes = true
				l.readChar()
				startColumn = l.column
				goto endPrefixLoop
			}
		case 'r', 'R':
			if l.peekChar() == '\'' || l.peekChar() == '"' {
				isRaw = true
				l.readChar()
				startColumn = l.column
				goto endPrefixLoop
			}
		}
		break
	}
	endPrefixLoop:

	switch l.ch {
	case constants.AssignOperatorRune:
		if l.peekChar() == constants.AssignOperatorRune {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(EQ, literal)
		} else {
			tok = makeToken(ASSIGN, string(l.ch))
		}
	case constants.PlusOperatorRune:
		if l.peekChar() == constants.AssignOperatorRune { 
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(PLUS_EQ, literal)
		} else {
			tok = makeToken(PLUS, string(l.ch))
		}
	case constants.MinusSignRune:
		if l.peekChar() == constants.AssignOperatorRune { 
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(MINUS_EQ, literal)
		} else {
			tok = makeToken(MINUS, string(l.ch))
		}
	case constants.AsteriskRune:
		if l.peekChar() == constants.AsteriskRune {
			ch := l.ch
			l.readChar() 
			literal := string(ch) + string(l.ch)
			tok = makeToken(POW, literal)
		} else {
			tok = makeToken(ASTERISK, string(l.ch)) 
		}
	case constants.SlashRune:
		if l.peekChar() == constants.SlashRune {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(FLOOR_DIV, literal)
		} else {
			tok = makeToken(SLASH, string(l.ch))
		}
	case constants.PercentRune:
		tok = makeToken(PERCENT, string(l.ch))
	case constants.BangRune:
		if l.peekChar() == constants.AssignOperatorRune {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(NOT_EQ, literal)
		} else {
			tok = makeToken(BANG, string(l.ch))
		}
	case constants.LessThanOpRune:
		if l.peekChar() == constants.AssignOperatorRune {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(LT_EQ, literal)
		} else if l.peekChar() == constants.LessThanOpRune { 
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(LSHIFT, literal)
		} else {
			tok = makeToken(LT, string(l.ch))
		}
	case constants.GreaterThanOpRune:
		if l.peekChar() == constants.AssignOperatorRune {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(GT_EQ, literal)
		} else if l.peekChar() == constants.GreaterThanOpRune { 
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(RSHIFT, literal)
		} else {
			tok = makeToken(GT, string(l.ch))
		}
	case constants.CommaRune:
		tok = makeToken(COMMA, string(l.ch))
	case constants.SemicolonRune:
		tok = makeToken(SEMICOLON, string(l.ch))
	case constants.ColonRune:
		tok = makeToken(COLON, string(l.ch))
	case constants.OpenParenRune:
		tok = makeToken(LPAREN, string(l.ch))
	case constants.CloseParenRune:
		tok = makeToken(RPAREN, string(l.ch))
	case constants.OpenBraceRune:
		tok = makeToken(LBRACE, string(l.ch))
	case constants.CloseBraceRune:
		tok = makeToken(RBRACE, string(l.ch))
	case constants.OpenBracketRune:
		tok = makeToken(LBRACKET, string(l.ch))
	case constants.CloseBracketRune:
		tok = makeToken(RBRACKET, string(l.ch))
	case constants.DotRune:
		tok = makeToken(DOT, string(l.ch))
	case constants.AtRune:
		tok = makeToken(AT, string(l.ch))
	case '&':
		tok = makeToken(BITWISE_AND, string(l.ch))
	case '|':
		tok = makeToken(BITWISE_OR, string(l.ch))
	case '^':
		tok = makeToken(BITWISE_XOR, string(l.ch))
	case '~':
		tok = makeToken(BITWISE_NOT, string(l.ch))

	// --- Updated String Handling ---
	case constants.DoubleQuoteRune, constants.SingleQuoteRune:
		quoteChar := l.ch
		stringStartCol := startColumn
		isTriple := false

		originalLiteralStartPos := l.position

		p1 := l.peekChar()
		p2 := l.peekCharN(2)

		if l.ch == p1 && l.ch == p2 {
			isTriple = true
			l.readChar()
			l.readChar()
			l.readChar() 
		} else {
			l.readChar() 
		}

		content, ok := l.readStringOrBytesContent(quoteChar, isTriple, isBytes, isRaw)

		originalLiteralEndPos := l.position
		originalLiteral := l.input[originalLiteralStartPos:originalLiteralEndPos]
		if isBytes {
			originalLiteral = constants.CharB + originalLiteral
		}

		if !ok { 
			tok = Token{Type: ILLEGAL, Literal: constants.LexerUnterminatedStringOrBytes, Line: startLine, Column: stringStartCol}
		} else {
			tokenType := STRING
			if isBytes {
				tokenType = BYTES
				tok = Token{Type: tokenType, Literal: originalLiteral, Line: startLine, Column: stringStartCol}
			} else if isFString { 
				tokenType = FSTRING
				tok = Token{Type: tokenType, Literal: content, Line: startLine, Column: stringStartCol}
			} else {
				tok = Token{Type: tokenType, Literal: content, Line: startLine, Column: stringStartCol}
			}
		}
		return tok

	default:
		if isLetter(l.ch) {
			identStartCol := startColumn
			literal := l.readIdentifier() 
			tokType := LookupIdent(literal)
			tok = Token{Type: tokType, Literal: literal, Line: startLine, Column: identStartCol}
			return tok 
		} else if isDigit(l.ch) {
			// Check for prefixed numbers (0x, 0X, 0o, 0O, 0b, 0B)
			if l.ch == '0' {
				p := l.peekChar()
				if p == 'x' || p == 'X' || p == 'o' || p == 'O' || p == 'b' || p == 'B' {
					numStartCol := startColumn
					literal := l.readPrefixedNumber() 
					tok = Token{Type: INT, Literal: literal, Line: startLine, Column: numStartCol}
					return tok 
				}
			}
			numStartCol := startColumn
			literal := l.readNumber() 
			tokType := INT
			for _, r := range literal {
				if r == constants.DotRune {
					tokType = FLOAT
					break
				}
			}
			tok = Token{Type: tokType, Literal: literal, Line: startLine, Column: numStartCol}
			return tok
		} else {
			tok = makeToken(ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) handleEOFIndentation() Token {
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	eofLine := l.line
	eofColumn := l.column 

	for len(l.indentStack) > 1 {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		dedentToken := Token{Type: DEDENT, Literal: constants.DedentLiteral, Line: eofLine, Column: 1}
		l.pendingTokens = append(l.pendingTokens, dedentToken)
	}

	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	eofToken := Token{Type: EOF, Literal: constants.EmptyString, Line: eofLine, Column: eofColumn} 
	return eofToken
}

func (l *Lexer) readIdentifier() string {
	startPosition := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	ident := l.input[startPosition:l.position]

	if ident == constants.NotKeyword && l.ch == constants.SpaceRune && l.peekChar() == constants.CharIRune && l.peekCharN(2) == constants.CharNRune {
		tempLexer := *l
		tempLexer.readChar()

		nextIdentStart := tempLexer.position
		for isLetter(tempLexer.ch) {
			tempLexer.readChar()
		}
		nextIdent := tempLexer.input[nextIdentStart:tempLexer.position]

		if nextIdent == constants.InKeyword {
			l.readChar()
			l.readChar()
			l.readChar()
			return constants.NotInKeyword
		}
	} else if ident == constants.IsKeyword && l.ch == constants.SpaceRune && l.peekChar() == constants.CharNRune && l.peekCharN(2) == constants.CharORune && l.peekCharN(3) == constants.CharTRune {
		tempLexer := *l
		tempLexer.readChar()
		nextIdentStart := tempLexer.position
		for isLetter(tempLexer.ch) {
			tempLexer.readChar()
		}
		nextIdent := tempLexer.input[nextIdentStart:tempLexer.position]

		if nextIdent == constants.NotKeyword {
			l.readChar()
			l.readChar()
			l.readChar()
			l.readChar()
			return constants.IsNotKeyword
		}
	}

	return ident
}

func LookupIdent(ident string) TokenType {
	if tokType, ok := keywords[ident]; ok {
		return tokType
	}
	if ident == constants.NotInKeyword {
		return NOT_IN
	}
	if ident == constants.IsNotKeyword {
		return IS_NOT
	}
	return IDENT
}

func (l *Lexer) readNumber() string {
	startPosition := l.position
	hasDecimal := false
	for isDigit(l.ch) || (l.ch == constants.DotRune && !hasDecimal && isDigit(l.peekChar())) {
		if l.ch == constants.DotRune {
			hasDecimal = true
		}
		l.readChar()
	}
	endPosition := l.position
	literal := l.input[startPosition:endPosition]
	return literal
}

func (l *Lexer) readStringOrBytesContent(quote rune, isTriple bool, isBytes bool, isRaw bool) (string, bool) {
	var out bytes.Buffer

	for {
		if l.ch == 0 {
			return constants.EmptyString, false
		}

		if l.ch == constants.BackslashRune { 
			if isRaw {
				// Raw strings preserve the backslash
				out.WriteRune(l.ch)
				l.readChar()
				if l.ch == 0 {
					return constants.EmptyString, false
				}
				// Write the next character (it might be the quote, escaping it from terminating the string)
				out.WriteRune(l.ch)
				l.readChar()
				continue
			}

			l.readChar()
			escapeChar := l.ch
			if escapeChar == 0 {
				return constants.EmptyString, false
			}

			validByteEscape := true
			var escapedRune rune = 0 
			handled := true

			switch escapeChar {
			case constants.CharNRune:
				escapedRune = constants.NewlineRune
			case constants.CharTRune:
				escapedRune = constants.TabRune
			case constants.CharRRune:
				escapedRune = constants.CarriageReturnRune
			case constants.CharBRune:
				escapedRune = constants.BackspaceRune 
			case constants.CharFRune:
				escapedRune = constants.FormFeedRune 
			case constants.BackslashRune:
				escapedRune = constants.BackslashRune
			case constants.SingleQuoteRune:
				escapedRune = constants.SingleQuoteRune
			case constants.DoubleQuoteRune:
				escapedRune = constants.DoubleQuoteRune
			case constants.CharZero, constants.CharOne, constants.CharTwo, constants.CharThree, constants.CharFour, constants.CharFive, constants.CharSix, constants.CharSeven: 
				if !isBytes {
					validByteEscape = false
					handled = false
				} else {
					validByteEscape = false
					handled = false
				}
			case constants.CharXRune:
				l.readChar() 
				h1 := l.ch
				l.readChar() 
				h2 := l.ch
				hexStr := string([]rune{h1, h2})
				byteVal, err := strconv.ParseUint(hexStr, 16, 8)
				if err != nil {
					out.WriteRune(constants.BackslashRune)
					out.WriteRune(constants.CharXRune)
					out.WriteRune(h1)
					out.WriteRune(h2)
					validByteEscape = false
					handled = false 
				} else {
					out.WriteByte(byte(byteVal))
				}
			case constants.NewlineRune: 
				l.line++
				l.column = 0
			case constants.CharURune, constants.CharUUpperRune, constants.CharNUpperRune:
				if isBytes {
					validByteEscape = false
				}
				handled = false 
			default: 
				if isBytes {
					validByteEscape = false
				} 
				handled = false 
			}

			if isBytes && !validByteEscape {
				handled = false
			}

			if handled && escapedRune != 0 { 
				out.WriteRune(escapedRune)
			} else if !handled { 
				out.WriteRune(constants.BackslashRune)
				out.WriteRune(escapeChar)
			}
			l.readChar() 

		} else if l.ch == quote { 
			if isTriple {
				if l.peekChar() == quote && l.peekCharN(2) == quote {
					l.readChar()
					l.readChar()
					l.readChar()              
					return out.String(), true 
				} else {
					out.WriteRune(l.ch)
					l.readChar() 
				}
			} else { 
				l.readChar()              
				return out.String(), true 
			}
		} else if l.ch == constants.NewlineRune { 
			if !isTriple && !isBytes { 
				return constants.EmptyString, false 
			}
			out.WriteRune(l.ch)
			l.readChar()
			l.line++
			l.column = 0
		} else { 
			out.WriteRune(l.ch)
			l.readChar()
		}
	}
}

func (l *Lexer) readString(quote rune, isTriple bool) (string, bool) {
	var out bytes.Buffer

	for {
		if l.ch == 0 { 
			return constants.EmptyString, false 
		}

		if l.ch == constants.BackslashRune { 
			l.readChar() 
			switch l.ch {
			case constants.CharNRune:
				out.WriteRune(constants.NewlineRune)
			case constants.CharTRune:
				out.WriteRune(constants.TabRune)
			case constants.BackslashRune:
				out.WriteRune(constants.BackslashRune)
			case constants.DoubleQuoteRune:
				out.WriteRune(constants.DoubleQuoteRune)
			case constants.SingleQuoteRune:
				out.WriteRune(constants.SingleQuoteRune)
			case 0: 
				return constants.EmptyString, false 
			case constants.NewlineRune: 
				l.line++
				l.column = 0
			default:
				out.WriteRune(constants.BackslashRune)
				out.WriteRune(l.ch)
			}
			l.readChar() 
		} else if l.ch == quote { 
			if isTriple {
				if l.peekChar() == quote && l.peekCharN(2) == quote {
					l.readChar()              
					l.readChar()              
					l.readChar()              
					return out.String(), true 
				} else {
					out.WriteRune(l.ch)
					l.readChar() 
				}
			} else { 
				l.readChar()              
				return out.String(), true 
			}
		} else if l.ch == constants.NewlineRune { 
			out.WriteRune(l.ch) 
			l.readChar()        
			l.line++
			l.column = 0 
		} else { 
			out.WriteRune(l.ch)
			l.readChar() 
		}
	}
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == constants.UnderscoreRune
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}