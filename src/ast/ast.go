package ast

import (
	"fmt"
	"lizalang/token"
	"strconv"
)

type Program struct {
	Namespaces map[string]BodyStatement
}

type Node interface{}

type Expression interface {
	Node
	expr()
	String() string
}

type BinaryExpression struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

func (*BinaryExpression) expr() {}
func (be *BinaryExpression) String() string {
	if be.Op.Value == "[" {
		return fmt.Sprintf("(%s %s %s])", be.Left.String(), be.Op.Value.(string), be.Right.String())
	}
	return fmt.Sprintf("(%s %s %s)", be.Left.String(), be.Op.Value.(string), be.Right.String())
}

type UnaryExpression struct {
	Prefix token.Token
	Value  Expression
}

func (*UnaryExpression) expr() {}
func (ue *UnaryExpression) String() string {
	return fmt.Sprintf("%s(%s)", ue.Prefix.Value.(string), ue.Value.String())
}

type FunctionCall struct {
	Identifier token.Token
	Args       []Expression
}

func (*FunctionCall) expr() {}
func (fc *FunctionCall) String() string {
	args := ""
	for i, arg := range fc.Args {
		if i == 0 {
			args += arg.String()
		} else {
			args += "," + arg.String()
		}
	}
	return fmt.Sprintf("%s(%s)", fc.Identifier.Value.(string), args)
}

type VariableExpression struct {
	Value token.Token
}

func (*VariableExpression) expr() {}
func (ve *VariableExpression) String() string {
	return ve.Value.Value.(string)
}

type LiteralExpression struct {
	Value token.Token
}

func (*LiteralExpression) expr() {}
func (le *LiteralExpression) String() string {
	id := ""
	switch le.Value.Value.(type) {
	case (int64):
		id = strconv.Itoa(int(le.Value.Value.(int64)))
	case (string):
		id = le.Value.Value.(string)
	case (float64):
		id = fmt.Sprintf("%f", le.Value.Value.(float64))
	}
	return id
}

type ArrayExpression struct {
	Elements []Expression
}

func (ae *ArrayExpression) String() string {
	var elements string
	for i, element := range ae.Elements {
		if i != 0 {
			elements += ", "
		}
		elements += element.String()
	}
	return fmt.Sprintf("[%s]", elements)
}
func (*ArrayExpression) expr() {}

type InvalidExpression struct{}

func (*InvalidExpression) expr()          {}
func (*InvalidExpression) String() string { return "()" }

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
	Body        Node
	Alternative Node
}

func (*IfStatement) stmt() {}

type ReturnStatement struct {
	ReturnValue Expression
}

func (*ReturnStatement) stmt() {}

type FunctionDeclarationStatement struct {
	Name token.Token
	Type Type
	Args []VariableDeclarationStatement
	Body Node
}

func (*FunctionDeclarationStatement) stmt() {}

type VariableDeclarationStatement struct {
	Identifier token.Token
	Type       Type
	Value      Expression
	Mutable    bool
}

func (*VariableDeclarationStatement) stmt() {}

type VariableAssignmentStatement struct {
	Target Expression
	Value  Expression
}

func (*VariableAssignmentStatement) stmt() {}

type ForStatement struct {
	Init      VariableDeclarationStatement
	Condition Expression
	Post      VariableAssignmentStatement
	Body      Node
}

func (*ForStatement) stmt() {}

type Type interface {
	T() string
}

type Int token.Token

func (*Int) T() string { return "int" }

type Float token.Token

func (*Float) T() string { return "float" }

type String token.Token

func (*String) T() string { return "string" }

type Bool token.Token

func (*Bool) T() string { return "bool" }

type Array struct {
	Type Type
	Size Expression
}

func (a *Array) T() string { return fmt.Sprintf("array of %s", a.Type.T()) }

type Void token.Token

func (*Void) T() string { return "void" }
