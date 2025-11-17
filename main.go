package main

import (
	"lizalang/lexer"
	"lizalang/parser"
	"lizalang/tests"
	//"lizalang/token"
)

func main() {
	code := `-(1+2)*3`
	code = string(append([]byte(code), 0))
	lex := lexer.New(code, "nil")
	lex.Lex()
	par := parser.New(lex.Tokens)
	tests.TestParseExpression(par)
}
