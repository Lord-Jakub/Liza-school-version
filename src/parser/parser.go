package parser

import (
	"fmt"
	"lizalang/ast"
	"lizalang/lexer"
	"lizalang/token"
	"lizalang/utils"
	"os"
	"slices"
)

// Parser transforms a token stream into an AST.
// It maintains a cursor with lookahead and accumulates errors.
type Parser struct {
	Tokens  []token.Token
	Pos     int
	CurTok  token.Token
	NextTok token.Token
	Program ast.Program
	Errors  []error
}

// New initializes a parser with token stream and prepares lookahead.
func New(tokens []token.Token) *Parser {
	parser := &Parser{
		Tokens:  tokens,
		Pos:     0,
		Program: ast.Program{Namespaces: make(map[string]ast.BodyStatement, 0)},
		Errors:  make([]error, 0),
	}

	// Initialize current and next token
	if len(tokens) > 2 {
		parser.CurTok = tokens[0]
		parser.NextTok = tokens[1]
	}

	return parser
}

// Advance moves parser forward by one token.
func (parser *Parser) Advance() {
	parser.Pos++
	parser.CurTok = parser.NextTok

	// Maintain lookahead unless EOF reached
	if parser.NextTok.Type != token.EOF {
		parser.NextTok = parser.Tokens[parser.Pos+1]
	}
}

// Operator precedence table used by Pratt parser.
var (
	precedence = map[string]int{
		"[":  100,
		"&&": 2, "||": 2,
		"==": 3, "!=": 3, "<=": 3, ">=": 3, "<": 3, ">": 3,
		"+": 4, "-": 4,
		"*": 5, "/": 5, "%": 5,
		"^": 6,
		".": 7,
	}

	// Defines associativity for operators
	isRightAssociative = map[string]bool{
		"[":  false,
		"&&": false, "||": false,
		"==": false, "!=": false, "<=": false, ">=": false, "<": false, ">": false,
		"+": false, "-": false,
		"*": false, "/": false, "%": false,
		"^": true,
		".": true,
	}
)

// ParseExpression implements Pratt parsing.
// precLimit controls how far the parser can bind operators.
func (parser *Parser) ParseExpression(precLimit int) ast.Expression {
	left := parser.ParseExpressionLeft()

	for {
		op := parser.NextTok

		// Only continue if next token is a valid infix operator
		if op.Type != token.Operator && op.Type != token.OpenBracket && op.Type != token.Dot {
			break
		}

		// Stop if precedence is too low
		if precedence[op.Value.(string)] <= precLimit {
			break
		}

		// Consume operator
		parser.Advance()
		parser.Advance()

		newLimit := precedence[op.Value.(string)]
		if isRightAssociative[op.Value.(string)] {
			newLimit-- // allow right-binding
		}

		var right ast.Expression

		// Special handling for array indexing: a[b]
		if op.Type == token.OpenBracket {
			right = parser.ParseExpression(0)
			if parser.NextTok.Type == token.CloseBracket {
				parser.Advance()
			}
		} else {
			right = parser.ParseExpression(newLimit)
		}

		// Build binary AST node
		left = &ast.BinaryExpression{left, op, right}
	}

	return left
}

