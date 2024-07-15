package interpreter

import (
	"fmt"

	genalphatypes "bobik.squidwock.com/root/genalphalang/genalpha"
)

type STDFunction func(args []string) Result

var STDFunctions = map[string]STDFunction{
	"std.print": func(args []string) Result {
		for _, arg := range args {
			fmt.Print(arg)
		}
		return Result{
			Type: genalphatypes.ASTNodeTypeNone,
		}
	},
	"std.println": func(args []string) Result {
		for _, arg := range args {
			fmt.Println(arg)
		}
		return Result{
			Type: genalphatypes.ASTNodeTypeNone,
		}
	},
	"std.exit": func(args []string) Result {
		if len(args) == 0 {
			panic("exit")
		}
		panic(args[0])
	},
}
