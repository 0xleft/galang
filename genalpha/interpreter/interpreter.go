package interpreter

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	genalphatypes "bobik.squidwock.com/root/genalphalang/genalpha"
	"bobik.squidwock.com/root/genalphalang/genalpha/lexer"
	"bobik.squidwock.com/root/genalphalang/genalpha/parser"
	"bobik.squidwock.com/root/genalphalang/genalpha/utils"
)

type Scope struct {
	Variables []Variable
}

type Function struct {
	Name string
	Args []genalphatypes.ASTNode
	Body []genalphatypes.ASTNode
}

type Result struct {
	Type  genalphatypes.ASTNodeType
	Value string
}

type Variable struct {
	Name  string
	Type  genalphatypes.ASTNodeType
	Value string
}

type InterpreterState struct {
	Functions []Function

	ScopeStack  []Scope
	LocalScope  Scope
	GlobalScope Scope
}

func Interpret(ast *genalphatypes.ASTNode, filename string) {
	var interpreterState = InterpreterState{}

	if ast.Type != genalphatypes.ASTNodeTypeProgram {
		panic("Invalid AST type, parent should be a program node")
	}

	saveAST(*ast, filename, filename+"+")

	for _, child := range ast.Children {
		interpretNode(&interpreterState, child)
	}

	for _, function := range interpreterState.Functions {
		if function.Name == "main" {
			for _, instructionNode := range function.Body {
				interpretNode(&interpreterState, instructionNode)
			}

			return
		}
	}

	panic("No main function found, please declare a main function with the name 'main'")
}

func newScope(interpreterState *InterpreterState, scope Scope) {
	interpreterState.ScopeStack = append(interpreterState.ScopeStack, interpreterState.LocalScope)
	interpreterState.LocalScope = scope
}

func popScope(interpreterState *InterpreterState) {
	if len(interpreterState.ScopeStack) == 0 {
		panic("No scope to pop")
	}

	interpreterState.LocalScope = interpreterState.ScopeStack[len(interpreterState.ScopeStack)-1]
	interpreterState.ScopeStack = interpreterState.ScopeStack[:len(interpreterState.ScopeStack)-1]
}