// ParseExpressionLeft parses prefix expressions
func (parser *Parser) ParseExpressionLeft() ast.Expression {
	tok := parser.CurTok

	switch tok.Type {
	case token.Int, token.Float, token.String:
		// Literal values
		return &ast.LiteralExpression{tok}

	case token.Identifier:
		// Variable or function call
		switch parser.NextTok.Type {
		case token.OpenParen:
			functionCall := parser.ParseFunctionCall()
			return &functionCall
		default:
			return &ast.VariableExpression{parser.CurTok}
		}

	case token.Operator:
		// Prefix unary operator
		parser.Advance()
		return &ast.UnaryExpression{tok, parser.ParseExpression(precedence[tok.Value.(string)])}

	case token.OpenParen:
		// Parenthesized expression
		parser.Advance()
		expr := parser.ParseExpression(0)
		parser.Advance()
		return expr

	case token.OpenBracket:
		// Array literal
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

// Parse parses the entire file including namespace and imports.
func (parser *Parser) Parse(root, path string) {
	// Seek "namespace" declaration
	for parser.CurTok.Type != token.Keyword && parser.CurTok.Value != "namespace" {
		if parser.CurTok.Type == token.EOF {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error: no namespace defined"))
			return
		}
		parser.Advance()
	}

	parser.Advance()

	// Parse namespace name
	namespaceName := ""
	if parser.CurTok.Type == token.Identifier {
		if val, ok := parser.CurTok.Value.(string); ok {
			namespaceName = val
		}
	} else {
		parser.Errors = append(parser.Errors, fmt.Errorf("Error: namespace not named"))
	}
	parser.Advance()

	for parser.CurTok.Type == token.NewInstruction {
		parser.Advance()
	}

	// Parse imports
	var imports []string
	for parser.CurTok.Type == token.Keyword && parser.CurTok.Value == "import" {
		if parser.NextTok.Type == token.String {
			parser.Advance()
			imports = append(imports, parser.CurTok.Value.(string))
			parser.Advance()
		} else {
			parser.Errors = append(parser.Errors, fmt.Errorf("Error at line %d: expected string", parser.NextTok.Line))
			parser.Advance()
		}

		for parser.CurTok.Type == token.NewInstruction {
			parser.Advance()
		}
	}

	// Parse main body
	currentBody := parser.ParseBody()

	// Merge into namespace
	if _, ok := parser.Program.Namespaces[namespaceName]; !ok {
		parser.Program.Namespaces[namespaceName] = ast.BodyStatement{}
	}

	ns := parser.Program.Namespaces[namespaceName]
	ns.Nodes = append(ns.Nodes, currentBody.Nodes...)
	parser.Program.Namespaces[namespaceName] = ns

	// Recursively parse imported files
	for _, imported := range imports {
		files, err := utils.GetFilesOfDir(imported, root, path)
		if err != nil {
			parser.Errors = append(parser.Errors, err)
			continue
		}

		for _, file := range files {
			code, err := os.ReadFile(file)
			if err != nil {
				parser.Errors = append(parser.Errors, err)
				continue
			}

			// Lex and parse imported file
			lex := lexer.New(string(code), file)
			lex.Lex()

			if len(lex.Tokens) < 2 {
				continue
			}

			subParser := &Parser{
				Tokens:  lex.Tokens,
				Program: parser.Program,
				Pos:     0,
				CurTok:  lex.Tokens[0],
				NextTok: lex.Tokens[1],
			}

			subParser.Parse(root, path)
			parser.Errors = append(parser.Errors, subParser.Errors...)
		}
	}
}

// ParseNode parses a single top-level or block-level construct.
func (parser *Parser) ParseNode() ast.Node {
	for parser.CurTok.Type == token.NewInstruction {
		parser.Advance()
	}

	var node ast.Node

	switch parser.CurTok.Type {
	case token.Keyword:
		// Parse keywords
		if slices.Contains(token.Types, parser.CurTok.Value.(string)) {
			node = parser.ParseVariableDeclaration()
		}

		switch parser.CurTok.Value {
		case "func":
			node = parser.ParseFunctionDeclaration()
		case "if":
			node = parser.ParseIfStatement()
		case "for":
			node = parser.ParseForStatement()
		case "return":
			node = parser.ParseReturnStatement()
		case "constant":
			parser.Advance()
			constant := parser.ParseVariableDeclaration()
			constant.Mutable = false
			node = constant
		}

	case token.Identifier:
		// Assignment or function call
		switch parser.NextTok.Type {
		case token.Equal, token.OpenBracket:
			node = parser.ParseVariableAssigment()
		case token.OpenParen:
			node = parser.ParseFunctionCall()
		default:
			parser.Errors = append(parser.Errors,
				fmt.Errorf("Invalid token [%v] (of type %s) on a line %d, file %s",
					parser.CurTok.Value, parser.CurTok.Type, parser.CurTok.Line, parser.CurTok.File))
		}

	case token.OpenBrace:
		// Detect body of nodes marked by {}
		node = parser.ParseBody()

	case token.CloseBrace:
		return nil

	default:
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Invalid token [%v] (of type %s) on a line %d, file %s",
				parser.CurTok.Value, parser.CurTok.Type, parser.CurTok.Line, parser.CurTok.File))
	}

	return node
}

