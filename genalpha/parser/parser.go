package parser

import (
	"fmt"
	"slices"

	genalphatypes "bobik.squidwock.com/root/genalphalang/genalpha"
	"bobik.squidwock.com/root/genalphalang/genalpha/lexer"
)

type ProgramState int

const (
	ProgramStateNormal ProgramState = iota
	ProgramStateFunctionDeclaration
	ProgramStateVariableDeclaration
	ProgramStateVariableAssignment
	ProgramStateFunctionCall
	ProgramStateIf
	ProgramStateWhile
	ProgramStateReturn
	ProgramStateImport
)

type ParserState struct {
	ProgramState        ProgramState
	Line                int
	TokenIndex          int
	OpenBrackets        int
	OpenCurly           int
	DeclarationCount    int // loops, if, function
	IsArgList           bool
	IsFuncBlock         bool
	ASTNodeFunc         genalphatypes.ASTNode // for constructing function declaration
	ASTNodeCall         genalphatypes.ASTNode // for constructing function call
	ASTNodeExpr         genalphatypes.ASTNode // for constructing expressions
	ASTRoot             genalphatypes.ASTNode
	ASTNodeDecl         genalphatypes.ASTNode
	ASTNodeAssign       genalphatypes.ASTNode
	ASTNodeReturn       genalphatypes.ASTNode
	ASTNodeImport       genalphatypes.ASTNode
	ASTNodeMemberAccess genalphatypes.ASTNode
	ASTNodeNamespace    genalphatypes.ASTNode
	ASTNodeParent       *genalphatypes.ASTNode   // for nested blocks
	ASTNodeStack        []*genalphatypes.ASTNode // for nested blocks too
}

// just a huge state machine
func Parse(tokens []lexer.Token) genalphatypes.ASTNode {
	var parserState = ParserState{
		ProgramState: ProgramStateNormal,
		ASTRoot: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeProgram,
		},
		ASTNodeFunc: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeFunctionDeclaration,
		},
		ASTNodeCall: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeFunctionCall,
		},
		ASTNodeExpr: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeExpression,
		},
		ASTNodeDecl: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeVariableDeclaration,
		},
		ASTNodeAssign: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeVariableAssignment,
		},
		ASTNodeReturn: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeReturn,
		},
		ASTNodeImport: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeImport,
		},
		ASTNodeMemberAccess: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeMemberAccess,
		},
		ASTNodeNamespace: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeIdentifier,
		},
	}

	for _, token := range tokens {
		if token.Type == lexer.TokenTypeWhitespace || token.Type == lexer.TokenTypeComment {
			continue
		}

		if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordIf) {
			parserState.DeclarationCount++
		}
		if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordWhile) {
			parserState.DeclarationCount++
		}
		if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordFunc) {
			parserState.DeclarationCount++
		}
		if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordEnd) {
			parserState.DeclarationCount--
		}
		if token.Type == lexer.TokenTypePunctuation && token.Value == "(" {
			parserState.OpenBrackets++
		}
		if token.Type == lexer.TokenTypePunctuation && token.Value == ")" {
			parserState.OpenBrackets--
		}
		if token.Type == lexer.TokenTypePunctuation && token.Value == "{" {
			parserState.OpenCurly++
		}
		if token.Type == lexer.TokenTypePunctuation && token.Value == "}" {
			parserState.OpenCurly--
		}

		if token.Type == lexer.TokenTypeNewline {
			parserState.Line++

			if parserState.OpenBrackets != 0 {
				panic("PARSER: Mismatched brackets") // todo where
			}
			if parserState.OpenCurly != 0 {
				panic("PARSER: Mismatched curly brackets") // todo where
			}
		}

		if parseFunctionDeclarationLogic(&parserState, token) {
			continue
		}

		if parseVariableDeclaration(&parserState, token) {
			continue
		}

		if parseFunctionCall(&parserState, token) {
			continue
		}

		if parseIfWhile(&parserState, token) {
			continue
		}

		if parseReturn(&parserState, token) {
			continue
		}

		if parseImport(&parserState, token) {
			continue
		}

		if parseAssignment(&parserState, token) {
			continue
		}

		parserState.TokenIndex++
	}

	PrintAST(parserState.ASTRoot)
	if parserState.DeclarationCount != 0 {
		panic("PARSER: Mismatched declarations, meaning you are missing end somewhere")
	}

	return parserState.ASTRoot
}

func parseAssignment(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeIdentifier && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateVariableAssignment
		parserState.ASTNodeExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNoteTypeExpression,
		}
		parserState.ASTNodeAssign = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeVariableAssignment,
		}
		parserState.ASTNodeAssign.Children = append(parserState.ASTNodeAssign.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeIdentifier,
			Value: token.Value,
		})
		return true
	}

	if parserState.ProgramState == ProgramStateVariableAssignment {
		if token.Type == lexer.TokenTypeOperator && token.Value == "=" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}

		if token.Type == lexer.TokenTypeNewline && parserState.IsArgList {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeAssign.Children = append(parserState.ASTNodeAssign.Children, parserState.ASTNodeExpr)

			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false
			fmt.Println(parserState.ASTNodeAssign)
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeAssign)
			return true
		}

		if parserState.IsArgList {
			parseExpression(parserState, token)
			return true
		}
	}

	return false
}

