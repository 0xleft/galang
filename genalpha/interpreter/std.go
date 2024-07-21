package interpreter

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/term"

	genalphatypes "bobik.squidwock.com/root/gal/genalpha"
)

type STDFunction func(args []Variable) Variable

var STDFunctions = map[string]STDFunction{
	"std.print": func(args []Variable) Variable {
		for _, arg := range args {
			fmt.Print(arg.Value)
		}
		return Variable{
			Type: genalphatypes.ASTNodeTypeNone,
		}
	},
	"std.println": func(args []Variable) Variable {
		for _, arg := range args {
			fmt.Println(arg.Value)
		}
		return Variable{
			Type: genalphatypes.ASTNodeTypeNone,
		}
	},
	"std.exit": func(args []Variable) Variable {
		if len(args) == 0 {
			panic("exit")
		}
		panic(args[0].Value)
	},
	"std.len": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("std.len expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString {
			panic("std.len expects a string argument")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprintf("%d", len(args[0].Value)),
		}
	},
	"std.split": func(args []Variable) Variable {
		if len(args) != 2 {
			panic("std.split expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeString {
			panic("std.split expects string arguments")
		}

		toSplit := args[0]
		separator := args[1]

		parts := strings.Split(toSplit.Value, separator.Value)

		results := map[string]*Variable{}
		for i, part := range parts {
			results[fmt.Sprint(i)] = &Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: part,
			}
		}

		return Variable{
			Type:     genalphatypes.ASTNodeTypeArray,
			Value:    fmt.Sprint(len(results)),
			Indecies: results,
		}
	},
	"std.join": func(args []Variable) Variable {
		if len(args) != 2 {
			panic("std.join expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeArray && args[1].Type != genalphatypes.ASTNodeTypeString {
			panic("std.join expects array and string arguments")
		}

		array := args[0]
		separator := args[1]

		parts := []string{}
		for i := 0; i < len(array.Indecies); i++ {
			parts = append(parts, array.Indecies[fmt.Sprint(i)].Value)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: strings.Join(parts, separator.Value),
		}
	},
	"std.read": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("std.read expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString {
			panic("std.read expects string argument")
		}

		file, err := os.Open(args[0].Value)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}
		defer file.Close()

		contents, err := io.ReadAll(file)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: string(contents),
		}
	},
	"std.write": func(args []Variable) Variable {
		if len(args) != 2 {
			panic("std.write expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeString {
			panic("std.write expects string arguments")
		}

		file, err := os.Create(args[0].Value)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}
		defer file.Close()

		_, err = file.WriteString(args[1].Value)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: string(genalphatypes.KeywordTrue),
		}
	},
	"std.exists": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("std.exists expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString {
			panic("std.exists expects string argument")
		}

		_, err := os.Stat(args[0].Value)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: string(genalphatypes.KeywordTrue),
		}
	},
	"std.shell": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("std.shell expects exactly 1 argument")
		}

		cmd := exec.Command(args[0].Value)
		output, err := cmd.Output()
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: err.Error(),
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: string(output),
		}
	},
	"std.inputln": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("std.inputln expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString {
			panic("std.inputln expects string argument")
		}

		fmt.Print(args[0].Value)

		var input string
		fmt.Scanln(&input)

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: input,
		}
	},
	"std.binput": func(args []Variable) Variable {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		b := make([]byte, 1)
		_, err = os.Stdin.Read(b)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}

		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: "",
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprintf("%d", b[0]),
		}
	},
	"std.input": func(args []Variable) Variable {
		if len(args) != 2 {
			panic("std.input expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeNumber {
			panic("std.inputln expects string and a number argument")
		}

		length, err := strconv.Atoi(args[1].Value)
		if err != nil {
			panic("std.inputln expects a number argument")
		}

		fmt.Print(args[0].Value)

		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: "",
			}
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		b := make([]byte, length)
		for i := 0; i < length; i++ {
			_, err := os.Stdin.Read(b[i : i+1])
			if err != nil {
				return Variable{
					Type:  genalphatypes.ASTNodeTypeString,
					Value: "",
				}
			}
		}

		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: "",
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: string(b),
		}
	},
	"term.term_width": func(args []Variable) Variable {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNumber,
				Value: "0",
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprintf("%d", width),
		}
	},
	"term.term_height": func(args []Variable) Variable {
		_, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNumber,
				Value: "0",
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprintf("%d", height),
		}
	},
}
