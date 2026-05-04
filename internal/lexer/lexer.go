// pylearn/internal/lexer/lexer.go
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

// New ... - No changes needed here
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

// readChar ... - No changes needed here
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
		l.position = l.readPosition
		return
	}

	// Store position before reading the rune to know where it starts
	// (Though we don't strictly need prevPos for the simple column increment)
	// prevPos := l.position

	r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
	l.position = l.readPosition // position is the start of the current rune
	l.ch = r
	l.readPosition += size // readPosition is the start of the *next* rune

	// Simple column increment. Resetting happens when '\n' is consumed.
	l.column += 1

	// Note: The more complex calculateColumn method discussed in comments
	// was deemed unnecessary for this implementation. The simple increment
	// combined with resetting column to 0 on newline consumption is used.
}

// calculateColumn: Helper to determine the column based on the last newline
// This isn't strictly necessary if we reset column=0 on newline consumption,
// but can be useful for more complex scenarios or debugging.
// For this implementation, simply incrementing in readChar and resetting on \n
// in the main loop is sufficient. Let's stick to the simpler approach for now.
// We'll keep the column increment in readChar and reset in NextToken/readString.

// peekChar ... - No changes needed here
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) PeekChar() rune { // Renamed to PeekChar
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// peekCharN checks N characters ahead. Returns 0 if EOF is reached before N chars.
// Used for checking triple quotes and multiline comment starters.
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
	// startCol := l.column // Column where '#' started
	// startLine := l.line
	//##fmt.Printf("DEBUG LEXER: Reading single-line comment starting at L%d C%d\n", startLine, startCol)
	for l.ch != constants.NewlineRune && l.ch != 0 {
		l.readChar()
	}
	// Do not consume the newline/EOF, let the main loop handle it.
	//##fmt.Printf("DEBUG LEXER: Finished single-line comment ending at L%d C%d (next char %q)\n", l.line, l.column, l.ch)
}
// // readSingleLineComment reads until a newline character (\n or \r) or EOF.
// func (l *Lexer) readSingleLineComment() {
// 	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
// 		l.readChar()
// 	}
// }

// readMultilineComment reads /* ... */ style comments
func (l *Lexer) readMultilineComment() Token {
	startLine := l.line
	startCol := l.column - 1 // The '/' was already read and column incremented

	// Consume the '*'
	l.readChar() // l.ch is now the character after '/*'

	for {
		if l.ch == 0 { // Unterminated comment
			return Token{Type: ILLEGAL, Literal: constants.LexerUnterminatedMultilineComment, Line: startLine, Column: startCol}
		}
		if l.ch == constants.AsteriskRune && l.peekChar() == constants.SlashRune {
			l.readChar() // Consume '*'
			l.readChar() // Consume '/'
			//##fmt.Printf("DEBUG LEXER: Finished multiline comment ending at L%d C%d (next char %q)\n", l.line, l.column, l.ch)
			return Token{Type: COMMENT, Literal: constants.SlashAsteriskComment, Line: startLine, Column: startCol} // Or just return nil/skip indicator
		}
		if l.ch == constants.NewlineRune {
			// Consume newline, update line/col count
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
			// It's a carriage return. Skip it, but do not count it
			// as indentation. The subsequent '\n' will be handled by
			// the main NextToken loop.
			l.readChar()
		} else {
			// Not a space or carriage return, stop measuring.
			break
		}
	}

	if l.ch == '\t' {
		// Return error, use the column where the tab was found
		return -1 * (startColForError + indent + 1) // Negative indicates error, value is column
	}
	return indent
}

