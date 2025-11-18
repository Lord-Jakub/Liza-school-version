package tests

import (
	"encoding/json"
	"fmt"
	"strconv"

	"lizalang/ast"
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
	fmt.Println(PrintExpression(data))
}

func PrintExpression(e ast.Expression) string {
	switch v := e.(type) {

	case *ast.BinaryExpression:
		return "(" + PrintExpression(v.Left) + v.Op.Value.(string) + PrintExpression(v.Right) + ")"

	case *ast.UnaryExpression:
		return "(" + strconv.Itoa(int(v.Prefix.Value.(int64))) + PrintExpression(v.Value) + ")"

	/*case *ast.FunctionCall:
		args := make([]string, len(v.Args))
		for i, a := range v.Args {
			args[i] = PrintExpression(a)
		}
		return v.Identifier.Literal + "(" + strings.Join(args, ", ") + ")"

	case *ast.IdentifierExpression:
		return v.Value.Literal
	*/
	case *ast.LiteralExpression:
		return strconv.Itoa(int(v.Value.Value.(int64)))
	}

	return ""
}

func TestParser(parser *parser.Parser) {
	data := parser.ParseNamespace()
	jsonExpression, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println(string(jsonExpression))
	for _, err := range parser.Errors {
		fmt.Println(err)
	}
}
