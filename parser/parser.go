package parser

import (
	"fmt"
	"slices"

	"lizalang/ast"
	"lizalang/token"
)

type Parser struct {
	Tokens  []token.Token
	Pos     int
	CurTok  token.Token
	NextTok token.Token
	Program ast.Program
	Errors  []error
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
		parser.NextTok = tokens[1]
	}

	return parser
}

func (parser *Parser) Advance() {
	parser.Pos++
	parser.CurTok = parser.NextTok
	if parser.NextTok.Type != token.EOF {
		parser.NextTok = parser.Tokens[parser.Pos+1]
	}
}

var (
	precedence         = map[string]int{"[": 1, "&&": 2, "||": 2, "==": 3, "!=": 3, "<=": 3, ">=": 3, "<": 3, ">": 3, "+": 4, "-": 4, "*": 5, "/": 5, "^": 6}
	isRightAssociative = map[string]bool{"[": false, "&&": false, "||": false, "==": false, "!=": false, "<=": false, ">=": false, "<": false, ">": false, "+": false, "-": false, "*": false, "/": false, "^": true}
)

func (parser *Parser) ParseExpression(precLimit int) ast.Expression {
	left := parser.ParseExpressionLeft()

	for {
		op := parser.NextTok

		if op.Type != token.Operator && op.Type != token.OpenBracket {
			break
		}
		if precedence[op.Value.(string)] <= precLimit {
			break
		}
		parser.Advance()
		parser.Advance()
		newLimit := precedence[op.Value.(string)]
		if isRightAssociative[op.Value.(string)] {
			newLimit--
		}
		if op.Type == token.OpenBracket {
			newLimit = 0
		}
		right := parser.ParseExpression(newLimit)
		if parser.NextTok.Type == token.CloseBracket {
			parser.Advance()
		}
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
		switch parser.NextTok.Type {
		case token.OpenParen:
			functionCall := parser.ParseFunctionCall()
			return &functionCall
		// case token.OpenBracket:
		default:
			return &ast.VariableExpression{parser.CurTok}
		}
	case token.Operator:
		parser.Advance()
		return &ast.UnaryExpression{tok, parser.ParseExpression(precedence[tok.Value.(string)])}
	case token.OpenParen:
		parser.Advance()
		expr := parser.ParseExpression(0)
		parser.Advance()
		return expr
	case token.OpenBracket:
		parser.Advance()
		var arr ast.ArrayExpression
		if parser.CurTok.Type != token.CloseBracket {
			arr.Elements = append(arr.Elements, parser.ParseExpression(0))
			parser.Advance() // TODO: error handling if token to consume isn't comma
		}
		for parser.CurTok.Type == token.Comma {
			parser.Advance()
			arr.Elements = append(arr.Elements, parser.ParseExpression(0))

			parser.Advance()
		}
		return &arr
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
		parser.Advance()
	}
	parser.Advance()
	if parser.CurTok.Type == token.Identifier {
		namespace.Name = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error: namespace not named"))
		return ast.Namespace{}
	}
	parser.Advance()
	namespace.Body = parser.ParseBody()
	return namespace
}

func (parser *Parser) ParseBody() ast.BodyStatement {
	body := ast.BodyStatement{}
	var endToken token.TokenType
	if parser.CurTok.Type == token.OpenBrace {
		endToken = token.CloseBrace
		parser.Advance()
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
			case "if":
				node = parser.ParseIfStatement()
				break
			case "for":
				node = parser.ParseForStatement()
				break
			case "return":
				node = parser.ParseReturnStatement()
				break
			}
			break
		case token.Identifier:
			switch parser.NextTok.Type {
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
		parser.Advance()
		if !isNewIns {
			body.Nodes = append(body.Nodes, node)
		}
	}

	return body
}

func (parser *Parser) ParseVariableDeclaration() ast.VariableDeclarationStatement {
	var variable ast.VariableDeclarationStatement
	variable.Type = parser.CurTok
	if parser.NextTok.Type == token.Identifier {
		parser.Advance()
		variable.Identifier = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Syntax error at line %d: expected identifier, got %s", parser.NextTok.Line, parser.NextTok.Value))
		return ast.VariableDeclarationStatement{}
	}
	if parser.NextTok.Type == token.Equal {
		parser.Advance()
		parser.Advance()
		variable.Value = parser.ParseExpression(0)
	}
	return variable
}

func (parser *Parser) ParseVariableAssigment() ast.VariableAssignmentStatement {
	var variable ast.VariableAssignmentStatement
	variable.Target = parser.CurTok
	parser.Advance()
	parser.Advance()
	variable.Value = parser.ParseExpression(0)
	return variable
}

