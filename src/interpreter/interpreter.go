package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/environment"
	"lizalang/interpreter/eval"
)

var (
	Namespaces map[string]*environment.Environment = make(map[string]*environment.Environment)
	BuildIns   map[string]func(...ast.Expression)  = make(map[string]func(...ast.Expression))
)

func Init() {
	BuildIns["print"] = func(args ...ast.Expression) {
		for _, arg := range args {
			argVal, err := eval.Eval(arg)
			if err != nil {
				panic(err) // TODO: also handle err
			}
			fmt.Print(argVal.GetValue())
		}
	}
}

func GetNamespaces(program *ast.Program) {
	for namespaceName, namespace := range program.Namespaces {
		env := environment.New()
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

func Interpret(body *ast.BodyStatement, env *environment.Environment) error {
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
			if function, ok := BuildIns[functionCall.Identifier.Value.(string)]; ok {
				function(functionCall.Args...)
				break
			}
			function, ok := env.GetFunc(functionCall.Identifier.Value.(string))
			if !ok {
				return fmt.Errorf("function %s is not declared", functionCall.Identifier.Value.(string))
			}
			if len(function.Args) != len(functionCall.Args) {
				return fmt.Errorf("function %s need %d arguments, provided %d", functionCall.Identifier.Value.(string), len(function.Args), len(functionCall.Args))
			}
			for i := range function.Args {
				function.Args[i].Value = functionCall.Args[i]
			}
			funcEnv := environment.New()
			funcEnv.Outer = env
			Interpret(&function.Body, &funcEnv)
			break
		}
	}
	return nil
}
