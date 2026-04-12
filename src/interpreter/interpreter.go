package interpreter

import (
	"bufio"
	"fmt"
	"lizalang/ast"
	"lizalang/interpreter/object"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	// Namespaces holds global environments for each namespace.
	Namespaces map[string]*Environment = make(map[string]*Environment)

	// BuildIns maps builtin function names to their implementations.
	BuildIns map[string]func(*Environment, []ast.Expression) error = make(map[string]func(*Environment, []ast.Expression) error)
)

// Init registers all builtin functions into the interpreter.
func Init() {
	// print(...) - prints evaluated arguments
	var writer = bufio.NewWriter(os.Stdout)
	BuildIns["print"] = func(env *Environment, args []ast.Expression) error {
		for _, arg := range args {
			argVal, err := Eval(arg, env)
			if err != nil {
				return err
			}
			if argVal == nil {
				fmt.Fprint(writer, "<nil>")
				continue
			}

			switch v := argVal.(type) {
			case *object.ArrayObject:
				fmt.Fprint(writer, "[")
				for i, elem := range v.Value {
					if i > 0 {
						fmt.Fprint(writer, ", ")
					}
					if elem != nil {
						fmt.Fprint(writer, elem.GetValue())
					} else {
						fmt.Fprint(writer, "null")
					}
				}
				fmt.Fprint(writer, "]")
			default:
				fmt.Fprint(writer, argVal.GetValue())
			}
		}
		writer.Flush()
		return nil
	}
	// read() - reads user input from console
	var stdinReader = bufio.NewReader(os.Stdin)
	BuildIns["read"] = func(env *Environment, args []ast.Expression) error {
		if args != nil {
			return fmt.Errorf("read() doesn't accept arguments")
		}
		text, err := stdinReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		value := strings.TrimSpace(text)

		var obj object.Object
		obj = &object.StringObject{Value: value}
		env.Return = &obj
		return nil
	}

	// len(array) - returns length of an array
	BuildIns["len"] = func(env *Environment, args []ast.Expression) error {
		if len(args) > 1 {
			return fmt.Errorf("Too many argumetns")
		} else if len(args) < 1 {
			return fmt.Errorf("Not enough argumetns")
		}

		value, err := Eval(args[0], env)
		if err != nil {
			return err
		}

		arr, ok := value.(*object.ArrayObject)
		if !ok {
			return fmt.Errorf("argument must be array")
		}

		var obj object.Object
		obj = &object.IntObject{int64(arr.Len)}

		env.Return = &obj
		return nil
	}

	// IntToFloat(int) - converts int to float
	BuildIns["IntToFloat"] = func(env *Environment, args []ast.Expression) error {
		if len(args) != 1 {
			return fmt.Errorf("You can have only one argument in IntToFloat()")
		}

		argVal, err := Eval(args[0], env)
		if err != nil {
			return err
		}

		if argVal.Type() != object.Int {
			return fmt.Errorf("IntToFloat() accept only int as an argument, not %s", argVal.Type())
		}

		var obj object.Object
		obj = &object.FloatObject{float64(argVal.GetValue().(int64))}
		env.Return = &obj
		return nil
	}

	// FloatToInt(float) - converts float to int
	BuildIns["FloatToInt"] = func(env *Environment, args []ast.Expression) error {
		if len(args) != 1 {
			return fmt.Errorf("You can have only one argument in FloatToInt()")
		}

		argVal, err := Eval(args[0], env)
		if err != nil {
			return err
		}

		if argVal.Type() != object.Float {
			return fmt.Errorf("FloatToInt() accept only float as an argument, not %s", argVal.Type())
		}

		var obj object.Object
		obj = &object.IntObject{int64(argVal.GetValue().(float64))}
		env.Return = &obj
		return nil
	}
	// IntToString(int) - converts int to String
	BuildIns["IntToString"] = func(env *Environment, args []ast.Expression) error {
		if len(args) != 1 {
			return fmt.Errorf("You can have only one argument in IntToString()")
		}

		argVal, err := Eval(args[0], env)
		if err != nil {
			return err
		}

		if argVal.Type() != object.Int {
			return fmt.Errorf("IntToString() accept only int as an argument, not %s", argVal.Type())
		}

		var obj object.Object
		obj = &object.StringObject{strconv.Itoa(int(argVal.GetValue().(int64)))}
		env.Return = &obj
		return nil
	}
	// StringToInt(String) - converts string to int
	BuildIns["StringToInt"] = func(env *Environment, args []ast.Expression) error {
		if len(args) != 1 {
			return fmt.Errorf("You can have only one argument in StringToInt()")
		}
		argVal, err := Eval(args[0], env)
		if err != nil {
			return err
		}

		if argVal.Type() != object.String {
			return fmt.Errorf("StringToInt() accept only String as an argument, not %s", argVal.Type())
		}
		num, err := strconv.Atoi(argVal.GetValue().(string))
		if err != nil {
			return err
		}
		var obj object.Object
		obj = &object.IntObject{int64(num)}
		env.Return = &obj
		return nil
	}
	// exit(int n) exits a program with exit code n
	BuildIns["exit"] = func(env *Environment, args []ast.Expression) error {
		if len(args) != 1 {
			return fmt.Errorf("You must have one argument in exit()")
		}

		argVal, err := Eval(args[0], env)
		if err != nil {
			return err
		}

		if argVal.Type() != object.Int {
			return fmt.Errorf("exit() accept only int as an argument, not %s", argVal.Type())
		}

		os.Exit(int(argVal.GetValue().(int64)))
		return nil
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
			err := setValueInArray(value, target.Value.(*object.ArrayObject), indexes)
			if err != nil {
				return err
			}
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
		if (forStmt.Post != ast.VariableAssignmentStatement{}) {
			forStmt.Body = ast.BodyStatement{[]ast.Node{forStmt.Body, forStmt.Post}}
		}

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
			err := Interpret(n, &scope)
			if err != nil {
				fmt.Println()

				log.Fatal(fmt.Errorf("File %s, line %d: %s", n.File(), n.Line(), err.Error()))
			}

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
func setValueInArray(value object.Object, array *object.ArrayObject, indexes []int64) error {
	// Base case: assign directly if type matches or slot is empty
	if len(indexes) == 1 {
		if array.Value[int(indexes[0])] == nil || array.Value[int(indexes[0])].Type() == value.Type() {
			array.Value[int(indexes[0])] = value
			return nil
		}
		return fmt.Errorf("type mismatch: cannot assign %s to %s", value.Type(), array.Value[int(indexes[0])].Type())
	}

	// Recursive descent into nested arrays
	if len(indexes) > 1 {
		if arr, ok := array.Value[int(indexes[0])].(*object.ArrayObject); ok {
			return setValueInArray(value, arr, indexes[1:])
		} else {
			return fmt.Errorf("index %d is not an array", indexes[0])
		}
	} else {
		return fmt.Errorf("no indexes provided")
	}
}