// Add this new helper function to the file.
func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// Add this new function to read the full hex number.
func (l *Lexer) readHexNumber() string {
	startPosition := l.position
	// Consume the '0' and the 'x'/'X'
	l.readChar()
	l.readChar()
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// --- *** REVISED NextToken *** ---
func (l *Lexer) NextToken() Token {
	// 1. Process pending DEDENT tokens
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		//##fmt.Printf("DEBUG LEXER: Returning pending token: %s\n", tok.String())
		return tok
	}

	// 2. Loop to skip whitespace, comments, handle newlines, and process indentation
	for {
		// Add a universal rule to always skip carriage returns. This makes the lexer
		// agnostic to line endings (LF vs CRLF).
		if l.ch == '\r' {
			l.readChar()
			continue // Restart the loop with the next character
		}
		// Handle indentation logic ONLY if we are at the logical start of a line
		if l.atLineStart {
			//##fmt.Printf("DEBUG LEXER: At line start (line %d, col %d, char %q)\n", l.line, l.column, l.ch)

			// A. Measure leading whitespace
			indentStartColumn := l.column      // Column where indent measurement begins
			currentIndent := l.measureIndent() // Consumes spaces, l.ch is now first non-space/non-tab

			if currentIndent < 0 { // Tab error, value is -column
				return Token{Type: ILLEGAL, Literal: constants.LexerIndentationTabError, Line: l.line, Column: -currentIndent}
			}

			// B. Check if line is blank, comment-only, or EOF after spaces
			isSignificantLine := true
			isCommentLine := false
			if l.ch == constants.HashRune { // Single-line comment
				isSignificantLine = false
				isCommentLine = true
				//##fmt.Printf("DEBUG LEXER: Comment-only line detected after indent %d.\n", currentIndent)
			} else if l.ch == constants.SlashRune && l.peekChar() == constants.AsteriskRune { // Multiline comment start
				isSignificantLine = false
				isCommentLine = true
				//##fmt.Printf("DEBUG LEXER: Multiline comment detected at line start after indent %d.\n", currentIndent)
			} else if l.ch == constants.NewlineRune { // Blank line
				isSignificantLine = false
				//##fmt.Printf("DEBUG LEXER: Blank line detected after indent %d.\n", currentIndent)
			} else if l.ch == 0 { // EOF after spaces
				isSignificantLine = false
				//##fmt.Printf("DEBUG LEXER: EOF found at line start processing (after indent %d)\n", currentIndent)
				return l.handleEOFIndentation() // Generate necessary DEDENTs before EOF
			}

			// C. Process Indentation ONLY for significant lines
			if isSignificantLine {
				l.atLineStart = false // Processing actual code now, clear the flag
				//##fmt.Printf("DEBUG LEXER: Significant line. Measured indent: %d (char is now %q)\n", currentIndent, l.ch)

				lastIndent := l.indentStack[len(l.indentStack)-1]

				if currentIndent > lastIndent {
					//##fmt.Printf("DEBUG LEXER: Indent detected (%d > %d)\n", currentIndent, lastIndent)
					l.indentStack = append(l.indentStack, currentIndent)
					// Column for INDENT is typically 1, representing the line change
					indentToken := Token{Type: INDENT, Literal: constants.IndentLiteral, Line: l.line, Column: 1}
					//##fmt.Printf("DEBUG LEXER: Returning token: %s\n", indentToken.String())
					return indentToken // Return INDENT immediately

				} else if currentIndent < lastIndent { // DEDENT case
					//##fmt.Printf("DEBUG LEXER: Dedent needed (%d < %d) Stack before pop: %v\n", currentIndent, lastIndent, l.indentStack)
					for len(l.indentStack) > 1 && currentIndent < l.indentStack[len(l.indentStack)-1] {
						l.indentStack = l.indentStack[:len(l.indentStack)-1]
						// Column for DEDENT is typically 1
						dedentToken := Token{Type: DEDENT, Literal: constants.DedentLiteral, Line: l.line, Column: 1}
						l.pendingTokens = append(l.pendingTokens, dedentToken)
					}
					//##fmt.Printf("DEBUG LEXER: Stack after pop: %v\n", l.indentStack)

					// Check for mismatch AFTER popping
					if len(l.indentStack) == 0 || currentIndent != l.indentStack[len(l.indentStack)-1] {
						//##fmt.Printf("DEBUG LEXER: Indentation dedent error (%d != stack top %d)\n", currentIndent, l.indentStack[len(l.indentStack)-1])
						// Report error at the column where the indentation started
						return Token{Type: ILLEGAL, Literal: constants.LexerIndentationDedentError, Line: l.line, Column: indentStartColumn + 1}
					}

					// Return first pending DEDENT now
					if len(l.pendingTokens) > 0 {
						tok := l.pendingTokens[0]
						l.pendingTokens = l.pendingTokens[1:]
						//##fmt.Printf("DEBUG LEXER: Returning first pending DEDENT token: %s\n", tok.String())
						return tok // Return immediately
					}
				} else { // currentIndent == lastIndent
					//##fmt.Printf("DEBUG LEXER: Indent matches stack top (%d == %d). Processing char %q\n", currentIndent, lastIndent, l.ch)
					// Indentation matches, proceed to lex l.ch
				}
				// If we reached here, break the skipping loop below to process l.ch.

			} else { // Blank or comment line - SKIP INDENTATION LOGIC
				//##fmt.Printf("DEBUG LEXER: Skipping indent logic for blank/comment line (char %q)\n", l.ch)
				if isCommentLine {
					if l.ch == constants.HashRune {
						l.readSingleLineComment() // Consumes comment until \n or EOF
					} else if l.ch == constants.SlashRune && l.peekChar() == constants.AsteriskRune {
						l.readChar()                             // Consume the '/' before calling
						commentToken := l.readMultilineComment() // Consumes /* ... */
						if commentToken.Type == ILLEGAL {
							return commentToken // Return unterminated comment error
						}
						// Successfully consumed multiline comment.
						// We need to continue the *outer loop* because the character
						// after the comment might be significant (e.g., EOF, newline, or even code
						// if the comment didn't end with a newline).
						// Crucially, we are STILL considered at the start of a logical line
						// if the comment ended exactly at a newline boundary or EOF.
						// If the comment ended mid-line, atLineStart should become false.

						// Let's simplify: after consuming a comment at the start of the line,
						// we just continue the loop. If the next char is \n, it's handled.
						// If it's EOF, it's handled. If it's code, the `atLineStart` logic
						// runs again on the next iteration. This seems safest.
						continue // Restart the outer loop
					}
					// readSingleLineComment leaves l.ch as \n or 0.
				}
				// For blank lines, l.ch is '\n'.
				// For comment lines ending in \n, l.ch is \n.
				// Keep l.atLineStart = true, the newline logic below will handle it.
				// Fall through to newline handling or EOF handling.
			}
		} // end if l.atLineStart

		// 3. Skip non-leading horizontal whitespace (only runs if not atLineStart)
		for !l.atLineStart && (l.ch == constants.SpaceRune || l.ch == constants.TabRune || l.ch == constants.CarriageReturnRune) {
			//##fmt.Printf("DEBUG LEXER: Skipping non-leading whitespace char %q\n", l.ch)
			l.readChar()
		}

		// 4. Handle Newlines (This runs after atLineStart logic)
		if l.ch == constants.NewlineRune {
			//##fmt.Printf("DEBUG LEXER: Handling newline (line %d -> %d)\n", l.line, l.line+1)
			l.readChar() // Consume the newline
			l.line++
			l.column = 0         // Reset column
			l.atLineStart = true // Mark start of next logical line
			continue             // Restart the loop to handle potential indentation on the new line
		}

		// 5. Skip comments (if not at the start of a line)
		if l.ch == constants.HashRune {
			//##fmt.Printf("DEBUG LEXER: Handling non-leading single-line comment\n")
			l.readSingleLineComment() // Reads until newline or EOF
			continue                  // Loop will continue, hitting newline or EOF next
		}
		if l.ch == constants.SlashRune && l.peekChar() == constants.AsteriskRune {
			//##fmt.Printf("DEBUG LEXER: Handling non-leading multiline comment\n")
			startColComment := l.column // Column of the '/'
			l.readChar()                // Consume '/' before calling
			commentToken := l.readMultilineComment()
			if commentToken.Type == ILLEGAL {
				// Ensure the column reported is where the /* started
				commentToken.Column = startColComment
				return commentToken
			}
			// Consumed the comment, continue skipping
			continue
		}

		// 6. Handle EOF (if encountered after skipping other things)
		if l.ch == 0 {
			//##fmt.Printf("DEBUG LEXER: EOF encountered after skipping loop\n")
			return l.handleEOFIndentation()
		}

		// 7. If we reach here, l.ch is the first significant character of a token. Break the loop.
		//##fmt.Printf("DEBUG LEXER: Breaking skip loop, ready to lex char %q at L%d C%d\n", l.ch, l.line, l.column)
		break // Exit the skipping loop
	} // End of the skipping loop

	// --- Regular Token Lexing (after skipping loop) ---
	startLine := l.line
	// Use the column where the significant character was first encountered
	startColumn := l.column

	//##fmt.Printf("DEBUG LEXER: Lexing regular token starting L%d, C%d (char %q)\n", startLine, startColumn, l.ch)

	var tok Token
	makeToken := func(typ TokenType, lit string) Token {
		return Token{Type: typ, Literal: lit, Line: startLine, Column: startColumn}
	}

	// --- BYTE FIX HERE: Check for 'b' or 'f' prefix BEFORE checking for string quotes ---
	isBytes := false
	isFString := false
	// Loop to handle multiple prefix characters like 'f', 'r', 'b', 'u'
	// This correctly handles f"", fr"", b"", etc.
	for {
		switch l.ch {
		case 'f', 'F':
			// If we see a quote next, this prefix is confirmed. Consume it and continue.
			if l.peekChar() == '\'' || l.peekChar() == '"' {
				isFString = true
				l.readChar() // Consume the 'f' or 'F'
				startColumn = l.column
				// After consuming, break the loop to process the quote.
				goto endPrefixLoop
			}
		case 'b', 'B':
			if l.peekChar() == '\'' || l.peekChar() == '"' {
				isBytes = true
				l.readChar() // Consume the 'b' or 'B'
				startColumn = l.column
				goto endPrefixLoop
			}
		// Add cases for 'r' (raw) or 'u' (unicode) prefixes here if you support them
		// case 'r', 'R': ...
		}
		// If the current character is not a valid prefix followed by a quote,
		// break the loop and let it be handled as an identifier or something else.
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
		if l.peekChar() == constants.AssignOperatorRune { // Check for +=
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(PLUS_EQ, literal)
		} else {
			tok = makeToken(PLUS, string(l.ch))
		}
	case constants.MinusSignRune:
		if l.peekChar() == constants.AssignOperatorRune { // Check for -=
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(MINUS_EQ, literal)
		} else {
			tok = makeToken(MINUS, string(l.ch))
		}
	// 	tok = makeToken(ASTERISK, string(l.ch)) // Note: Multiline comment /* handled above
	case constants.AsteriskRune:
		if l.peekChar() == constants.AsteriskRune {
			ch := l.ch
			l.readChar() // Consume the first '*'
			literal := string(ch) + string(l.ch)
			tok = makeToken(POW, literal)
		} else {
			tok = makeToken(ASTERISK, string(l.ch)) // Note: Multiline comment /* handled above
		}
	// case constants.SlashRune:
	// 	tok = makeToken(SLASH, string(l.ch)) // Note: Multiline comment /* handled above
	case constants.SlashRune:
		// --- THIS IS THE FIX ---
		if l.peekChar() == constants.SlashRune {
			ch := l.ch
			l.readChar() // Consume the first '/'
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
		} else if l.peekChar() == constants.LessThanOpRune { // <<< ADD THIS BLOCK
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(LSHIFT, literal)
		} else {
			tok = makeToken(LT, string(l.ch))
		}
		// if l.peekChar() == constants.AssignOperatorRune {
		// 	ch := l.ch
		// 	l.readChar()
		// 	literal := string(ch) + string(l.ch)
		// 	tok = makeToken(LT_EQ, literal)
		// } else {
		// 	tok = makeToken(LT, string(l.ch))
		// }
	case constants.GreaterThanOpRune:
		if l.peekChar() == constants.AssignOperatorRune {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(GT_EQ, literal)
		} else if l.peekChar() == constants.GreaterThanOpRune { // <<< ADD THIS BLOCK
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = makeToken(RSHIFT, literal)
		} else {
			tok = makeToken(GT, string(l.ch))
		}
		// if l.peekChar() == constants.AssignOperatorRune {
		// 	ch := l.ch
		// 	l.readChar()
		// 	literal := string(ch) + string(l.ch)
		// 	tok = makeToken(GT_EQ, literal)
		// } else {
		// 	tok = makeToken(GT, string(l.ch))
		// }
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

		// Check for triple quotes
		p1 := l.peekChar()
		p2 := l.peekCharN(2) // Peek second char ahead

		if l.ch == p1 && l.ch == p2 {
			isTriple = true
			//##fmt.Printf("DEBUG LEXER: Detected triple quote %q starting L%d C%d\n", quoteChar, startLine, stringStartCol)
			// Consume the three opening quotes
			l.readChar()
			l.readChar()
			l.readChar() // l.ch is now the first char *inside* the string
		} else {
			//##fmt.Printf("DEBUG LEXER: Detected single quote %q starting L%d C%d\n", quoteChar, startLine, stringStartCol)
			// Consume the single opening quote
			l.readChar() // l.ch is now the first char *inside* the string
		}

		// Read the string/bytes content
		// Pass quoteChar, isTriple AND isBytes flag
		content, ok := l.readStringOrBytesContent(quoteChar, isTriple, isBytes) // Renamed helper

		// Store the *original* literal including quotes and prefix
		originalLiteralEndPos := l.position
		originalLiteral := l.input[originalLiteralStartPos:originalLiteralEndPos]
		if isBytes {
			originalLiteral = constants.CharB + originalLiteral // Prepend 'b' back for the token literal
		}

		if !ok { // Unterminated
			tok = Token{Type: ILLEGAL, Literal: constants.LexerUnterminatedStringOrBytes, Line: startLine, Column: stringStartCol}
		} else {
			tokenType := STRING
			if isBytes {
				tokenType = BYTES
				// For BYTES token, Literal should store the source representation
				// Content is used by the parser later
				tok = Token{Type: tokenType, Literal: originalLiteral, Line: startLine, Column: stringStartCol}
			}else if isFString { 
				tokenType = FSTRING
				// For FSTRING, the literal is the raw content inside the quotes.
				tok = Token{Type: tokenType, Literal: content, Line: startLine, Column: stringStartCol}
			} else {
				// For STRING token, Literal can store the unescaped content
				tok = Token{Type: tokenType, Literal: content, Line: startLine, Column: stringStartCol}
			}
		}
		return tok // Return early

	// --- End Updated String Handling ---

	default:
		if isLetter(l.ch) {
			identStartCol := startColumn
			literal := l.readIdentifier() // Handles advancement
			tokType := LookupIdent(literal)
			tok = Token{Type: tokType, Literal: literal, Line: startLine, Column: identStartCol}
			//##fmt.Printf("DEBUG LEXER: Returning token: %s\n", tok.String())
			return tok // Return early as readIdentifier advanced
		} else if isDigit(l.ch) {
			// Check for hexadecimal prefix '0x' or '0X'
			if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
				numStartCol := startColumn
				literal := l.readHexNumber() // Use the new helper
				tok = Token{Type: INT, Literal: literal, Line: startLine, Column: numStartCol}
				return tok // Return early
			}
			numStartCol := startColumn
			literal := l.readNumber() // Handles advancement
			tokType := INT
			for _, r := range literal {
				if r == constants.DotRune {
					tokType = FLOAT
					break
				}
			}
			tok = Token{Type: tokType, Literal: literal, Line: startLine, Column: numStartCol}
			//##fmt.Printf("DEBUG LEXER: Returning token: %s\n", tok.String())
			return tok // Return early as readNumber advanced
		} else {
			// Genuine illegal character
			//##fmt.Printf("DEBUG LEXER: Creating ILLEGAL token for char %q\n", l.ch)
			tok = makeToken(ILLEGAL, string(l.ch))
			// Let the code below advance past the illegal character
		}
	}

	// Advance past the token character ONLY if we didn't return early
	// (String, Identifier, Number cases return early)
	l.readChar()

	//##fmt.Printf("DEBUG LEXER: Returning token (default path): %s\n", tok.String())
	return tok
}

