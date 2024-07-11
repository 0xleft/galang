package interpreter

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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
	Args []string
	Body genalphatypes.ASTNode
}

type Variable struct {
	Name  string
	Value string
}

type InterpreterState struct {
	Functions []Function
	Variables []Variable
}

func Interpret(ast *genalphatypes.ASTNode) {
	if ast.Type != genalphatypes.ASTNodeTypeProgram {
		panic("Invalid AST type, parent should be a program node")
	}

	for _, child := range ast.Children {
		interpretNode(child)
	}
}

func interpretNode(node genalphatypes.ASTNode) {
	switch node.Type {
	case genalphatypes.ASTNodeTypeFunctionDeclaration:
		//interpretFunctionDeclaration(node)
	case genalphatypes.ASTNodeTypeVariableDeclaration:
		//interpretVariableDeclaration(node)
	case genalphatypes.ASTNodeTypeVariableAssignment:
		//interpretVariableAssignment(node)
	case genalphatypes.ASTNodeTypeIf:
		//interpretIf(node)
	case genalphatypes.ASTNodeTypeWhile:
		//interpretWhile(node)
	case genalphatypes.ASTNodeTypeReturn:
		//interpretReturn(node)
	case genalphatypes.ASTNodeTypeImport:
		//interpretImport(node)
	case genalphatypes.ASTNodeTypeBlock:
		//interpretBlock(node)
	default:
		panic("Invalid AST node type")
	}
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
	var firstLine = readFirstLineFile(filename)
	if firstLine == "" {
		// todo actualy parse here and then save the ast

		var contents = utils.ReadContents(filename)
		var tokens = lexer.Lex(contents, filename)
		var ast = parser.Parse(tokens)
		saveAST(ast, filename)

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
