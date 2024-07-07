package compiler

import "bobik.squidwock.com/root/genalphalang/genalphalang/parser"

type Platform int

const (
	PlatformX86 Platform = iota
	PlatformX64
	PlatformARM
	PlatformARM64
)

func GenerateASM(ast *parser.ASTNode) {
	panic("Not implemented")
}
