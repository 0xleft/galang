package interpreter

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"

	genalphatypes "bobik.squidwock.com/root/gal/genalpha"
	"bobik.squidwock.com/root/gal/genalpha/lexer"
	"bobik.squidwock.com/root/gal/genalpha/parser"
	"bobik.squidwock.com/root/gal/genalpha/utils"
)

type Scope struct {
	Variables map[string]*Variable
}

type Function struct {
	Name string
	Args []genalphatypes.ASTNode
	Body []genalphatypes.ASTNode
}

type Variable struct {
	Type     genalphatypes.ASTNodeType
	Value    string
	Indecies map[string]*Variable
}

type InterpreterState struct {
	Functions map[string]Function

	ScopeStack  []Scope
	LocalScope  Scope
	GlobalScope Scope

	ImportedFiles []string
}

func Interpret(ast *genalphatypes.ASTNode, args []string, filename string) {
	interpreterState := InterpreterState{
		Functions: map[string]Function{},
		GlobalScope: Scope{
			Variables: map[string]*Variable{},
		},
		LocalScope: Scope{
			Variables: map[string]*Variable{},
		},
	}

	interpreterState.LocalScope.Variables["args"] = &Variable{
		Type:     genalphatypes.ASTNodeTypeNumber,
		Value:    fmt.Sprint(len(args)),
		Indecies: map[string]*Variable{},
	}
	for i, arg := range args {
		interpreterState.LocalScope.Variables["args"].Indecies[fmt.Sprint(i)] = &Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: arg,
		}
	}

	if ast.Type != genalphatypes.ASTNodeTypeProgram {
		panic("Invalid AST type, parent should be a program node")
	}

	for _, child := range ast.Children {
		interpretNode(&interpreterState, child, filename)
	}

	for _, function := range interpreterState.Functions {
		if function.Name == "main" {
			for _, instructionNode := range function.Body {
				variable := interpretNode(&interpreterState, instructionNode, "")
				if variable.Type != genalphatypes.ASTNodeTypeNone {
					fmt.Println("Program exited with code:", variable.Value)
					return
				}
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

func interpretNode(interpreterState *InterpreterState, node genalphatypes.ASTNode, filename string) Variable {
	switch node.Type {
	case genalphatypes.ASTNodeTypeMemberAssignment:
		interpretMemberAssignment(interpreterState, node)
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeFunctionDeclaration:
		interpretFunctionDeclaration(interpreterState, node)
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeVariableDeclaration:
		interpretVariableDeclaration(interpreterState, node)
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeVariableAssignment:
		interpretVariableAssignment(interpreterState, node)
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeIf:
		return interpretIf(interpreterState, node)
	case genalphatypes.ASTNodeTypeWhile:
		return interpretWhile(interpreterState, node)
	case genalphatypes.ASTNodeTypeReturn:
		return interpretReturn(interpreterState, node)
	case genalphatypes.ASTNodeTypeImport:
		interpretImport(interpreterState, node, filename)
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	case genalphatypes.ASTNodeTypeFunctionCall:
		return resolveFunctionCall(interpreterState, node)
	case genalphatypes.ASTNodeTypeFunctionArgument:
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	default:
		fmt.Println(node.Type, node.Value, filename)
		panic("Invalid AST node type" + fmt.Sprint(node.Type))
	}
}

func interpretFunctionDeclaration(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
	name := node.Children[0].Value
	args := []genalphatypes.ASTNode{}
	bodyStart := 0

	for _, arg := range node.Children {
		if arg.Type == genalphatypes.ASTNodeTypeFunctionArgument {
			args = append(args, arg)
			bodyStart++
		}
	}

	bodyStart++

	function := Function{
		Name: name,
		Args: args,
		Body: node.Children[bodyStart:],
	}

	if interpreterState.Functions[name].Name != "" {
		panic("Function " + name + " already declared")
	}

	interpreterState.Functions[name] = function
}

func interpretMemberAssignment(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	name := node.Children[0].Value
	index := resolveExpression(interpreterState, node.Children[1])
	value := resolveExpression(interpreterState, node.Children[2])

	variable := interpreterState.LocalScope.Variables[name]
	if variable != nil {
		if variable.Indecies == nil {
			variable.Indecies = map[string]*Variable{}
		}

		variable.Indecies[index.Value] = &Variable{
			Type:     value.Type,
			Value:    value.Value,
			Indecies: value.Indecies,
		}

		interpreterState.LocalScope.Variables[name] = variable

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	}

	variable = interpreterState.GlobalScope.Variables[name]
	if variable != nil {
		if variable.Indecies == nil {
			variable.Indecies = map[string]*Variable{}
		}

		variable.Indecies[index.Value] = &Variable{
			Type:     value.Type,
			Value:    value.Value,
			Indecies: value.Indecies,
		}

		interpreterState.GlobalScope.Variables[name] = variable

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	}

	return Variable{
		Type:  genalphatypes.ASTNodeTypeNone,
		Value: "",
	}
}

func resolveExpression(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	if node.Type == genalphatypes.ASTNodeTypeIdentifier {
		return resolveIdentifier(interpreterState, node)
	}

	if node.Type == genalphatypes.ASTNodeTypeNumber {
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: node.Value,
		}
	}

	if node.Type == genalphatypes.ASTNodeTypeString {
		return Variable{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: node.Value,
		}
	}

	if node.Type == genalphatypes.ASTNodeTypeBoolean {
		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: node.Value,
		}
	}

	if node.Type == genalphatypes.ASTNodeTypeBlock {
		return interpretBlock(interpreterState, node)
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

	if node.Type == genalphatypes.ASTNodeTypeMemberAccess {
		return resolveMemberAccess(interpreterState, node)
	}

	if node.Type == genalphatypes.ASTNodeTypeExpression {
		// empty node
		if len(node.Children) == 0 {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}

		return resolveExpression(interpreterState, node.Children[0])
	}

	if node.Type == genalphatypes.ASTNodeTypeNone {
		return Variable{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: "",
		}
	}

	panic("Invalid expression node type " + fmt.Sprint(node.Type))
}

func resolveMemberAccess(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	name := node.Children[0].Value

	variable := interpreterState.LocalScope.Variables[name]
	if variable == nil {
		variable = interpreterState.GlobalScope.Variables[name]
	}
	if variable == nil {
		panic("Variable " + name + " not found")
	}

	index := resolveExpression(interpreterState, node.Children[1])
	value := variable.Indecies[index.Value]
	if value == nil {
		value = variable.Indecies[fmt.Sprint(utils.ParseNumber(index.Value))]
		if value == nil {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}
	}

	return *value
}

func resolveIdentifier(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	name := node.Value

	variable := interpreterState.LocalScope.Variables[name]
	if variable == nil {
		variable = interpreterState.GlobalScope.Variables[name]
	}
	if variable == nil {
		panic("Variable " + name + " not found")
	}

	return *variable
}

func resolveBinaryOperation(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	left := resolveExpression(interpreterState, node.Children[0])
	right := resolveExpression(interpreterState, node.Children[1])

	switch node.Value {
	case "+":
		switch left.Type {
		// do we keep it like javascript bullshit or check if both are the same
		case genalphatypes.ASTNodeTypeNumber:
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNumber,
				Value: fmt.Sprint(utils.ParseNumber(left.Value) + utils.ParseNumber(right.Value)),
			}
		case genalphatypes.ASTNodeTypeString:
			return Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: left.Value + right.Value,
			}
		default:
			return Variable{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: left.Value + right.Value,
			}
		}

	case "-":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation -")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprint(utils.ParseNumber(left.Value) - utils.ParseNumber(right.Value)),
		}
	case "*":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation *")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprint(utils.ParseNumber(left.Value) * utils.ParseNumber(right.Value)),
		}
	case "**":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation **")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprint(math.Pow(utils.ParseNumber(left.Value), utils.ParseNumber(right.Value))),
		}
	case "/":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation /")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprint(utils.ParseNumber(left.Value) / utils.ParseNumber(right.Value)),
		}
	case "%":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation %")
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: fmt.Sprint(int(utils.ParseNumber(left.Value)) % int(utils.ParseNumber(right.Value))),
		}
	case "==":
		value := string(genalphatypes.KeywordFalse)
		if left.Value == right.Value {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "===":
		if left.Type != right.Type {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordFalse),
			}
		}

		value := string(genalphatypes.KeywordFalse)
		if left.Value == right.Value {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "!=":
		value := string(genalphatypes.KeywordFalse)
		if left.Value != right.Value {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "!==":
		if left.Type != right.Type {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeBoolean,
				Value: string(genalphatypes.KeywordTrue),
			}
		}

		value := string(genalphatypes.KeywordFalse)
		if left.Value != right.Value {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case ">":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation >")
		}

		value := string(genalphatypes.KeywordFalse)
		if utils.ParseNumber(left.Value) > utils.ParseNumber(right.Value) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "<":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation <")
		}

		value := string(genalphatypes.KeywordFalse)
		if utils.ParseNumber(left.Value) < utils.ParseNumber(right.Value) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case ">=":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation <")
		}

		value := string(genalphatypes.KeywordFalse)
		if utils.ParseNumber(left.Value) >= utils.ParseNumber(right.Value) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "<=":
		if left.Type != genalphatypes.ASTNodeTypeNumber {
			panic("Invalid operand type for binary operation <")
		}

		value := string(genalphatypes.KeywordFalse)
		if utils.ParseNumber(left.Value) <= utils.ParseNumber(right.Value) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "&&":
		if left.Type != genalphatypes.ASTNodeTypeBoolean {
			panic("Invalid operand type for binary operation &&")
		}

		value := string(genalphatypes.KeywordFalse)
		if left.Value == string(genalphatypes.KeywordTrue) && right.Value == string(genalphatypes.KeywordTrue) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	case "||":
		if left.Type != genalphatypes.ASTNodeTypeBoolean {
			panic("Invalid operand type for binary operation ||")
		}

		value := string(genalphatypes.KeywordFalse)
		if left.Value == string(genalphatypes.KeywordTrue) || right.Value == string(genalphatypes.KeywordTrue) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	default:
		panic("Invalid binary operation " + node.Value)
	}
}

