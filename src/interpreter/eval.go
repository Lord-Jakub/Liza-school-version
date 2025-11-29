package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
	"math"
)

func Eval(expression ast.Expression, env *Environment) (object.Object, error) {
	var obj object.Object
	switch expression.(type) {
	case (*ast.LiteralExpression):
		literal := expression.(*ast.LiteralExpression)
		return EvalLiteral(literal), nil
	case (*ast.BinaryExpression):
		binary := expression.(*ast.BinaryExpression)
		return EvalBinary(binary, env)
	case (*ast.UnaryExpression):
		unary := expression.(*ast.UnaryExpression)
		return EvalUnary(unary, env)
	case (*ast.VariableExpression):
		variableEx := expression.(*ast.VariableExpression)
		variable, ok := env.GetVar(variableEx.String())
		if !ok {
			return &object.VoidObject{}, fmt.Errorf("variable %s does not exist", variableEx.String())
		}
		return variable.Value, nil
	case (*ast.ArrayExpression):
		array := expression.(*ast.ArrayExpression)
		return EvalArray(array, env)
	case (*ast.FunctionCall):
		functionCall := expression.(*ast.FunctionCall)
		retEnv, err := env.CallFunction(functionCall)
		if retEnv.Return == nil {
			return nil, err
		}
		returnVal := *retEnv.Return
		retEnv.Return = nil
		return returnVal, err
	default:
		return &object.VoidObject{}, nil

	}
	return obj, nil
}

func EvalArray(array *ast.ArrayExpression, env *Environment) (object.Object, error) {
	var elements []object.Object
	var elementType object.Type
	for i, element := range array.Elements {
		elementObj, err := Eval(element, env)
		if err != nil {
			return &object.VoidObject{}, err
		}
		if i == 0 {
			elementType = elementObj.Type()
		}
		if elementObj.Type() != elementType {
			err = fmt.Errorf("cannot use element of type %s in array of type %s", elementObj.Type(), elementType)
			if err != nil {
				return &object.VoidObject{}, err
			}
		}
		elements = append(elements, elementObj)
	}
	return &object.ArrayObject{len(elements), elementType, elements}, nil
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

func EvalUnary(unary *ast.UnaryExpression, env *Environment) (object.Object, error) {
	switch unary.Prefix.Value {
	case "-":
		value, err := Eval(unary.Value, env)
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
		return Eval(unary.Value, env)
	case "!":
		value, err := Eval(unary.Value, env)
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

func EvalBinary(binary *ast.BinaryExpression, env *Environment) (object.Object, error) {
	if binary.Op.Value.(string) == "." {
		switch binary.Left.(type) {
		case *ast.VariableExpression:
			if namespace, ok := Namespaces[binary.Left.(*ast.VariableExpression).Value.Value.(string)]; ok {
				return Eval(binary.Right, namespace)
			}
			break
		}
	}
	left, err := Eval(binary.Left, env)
	if err != nil {
		return &object.VoidObject{}, err
	}
	op := binary.Op.Value.(string)
	right, err := Eval(binary.Right, env)
	if err != nil {
		return &object.VoidObject{}, err
	}

	if op == "[" {
		array, ok := left.(*object.ArrayObject)
		if !ok {
			return &object.VoidObject{}, fmt.Errorf("expected array, got %s", left.Type())
		}
		index, ok := right.(*object.IntObject)
		if !ok {
			return &object.VoidObject{}, fmt.Errorf("expected int, got %s", left.Type())
		}
		if index.Value >= int64(array.Len) {
			return &object.VoidObject{}, fmt.Errorf("Index %d out of bounds (array lenght is %d)", index.Value, array.Len)
		}
		return array.Value[index.Value], nil
	}
	if right.Type() != left.Type() {
		return &object.VoidObject{}, fmt.Errorf("cannot use %s on types %s and %s", op, right.Type(), left.Type())
	}
	if op == "==" {
		return &object.BoolObject{left.GetValue() == right.GetValue()}, nil
	}
	if op == "!=" {
		return &object.BoolObject{left.GetValue() != right.GetValue()}, nil
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
		case "<":
			return &object.BoolObject{leftInt.Value < rightInt.Value}, nil
		case ">":
			return &object.BoolObject{leftInt.Value > rightInt.Value}, nil
		case "<=":
			return &object.BoolObject{leftInt.Value <= rightInt.Value}, nil
		case ">=":
			return &object.BoolObject{leftInt.Value >= rightInt.Value}, nil

		}
	case object.Bool:
		rightBool := right.(*object.BoolObject)
		leftBool := left.(*object.BoolObject)
		switch op {
		case "&&":
			// fmt.Printf("%d %s %d = %d\n", leftBool.Value, op, righBool.Value, leftBool.Value && rightBool.Value)
			return &object.BoolObject{leftBool.Value && rightBool.Value}, nil
		case "||":
			// fmt.Printf("%d %s %d = %d\n", leftBool.Value, op, righBool.Value, leftBool.Value || rightBool.Value)
			return &object.BoolObject{leftBool.Value || rightBool.Value}, nil
		}
	case object.Float:
		rightFloat := right.(*object.FloatObject)
		leftFloat := left.(*object.FloatObject)
		switch op {
		case "+":
			// fmt.Printf("%d %s %d = %d\n", leftFloat.Value, op, rightFloat.Value, leftFloat.Value+rightFloat.Value)
			return &object.FloatObject{leftFloat.Value + rightFloat.Value}, nil
		case "-":
			// fmt.Printf("%d %s %d = %d\n", leftFloat.Value, op, rightFloat.Value, leftFloat.Value-rightFloat.Value)
			return &object.FloatObject{leftFloat.Value - rightFloat.Value}, nil
		case "*":
			// fmt.Printf("%d %s %d = %d\n", leftFloat.Value, op, rightFloat.Value, leftFloat.Value*rightFloat.Value)
			return &object.FloatObject{leftFloat.Value * rightFloat.Value}, nil
		case "/":
			// fmt.Printf("%d %s %d = %d\n", leftFloat.Value, op, rightFloat.Value, leftFloat.Value/rightFloat.Value)
			return &object.FloatObject{leftFloat.Value / rightFloat.Value}, nil
		case "^":
			// fmt.Printf("%d %s %d = %d\n", leftFloat.Value, op, rightFloat.Value, math.Pow(leftFloat.Value, rightFloat.Value))
			return &object.FloatObject{math.Pow(leftFloat.Value, rightFloat.Value)}, nil
		case "<":
			return &object.BoolObject{leftFloat.Value < rightFloat.Value}, nil
		case ">":
			return &object.BoolObject{leftFloat.Value > rightFloat.Value}, nil
		case "<=":
			return &object.BoolObject{leftFloat.Value <= rightFloat.Value}, nil
		case ">=":
			return &object.BoolObject{leftFloat.Value >= rightFloat.Value}, nil
		}

	case object.String:
		rightString := right.(*object.StringObject)
		leftString := left.(*object.StringObject)
		switch op {
		case "+":
			// fmt.Printf("%d %s %d = %d\n", leftString.Value, op, rightString.Value, leftString.Value+rightString.Value)
			return &object.StringObject{leftString.Value + rightString.Value}, nil
		}

	}
	return nil, nil
}
