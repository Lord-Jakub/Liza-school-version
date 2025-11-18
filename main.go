package main

import (
	"lizalang/lexer"
	"lizalang/parser"
	"lizalang/tests"
	//"lizalang/token"
)

func main() {
	code := `namespace test
	func foo(int x, int y) int{
	x = y
	}
	func main(){
	int i
	i = 1*5+(3-2)/5^2^2
	foo(i, 5*6)
	}
	`
	code = string(append([]byte(code), 0))
	lex := lexer.New(code, "nil")
	lex.Lex()
	par := parser.New(lex.Tokens)
	// tests.TestParseExpression(par)
	tests.TestParser(par)
}
