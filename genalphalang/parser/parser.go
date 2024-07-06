package parser

import (
	"fmt"

	"bobik.squidwock.com/root/genalphalang/genalphalang/lexer"
)

type ASTNodeType int

const (
	ASTNodeTypeProgram ASTNodeType = iota
	ASTNodeTypeFunctionDeclaration
	ASTNodeTypeFunctionCall
	ASTNodeTypeVariableDeclaration
	ASTNodeTypeVariableAssignment
	ASTNoteTypeExpression
	ASTNodeTypeIf
	ASTNodeTypeWhile
	ASTNodeTypeReturn
	ASTNodeTypeBinaryOperation
	ASTNodeTypeUnaryOperation
	ASTNodeTypeIdentifier
	ASTNodeTypeNumber
	ASTNodeTypeString
	ASTNodeTypeBoolean
	ASTNodeTypeArray
	ASTNodeTypeFunctionArgument
	ASTNodeTypeIndex
	ASTNodeTypeMemberAccess
	ASTNodeTypeBlock
)

type ASTNode struct {
	Type     ASTNodeType
	Children []ASTNode
	Value    string
	Line     int
}

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
)

type ParserState struct {
	ProgramState     ProgramState
	Line             int
	TokenIndex       int
	OpenBrackets     int
	OpenCurly        int
	DeclarationCount int // loops, if, function
	IsArgList        bool
	IsFuncBlock      bool
	ASTNodeFunc      ASTNode // for constructing function declaration
	ASTNodeCall      ASTNode // for constructing function call
	ASTNodeExpr      ASTNode // for constructing expressions
	ASTRoot          ASTNode
	ASTNodeDecl      ASTNode
	ASTNodeAssign    ASTNode
	ASTNodeIf        ASTNode
	ASTNodeWhile     ASTNode
	ASTNodeReturn    ASTNode
}

// just a huge state machine
func Parse(tokens []lexer.Token) ASTNode {
	var parserState = ParserState{
		ProgramState: ProgramStateNormal,
		ASTRoot: ASTNode{
			Type: ASTNodeTypeProgram,
		},
		ASTNodeFunc: ASTNode{
			Type: ASTNodeTypeFunctionDeclaration,
		},
		ASTNodeCall: ASTNode{
			Type: ASTNodeTypeFunctionCall,
		},
		ASTNodeExpr: ASTNode{
			Type: ASTNoteTypeExpression,
		},
		ASTNodeDecl: ASTNode{
			Type: ASTNodeTypeVariableDeclaration,
		},
		ASTNodeAssign: ASTNode{
			Type: ASTNodeTypeVariableAssignment,
		},
		ASTNodeIf: ASTNode{
			Type: ASTNodeTypeIf,
		},
		ASTNodeWhile: ASTNode{
			Type: ASTNodeTypeWhile,
		},
		ASTNodeReturn: ASTNode{
			Type: ASTNodeTypeReturn,
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
			continue
		}

		if parseFunctionDeclarationLogic(&parserState, token) {
			continue
		}

		if parseFunctionCall(&parserState, token) {
			continue
		}

		if parseVariableDeclaration(&parserState, token) {
			continue
		}

		if parseIf(&parserState, token) {
			continue
		}

		if parseWhile(&parserState, token) {
			continue
		}

		if parseReturn(&parserState, token) {
			continue
		}

		parserState.TokenIndex++
	}

	fmt.Println(parserState.ASTRoot, parserState.DeclarationCount, parserState.OpenBrackets, parserState.OpenCurly)
	if parserState.DeclarationCount != 0 {
		panic("PARSER: Mismatched declarations, meaning you are missing end somewhere")
	}

	return parserState.ASTRoot
}

// fixes the expression so it can be evaluated
func resolveExpression(ASTNode *ASTNode) {
	// todo
}

func parseIf(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordIf) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateIf
		return true
	}

	if parserState.ProgramState == ProgramStateIf {

	}

	return false
}

func parseWhile(parserState *ParserState, token lexer.Token) bool {
	return false
}

func parseReturn(parserState *ParserState, token lexer.Token) bool {
	return false
}

