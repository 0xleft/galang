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
		for _, value := range array.Indecies {
			parts = append(parts, value.Value)
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
	"std.char": func(args []Variable) Variable {
		// returns a string from a keycode
		if len(args) != 1 {
			panic("std.char expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeNumber {
			panic("std.char expects a number argument")
		}

		char, err := strconv.Atoi(args[0].Value)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: string(rune(char)),
		}
	},
	"std.insert": func(args []Variable) Variable {
		if len(args) != 3 {
			panic("std.insert expects exactly 3 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeNumber && args[2].Type != genalphatypes.ASTNodeTypeString {
			panic("std.insert expects string, number, and string arguments")
		}

		str := args[0].Value
		index, err := strconv.Atoi(args[1].Value)
		if err != nil {
			panic("std.insert expects a number argument")
		}
		insert := args[2].Value

		str = str[:index] + insert + str[index:]

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: str,
		}
	},
	"std.slice": func(args []Variable) Variable {
		if len(args) != 3 {
			panic("std.slice expects exactly 3 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeNumber && args[2].Type != genalphatypes.ASTNodeTypeNumber {
			panic("std.slice expects string, number, and number arguments")
		}

		str := args[0].Value
		start, err := strconv.Atoi(args[1].Value)
		if err != nil {
			panic("std.slice expects a number argument")
		}
		end, err := strconv.Atoi(args[2].Value)
		if err != nil {
			panic("std.slice expects a number argument")
		}

		str = str[start:end]

		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: str,
		}
	},
	"std.len_indecies": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("len expects exactly 1 argument")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprintf("%d", len(args[0].Indecies)),
		}
	},
	"std.input": func(args []Variable) Variable {
		if len(args) != 2 {
			panic("std.input expects exactly 2 arguments")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeString && args[1].Type != genalphatypes.ASTNodeTypeNumber {
			panic("std.input expects string and a number argument")
		}

		length, err := strconv.Atoi(args[1].Value)
		if err != nil {
			panic("std.input expects a number argument")
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
	"std.writable": func(args []Variable) Variable {
		if len(args) != 1 {
			panic("std.writable expects exactly 1 argument")
		}
		if args[0].Type != genalphatypes.ASTNodeTypeNumber {
			panic("std.writable expects a number argument")
		}

		// returns if the number which is a keycode is a writable character
		char, err := strconv.Atoi(args[0].Value)
		if err != nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}

		writable_chars := []string{
			" ", "!", "\"", "#", "$", "%", "&", "'", "(", ")", "*", "+", ",", "-", ".", "/",
			"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ":", ";", "<", "=", ">", "?",
			"@", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O",
			"P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "[", "\\", "]", "^", "_",
			"`", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o",
			"p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "{", "|", "}", "~", "\t",
		}

		value := string(genalphatypes.KeywordFalse)
		if strings.Contains(strings.Join(writable_chars, ""), string(rune(char))) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
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