func parseImport(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordImport) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateImport
		parserState.ASTNodeImport = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeImport,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateImport {
		if token.Type == lexer.TokenTypeString {
			parserState.ASTNodeImport.Children = append(parserState.ASTNodeImport.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeString,
				Value: token.Value,
			})
			parserState.ProgramState = ProgramStateNormal
			parserState.ASTRoot.Children = append(parserState.ASTRoot.Children, parserState.ASTNodeImport)
			return true
		}
	}

	return false
}

func parseIfWhile(parserState *ParserState, token lexer.Token) bool {
	if (token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordIf) && parserState.ProgramState == ProgramStateNormal) ||
		(token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordWhile) && parserState.ProgramState == ProgramStateNormal) {
		parserState.ASTNodeStack = append(parserState.ASTNodeStack, parserState.ASTNodeParent)
		var nodeType = genalphatypes.ASTNodeTypeIf
		var programState = ProgramStateIf
		if token.Value == string(lexer.KeywordWhile) {
			nodeType = genalphatypes.ASTNodeTypeWhile
			programState = ProgramStateWhile
		}
		parserState.ASTNodeParent = &genalphatypes.ASTNode{
			Type: nodeType,
		}
		parserState.ASTNodeExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNoteTypeExpression,
		}
		parserState.ProgramState = programState
		parserState.IsArgList = true
		return true
	}

	if parserState.ProgramState == ProgramStateIf || parserState.ProgramState == ProgramStateWhile {
		if token.Type == lexer.TokenTypeNewline {
			// append expression
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeExpr)
			// append if block to parent
			parserState.ASTNodeStack[len(parserState.ASTNodeStack)-1].Children = append(parserState.ASTNodeStack[len(parserState.ASTNodeStack)-1].Children, *parserState.ASTNodeParent)
			parserState.ASTNodeParent = parserState.ASTNodeStack[len(parserState.ASTNodeStack)-1]
			parserState.ASTNodeStack = parserState.ASTNodeStack[:len(parserState.ASTNodeStack)-1]
			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false

			return true
		}

		if parserState.IsArgList {
			parseExpression(parserState, token)
			return true
		}
	}

	return false
}

func parseReturn(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordReturn) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateReturn
		parserState.ASTNodeExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNoteTypeExpression,
		}
		parserState.ASTNodeReturn = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeReturn,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateReturn {
		if token.Type == lexer.TokenTypeNewline {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeReturn.Children = append(parserState.ASTNodeReturn.Children, parserState.ASTNodeExpr)
			parserState.ProgramState = ProgramStateNormal
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeReturn)
			return true
		}

		parseExpression(parserState, token)
		return true
	}

	return false
}

func parseVariableDeclaration(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordVar) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateVariableDeclaration
		parserState.ASTNodeExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNoteTypeExpression,
		}
		parserState.ASTNodeDecl = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeVariableDeclaration,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateVariableDeclaration {
		/// fmt.Println(parserState.ASTNodeDecl)
		if token.Type == lexer.TokenTypeIdentifier && !parserState.IsArgList {
			parserState.ASTNodeDecl.Children = append(parserState.ASTNodeDecl.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			return true
		}

		if token.Type == lexer.TokenTypeOperator && token.Value == "=" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}

		if token.Type == lexer.TokenTypeNewline && parserState.IsArgList {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeDecl.Children = append(parserState.ASTNodeDecl.Children, parserState.ASTNodeExpr)

			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeDecl)
			return true
		}

		if parserState.IsArgList {
			parseExpression(parserState, token)
			return true
		}
	}

	return false
}

func parseFunctionCall(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordCall) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateFunctionCall
		parserState.ASTNodeCall = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeFunctionCall,
		}
		parserState.ASTNodeExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNoteTypeExpression,
		}
		return true
	}

	// fire funcName(args)
	if parserState.ProgramState == ProgramStateFunctionCall {
		if token.Type == lexer.TokenTypeIdentifier && !parserState.IsArgList {
			parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			return true
		}

		// before args because we want to be able to end
		if token.Type == lexer.TokenTypePunctuation && token.Value == ")" {
			// append last arg
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, parserState.ASTNodeExpr)
			parserState.IsArgList = false

			parserState.ProgramState = ProgramStateNormal
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeCall)
			return true
		}

		// args could be expressions
		if parserState.IsArgList {
			if token.Type == lexer.TokenTypePunctuation && token.Value == "," {
				if parserState.ASTNodeExpr.Type == 0 {
					panic("PARSER: Expected expression")
				}
				fixExpression(&parserState.ASTNodeExpr)
				parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, parserState.ASTNodeExpr)
				return true
			}

			parseExpression(parserState, token)

			return true
		}

		// end of args
		if token.Type == lexer.TokenTypePunctuation && token.Value == "(" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}
	}

	return false
}

