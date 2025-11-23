package main

import (
	"fmt"
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
	code2 := "1+1 == 2 && 1==1"
	code2 = string(append([]byte(code2), 0))
	code3 := "5.5*1.1/3.0"
	code3 = string(append([]byte(code3), 0))
	code4 := "\"Hello \" + \"World\""
	code4 = string(append([]byte(code4), 0))
	code5 := "[[1, 2], [3,4],[5,6]]"
	code5 = string(append([]byte(code5), 0))

	lex := lexer.New(code, "nil")
	lex.Lex()
	// tests.TestLexer(lex)
	// par := parser.New(lex.Tokens)
	// tests.TestParseExpression(par)
	// par.Parse()
	// tests.TestParser(par)
	fmt.Printf("%s = ", code)
	par2 := parser.New(lex.Tokens)
	tests.TestEval(par2.ParseExpression(0))

	lex2 := lexer.New(code2, "")
	lex2.Lex()
	fmt.Printf("%s = ", code2)
	tests.TestEval(parser.New(lex2.Tokens).ParseExpression(0))

	lex3 := lexer.New(code3, "")
	lex3.Lex()
	fmt.Printf("%s = ", code3)
	tests.TestEval(parser.New(lex3.Tokens).ParseExpression(0))

	lex4 := lexer.New(code4, "")
	lex4.Lex()
	fmt.Printf("%s = ", code4)
	tests.TestEval(parser.New(lex4.Tokens).ParseExpression(0))

	lex5 := lexer.New(code5, "")
	lex5.Lex()
	fmt.Printf("%s = ", code5)
	tests.TestEval(parser.New(lex5.Tokens).ParseExpression(0))
}