// ParseBody parses block of nodes
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
		node := parser.ParseNode()

		if parser.CurTok.Type == endToken {
			body.Nodes = append(body.Nodes, node)
			break
		}

		parser.Advance()
		body.Nodes = append(body.Nodes, node)
	}

	if parser.CurTok.Type == token.CloseBrace {
		parser.Advance()
	}

	return body
}

// ParseVariableDeclaration parses variable declarations.
func (parser *Parser) ParseVariableDeclaration() ast.VariableDeclarationStatement {
	var variable ast.VariableDeclarationStatement
	variable.Mutable = true

	variable.Type = parser.ParseType()

	if parser.NextTok.Type == token.CloseBracket {
		parser.Advance()
	}

	if parser.NextTok.Type == token.Identifier {
		parser.Advance()
		variable.Identifier = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Syntax error at line %d: expected identifier, got %s",
				parser.NextTok.Line, parser.NextTok.Value))
		return ast.VariableDeclarationStatement{}
	}

	if parser.NextTok.Type == token.NewInstruction {
		parser.Advance()
	}

	if parser.NextTok.Type == token.Equal {
		parser.Advance()
		parser.Advance()
		variable.Value = parser.ParseExpression(0)
	}

	return variable
}

// ParseVariableAssigment parses assignment expressions.
func (parser *Parser) ParseVariableAssigment() ast.VariableAssignmentStatement {
	var variable ast.VariableAssignmentStatement
	variable.Target = parser.ParseExpression(0)

	parser.Advance()
	parser.Advance()

	variable.Value = parser.ParseExpression(0)
	return variable
}

// ParseFunctionCall parses function call and its arguments.
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
			parser.Errors = append(parser.Errors,
				fmt.Errorf("Error at line %d: expected , or ), got %s",
					parser.CurTok.Line, parser.CurTok.Value))
		}
	}

	if parser.CurTok.Type == token.EOF {
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Error at line %d: func call not closed", line))
		return ast.FunctionCall{}
	}

	return function
}

// ParseFunctionDeclaration parses function definition including args and body.
func (parser *Parser) ParseFunctionDeclaration() ast.FunctionDeclarationStatement {
	var function ast.FunctionDeclarationStatement
	function.Type = &ast.Void{} // default return type

	if parser.NextTok.Type == token.Identifier {
		parser.Advance()
		function.Name = parser.CurTok
	} else {
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Error at line %d: expected identifier, got %s",
				parser.CurTok.Line, parser.NextTok.Value))
		return ast.FunctionDeclarationStatement{}
	}

	parser.Advance()

	// Parse arguments
	for parser.CurTok.Type != token.CloseParen && parser.CurTok.Type != token.EOF {
		parser.Advance()

		if parser.CurTok.Type == token.CloseParen {
			break
		}

		function.Args = append(function.Args, parser.ParseVariableDeclaration())
		parser.Advance()

		if parser.CurTok.Type != token.Comma && parser.CurTok.Type != token.CloseParen {
			parser.Errors = append(parser.Errors,
				fmt.Errorf("Error at line %d: expected , or ), got %s",
					parser.CurTok.Line, parser.CurTok.Value))
			return ast.FunctionDeclarationStatement{}
		}
	}

	parser.Advance()

	// Optional return type
	if parser.CurTok.Type == token.Keyword && slices.Contains(token.Types, parser.CurTok.Value.(string)) {
		function.Type = parser.ParseType()
		parser.Advance()
	}

	function.Body = parser.ParseNode()
	return function
}

