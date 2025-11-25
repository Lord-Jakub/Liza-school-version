package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
)

var (
	Namespaces map[string]*Environment                          = make(map[string]*Environment)
	BuildIns   map[string]func(*Environment, ...ast.Expression) = make(map[string]func(*Environment, ...ast.Expression))
)

func Init() {
	BuildIns["print"] = func(env *Environment, args ...ast.Expression) {
		for _, arg := range args {
			argVal, err := Eval(arg, env)
			if err != nil {
				panic(err) // TODO: also handle err
			}
			fmt.Print(argVal.GetValue())
		}
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
				Interpret(&ifStmt.Body, &scope)
				if scope.Return != nil {
					env.Return = scope.Return
					return nil
				}
			} else {
				scope := NewEnv()
				scope.Outer = env
				Interpret(&ifStmt.Alternative, &scope)
				if scope.Return != nil {
					env.Return = scope.Return
					return nil
				}
			}
			break
		case (ast.VariableAssignmentStatement):
			varStmt := node.(ast.VariableAssignmentStatement)
			target, ok := env.GetVar(varStmt.Target.Value.(string))
			if !ok {
				return fmt.Errorf("variable %s is not declared", varStmt.Target.Value.(string))
			}
			if !target.Mutable {
				return fmt.Errorf("variable %s isn't mutable", varStmt.Target.Value.(string))
			}
			value, err := Eval(varStmt.Value, env)
			if err != nil {
				return err
			}
			if value.Type() != target.Type {
				return fmt.Errorf("cannot assign value of type %s to variable of type %s", value.Type(), target.Type)
			}
			target.Value = value
			break
		}
	}
	return nil
}
