package object

import "fmt"

type Object interface {
	Type() Type
}

type Type string

const (
	Int    = "int"
	Float  = "float"
	Bool   = "bool"
	String = "string"
	Void   = "void"
)

type IntObject struct {
	Value int64
}

func (*IntObject) Type() Type {
	return Int
}

type FloatObject struct {
	Value float64
}

func (*FloatObject) Type() Type {
	return Float
}

type BoolObject struct {
	Value bool
}

func (*BoolObject) Type() Type {
	return Bool
}

type StringObject struct {
	Value string
}

func (*StringObject) Type() Type {
	return String
}

type VoidObject struct {
	Value any
}

func (*VoidObject) Type() Type {
	return Void
}

type ArrayObject struct {
	Len         int
	ElementType Type
	Value       []Object
}

func (ao *ArrayObject) Type() Type {
	return Type(fmt.Sprintf("array of %s", string(ao.ElementType)))
}