// handles only function declaration: args, identity of func and end and pushes to AST root
func parseFunctionDeclarationLogic(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordFunc) && parserState.ProgramState == ProgramStateNormal {
		if parserState.DeclarationCount != 1 {
			panic("PARSER: Function declaration should be on top level")
		}

		parserState.ProgramState = ProgramStateFunctionDeclaration
		parserState.ASTNodeFunc = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeFunctionDeclaration,
		}
		parserState.ASTNodeParent = &parserState.ASTNodeFunc
		return true
	}

	if parserState.ProgramState == ProgramStateFunctionDeclaration {
		// args and function name
		if token.Type == lexer.TokenTypeIdentifier && !parserState.IsFuncBlock {
			var identType = genalphatypes.ASTNodeTypeIdentifier
			if parserState.IsArgList {
				identType = genalphatypes.ASTNodeTypeFunctionArgument
			}

			parserState.ASTNodeFunc.Children = append(parserState.ASTNodeFunc.Children, genalphatypes.ASTNode{
				Type:  identType,
				Value: token.Value,
			})

			parserState.IsArgList = true
			return true
		}

		if token.Type == lexer.TokenTypePunctuation && token.Value == "{" {
			return true
		}

		if token.Type == lexer.TokenTypePunctuation && token.Value == "}" {
			if parserState.OpenCurly == 0 {
				parserState.IsArgList = false
				parserState.IsFuncBlock = true
				parserState.ProgramState = ProgramStateNormal
			}

			return true
		}
	}

	// end
	if parserState.IsFuncBlock {
		if parserState.DeclarationCount == 0 { // only when we sure its end of function
			if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordEnd) {
				parserState.ProgramState = ProgramStateNormal
				parserState.IsArgList = false
				parserState.IsFuncBlock = false
				parserState.ASTRoot.Children = append(parserState.ASTRoot.Children, parserState.ASTNodeFunc)
				return true
			}
		}
	}

	return false
}

func parseExpression(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeIdentifier {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeIdentifier,
			Value: token.Value,
		})
		return true
	}

	if token.Type == lexer.TokenTypeNumber {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: token.Value,
		})
		return true
	}

	if token.Type == lexer.TokenTypeString {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: token.Value,
		})
		return true
	}

	if token.Value == string(lexer.KeywordTrue) || token.Value == string(lexer.KeywordFalse) {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: token.Value,
		})
		return true
	}

	if token.Type == lexer.TokenTypeOperator {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeOperator,
			Value: token.Value,
		})
		return true
	}

	if token.Type == lexer.TokenTypePunctuation && token.Value == "(" {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeBlock,
			Value: token.Value,
		})

		return true
	}

	if token.Type == lexer.TokenTypePunctuation && token.Value == ")" {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeBlock,
			Value: token.Value,
		})

		return true
	}

	return false
}

// makes it so its correct order of operations
func fixExpression(expression *genalphatypes.ASTNode) {
	makeBlocks(expression)
}

func makeBlocks(expression *genalphatypes.ASTNode) {
	var block = genalphatypes.ASTNode{
		Type: genalphatypes.ASTNodeTypeBlock,
	}

	// [ (, 1, +, (2, -, 20, ), ), +, 5 ]
	//   ^                      ^
	//   blockStart             i
	// we replace the entire block with just a single genalphatypes.ASTNode of type block
	// so it becomes
	// [ BLOCK, +, 5 ]
	// and i and blockstart are reset to 0

	stack := []int{} // stack of indexes of "("
	for i := 0; i < len(expression.Children); i++ {
		switch expression.Children[i].Value {
		case "(":
			if expression.Children[i].Type == genalphatypes.ASTNodeTypeBlock {
				stack = append(stack, i)
			}
		case ")":
			if expression.Children[i].Type != genalphatypes.ASTNodeTypeBlock {
				continue
			}
			if len(stack) == 0 {
				continue
			}
			blockStart := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			var astCopy = slices.Clone(expression.Children)
			block.Children = astCopy[blockStart+1 : i]

			expression.Children = append(expression.Children[:blockStart], block)
			expression.Children = append(expression.Children, astCopy[i+1:]...)

			makeBlocks(&block)
			block = genalphatypes.ASTNode{}
			i = blockStart // Reset i to blockStart to continue processing
		}
	}
}

func PrintAST(ast genalphatypes.ASTNode) {
	fmt.Println("AST:")
	printASTNode(ast, 0)
}

func printASTNode(ast genalphatypes.ASTNode, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
	fmt.Printf("%d: %s\n", ast.Type, ast.Value)
	for _, child := range ast.Children {
		printASTNode(child, level+1)
	}
}
