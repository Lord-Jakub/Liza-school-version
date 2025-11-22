package main

import (
	"lizalang/lexer"
	"lizalang/parser"
	"lizalang/tests"
	//"lizalang/token"
)

func main() {
	/*code := `namespace Algorithms

	// Global variable test
	float PI = 3.14159

	// Function declaration test
	func fibonacci(int n) int {
	    // If statement test
	    if (n <= 1) {
	        return n
	    }
	    // Recursion and binary operators
	    return fibonacci(n - 1) + fibonacci(n - 2)
	}

	func main() void {
	    // Array declaration test
	    int[] numbers = [10, 20, 30, 40, 50]

	    // Variable declaration
	    int sum = 0

	    // For loop test
	    for int i = 0; i < 5; i = i + 1 {
	        // Array Indexing Test (uses your infix logic)
	        int val = numbers[i]

	        // Assignment and math
	        sum = sum + val
	    }

	    // Complex expression test
	    int result = (sum * 2) / 5
	}`*/
	code := "(20/(1+1)*5-2*4+(66*5+6))*3-2^6"
	code = string(append([]byte(code), 0))
	lex := lexer.New(code, "nil")
	lex.Lex()
	// tests.TestLexer(lex)
	par := parser.New(lex.Tokens)
	tests.TestParseExpression(par)
	// par.Parse()
	// tests.TestParser(par)
	par2 := parser.New(lex.Tokens)
	tests.TestEval(par2.ParseExpression(0))
}