func resolveUnaryOperation(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	operand := resolveExpression(interpreterState, node.Children[0])

	switch node.Value {
	case "!":
		if operand.Type != genalphatypes.ASTNodeTypeBoolean {
			panic("Invalid operand type for unary operation !")
		}

		value := string(genalphatypes.KeywordFalse)
		if operand.Value == string(genalphatypes.KeywordFalse) {
			value = string(genalphatypes.KeywordTrue)
		}

		return Variable{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: value,
		}
	default:
		panic("Invalid unary operation " + node.Value)
	}
}

func resolveStdFunctionCall(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	name := node.Children[0].Value

	stdFunction := STDFunctions[name] // from std.go
	if stdFunction == nil {
		panic("Function " + name + " not found")
	}

	args := []Variable{}
	for _, argNode := range node.Children[1:] {
		arg := resolveExpression(interpreterState, argNode)
		args = append(args, arg)
	}

	return stdFunction(args)
}

func resolveFunctionCall(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	name := node.Children[0].Value

	function := interpreterState.Functions[name]
	if function.Name == "" {
		return resolveStdFunctionCall(interpreterState, node)
	}

	// todo check if actualy correct?
	if len(function.Args) > len(node.Children)-1 {
		panic("Invalid number of arguments for function " + name)
	}

	scope := Scope{
		Variables: map[string]*Variable{},
	}

	for i, arg := range function.Args {
		argValue := resolveExpression(interpreterState, node.Children[i+1])
		scope.Variables[arg.Value] = &argValue
	}

	newScope(interpreterState, scope)
	for _, instructionNode := range function.Body {
		variable := interpretNode(interpreterState, instructionNode, "")
		if variable.Type != genalphatypes.ASTNodeTypeNone {
			popScope(interpreterState)
			return variable
		}
	}
	popScope(interpreterState)

	return Variable{
		Type:  genalphatypes.ASTNodeTypeNone,
		Value: "",
	}
}

