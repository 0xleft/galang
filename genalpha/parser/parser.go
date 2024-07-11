package parser

import (
	"fmt"
	"slices"

	genalphatypes "bobik.squidwock.com/root/genalphalang/genalpha"
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
	ProgramStateMemberAccess
	ProgramStateMemberAssign
	ProgramStateFunctionCallExpression
)

type ParserState struct {
	ProgramState            ProgramState
	PreviousState           ProgramState
	Line                    int
	TokenIndex              int
	OpenBrackets            int
	OpenCurly               int
	DeclarationCount        int  // loops, if, function
	IsMemberAccessExpr      bool // only used for member access expression
	IsCallExpr              bool
	IsArgList               bool
	IsFuncBlock             bool
	ASTNodeFunc             genalphatypes.ASTNode // for constructing function declaration
	ASTNodeCall             genalphatypes.ASTNode // for constructing function call
	ASTNodeCallExpr         genalphatypes.ASTNode // for constructing function call in expression
	ASTNodeCallTempExpr     genalphatypes.ASTNode // used for expressions in function call args
	ASTNodeExpr             genalphatypes.ASTNode // for constructing expressions
	ASTRoot                 genalphatypes.ASTNode
	ASTNodeDecl             genalphatypes.ASTNode
	ASTNodeAssign           genalphatypes.ASTNode
	ASTNodeReturn           genalphatypes.ASTNode
	ASTNodeImport           genalphatypes.ASTNode
	ASTNodeMemberAccess     genalphatypes.ASTNode
	ASTNodeMemberAccessExpr genalphatypes.ASTNode
	ASTNodeMemberAssign     genalphatypes.ASTNode
	ASTNodeParent           *genalphatypes.ASTNode   // for nested blocks
	ASTNodeStack            []*genalphatypes.ASTNode // for nested blocks too
}

// just a huge state machine
func Parse(tokens []genalphatypes.Token) genalphatypes.ASTNode {
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
		ASTNodeMemberAssign: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeMemberAssignment,
		},
		ASTNodeCallTempExpr: genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeExpression,
		},
	}

	for _, token := range tokens {
		if token.Type == genalphatypes.TokenTypeWhitespace || token.Type == genalphatypes.TokenTypeComment {
			continue
		}

		if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordIf) {
			parserState.DeclarationCount++
		}
		if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordWhile) {
			parserState.DeclarationCount++
		}
		if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordFunc) {
			parserState.DeclarationCount++
		}
		if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordEnd) {
			parserState.DeclarationCount--
		}
		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "(" {
			parserState.OpenBrackets++
		}
		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == ")" {
			parserState.OpenBrackets--
		}
		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "{" {
			parserState.OpenCurly++
		}
		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "}" {
			parserState.OpenCurly--
		}

		if token.Type == genalphatypes.TokenTypeNewline {
			parserState.Line++

			if parserState.OpenBrackets != 0 {
				panic("PARSER: Mismatched brackets") // todo where
			}
			if parserState.OpenCurly != 0 {
				panic("PARSER: Mismatched curly brackets") // todo where
			}
		}

		if parserState.IsArgList {
			if parseMemberAccess(&parserState, token) {
				continue
			}
			if parseFunctionCallExpression(&parserState, token) {
				continue
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

		if parseMemberAssignment(&parserState, token) {
			continue
		}

		parserState.TokenIndex++
	}

	if parserState.DeclarationCount != 0 {
		panic("PARSER: Mismatched declarations, meaning you are missing end somewhere")
	}

	return parserState.ASTRoot
}

// todo
func resetExpression(parserState *ParserState) {
	parserState.ASTNodeExpr = genalphatypes.ASTNode{
		Type: genalphatypes.ASTNodeTypeExpression,
	}
}

func parseMemberAccess(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "[" {
		parserState.PreviousState = parserState.ProgramState
		parserState.ProgramState = ProgramStateMemberAccess
		parserState.ASTNodeMemberAccess = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeMemberAccess,
		}
		parserState.ASTNodeMemberAccessExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeExpression,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateMemberAccess {
		if token.Type == genalphatypes.TokenTypeIdentifier && !parserState.IsMemberAccessExpr {
			parserState.ASTNodeMemberAccess.Children = append(parserState.ASTNodeMemberAccess.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			parserState.IsMemberAccessExpr = true

			return true
		}

		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "]" {
			fixExpression(&parserState.ASTNodeMemberAccessExpr)
			parserState.ProgramState = parserState.PreviousState
			parserState.ASTNodeMemberAccess.Children = append(parserState.ASTNodeMemberAccess.Children, parserState.ASTNodeMemberAccessExpr)
			parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, parserState.ASTNodeMemberAccess)

			parserState.IsMemberAccessExpr = false

			return false
		}

		if parserState.IsArgList {
			parseExpression(parserState, token, &parserState.ASTNodeMemberAccessExpr)
			return true
		}
	}

	return false
}

