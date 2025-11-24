package object

import "fmt"

type Object interface {
	Type() Type
	GetValue() any
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

func (io *IntObject) GetValue() any {
	return io.Value
}

type FloatObject struct {
	Value float64
}

func (*FloatObject) Type() Type {
	return Float
}

func (fo *FloatObject) GetValue() any {
	return fo.Value
}

type BoolObject struct {
	Value bool
}

func (*BoolObject) Type() Type {
	return Bool
}

func (bo *BoolObject) GetValue() any {
	return bo.Value
}

type StringObject struct {
	Value string
}

func (*StringObject) Type() Type {
	return String
}

func (so *StringObject) GetValue() any {
	return so.Value
}

type VoidObject struct {
	Value any
}

func (vo *VoidObject) GetValue() any {
	return vo.Value
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

func (ao *ArrayObject) GetValue() any {
	return ao.Value
}
