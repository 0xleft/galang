package main

import (
	"flag"
	"fmt"
	"os"

	"bobik.squidwock.com/root/gal/genalpha/interpreter"
	"bobik.squidwock.com/root/gal/genalpha/lexer"
	"bobik.squidwock.com/root/gal/genalpha/parser"
	"bobik.squidwock.com/root/gal/genalpha/utils"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: ", r)
		}
	}()

	inlineScript := flag.String("c", "", "Inline script to run")

	flag.Parse()
	args := flag.Args()

	if *inlineScript == "" && len(args) == 0 {
		fmt.Println("Usage: gal <filename>")
		return
	}

	if *inlineScript != "" {
		tokens := lexer.Lex(*inlineScript)
		ast := parser.Parse(tokens)
		interpreter.Interpret(&ast, []string{}, "")
		return
	}

	filename := args[0]
	contents := utils.ReadContents(filename)
	tokens := lexer.Lex(contents)
	ast := parser.Parse(tokens)
	interpreter.Interpret(&ast, os.Args, filename)
}