func parseMemberAssignment(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "[" && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateMemberAssign
		parserState.ASTNodeMemberAssign = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeMemberAssignment,
		}
		resetExpression(parserState)
		return true
	}

	if parserState.ProgramState == ProgramStateMemberAssign {
		if token.Type == genalphatypes.TokenTypeIdentifier && !parserState.IsMemberAccessExpr && !parserState.IsArgList {
			parserState.ASTNodeMemberAssign.Children = append(parserState.ASTNodeMemberAssign.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			parserState.IsMemberAccessExpr = true
			return true
		}

		if token.Type == genalphatypes.TokenTypeOperator && token.Value == "=" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}

		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "]" {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeMemberAssign.Children = append(parserState.ASTNodeMemberAssign.Children, parserState.ASTNodeExpr)

			resetExpression(parserState)

			parserState.IsMemberAccessExpr = false

			return true
		}

		if token.Type == genalphatypes.TokenTypeNewline && parserState.IsArgList {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeMemberAssign.Children = append(parserState.ASTNodeMemberAssign.Children, parserState.ASTNodeExpr)

			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeMemberAssign)
			return true
		}

		if parserState.IsArgList || parserState.IsMemberAccessExpr {
			parseExpression(parserState, token, nil)
			return true
		}
	}

	return false
}

func parseAssignment(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeIdentifier && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateVariableAssignment
		resetExpression(parserState)
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
		if token.Type == genalphatypes.TokenTypeOperator && token.Value == "=" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}

		if token.Type == genalphatypes.TokenTypeNewline && parserState.IsArgList {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeAssign.Children = append(parserState.ASTNodeAssign.Children, parserState.ASTNodeExpr)

			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeAssign)
			return true
		}

		if parserState.IsArgList {
			parseExpression(parserState, token, nil)
			return true
		}
	}

	return false
}

func parseImport(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordImport) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateImport
		parserState.ASTNodeImport = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeImport,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateImport {
		if token.Type == genalphatypes.TokenTypeString {
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

func parseIfWhile(parserState *ParserState, token genalphatypes.Token) bool {
	if (token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordIf) && parserState.ProgramState == ProgramStateNormal) ||
		(token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordWhile) && parserState.ProgramState == ProgramStateNormal) {
		parserState.ASTNodeStack = append(parserState.ASTNodeStack, parserState.ASTNodeParent)
		var nodeType = genalphatypes.ASTNodeTypeIf
		var programState = ProgramStateIf
		if token.Value == string(genalphatypes.KeywordWhile) {
			nodeType = genalphatypes.ASTNodeTypeWhile
			programState = ProgramStateWhile
		}
		parserState.ASTNodeParent = &genalphatypes.ASTNode{
			Type: nodeType,
		}
		resetExpression(parserState)
		parserState.ProgramState = programState
		parserState.IsArgList = true
		return true
	}

	if parserState.ProgramState == ProgramStateIf || parserState.ProgramState == ProgramStateWhile {
		if token.Type == genalphatypes.TokenTypeNewline {
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
			parseExpression(parserState, token, nil)
			return true
		}
	}

	return false
}

func parseReturn(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordReturn) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateReturn
		resetExpression(parserState)
		parserState.ASTNodeReturn = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeReturn,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateReturn {
		if token.Type == genalphatypes.TokenTypeNewline {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeReturn.Children = append(parserState.ASTNodeReturn.Children, parserState.ASTNodeExpr)
			parserState.ProgramState = ProgramStateNormal
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeReturn)
			return true
		}

		parseExpression(parserState, token, nil)
		return true
	}

	return false
}

func parseVariableDeclaration(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordVar) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateVariableDeclaration
		resetExpression(parserState)
		parserState.ASTNodeDecl = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeVariableDeclaration,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateVariableDeclaration {
		if token.Type == genalphatypes.TokenTypeIdentifier && !parserState.IsArgList {
			parserState.ASTNodeDecl.Children = append(parserState.ASTNodeDecl.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			return true
		}

		if token.Type == genalphatypes.TokenTypeOperator && token.Value == "=" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}

		if token.Type == genalphatypes.TokenTypeNewline && parserState.IsArgList {
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeDecl.Children = append(parserState.ASTNodeDecl.Children, parserState.ASTNodeExpr)

			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false
			parserState.ASTNodeParent.Children = append(parserState.ASTNodeParent.Children, parserState.ASTNodeDecl)
			return true
		}

		if parserState.IsArgList {
			parseExpression(parserState, token, nil)
			return true
		}
	}

	return false
}

