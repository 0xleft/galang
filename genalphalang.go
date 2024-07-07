package main

import (
	"fmt"
	"os"

	"bobik.squidwock.com/root/genalphalang/genalphalang/interpreter"
	"bobik.squidwock.com/root/genalphalang/genalphalang/lexer"
	"bobik.squidwock.com/root/genalphalang/genalphalang/parser"
	"bobik.squidwock.com/root/genalphalang/genalphalang/utils"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Execution failed: ", r)
		}
	}()

	if len(os.Args) < 2 {
		fmt.Println("Usage: genalphalang <filename>")
		return
	}

	filename := os.Args[1]
	contents := utils.ReadContents(filename)
	tokens := lexer.Lex(contents, filename)
	fmt.Println(tokens)
	ast := parser.Parse(tokens)
	interpreter.Interpret(&ast)
}
