package tests

import (
	"encoding/json"
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter"
	"lizalang/lexer"
	"lizalang/object"
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
	val, err := interpreter.Eval(expr)
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
