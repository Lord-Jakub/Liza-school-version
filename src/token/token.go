package token

// Token is a structure representing lexical tokens of this language
type Token struct {
	Type  TokenType
	Value any
	Line  int
	File  string
}

// TokenType is alias for string used to represent type of token
type TokenType string

// Types of tokens are represented as constants
const (
	Invalid        = "INVALID"
	EOF            = "EOF"
	OpenParen      = "OPENPAREN"
	CloseParen     = "CLOSEPAREN"
	OpenBrace      = "OPENBRACE"
	CloseBrace     = "CLOSEBRACE"
	Operator       = "OPERATOR"
	Backslash      = "BACKSLASH"
	NewInstruction = "NEWINSTRUCTION"
	SingleQuote    = "SINGLEQUOTE"
	DoubleQuote    = "DOUBLEQUOTE"
	Equal          = "EQUAL"
	Comma          = "COMMA"
	Int            = "INT"
	String         = "STRING"
	Identifier     = "IDENTIFIER"
	Keyword        = "KEYWORD"
	Float          = "FLOAT"
	OpenBracket    = "OPENBRACKET"
	CloseBracket   = "CLOSEBRACKET"
	Dot            = "DOT"
)

// Tokens that can be represented as one symbol are defined in lookup table
var SymbolMap = map[rune]TokenType{
	'(':  OpenParen,
	')':  CloseParen,
	'{':  OpenBrace,
	'}':  CloseBrace,
	'+':  Operator,
	'-':  Operator,
	'*':  Operator,
	'/':  Operator,
	'\\': Backslash,
	'=':  Equal,
	'!':  Operator,
	',':  Comma,
	'<':  Operator,
	'>':  Operator,
	';':  NewInstruction,
	'[':  OpenBracket,
	']':  CloseBracket,
	'^':  Operator,
	'.':  Dot,
	'%':  Operator,
}

var (
	// KeyWords is a slice of all reserved words in the language.
	KeyWords = []string{"if", "else", "for", "func", "return", "string", "int", "float", "bool", "void", "namespace", "constant", "import"}
	// Types defines the built-in primitive type identifiers.
	Types = []string{"string", "int", "float", "bool", "void"}
)
