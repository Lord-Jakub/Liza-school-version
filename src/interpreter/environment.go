package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
)

// Environment represents a scope
// It stores variables, functions, and links to an outer scope.
type Environment struct {
	StoreVars  map[string]*Variable
	StoreFuncs map[string]*ast.FunctionDeclarationStatement
	Outer      *Environment
	Return     *object.Object
	Namespace  string
}

// NewEnv creates a new environment with built-in constants for true/false
func NewEnv() Environment {
	env := Environment{
		make(map[string]*Variable),
		make(map[string]*ast.FunctionDeclarationStatement),
		nil,
		nil,
		"",
	}

	// Define immutable boolean constants
	trueObj := object.BoolObject{true}
	trueConst := Variable{"true", object.Bool, false, &trueObj}
	env.StoreVars["true"] = &trueConst

	falseObj := object.BoolObject{false}
	falseConst := Variable{"false", object.Bool, false, &falseObj}
	env.StoreVars["false"] = &falseConst

	return env
}

// Variable represents a runtime variable.
type Variable struct {
	Name    string
	Type    object.Type
	Mutable bool
	Value   object.Object
}

// GetVar resolves a variable by name, searching outer scopes if necessary.
func (env *Environment) GetVar(name string) (*Variable, bool) {
	variable, ok := env.StoreVars[name]
	if !ok && env.Outer != nil {
		variable, ok = env.Outer.GetVar(name)
	}
	return variable, ok
}

// DeclareVar declares a new variable in the current scope.
// It evaluates the initial value (if present) and enforces type correctness.
func (env *Environment) DeclareVar(v ast.VariableDeclarationStatement) (*Variable, error) {
	var vValue object.Object

	// Evaluate initial value if provided
	if v.Value != nil {
		var err error
		vValue, err = Eval(v.Value, env)
		if err != nil {
			return &Variable{}, err
		}
	}

	// Type check against declared type
	if vValue != nil && vValue.Type() != object.Type(v.Type.T()) {
		return &Variable{}, fmt.Errorf("Can't assign value of type %s to variable of type %s", vValue.Type(), v.Type.T())
	}

	// Handle uninitialized arrays
	if vValue == nil {
		if _, ok := v.Type.(*ast.Array); ok {
			lenght, err := Eval(v.Type.(*ast.Array).Size, env)

			// Array size must be integer
			if lenght.GetValue() != nil && lenght.Type() != object.Int {
				return nil, fmt.Errorf("array len must be integer")
			}

			lenghtInt := 0
			if lenght.GetValue() != nil {
				lenghtInt = int(lenght.GetValue().(int64))
			}
			if err != nil {
				return nil, err
			}

			// Allocate array with given length
			vValue = &object.ArrayObject{
				Len:         lenghtInt,
				ElementType: object.Type(v.Type.T()),
				Value:       make([]object.Object, lenghtInt),
			}
		}
	}

	variable := &Variable{
		Name:    v.Identifier.Value.(string),
		Type:    object.Type(v.Type.T()),
		Mutable: v.Mutable,
		Value:   vValue,
	}

	// Prevent redeclaration
	_, ok := env.StoreVars[variable.Name]
	if !ok && env.Outer != nil {
		_, ok = env.Outer.GetVar(variable.Name)
	}
	if ok {
		return nil, fmt.Errorf("cannot redeclare variable %s", variable.Name)
	}

	env.StoreVars[variable.Name] = variable
	return variable, nil
}

// GetFunc find a function by name
func (env *Environment) GetFunc(name string) (*ast.FunctionDeclarationStatement, bool) {
	function, ok := env.StoreFuncs[name]
	if !ok && env.Outer != nil {
		function, ok = env.Outer.GetFunc(name)
	}
	return function, ok
}

// DeclareFunc registers a function in the current scope
func (env *Environment) DeclareFunc(name string, function *ast.FunctionDeclarationStatement) error {
	_, ok := env.StoreFuncs[name]
	if !ok && env.Outer != nil {
		_, ok = env.Outer.GetFunc(name)
	}
	if ok {
		return fmt.Errorf("cannot redeclare function %s", name)
	}

	env.StoreFuncs[name] = function
	return nil
}

// CallFunction executes a function call.
func (env *Environment) CallFunction(functionCall *ast.FunctionCall) (*Environment, error) {

	// Built-in functions
	if function, ok := BuildIns[functionCall.Identifier.Value.(string)]; ok {
		err := function(env, functionCall.Args)
		return env, err
	}

	// User-defined function
	function, ok := env.GetFunc(functionCall.Identifier.Value.(string))
	if !ok {
		return env, fmt.Errorf("function %s is not declared", functionCall.Identifier.Value.(string))
	}

	// Argument count check
	if len(function.Args) != len(functionCall.Args) {
		return env, fmt.Errorf(
			"function %s need %d arguments, provided %d",
			functionCall.Identifier.Value.(string),
			len(function.Args),
			len(functionCall.Args),
		)
	}

	// Create new function scope
	funcEnv := NewEnv()
	funcEnv.Namespace = env.Namespace

	// Functions are resolved from namespace-level environment
	funcEnv.Outer = Namespaces[funcEnv.Namespace]

	// Pass arguments
	for i, arg := range functionCall.Args {
		variable, _ := funcEnv.DeclareVar(function.Args[i])
		var err error
		variable.Value, err = Eval(arg, env)
		if err != nil {
			return env, err
		}
	}

	// Execute function body
	err := Interpret(function.Body, &funcEnv)

	// If no return, just return environment
	if funcEnv.Return == nil {
		return &funcEnv, err
	}

	retVal := *funcEnv.Return

	// Enforce return type correctness
	if string(retVal.Type()) != function.Type.T() {
		return env, fmt.Errorf(
			"cannot return type %s in function of type %s",
			retVal.Type(),
			function.Type.T(),
		)
	}

	return &funcEnv, err
}
