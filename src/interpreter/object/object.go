package object

import "fmt"

type Object interface {
	Type() Type
	GetValue() any
	Copy() Object
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

func (*IntObject) Type() Type { return Int }

func (io *IntObject) GetValue() any { return io.Value }

func (io *IntObject) Copy() Object {
	if io == nil {
		return &IntObject{Value: 0}
	}
	return &IntObject{Value: io.Value}
}

type FloatObject struct {
	Value float64
}

func (*FloatObject) Type() Type { return Float }

func (fo *FloatObject) GetValue() any { return fo.Value }

func (fo *FloatObject) Copy() Object {
	if fo == nil {
		return &FloatObject{Value: 0.0}
	}
	return &FloatObject{Value: fo.Value}
}

type BoolObject struct {
	Value bool
}

func (*BoolObject) Type() Type { return Bool }

func (bo *BoolObject) GetValue() any { return bo.Value }

func (bo *BoolObject) Copy() Object {
	if bo == nil {
		return &BoolObject{Value: false}
	}
	return &BoolObject{Value: bo.Value}
}

type StringObject struct {
	Value string
}

func (*StringObject) Type() Type { return String }

func (so *StringObject) GetValue() any { return so.Value }

func (so *StringObject) Copy() Object {
	if so == nil {
		return &StringObject{Value: ""}
	}
	return &StringObject{Value: so.Value}
}

type VoidObject struct {
	Value any
}

func (*VoidObject) Type() Type { return Void }

func (vo *VoidObject) GetValue() any {
	if vo == nil {
		return nil
	}
	return vo.Value
}

func (vo *VoidObject) Copy() Object {
	if vo == nil {
		return &VoidObject{}
	}
	return &VoidObject{Value: vo.Value}
}

type ArrayObject struct {
	Len         int
	ElementType Type
	Value       []Object
}

func (ao *ArrayObject) Type() Type {
	if ao == nil {
		return "array"
	}
	return Type(fmt.Sprintf("array of %s", ao.ElementType))
}

func (ao *ArrayObject) GetValue() any {
	if ao == nil {
		return nil
	}
	return ao.Value
}

func (ao *ArrayObject) Copy() Object {
	if ao == nil {
		return &ArrayObject{
			Len:         0,
			ElementType: Void,
			Value:       nil,
		}
	}

	newElements := make([]Object, len(ao.Value))
	for i, val := range ao.Value {
		if val == nil {
			newElements[i] = nil
		} else {
			newElements[i] = val.Copy()
		}
	}

	return &ArrayObject{
		Len:         ao.Len,
		ElementType: ao.ElementType,
		Value:       newElements,
	}
}
