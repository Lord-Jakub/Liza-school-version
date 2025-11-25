package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
)

type Environment struct {
	StoreVars  map[string]*Variable
	StoreFuncs map[string]*ast.FunctionDeclarationStatement
	Outer      *Environment
	Return     *object.Object
}

func NewEnv() Environment {
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

func (env *Environment) DeclareVar(v ast.VariableDeclarationStatement) (*Variable, error) {
	var vValue object.Object
	if v.Value != nil {
		var err error
		vValue, err = Eval(v.Value, env)
		if err != nil {
			// TODO: handle errs
			panic(err)
		}
	}
	if vValue != nil && vValue.Type() != object.Type(v.Type.T()) {
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
		return nil, fmt.Errorf("cannot redeclare variable %s", variable.Name)
	}
	env.StoreVars[variable.Name] = variable
	return variable, nil
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

func (env *Environment) CallFunction(functionCall *ast.FunctionCall) (*Environment, error) {
	if function, ok := BuildIns[functionCall.Identifier.Value.(string)]; ok {
		function(env, functionCall.Args...)
		return env, nil
	}
	function, ok := env.GetFunc(functionCall.Identifier.Value.(string))
	if !ok {
		return env, fmt.Errorf("function %s is not declared", functionCall.Identifier.Value.(string))
	}
	if len(function.Args) != len(functionCall.Args) {
		return env, fmt.Errorf("function %s need %d arguments, provided %d", functionCall.Identifier.Value.(string), len(function.Args), len(functionCall.Args))
	}

	funcEnv := NewEnv()
	for i, arg := range functionCall.Args {
		variable, _ := funcEnv.DeclareVar(function.Args[i])
		var err error
		variable.Value, err = Eval(arg, env)
		if err != nil {
			return env, err
		}
	}
	funcEnv.Outer = env
	err := Interpret(&function.Body, &funcEnv)
	retVal := *funcEnv.Return
	if string(retVal.Type()) != function.Type.T() {
		return env, fmt.Errorf("cannot return type %s in function of type %s", retVal.Type(), function.Type.T())
	}
	return &funcEnv, err
}
