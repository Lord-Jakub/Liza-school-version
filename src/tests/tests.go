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
	println(val.(*object.IntObject).Value)
}
