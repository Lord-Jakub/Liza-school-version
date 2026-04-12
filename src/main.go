package main

import (
	"encoding/json"
	"fmt"
	"lizalang/interpreter"
	"lizalang/lexer"
	"lizalang/parser"
	"log"
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
		log.Fatal(err)
	}
	exec, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	ctx := Context{
		Root:      filepath.Dir(exec),
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
		log.Fatal(err)
	}
	lex := lexer.New(string(file), ctx.File)
	lex.Lex()
	if len(lex.Errors) > 0 {
		for _, err := range lex.Errors {
			fmt.Print(err.Error())
		}
		return
	}
	par := parser.New(lex.Tokens)
	par.Parse(ctx.Root, ctx.Path)
	if len(par.Errors) > 0 {
		for _, err := range par.Errors {
			fmt.Println(err)
		}
		return
	}
	if ctx.AST {
		data := par.Program
		jsonAST, _ := json.MarshalIndent(data, "", "\t")
		err := os.WriteFile(filepath.Join(ctx.Path, ctx.File[:len(ctx.File)-3]+"AST.json"), jsonAST, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	if ctx.Interpret {
		//utils.PrintData(par.Program.Namespaces)
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