func parseFunctionCallExpression(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordCall) {
		parserState.PreviousState = parserState.ProgramState
		parserState.ProgramState = ProgramStateFunctionCallExpression
		parserState.ASTNodeCallExpr = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeFunctionCall,
		}

		return true
	}

	if parserState.ProgramState == ProgramStateFunctionCallExpression {
		if token.Type == genalphatypes.TokenTypeIdentifier && !parserState.IsCallExpr {
			parserState.ASTNodeCallExpr.Children = append(parserState.ASTNodeCallExpr.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			return true
		}

		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "(" && !parserState.IsCallExpr {
			parserState.IsCallExpr = true
			return true
		}

		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == ")" {
			// append last arg
			fixExpression(&parserState.ASTNodeExpr)
			parserState.ASTNodeCallExpr.Children = append(parserState.ASTNodeCallExpr.Children, parserState.ASTNodeCallTempExpr)
			parserState.IsCallExpr = false

			parserState.ProgramState = parserState.PreviousState
			parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, parserState.ASTNodeCallExpr)
			return true
		}

		if parserState.IsCallExpr {
			if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "," {
				if parserState.ASTNodeExpr.Type == 0 {
					panic("PARSER: Expected expression")
				}
				fixExpression(&parserState.ASTNodeExpr)
				parserState.ASTNodeCallExpr.Children = append(parserState.ASTNodeCallExpr.Children, parserState.ASTNodeCallTempExpr)

				parserState.ASTNodeCallTempExpr = genalphatypes.ASTNode{
					Type: genalphatypes.ASTNodeTypeExpression,
				}

				return true
			}

			parseExpression(parserState, token, &parserState.ASTNodeCallTempExpr)
			return true
		}
	}

	return false
}

func parseFunctionCall(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordCall) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateFunctionCall
		parserState.ASTNodeCall = genalphatypes.ASTNode{
			Type: genalphatypes.ASTNodeTypeFunctionCall,
		}
		resetExpression(parserState)
		return true
	}

	// fire funcName(args)
	if parserState.ProgramState == ProgramStateFunctionCall {
		if token.Type == genalphatypes.TokenTypeIdentifier && !parserState.IsArgList {
			parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, genalphatypes.ASTNode{
				Type:  genalphatypes.ASTNodeTypeIdentifier,
				Value: token.Value,
			})

			return true
		}

		// before args because we want to be able to end
		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == ")" {
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
			if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "," {
				if parserState.ASTNodeExpr.Type == 0 {
					panic("PARSER: Expected expression")
				}
				fixExpression(&parserState.ASTNodeExpr)
				parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, parserState.ASTNodeExpr)
				return true
			}

			parseExpression(parserState, token, nil)

			return true
		}

		// end of args
		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "(" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}
	}

	return false
}

// handles only function declaration: args, identity of func and end and pushes to AST root
func parseFunctionDeclarationLogic(parserState *ParserState, token genalphatypes.Token) bool {
	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordFunc) && parserState.ProgramState == ProgramStateNormal {
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
		if token.Type == genalphatypes.TokenTypeIdentifier && !parserState.IsFuncBlock {
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

		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "{" {
			return true
		}

		if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "}" {
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
			if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordEnd) {
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

func parseExpression(parserState *ParserState, token genalphatypes.Token, expression *genalphatypes.ASTNode) bool {
	var exprNode = &parserState.ASTNodeExpr
	if expression != nil {
		exprNode = expression
	}

	if token.Type == genalphatypes.TokenTypeIdentifier {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeIdentifier,
			Value: token.Value,
		})
		return true
	}

	if token.Type == genalphatypes.TokenTypeNumber {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeNumber,
			Value: token.Value,
		})
		return true
	}

	if token.Type == genalphatypes.TokenTypeKeyword && token.Value == string(genalphatypes.KeywordNone) {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeNone,
			Value: token.Value,
		})
		return true
	}

	if token.Type == genalphatypes.TokenTypeString {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeString,
			Value: token.Value,
		})
		return true
	}

	if token.Value == string(genalphatypes.KeywordTrue) || token.Value == string(genalphatypes.KeywordFalse) {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeBoolean,
			Value: token.Value,
		})
		return true
	}

	if token.Type == genalphatypes.TokenTypeOperator {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeOperator,
			Value: token.Value,
		})
		return true
	}

	if token.Type == genalphatypes.TokenTypePunctuation && token.Value == "(" {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeBlock,
			Value: token.Value,
		})

		return true
	}

	if token.Type == genalphatypes.TokenTypePunctuation && token.Value == ")" {
		exprNode.Children = append(exprNode.Children, genalphatypes.ASTNode{
			Type:  genalphatypes.ASTNodeTypeBlock,
			Value: token.Value,
		})

		return true
	}

	return false
}

