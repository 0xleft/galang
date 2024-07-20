package interpreter

import (
	"fmt"
	"strings"

	genalphatypes "bobik.squidwock.com/root/gal/genalpha"
)

type STDFunction func(args []Result) Result

var STDFunctions = map[string]STDFunction{
	"std.print": func(args []Result) Result {
		for _, arg := range args {
			fmt.Print(arg.Value)
		}
		return Result{
			Type: genalphatypes.ASTNodeTypeNone,
		}
	},
	"std.println": func(args []Result) Result {
		for _, arg := range args {
			fmt.Println(arg.Value)
		}
		return Result{
			Type: genalphatypes.ASTNodeTypeNone,
		}
	},
	"std.exit": func(args []Result) Result {
		if len(args) == 0 {
			panic("exit")
		}
		panic(args[0].Value)
	},
	"std.len": func(args []Result) Result {
		if len(args) != 1 {
			panic("std.len expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString {
			panic("std.len expects a string argument")
		}

		return Result{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprintf("%d", len(args[0].Value)),
		}
	},
	"std.split": func(args []Result) Result {
		if len(args) != 2 {
			panic("std.split expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeString {
			panic("std.split expects string arguments")
		}

		toSplit := args[0]
		separator := args[1]

		parts := strings.Split(toSplit.Value, separator.Value)

		results := make([]Result, len(parts))
		for i, part := range parts {
			results[i] = Result{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: part,
			}
		}

		return Result{
			Type:   genalphatypes.ASTNodeTypeArray,
			Value:  fmt.Sprint(len(results)),
			Values: results,
		}
	},
}
