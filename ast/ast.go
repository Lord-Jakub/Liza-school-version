package ast

import "lizalang/token"

type Program struct {
	Namespaces []Namespace
}
type Namespace struct {
	Name token.Token
	Body BodyStatement
}

type Node interface{}

type Expression interface {
	Node
	expr()
}

type BinaryExpression struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

func (*BinaryExpression) expr() {}

type UnaryExpression struct {
	Prefix token.Token
	Value  Expression
}

func (UnaryExpression) expr() {}

type FunctionCall struct {
	Identifier token.Token
	Args       []Expression
}

func (*FunctionCall) expr() {}

type IdentifierExpression struct {
	Value token.Token
}

func (*IdentifierExpression) expr() {}

type LiteralExpression struct {
	Value token.Token
}

func (*LiteralExpression) expr() {}

type InvalidExpression struct{}

func (*InvalidExpression) expr() {}

type Statement interface {
	Node
	stmt()
}

type BodyStatement struct {
	Nodes []Node
}

func (*BodyStatement) stmt() {}

type IfStatement struct {
	Condition   Expression
	Body        BodyStatement
	Alternative BodyStatement
}

func (*IfStatement) stmt() {}

type ReturnStatement struct {
	ReturnValue Expression
}

func (*ReturnStatement) stmt() {}

type FunctionDeclarationStatement struct {
	Name token.Token
	Type token.Token
	Args any // idk, i will solve this later
	Body BodyStatement
}

func (*FunctionDeclarationStatement) stmt() {}

type VariableDeclarationStatement struct {
	Identifier token.Token
	Type       token.Token
	Value      Expression
}

func (*VariableDeclarationStatement) stmt() {}

type VariableAssignmentStatement struct {
	Target IdentifierExpression
	Value  Expression
}

func (*VariableAssignmentStatement) stmt() {}

type ForStatement struct {
	Init      Statement
	Condition Expression
	Post      Statement
	Body      BodyStatement
}

func (*ForStatement) stmt() {}
