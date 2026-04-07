//go:build en

package constants

// pylearn/internal/lexer/token.go
const (
	LexerTokenFmtString   = "Token{Type: %s, Literal: %s, Line: %d, Column: %d}"
	LexerTokenQuoteFormat = "%q"
)

// pylearn/internal/lexer/lexer.go
const (
	LexerIndentationTabError          = "indentation tab error"
	LexerUnterminatedMultilineComment = "Unterminated multiline comment"
	LexerIndentationDedentError       = "indentation dedent error"
	LexerUnterminatedStringOrBytes    = "Unterminated string/bytes"
	LexerIllegalCharacter             = "illegal character" // Generic message for unrecognized char
)

// pylearn/internal/lexer/token.go
const (
	LexerKeywordDef      = "def"
	LexerKeywordTrue     = "True"
	LexerKeywordFalse    = "False"
	LexerKeywordIf       = "if"
	LexerKeywordElif     = "elif"
	LexerKeywordElse     = "else"
	LexerKeywordReturn   = "return"
	LexerKeywordFor      = "for"
	LexerKeywordWhile    = "while"
	LexerKeywordIn       = "in"
	LexerKeywordAnd      = "and"
	LexerKeywordOr       = "or"
	LexerKeywordNot      = "not"
	LexerKeywordNone     = "None"
	LexerKeywordBreak    = "break"
	LexerKeywordContinue = "continue"
	LexerKeywordClass    = "class"
	LexerKeywordPass     = "pass"
	LexerKeywordImport   = "import"
	LexerKeywordFrom     = "from"
	LexerKeywordAs       = "as"
	LexerKeywordTry      = "try"
	LexerKeywordExcept   = "except"
	LexerKeywordRaise    = "raise"
	LexerKeywordFinally  = "finally"
	LexerKeywordDel      = "del"
	LexerKeywordGlobal   = "global"
	LexerKeywordNonlocal = "nonlocal"
	LexerKeywordWith     = "with"
	LexerKeywordAsync    = "async"
	LexerKeywordAwait    = "await"
	LexerKeywordIs       = "is"
	LexerKeywordYield    = "yield"
	LexerKeywordAssert   = "assert"
	LexerKeywordLambda   = "lambda"
)

const (
	LexerTokenTypeIllegal = "ILLEGAL"
	LexerTokenTypeEOF     = "EOF"
	LexerTokenTypeIndent  = "INDENT"
	LexerTokenTypeDedent  = "DEDENT"
	LexerTokenTypeNewline = "NEWLINE"
	// Identifiers + Literals
	LexerTokenTypeIdent   = "IDENT"
	LexerTokenTypeInt     = "INT"
	LexerTokenTypeFloat   = "FLOAT"
	LexerTokenTypeString  = "STRING"
	LexerTokenTypeFString = "FSTRING"
	LexerTokenTypeBytes   = "BYTES"
	LexerTokenTypeBool    = "BOOL"

	LexerTokenTypeIs    = "is"
	LexerTokenTypeIsNot = "is not"
	LexerTokenTypeNotIn = "not in"

	LexerTokenTypeStar = "STAR" // <<< ADD THIS for 'import *' symbol
	// Keywords
	LexerTokenTypeFunction = "FUNCTION" // def
	LexerTokenTypeTrue     = "TRUE"     // True
	LexerTokenTypeFalse    = "FALSE"    // False
	LexerTokenTypeIf       = "IF"       // if
	LexerTokenTypeElif     = "ELIF"     // elif
	LexerTokenTypeElse     = "ELSE"     // else
	LexerTokenTypeReturn   = "RETURN"   // return
	LexerTokenTypeFor      = "FOR"      // for
	LexerTokenTypeWhile    = "WHILE"    // while
	LexerTokenTypeIn       = "IN"       // in
	LexerTokenTypeAnd      = "AND"      // and
	LexerTokenTypeOr       = "OR"       // or
	LexerTokenTypeNot      = "NOT"      // not
	LexerTokenTypeNil      = "NIL"      // None
	LexerTokenTypeBreak    = "BREAK"    // break
	LexerTokenTypeContinue = "CONTINUE" // continue
	LexerTokenTypeClass    = "CLASS"
	LexerTokenTypePass     = "PASS"
	LexerTokenTypeImport   = "IMPORT" // import
	LexerTokenTypeFrom     = "FROM"
	LexerTokenTypeAs       = "AS"
	LexerTokenTypeTry      = "TRY"
	LexerTokenTypeExcept   = "EXCEPT"
	LexerTokenTypeRaise    = "RAISE"
	LexerTokenTypeFinally  = "FINALLY"
	LexerTokenTypeDel      = "DEL"
	LexerTokenTypeGlobal   = "GLOBAL"
	LexerTokenTypeNonLocal = "NONLOCAL"
	LexerTokenTypeComment  = "COMMENT"
	LexerTokenTypeAsync    = "ASYNC"
	LexerTokenTypeAwait    = "AWAIT"
	LexerTokenTypeYield    = "YIELD"
	LexerTokenTypeWith     = "WITH"
	LexerTokenTypeAssert   = "ASSERT"
	LexerTokenTypeLambda   = "LAMBDA"
)
