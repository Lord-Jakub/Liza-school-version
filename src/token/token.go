package token

type Token struct {
	Type  TokenType
	Value any
	Line  int
	File  string
}
type TokenType string

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
)

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
}

var (
	KeyWords = []string{"if", "else", "for", "func", "return", "string", "int", "float", "bool", "void", "namespace", "constant", "import"}
	Types    = []string{"string", "int", "float", "bool", "void"}
)
