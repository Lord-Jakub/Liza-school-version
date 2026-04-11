package ast

import (
	"fmt"
	"lizalang/token"
	"strconv"
)

// Program represents the root of the AST.
// It organizes code into namespaces, each containing a body of statements.
type Program struct {
	Namespaces map[string]BodyStatement
}

// Node is a marker interface for all AST nodes.
type Node interface{}

// Expression represents any evaluatable construct in the language.
type Expression interface {
	Node
	expr()          // marker method to distinguish expressions
	String() string // returns a string representation (mainly for debugging)
}

// BinaryExpression represents a binary operation
type BinaryExpression struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

func (*BinaryExpression) expr() {}

// String returns a parenthesized representation of the binary expression.
func (be *BinaryExpression) String() string {
	// Special handling for array indexing syntax
	if be.Op.Value == "[" {
		return fmt.Sprintf("(%s %s %s])", be.Left.String(), be.Op.Value.(string), be.Right.String())
	}
	return fmt.Sprintf("(%s %s %s)", be.Left.String(), be.Op.Value.(string), be.Right.String())
}

// UnaryExpression represents prefix unary operations
type UnaryExpression struct {
	Prefix token.Token
	Value  Expression
}

func (*UnaryExpression) expr() {}

func (ue *UnaryExpression) String() string {
	return fmt.Sprintf("%s(%s)", ue.Prefix.Value.(string), ue.Value.String())
}

// FunctionCall represents a function invocation with arguments.
type FunctionCall struct {
	Identifier token.Token
	Args       []Expression
}

func (*FunctionCall) expr() {}

// String builds a comma-separated argument list.
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

// VariableExpression represents usage of a variable identifier.
type VariableExpression struct {
	Value token.Token
}

func (*VariableExpression) expr() {}

func (ve *VariableExpression) String() string {
	return ve.Value.Value.(string)
}

// LiteralExpression represents constant values (int, float, string, etc.).
type LiteralExpression struct {
	Value token.Token
}

func (*LiteralExpression) expr() {}

// String converts the literal token value into its string form.
func (le *LiteralExpression) String() string {
	id := ""
	switch le.Value.Value.(type) {
	case int64:
		id = strconv.Itoa(int(le.Value.Value.(int64)))
	case string:
		id = le.Value.Value.(string)
	case float64:
		id = fmt.Sprintf("%f", le.Value.Value.(float64))
	}
	return id
}

// ArrayExpression represents an array literal.
type ArrayExpression struct {
	Elements []Expression
}

func (*ArrayExpression) expr() {}

// String joins elements into a comma-separated list inside brackets.
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

// InvalidExpression represents a placeholder for malformed expressions.
type InvalidExpression struct{}

func (*InvalidExpression) expr()          {}
func (*InvalidExpression) String() string { return "()" }

// Statement represents any executable construct.
type Statement interface {
	Node
	stmt() // marker method to distinguish statements
}

// BodyStatement represents a block of nodes (statements or expressions).
type BodyStatement struct {
	Nodes []Node
}

func (*BodyStatement) stmt() {}

// IfStatement represents conditional branching with optional alternative.
type IfStatement struct {
	Condition   Expression
	Body        Node
	Alternative Node
}

func (*IfStatement) stmt() {}

// ReturnStatement represents returning a value from a function.
type ReturnStatement struct {
	ReturnValue Expression
}

func (*ReturnStatement) stmt() {}

// FunctionDeclarationStatement represents a function definition.
type FunctionDeclarationStatement struct {
	Name token.Token
	Type Type
	Args []VariableDeclarationStatement
	Body Node
}

func (*FunctionDeclarationStatement) stmt() {}

// VariableDeclarationStatement represents variable definition.
type VariableDeclarationStatement struct {
	Identifier token.Token
	Type       Type
	Value      Expression
	Mutable    bool
}

func (*VariableDeclarationStatement) stmt() {}

// VariableAssignmentStatement represents assigning a value to a target.
type VariableAssignmentStatement struct {
	Target Expression
	Value  Expression
}

func (*VariableAssignmentStatement) stmt() {}

// ForStatement represents a classical for loop with init, condition, and post.
type ForStatement struct {
	Init      VariableDeclarationStatement
	Condition Expression
	Post      VariableAssignmentStatement
	Body      Node
}

func (*ForStatement) stmt() {}

// Type represents a type in the language type system.
type Type interface {
	T() string // returns string representation of the type
}

// Primitive type wrappers (based on tokens)

type Int token.Token

func (*Int) T() string { return "int" }

type Float token.Token

func (*Float) T() string { return "float" }

type String token.Token

func (*String) T() string { return "string" }

type Bool token.Token

func (*Bool) T() string { return "bool" }

// Array represents an array type with element type and size expression.
type Array struct {
	Type Type
	Size Expression
}

func (a *Array) T() string { return fmt.Sprintf("array of %s", a.Type.T()) }

// Void represents absence of a value
type Void token.Token

func (*Void) T() string { return "void" }