func interpretNode(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	switch node.Type {
	case genalphatypes.ASTNodeTypeFunctionDeclaration:
		interpretFunctionDeclaration(interpreterState, node)
		return Result{
			Type:  genalphatypes.ASTNodeTypeUnknown,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeVariableDeclaration:
		interpretVariableDeclaration(interpreterState, node)
		return Result{
			Type:  genalphatypes.ASTNodeTypeUnknown,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeVariableAssignment:
		interpretVariableAssignment(interpreterState, node)
		return Result{
			Type:  genalphatypes.ASTNodeTypeUnknown,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeIf:
		return interpretIf(interpreterState, node)
	case genalphatypes.ASTNodeTypeWhile:
		return interpretWhile(interpreterState, node)
	case genalphatypes.ASTNodeTypeReturn:
		return interpretReturn(interpreterState, node)
	case genalphatypes.ASTNodeTypeImport:
		interpretImport(interpreterState, node)
		return Result{
			Type:  genalphatypes.ASTNodeTypeUnknown,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeBlock:
		return interpretBlock(interpreterState, node)
	case genalphatypes.ASTNodeTypeFunctionCall:
		return resolveFunctionCall(interpreterState, node)
	case genalphatypes.ASTNodeTypeFunctionArgument:
		return Result{
			Type:  genalphatypes.ASTNodeTypeUnknown,
			Value: "",
		}
	default:
		panic("Invalid AST node type" + fmt.Sprint(node.Type))
	}
}

func interpretFunctionDeclaration(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
	var name = node.Children[0].Value
	var args = []genalphatypes.ASTNode{}
	var bodyStart = 1
	for _, arg := range node.Children {
		if arg.Type == genalphatypes.ASTNodeTypeFunctionArgument {
			args = append(args, arg)
			bodyStart++
		} else {
			break
		}
	}

	var function = Function{
		Name: name,
		Args: args,
		Body: node.Children[bodyStart:],
	}

	interpreterState.Functions = append(interpreterState.Functions, function)
}

func resolveExpression(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	if node.Type == genalphatypes.ASTNodeTypeIdentifier {
		return resolveIdentifier(interpreterState, node)
	}

	if node.Type == genalphatypes.ASTNodeTypeNumber {
		return Result{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: node.Value,
		}
	}

	if node.Type == genalphatypes.ASTNodeTypeString {
		return Result{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: node.Value,
		}
	}

	if node.Type == genalphatypes.ASTNodeTypeBoolean {
		return Result{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: node.Value,
		}
	}

	if node.Type == genalphatypes.ASTNodeTypeBinaryOperation {
		return resolveBinaryOperation(interpreterState, node)
	}

	if node.Type == genalphatypes.ASTNodeTypeUnaryOperation {
		return resolveUnaryOperation(interpreterState, node)
	}

	if node.Type == genalphatypes.ASTNodeTypeFunctionCall {
		return resolveFunctionCall(interpreterState, node)
	}

	panic("Invalid expression node type " + fmt.Sprint(node.Type))
}

func resolveIdentifier(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	var name = node.Value
	for _, variable := range interpreterState.LocalScope.Variables {
		if variable.Name == name {
			return Result{
				Type:  variable.Type,
				Value: variable.Value,
			}
		}
	}

	for _, variable := range interpreterState.GlobalScope.Variables {
		if variable.Name == name {
			return Result{
				Type:  variable.Type,
				Value: variable.Value,
			}
		}
	}

	panic("Variable " + name + " not found")
}

func resolveBinaryOperation(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	var left = resolveExpression(interpreterState, node.Children[0])
	var right = resolveExpression(interpreterState, node.Children[1])

	// todo
	var _ = left.Type & right.Type

	switch node.Value {
	case "+":
	case "-":
	case "*":
	case "/":
	case "%":
	case "==":
	case "!=":
	case ">":
	case "<":
	case ">=":
	case "<=":
	default:
		panic("Invalid binary operation " + node.Value)
	}

	return Result{
		Type: genalphatypes.ASTNodeTypeUnknown,
	}
}

func resolveUnaryOperation(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	var operand = resolveExpression(interpreterState, node.Children[0])

	switch node.Value {
	case "!":
		if operand.Type != genalphatypes.ASTNodeTypeBoolean {
			panic("Invalid operand type for unary operation !")
		}

		var value = string(genalphatypes.KeywordFalse)
		if operand.Value == string(genalphatypes.KeywordFalse) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Result{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	default:
		panic("Invalid unary operation " + node.Value)
	}

	return Result{
		Type: genalphatypes.ASTNodeTypeUnknown,
	}
}

func resolveFunctionCall(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	var name = node.Children[0].Value
	for _, function := range interpreterState.Functions {
		if function.Name == name {
			if len(function.Args) != len(node.Children)-1 {
				panic("Invalid number of arguments for function " + name)
			}

			var scope = Scope{}
			for i, arg := range function.Args {
				var argValue = resolveExpression(interpreterState, node.Children[i+1])
				scope.Variables = append(scope.Variables, Variable{
					Name:  arg.Value,
					Type:  arg.Type,
					Value: argValue.Value,
				})
			}

			newScope(interpreterState, scope)
			for _, instructionNode := range function.Body {
				var result = interpretNode(interpreterState, instructionNode)
				if result.Type != genalphatypes.ASTNodeTypeUnknown {
					popScope(interpreterState)
					return result
				}
			}
			popScope(interpreterState)

			return Result{
				Type:  genalphatypes.ASTNodeTypeUnknown,
				Value: "",
			}
		}
	}

	panic("Function " + name + " not found")
}

func interpretBlock(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	return Result{
		Type:  genalphatypes.ASTNodeTypeUnknown,
		Value: "",
	}
}

func interpretIf(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	return Result{
		Type:  genalphatypes.ASTNodeTypeUnknown,
		Value: "",
	}
}

func interpretWhile(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	return Result{
		Type:  genalphatypes.ASTNodeTypeUnknown,
		Value: "",
	}
}

func interpretReturn(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	return Result{
		Type:  genalphatypes.ASTNodeTypeUnknown,
		Value: "",
	}
}

func interpretImport(interpreterState *InterpreterState, node genalphatypes.ASTNode) Result {
	if len(node.Children) != 1 {
		panic("import should be done with one argument, the file to import, such as 'gyat \"test.gal\"'")
	}

	var filename = node.Children[0].Value
	var isString = node.Children[0].Type == genalphatypes.ASTNodeTypeString
	if !isString {
		panic("import should be done with a string argument, the file to import, such as 'gyat \"test.gal\"'")
	}

	var ast = loadAST(filename)
	for _, child := range ast.Children {
		interpretNode(interpreterState, child)
	}

	return Result{
		Type:  genalphatypes.ASTNodeTypeUnknown,
		Value: "",
	}
}

func interpretVariableDeclaration(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

func interpretVariableAssignment(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

// returns the sha256 hash of the given ast
func sha256Hash(content string) string {
	var shaBytes = sha256.Sum256([]byte(content))
	return hex.EncodeToString(shaBytes[:])
}

func getNodeBytes(node genalphatypes.ASTNode) []byte {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)

	err := enc.Encode(node)
	if err != nil {
		panic("Error encoding node")
	}

	return network.Bytes()
}

func getNodeFromBytes(nodeBytes []byte) genalphatypes.ASTNode {
	var node genalphatypes.ASTNode
	dec := gob.NewDecoder(bytes.NewReader(nodeBytes))

	err := dec.Decode(&node)
	if err != nil {
		panic("Error decoding node")
	}

	return node
}

func readFirstLineFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := scanner.Text()
		return firstLine
	}

	return ""
}

func writeToFile(filename string, data string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		panic(err)
	}
}

// get everything except the first line of the file
func readNotFirstLineFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	scanner.Scan()
	return scanner.Text()
}

// load and save are for when importing new files so we dont have to lex and parse them again (optimization and easier to work with)
func loadAST(filename string) genalphatypes.ASTNode {
	var firstLine = readFirstLineFile(filename + "+")
	if firstLine == "" {
		// parse here and then save the ast
		var contents = utils.ReadContents(filename)
		var tokens = lexer.Lex(contents, filename)
		var ast = parser.Parse(tokens)
		saveAST(ast, filename, filename+"+")

		return ast
	}

	var nodeString = readNotFirstLineFile(filename)
	if len(nodeString) == 0 {
		return genalphatypes.ASTNode{}
	}

	nodeString = strings.ReplaceAll(nodeString, "\\n", "\n")

	return getNodeFromBytes([]byte(nodeString))
}

func saveAST(ast genalphatypes.ASTNode, sourceFilename string, filename string) {
	// format:
	// - sha256 hash of the ast
	// - the rest of the ast in gob format

	// first we need to get the sha256 hash of the ast
	var contents = utils.ReadContents(sourceFilename)
	var sha = sha256Hash(contents)
	if readFirstLineFile(filename) == sha {
		return
	}

	// replace any new lines with \\n
	var bytes = string(getNodeBytes(ast))
	bytes = strings.ReplaceAll(bytes, "\n", "\\n")

	writeToFile(filename, sha+"\n"+bytes)
}
