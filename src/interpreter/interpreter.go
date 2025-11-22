package interpreter

import (
	"fmt"

	"lizalang/ast"
	"lizalang/object"
)

func Eval(expression ast.Expression) (object.Object, error) {
	var obj object.Object
	switch expression.(type) {
	case (*ast.LiteralExpression):
		literal := expression.(*ast.LiteralExpression)
		return EvalLiteral(literal), nil
	case (*ast.BinaryExpression):
		binary := expression.(*ast.BinaryExpression)
		return EvalBinary(binary)
	case (*ast.UnaryExpression):
		unary := expression.(*ast.UnaryExpression)
		return EvalUnary(unary)
	case (*ast.VariableExpression):
		break
	case (*ast.ArrayExpression):
		break
	default:
		return &object.VoidObject{}, nil

	}
	return obj, nil
}

func EvalLiteral(literal *ast.LiteralExpression) object.Object {
	switch literal.Value.Type {
	case "INT":
		return &object.IntObject{literal.Value.Value.(int64)}
	case "FLOAT":
		return &object.FloatObject{literal.Value.Value.(float64)}
	case "BOOL":
		return &object.BoolObject{literal.Value.Value.(bool)}
	case "STRING":
		return &object.StringObject{literal.Value.Value.(string)}
	default:
		return &object.VoidObject{}
	}
}

func EvalUnary(unary *ast.UnaryExpression) (object.Object, error) {
	switch unary.Prefix.Value {
	case "-":
		value, err := Eval(unary.Value)
		if err != nil {
			return &object.VoidObject{}, err
		}
		if value.Type() == object.Int {
			value.(*object.IntObject).Value = -value.(*object.IntObject).Value
			return value, nil
		} else if value.Type() == object.Float {
			value.(*object.FloatObject).Value = -value.(*object.FloatObject).Value
			return value, nil
		} else {
			return value, fmt.Errorf("cannot use prefix - on %s", value.Type())
		}
	case "+":
		return Eval(unary.Value)
	case "!":
		value, err := Eval(unary.Value)
		if err != nil {
			return &object.VoidObject{}, err
		}
		if value.Type() == object.Bool {
			value.(*object.BoolObject).Value = !value.(*object.BoolObject).Value
			return value, nil
		} else {
			return value, fmt.Errorf("cannot use prefix - on %s", value.Type())
		}

	default:
		return &object.VoidObject{}, fmt.Errorf("unexpected +, -, or !, got %s", unary.Prefix.Value.(string))
	}
}

func EvalBinary(binary *ast.BinaryExpression) (object.Object, error) {
	left, err := Eval(binary.Left)
	if err != nil {
		return &object.VoidObject{}, err
	}
	op := binary.Op.Value.(string)
	right, err := Eval(binary.Right)
	if err != nil {
		return &object.VoidObject{}, err
	}
	if right.Type() != left.Type() {
		return &object.VoidObject{}, fmt.Errorf("cannot use %s on types %s and %s", op, right.Type(), left.Type())
	}
	switch right.Type() {
	case object.Int:
		rightInt := right.(*object.IntObject)
		leftInt := left.(*object.IntObject)
		switch op {
		case "+":
			// fmt.Printf("%d %s %d = %d\n", leftInt.Value, op, rightInt.Value, leftInt.Value+rightInt.Value)
			return &object.IntObject{leftInt.Value + rightInt.Value}, nil
		case "-":
			// fmt.Printf("%d %s %d = %d\n", leftInt.Value, op, rightInt.Value, leftInt.Value-rightInt.Value)
			return &object.IntObject{leftInt.Value - rightInt.Value}, nil
		case "*":
			// fmt.Printf("%d %s %d = %d\n", leftInt.Value, op, rightInt.Value, leftInt.Value*rightInt.Value)
			return &object.IntObject{leftInt.Value * rightInt.Value}, nil
		case "/":
			// fmt.Printf("%d %s %d = %d\n", leftInt.Value, op, rightInt.Value, leftInt.Value/rightInt.Value)
			return &object.IntObject{leftInt.Value / rightInt.Value}, nil
		case "^":
			pow := leftInt.Value
			if rightInt.Value == 0 {
				pow = 1
			} else {
				for i := 1; i < int(rightInt.Value); i++ {
					pow *= leftInt.Value
				}
			}

			// fmt.Printf("%d %s %d = %d\n", leftInt.Value, op, rightInt.Value, pow)
			return &object.IntObject{pow}, nil
		}
	case object.Bool:
	case object.Float:
	case object.String:
	}
	return nil, nil
}
