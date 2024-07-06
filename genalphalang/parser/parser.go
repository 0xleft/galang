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
	ProgramStateIf
	ProgramStateWhile
	ProgramStateReturn
)

type ParserState struct {
	ProgramState ProgramState
	Line         int
	TokenIndex   int
	OpenBrackets int
	OpenCurly    int
	IsArgList    bool
	IsFuncBlock  bool
	ASTNodeFunc  ASTNode // for constructing function declaration
	ASTRoot      ASTNode
}

// just a huge state machine
func Parse(tokens []lexer.Token) ASTNode {
	var parserState = ParserState{
		ProgramState: ProgramStateNormal,
		ASTRoot: ASTNode{
			Type: ASTNodeTypeProgram,
		},
	}

	for _, token := range tokens {
		if token.Type == lexer.TokenTypeWhitespace || token.Type == lexer.TokenTypeComment {
			continue
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

		parserState.TokenIndex++
	}

	fmt.Println(parserState.ASTRoot)
	return parserState.ASTRoot
}

func parseFunctionBlock(parserState *ParserState, token lexer.Token) bool {
	if parserState.ProgramState == ProgramStateFunctionDeclaration {
		if parserState.IsFuncBlock {
			return true
		}
	}
}

// handles only function declaration: args, identity of func and end and pushes to AST root
func parseFunctionDeclarationLogic(parserState *ParserState, token lexer.Token) bool {
	if token.Type == lexer.TokenTypeKeyword && token.Value == string(lexer.KeywordFunc) {
		parserState.ProgramState = ProgramStateFunctionDeclaration
		parserState.ASTNodeFunc = ASTNode{
			Type: ASTNodeTypeFunctionDeclaration,
		}
		return true
	}

	if parserState.ProgramState == ProgramStateFunctionDeclaration {
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

		// todo remove?
		if token.Type == lexer.TokenTypePunctuation && token.Value == "{" {
			return true
		}

		if token.Type == lexer.TokenTypePunctuation && token.Value == "}" {
			if parserState.OpenCurly == 0 {
				parserState.IsArgList = false
				parserState.IsFuncBlock = true
			}

			return true
		}

		if parserState.IsFuncBlock {
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
