package main

import (
	"encoding/json"
	"lizalang/interpreter"
	"lizalang/lexer"
	"lizalang/parser"
	"os"
	"path/filepath"
)

//"lizalang/token"

type Context struct {
	Root      string
	Path      string
	File      string
	AST       bool
	Interpret bool
}

func ParseArgs(args []string) *Context {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	ctx := Context{
		Root:      args[0],
		Path:      wd,
		File:      "main.li",
		AST:       false,
		Interpret: true,
	}
	for i, arg := range args {
		if i > 0 {
			if arg[0] == '-' {
				switch arg[0:] {
				case "-AST":
					ctx.AST = true
					break
				}
			} else {
				ctx.File = arg
			}
		}
	}
	return &ctx
}

func main() {
	ctx := ParseArgs(os.Args)
	file, err := os.ReadFile(filepath.Join(ctx.Path, ctx.File))
	if err != nil {
		panic(err)
	}
	lex := lexer.New(string(file), ctx.File)
	lex.Lex()

	par := parser.New(lex.Tokens)
	par.Parse()
	if ctx.AST {
		data := par.Program
		jsonAST, _ := json.MarshalIndent(data, "", "\t")
		err := os.WriteFile(filepath.Join(ctx.Path, ctx.File[:len(ctx.File)-3]+"AST.json"), jsonAST, 0644)
		if err != nil {
			panic(err)
		}
	}
	if ctx.Interpret {
		interpreter.GetNamespaces(&par.Program)
		interpreter.Init()
		env := interpreter.Namespaces["main"]
		env.Namespace = "main"
		main, _ := interpreter.Namespaces["main"].GetFunc("main")
		interpreter.Interpret(main.Body, env)

		/*jsonData, _ := json.MarshalIndent(interpreter.Namespaces, "", "\t")
		println(string(jsonData))*/
	}
}
