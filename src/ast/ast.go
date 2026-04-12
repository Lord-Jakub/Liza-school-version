package ast

import (
	"fmt"
	"lizalang/token"
	"strconv"
)

// Program represents the root of the AST.
type Program struct {
	Namespaces map[string]BodyStatement
}

func (p *Program) Line() int {
	for _, ns := range p.Namespaces {
		return ns.Line()
	}
	return 0
}

func (p *Program) File() string {
	for _, ns := range p.Namespaces {
		return ns.File()
	}
	return ""
}

// Node is a marker interface for all AST nodes.
type Node interface {
	Line() int
	File() string
}

type Expression interface {
	Node
	expr()
	String() string
}

// BinaryExpression

type BinaryExpression struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

func (*BinaryExpression) expr() {}

func (be *BinaryExpression) Line() int {
	return be.Op.Line
}

func (be *BinaryExpression) File() string {
	return be.Op.File
}

func (be *BinaryExpression) String() string {
	if be.Op.Value == "[" {
		return fmt.Sprintf("(%s %s %s])", be.Left.String(), be.Op.Value.(string), be.Right.String())
	}
	return fmt.Sprintf("(%s %s %s)", be.Left.String(), be.Op.Value.(string), be.Right.String())
}

// UnaryExpression

type UnaryExpression struct {
	Prefix token.Token
	Value  Expression
}

func (*UnaryExpression) expr() {}

func (ue *UnaryExpression) Line() int {
	return ue.Prefix.Line
}

func (ue *UnaryExpression) File() string {
	return ue.Prefix.File
}

func (ue *UnaryExpression) String() string {
	return fmt.Sprintf("%s(%s)", ue.Prefix.Value.(string), ue.Value.String())
}

// FunctionCall

type FunctionCall struct {
	Identifier token.Token
	Args       []Expression
}

func (*FunctionCall) expr() {}

func (fc FunctionCall) Line() int {
	return fc.Identifier.Line
}

func (fc FunctionCall) File() string {
	return fc.Identifier.File
}

func (fc *FunctionCall) String() string {
	args := ""
	for i, arg := range fc.Args {
		if i != 0 {
			args += ","
		}
		args += arg.String()
	}
	return fmt.Sprintf("%s(%s)", fc.Identifier.Value.(string), args)
}

// VariableExpression

type VariableExpression struct {
	Value token.Token
}

func (*VariableExpression) expr() {}

func (ve *VariableExpression) Line() int {
	return ve.Value.Line
}

func (ve *VariableExpression) File() string {
	return ve.Value.File
}

func (ve *VariableExpression) String() string {
	return ve.Value.Value.(string)
}

// LiteralExpression

type LiteralExpression struct {
	Value token.Token
}

func (*LiteralExpression) expr() {}

func (le *LiteralExpression) Line() int {
	return le.Value.Line
}

func (le *LiteralExpression) File() string {
	return le.Value.File
}

func (le *LiteralExpression) String() string {
	switch v := le.Value.Value.(type) {
	case int64:
		return strconv.Itoa(int(v))
	case string:
		return v
	case float64:
		return fmt.Sprintf("%f", v)
	}
	return ""
}

// ArrayExpression

type ArrayExpression struct {
	Elements []Expression
}

func (*ArrayExpression) expr() {}

func (ae *ArrayExpression) Line() int {
	if len(ae.Elements) > 0 {
		return ae.Elements[0].Line()
	}
	return 0
}

func (ae *ArrayExpression) File() string {
	if len(ae.Elements) > 0 {
		return ae.Elements[0].File()
	}
	return ""
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

// InvalidExpression

type InvalidExpression struct{}

func (*InvalidExpression) expr() {}

func (*InvalidExpression) Line() int { return 0 }

func (*InvalidExpression) File() string { return "" }

func (*InvalidExpression) String() string { return "()" }

type Statement interface {
	Node
	stmt()
}

// BodyStatement

type BodyStatement struct {
	Nodes []Node
}

func (*BodyStatement) stmt() {}

func (bs BodyStatement) Line() int {
	if len(bs.Nodes) > 0 {
		return bs.Nodes[0].Line()
	}
	return 0
}

func (bs BodyStatement) File() string {
	if len(bs.Nodes) > 0 {
		return bs.Nodes[0].File()
	}
	return ""
}

// IfStatement

type IfStatement struct {
	Condition   Expression
	Body        Node
	Alternative Node
}

func (*IfStatement) stmt() {}

func (is IfStatement) Line() int {
	return is.Condition.Line()
}

func (is IfStatement) File() string {
	return is.Condition.File()
}

// ReturnStatement

type ReturnStatement struct {
	ReturnValue Expression
}

func (*ReturnStatement) stmt() {}

func (rs ReturnStatement) Line() int {
	return rs.ReturnValue.Line()
}

func (rs ReturnStatement) File() string {
	return rs.ReturnValue.File()
}

// FunctionDeclarationStatement

type FunctionDeclarationStatement struct {
	Name token.Token
	Type Type
	Args []VariableDeclarationStatement
	Body Node
}

func (*FunctionDeclarationStatement) stmt() {}

func (fds FunctionDeclarationStatement) Line() int {
	return fds.Name.Line
}

func (fds FunctionDeclarationStatement) File() string {
	return fds.Name.File
}

// VariableDeclarationStatement

type VariableDeclarationStatement struct {
	Identifier token.Token
	Type       Type
	Value      Expression
	Mutable    bool
}

func (*VariableDeclarationStatement) stmt() {}

func (vds VariableDeclarationStatement) Line() int {
	return vds.Identifier.Line
}

func (vds VariableDeclarationStatement) File() string {
	return vds.Identifier.File
}

// VariableAssignmentStatement

type VariableAssignmentStatement struct {
	Target Expression
	Value  Expression
}

func (*VariableAssignmentStatement) stmt() {}

func (vas VariableAssignmentStatement) Line() int {
	return vas.Target.Line()
}

func (vas VariableAssignmentStatement) File() string {
	return vas.Target.File()
}

// ForStatement

type ForStatement struct {
	Init      VariableDeclarationStatement
	Condition Expression
	Post      VariableAssignmentStatement
	Body      Node
}

func (*ForStatement) stmt() {}

func (fs ForStatement) Line() int {
	return fs.Init.Line()
}

func (fs ForStatement) File() string {
	return fs.Init.File()
}

type Type interface {
	T() string
}

// Primitive types

type Int token.Token

func (*Int) T() string { return "int" }

type Float token.Token

func (*Float) T() string { return "float" }

type String token.Token

func (*String) T() string { return "string" }

type Bool token.Token

func (*Bool) T() string { return "bool" }

// Array type

type Array struct {
	Type Type
	Size Expression
}

func (a Array) Line() int {
	if a.Size != nil {
		return a.Size.Line()
	}
	return 0
}

func (a Array) File() string {
	if a.Size != nil {
		return a.Size.File()
	}
	return ""
}

func (a *Array) T() string {
	return fmt.Sprintf("array of %s", a.Type.T())
}

// Void

type Void token.Token

func (*Void) T() string { return "void" }
