package genalphatypes

type ASTNodeType int

const (
	ASTNodeTypeProgram ASTNodeType = iota
	ASTNodeTypeExpression
	ASTNodeTypeFunctionDeclaration
	ASTNodeTypeFunctionCall
	ASTNodeTypeVariableDeclaration
	ASTNodeTypeVariableAssignment
	ASTNoteTypeExpression
	ASTNodeTypeIf
	ASTNodeTypeImport
	ASTNodeTypeWhile
	ASTNodeTypeReturn
	ASTNodeTypeOperator
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
}
