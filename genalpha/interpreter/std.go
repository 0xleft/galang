package interpreter

import (
	"fmt"
)

type STDFunction func(args []string) string

var STDFunctions = map[string]STDFunction{
	"std.print": func(args []string) string {
		for _, arg := range args {
			fmt.Print(arg)
		}
		return ""
	},
	"std.println": func(args []string) string {
		for _, arg := range args {
			fmt.Println(arg)
		}
		return ""
	},
	"std.exit": func(args []string) string {
		if len(args) == 0 {
			panic("exit")
		}
		panic(args[0])
	},
	"std.eval": func(args []string) string {
		return ""
	},
	"std.exec": func(args []string) string {
		if len(args) == 0 {
			return ""
		}
		return ""
	},
	// execute a system call
	"std.syscall": func(args []string) string {
		return ""
	},
}