// handleEOFIndentation ... - No changes needed here
func (l *Lexer) handleEOFIndentation() Token {
	//##fmt.Printf("DEBUG LEXER: handleEOFIndentation called (line %d, col %d, stack %v)\n", l.line, l.column, l.indentStack)
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		//##fmt.Printf("DEBUG LEXER: Returning pending token from EOF handler: %s\n", tok.String())
		return tok
	}

	eofLine := l.line
	eofColumn := l.column // Use current column at EOF

	// Generate pending DEDENTs
	for len(l.indentStack) > 1 {
		//##fmt.Printf("DEBUG LEXER: EOF handler popping indent %d\n", l.indentStack[len(l.indentStack)-1])
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		// DEDENT column is typically 1, associated with the implicit line change/end
		dedentToken := Token{Type: DEDENT, Literal: constants.DedentLiteral, Line: eofLine, Column: 1}
		l.pendingTokens = append(l.pendingTokens, dedentToken)
	}

	// Return first pending DEDENT if any
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		//##fmt.Printf("DEBUG LEXER: Returning generated DEDENT from EOF handler: %s\n", tok.String())
		return tok
	}

	// Return the final EOF token
	eofToken := Token{Type: EOF, Literal: constants.EmptyString, Line: eofLine, Column: eofColumn} // <--- MODIFIED LITERAL
	//##fmt.Printf("DEBUG LEXER: Returning final EOF token: %s\n", eofToken.String())
	return eofToken
}

