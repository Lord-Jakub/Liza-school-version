package interpreter

import (
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
)

var (
	// Namespaces holds global environments for each namespace.
	Namespaces map[string]*Environment = make(map[string]*Environment)

	// BuildIns maps builtin function names to their implementations.
	BuildIns map[string]func(*Environment, []ast.Expression) = make(map[string]func(*Environment, []ast.Expression))
)

// Init registers all builtin functions into the interpreter.
func Init() {
	// print(...) - prints evaluated arguments without newline
	BuildIns["print"] = func(env *Environment, args []ast.Expression) {
		for _, arg := range args {
			argVal, err := Eval(arg, env)
			if err != nil {
				panic(err) // TODO: replace with proper error propagation
			}
			fmt.Print(argVal.GetValue())
		}
	}

	// len(array) - returns length of an array
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

	// IntToFloat(int) - converts int to float
	BuildIns["IntToFloat"] = func(env *Environment, args []ast.Expression) {
		if len(args) != 1 {
			panic(fmt.Errorf("You can have only one argument in IntToFloat()"))
		}

		argVal, err := Eval(args[0], env)
		if err != nil {
			panic(err)
		}

		if argVal.Type() != object.Int {
			panic(fmt.Errorf("IntToFloat() accept only int as an argument, not %s", argVal.Type()))
		}

		var obj object.Object
		obj = &object.FloatObject{float64(argVal.GetValue().(int64))}
		env.Return = &obj
	}

	// FloatToInt(float) - converts float to int
	BuildIns["FloatToInt"] = func(env *Environment, args []ast.Expression) {
		if len(args) != 1 {
			panic(fmt.Errorf("You can have only one argument in FloatToInt()"))
		}

		argVal, err := Eval(args[0], env)
		if err != nil {
			panic(err)
		}

		if argVal.Type() != object.Float {
			panic(fmt.Errorf("FloatToInt() accept only float as an argument, not %s", argVal.Type()))
		}

		var obj object.Object
		obj = &object.IntObject{int64(argVal.GetValue().(float64))}
		env.Return = &obj
	}
}

// GetNamespaces initializes environments for each namespace in the program.
// It pre-registers functions and global variables.
func GetNamespaces(program *ast.Program) {
	for namespaceName, namespace := range program.Namespaces {
		env := NewEnv()
		env.Namespace = namespaceName

		for _, node := range namespace.Nodes {
			switch node.(type) {
			case ast.FunctionDeclarationStatement:
				f := node.(ast.FunctionDeclarationStatement)
				env.DeclareFunc(f.Name.Value.(string), &f)

			case ast.VariableDeclarationStatement:
				v := node.(ast.VariableDeclarationStatement)
				env.DeclareVar(v)
			}

			// Store environment per namespace
			Namespaces[namespaceName] = &env
		}
	}
}

// Interpret executes a single AST node within a given environment.
func Interpret(node ast.Node, env *Environment) error {
	switch node.(type) {

	case ast.FunctionDeclarationStatement:
		// Register function in current scope
		f := node.(ast.FunctionDeclarationStatement)
		env.DeclareFunc(f.Name.Value.(string), &f)

	case ast.VariableDeclarationStatement:
		// Declare variable in current scope
		v := node.(ast.VariableDeclarationStatement)
		env.DeclareVar(v)

	case ast.FunctionCall:
		// Execute function call
		functionCall := node.(ast.FunctionCall)
		_, err := env.CallFunction(&functionCall)
		if err != nil {
			return err
		}

	case ast.ReturnStatement:
		// Evaluate return value
		retVal, err := Eval(node.(ast.ReturnStatement).ReturnValue, env)
		env.Return = &retVal
		return err

	case ast.IfStatement:
		// Evaluate condition
		ifStmt := node.(ast.IfStatement)
		condition, err := Eval(ifStmt.Condition, env)
		if err != nil {
			return err
		}

		if condition.Type() != object.Bool {
			return fmt.Errorf("non-bool condition is not allowed")
		}

		// Execute appropriate branch in a new scope
		scope := NewEnv()
		scope.Outer = env
		scope.Namespace = env.Namespace

		if condition.GetValue().(bool) {
			Interpret(ifStmt.Body, &scope)
		} else {
			Interpret(ifStmt.Alternative, &scope)
		}

		// Propagate return upward
		if scope.Return != nil {
			env.Return = scope.Return
			return nil
		}

	case ast.VariableAssignmentStatement:
		varStmt := node.(ast.VariableAssignmentStatement)

		// Resolve variable name (needed because of arrays)
		name, err := GetVariableName(varStmt.Target)
		if err != nil {
			return err
		}

		target, ok := env.GetVar(name)

		// Extract array indexing chain if present
		var indexes []int64
		arr, isBinExpr := varStmt.Target.(*ast.BinaryExpression)
		isArray := isBinExpr && arr.Op.Value == "["

		for isBinExpr {
			indexObj, err := Eval(arr.Right, env)
			if err != nil {
				return err
			}

			if indexObj.Type() == object.Int {
				indexes = append(indexes, indexObj.GetValue().(int64))
			}

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

		// Type check
		if value.Type() != target.Type && !isArray {
			return fmt.Errorf("cannot assign value of type %s to variable of type %s", value.Type(), target.Type)
		}

		if !isArray {
			target.Value = value
		} else {
			// Reverse indexes (collected inside-out)
			for i, j := 0, len(indexes)-1; i < j; i, j = i+1, j-1 {
				indexes[i], indexes[j] = indexes[j], indexes[i]
			}
			setValueInArray(value, target.Value.(*object.ArrayObject), indexes)
		}

	case ast.ForStatement:
		forStmt := node.(ast.ForStatement)

		// Create loop scope
		scope := NewEnv()
		scope.Outer = env
		scope.Namespace = env.Namespace

		// Initialize loop variable if present
		if (forStmt.Init != ast.VariableDeclarationStatement{}) {
			scope.DeclareVar(forStmt.Init)
		}

		// Evaluate initial condition
		conditionVal, err := Eval(forStmt.Condition, &scope)
		if err != nil {
			return err
		}

		condition, ok := conditionVal.GetValue().(bool)
		if !ok {
			return fmt.Errorf("non-bool condition is not allowed")
		}

		// Transform body to include post-expression
		forStmt.Body = ast.BodyStatement{[]ast.Node{forStmt.Body, forStmt.Post}}

		for condition {
			Interpret(forStmt.Body, &scope)

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

	case ast.BodyStatement:
		body := node.(ast.BodyStatement)

		// Create new scope for block
		scope := NewEnv()
		scope.Outer = env
		scope.Namespace = env.Namespace

		for _, n := range body.Nodes {
			Interpret(n, &scope)

			// Propagate return upward
			if scope.Return != nil {
				env.Return = scope.Return
				return nil
			}
		}
	}

	return nil
}

// GetVariableName extracts base identifier from expression.
func GetVariableName(expression ast.Expression) (string, error) {
	if binary, ok := expression.(*ast.BinaryExpression); ok {
		return GetVariableName(binary.Left)
	} else if variable, ok := expression.(*ast.VariableExpression); ok {
		return variable.Value.Value.(string), nil
	}
	return "", fmt.Errorf("not identifier")
}

// setValueInArray assigns value into (possibly nested) array using index chain
func setValueInArray(value object.Object, array *object.ArrayObject, indexes []int64) {
	// Base case: assign directly if type matches or slot is empty
	if array.Value[int(indexes[0])] == nil || array.Value[int(indexes[0])].Type() == value.Type() {
		array.Value[int(indexes[0])] = value
		return
	}

	// Recursive descent into nested arrays
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
