package parser

import (
	"fmt"
	"slices"

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
		newLimit := precedence[op.Value.(string)]
		if isRightAssociative[op.Value.(string)] {
			newLimit--
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
	case token.Identifier:
		switch parser.TokAfter.Type {
		case token.OpenParen:
			functionCall := parser.ParseFunctionCall()
			return &functionCall
		case token.OpenBracket:
		default:
			return &ast.VariableExpression{parser.CurTok}
		}
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

func (parser *Parser) ParseNamespace() ast.Namespace {
	namespace := ast.Namespace{}
	for parser.CurTok.Type != token.Keyword && parser.CurTok.Value != "namespace" {
		if parser.CurTok.Type == token.EOF {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error: no namespace defined"))
			return ast.Namespace{}
		}
		parser.NextToken()
	}
	parser.NextToken()
	if parser.CurTok.Type == token.Identifier {
		namespace.Name = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error: namespace not named"))
		return ast.Namespace{}
	}
	parser.NextToken()
	namespace.Body = parser.ParseBody()
	return namespace
}

func (parser *Parser) ParseBody() ast.BodyStatement {
	body := ast.BodyStatement{}
	var endToken token.TokenType
	if parser.CurTok.Type == token.OpenBrace {
		endToken = token.CloseBrace
		parser.NextToken()
	} else {
		endToken = token.EOF
	}
	for parser.CurTok.Type != endToken {
		var node ast.Node
		isNewIns := false
		switch parser.CurTok.Type {
		case token.Keyword:
			switch parser.CurTok.Value {
			case "int":
				node = parser.ParseVariableDeclaration()
				break
			case "float":
				node = parser.ParseVariableDeclaration()
				break
			case "string":
				node = parser.ParseVariableDeclaration()
				break
			case "func":
				node = parser.ParseFunctionDeclaration()
				break
			}
			break
		case token.Identifier:
			switch parser.TokAfter.Type {
			case token.Equal:
				node = parser.ParseVariableAssigment()
				break
			case token.OpenParen:
				node = parser.ParseFunctionCall()
				break
			case token.OpenBracket:
				break
			}
			break
		case token.String:
			break
		case token.Int:
			break
		case token.Float:
			break
		case token.NewInstruction:
			isNewIns = true
			break
		}
		parser.NextToken()
		if !isNewIns {
			body.Nodes = append(body.Nodes, node)
		}
	}

	return body
}

func (parser *Parser) ParseVariableDeclaration() ast.VariableDeclarationStatement {
	var variable ast.VariableDeclarationStatement
	variable.Type = parser.CurTok
	if parser.TokAfter.Type == token.Identifier {
		parser.NextToken()
		variable.Identifier = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Syntax error at line %d: expected identifier, got %s", parser.TokAfter.Line, parser.TokAfter.Value))
		return ast.VariableDeclarationStatement{}
	}
	if parser.TokAfter.Type == token.Equal {
		parser.NextToken()
		parser.NextToken()
		variable.Value = parser.ParseExpression(0)
	}
	return variable
}

func (parser *Parser) ParseVariableAssigment() ast.VariableAssignmentStatement {
	var variable ast.VariableAssignmentStatement
	variable.Target = parser.CurTok
	parser.NextToken()
	parser.NextToken()
	variable.Value = parser.ParseExpression(0)
	return variable
}

func (parser *Parser) ParseFunctionCall() ast.FunctionCall {
	var function ast.FunctionCall
	function.Identifier = parser.CurTok
	parser.NextToken()
	line := parser.CurTok.Line
	for parser.CurTok.Type != token.CloseParen && parser.CurTok.Type != token.EOF {
		parser.NextToken()
		if parser.CurTok.Type == token.CloseParen {
			break
		}
		function.Args = append(function.Args, parser.ParseExpression(0))
		parser.NextToken()
		if parser.CurTok.Type != token.Comma && parser.CurTok.Type != token.CloseParen {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected , or ), got %s", parser.CurTok.Line, parser.CurTok.Value))
		}
	}
	if parser.CurTok.Type == token.EOF {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: func call not closed", line))
		return ast.FunctionCall{}
	}
	return function
}

func (parser *Parser) ParseFunctionDeclaration() ast.FunctionDeclarationStatement {
	var function ast.FunctionDeclarationStatement
	if parser.TokAfter.Type == token.Identifier {
		parser.NextToken()
		function.Name = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected identifier, got %s", parser.CurTok.Line, parser.TokAfter.Value))
		return ast.FunctionDeclarationStatement{}
	}
	parser.NextToken()
	for parser.CurTok.Type != token.CloseParen && parser.CurTok.Type != token.EOF {
		parser.NextToken()
		if parser.CurTok.Type == token.CloseParen {
			break
		}
		function.Args = append(function.Args, parser.ParseVariableDeclaration())
		parser.NextToken()
		if parser.CurTok.Type != token.Comma && parser.CurTok.Type != token.CloseParen {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected , or ), got %s", parser.CurTok.Line, parser.CurTok.Value))
			return ast.FunctionDeclarationStatement{}
		}
	}
	parser.NextToken()
	if parser.CurTok.Type == token.Keyword && slices.Contains(token.Types, parser.CurTok.Value.(string)) {
		function.Type = parser.CurTok
		parser.NextToken()
	}
	if parser.CurTok.Type == token.OpenBrace {
		if function.Type.Type == "" {
			function.Type = token.Token{Type: token.Keyword, Value: "void"}
		}
		function.Body = parser.ParseBody()
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected type or {, got %s", parser.CurTok.Line, parser.CurTok.Value))
		return ast.FunctionDeclarationStatement{}
	}
	return function
}
