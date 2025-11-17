package parser

import (
	"lizalang/ast"
	"lizalang/token"
)

type Parser struct {
	Tokens   []token.Token
	Pos      int
	CurTok   token.Token
	TokAfter token.Token
	Program  ast.Program
	Errors   []error
}

func New(tokens []token.Token) *Parser {
	parser := &Parser{
		Tokens:  tokens,
		Pos:     0,
		Program: ast.Program{},
		Errors:  make([]error, 0),
	}
	if len(tokens) > 2 {
		parser.CurTok = tokens[0]
		parser.TokAfter = tokens[1]
	}

	return parser
}

func (parser *Parser) NextToken() {
	parser.Pos++
	parser.CurTok = parser.TokAfter
	if parser.TokAfter.Type != token.EOF {
		parser.TokAfter = parser.Tokens[parser.Pos+1]
	}
}

var (
	precedence         = map[string]int{"+": 1, "-": 1, "*": 2, "/": 2, "^": 3}
	isRightAssociative = map[string]bool{"+": false, "-": false, "*": false, "/": false, "^": true}
)

func (parser *Parser) ParseExpression(precLimit int) ast.Expression {
	left := parser.ParseExpressionLeft()

	for {
		op := parser.TokAfter
		if op.Type != token.Operator {
			break
		}
		if precedence[op.Value.(string)] <= precLimit {
			break
		}
		parser.NextToken()
		parser.NextToken()
		newLimit := precLimit
		if isRightAssociative[op.Value.(string)] {
			precLimit++
		}
		right := parser.ParseExpression(newLimit)
		left = &ast.BinaryExpression{left, op, right}
	}

	return left
}

func (parser *Parser) ParseExpressionLeft() ast.Expression {
	tok := parser.CurTok
	switch tok.Type {
	case token.Int:
		return &ast.LiteralExpression{tok}
	case token.Float:
		return &ast.LiteralExpression{tok}
	case token.String:
		return &ast.LiteralExpression{tok}
	case token.Operator:
		parser.NextToken()
		return &ast.UnaryExpression{tok, parser.ParseExpression(precedence[tok.Value.(string)])}
	case token.OpenParen:
		parser.NextToken()
		expr := parser.ParseExpression(0)
		parser.NextToken()
		return expr
	}
	return &ast.InvalidExpression{}
}