// ParseIfStatement parses an if-else construct.
func (parser *Parser) ParseIfStatement() ast.IfStatement {
	var ifStatement ast.IfStatement

	parser.Advance()

	// Parse condition expression
	ifStatement.Condition = parser.ParseExpression(0)

	parser.Advance()

	// Parse main body
	ifStatement.Body = parser.ParseNode()

	for parser.NextTok.Type == token.NewInstruction {
		parser.Advance()
	}

	// Optional else branch
	if parser.NextTok.Type == token.Keyword && parser.NextTok.Value == "else" {
		parser.Advance()
		ifStatement.Alternative = parser.ParseNode()
	}

	return ifStatement
}

// ParseReturnStatement parses a return statement.
func (parser *Parser) ParseReturnStatement() ast.ReturnStatement {
	var ret ast.ReturnStatement

	parser.Advance()

	// Parse returned expression
	ret.ReturnValue = parser.ParseExpression(0)

	return ret
}

// ParseForStatement parses a for loop
func (parser *Parser) ParseForStatement() ast.ForStatement {
	var forStmt ast.ForStatement

	parser.Advance()

	for parser.CurTok.Type == token.NewInstruction {
		parser.Advance()
	}

	// Parse initialization (must be variable declaration)
	if parser.CurTok.Type == token.Keyword {
		if slices.Contains(token.Types, parser.CurTok.Value.(string)) {
			forStmt.Init = parser.ParseVariableDeclaration()
		} else {
			parser.Errors = append(parser.Errors,
				fmt.Errorf("Error at line %d: expected variable declaration, got %s",
					parser.CurTok.Line, parser.CurTok.Type))
			return ast.ForStatement{}
		}

		// Expect instruction separator (acts like ;)
		if parser.NextTok.Type != token.NewInstruction {
			parser.Errors = append(parser.Errors,
				fmt.Errorf("Error at line %d: expected ;, got %s",
					parser.NextTok.Line, parser.NextTok.Type))
			return ast.ForStatement{}
		}

		parser.Advance()
		parser.Advance()
	}

	// Parse loop condition
	forStmt.Condition = parser.ParseExpression(0)

	// Expect separator or start of body
	if parser.NextTok.Type != token.NewInstruction && parser.NextTok.Type != token.OpenBrace {
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Error at line %d: expected ; or {, got %s",
				parser.NextTok.Line, parser.NextTok.Type))
		return ast.ForStatement{}
	}

	parser.Advance()

	// Parse post-expression (assignment)
	if parser.NextTok.Type == token.Identifier {
		parser.Advance()
		forStmt.Post = parser.ParseVariableAssigment()
		parser.Advance()
	}

	// Parse loop body
	forStmt.Body = parser.ParseNode()

	return forStmt
}

// ParseType parses a (possibly nested, i.e. array) type.
func (parser *Parser) ParseType() ast.Type {
	t := parser.ParseSimpleType()

	// Handle array type suffixes
	for parser.NextTok.Type == token.OpenBracket {
		parser.Advance()
		parser.Advance()

		var a ast.Array

		// Optional array size expression
		if parser.CurTok.Type != token.CloseBracket {
			a.Size = parser.ParseExpression(0)
			parser.Advance()
		}

		a.Type = t
		t = &a // wrap previous type

	}

	return t
}

// ParseSimpleType parses a base (non-composite) type.
func (parser *Parser) ParseSimpleType() ast.Type {
	typ, ok := parser.CurTok.Value.(string)
	if !ok {
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Error at line %d: expected type, got %s",
				parser.CurTok.Line, parser.CurTok.Value))
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
		parser.Errors = append(parser.Errors,
			fmt.Errorf("Error at line %d: expected type, got %s",
				parser.CurTok.Line, parser.CurTok.Value))
		return &ast.Void{}
	}
}
