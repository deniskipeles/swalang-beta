//go:build en
// pylearn/internal/constants/all_en.go
package constants

// pylearn/internal/object/exceptions.go
// const EXCEPTION_MRO_COMPUTATION_FAILED = "FATAL: Could not compute MRO for built-in exception '%s': %v\n"

const (
	SingleQuoteRune    = '\''
	DoubleQuoteRune    = '"'
	OpenParenRune      = '('
	CloseParenRune     = ')'
	SemicolonRune      = ';'
	ColonRune          = ':'
	OpenBraceRune      = '{'
	CloseBraceRune     = '}'
	OpenBracketRune    = '['
	CloseBracketRune   = ']'
	DotRune            = '.'
	AtRune             = '@'
	HashRune           = '#'
	AsteriskRune       = '*'
	SlashRune          = '/'
	PercentRune        = '%'
	PlusOperatorRune   = '+'
	MinusSignRune      = '-'
	BangRune           = '!'
	AssignOperatorRune = '='
	LessThanOpRune     = '<'
	GreaterThanOpRune  = '>'
	CommaRune          = ','
	SpaceRune          = ' '
	TabRune            = '\t'
	CarriageReturnRune = '\r'
	NewlineRune        = '\n'
	BackspaceRune      = '\b'
	FormFeedRune       = '\f'
	VerticalTabRune    = '\v'
	BackslashRune      = '\\'
	UnderscoreRune     = '_'
)
const (
	SingleQuoteByte    = '\''
	DoubleQuoteByte    = '"'
	OpenParenByte      = '('
	CloseParenByte     = ')'
	SemicolonByte      = ';'
	ColonByte          = ':'
	OpenBraceByte      = '{'
	CloseBraceByte     = '}'
	OpenBracketByte    = '['
	CloseBracketByte   = ']'
	DotByte            = '.'
	AtByte             = '@'
	HashByte           = '#'
	AsteriskByte       = '*'
	SlashByte          = '/'
	PercentByte        = '%'
	PlusOperatorByte   = '+'
	MinusSignByte      = '-'
	BangByte           = '!'
	AssignOperatorByte = '='
	LessThanOpByte     = '<'
	GreaterThanOpByte  = '>'
	CommaByte          = ','
	SpaceByte          = ' '
	TabByte            = '\t'
	CarriageReturnByte = '\r'
	NewlineByte        = '\n'
	BackspaceByte      = '\b'
	FormFeedByte       = '\f'
	VerticalTabByte    = '\v'
	BackslashByte      = '\\'
	UnderscoreByte     = '_'
)

// Relational Operators
const (
	EqualsOperator         = "=="
	NotEqualsOperator      = "!="
	LessThanOp             = "<"
	LessThanEqualsOp       = "<="
	GreaterThanOp          = ">"
	GreaterThanEqualsOp    = ">="
	AssignOperator         = "="
	EqOperator             = "=="
	NotEqOperator          = "!="
	LessThanOrEqualToOp    = "<="
	GreaterThanOrEqualToOp = ">="
)
const (
	Space                  = " "
	EmptyString            = ""
	SingleQuote            = "'"
	DoubleQuote            = "\""
	Backslash              = "\\"
	Slash                  = "/"
	BangOperator           = "!"
	QuestionMark           = "?"
	AsteriskOperator       = "*"
	PowOperator            = "**"
	DoubleAsteriskOperator = "**"
	CommaOperator          = ","
	CommaWithSpace         = ", "
	SemicolonOperator      = ";"
	Semicolon              = ";"
	SemicolonWithSpace     = "; "
	ColonOperator          = ":"
	OpenParenOperator      = "("
	CloseParenOperator     = ")"
	OpenBraceOperator      = "{"
	CloseBraceOperator     = "}"
	OpenBracketOperator    = "["
	CloseBracketOperator   = "]"
	DotOperator            = "."
	AtOperator             = "@"
	HashOperator           = "#"

	SlashOperator   = "/"
	PercentOperator = "%"
	PlusOperator    = "+"
	MinusOperator   = "-"
	Underscore      = "_"

	PyExtension = ".py"

	SlashAsteriskComment = "/*...*/"
	IndentLiteral        = "INDENT"
	DedentLiteral        = "DEDENT"

	SemiColon          = ";"
	SemiColonWithSpace = "; "
	OpenParen          = "("
	Newline        = "\n"
	WindowsNewline = "\r\n"
	PlusChar       = "+"
	CloseParen     = ")"
	MinusSign      = "-"
	PlusSign       = "+"
	Dot            = "."
)

// String numbers
const (
	ZeroString = "0"
	Zero       = "0"
	One        = "1"
	Two        = "2"
	Three      = "3"
	Four       = "4"
	Five       = "5"
	Six        = "6"
	Seven      = "7"
	Eight      = "8"
	Nine       = "9"
)

// format specifiers
const (
	FormatInt                  = "%d"
	FormatHex                  = "%x"
	FormatOct                  = "%o"
	FormatBin                  = "%b"
	FormatFloat                = "%f"
	FormatScientific           = "%e"
	DoubleQuoteFormat          = "%q"
	ErrorFormat                = "%s"
	StringFormat               = "%s"
	FormatFloatNoTrailingZeros = "%g"

	HexEscapeFormat                    = "\\x%02x"
	OctEscapeFormat                    = "\\%03o"
	BinEscapeFormat                    = "\\%08b"
	UnicodeEscapeFormat                = "\\u%04x"
	UniversalCharacterNameEscapeFormat = "\\U%08x"

	FormatSpecifierD = "d"
	FormatSpecifierX = "x"
	FormatSpecifierO = "o"
	FormatSpecifierB = "b"
	FormatSpecifierF = "f"
	FormatSpecifierE = "e"
	FormatSpecifierS = "s"
)

// error strings
const (
	AttributeErrorColon = "AttributeErrorColon:"
	ErrorKeyword        = "Error"
	ExceptionKeyword    = "Exception"
)

// Rune Numbers
const (
	CharZero  = '0'
	CharOne   = '1'
	CharTwo   = '2'
	CharThree = '3'
	CharFour  = '4'
	CharFive  = '5'
	CharSix   = '6'
	CharSeven = '7'
	CharEight = '8'
	CharNine  = '9'
)

// CharRunes
const (
	CharBRune      = 'b'
	CharBUpperRune = 'B'
	CharNRune      = 'n'
	CharNUpperRune = 'N'
	CharFRune      = 'f'
	CharFFRune     = 'F'
	CharIRune      = 'i'
	CharIUpperRune = 'I'
	CharTRune      = 't'
	CharTUpperRune = 'T'
	CharRRune      = 'r'
	CharRUpperRune = 'R'
	CharORune      = 'o'
	CharOUpperRune = 'O'
	CharURune      = 'u'
	CharUUpperRune = 'U'
	CharVRune      = 'v'
	CharVUpperRune = 'V'
	CharXRune      = 'x'
	CharXUpperRune = 'X'
)

// Char
const (
	CharB = "b"
	CharN = "n"
	CharF = "f"
	CharT = "t"
	CharR = "r"
	CharO = "o"
	CharU = "u"
	CharX = "x"
)