// // readIdentifier ... - No changes needed here
//
//	func (l *Lexer) readIdentifier() string {
//		startPosition := l.position
//		for isLetter(l.ch) || isDigit(l.ch) {
//			l.readChar()
//		}
//		return l.input[startPosition:l.position]
//	}
//
// --- *** REVISED readIdentifier *** ---
func (l *Lexer) readIdentifier() string {
	startPosition := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	ident := l.input[startPosition:l.position]

	// Check for multi-word operators starting with this identifier
	if ident == constants.NotKeyword && l.ch == constants.SpaceRune && l.peekChar() == constants.CharIRune && l.peekCharN(2) == constants.CharNRune {
		// Possible "not in". We must confirm the next word is "in" and not part of another identifier.
		// A simple way is to peek ahead. A more robust way involves backtracking, but peeking is often sufficient.
		// Let's create a temporary lexer to look ahead.
		tempLexer := *l
		tempLexer.readChar() // Consume the space

		nextIdentStart := tempLexer.position
		for isLetter(tempLexer.ch) {
			tempLexer.readChar()
		}
		nextIdent := tempLexer.input[nextIdentStart:tempLexer.position]

		if nextIdent == constants.InKeyword {
			// It's "not in". Consume the characters in the main lexer.
			l.readChar() // consume space
			l.readChar() // consume 'i'
			l.readChar() // consume 'n'
			return constants.NotInKeyword
		}
	} else if ident == constants.IsKeyword && l.ch == constants.SpaceRune && l.peekChar() == constants.CharNRune && l.peekCharN(2) == constants.CharORune && l.peekCharN(3) == constants.CharTRune {
		// Possible "is not". Similar lookahead logic.
		tempLexer := *l
		tempLexer.readChar() // Consume space
		nextIdentStart := tempLexer.position
		for isLetter(tempLexer.ch) {
			tempLexer.readChar()
		}
		nextIdent := tempLexer.input[nextIdentStart:tempLexer.position]

		if nextIdent == constants.NotKeyword {
			// It's "is not". Consume.
			l.readChar() // space
			l.readChar() // n
			l.readChar() // o
			l.readChar() // t
			return constants.IsNotKeyword
		}
	}

	return ident
}

