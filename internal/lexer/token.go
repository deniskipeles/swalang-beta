package lexer

import (
	"fmt"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// TokenType remains the same...
type TokenType string

// Token struct remains the same...
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// String() method remains the same...
func (t Token) String() string {
	lit := t.Literal
	if t.Type == INDENT || t.Type == DEDENT || t.Type == EOF || t.Type == ILLEGAL {
		lit = string(t.Type)
	} else {
		lit = fmt.Sprintf(constants.LexerTokenQuoteFormat, t.Literal)
	}
	return fmt.Sprintf(constants.LexerTokenFmtString,
		t.Type, lit, t.Line, t.Column)
}

const (
	// Special tokens
	ILLEGAL TokenType = constants.LexerTokenTypeIllegal // Illegal token
	EOF     TokenType = constants.LexerTokenTypeEOF     // End of file

	// Indentation
	INDENT  TokenType = constants.LexerTokenTypeIndent  // Indentation
	DEDENT  TokenType = constants.LexerTokenTypeDedent  // Dedentation
	NEWLINE TokenType = constants.LexerTokenTypeNewline // Newline

	// Identifiers + Literals
	IDENT   TokenType = constants.LexerTokenTypeIdent   // Identifier
	INT     TokenType = constants.LexerTokenTypeInt     // Integer literal
	FLOAT   TokenType = constants.LexerTokenTypeFloat   // Floating-point literal
	STRING  TokenType = constants.LexerTokenTypeString  // String literal
	FSTRING TokenType = constants.LexerTokenTypeFString // Formatted string literal
	BYTES   TokenType = constants.LexerTokenTypeBytes   // Bytes literal
	BOOL    TokenType = constants.LexerTokenTypeBool    // Boolean literal
	// CHAR    TokenType = constants.LexerTokenTypeChar    // Character literal

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	PLUS_EQ  TokenType = "+="
	MINUS    TokenType = "-"
	MINUS_EQ TokenType = "-="
	// MUL_EQ      TokenType = "*="
	// DIV_EQ      TokenType = "/="
	// FLOOR_DIV_EQ TokenType = "//="
	// POW_EQ      TokenType = "**="
	// MOD_EQ      TokenType = "%="
	// PLUS_PLUS   TokenType = "++"
	// MINUS_MINUS TokenType = "--"
	BANG        TokenType = "!"
	ASTERISK    TokenType = "*"
	POW         TokenType = "**"
	SLASH       TokenType = "/"
	FLOOR_DIV   TokenType = "//"
	PERCENT     TokenType = "%"
	LT          TokenType = "<"
	GT          TokenType = ">"
	EQ          TokenType = "=="
	NOT_EQ      TokenType = "!="
	LT_EQ       TokenType = "<="
	GT_EQ       TokenType = ">="
	IS          TokenType = constants.LexerTokenTypeIs // is operator 
	IS_NOT      TokenType = constants.LexerTokenTypeIsNot // is not operator
	NOT_IN      TokenType = constants.LexerTokenTypeNotIn // not in operator
	BITWISE_AND TokenType = "&"
	LSHIFT      TokenType = "<<"
	RSHIFT      TokenType = ">>"
	// BITWISE_OR  TokenType = "|"
	// BITWISE_XOR TokenType = "^"
	// BITWISE_NOT TokenType = "~"

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
	DOT       TokenType = "."
	AT        TokenType = "@"
	STAR      TokenType = constants.LexerTokenTypeStar // <<< ADD THIS for 'import *' symbol

	// Keywords
	FUNCTION TokenType = constants.LexerTokenTypeFunction // def
	TRUE     TokenType = constants.LexerTokenTypeTrue     // True
	FALSE    TokenType = constants.LexerTokenTypeFalse    // False
	IF       TokenType = constants.LexerTokenTypeIf       // if
	ELIF     TokenType = constants.LexerTokenTypeElif     // elif
	ELSE     TokenType = constants.LexerTokenTypeElse     // else
	RETURN   TokenType = constants.LexerTokenTypeReturn   // return
	FOR      TokenType = constants.LexerTokenTypeFor      // for
	WHILE    TokenType = constants.LexerTokenTypeWhile    // while
	IN       TokenType = constants.LexerTokenTypeIn       // in
	AND      TokenType = constants.LexerTokenTypeAnd      // and
	OR       TokenType = constants.LexerTokenTypeOr       // or
	NOT      TokenType = constants.LexerTokenTypeNot      // not
	NIL      TokenType = constants.LexerTokenTypeNil      // None
	BREAK    TokenType = constants.LexerTokenTypeBreak    // break
	CONTINUE TokenType = constants.LexerTokenTypeContinue // continue
	CLASS    TokenType = constants.LexerTokenTypeClass // class
	PASS     TokenType = constants.LexerTokenTypePass // pass
	IMPORT   TokenType = constants.LexerTokenTypeImport // import
	FROM     TokenType = constants.LexerTokenTypeFrom
	AS       TokenType = constants.LexerTokenTypeAs
	TRY      TokenType = constants.LexerTokenTypeTry
	EXCEPT   TokenType = constants.LexerTokenTypeExcept
	RAISE    TokenType = constants.LexerTokenTypeRaise
	FINALLY  TokenType = constants.LexerTokenTypeFinally
	DEL      TokenType = constants.LexerTokenTypeDel
	GLOBAL   TokenType = constants.LexerTokenTypeGlobal
	// NONLOCAL TokenType = "NONLOCAL"

	COMMENT TokenType =  constants.LexerTokenTypeComment
	ASYNC   TokenType =  constants.LexerTokenTypeAsync
	AWAIT   TokenType =  constants.LexerTokenTypeAwait
	YIELD   TokenType = constants.LexerTokenTypeYield
	WITH    TokenType = constants.LexerTokenTypeWith
	ASSERT  TokenType = constants.LexerTokenTypeAssert
	LAMBDA  TokenType = constants.LexerTokenTypeLambda
)

// keywords map - REMOVE the "print" entry
var keywords = map[string]TokenType{
	constants.LexerKeywordDef:      FUNCTION,
	constants.LexerKeywordTrue:     TRUE,
	constants.LexerKeywordFalse:    FALSE,
	constants.LexerKeywordIf:       IF,
	constants.LexerKeywordElif:     ELIF,
	constants.LexerKeywordElse:     ELSE,
	constants.LexerKeywordReturn:   RETURN,
	constants.LexerKeywordFor:      FOR,
	constants.LexerKeywordWhile:    WHILE,
	constants.LexerKeywordIn:       IN,
	constants.LexerKeywordAnd:      AND,
	constants.LexerKeywordOr:       OR,
	constants.LexerKeywordNot:      NOT,
	constants.LexerKeywordNone:     NIL,
	constants.LexerKeywordBreak:    BREAK,
	constants.LexerKeywordContinue: CONTINUE,
	constants.LexerKeywordClass:    CLASS,
	constants.LexerKeywordPass:     PASS,
	constants.LexerKeywordImport:   IMPORT,
	constants.LexerKeywordFrom:     FROM,
	constants.LexerKeywordAs:       AS,
	constants.LexerKeywordWith:     WITH,
	constants.LexerKeywordTry:      TRY,
	constants.LexerKeywordExcept:   EXCEPT,
	constants.LexerKeywordRaise:    RAISE,
	constants.LexerKeywordFinally:  FINALLY,
	constants.LexerKeywordDel:      DEL,
	constants.LexerKeywordGlobal:   GLOBAL,
	constants.LexerKeywordAsync:    ASYNC,
	constants.LexerKeywordAwait:    AWAIT,
	constants.LexerKeywordIs:       IS,
	constants.LexerKeywordYield:    YIELD,
	constants.LexerKeywordAssert:   ASSERT,
	constants.LexerKeywordLambda:   LAMBDA,
}
