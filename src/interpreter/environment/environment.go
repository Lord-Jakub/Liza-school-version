package environment

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/eval"
	"lizalang/interpreter/object"
)

type Environment struct {
	StoreVars  map[string]*Variable
	StoreFuncs map[string]*ast.FunctionDeclarationStatement
	Outer      *Environment
	Return     *object.Object
}

func New() Environment {
	return Environment{
		make(map[string]*Variable),
		make(map[string]*ast.FunctionDeclarationStatement),
		nil,
		nil,
	}
}

type Variable struct {
	Name    string
	Type    object.Type
	Mutable bool
	Value   object.Object
}

func (env *Environment) GetVar(name string) (*Variable, bool) {
	variable, ok := env.StoreVars[name]
	if !ok {
		if env.Outer != nil {
			variable, ok = env.Outer.GetVar(name)
		}
	}
	return variable, ok
}

func (env *Environment) DeclareVar(v ast.VariableDeclarationStatement) error {
	var vValue object.Object
	if v.Value != nil {
		var err error
		vValue, err = eval.Eval(v.Value)
		if err != nil {
			// TODO: handle errs
			panic(err)
		}
	}
	if vValue.Type() != object.Type(v.Type.T()) {
		// TODO: handle err
		panic("wrong type bro")
	}
	variable := &Variable{
		Name:    v.Identifier.Value.(string),
		Type:    object.Type(v.Type.T()),
		Mutable: v.Mutable,
		Value:   vValue,
	}

	_, ok := env.StoreVars[variable.Name]
	if !ok {
		if env.Outer != nil {
			_, ok = env.Outer.GetVar(variable.Name)
		}
	}
	if ok {
		return fmt.Errorf("cannot redeclare variable %s", variable.Name)
	}
	env.StoreVars[variable.Name] = variable
	return nil
}

func (env *Environment) GetFunc(name string) (*ast.FunctionDeclarationStatement, bool) {
	function, ok := env.StoreFuncs[name]
	if !ok {
		if env.Outer != nil {
			function, ok = env.Outer.GetFunc(name)
		}
	}
	return function, ok
}

func (env *Environment) DeclareFunc(name string, function *ast.FunctionDeclarationStatement) error {
	_, ok := env.StoreFuncs[name]
	if !ok {
		if env.Outer != nil {
			_, ok = env.Outer.GetFunc(name)
		}
	}
	if ok {
		return fmt.Errorf("cannot redeclare function %s", name)
	}
	env.StoreFuncs[name] = function
	return nil
}