// only case is when we have brackets in an expression
func interpretBlock(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	return resolveExpression(interpreterState, node.Children[0])
}

func interpretIf(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	condition := resolveExpression(interpreterState, node.Children[0])
	if condition.Type != genalphatypes.ASTNodeTypeBoolean {
		panic("Invalid condition type for if statement, got: " + condition.Value)
	}

	if condition.Value == string(genalphatypes.KeywordTrue) {
		for _, instructionNode := range node.Children[1:] {
			variable := interpretNode(interpreterState, instructionNode, "")
			if variable.Type != genalphatypes.ASTNodeTypeNone {
				return variable
			}
		}
	}

	return Variable{
		Type:  genalphatypes.ASTNodeTypeNone,
		Value: "",
	}
}

func interpretWhile(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	for {
		condition := resolveExpression(interpreterState, node.Children[0])
		if condition.Type != genalphatypes.ASTNodeTypeBoolean {
			panic("Invalid condition type for while statement")
		}

		if condition.Value == string(genalphatypes.KeywordFalse) {
			break
		}

		for _, instructionNode := range node.Children[1:] {
			variable := interpretNode(interpreterState, instructionNode, "")
			if variable.Type != genalphatypes.ASTNodeTypeNone {
				return variable
			}
		}
	}

	return Variable{
		Type:  genalphatypes.ASTNodeTypeNone,
		Value: "",
	}
}

func interpretReturn(interpreterState *InterpreterState, node genalphatypes.ASTNode) Variable {
	if len(node.Children) != 1 {
		panic("return should be done with one argument, the value to return, such as 'rizzult \"returned value\"'")
	}

	return resolveExpression(interpreterState, node.Children[0])
}