func (parser *Parser) ParseFunctionCall() ast.FunctionCall {
	var function ast.FunctionCall
	function.Identifier = parser.CurTok
	parser.Advance()
	line := parser.CurTok.Line
	for parser.CurTok.Type != token.CloseParen && parser.CurTok.Type != token.EOF {
		parser.Advance()
		if parser.CurTok.Type == token.CloseParen {
			break
		}
		function.Args = append(function.Args, parser.ParseExpression(0))
		parser.Advance()
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
	if parser.NextTok.Type == token.Identifier {
		parser.Advance()
		function.Name = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected identifier, got %s", parser.CurTok.Line, parser.NextTok.Value))
		return ast.FunctionDeclarationStatement{}
	}
	parser.Advance()
	for parser.CurTok.Type != token.CloseParen && parser.CurTok.Type != token.EOF {
		parser.Advance()
		if parser.CurTok.Type == token.CloseParen {
			break
		}
		function.Args = append(function.Args, parser.ParseVariableDeclaration())
		parser.Advance()
		if parser.CurTok.Type != token.Comma && parser.CurTok.Type != token.CloseParen {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected , or ), got %s", parser.CurTok.Line, parser.CurTok.Value))
			return ast.FunctionDeclarationStatement{}
		}
	}
	parser.Advance()
	if parser.CurTok.Type == token.Keyword && slices.Contains(token.Types, parser.CurTok.Value.(string)) {
		function.Type = parser.CurTok
		parser.Advance()
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

func (parser *Parser) ParseIfStatement() ast.IfStatement {
	var ifStatement ast.IfStatement
	parser.Advance()
	ifStatement.Condition = parser.ParseExpression(0)
	line := parser.CurTok.Line
	for parser.CurTok.Type != token.OpenBrace {
		parser.Advance()
		if parser.CurTok.Type == token.EOF {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: missing {", line))
		}
	}
	ifStatement.Body = parser.ParseBody()
	for parser.NextTok.Type == token.NewInstruction {
		parser.Advance()
	}
	if parser.NextTok.Type == token.Keyword && parser.NextTok.Value == "else" {
		parser.Advance()
		if parser.NextTok.Type == token.Keyword && parser.NextTok.Value == "if" {
			parser.Advance()
			ifStatement.Alternative.Nodes = append(ifStatement.Alternative.Nodes, parser.ParseIfStatement())
		} else {
			for parser.CurTok.Type != token.OpenBrace {
				parser.Advance()
			}
			ifStatement.Alternative = parser.ParseBody()
		}
	}
	return ifStatement
}

func (parser *Parser) ParseReturnStatement() ast.ReturnStatement {
	var ret ast.ReturnStatement
	parser.Advance()
	ret.ReturnValue = parser.ParseExpression(0)
	return ret
}

func (parser *Parser) ParseForStatement() ast.ForStatement {
	var forStmt ast.ForStatement
	parser.Advance()
	for parser.CurTok.Type == token.NewInstruction {
		parser.Advance()
	}

	if parser.CurTok.Type == token.Keyword {
		if slices.Contains(token.Types, parser.CurTok.Value.(string)) {
			forStmt.Init = parser.ParseVariableDeclaration()
		} else {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected variable declaration, got %s", parser.CurTok.Line, parser.CurTok.Type))
			return ast.ForStatement{}
		}
		if parser.NextTok.Type != token.NewInstruction {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected ;, got %s", parser.NextTok.Line, parser.NextTok.Type))
			return ast.ForStatement{}
		}
		parser.Advance()

		parser.Advance()
	}
	forStmt.Condition = parser.ParseExpression(0)
	if parser.NextTok.Type != token.NewInstruction && parser.NextTok.Type != token.OpenBrace {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected ; or {, got %s", parser.NextTok.Line, parser.NextTok.Type))
		return ast.ForStatement{}
	}
	parser.Advance()

	if parser.NextTok.Type == token.Identifier {
		parser.Advance()
		forStmt.Post = parser.ParseVariableAssigment()
	}
	line := parser.CurTok.Line
	for parser.CurTok.Type != token.OpenBrace && parser.CurTok.Type != token.EOF {
		if parser.CurTok.Type == token.EOF {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected { got EOF", line))
			return ast.ForStatement{}
		}
		parser.Advance()
	}
	forStmt.Body = parser.ParseBody()
	return forStmt
}