// This function must now handle the multi-word literals
func LookupIdent(ident string) TokenType {
	if tokType, ok := keywords[ident]; ok {
		return tokType
	}
	// Check for the multi-word tokens we synthesized in readIdentifier
	if ident == constants.NotInKeyword {
		return NOT_IN
	}
	if ident == constants.IsNotKeyword {
		return IS_NOT
	}
	return IDENT
}

// readNumber ... - No changes needed here
func (l *Lexer) readNumber() string {
	startPosition := l.position
	//##fmt.Printf("DEBUG LEXER: readNumber starting at pos %d (char %q)\n", startPosition, l.ch)
	hasDecimal := false
	for isDigit(l.ch) || (l.ch == constants.DotRune && !hasDecimal && isDigit(l.peekChar())) {
		if l.ch == constants.DotRune {
			hasDecimal = true
		}
		l.readChar()
	}
	endPosition := l.position
	literal := l.input[startPosition:endPosition]
	//##fmt.Printf("DEBUG LEXER: readNumber returning literal %q (ends at pos %d, char %q)\n", literal, endPosition, l.ch)
	return literal
}

// --- RENAME and MODIFY readString -> readStringOrBytesContent ---
// readStringOrBytesContent reads the content of a string or bytes literal.
// Returns the UNESCAPED string content and success bool.
// For bytes, validates escapes but returns string content (parser will convert).
func (l *Lexer) readStringOrBytesContent(quote rune, isTriple bool, isBytes bool) (string, bool) {
	var out bytes.Buffer

	for {
		if l.ch == 0 {
			return constants.EmptyString, false
		} // Unterminated

		if l.ch == constants.BackslashRune { // Handle escape sequences
			l.readChar()
			escapeChar := l.ch
			if escapeChar == 0 {
				return constants.EmptyString, false
			} // Unterminated after backslash

			// Validate escapes for bytes literals
			validByteEscape := true
			var escapedRune rune = 0 // Default for non-rune escapes
			handled := true

			switch escapeChar {
			case constants.CharNRune:
				escapedRune = constants.NewlineRune
			case constants.CharTRune:
				escapedRune = constants.TabRune
			case constants.CharRRune:
				escapedRune = constants.CarriageReturnRune
			case constants.CharBRune:
				escapedRune = constants.BackspaceRune // Backspace (less common)
			case constants.CharFRune:
				escapedRune = constants.FormFeedRune // Form feed (less common)
			case constants.BackslashRune:
				escapedRune = constants.BackslashRune
			case constants.SingleQuoteRune:
				escapedRune = constants.SingleQuoteRune
			case constants.DoubleQuoteRune:
				escapedRune = constants.DoubleQuoteRune
			case constants.CharZero, constants.CharOne, constants.CharTwo, constants.CharThree, constants.CharFour, constants.CharFive, constants.CharSix, constants.CharSeven: // Octal escapes (e.g., \0, \12, \377)
				// Python allows \0 -> NUL, but other octal escapes aren't typical in modern code
				// For simplicity, let's only handle \0 maybe? Or disallow?
				// Let's treat as literal backslash + digit for now if isBytes is false
				if !isBytes {
					validByteEscape = false
					handled = false
				} else {
					// TODO: Implement octal escape parsing if needed
					validByteEscape = false
					handled = false // Disallow for now
				}

			case constants.CharXRune: // Hex escapes \xHH
				l.readChar() // Read first hex digit
				h1 := l.ch
				l.readChar() // Read second hex digit
				h2 := l.ch
				hexStr := string([]rune{h1, h2})
				byteVal, err := strconv.ParseUint(hexStr, 16, 8)
				if err != nil {
					// Invalid hex escape sequence
					// Append backslash, x, and the two chars read
					out.WriteRune(constants.BackslashRune)
					out.WriteRune(constants.CharXRune)
					out.WriteRune(h1)
					out.WriteRune(h2)
					validByteEscape = false
					handled = false // Mark as unhandled for bytes check
				} else {
					// Write the actual byte value
					out.WriteByte(byte(byteVal))
					// No rune needed for byte escapes
				}

			case constants.NewlineRune: // Escaped newline (ignore backslash and newline)
				l.line++
				l.column = 0
				// Do not write anything to buffer

			// Unicode escapes are ONLY valid for strings, not bytes
			case constants.CharURune, constants.CharUUpperRune, constants.CharNUpperRune:
				if isBytes {
					validByteEscape = false
				}
				handled = false // Treat as literal backslash + char

			default: // Unknown escape sequence
				if isBytes {
					validByteEscape = false
				} // Invalid in bytes
				handled = false // Treat as literal backslash + char
			} // end switch escapeChar

			if isBytes && !validByteEscape {
				// Return error or handle? Let's return error via ok=false maybe?
				// For now, let lexer produce ILLEGAL token later by writing literal backslash+char
				// (This requires parser to maybe re-validate bytes content)
				handled = false
			}

			if handled && escapedRune != 0 { // If it was a standard char escape like \n, \t, \\
				out.WriteRune(escapedRune)
			} else if !handled { // If not handled above (unknown, unicode in bytes, octal)
				out.WriteRune(constants.BackslashRune)
				out.WriteRune(escapeChar)
			}
			l.readChar() // Consume the char after escape sequence (or last hex digit)

		} else if l.ch == quote { // Potential closing quote
			if isTriple {
				if l.peekChar() == quote && l.peekCharN(2) == quote {
					l.readChar()
					l.readChar()
					l.readChar()              // Consume closing quotes
					return out.String(), true // Terminated
				} else {
					out.WriteRune(l.ch)
					l.readChar() // Just a regular quote
				}
			} else { // Single quoted
				l.readChar()              // Consume closing quote
				return out.String(), true // Terminated
			}
		} else if l.ch == constants.NewlineRune { // Newline within literal
			if !isTriple && !isBytes { // Single-quoted strings cannot contain unescaped newlines
				// This should ideally be caught earlier or return error
				return constants.EmptyString, false // Treat as unterminated/error
			}
			out.WriteRune(l.ch)
			l.readChar()
			l.line++
			l.column = 0
		} else { // Regular character
			// Check validity for bytes literals
			if isBytes && l.ch >= utf8.RuneSelf {
				// Non-ASCII character in bytes literal (error)
				// How to signal error? Let parser handle? Or make lexer return ILLEGAL?
				// Let's write the rune for now, parser can validate later.
			}
			out.WriteRune(l.ch)
			l.readChar()
		}
	}
}

