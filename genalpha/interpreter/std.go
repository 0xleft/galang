package interpreter

import "fmt"

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
}