func parseVariableDeclaration(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordVar) && parserState.ProgramState == ProgramStateNormal {
		parserState.ProgramState = ProgramStateVariableDeclaration
		return true
	}

	if parserState.ProgramState == ProgramStateVariableDeclaration {
		/// fmt.Println(parserState.ASTNodeDecl)
		if token.Type == lexer.TokenTypeIdentifier && !parserState.IsArgList {
			parserState.ASTNodeDecl.Children = append(parserState.ASTNodeDecl.Children, ASTNode{
				Type:  ASTNodeTypeIdentifier,
				Value: token.Value,
				Line:  parserState.Line,
			})

			return true
		}

		if token.Type == lexer.TokenTypeOperator && token.Value == "=" && !parserState.IsArgList {
			parserState.IsArgList = true
			return true
		}

		if token.Type == lexer.TokenTypePunctuation && token.Value == ";" && parserState.IsArgList {
			parserState.ASTNodeDecl.Children = append(parserState.ASTNodeDecl.Children, parserState.ASTNodeExpr)

			parserState.ProgramState = ProgramStateNormal
			parserState.IsArgList = false
			parserState.ASTNodeFunc.Children = append(parserState.ASTNodeFunc.Children, parserState.ASTNodeDecl)
			parserState.ASTNodeDecl = ASTNode{
				Type: ASTNodeTypeVariableDeclaration,
			}
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
		return true
	}

	// fire funcName(args)
	if parserState.ProgramState == ProgramStateFunctionCall {
		if token.Type == lexer.TokenTypeIdentifier && !parserState.IsArgList {
			parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, ASTNode{
				Type:  ASTNodeTypeIdentifier,
				Value: token.Value,
				Line:  parserState.Line,
			})

			return true
		}

		// before args because we want to be able to end
		if token.Type == lexer.TokenTypePunctuation && token.Value == ")" {
			// append last arg
			parserState.ASTNodeCall.Children = append(parserState.ASTNodeCall.Children, parserState.ASTNodeExpr)
			parserState.IsArgList = false

			parserState.ProgramState = ProgramStateNormal
			parserState.ASTNodeFunc.Children = append(parserState.ASTNodeFunc.Children, parserState.ASTNodeCall)
			parserState.ASTNodeCall = ASTNode{
				Type: ASTNodeTypeFunctionCall,
			}
			parserState.ASTNodeExpr = ASTNode{
				Type: ASTNoteTypeExpression,
			}
			return true
		}

		// args could be expressions
		if parserState.IsArgList {
			if token.Type == lexer.TokenTypePunctuation && token.Value == "," {
				if parserState.ASTNodeExpr.Type == 0 {
					panic("PARSER: Expected expression")
				}
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
		parserState.ProgramState = ProgramStateFunctionDeclaration
		parserState.ASTNodeFunc = ASTNode{
			Type: ASTNodeTypeFunctionDeclaration,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateFunctionDeclaration {
		// args and function name
		if token.Type == lexer.TokenTypeIdentifier && !parserState.IsFuncBlock {
			var identType = ASTNodeTypeIdentifier
			if parserState.IsArgList {
				identType = ASTNodeTypeFunctionArgument
			}

			parserState.ASTNodeFunc.Children = append(parserState.ASTNodeFunc.Children, ASTNode{
				Type:  identType,
				Value: token.Value,
				Line:  parserState.Line,
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
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeIdentifier,
			Value: token.Value,
			Line:  parserState.Line,
		})
		return true
	}

	if token.Type == lexer.TokenTypeNumber {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeNumber,
			Value: token.Value,
			Line:  parserState.Line,
		})
		return true
	}

	// fmt.Printf("token: %v\n", token)
	if token.Type == lexer.TokenTypeString {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeString,
			Value: token.Value,
			Line:  parserState.Line,
		})
		return true
	}

	if token.Value == string(lexer.KeywordTrue) || token.Value == string(lexer.KeywordFalse) {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeBoolean,
			Value: token.Value,
			Line:  parserState.Line,
		})
		return true
	}

	if token.Type == lexer.TokenTypeOperator {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeBinaryOperation,
			Value: token.Value,
			Line:  parserState.Line,
		})
		return true
	}

	if token.Type == lexer.TokenTypePunctuation && token.Value == "(" {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeBlock,
			Value: token.Value,
			Line:  parserState.Line,
		})

		return true
	}

	if token.Type == lexer.TokenTypePunctuation && token.Value == ")" {
		parserState.ASTNodeExpr.Children = append(parserState.ASTNodeExpr.Children, ASTNode{
			Type:  ASTNodeTypeBlock,
			Value: token.Value,
			Line:  parserState.Line,
		})

		return true
	}

	return false
}
