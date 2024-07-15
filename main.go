package main

import (
	"fmt"
	"os"

	"bobik.squidwock.com/root/genalphalang/genalpha/interpreter"
	"bobik.squidwock.com/root/genalphalang/genalpha/lexer"
	"bobik.squidwock.com/root/genalphalang/genalpha/parser"
	"bobik.squidwock.com/root/genalphalang/genalpha/utils"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: genalphalang <filename>")
		return
	}

	filename := os.Args[1]

	contents := utils.ReadContents(filename)
	tokens := lexer.Lex(contents, filename)
	ast := parser.Parse(tokens)
	interpreter.Interpret(&ast)
}
