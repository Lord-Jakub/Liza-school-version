package parser

import (
	"fmt"
	"lizalang/ast"
	"lizalang/token"
	"slices"
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
		Program: ast.Program{Namespaces: make(map[string]ast.BodyStatement, 0)},
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
			parser.Advance()
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

func (parser *Parser) Parse() {
	namespaceName := ""
	for parser.CurTok.Type != token.Keyword && parser.CurTok.Value != "namespace" {
		if parser.CurTok.Type == token.EOF {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error: no namespace defined"))
		}
		parser.Advance()
	}
	parser.Advance()
	if parser.CurTok.Type == token.Identifier {
		namespaceName = parser.CurTok.Value.(string)
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error: namespace not named"))
	}
	parser.Advance()
	for parser.CurTok.Type == token.NewInstruction {
		parser.Advance()
	}
	for parser.CurTok.Type == token.Keyword && parser.CurTok.Value.(string) == "import" {
		if parser.NextTok.Type == token.String {
			parser.Advance()
			if _, ok := parser.Program.Namespaces[parser.CurTok.Value.(string)]; ok {
				// TODO: add namespace to namespaces to parse queue
			}
		} else {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected string, got %s", parser.NextTok.Line, parser.NextTok.Value))
			parser.Advance()
			for parser.CurTok.Type == token.NewInstruction {
				parser.Advance()
			}
		}
	}
	namespaceBody, ok := parser.Program.Namespaces[namespaceName]
	if !ok {
		parser.Program.Namespaces[namespaceName] = ast.BodyStatement{}
	}

	namespaceBody.Nodes = append(parser.Program.Namespaces[namespaceName].Nodes, parser.ParseBody().Nodes...)
	parser.Program.Namespaces[namespaceName] = namespaceBody
	// TODO: look for namespaces in queue and parse them
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
			if slices.Contains(token.Types, parser.CurTok.Value.(string)) {
				node = parser.ParseVariableDeclaration()
			}
			switch parser.CurTok.Value {
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
			case "constant":
				parser.Advance()
				constant := parser.ParseVariableDeclaration()
				constant.Mutable = false
				node = constant
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
	variable.Mutable = true
	variable.Type = parser.ParseType()
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
		function.Type = parser.ParseType()
		parser.Advance()
	}
	if parser.CurTok.Type == token.OpenBrace {
		if function.Type == nil {
			function.Type = &ast.Void{}
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

func (parser *Parser) ParseType() ast.Type {
	t := parser.ParseSimpleType()
	for parser.NextTok.Type == token.OpenBracket {
		parser.Advance()
		parser.Advance()
		var a ast.Array
		if parser.CurTok.Type != token.CloseBracket {
			a.Size = parser.ParseExpression(0)
			parser.Advance()
		}
		a.Type = t
		t = &a
	}
	return t
}

func (parser *Parser) ParseSimpleType() ast.Type {
	typ, ok := parser.CurTok.Value.(string)
	if !ok {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected type, got %s", parser.CurTok.Line, parser.CurTok.Value))
		return &ast.Void{}
	}
	switch typ {
	case "int":
		i := ast.Int(parser.CurTok)
		return &i
	case "float":
		f := ast.Float(parser.CurTok)
		return &f
	case "string":
		s := ast.String(parser.CurTok)
		return &s
	case "bool":
		b := ast.Bool(parser.CurTok)
		return &b
	case "void":
		v := ast.Void(parser.CurTok)
		return &v
	default:
		parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected type, got %s", parser.CurTok.Line, parser.CurTok.Value))
		return &ast.Void{}
	}
}
