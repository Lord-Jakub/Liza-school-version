package main

import (
	"lizalang/lexer"
	"lizalang/parser"
	"lizalang/tests"
	//"lizalang/token"
)

func main() {
	code := `namespace test
	for int i; i<10; i = i+1{
	return array[i]
	}
	`
	// code := "!(5*3==15 && 5!=3)"
	code = string(append([]byte(code), 0))
	lex := lexer.New(code, "nil")
	lex.Lex()
	// tests.TestLexer(lex)
	par := parser.New(lex.Tokens)
	// tests.TestParseExpression(par)
	tests.TestParser(par)
}