// makes it so its correct order of operations
func fixExpression(expression *genalphatypes.ASTNode) {
	fixOperators(expression)
	makeBlocks(expression)
	orderOperations(expression)
}

// basicaly just merge operator if there is 2 in a row
func fixOperators(expression *genalphatypes.ASTNode) {
	var i int
	for {
		if i >= len(expression.Children) {
			break
		}

		if expression.Children[i].Type == genalphatypes.ASTNodeTypeOperator {
			if i == 0 || i == len(expression.Children)-1 {
				panic("PARSER: Operator at start or end of expression")
			}

			if expression.Children[i].Type == genalphatypes.ASTNodeTypeOperator && expression.Children[i+1].Type == genalphatypes.ASTNodeTypeOperator {
				expression.Children[i].Value += expression.Children[i+1].Value
				expression.Children = append(expression.Children[:i+1], expression.Children[i+2:]...)
				i = 0
				continue
			}

		}

		i++
	}
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
			block = genalphatypes.ASTNode{
				Type: genalphatypes.ASTNodeTypeBlock,
			}
			i = blockStart // Reset i to blockStart to continue processing
		}
	}
}

func binarySplit(expression *genalphatypes.ASTNode, where int) {
	var childrenTemp = slices.Clone(expression.Children)

	var binaryBlock = genalphatypes.ASTNode{
		Type:  genalphatypes.ASTNodeTypeBinaryOperation,
		Value: expression.Children[where].Value,
		Children: []genalphatypes.ASTNode{
			{
				Type:     genalphatypes.ASTNodeTypeBlock,
				Children: childrenTemp[:where],
			},
			{
				Type:     genalphatypes.ASTNodeTypeBlock,
				Children: childrenTemp[where+1:],
			},
		},
	}

	expression.Children = []genalphatypes.ASTNode{binaryBlock}

	for _, child := range expression.Children {
		orderOperations(&child)
	}
}

func orderOperations(expression *genalphatypes.ASTNode) {
	var i int

	var operationType = 0

	for {
		if i >= len(expression.Children) {
			if operationType >= 6 {
				break
			}

			i = 0
			operationType++

			if i >= len(expression.Children) {
				break
			}
		}

		if expression.Children[i].Type == genalphatypes.ASTNodeTypeBlock {
			orderOperations(&expression.Children[i])
		}

		if expression.Children[i].Type == genalphatypes.ASTNodeTypeOperator {
			if i == 0 || i == len(expression.Children)-1 {
				panic("PARSER: Operator at start or end of expression")
			}

			switch operationType {
			case 0:
				if expression.Children[i].Value == "&&" {
					binarySplit(expression, i)
					i = 0
				}
			case 1:
				if expression.Children[i].Value == "||" {
					binarySplit(expression, i)
					i = 0
				}
			case 2:
				if expression.Children[i].Value == "==" || expression.Children[i].Value == "!=" || expression.Children[i].Value == "<" || expression.Children[i].Value == ">" || expression.Children[i].Value == "<=" || expression.Children[i].Value == ">=" {
					binarySplit(expression, i)
					i = 0
				}
			case 3:
				if expression.Children[i].Value == "+" || expression.Children[i].Value == "-" {
					binarySplit(expression, i)
					i = 0
				}
			case 4:
				if expression.Children[i].Value == "*" || expression.Children[i].Value == "/" || expression.Children[i].Value == "%" || expression.Children[i].Value == "^" || expression.Children[i].Value == "&" || expression.Children[i].Value == "|" {
					binarySplit(expression, i)
					i = 0
				}
			case 5:
				if expression.Children[i].Value == "!" {
					var unaryBlock = genalphatypes.ASTNode{
						Type:  genalphatypes.ASTNodeTypeUnaryOperation,
						Value: expression.Children[i].Value,
						Children: []genalphatypes.ASTNode{
							{
								Type:     genalphatypes.ASTNodeTypeBlock,
								Children: expression.Children[i+1:],
							},
						},
					}

					expression.Children = []genalphatypes.ASTNode{unaryBlock}

					for _, child := range expression.Children {
						orderOperations(&child)
					}
				}
			}
		}
		i++
	}
}

func PrintAST(ast genalphatypes.ASTNode, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
	fmt.Printf("%d: %s\n", ast.Type, ast.Value)
	for _, child := range ast.Children {
		PrintAST(child, level+1)
	}
}
