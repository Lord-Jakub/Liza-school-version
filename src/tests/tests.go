package tests

import (
	"encoding/json"
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/eval"
	"lizalang/interpreter/object"
	"lizalang/lexer"
	"lizalang/parser"
)

func TestLexer(lex *lexer.Lexer) {
	for _, tok := range lex.Tokens {
		if _, ok := tok.Value.(rune); ok {
			fmt.Printf("%s:%s\n", tok.Type, string(tok.Value.(rune)))
		} else {
			fmt.Printf("%s:%s\n", tok.Type, tok.Value)
		}
	}
	for _, err := range lex.Errors {
		fmt.Println(err)
	}
}

func TestParseExpression(parser *parser.Parser) {
	data := parser.ParseExpression(0)
	// jsonExpression, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println(data.String())
}

func TestParser(parser *parser.Parser) {
	data := parser.Program
	jsonExpression, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println(string(jsonExpression))
	for _, err := range parser.Errors {
		fmt.Println(err)
	}
}

func TestEval(expr ast.Expression) {
	val, err := eval.Eval(expr)
	if err != nil {
		fmt.Println(err)
		return
	}
	PrintVal(val)
	fmt.Println()
	fmt.Println()
}

func PrintVal(val object.Object) {
	v := val.GetValue()
	switch v := v.(type) {
	case (int64):
		fmt.Printf("%d", v)
	case (float64):
		fmt.Printf("%f", v)
	case (bool):
		fmt.Printf("%t", v)
	case (string):
		fmt.Printf("%s", v)
	case ([]object.Object):
		fmt.Print("[")
		for i, el := range v {
			if i != 0 {
				fmt.Print(",")
			}
			PrintVal(el)
		}
		fmt.Print("]")

	}
}

/*
code := `namespace Algorithms

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
	}
	code := "(20/(1+1)*5-2*4+(66*5+6))*3-2^([1, 2, 3, 4, 5, 6][5])"
	code = string(append([]byte(code), 0))
	code2 := "1+1 == 2 && 1==1"
	code2 = string(append([]byte(code2), 0))
	code3 := "5.5*1.1/3.0"
	code3 = string(append([]byte(code3), 0))
	code4 := "\"Hello \" + \"World\""
	code4 = string(append([]byte(code4), 0))
	code5 := "[[[1, 2], [3,4],[5,6]], [[5, 6], [7,8],[9,10]]][0][2]"
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
*/