func interpretImport(interpreterState *InterpreterState, node genalphatypes.ASTNode, parentFilename string) Variable {
	if len(node.Children) != 1 {
		panic("import should be done with one argument, the file to import, such as 'gyat \"test.gal\"'")
	}

	filename := node.Children[0].Value
	newFilename := parentFilename[:strings.LastIndex(parentFilename, "/")]

	importedFilename := newFilename + "/" + filename
	if !strings.HasSuffix(filename, ".gal") {
		// this means we are importing a package
		// check in the parent directory
		// remove filename from parentFilename

		importedFilename = newFilename + "/" + filename + "/__.gal"
		if !utils.FileExists(importedFilename) {
			// check the installed packages directory
			importedFilename = utils.GetInstalledPackagesDirectory() + filename + "/__.gal"
			if !utils.FileExists(importedFilename) {
				fmt.Println(importedFilename)
				panic("Package " + filename + " not found" + " in " + newFilename)
			}
		}
	}

	isString := node.Children[0].Type == genalphatypes.ASTNodeTypeString
	if !isString {
		panic("import should be done with a string argument, the file to import, such as 'gyat \"test.gal\"'")
	}

	for _, importedFile := range interpreterState.ImportedFiles {
		if importedFile == importedFilename {
			return Variable{
				Type:  genalphatypes.ASTNodeTypeNone,
				Value: "",
			}
		}
	}

	interpreterState.ImportedFiles = append(interpreterState.ImportedFiles, importedFilename)
	ast := loadAST(importedFilename)
	for _, child := range ast.Children {
		interpretNode(interpreterState, child, importedFilename)
	}

	return Variable{
		Type:  genalphatypes.ASTNodeTypeNone,
		Value: "",
	}
}

func interpretVariableDeclaration(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
	name := node.Children[0].Value
	value := resolveExpression(interpreterState, node.Children[1])

	if strings.HasPrefix(name, "GLOBAL_") {
		interpreterState.GlobalScope.Variables[name] = &value
		return
	}

	interpreterState.LocalScope.Variables[name] = &value
}

func interpretVariableAssignment(interpreterState *InterpreterState, node genalphatypes.ASTNode) {
	name := node.Children[0].Value
	value := resolveExpression(interpreterState, node.Children[1])

	variable := interpreterState.LocalScope.Variables[name]
	if variable != nil {
		originalIndecies := variable.Indecies
		for key, val := range value.Indecies {
			if val != nil {
				originalIndecies[key] = val
			}
		}

		value.Indecies = originalIndecies
		interpreterState.LocalScope.Variables[name] = &value
		return
	}

	variable = interpreterState.GlobalScope.Variables[name]
	if variable != nil {

		originalIndecies := variable.Indecies
		for key, val := range value.Indecies {
			if val != nil {
				originalIndecies[key] = val
			}
		}

		value.Indecies = originalIndecies
		interpreterState.GlobalScope.Variables[name] = &value
		return
	}

	panic("Variable " + name + " not found")
}

// returns the sha256 hash of the given ast
func sha256Hash(content string) string {
	shaBytes := sha256.Sum256([]byte(content))
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
		panic(err)
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
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	scanner.Scan()
	return scanner.Text()
}

// load and save are for when importing new files so we dont have to lex and parse them again (optimization and easier to work with)
func loadAST(filename string) genalphatypes.ASTNode {
	firstLine := readFirstLineFile(filename + "+")
	if firstLine == "" {
		// parse here and then save the ast
		contents := utils.ReadContents(filename)
		tokens := lexer.Lex(contents)
		ast := parser.Parse(tokens)
		// saveAST(ast, filename, filename+"+")
		return ast
	}

	node := readNotFirstLineFile(filename + "+")
	node = strings.ReplaceAll(node, "\\n", "\n")

	if len(node) == 0 {
		panic("Error reading ast from file")
	}

	return getNodeFromBytes([]byte(node))
}

func saveAST(ast genalphatypes.ASTNode, sourceFilename string, filename string) {
	// format:
	// - sha256 hash of the ast
	// - the rest of the ast in gob format

	// first we need to get the sha256 hash of the ast
	contents := utils.ReadContents(sourceFilename)
	sha := sha256Hash(contents)
	if readFirstLineFile(filename) == sha {
		return
	}

	// replace any new lines with \\n
	bytes := string(getNodeBytes(ast))
	bytes = strings.ReplaceAll(bytes, "\n", "\\n")

	writeToFile(filename, sha+"\n"+bytes)
}
