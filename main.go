package main

import (
	"lizalang/lexer"
	"lizalang/parser"
	"lizalang/tests"
	//"lizalang/token"
)

func main() {
	code := `[[1,2],[3,4],[5,arr[6]]]`
	// code := "!(5*3==15 && 5!=3)"
	code = string(append([]byte(code), 0))
	lex := lexer.New(code, "nil")
	lex.Lex()
	// tests.TestLexer(lex)
	par := parser.New(lex.Tokens)
	tests.TestParseExpression(par)
	// tests.TestParser(par)
}
