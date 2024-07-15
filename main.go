package main

import (
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
			fmt.Println(r)
		}
	}()

	if len(os.Args) < 2 {
		fmt.Println("Usage: genalphalang <filename>")
		return
	}

	filename := os.Args[1]

	contents := utils.ReadContents(filename)
	tokens := lexer.Lex(contents)
	ast := parser.Parse(tokens)
	interpreter.Interpret(&ast, os.Args[1:])
}
