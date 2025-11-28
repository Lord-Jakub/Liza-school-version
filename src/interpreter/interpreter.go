package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
)

var (
	Namespaces map[string]*Environment                         = make(map[string]*Environment)
	BuildIns   map[string]func(*Environment, []ast.Expression) = make(map[string]func(*Environment, []ast.Expression))
)

func Init() {
	BuildIns["print"] = func(env *Environment, args []ast.Expression) {
		for _, arg := range args {
			argVal, err := Eval(arg, env)
			if err != nil {
				panic(err) // TODO: also handle err
			}
			fmt.Print(argVal.GetValue())
		}
	}
	BuildIns["len"] = func(env *Environment, args []ast.Expression) {
		if len(args) > 1 {
			panic(fmt.Errorf("Too many argumetns"))
		} else if len(args) < 1 {
			panic(fmt.Errorf("Not enough argumetns"))
		}
		value, err := Eval(args[0], env)
		if err != nil {
			panic(err)
		}
		arr, ok := value.(*object.ArrayObject)
		if !ok {
			panic(fmt.Errorf("argument must be array"))
		}
		var obj object.Object
		obj = &object.IntObject{int64(arr.Len)}
		env.Return = &obj
	}
}

func GetNamespaces(program *ast.Program) {
	for namespaceName, namespace := range program.Namespaces {
		env := NewEnv()
		for _, node := range namespace.Nodes {
			switch node.(type) {
			case (ast.FunctionDeclarationStatement):
				f := node.(ast.FunctionDeclarationStatement)
				env.DeclareFunc(f.Name.Value.(string), &f)
				break
			case (ast.VariableDeclarationStatement):
				v := node.(ast.VariableDeclarationStatement)

				env.DeclareVar(v)
				break
			}
			Namespaces[namespaceName] = &env
		}
	}
}

func Interpret(body *ast.BodyStatement, env *Environment) error {
	for _, node := range body.Nodes {
		switch node.(type) {
		case (ast.FunctionDeclarationStatement):
			f := node.(ast.FunctionDeclarationStatement)
			env.DeclareFunc(f.Name.Value.(string), &f)
			break
		case (ast.VariableDeclarationStatement):
			v := node.(ast.VariableDeclarationStatement)
			env.DeclareVar(v)
			break
		case (ast.FunctionCall):
			functionCall := node.(ast.FunctionCall)
			_, err := env.CallFunction(&functionCall)
			if err != nil {
				return err
			}
			break
		case (ast.ReturnStatement):
			retVal, err := Eval(node.(ast.ReturnStatement).ReturnValue, env)
			env.Return = &retVal
			return err
		case (ast.IfStatement):
			ifStmt := node.(ast.IfStatement)
			condition, err := Eval(ifStmt.Condition, env)
			if err != nil {
				return err
			}
			if condition.Type() != object.Bool {
				return fmt.Errorf("non-bool condition is not allowed")
			}
			if condition.GetValue().(bool) {
				scope := NewEnv()
				scope.Outer = env
				scope.Namespace = env.Namespace
				Interpret(&ifStmt.Body, &scope)
				if scope.Return != nil {
					env.Return = scope.Return
					return nil
				}
			} else {
				scope := NewEnv()
				scope.Outer = env
				scope.Namespace = env.Namespace
				Interpret(&ifStmt.Alternative, &scope)
				if scope.Return != nil {
					env.Return = scope.Return
					return nil
				}
			}
			break
		case (ast.VariableAssignmentStatement):
			varStmt := node.(ast.VariableAssignmentStatement)
			name, err := GetVariableName(varStmt.Target)
			if err != nil {
				return err
			}
			target, ok := env.GetVar(name)

			var indexes []int64
			arr, isBinExpr := varStmt.Target.(*ast.BinaryExpression)
			isArray := isBinExpr
			for isBinExpr {
				//arr, _ := varStmt.Target.(*ast.BinaryExpression)
				indexObj, err := Eval(arr.Right, env)
				if err != nil {
					return err
				}
				indexes = append(indexes, indexObj.GetValue().(int64))
				arr, isBinExpr = arr.Left.(*ast.BinaryExpression)
			}
			if !ok {
				return fmt.Errorf("variable %s is not declared", name)
			}
			if !target.Mutable {
				return fmt.Errorf("variable %s isn't mutable", name)
			}
			value, err := Eval(varStmt.Value, env)
			if err != nil {
				return err
			}

			if value.Type() != target.Type && !isArray {
				return fmt.Errorf("cannot assign value of type %s to variable of type %s", value.Type(), target.Type)
			}
			if !isArray {
				target.Value = value
			} else {
				for i, j := 0, len(indexes)-1; i < j; i, j = i+1, j-1 {
					indexes[i], indexes[j] = indexes[j], indexes[i]
				}
				setValueInArray(value, target.Value.(*object.ArrayObject), indexes)
			}
			break
		case (ast.ForStatement):
			forStmt := node.(ast.ForStatement)

			scope := NewEnv()
			scope.Outer = env
			scope.Namespace = env.Namespace
			scope.DeclareVar(forStmt.Init)
			conditionVal, err := Eval(forStmt.Condition, &scope)
			if err != nil {
				return err
			}

			condition, ok := conditionVal.GetValue().(bool)
			if !ok {
				return fmt.Errorf("non-bool condition is not allowed")
			}

			forStmt.Body.Nodes = append(forStmt.Body.Nodes, forStmt.Post)

			for condition {
				Interpret(&forStmt.Body, &scope)
				if scope.Return != nil {
					env.Return = scope.Return
					return nil
				}
				conditionVal, err = Eval(forStmt.Condition, &scope)
				if err != nil {
					return err
				}
				condition = conditionVal.GetValue().(bool)
			}
			break

		}
	}
	return nil
}

/*
	func GetVarFromExpression(expression ast.Expression, env *Environment) (*Variable, bool){
		if arrIdx, ok := expression.(*ast.BinaryExpression); ok{
					var array *Variable
					array, ok = GetVarFromExpression(arrIdx.Left, env)
					index, err:= Eval(arrIdx.Right, env)
					panic(err)
					valAtIndex := array.Value.(*object.ArrayObject).Value[index.GetValue().(int64)]
					return
				} else {

				}
	}
*/
func GetVariableName(expression ast.Expression) (string, error) {
	if binary, ok := expression.(*ast.BinaryExpression); ok {
		return GetVariableName(binary.Left)
	} else if variable, ok := expression.(*ast.VariableExpression); ok {
		return variable.Value.Value.(string), nil
	} else {
		return "", fmt.Errorf("not identifier")
	}
}
func setValueInArray(value object.Object, array *object.ArrayObject, indexes []int64) {
	if array.Value[int(indexes[0])] == nil || array.Value[int(indexes[0])].Type() == value.Type() {
		array.Value[int(indexes[0])] = value
	} else {
		if len(indexes) > 0 {
			if arr, ok := array.Value[int(indexes[0])].(*object.ArrayObject); ok {
				indexes = indexes[1:]
				setValueInArray(value, arr, indexes)
			} else {
				panic("")
			}
		} else {
			panic("")
		}
	}
}