// --- Updated readString ---
// readString reads the content of a string literal (single or triple quoted).
// It handles basic escape sequences and newlines within multiline strings.
// It consumes the closing quote(s).
// Returns the string content and a boolean indicating success (true) or unterminated (false).
func (l *Lexer) readString(quote rune, isTriple bool) (string, bool) {
	var out bytes.Buffer // Use bytes.Buffer for efficient string building

	for {
		if l.ch == 0 { // EOF before closing quote
			return constants.EmptyString, false // Unterminated
		}

		if l.ch == constants.BackslashRune { // Handle escape sequences
			l.readChar() // Consume '\'
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
			// Add more escapes as needed (e.g., \r, \b, \f, \uXXXX, etc.)
			case 0: // EOF after backslash
				return constants.EmptyString, false // Unterminated
			case constants.NewlineRune: // Escaped newline (ignore backslash and newline)
				l.line++
				l.column = 0
				// Consume the newline character without adding anything to buffer
			default:
				// Unknown escape sequence, treat as literal backslash + char
				out.WriteRune(constants.BackslashRune)
				out.WriteRune(l.ch)
			}
			l.readChar() // Consume the character after '\' or the escaped char
		} else if l.ch == quote { // Potential closing quote
			if isTriple {
				// Need two more quotes
				if l.peekChar() == quote && l.peekCharN(2) == quote {
					l.readChar()              // Consume first closing quote
					l.readChar()              // Consume second closing quote
					l.readChar()              // Consume third closing quote
					return out.String(), true // Properly terminated triple quote
				} else {
					// Just a regular quote char inside the triple-quoted string
					out.WriteRune(l.ch)
					l.readChar() // Consume the quote
				}
			} else { // Single quoted string
				l.readChar()              // Consume the closing quote
				return out.String(), true // Properly terminated single quote
			}
		} else if l.ch == constants.NewlineRune { // Handle newline within the string
			out.WriteRune(l.ch) // Keep the newline in the string literal
			l.readChar()        // Consume newline
			l.line++
			l.column = 0 // Reset column
		} else { // Regular character
			out.WriteRune(l.ch)
			l.readChar() // Consume the character
		}
	}
	// Loop only exits via return statements (terminated or unterminated)
}

// isLetter, isDigit remain the same...
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == constants.UnderscoreRune
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
