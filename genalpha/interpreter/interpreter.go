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

// there is a local scope for each function and a global scope that includes everything declared in the main function
type Scope struct {
	Variables []Variable
}

type Function struct {
	Name string
	Args []genalphatypes.ASTNode
	Body []genalphatypes.ASTNode
}

type Variable struct {
	Name  string
	Type  genalphatypes.ASTNodeType
	Value string
}

type InterpreterState struct {
	Functions   []Function
	ScopeStack  []Scope
	ReturnStack []genalphatypes.ASTNode
	ReturnIndex int

	Scope       Scope
	GlobalScope Scope
}

func Interpret(ast *genalphatypes.ASTNode) {
	var interpreterState = InterpreterState{}

	if ast.Type != genalphatypes.ASTNodeTypeProgram {
		panic("Invalid AST type, parent should be a program node")
	}

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

func interpretNode(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
	switch node.Type {
	case genalphatypes.ASTNodeTypeFunctionDeclaration:
		interpretFunctionDeclaration(interpreterState, node)
	case genalphatypes.ASTNodeTypeVariableDeclaration:
		interpretVariableDeclaration(interpreterState, node)
	case genalphatypes.ASTNodeTypeVariableAssignment:
		interpretVariableAssignment(interpreterState, node)
	case genalphatypes.ASTNodeTypeIf:
		interpretIf(interpreterState, node)
	case genalphatypes.ASTNodeTypeWhile:
		interpretWhile(interpreterState, node)
	case genalphatypes.ASTNodeTypeReturn:
		interpretReturn(interpreterState, node)
	case genalphatypes.ASTNodeTypeImport:
		interpretImport(interpreterState, node)
	case genalphatypes.ASTNodeTypeBlock:
		interpretBlock(interpreterState, node)
	case genalphatypes.ASTNodeTypeFunctionCall:
		interpreterFunctionCall(interpreterState, node)
	case genalphatypes.ASTNodeTypeFunctionArgument:
		// do nothing
		return
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

func interpreterFunctionCall(interpreterState *InterpreterState, node genalphatypes.ASTNode) {

}

func interpretBlock(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

func interpretIf(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

func interpretWhile(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

func interpretReturn(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

func interpretImport(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
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
}

func interpretVariableDeclaration(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

func interpretVariableAssignment(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
}

// returns the sha256 hash of the given ast
func sha256AST(node genalphatypes.ASTNode) string {
	var bytes = getNodeBytes(node)
	var shaBytes = sha256.Sum256(bytes)
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
		saveAST(ast, filename+"+")

		return ast
	}

	var nodeString = readNotFirstLineFile(filename)
	if len(nodeString) == 0 {
		return genalphatypes.ASTNode{}
	}

	nodeString = strings.ReplaceAll(nodeString, "\\n", "\n")

	return getNodeFromBytes([]byte(nodeString))
}

func saveAST(ast genalphatypes.ASTNode, filename string) {
	// format:
	// - sha256 hash of the ast
	// - the rest of the ast in gob format

	// first we need to get the sha256 hash of the ast
	var sha = sha256AST(ast)
	if readFirstLineFile(filename) == sha {
		return
	}

	// replace any new lines with \\n
	var bytes = string(getNodeBytes(ast))
	bytes = strings.ReplaceAll(bytes, "\n", "\\n")

	writeToFile(filename, sha+"\n"+bytes)
}
