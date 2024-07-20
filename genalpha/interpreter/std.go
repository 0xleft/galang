package interpreter

import (
	"fmt"
	"io"
	"os"
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
	"std.join": func(args []Result) Result {
		if len(args) != 2 {
			panic("std.join expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeArray && args[1].Type != genalphatypes.ASTNodeTypeString {
			panic("std.join expects array and string arguments")
		}

		array := args[0]
		separator := args[1]

		parts := make([]string, len(array.Values))
		for i, part := range array.Values {
			parts[i] = part.Value
		}

		return Result{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: strings.Join(parts, separator.Value),
		}
	},
	"std.read": func(args []Result) Result {
		if len(args) != 1 {
			panic("std.read expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString {
			panic("std.read expects string argument")
		}

		file, err := os.Open(args[0].Value)
		if err != nil {
			return Result{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}
		defer file.Close()

		contents, err := io.ReadAll(file)
		if err != nil {
			return Result{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}

		return Result{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: string(contents),
		}
	},
	"std.write": func(args []Result) Result {
		if len(args) != 2 {
			panic("std.write expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeString {
			panic("std.write expects string arguments")
		}

		file, err := os.Create(args[0].Value)
		if err != nil {
			return Result{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}
		defer file.Close()

		_, err = file.WriteString(args[1].Value)
		if err != nil {
			return Result{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}

		return Result{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: string(genalphatypes.KeywordTrue),
		}
	},
}
